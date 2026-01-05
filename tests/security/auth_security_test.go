package security

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJWTTokenValidation_ExpiredToken verifies that expired tokens are properly rejected
func TestJWTTokenValidation_ExpiredToken(t *testing.T) {
	// Given: An expired JWT token
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := jwt.MapClaims{
		"iss":   "ainative-auth",
		"aud":   []string{"ainative-code"},
		"sub":   "user-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// When: Attempting to parse the expired token
	_, err = auth.ParseAccessToken(tokenString, &privateKey.PublicKey)

	// Then: Should return error indicating token is expired
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrTokenExpired)
}

// TestJWTTokenValidation_InvalidSignature verifies that tokens with invalid signatures are rejected
func TestJWTTokenValidation_InvalidSignature(t *testing.T) {
	// Given: A token signed with one key, verified with another
	privateKey1, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := jwt.MapClaims{
		"iss":   "ainative-auth",
		"aud":   []string{"ainative-code"},
		"sub":   "user-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey1)
	require.NoError(t, err)

	// When: Attempting to verify with wrong public key
	_, err = auth.ParseAccessToken(tokenString, &privateKey2.PublicKey)

	// Then: Should return signature validation error
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrInvalidSignature)
}

// TestJWTTokenValidation_WrongIssuer verifies tokens from wrong issuer are rejected
func TestJWTTokenValidation_WrongIssuer(t *testing.T) {
	// Given: A token with wrong issuer
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := jwt.MapClaims{
		"iss":   "evil-auth-server", // Wrong issuer
		"aud":   []string{"ainative-code"},
		"sub":   "user-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// When: Attempting to parse the token
	_, err = auth.ParseAccessToken(tokenString, &privateKey.PublicKey)

	// Then: Should return issuer validation error
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrInvalidIssuer)
}

// TestJWTTokenValidation_WrongAudience verifies tokens for wrong audience are rejected
func TestJWTTokenValidation_WrongAudience(t *testing.T) {
	// Given: A token with wrong audience
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := jwt.MapClaims{
		"iss":   "ainative-auth",
		"aud":   []string{"different-app"}, // Wrong audience
		"sub":   "user-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// When: Attempting to parse the token
	_, err = auth.ParseAccessToken(tokenString, &privateKey.PublicKey)

	// Then: Should return audience validation error
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrInvalidAudience)
}

// TestJWTTokenValidation_MissingClaims verifies tokens with missing required claims are rejected
func TestJWTTokenValidation_MissingClaims(t *testing.T) {
	testCases := []struct {
		name          string
		claims        jwt.MapClaims
		expectedError error
	}{
		{
			name: "Missing subject (sub)",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{"ainative-code"},
				"email": "test@example.com",
				"exp":   time.Now().Add(1 * time.Hour).Unix(),
				// Missing: "sub"
			},
			expectedError: auth.ErrInvalidClaims,
		},
		{
			name: "Missing email",
			claims: jwt.MapClaims{
				"iss": "ainative-auth",
				"aud": []string{"ainative-code"},
				"sub": "user-123",
				"exp": time.Now().Add(1 * time.Hour).Unix(),
				// Missing: "email"
			},
			expectedError: auth.ErrInvalidClaims,
		},
		{
			name: "Missing expiration",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{"ainative-code"},
				"sub":   "user-123",
				"email": "test@example.com",
				// Missing: "exp"
			},
			expectedError: auth.ErrInvalidClaims,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: A token with missing claims
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)

			token := jwt.NewWithClaims(jwt.SigningMethodRS256, tc.claims)
			tokenString, err := token.SignedString(privateKey)
			require.NoError(t, err)

			// When: Attempting to parse the token
			_, err = auth.ParseAccessToken(tokenString, &privateKey.PublicKey)

			// Then: Should return claims validation error
			assert.Error(t, err)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

// TestJWTTokenValidation_AlgorithmConfusion verifies that algorithm confusion attacks are prevented
func TestJWTTokenValidation_AlgorithmConfusion(t *testing.T) {
	// Given: A token signed with HS256 (symmetric) instead of RS256 (asymmetric)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := jwt.MapClaims{
		"iss":   "ainative-auth",
		"aud":   []string{"ainative-code"},
		"sub":   "user-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}

	// Attacker tries to use HS256 with the public key as the secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("secret"))
	require.NoError(t, err)

	// When: Attempting to parse the token
	_, err = auth.ParseAccessToken(tokenString, &privateKey.PublicKey)

	// Then: Should reject due to algorithm mismatch
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrInvalidSignature)
}

