package oauth_test

import (
	"strings"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/auth/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePKCECodePair(t *testing.T) {
	t.Run("generates valid code pair", func(t *testing.T) {
		pair, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)
		assert.NotEmpty(t, pair.Verifier)
		assert.NotEmpty(t, pair.Challenge)
		assert.Equal(t, "S256", pair.ChallengeMethod)
	})

	t.Run("verifier has correct length", func(t *testing.T) {
		pair, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)
		assert.Equal(t, oauth.PKCECodeVerifierLength, len(pair.Verifier))
	})

	t.Run("verifier contains only valid characters", func(t *testing.T) {
		pair, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
		for _, c := range pair.Verifier {
			assert.Contains(t, validChars, string(c))
		}
	})

	t.Run("generates different verifiers each time", func(t *testing.T) {
		pair1, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		pair2, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		assert.NotEqual(t, pair1.Verifier, pair2.Verifier)
		assert.NotEqual(t, pair1.Challenge, pair2.Challenge)
	})

	t.Run("challenge is base64url encoded", func(t *testing.T) {
		pair, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		// Base64URL should not contain + / = characters
		assert.NotContains(t, pair.Challenge, "+")
		assert.NotContains(t, pair.Challenge, "/")
		assert.NotContains(t, pair.Challenge, "=")
	})

	t.Run("challenge is consistent for same verifier", func(t *testing.T) {
		pair1, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		// Verify the challenge matches the verifier
		isValid := oauth.VerifyCodeChallenge(pair1.Verifier, pair1.Challenge)
		assert.True(t, isValid)
	})
}

func TestGeneratePKCECodePairWithLength(t *testing.T) {
	t.Run("generates verifier with custom length", func(t *testing.T) {
		length := 50
		pair, err := oauth.GeneratePKCECodePairWithLength(length)
		require.NoError(t, err)
		assert.Equal(t, length, len(pair.Verifier))
	})

	t.Run("accepts minimum length", func(t *testing.T) {
		pair, err := oauth.GeneratePKCECodePairWithLength(oauth.PKCECodeVerifierMinLength)
		require.NoError(t, err)
		assert.Equal(t, oauth.PKCECodeVerifierMinLength, len(pair.Verifier))
	})

	t.Run("accepts maximum length", func(t *testing.T) {
		pair, err := oauth.GeneratePKCECodePairWithLength(oauth.PKCECodeVerifierMaxLength)
		require.NoError(t, err)
		assert.Equal(t, oauth.PKCECodeVerifierMaxLength, len(pair.Verifier))
	})

	t.Run("rejects length too short", func(t *testing.T) {
		_, err := oauth.GeneratePKCECodePairWithLength(oauth.PKCECodeVerifierMinLength - 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be between")
	})

	t.Run("rejects length too long", func(t *testing.T) {
		_, err := oauth.GeneratePKCECodePairWithLength(oauth.PKCECodeVerifierMaxLength + 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be between")
	})
}

func TestValidateCodeVerifier(t *testing.T) {
	t.Run("validates correct verifier", func(t *testing.T) {
		pair, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		err = oauth.ValidateCodeVerifier(pair.Verifier)
		assert.NoError(t, err)
	})

	t.Run("rejects verifier too short", func(t *testing.T) {
		shortVerifier := strings.Repeat("a", oauth.PKCECodeVerifierMinLength-1)
		err := oauth.ValidateCodeVerifier(shortVerifier)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too short")
	})

	t.Run("rejects verifier too long", func(t *testing.T) {
		longVerifier := strings.Repeat("a", oauth.PKCECodeVerifierMaxLength+1)
		err := oauth.ValidateCodeVerifier(longVerifier)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too long")
	})

	t.Run("rejects invalid characters", func(t *testing.T) {
		invalidVerifier := strings.Repeat("a", oauth.PKCECodeVerifierMinLength) + "!"
		err := oauth.ValidateCodeVerifier(invalidVerifier)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character")
	})

	t.Run("accepts all valid characters", func(t *testing.T) {
		// Test all valid characters (66 chars total)
		validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
		// validChars is already longer than minimum (66 > 43), so just use it
		err := oauth.ValidateCodeVerifier(validChars)
		assert.NoError(t, err)
	})
}

func TestVerifyCodeChallenge(t *testing.T) {
	t.Run("verifies matching verifier and challenge", func(t *testing.T) {
		pair, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		isValid := oauth.VerifyCodeChallenge(pair.Verifier, pair.Challenge)
		assert.True(t, isValid)
	})

	t.Run("rejects mismatched verifier and challenge", func(t *testing.T) {
		pair1, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		pair2, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		// Use verifier from pair1 with challenge from pair2
		isValid := oauth.VerifyCodeChallenge(pair1.Verifier, pair2.Challenge)
		assert.False(t, isValid)
	})

	t.Run("rejects invalid challenge", func(t *testing.T) {
		pair, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		isValid := oauth.VerifyCodeChallenge(pair.Verifier, "invalid-challenge")
		assert.False(t, isValid)
	})
}

func TestPKCEConstants(t *testing.T) {
	t.Run("verify constants", func(t *testing.T) {
		assert.Equal(t, 64, oauth.PKCECodeVerifierLength)
		assert.Equal(t, 43, oauth.PKCECodeVerifierMinLength)
		assert.Equal(t, 128, oauth.PKCECodeVerifierMaxLength)
		assert.Equal(t, "S256", oauth.PKCEChallengeMethod)
	})
}

func TestPKCESecurityProperties(t *testing.T) {
	t.Run("high entropy in verifier", func(t *testing.T) {
		// Generate multiple verifiers and ensure they're unique
		verifiers := make(map[string]bool)
		iterations := 100

		for i := 0; i < iterations; i++ {
			pair, err := oauth.GeneratePKCECodePair()
			require.NoError(t, err)
			verifiers[pair.Verifier] = true
		}

		// All verifiers should be unique
		assert.Equal(t, iterations, len(verifiers))
	})

	t.Run("challenge is deterministic", func(t *testing.T) {
		pair, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		// Verify same verifier produces same challenge
		for i := 0; i < 10; i++ {
			isValid := oauth.VerifyCodeChallenge(pair.Verifier, pair.Challenge)
			assert.True(t, isValid)
		}
	})

	t.Run("challenge is not reversible", func(t *testing.T) {
		pair, err := oauth.GeneratePKCECodePair()
		require.NoError(t, err)

		// Challenge should not contain the verifier
		assert.NotContains(t, pair.Challenge, pair.Verifier)

		// Challenge should be different from verifier
		assert.NotEqual(t, pair.Challenge, pair.Verifier)
	})
}
