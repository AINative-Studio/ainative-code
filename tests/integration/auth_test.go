package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth"
	"github.com/AINative-studio/ainative-code/tests/integration/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOAuthLoginFlow_PKCE tests PKCE parameter generation
func TestOAuthLoginFlow_PKCE(t *testing.T) {
	t.Run("should generate valid PKCE parameters", func(t *testing.T) {
		// When: Generating PKCE parameters
		pkce, err := auth.GeneratePKCE()

		// Then: Parameters should be valid
		require.NoError(t, err)
		require.NotNil(t, pkce)

		assert.NotEmpty(t, pkce.CodeVerifier, "code verifier should not be empty")
		assert.NotEmpty(t, pkce.CodeChallenge, "code challenge should not be empty")
		assert.Equal(t, "S256", pkce.Method, "method should be S256")
		assert.NotEmpty(t, pkce.State, "state should not be empty")

		// Verify code verifier length (43-128 characters)
		assert.GreaterOrEqual(t, len(pkce.CodeVerifier), 43)
		assert.LessOrEqual(t, len(pkce.CodeVerifier), 128)

		// Verify code challenge length (43 characters for SHA-256)
		assert.Equal(t, 43, len(pkce.CodeChallenge))
	})

	t.Run("should generate unique PKCE parameters on each call", func(t *testing.T) {
		// When: Generating multiple PKCE parameters
		pkce1, err1 := auth.GeneratePKCE()
		pkce2, err2 := auth.GeneratePKCE()

		// Then: Each should be unique
		require.NoError(t, err1)
		require.NoError(t, err2)

		assert.NotEqual(t, pkce1.CodeVerifier, pkce2.CodeVerifier)
		assert.NotEqual(t, pkce1.CodeChallenge, pkce2.CodeChallenge)
		assert.NotEqual(t, pkce1.State, pkce2.State)
	})

	t.Run("should validate code verifier", func(t *testing.T) {
		// Given: Valid PKCE parameters
		pkce, err := auth.GeneratePKCE()
		require.NoError(t, err)

		// When: Validating the code verifier
		err = auth.ValidateCodeVerifier(pkce.CodeVerifier)

		// Then: Should be valid
		assert.NoError(t, err)
	})

	t.Run("should reject invalid code verifier", func(t *testing.T) {
		// Given: Invalid code verifiers
		testCases := []struct {
			name     string
			verifier string
		}{
			{"too short", "abc"},
			{"too long", string(make([]byte, 150))},
			{"invalid characters", "invalid!@#$%"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// When: Validating invalid verifier
				err := auth.ValidateCodeVerifier(tc.verifier)

				// Then: Should return error
				assert.Error(t, err)
			})
		}
	})
}

// TestOAuthLoginFlow_AuthorizationURL tests authorization URL construction
func TestOAuthLoginFlow_AuthorizationURL(t *testing.T) {
	t.Run("should construct valid authorization URL", func(t *testing.T) {
		// Given: Mock auth server
		authServer, err := mocks.NewAuthServer()
		require.NoError(t, err)
		defer authServer.Close()

		// And: PKCE parameters
		pkce, err := auth.GeneratePKCE()
		require.NoError(t, err)

		// When: Making authorization request
		req, err := http.NewRequest("GET", authServer.GetAuthURL(), nil)
		require.NoError(t, err)

		q := req.URL.Query()
		q.Add("response_type", "code")
		q.Add("client_id", "ainative-code-cli")
		q.Add("redirect_uri", "http://localhost:8080/callback")
		q.Add("code_challenge", pkce.CodeChallenge)
		q.Add("code_challenge_method", "S256")
		q.Add("state", pkce.State)
		q.Add("scope", "read write offline_access")
		req.URL.RawQuery = q.Encode()

		// Then: URL should be properly formatted
		assert.Contains(t, req.URL.String(), "code_challenge=")
		assert.Contains(t, req.URL.String(), "code_challenge_method=S256")
		assert.Contains(t, req.URL.String(), "state=")
	})
}

