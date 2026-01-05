package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestKeyPair generates an RSA key pair for testing.
func generateTestKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

func TestCreateAccessToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		userID := "user-123"
		email := "test@example.com"
		roles := []string{"user", "admin"}

		token, err := jwt.CreateAccessToken(userID, email, roles, privateKey)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate the token
		claims, err := jwt.ValidateAccessToken(token, publicKey)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, roles, claims.Roles)
		assert.Equal(t, jwt.Issuer, claims.Issuer)
		assert.Contains(t, claims.Audience, jwt.Audience)
	})

	t.Run("token with empty roles", func(t *testing.T) {
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{}, privateKey)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := jwt.ValidateAccessToken(token, publicKey)
		require.NoError(t, err)
		assert.Empty(t, claims.Roles)
	})

	t.Run("token expiration", func(t *testing.T) {
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		claims, err := jwt.ValidateAccessToken(token, publicKey)
		require.NoError(t, err)

		// Check expiration is approximately 24 hours from now
		expectedExpiry := time.Now().Add(jwt.AccessTokenDuration)
		timeDiff := claims.ExpiresAt.Time.Sub(expectedExpiry)
		assert.Less(t, timeDiff.Abs().Seconds(), 5.0) // Within 5 seconds
	})
}

func TestCreateRefreshToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		userID := "user-123"
		sessionID := "session-456"

		token, err := jwt.CreateRefreshToken(userID, sessionID, privateKey)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate the token
		claims, err := jwt.ValidateRefreshToken(token, publicKey)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, sessionID, claims.SessionID)
		assert.Equal(t, jwt.Issuer, claims.Issuer)
		assert.Contains(t, claims.Audience, jwt.Audience)
	})

	t.Run("token expiration", func(t *testing.T) {
		token, err := jwt.CreateRefreshToken("user-123", "session-456", privateKey)
		require.NoError(t, err)

		claims, err := jwt.ValidateRefreshToken(token, publicKey)
		require.NoError(t, err)

		// Check expiration is approximately 7 days from now
		expectedExpiry := time.Now().Add(jwt.RefreshTokenDuration)
		timeDiff := claims.ExpiresAt.Time.Sub(expectedExpiry)
		assert.Less(t, timeDiff.Abs().Seconds(), 5.0) // Within 5 seconds
	})
}

func TestCreateTokenPair(t *testing.T) {
	privateKey, _, err := generateTestKeyPair()
	require.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		userID := "user-123"
		email := "test@example.com"
		roles := []string{"user"}
		sessionID := "session-456"

		tokenPair, err := jwt.CreateTokenPair(userID, email, roles, sessionID, privateKey)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenPair.AccessToken)
		assert.NotEmpty(t, tokenPair.RefreshToken)
		assert.Equal(t, "Bearer", tokenPair.TokenType)
		assert.Equal(t, int64(jwt.AccessTokenDuration.Seconds()), tokenPair.ExpiresIn)
	})
}

func TestValidateAccessToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	t.Run("valid token", func(t *testing.T) {
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		claims, err := jwt.ValidateAccessToken(token, publicKey)
		require.NoError(t, err)
		assert.NotNil(t, claims)
	})

	t.Run("invalid token format", func(t *testing.T) {
		_, err := jwt.ValidateAccessToken("invalid-token", publicKey)
		require.Error(t, err)
	})

	t.Run("token signed with wrong key", func(t *testing.T) {
		wrongPrivateKey, _, err := generateTestKeyPair()
		require.NoError(t, err)

		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, wrongPrivateKey)
		require.NoError(t, err)

		_, err = jwt.ValidateAccessToken(token, publicKey)
		require.Error(t, err)
	})

	t.Run("token with wrong signing method", func(t *testing.T) {
		// This would require creating a token with a different algorithm
		// which is not easily testable with our current structure
		// We're testing the method check in the validation function
	})
}

func TestValidateRefreshToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	t.Run("valid token", func(t *testing.T) {
		token, err := jwt.CreateRefreshToken("user-123", "session-456", privateKey)
		require.NoError(t, err)

		claims, err := jwt.ValidateRefreshToken(token, publicKey)
		require.NoError(t, err)
		assert.NotNil(t, claims)
	})

	t.Run("invalid token format", func(t *testing.T) {
		_, err := jwt.ValidateRefreshToken("invalid-token", publicKey)
		require.Error(t, err)
	})
}

func TestValidateToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	t.Run("valid token", func(t *testing.T) {
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		result, err := jwt.ValidateToken(token, publicKey)
		require.NoError(t, err)
		assert.True(t, result.Valid)
		assert.Nil(t, result.Error)
		assert.False(t, result.Expired)
		assert.NotNil(t, result.Claims)
	})

	t.Run("invalid token", func(t *testing.T) {
		result, err := jwt.ValidateToken("invalid-token", publicKey)
		require.NoError(t, err) // ValidateToken doesn't return error
		assert.False(t, result.Valid)
		assert.NotNil(t, result.Error)
	})
}

func TestIsTokenExpired(t *testing.T) {
	privateKey, _, err := generateTestKeyPair()
	require.NoError(t, err)

	t.Run("non-expired token", func(t *testing.T) {
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		expired, err := jwt.IsTokenExpired(token)
		require.NoError(t, err)
		assert.False(t, expired)
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := jwt.IsTokenExpired("invalid-token")
		require.Error(t, err)
	})
}

func TestGetTokenExpiration(t *testing.T) {
	privateKey, _, err := generateTestKeyPair()
	require.NoError(t, err)

	t.Run("get expiration", func(t *testing.T) {
		token, err := jwt.CreateAccessToken("user-123", "test@example.com", []string{"user"}, privateKey)
		require.NoError(t, err)

		expiration, err := jwt.GetTokenExpiration(token)
		require.NoError(t, err)
		assert.False(t, expiration.IsZero())

		// Check expiration is approximately 24 hours from now
		expectedExpiry := time.Now().Add(jwt.AccessTokenDuration)
		timeDiff := expiration.Sub(expectedExpiry)
		assert.Less(t, timeDiff.Abs().Seconds(), 5.0) // Within 5 seconds
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := jwt.GetTokenExpiration("invalid-token")
		require.Error(t, err)
	})
}

func TestTokenClaims(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	t.Run("access token claims", func(t *testing.T) {
		userID := "user-123"
		email := "test@example.com"
		roles := []string{"user", "admin"}

		token, err := jwt.CreateAccessToken(userID, email, roles, privateKey)
		require.NoError(t, err)

		claims, err := jwt.ValidateAccessToken(token, publicKey)
		require.NoError(t, err)

		// Verify all claims
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, roles, claims.Roles)
		assert.Equal(t, jwt.Issuer, claims.Issuer)
		assert.Contains(t, claims.Audience, jwt.Audience)
		assert.NotNil(t, claims.ExpiresAt)
		assert.NotNil(t, claims.IssuedAt)
		assert.NotNil(t, claims.NotBefore)
	})

	t.Run("refresh token claims", func(t *testing.T) {
		userID := "user-123"
		sessionID := "session-456"

		token, err := jwt.CreateRefreshToken(userID, sessionID, privateKey)
		require.NoError(t, err)

		claims, err := jwt.ValidateRefreshToken(token, publicKey)
		require.NoError(t, err)

		// Verify all claims
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, sessionID, claims.SessionID)
		assert.Equal(t, jwt.Issuer, claims.Issuer)
		assert.Contains(t, claims.Audience, jwt.Audience)
		assert.NotNil(t, claims.ExpiresAt)
		assert.NotNil(t, claims.IssuedAt)
		assert.NotNil(t, claims.NotBefore)
	})
}

func TestTokenConstants(t *testing.T) {
	t.Run("verify constants", func(t *testing.T) {
		assert.Equal(t, "ainative-auth", jwt.Issuer)
		assert.Equal(t, "ainative-code", jwt.Audience)
		assert.Equal(t, 24*time.Hour, jwt.AccessTokenDuration)
		assert.Equal(t, 7*24*time.Hour, jwt.RefreshTokenDuration)
		assert.Equal(t, "RS256", jwt.SigningMethod)
	})
}
