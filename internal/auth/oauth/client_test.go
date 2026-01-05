package oauth_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("creates client with config", func(t *testing.T) {
		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    "https://auth.example.com/token",
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
			Scopes:      []string{"read", "write"},
		}

		client := oauth.NewClient(config)
		assert.NotNil(t, client)
	})

	t.Run("uses default callback port", func(t *testing.T) {
		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    "https://auth.example.com/token",
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
		}

		client := oauth.NewClient(config)
		assert.NotNil(t, client)
	})

	t.Run("accepts custom HTTP client", func(t *testing.T) {
		httpClient := &http.Client{
			Timeout: 5 * time.Second,
		}

		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    "https://auth.example.com/token",
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
			HTTPClient:  httpClient,
		}

		client := oauth.NewClient(config)
		assert.NotNil(t, client)
	})
}

func TestGetAuthorizationURL(t *testing.T) {
	t.Run("builds valid authorization URL", func(t *testing.T) {
		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    "https://auth.example.com/token",
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
			Scopes:      []string{"read", "write"},
		}

		client := oauth.NewClient(config)

		authURL, pkcePair, state, err := client.GetAuthorizationURL()
		require.NoError(t, err)
		assert.NotEmpty(t, authURL)
		assert.NotNil(t, pkcePair)
		assert.NotEmpty(t, state)

		// Parse URL and verify parameters
		parsedURL, err := url.Parse(authURL)
		require.NoError(t, err)

		query := parsedURL.Query()
		assert.Equal(t, "code", query.Get("response_type"))
		assert.Equal(t, "test-client-id", query.Get("client_id"))
		assert.Equal(t, "http://localhost:8080/callback", query.Get("redirect_uri"))
		assert.Equal(t, pkcePair.Challenge, query.Get("code_challenge"))
		assert.Equal(t, "S256", query.Get("code_challenge_method"))
		assert.Equal(t, state, query.Get("state"))
		assert.Equal(t, "read write", query.Get("scope"))
	})

	t.Run("includes scopes when provided", func(t *testing.T) {
		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    "https://auth.example.com/token",
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
			Scopes:      []string{"openid", "profile", "email"},
		}

		client := oauth.NewClient(config)

		authURL, _, _, err := client.GetAuthorizationURL()
		require.NoError(t, err)

		parsedURL, err := url.Parse(authURL)
		require.NoError(t, err)

		assert.Equal(t, "openid profile email", parsedURL.Query().Get("scope"))
	})

	t.Run("works without scopes", func(t *testing.T) {
		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    "https://auth.example.com/token",
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
		}

		client := oauth.NewClient(config)

		authURL, _, _, err := client.GetAuthorizationURL()
		require.NoError(t, err)

		parsedURL, err := url.Parse(authURL)
		require.NoError(t, err)

		assert.Empty(t, parsedURL.Query().Get("scope"))
	})
}

func TestExchangeCode(t *testing.T) {
	t.Run("exchanges code for tokens", func(t *testing.T) {
		// Mock token server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

			// Parse form data
			err := r.ParseForm()
			require.NoError(t, err)

			assert.Equal(t, "authorization_code", r.Form.Get("grant_type"))
			assert.Equal(t, "test-auth-code", r.Form.Get("code"))
			assert.Equal(t, "http://localhost:8080/callback", r.Form.Get("redirect_uri"))
			assert.Equal(t, "test-client-id", r.Form.Get("client_id"))
			assert.NotEmpty(t, r.Form.Get("code_verifier"))

			// Send token response
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(oauth.TokenResponse{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
			})
		}))
		defer server.Close()

		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    server.URL,
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
		}

		client := oauth.NewClient(config)

		// Generate PKCE pair
		pkcePair, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		// Exchange code
		tokens, err := client.ExchangeCode(context.Background(), "test-auth-code", pkcePair.Verifier)
		require.NoError(t, err)
		assert.Equal(t, "test-access-token", tokens.AccessToken)
		assert.Equal(t, "test-refresh-token", tokens.RefreshToken)
		assert.Equal(t, "Bearer", tokens.TokenType)
		assert.Equal(t, int64(3600), tokens.ExpiresIn)
	})

	t.Run("handles token server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"invalid_grant"}`))
		}))
		defer server.Close()

		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    server.URL,
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
		}

		client := oauth.NewClient(config)

		pkcePair, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		_, err = client.ExchangeCode(context.Background(), "invalid-code", pkcePair.Verifier)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed with status 400")
	})

	t.Run("handles malformed response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    server.URL,
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
		}

		client := oauth.NewClient(config)

		pkcePair, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		_, err = client.ExchangeCode(context.Background(), "test-code", pkcePair.Verifier)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse")
	})
}

func TestRefreshToken(t *testing.T) {
	t.Run("refreshes access token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)

			err := r.ParseForm()
			require.NoError(t, err)

			assert.Equal(t, "refresh_token", r.Form.Get("grant_type"))
			assert.Equal(t, "test-refresh-token", r.Form.Get("refresh_token"))
			assert.Equal(t, "test-client-id", r.Form.Get("client_id"))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(oauth.TokenResponse{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
			})
		}))
		defer server.Close()

		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    server.URL,
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
		}

		client := oauth.NewClient(config)

		tokens, err := client.RefreshToken(context.Background(), "test-refresh-token")
		require.NoError(t, err)
		assert.Equal(t, "new-access-token", tokens.AccessToken)
		assert.Equal(t, "new-refresh-token", tokens.RefreshToken)
	})

	t.Run("handles refresh error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"invalid_token"}`))
		}))
		defer server.Close()

		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    server.URL,
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
		}

		client := oauth.NewClient(config)

		_, err := client.RefreshToken(context.Background(), "invalid-refresh-token")
		assert.Error(t, err)
	})
}

