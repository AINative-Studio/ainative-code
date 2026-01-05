package jwt_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIValidator(t *testing.T) {
	t.Run("creates API validator", func(t *testing.T) {
		validator := jwt.NewAPIValidator("https://api.example.com", nil)
		assert.NotNil(t, validator)
		assert.NotNil(t, validator.GetValidator())
	})

	t.Run("accepts custom HTTP client", func(t *testing.T) {
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		validator := jwt.NewAPIValidator("https://api.example.com", client)
		assert.NotNil(t, validator)
	})
}

func TestAPIValidator_ValidateAccessToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	publicKeyPEM, err := jwt.FormatPublicKeyPEM(publicKey)
	require.NoError(t, err)

	t.Run("successful API validation", func(t *testing.T) {
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		// Mock API server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/validate" {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(jwt.APIValidationResponse{
					Valid:     true,
					Expired:   false,
					PublicKey: publicKeyPEM,
				})
			} else if r.URL.Path == "/api/auth/public-key" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"public_key": publicKeyPEM,
				})
			}
		}))
		defer server.Close()

		validator := jwt.NewAPIValidator(server.URL, nil)

		claims, err := validator.ValidateAccessToken(context.Background(), token)
		require.NoError(t, err)
		assert.Equal(t, "user-123", claims.UserID)
		assert.Equal(t, "test@example.com", claims.Email)
	})

	t.Run("API validation failure", func(t *testing.T) {
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		// Mock API server returning invalid response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/validate" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(jwt.APIValidationResponse{
					Valid:   false,
					Message: "token is invalid",
				})
			}
		}))
		defer server.Close()

		validator := jwt.NewAPIValidator(server.URL, nil)

		_, err = validator.ValidateAccessToken(context.Background(), token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is invalid")
	})

	t.Run("fallback to local validation on API error", func(t *testing.T) {
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		// Mock API server that fails
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/validate" {
				w.WriteHeader(http.StatusInternalServerError)
			} else if r.URL.Path == "/api/auth/public-key" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"public_key": publicKeyPEM,
				})
			}
		}))
		defer server.Close()

		validator := jwt.NewAPIValidator(server.URL, nil)

		// Should fall back to local validation
		claims, err := validator.ValidateAccessToken(context.Background(), token)
		require.NoError(t, err)
		assert.Equal(t, "user-123", claims.UserID)
	})

	t.Run("timeout handling", func(t *testing.T) {
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		// Mock API server with delay
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/validate" {
				time.Sleep(100 * time.Millisecond)
			}
		}))
		defer server.Close()

		// Create client with short timeout
		client := &http.Client{
			Timeout: 10 * time.Millisecond,
		}

		validator := jwt.NewAPIValidator(server.URL, client)

		// Should timeout and fall back
		ctx := context.Background()
		_, err = validator.ValidateAccessToken(ctx, token)
		// Error expected since both API and local validation will fail
		// (local has no cached key yet)
		assert.Error(t, err)
	})

	t.Run("caches public key from API response", func(t *testing.T) {
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/validate" {
				callCount++
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(jwt.APIValidationResponse{
					Valid:     true,
					PublicKey: publicKeyPEM,
				})
			} else if r.URL.Path == "/api/auth/public-key" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"public_key": publicKeyPEM,
				})
			}
		}))
		defer server.Close()

		validator := jwt.NewAPIValidator(server.URL, nil)

		// First call
		_, err = validator.ValidateAccessToken(context.Background(), token)
		require.NoError(t, err)

		// Verify key is cached
		cacheInfo := validator.GetValidator().GetCacheInfo()
		assert.True(t, cacheInfo.HasKey)
		assert.True(t, cacheInfo.IsValid)
	})
}

func TestAPIValidator_ValidateRefreshToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	publicKeyPEM, err := jwt.FormatPublicKeyPEM(publicKey)
	require.NoError(t, err)

	t.Run("successful API validation", func(t *testing.T) {
		token, err := jwt.CreateRefreshToken("user-123", "session-456", privateKey)
		require.NoError(t, err)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/validate" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(jwt.APIValidationResponse{
					Valid:     true,
					PublicKey: publicKeyPEM,
				})
			} else if r.URL.Path == "/api/auth/public-key" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"public_key": publicKeyPEM,
				})
			}
		}))
		defer server.Close()

		validator := jwt.NewAPIValidator(server.URL, nil)

		claims, err := validator.ValidateRefreshToken(context.Background(), token)
		require.NoError(t, err)
		assert.Equal(t, "user-123", claims.UserID)
		assert.Equal(t, "session-456", claims.SessionID)
	})

	t.Run("fallback on API failure", func(t *testing.T) {
		token, err := jwt.CreateRefreshToken("user-123", "session-456", privateKey)
		require.NoError(t, err)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/validate" {
				w.WriteHeader(http.StatusServiceUnavailable)
			} else if r.URL.Path == "/api/auth/public-key" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"public_key": publicKeyPEM,
				})
			}
		}))
		defer server.Close()

		validator := jwt.NewAPIValidator(server.URL, nil)

		claims, err := validator.ValidateRefreshToken(context.Background(), token)
		require.NoError(t, err)
		assert.Equal(t, "user-123", claims.UserID)
	})
}

