package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth/jwt"
)

const (
	// DefaultCallbackPort is the default port for the OAuth callback server
	DefaultCallbackPort = 8080

	// DefaultTimeout is the default timeout for HTTP requests
	DefaultTimeout = 30 * time.Second

	// DefaultCallbackPath is the default path for the OAuth callback
	DefaultCallbackPath = "/callback"
)

// Config represents the OAuth client configuration.
type Config struct {
	// AuthURL is the authorization endpoint URL
	AuthURL string

	// TokenURL is the token endpoint URL
	TokenURL string

	// ClientID is the OAuth client identifier
	ClientID string

	// RedirectURL is the callback URL (e.g., http://localhost:8080/callback)
	RedirectURL string

	// Scopes are the requested OAuth scopes
	Scopes []string

	// CallbackPort is the port for the local callback server (default: 8080)
	CallbackPort int

	// HTTPClient is the HTTP client for token requests (optional)
	HTTPClient *http.Client
}

// Client is an OAuth 2.0 client with PKCE support.
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new OAuth client with PKCE support.
func NewClient(config Config) *Client {
	if config.CallbackPort == 0 {
		config.CallbackPort = DefaultCallbackPort
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: DefaultTimeout,
		}
	}

	return &Client{
		config:     config,
		httpClient: config.HTTPClient,
	}
}

// Authenticate performs the OAuth 2.0 authorization code flow with PKCE.
//
// This method:
//  1. Generates PKCE code verifier and challenge
//  2. Builds authorization URL and opens it in browser
//  3. Starts local callback server to receive authorization code
//  4. Exchanges authorization code for tokens
//  5. Returns the token pair
//
// The user must authorize the application in their browser.
func (c *Client) Authenticate(ctx context.Context) (*jwt.TokenPair, error) {
	// Generate PKCE code pair
	pkcePair, err := GeneratePKCECodePair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE code pair: %w", err)
	}

	// Generate state parameter for CSRF protection
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Build authorization URL
	authURL := c.buildAuthorizationURL(pkcePair.Challenge, state)

	// Print authorization URL for user to open
	// In a real implementation, this would open the browser automatically
	fmt.Printf("Please visit this URL to authorize:\n%s\n", authURL)

	// Start callback server and get authorization code
	code, receivedState, err := c.startCallbackServer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to receive authorization code: %w", err)
	}

	// Verify state to prevent CSRF
	if receivedState != state {
		return nil, fmt.Errorf("state mismatch: possible CSRF attack")
	}

	// Exchange authorization code for tokens
	tokens, err := c.exchangeCodeForTokens(ctx, code, pkcePair.Verifier)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for tokens: %w", err)
	}

	return tokens, nil
}

// buildAuthorizationURL builds the OAuth authorization URL with PKCE parameters.
func (c *Client) buildAuthorizationURL(codeChallenge, state string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", c.config.ClientID)
	params.Set("redirect_uri", c.config.RedirectURL)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", PKCEChallengeMethod)
	params.Set("state", state)

	if len(c.config.Scopes) > 0 {
		params.Set("scope", strings.Join(c.config.Scopes, " "))
	}

	return c.config.AuthURL + "?" + params.Encode()
}

// startCallbackServer starts a local HTTP server to receive the authorization code.
func (c *Client) startCallbackServer(ctx context.Context) (code, state string, err error) {
	// Parse redirect URL to get path
	redirectURL, err := url.Parse(c.config.RedirectURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid redirect URL: %w", err)
	}

	callbackPath := redirectURL.Path
	if callbackPath == "" {
		callbackPath = DefaultCallbackPath
	}

	// Create channels for code and errors
	codeChan := make(chan string, 1)
	stateChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		// Extract code and state from query parameters
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if code == "" {
			errMsg := r.URL.Query().Get("error")
			if errMsg == "" {
				errMsg = "no authorization code received"
			}
			errChan <- fmt.Errorf("authorization failed: %s", errMsg)
			http.Error(w, "Authorization failed", http.StatusBadRequest)
			return
		}

		// Send code and state
		codeChan <- code
		stateChan <- state

		// Send success response
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Successful</title>
    <style>
        body { font-family: sans-serif; text-align: center; padding: 50px; }
        .success { color: #4CAF50; }
    </style>
</head>
<body>
    <h1 class="success">✓ Authentication Successful</h1>
    <p>You can close this window and return to the CLI.</p>
</body>
</html>
`)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", c.config.CallbackPort),
		Handler: mux,
		// Security: Prevent Slowloris attacks
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	// Wait for code or error
	select {
	case code = <-codeChan:
		state = <-stateChan
		// Shutdown server
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
		return code, state, nil
	case err = <-errChan:
		server.Shutdown(context.Background())
		return "", "", err
	case <-ctx.Done():
		server.Shutdown(context.Background())
		return "", "", ctx.Err()
	}
}

// exchangeCodeForTokens exchanges an authorization code for access and refresh tokens.
func (c *Client) exchangeCodeForTokens(ctx context.Context, code, codeVerifier string) (*jwt.TokenPair, error) {
	// Build token request
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.config.RedirectURL)
	data.Set("client_id", c.config.ClientID)
	data.Set("code_verifier", codeVerifier)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse token response
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Convert to TokenPair
	tokens := &jwt.TokenPair{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		TokenType:    tokenResp.TokenType,
	}

	return tokens, nil
}

// RefreshToken refreshes an access token using a refresh token.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	// Build refresh request
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", c.config.ClientID)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse token response
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Convert to TokenPair
	tokens := &jwt.TokenPair{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		TokenType:    tokenResp.TokenType,
	}

	return tokens, nil
}

// GetAuthorizationURL returns the authorization URL for manual flow.
func (c *Client) GetAuthorizationURL() (string, *PKCECodePair, string, error) {
	pkcePair, err := GeneratePKCECodePair()
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to generate PKCE code pair: %w", err)
	}

	state, err := generateState()
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to generate state: %w", err)
	}

	authURL := c.buildAuthorizationURL(pkcePair.Challenge, state)

	return authURL, pkcePair, state, nil
}

// ExchangeCode exchanges an authorization code for tokens (for manual flow).
func (c *Client) ExchangeCode(ctx context.Context, code, codeVerifier string) (*jwt.TokenPair, error) {
	return c.exchangeCodeForTokens(ctx, code, codeVerifier)
}

// TokenResponse represents the OAuth token endpoint response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
}

// generateState generates a random state parameter for CSRF protection.
func generateState() (string, error) {
	return generateCodeVerifier(32)
}
