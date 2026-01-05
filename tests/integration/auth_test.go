// +build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth/oauth"
	"github.com/stretchr/testify/suite"
)

// AuthIntegrationTestSuite tests OAuth authentication flow functionality.
type AuthIntegrationTestSuite struct {
	suite.Suite
	mockServer *httptest.Server
	cleanup    func()
}

// SetupTest runs before each test in the suite.
func (s *AuthIntegrationTestSuite) SetupTest() {
	// Create mock OAuth server
	mux := http.NewServeMux()

	// Token endpoint handler
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
			return
		}

		grantType := r.Form.Get("grant_type")

		w.Header().Set("Content-Type", "application/json")

		switch grantType {
		case "authorization_code":
			code := r.Form.Get("code")
			codeVerifier := r.Form.Get("code_verifier")

			if code == "" || codeVerifier == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
				return
			}

			// Return successful token response
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "mock_access_token_12345",
				"refresh_token": "mock_refresh_token_67890",
				"token_type":    "Bearer",
				"expires_in":    3600,
			})

		case "refresh_token":
			refreshToken := r.Form.Get("refresh_token")

			if refreshToken == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
				return
			}

			// Simulate token refresh
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "mock_new_access_token_54321",
				"refresh_token": "mock_new_refresh_token_09876",
				"token_type":    "Bearer",
				"expires_in":    3600,
			})

		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "unsupported_grant_type"})
		}
	})

	// Authorization endpoint handler
	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Validate required parameters
		query := r.URL.Query()
		codeChallenge := query.Get("code_challenge")
		challengeMethod := query.Get("code_challenge_method")
		state := query.Get("state")

		if codeChallenge == "" || challengeMethod == "" || state == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Return authorization success page (in real scenario, user would authenticate)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Authorization successful</body></html>"))
	})

	s.mockServer = httptest.NewServer(mux)
	s.cleanup = func() {
		s.mockServer.Close()
	}
}