func TestAPIValidator_ValidateToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	publicKeyPEM, err := jwt.FormatPublicKeyPEM(publicKey)
	require.NoError(t, err)

	t.Run("successful validation with claims", func(t *testing.T) {
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		expiresAt := time.Now().Add(24 * time.Hour)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/validate" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(jwt.APIValidationResponse{
					Valid:     true,
					Expired:   false,
					ExpiresAt: &expiresAt,
					PublicKey: publicKeyPEM,
				})
			} else if r.URL.Path == "/api/auth/public-key" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"public_key": publicKeyPEM,
				})
			}
		}))
		defer server.Close()

		validator := jwt.NewAPIValidator(server.URL, nil)

		result, err := validator.ValidateToken(context.Background(), token)
		require.NoError(t, err)
		assert.True(t, result.Valid)
		assert.False(t, result.Expired)
		assert.NotNil(t, result.Claims)
	})

	t.Run("validation failure with error message", func(t *testing.T) {
		token := "invalid-token"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/validate" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(jwt.APIValidationResponse{
					Valid:   false,
					Message: "invalid token format",
				})
			}
		}))
		defer server.Close()

		validator := jwt.NewAPIValidator(server.URL, nil)

		result, err := validator.ValidateToken(context.Background(), token)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.NotNil(t, result.Error)
		assert.Contains(t, result.Error.Error(), "invalid token format")
	})
}

func TestAPIValidator_InvalidateCache(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	publicKeyPEM, err := jwt.FormatPublicKeyPEM(publicKey)
	require.NoError(t, err)

	t.Run("invalidates cache", func(t *testing.T) {
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/validate" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(jwt.APIValidationResponse{
					Valid:     true,
					PublicKey: publicKeyPEM,
				})
			} else if r.URL.Path == "/api/auth/public-key" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"public_key": publicKeyPEM,
				})
			}
		}))
		defer server.Close()

		validator := jwt.NewAPIValidator(server.URL, nil)

		// Validate to populate cache
		_, err = validator.ValidateAccessToken(context.Background(), token)
		require.NoError(t, err)

		// Verify cache is populated
		cacheInfo := validator.GetValidator().GetCacheInfo()
		assert.True(t, cacheInfo.HasKey)

		// Invalidate cache
		err = validator.InvalidateCache()
		require.NoError(t, err)

		// Verify cache is cleared
		cacheInfo = validator.GetValidator().GetCacheInfo()
		assert.False(t, cacheInfo.HasKey)
	})
}

func TestAPIValidator_NetworkErrors(t *testing.T) {
	t.Run("handles connection refused", func(t *testing.T) {
		// Use invalid URL that will cause connection error
		validator := jwt.NewAPIValidator("http://localhost:1", nil)

		_, err := validator.ValidateAccessToken(context.Background(), "token")
		assert.Error(t, err)
	})

	t.Run("handles malformed response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/validate" {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("invalid json"))
			}
		}))
		defer server.Close()

		validator := jwt.NewAPIValidator(server.URL, nil)

		_, err := validator.ValidateAccessToken(context.Background(), "token")
		assert.Error(t, err)
	})

	t.Run("handles empty public key", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/public-key" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"public_key": "",
				})
			}
		}))
		defer server.Close()

		validator := jwt.NewAPIValidator(server.URL, nil)

		// Try to validate which will trigger key fetch
		_, err := validator.ValidateAccessToken(context.Background(), "token")
		assert.Error(t, err)
	})
}

func TestAPIValidator_ContextCancellation(t *testing.T) {
	t.Run("respects context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
		}))
		defer server.Close()

		validator := jwt.NewAPIValidator(server.URL, nil)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := validator.ValidateAccessToken(ctx, "token")
		assert.Error(t, err)
	})
}
