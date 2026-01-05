package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	t.Run("creates validator with key fetcher", func(t *testing.T) {
		fetcher := func() (string, error) {
			return "test-key", nil
		}

		validator := jwt.NewValidator(fetcher)
		assert.NotNil(t, validator)
	})
}

func TestValidator_ValidateAccessToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	publicKeyPEM, err := jwt.FormatPublicKeyPEM(publicKey)
	require.NoError(t, err)

	t.Run("successful validation with cached key", func(t *testing.T) {
		fetcher := func() (string, error) {
			return publicKeyPEM, nil
		}

		validator := jwt.NewValidator(fetcher)

		// Create a valid token
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		// Validate token
		claims, err := validator.ValidateAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user-123", claims.UserID)
		assert.Equal(t, "test@example.com", claims.Email)

		// Verify key is cached
		cacheInfo := validator.GetCacheInfo()
		assert.True(t, cacheInfo.HasKey)
		assert.True(t, cacheInfo.IsValid)
	})

	t.Run("invalidates cache on validation failure", func(t *testing.T) {
		fetchCount := 0
		fetcher := func() (string, error) {
			fetchCount++
			return publicKeyPEM, nil
		}

		validator := jwt.NewValidator(fetcher)

		// First validation should succeed
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		_, err = validator.ValidateAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, 1, fetchCount)

		// Try validating invalid token
		_, err = validator.ValidateAccessToken("invalid-token")
		assert.Error(t, err)

		// Cache should be invalidated, next validation should fetch again
		token2, err := jwt.CreateAccessToken("user-456", "test2@example.com", []string{"admin"}, privateKey)
		require.NoError(t, err)

		_, err = validator.ValidateAccessToken(token2)
		require.NoError(t, err)
		assert.Equal(t, 2, fetchCount) // Fetched again after invalidation
	})

	t.Run("handles key fetch error", func(t *testing.T) {
		fetcher := func() (string, error) {
			return "", errors.New("fetch failed")
		}

		validator := jwt.NewValidator(fetcher)

		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		_, err = validator.ValidateAccessToken(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get public key")
	})
}

func TestValidator_ValidateRefreshToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	publicKeyPEM, err := jwt.FormatPublicKeyPEM(publicKey)
	require.NoError(t, err)

	t.Run("successful validation", func(t *testing.T) {
		fetcher := func() (string, error) {
			return publicKeyPEM, nil
		}

		validator := jwt.NewValidator(fetcher)

		token, err := jwt.CreateRefreshToken("user-123", "session-456", privateKey)
		require.NoError(t, err)

		claims, err := validator.ValidateRefreshToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user-123", claims.UserID)
		assert.Equal(t, "session-456", claims.SessionID)
	})

	t.Run("invalidates cache on validation failure", func(t *testing.T) {
		fetchCount := 0
		fetcher := func() (string, error) {
			fetchCount++
			return publicKeyPEM, nil
		}

		validator := jwt.NewValidator(fetcher)

		// Validate invalid token
		_, err := validator.ValidateRefreshToken("invalid-token")
		assert.Error(t, err)

		// Next validation should fetch new key
		token, err := jwt.CreateRefreshToken("user-123", "session-456", privateKey)
		require.NoError(t, err)

		_, err = validator.ValidateRefreshToken(token)
		require.NoError(t, err)
		assert.Equal(t, 2, fetchCount)
	})
}

func TestValidator_ValidateToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	publicKeyPEM, err := jwt.FormatPublicKeyPEM(publicKey)
	require.NoError(t, err)

	t.Run("returns validation result", func(t *testing.T) {
		fetcher := func() (string, error) {
			return publicKeyPEM, nil
		}

		validator := jwt.NewValidator(fetcher)

		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		result, err := validator.ValidateToken(token)
		require.NoError(t, err)
		assert.True(t, result.Valid)
		assert.Nil(t, result.Error)
		assert.NotNil(t, result.Claims)
	})

	t.Run("returns error in result for invalid token", func(t *testing.T) {
		fetcher := func() (string, error) {
			return publicKeyPEM, nil
		}

		validator := jwt.NewValidator(fetcher)

		result, err := validator.ValidateToken("invalid-token")
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.NotNil(t, result.Error)
	})
}