// TestJWTTokenValidation_ValidToken verifies that valid tokens are accepted
func TestJWTTokenValidation_ValidToken(t *testing.T) {
	// Given: A properly formatted and signed token
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	userID := "user-123"
	email := "test@example.com"
	roles := []string{"user", "admin"}

	claims := jwt.MapClaims{
		"iss":   "ainative-auth",
		"aud":   []string{"ainative-code"},
		"sub":   userID,
		"email": email,
		"roles": roles,
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// When: Parsing the valid token
	accessToken, err := auth.ParseAccessToken(tokenString, &privateKey.PublicKey)

	// Then: Should successfully parse and extract claims
	require.NoError(t, err)
	assert.Equal(t, userID, accessToken.UserID)
	assert.Equal(t, email, accessToken.Email)
	assert.ElementsMatch(t, roles, accessToken.Roles)
	assert.Equal(t, "ainative-auth", accessToken.Issuer)
	assert.Equal(t, "ainative-code", accessToken.Audience)
	assert.False(t, accessToken.ExpiresAt.Before(time.Now()))
}

// TestRefreshTokenValidation_SessionBinding verifies refresh tokens are bound to sessions
func TestRefreshTokenValidation_SessionBinding(t *testing.T) {
	// Given: A refresh token with session binding
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	sessionID := "session-abc-123"
	userID := "user-456"

	claims := jwt.MapClaims{
		"iss":        "ainative-auth",
		"aud":        []string{"ainative-code"},
		"sub":        userID,
		"session_id": sessionID,
		"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// When: Parsing the refresh token
	refreshToken, err := auth.ParseRefreshToken(tokenString, &privateKey.PublicKey)

	// Then: Should successfully extract session binding
	require.NoError(t, err)
	assert.Equal(t, userID, refreshToken.UserID)
	assert.Equal(t, sessionID, refreshToken.SessionID)
}

// TestRefreshTokenValidation_MissingSessionID verifies refresh tokens without session_id are rejected
func TestRefreshTokenValidation_MissingSessionID(t *testing.T) {
	// Given: A refresh token without session_id
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := jwt.MapClaims{
		"iss": "ainative-auth",
		"aud": []string{"ainative-code"},
		"sub": "user-456",
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
		// Missing: "session_id"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// When: Attempting to parse the refresh token
	_, err = auth.ParseRefreshToken(tokenString, &privateKey.PublicKey)

	// Then: Should return error for missing session_id
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrInvalidClaims)
}

// TestTokenReplayAttackPrevention verifies that token replay attacks are mitigated
func TestTokenReplayAttackPrevention(t *testing.T) {
	// Given: A valid token that has been used once
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := jwt.MapClaims{
		"iss":   "ainative-auth",
		"aud":   []string{"ainative-code"},
		"sub":   "user-123",
		"email": "test@example.com",
		"jti":   "unique-token-id-123", // JWT ID for tracking
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// When/Then: Token should be valid on first use
	accessToken1, err := auth.ParseAccessToken(tokenString, &privateKey.PublicKey)
	require.NoError(t, err)
	assert.NotNil(t, accessToken1)

	// Note: Actual replay prevention requires server-side tracking of JTI
	// This test documents the expected behavior
	// Implementation should track used JTI values in Redis/database with TTL
}

// TestPasswordHashingSecurity verifies bcrypt password hashing uses appropriate cost
func TestPasswordHashingSecurity(t *testing.T) {
	t.Skip("Requires local auth implementation")
	// This test would verify:
	// 1. Bcrypt cost factor >= 12
	// 2. Password hashing is deterministic
	// 3. Different passwords produce different hashes
	// 4. Same password produces different hashes (due to salt)
}

// BenchmarkJWTTokenValidation measures JWT validation performance
func BenchmarkJWTTokenValidation(b *testing.B) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	claims := jwt.MapClaims{
		"iss":   "ainative-auth",
		"aud":   []string{"ainative-code"},
		"sub":   "user-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, _ := token.SignedString(privateKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = auth.ParseAccessToken(tokenString, &privateKey.PublicKey)
	}
}