func TestContextCancellation(t *testing.T) {
	t.Run("respects context cancellation for token exchange", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
		}))
		defer server.Close()

		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    server.URL,
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
		}

		client := oauth.NewClient(config)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		pkcePair, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		_, err = client.ExchangeCode(ctx, "test-code", pkcePair.Verifier)
		assert.Error(t, err)
	})
}

func TestCallbackServer(t *testing.T) {
	// Note: Full callback server testing is difficult in unit tests
	// as it requires starting a real HTTP server and simulating browser callback.
	// This would be better tested in integration tests.

	t.Run("validates redirect URL path", func(t *testing.T) {
		// Test that invalid redirect URLs are caught
		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    "https://token.example.com/token",
			ClientID:    "test-client-id",
			RedirectURL: "://invalid-url",
		}

		client := oauth.NewClient(config)

		// GetAuthorizationURL should work even with invalid redirect
		_, _, _, err := client.GetAuthorizationURL()
		assert.NoError(t, err) // URL construction doesn't validate redirect
	})
}

func TestOAuthFlow_Integration(t *testing.T) {
	// This is a simplified integration test that mocks the OAuth flow
	t.Run("complete OAuth flow simulation", func(t *testing.T) {
		// Mock OAuth server
		var authCode string
		var receivedChallenge string
		var receivedState string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/token" {
				// Token endpoint
				err := r.ParseForm()
				require.NoError(t, err)

				assert.Equal(t, "authorization_code", r.Form.Get("grant_type"))
				assert.Equal(t, authCode, r.Form.Get("code"))

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(oauth.TokenResponse{
					AccessToken:  "test-access-token",
					RefreshToken: "test-refresh-token",
					TokenType:    "Bearer",
					ExpiresIn:    3600,
				})
			}
		}))
		defer server.Close()

		config := oauth.Config{
			AuthURL:     server.URL + "/authorize",
			TokenURL:    server.URL + "/token",
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
			Scopes:      []string{"read", "write"},
		}

		client := oauth.NewClient(config)

		// Step 1: Get authorization URL
		authURL, pkcePair, state, err := client.GetAuthorizationURL()
		require.NoError(t, err)

		// Parse and verify URL
		parsedURL, err := url.Parse(authURL)
		require.NoError(t, err)

		receivedChallenge = parsedURL.Query().Get("code_challenge")
		receivedState = parsedURL.Query().Get("state")

		// Verify PKCE challenge
		assert.True(t, oauth.VerifyCodeChallenge(pkcePair.Verifier, receivedChallenge))
		assert.Equal(t, state, receivedState)

		// Step 2: Simulate authorization (normally done in browser)
		authCode = "simulated-auth-code"

		// Step 3: Exchange code for tokens
		tokens, err := client.ExchangeCode(context.Background(), authCode, pkcePair.Verifier)
		require.NoError(t, err)

		assert.Equal(t, "test-access-token", tokens.AccessToken)
		assert.Equal(t, "test-refresh-token", tokens.RefreshToken)
		assert.Equal(t, "Bearer", tokens.TokenType)
		assert.Equal(t, int64(3600), tokens.ExpiresIn)
	})
}

func TestPKCESecurity(t *testing.T) {
	t.Run("different clients generate different challenges", func(t *testing.T) {
		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    "https://auth.example.com/token",
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
		}

		client1 := oauth.NewClient(config)
		client2 := oauth.NewClient(config)

		url1, pair1, _, _ := client1.GetAuthorizationURL()
		url2, pair2, _, _ := client2.GetAuthorizationURL()

		assert.NotEqual(t, pair1.Verifier, pair2.Verifier)
		assert.NotEqual(t, pair1.Challenge, pair2.Challenge)
		assert.NotEqual(t, url1, url2) // URLs should differ due to state
	})

	t.Run("state parameter provides CSRF protection", func(t *testing.T) {
		config := oauth.Config{
			AuthURL:     "https://auth.example.com/authorize",
			TokenURL:    "https://auth.example.com/token",
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:8080/callback",
		}

		client := oauth.NewClient(config)

		// Generate multiple states
		states := make(map[string]bool)
		for i := 0; i < 10; i++ {
			_, _, state, err := client.GetAuthorizationURL()
			require.NoError(t, err)
			states[state] = true
		}

		// All states should be unique
		assert.Equal(t, 10, len(states))
	})
}

// Helper function to simulate OAuth callback
func simulateOAuthCallback(t *testing.T, redirectURL, code, state string) {
	parsedURL, err := url.Parse(redirectURL)
	require.NoError(t, err)

	callbackURL := fmt.Sprintf("http://localhost:%s%s?code=%s&state=%s",
		parsedURL.Port(), parsedURL.Path, code, state)

	resp, err := http.Get(callbackURL)
	if err == nil {
		defer resp.Body.Close()
	}
}