func TestValidator_PublicKeyCache(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	publicKeyPEM, err := jwt.FormatPublicKeyPEM(publicKey)
	require.NoError(t, err)

	t.Run("caches key with TTL", func(t *testing.T) {
		fetchCount := 0
		fetcher := func() (string, error) {
			fetchCount++
			return publicKeyPEM, nil
		}

		validator := jwt.NewValidator(fetcher)

		// First validation fetches key
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		_, err = validator.ValidateAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, 1, fetchCount)

		// Second validation uses cached key
		_, err = validator.ValidateAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, 1, fetchCount) // Still 1, used cache

		// Verify cache info
		cacheInfo := validator.GetCacheInfo()
		assert.True(t, cacheInfo.HasKey)
		assert.True(t, cacheInfo.IsValid)
		assert.Equal(t, jwt.PublicKeyCacheTTL, cacheInfo.TTL)
		assert.False(t, cacheInfo.CachedAt.IsZero())
		assert.False(t, cacheInfo.ExpiresAt.IsZero())
	})

	t.Run("invalidate cache manually", func(t *testing.T) {
		fetchCount := 0
		fetcher := func() (string, error) {
			fetchCount++
			return publicKeyPEM, nil
		}

		validator := jwt.NewValidator(fetcher)

		// Fetch and cache key
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		_, err = validator.ValidateAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, 1, fetchCount)

		// Invalidate cache
		err = validator.InvalidateCache()
		require.NoError(t, err)

		// Verify cache is empty
		cacheInfo := validator.GetCacheInfo()
		assert.False(t, cacheInfo.HasKey)
		assert.False(t, cacheInfo.IsValid)

		// Next validation fetches new key
		_, err = validator.ValidateAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, 2, fetchCount)
	})

	t.Run("custom invalidation function", func(t *testing.T) {
		invalidateCalled := false
		fetcher := func() (string, error) {
			return publicKeyPEM, nil
		}

		validator := jwt.NewValidator(fetcher)
		validator.SetInvalidateFunc(func() error {
			invalidateCalled = true
			return nil
		})

		err := validator.InvalidateCache()
		require.NoError(t, err)
		assert.True(t, invalidateCalled)
	})

	t.Run("custom invalidation function error", func(t *testing.T) {
		fetcher := func() (string, error) {
			return publicKeyPEM, nil
		}

		validator := jwt.NewValidator(fetcher)
		validator.SetInvalidateFunc(func() error {
			return errors.New("invalidation failed")
		})

		err := validator.InvalidateCache()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalidation failed")
	})
}

func TestValidator_NoKeyFetcher(t *testing.T) {
	t.Run("returns error when no key fetcher", func(t *testing.T) {
		validator := jwt.NewValidator(nil)

		_, err := validator.ValidateAccessToken("some-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no key fetcher configured")
	})
}

func TestParsePublicKeyPEM(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicKey := &privateKey.PublicKey

	t.Run("parse PKIX format", func(t *testing.T) {
		pemData, err := jwt.FormatPublicKeyPEM(publicKey)
		require.NoError(t, err)

		// This will be tested indirectly through validator tests
		assert.NotEmpty(t, pemData)
		assert.Contains(t, pemData, "BEGIN PUBLIC KEY")
		assert.Contains(t, pemData, "END PUBLIC KEY")
	})

	t.Run("invalid PEM data", func(t *testing.T) {
		validator := jwt.NewValidator(func() (string, error) {
			return "not-a-pem", nil
		})

		_, err := validator.ValidateAccessToken("token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get public key")
	})

	t.Run("invalid PEM type", func(t *testing.T) {
		validator := jwt.NewValidator(func() (string, error) {
			return "-----BEGIN CERTIFICATE-----\ndata\n-----END CERTIFICATE-----", nil
		})

		_, err := validator.ValidateAccessToken("token")
		assert.Error(t, err)
	})
}

func TestValidator_ConcurrentAccess(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	publicKeyPEM, err := jwt.FormatPublicKeyPEM(publicKey)
	require.NoError(t, err)

	t.Run("concurrent validations", func(t *testing.T) {
		fetcher := func() (string, error) {
			time.Sleep(10 * time.Millisecond) // Simulate network delay
			return publicKeyPEM, nil
		}

		validator := jwt.NewValidator(fetcher)

		// Create multiple tokens
		tokens := make([]string, 10)
		for i := 0; i < 10; i++ {
			token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
			require.NoError(t, err)
			tokens[i] = token
		}

		// Validate concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(token string) {
				_, err := validator.ValidateAccessToken(token)
				assert.NoError(t, err)
				done <- true
			}(tokens[i])
		}

		// Wait for all to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