// TearDownTest runs after each test in the suite.
func (s *AuthIntegrationTestSuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// TestOAuthAuthorizationURLGeneration tests generating OAuth authorization URL with PKCE.
func (s *AuthIntegrationTestSuite) TestOAuthAuthorizationURLGeneration() {
	// Given: An OAuth client configuration
	config := oauth.Config{
		AuthURL:      s.mockServer.URL + "/oauth/authorize",
		TokenURL:     s.mockServer.URL + "/oauth/token",
		ClientID:     "test_client_id",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read", "write"},
		CallbackPort: 8080,
	}

	client := oauth.NewClient(config)

	// When: Generating an authorization URL
	authURL, pkcePair, state, err := client.GetAuthorizationURL()

	// Then: URL should be generated with PKCE parameters
	s.Require().NoError(err, "Failed to generate authorization URL")
	s.NotEmpty(authURL, "Authorization URL should not be empty")
	s.NotNil(pkcePair, "PKCE pair should be generated")
	s.NotEmpty(state, "State parameter should be generated")

	// Verify PKCE code pair structure
	s.NotEmpty(pkcePair.Verifier, "Code verifier should not be empty")
	s.NotEmpty(pkcePair.Challenge, "Code challenge should not be empty")
	s.Equal("S256", pkcePair.ChallengeMethod, "Challenge method should be S256")

	// Verify verifier length requirements
	s.GreaterOrEqual(len(pkcePair.Verifier), 43, "Verifier should be at least 43 characters")
	s.LessOrEqual(len(pkcePair.Verifier), 128, "Verifier should be at most 128 characters")

	// Verify URL contains required parameters
	s.Contains(authURL, "client_id=test_client_id")
	s.Contains(authURL, "code_challenge=")
	s.Contains(authURL, "code_challenge_method=S256")
	s.Contains(authURL, "state="+state)
	s.Contains(authURL, "scope=read+write")
}

// TestPKCECodePairGeneration tests PKCE code verifier and challenge generation.
func (s *AuthIntegrationTestSuite) TestPKCECodePairGeneration() {
	// Given: Default PKCE configuration
	// When: Generating PKCE code pair
	pkcePair, err := oauth.GeneratePKCECodePair()

	// Then: Should generate valid code pair
	s.Require().NoError(err, "Failed to generate PKCE code pair")
	s.NotNil(pkcePair, "PKCE pair should not be nil")
	s.NotEmpty(pkcePair.Verifier, "Verifier should not be empty")
	s.NotEmpty(pkcePair.Challenge, "Challenge should not be empty")
	s.Equal("S256", pkcePair.ChallengeMethod, "Challenge method should be S256")

	// Verify verifier is cryptographically random (generate multiple and compare)
	pkcePair2, err := oauth.GeneratePKCECodePair()
	s.Require().NoError(err)
	s.NotEqual(pkcePair.Verifier, pkcePair2.Verifier, "Verifiers should be unique")
	s.NotEqual(pkcePair.Challenge, pkcePair2.Challenge, "Challenges should be unique")

	// Verify challenge can be validated
	isValid := oauth.VerifyCodeChallenge(pkcePair.Verifier, pkcePair.Challenge)
	s.True(isValid, "Challenge should match verifier")
}

// TestPKCECodePairWithCustomLength tests PKCE generation with custom verifier length.
func (s *AuthIntegrationTestSuite) TestPKCECodePairWithCustomLength() {
	// Given: Custom verifier length
	customLength := 100

	// When: Generating PKCE code pair with custom length
	pkcePair, err := oauth.GeneratePKCECodePairWithLength(customLength)

	// Then: Should generate code pair with specified length
	s.Require().NoError(err, "Failed to generate PKCE code pair with custom length")
	s.Len(pkcePair.Verifier, customLength, "Verifier should have custom length")

	// When: Using invalid length (too short)
	_, err = oauth.GeneratePKCECodePairWithLength(30)

	// Then: Should return error
	s.Error(err, "Should reject verifier length < 43")

	// When: Using invalid length (too long)
	_, err = oauth.GeneratePKCECodePairWithLength(150)

	// Then: Should return error
	s.Error(err, "Should reject verifier length > 128")
}

// TestTokenExchangeWorkflow tests exchanging authorization code for tokens.
func (s *AuthIntegrationTestSuite) TestTokenExchangeWorkflow() {
	// Given: OAuth client and authorization code
	config := oauth.Config{
		AuthURL:      s.mockServer.URL + "/oauth/authorize",
		TokenURL:     s.mockServer.URL + "/oauth/token",
		ClientID:     "test_client_id",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read", "write"},
		CallbackPort: 8080,
	}

	client := oauth.NewClient(config)
	ctx := context.Background()

	// Generate PKCE pair
	pkcePair, err := oauth.GeneratePKCECodePair()
	s.Require().NoError(err)

	// When: Exchanging authorization code for tokens
	authCode := "test_authorization_code"
	tokens, err := client.ExchangeCode(ctx, authCode, pkcePair.Verifier)

	// Then: Should receive token pair
	s.Require().NoError(err, "Failed to exchange code for tokens")
	s.NotNil(tokens, "Token pair should not be nil")
	s.Equal("mock_access_token_12345", tokens.AccessToken, "Access token should match")
	s.Equal("mock_refresh_token_67890", tokens.RefreshToken, "Refresh token should match")
	s.Equal("Bearer", tokens.TokenType, "Token type should be Bearer")
	s.Equal(int64(3600), tokens.ExpiresIn, "Expires in should be 3600 seconds")
}

// TestTokenRefreshWorkflow tests refreshing access tokens.
func (s *AuthIntegrationTestSuite) TestTokenRefreshWorkflow() {
	// Given: OAuth client and existing refresh token
	config := oauth.Config{
		AuthURL:      s.mockServer.URL + "/oauth/authorize",
		TokenURL:     s.mockServer.URL + "/oauth/token",
		ClientID:     "test_client_id",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read", "write"},
		CallbackPort: 8080,
	}

	client := oauth.NewClient(config)
	ctx := context.Background()
	existingRefreshToken := "mock_refresh_token_67890"

	// When: Refreshing the access token
	newTokens, err := client.RefreshToken(ctx, existingRefreshToken)

	// Then: Should receive new token pair
	s.Require().NoError(err, "Failed to refresh token")
	s.NotNil(newTokens, "New token pair should not be nil")
	s.Equal("mock_new_access_token_54321", newTokens.AccessToken, "New access token should be returned")
	s.Equal("mock_new_refresh_token_09876", newTokens.RefreshToken, "New refresh token should be returned")
	s.Equal("Bearer", newTokens.TokenType, "Token type should be Bearer")
	s.Equal(int64(3600), newTokens.ExpiresIn, "Expires in should be 3600 seconds")
}

// TestTokenValidation tests validating JWT token structure.
func (s *AuthIntegrationTestSuite) TestTokenValidation() {
	// Given: OAuth client and tokens from token exchange
	config := oauth.Config{
		AuthURL:      s.mockServer.URL + "/oauth/authorize",
		TokenURL:     s.mockServer.URL + "/oauth/token",
		ClientID:     "test_client_id",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read", "write"},
		CallbackPort: 8080,
	}

	client := oauth.NewClient(config)
	ctx := context.Background()

	pkcePair, err := oauth.GeneratePKCECodePair()
	s.Require().NoError(err)

	// When: Exchanging code for tokens
	tokens, err := client.ExchangeCode(ctx, "test_code", pkcePair.Verifier)

	// Then: Tokens should be returned and have valid structure
	s.Require().NoError(err, "Token exchange should succeed")
	s.NotNil(tokens, "Tokens should not be nil")
	s.NotEmpty(tokens.AccessToken, "Access token should not be empty")
	s.NotEmpty(tokens.RefreshToken, "Refresh token should not be empty")
	s.Equal("Bearer", tokens.TokenType, "Token type should be Bearer")
	s.Greater(tokens.ExpiresIn, int64(0), "Expires in should be positive")
}

// TestTokenExpirationHandling tests handling token expiration information.
func (s *AuthIntegrationTestSuite) TestTokenExpirationHandling() {
	// Given: OAuth client
	config := oauth.Config{
		AuthURL:      s.mockServer.URL + "/oauth/authorize",
		TokenURL:     s.mockServer.URL + "/oauth/token",
		ClientID:     "test_client_id",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read", "write"},
		CallbackPort: 8080,
	}

	client := oauth.NewClient(config)
	ctx := context.Background()

	pkcePair, err := oauth.GeneratePKCECodePair()
	s.Require().NoError(err)

	// When: Getting tokens
	tokens, err := client.ExchangeCode(ctx, "test_code", pkcePair.Verifier)
	s.Require().NoError(err)

	// Then: Should include expiration information
	s.Equal(int64(3600), tokens.ExpiresIn, "Should have expiration time")

	// When: Calculating expiration time
	expirationTime := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)

	// Then: Expiration should be in the future
	s.True(expirationTime.After(time.Now()), "Expiration should be in future")
}