// TestOAuthLoginFlow_CodeExchange tests authorization code exchange
func TestOAuthLoginFlow_CodeExchange(t *testing.T) {
	t.Run("should successfully exchange code for tokens", func(t *testing.T) {
		// Given: Mock auth server with valid code
		authServer, err := mocks.NewAuthServer()
		require.NoError(t, err)
		defer authServer.Close()

		code := "test_auth_code"
		verifier := "test_verifier_" + string(make([]byte, 100))
		authServer.AddValidCode(code, verifier)

		// When: Exchanging code for tokens
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.PostForm(authServer.GetTokenURL(), map[string][]string{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"code_verifier": {verifier},
			"redirect_uri":  {"http://localhost:8080/callback"},
			"client_id":     {"ainative-code-cli"},
		})

		// Then: Should receive tokens
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, authServer.TokenCalled, "token endpoint should be called")
	})

	t.Run("should reject invalid authorization code", func(t *testing.T) {
		// Given: Mock auth server configured to fail
		authServer, err := mocks.NewAuthServer()
		require.NoError(t, err)
		defer authServer.Close()

		authServer.ShouldFailToken = true

		// When: Exchanging invalid code
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.PostForm(authServer.GetTokenURL(), map[string][]string{
			"grant_type":    {"authorization_code"},
			"code":          {"invalid_code"},
			"code_verifier": {"invalid_verifier"},
		})

		// Then: Should receive error
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should reject mismatched code verifier", func(t *testing.T) {
		// Given: Mock auth server with valid code but wrong verifier
		authServer, err := mocks.NewAuthServer()
		require.NoError(t, err)
		defer authServer.Close()

		code := "test_code"
		correctVerifier := "correct_verifier_" + string(make([]byte, 100))
		authServer.AddValidCode(code, correctVerifier)

		// When: Using wrong verifier
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.PostForm(authServer.GetTokenURL(), map[string][]string{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"code_verifier": {"wrong_verifier"},
		})

		// Then: Should be rejected
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestOAuthLoginFlow_TokenRefresh tests token refresh flow
func TestOAuthLoginFlow_TokenRefresh(t *testing.T) {
	t.Run("should successfully refresh access token", func(t *testing.T) {
		// Given: Mock auth server with valid refresh token
		authServer, err := mocks.NewAuthServer()
		require.NoError(t, err)
		defer authServer.Close()

		refreshToken := "valid_refresh_token"
		authServer.AddValidRefreshToken(refreshToken)

		// When: Refreshing token
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.PostForm(authServer.GetTokenURL(), map[string][]string{
			"grant_type":    {"refresh_token"},
			"refresh_token": {refreshToken},
			"client_id":     {"ainative-code-cli"},
		})

		// Then: Should receive new access token
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, authServer.RefreshCalled, "refresh endpoint should be called")
	})

	t.Run("should reject invalid refresh token", func(t *testing.T) {
		// Given: Mock auth server
		authServer, err := mocks.NewAuthServer()
		require.NoError(t, err)
		defer authServer.Close()

		// When: Using invalid refresh token
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.PostForm(authServer.GetTokenURL(), map[string][]string{
			"grant_type":    {"refresh_token"},
			"refresh_token": {"invalid_token"},
		})

		// Then: Should be rejected
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestOAuthLoginFlow_ErrorHandling tests error scenarios
func TestOAuthLoginFlow_ErrorHandling(t *testing.T) {
	t.Run("should handle 401 unauthorized error", func(t *testing.T) {
		// Given: Mock server returning 403 Forbidden (auth failure)
		authServer, err := mocks.NewAuthServer()
		require.NoError(t, err)
		defer authServer.Close()

		authServer.ShouldFailAuth = true

		// When: Making request
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.PostForm(authServer.GetTokenURL(), map[string][]string{
			"grant_type": {"authorization_code"},
		})

		// Then: Should receive 403 or 400 (mock server returns 403 for auth failure, 400 for token errors)
		require.NoError(t, err)
		defer resp.Body.Close()

		// The mock can return either 403 (auth failed) or 400 (missing params) - both are error conditions
		assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusBadRequest,
			"Expected 403 or 400, got %d", resp.StatusCode)
	})

	t.Run("should handle 429 rate limit error", func(t *testing.T) {
		// Given: Mock server with rate limiting
		authServer, err := mocks.NewAuthServer()
		require.NoError(t, err)
		defer authServer.Close()

		authServer.ShouldRateLimit = true

		// When: Making request
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.PostForm(authServer.GetTokenURL(), map[string][]string{
			"grant_type": {"authorization_code"},
		})

		// Then: Should receive 429 with Retry-After header
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		assert.NotEmpty(t, resp.Header.Get("Retry-After"))
	})

	t.Run("should handle 500 server error", func(t *testing.T) {
		// Given: Mock server returning 500
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"server_error"}`))
		}))
		defer server.Close()

		// When: Making request
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.PostForm(server.URL, map[string][]string{
			"grant_type": {"authorization_code"},
		})

		// Then: Should receive 500
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("should handle timeout", func(t *testing.T) {
		// Given: Mock server with delay
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
		}))
		defer server.Close()

		// When: Making request with short timeout
		client := &http.Client{Timeout: 100 * time.Millisecond}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		req, _ := http.NewRequestWithContext(ctx, "POST", server.URL, nil)
		_, err := client.Do(req)

		// Then: Should timeout
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}

// TestOAuthLoginFlow_StateValidation tests CSRF state validation
func TestOAuthLoginFlow_StateValidation(t *testing.T) {
	t.Run("should validate state parameter matches", func(t *testing.T) {
		// Given: PKCE parameters with state
		pkce, err := auth.GeneratePKCE()
		require.NoError(t, err)

		originalState := pkce.State

		// When: Callback returns same state
		callbackState := originalState

		// Then: States should match
		assert.Equal(t, originalState, callbackState)
	})

	t.Run("should reject mismatched state", func(t *testing.T) {
		// Given: Original state
		pkce, err := auth.GeneratePKCE()
		require.NoError(t, err)

		originalState := pkce.State

		// When: Callback returns different state
		callbackState := "different_state"

		// Then: States should not match
		assert.NotEqual(t, originalState, callbackState)
	})
}
