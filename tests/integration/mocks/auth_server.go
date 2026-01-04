package mocks

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthServer represents a mock OAuth 2.0 authorization server
type AuthServer struct {
	Server           *httptest.Server
	PrivateKey       *rsa.PrivateKey
	PublicKey        *rsa.PublicKey
	AuthorizeCalled  bool
	TokenCalled      bool
	RefreshCalled    bool
	ValidCodes       map[string]string // code -> verifier mapping
	ValidRefreshTokens map[string]bool
	ShouldFailAuth   bool
	ShouldFailToken  bool
	ShouldRateLimit  bool
	ResponseDelay    time.Duration
}

// NewAuthServer creates a new mock authentication server
func NewAuthServer() (*AuthServer, error) {
	// Generate RSA key pair for JWT signing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	as := &AuthServer{
		PrivateKey:         privateKey,
		PublicKey:          &privateKey.PublicKey,
		ValidCodes:         make(map[string]string),
		ValidRefreshTokens: make(map[string]bool),
	}

	// Create test server with handler
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/authorize", as.handleAuthorize)
	mux.HandleFunc("/oauth/token", as.handleToken)
	mux.HandleFunc("/.well-known/jwks.json", as.handleJWKS)

	as.Server = httptest.NewServer(mux)
	return as, nil
}

// Close shuts down the mock server
func (as *AuthServer) Close() {
	as.Server.Close()
}

// AddValidCode adds a valid authorization code for testing
func (as *AuthServer) AddValidCode(code, verifier string) {
	as.ValidCodes[code] = verifier
}

// AddValidRefreshToken adds a valid refresh token for testing
func (as *AuthServer) AddValidRefreshToken(token string) {
	as.ValidRefreshTokens[token] = true
}

// handleAuthorize simulates the OAuth authorization endpoint
func (as *AuthServer) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	as.AuthorizeCalled = true

	if as.ResponseDelay > 0 {
		time.Sleep(as.ResponseDelay)
	}

	if as.ShouldFailAuth {
		http.Error(w, `{"error":"access_denied","error_description":"User denied authorization"}`, http.StatusForbidden)
		return
	}

	// Extract parameters
	codeChallenge := r.URL.Query().Get("code_challenge")
	challengeMethod := r.URL.Query().Get("code_challenge_method")
	state := r.URL.Query().Get("state")
	redirectURI := r.URL.Query().Get("redirect_uri")

	// Validate required parameters
	if codeChallenge == "" || challengeMethod != "S256" || state == "" {
		http.Error(w, `{"error":"invalid_request","error_description":"Missing required parameters"}`, http.StatusBadRequest)
		return
	}

	// Generate mock authorization code
	code := "mock_auth_code_" + generateRandomString(32)

	// Redirect to callback URL with code and state
	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, state)
	w.Header().Set("Location", redirectURL)
	w.WriteHeader(http.StatusFound)
}

// handleToken simulates the OAuth token endpoint
func (as *AuthServer) handleToken(w http.ResponseWriter, r *http.Request) {
	as.TokenCalled = true

	if as.ResponseDelay > 0 {
		time.Sleep(as.ResponseDelay)
	}

	if as.ShouldRateLimit {
		w.Header().Set("Retry-After", "60")
		http.Error(w, `{"error":"rate_limit_exceeded","error_description":"Too many requests"}`, http.StatusTooManyRequests)
		return
	}

	if as.ShouldFailToken {
		http.Error(w, `{"error":"invalid_grant","error_description":"Invalid authorization code"}`, http.StatusBadRequest)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		as.handleAuthorizationCodeGrant(w, r)
	case "refresh_token":
		as.handleRefreshTokenGrant(w, r)
	default:
		http.Error(w, `{"error":"unsupported_grant_type"}`, http.StatusBadRequest)
	}
}

// handleAuthorizationCodeGrant handles authorization code grant
func (as *AuthServer) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	verifier := r.FormValue("code_verifier")

	// Validate code and verifier
	expectedVerifier, exists := as.ValidCodes[code]
	if !exists || (expectedVerifier != "" && expectedVerifier != verifier) {
		http.Error(w, `{"error":"invalid_grant","error_description":"Invalid code or verifier"}`, http.StatusBadRequest)
		return
	}

	// Remove used code
	delete(as.ValidCodes, code)

	// Generate tokens
	accessToken, err := as.generateAccessToken("user123", "test@ainative.studio", []string{"read", "write"})
	if err != nil {
		http.Error(w, `{"error":"server_error"}`, http.StatusInternalServerError)
		return
	}

	refreshToken, err := as.generateRefreshToken("user123", "session123")
	if err != nil {
		http.Error(w, `{"error":"server_error"}`, http.StatusInternalServerError)
		return
	}

	// Store refresh token as valid
	as.AddValidRefreshToken(refreshToken)

	response := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"scope":         "read write offline_access",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRefreshTokenGrant handles refresh token grant
func (as *AuthServer) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	as.RefreshCalled = true

	refreshToken := r.FormValue("refresh_token")

	// Validate refresh token
	if !as.ValidRefreshTokens[refreshToken] {
		http.Error(w, `{"error":"invalid_grant","error_description":"Invalid refresh token"}`, http.StatusBadRequest)
		return
	}

	// Generate new access token
	accessToken, err := as.generateAccessToken("user123", "test@ainative.studio", []string{"read", "write"})
	if err != nil {
		http.Error(w, `{"error":"server_error"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        "read write offline_access",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleJWKS simulates the JWKS endpoint
func (as *AuthServer) handleJWKS(w http.ResponseWriter, r *http.Request) {
	// Return mock JWKS (simplified)
	response := map[string]interface{}{
		"keys": []interface{}{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// generateAccessToken creates a mock JWT access token
func (as *AuthServer) generateAccessToken(userID, email string, roles []string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"roles": roles,
		"iss":   "ainative-auth",
		"aud":   "ainative-code",
		"exp":   now.Add(1 * time.Hour).Unix(),
		"iat":   now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(as.PrivateKey)
}

// generateRefreshToken creates a mock JWT refresh token
func (as *AuthServer) generateRefreshToken(userID, sessionID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":        userID,
		"session_id": sessionID,
		"iss":        "ainative-auth",
		"aud":        "ainative-code",
		"exp":        now.Add(7 * 24 * time.Hour).Unix(),
		"iat":        now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(as.PrivateKey)
}

// generateRandomString generates a random string for testing
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// SimulateServerError configures the server to return 500 errors
func (as *AuthServer) SimulateServerError() {
	as.Server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"server_error","error_description":"Internal server error"}`, http.StatusInternalServerError)
	})
}

// SimulateTimeout configures the server to delay responses
func (as *AuthServer) SimulateTimeout(delay time.Duration) {
	as.ResponseDelay = delay
}

// GetAuthURL returns the authorization endpoint URL
func (as *AuthServer) GetAuthURL() string {
	return as.Server.URL + "/oauth/authorize"
}

// GetTokenURL returns the token endpoint URL
func (as *AuthServer) GetTokenURL() string {
	return as.Server.URL + "/oauth/token"
}