// TestInvalidTokenRejection tests rejecting invalid OAuth responses.
func (s *AuthIntegrationTestSuite) TestInvalidTokenRejection() {
	// Given: OAuth client with invalid server
	config := oauth.Config{
		AuthURL:      "http://invalid-server.local/oauth/authorize",
		TokenURL:     "http://invalid-server.local/oauth/token",
		ClientID:     "test_client_id",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read"},
		CallbackPort: 8080,
		HTTPClient: &http.Client{
			Timeout: 1 * time.Second,
		},
	}

	client := oauth.NewClient(config)
	ctx := context.Background()

	pkcePair, err := oauth.GeneratePKCECodePair()
	s.Require().NoError(err)

	// When: Attempting to exchange code with invalid server
	_, err = client.ExchangeCode(ctx, "test_code", pkcePair.Verifier)

	// Then: Should return error
	s.Error(err, "Should reject invalid token endpoint")
}

// TestOAuthErrorHandling tests error handling in OAuth flows.
func (s *AuthIntegrationTestSuite) TestOAuthErrorHandling() {
	// Given: OAuth client with invalid configuration
	config := oauth.Config{
		AuthURL:      "http://invalid-server.local/oauth/authorize",
		TokenURL:     "http://invalid-server.local/oauth/token",
		ClientID:     "test_client_id",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read"},
		CallbackPort: 8080,
		HTTPClient: &http.Client{
			Timeout: 1 * time.Second,
		},
	}

	client := oauth.NewClient(config)
	ctx := context.Background()

	pkcePair, err := oauth.GeneratePKCECodePair()
	s.Require().NoError(err)

	// When: Attempting to exchange code with invalid server
	_, err = client.ExchangeCode(ctx, "test_code", pkcePair.Verifier)

	// Then: Should return network error
	s.Error(err, "Should fail with invalid server")

	// When: Attempting to refresh token with invalid server
	_, err = client.RefreshToken(ctx, "test_refresh_token")

	// Then: Should return network error
	s.Error(err, "Should fail to refresh token with invalid server")
}

// TestConcurrentTokenOperations tests thread safety of token operations.
func (s *AuthIntegrationTestSuite) TestConcurrentTokenOperations() {
	// Given: OAuth client
	config := oauth.Config{
		AuthURL:      s.mockServer.URL + "/oauth/authorize",
		TokenURL:     s.mockServer.URL + "/oauth/token",
		ClientID:     "test_client_id",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read", "write"},
		CallbackPort: 8080,
	}

	client := oauth.NewClient(config)
	ctx := context.Background()

	// When: Making concurrent token refresh requests
	concurrentOps := 10
	done := make(chan bool, concurrentOps)
	errors := make(chan error, concurrentOps)

	for i := 0; i < concurrentOps; i++ {
		go func() {
			_, err := client.RefreshToken(ctx, "test_refresh_token")
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	// Wait for all operations to complete
	for i := 0; i < concurrentOps; i++ {
		<-done
	}
	close(errors)

	// Then: All operations should succeed
	s.Empty(errors, "No errors should occur during concurrent operations")
}

// TestAuthIntegrationTestSuite runs the test suite.
func TestAuthIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AuthIntegrationTestSuite))
}
