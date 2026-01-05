package keychain_test

import (
	"testing"

	"github.com/99designs/keyring"
	"github.com/AINative-studio/ainative-code/internal/auth/jwt"
	"github.com/AINative-studio/ainative-code/internal/auth/keychain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	t.Run("returns keychain instance", func(t *testing.T) {
		kc := keychain.Get()
		assert.NotNil(t, kc)
	})

	t.Run("returns same instance on multiple calls", func(t *testing.T) {
		kc1 := keychain.Get()
		kc2 := keychain.Get()
		assert.Equal(t, kc1, kc2)
	})
}

func TestNew(t *testing.T) {
	t.Run("creates keychain with custom config", func(t *testing.T) {
		config := keyring.Config{
			ServiceName:     "test-service",
			AllowedBackends: []keyring.BackendType{keyring.FileBackend},
			FileDir:         t.TempDir(),
			FilePasswordFunc: func(prompt string) (string, error) {
				return "test-password", nil
			},
		}

		kc, err := keychain.New(config)
		require.NoError(t, err)
		assert.NotNil(t, kc)
	})
}

func getTestKeychain(t *testing.T) keychain.Keychain {
	config := keyring.Config{
		ServiceName:     "test-ainative-code",
		AllowedBackends: []keyring.BackendType{keyring.FileBackend},
		FileDir:         t.TempDir(),
		FilePasswordFunc: func(prompt string) (string, error) {
			return "test-password", nil
		},
	}

	kc, err := keychain.New(config)
	require.NoError(t, err)
	return kc
}

func TestAccessToken(t *testing.T) {
	kc := getTestKeychain(t)

	t.Run("stores and retrieves access token", func(t *testing.T) {
		token := "test-access-token-12345"

		err := kc.SetAccessToken(token)
		require.NoError(t, err)

		retrieved, err := kc.GetAccessToken()
		require.NoError(t, err)
		assert.Equal(t, token, retrieved)
	})

	t.Run("updates existing access token", func(t *testing.T) {
		err := kc.SetAccessToken("first-token")
		require.NoError(t, err)

		err = kc.SetAccessToken("second-token")
		require.NoError(t, err)

		retrieved, err := kc.GetAccessToken()
		require.NoError(t, err)
		assert.Equal(t, "second-token", retrieved)
	})

	t.Run("returns error when token not found", func(t *testing.T) {
		freshKc := getTestKeychain(t)

		_, err := freshKc.GetAccessToken()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestRefreshToken(t *testing.T) {
	kc := getTestKeychain(t)

	t.Run("stores and retrieves refresh token", func(t *testing.T) {
		token := "test-refresh-token-67890"

		err := kc.SetRefreshToken(token)
		require.NoError(t, err)

		retrieved, err := kc.GetRefreshToken()
		require.NoError(t, err)
		assert.Equal(t, token, retrieved)
	})

	t.Run("returns error when token not found", func(t *testing.T) {
		freshKc := getTestKeychain(t)

		_, err := freshKc.GetRefreshToken()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestTokenPair(t *testing.T) {
	kc := getTestKeychain(t)

	t.Run("stores and retrieves token pair", func(t *testing.T) {
		tokens := &jwt.TokenPair{
			AccessToken:  "access-123",
			RefreshToken: "refresh-456",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}

		err := kc.SetTokenPair(tokens)
		require.NoError(t, err)

		retrieved, err := kc.GetTokenPair()
		require.NoError(t, err)
		assert.Equal(t, tokens.AccessToken, retrieved.AccessToken)
		assert.Equal(t, tokens.RefreshToken, retrieved.RefreshToken)
		assert.Equal(t, tokens.ExpiresIn, retrieved.ExpiresIn)
		assert.Equal(t, tokens.TokenType, retrieved.TokenType)
	})

	t.Run("also stores individual tokens", func(t *testing.T) {
		tokens := &jwt.TokenPair{
			AccessToken:  "pair-access-789",
			RefreshToken: "pair-refresh-012",
			ExpiresIn:    7200,
			TokenType:    "Bearer",
		}

		err := kc.SetTokenPair(tokens)
		require.NoError(t, err)

		// Should be able to retrieve individually
		accessToken, err := kc.GetAccessToken()
		require.NoError(t, err)
		assert.Equal(t, tokens.AccessToken, accessToken)

		refreshToken, err := kc.GetRefreshToken()
		require.NoError(t, err)
		assert.Equal(t, tokens.RefreshToken, refreshToken)
	})

	t.Run("rejects nil token pair", func(t *testing.T) {
		err := kc.SetTokenPair(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("returns error when pair not found", func(t *testing.T) {
		freshKc := getTestKeychain(t)

		_, err := freshKc.GetTokenPair()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestAPIKey(t *testing.T) {
	kc := getTestKeychain(t)

	t.Run("stores and retrieves API key", func(t *testing.T) {
		key := "api-key-secret-xyz"

		err := kc.SetAPIKey(key)
		require.NoError(t, err)

		retrieved, err := kc.GetAPIKey()
		require.NoError(t, err)
		assert.Equal(t, key, retrieved)
	})

	t.Run("returns error when key not found", func(t *testing.T) {
		freshKc := getTestKeychain(t)

		_, err := freshKc.GetAPIKey()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestUserEmail(t *testing.T) {
	kc := getTestKeychain(t)

	t.Run("stores and retrieves user email", func(t *testing.T) {
		email := "user@example.com"

		err := kc.SetUserEmail(email)
		require.NoError(t, err)

		retrieved, err := kc.GetUserEmail()
		require.NoError(t, err)
		assert.Equal(t, email, retrieved)
	})

	t.Run("returns error when email not found", func(t *testing.T) {
		freshKc := getTestKeychain(t)

		_, err := freshKc.GetUserEmail()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestDelete(t *testing.T) {
	kc := getTestKeychain(t)

	t.Run("deletes specific key", func(t *testing.T) {
		// Store multiple items
		kc.SetAccessToken("access-token")
		kc.SetRefreshToken("refresh-token")
		kc.SetAPIKey("api-key")

		// Delete access token
		err := kc.Delete(keychain.AccessTokenKey)
		require.NoError(t, err)

		// Access token should be gone
		_, err = kc.GetAccessToken()
		assert.Error(t, err)

		// Others should still exist
		_, err = kc.GetRefreshToken()
		assert.NoError(t, err)

		_, err = kc.GetAPIKey()
		assert.NoError(t, err)
	})

	t.Run("delete is idempotent", func(t *testing.T) {
		err := kc.Delete("non-existent-key")
		assert.NoError(t, err) // Should not error
	})
}

func TestDeleteAll(t *testing.T) {
	kc := getTestKeychain(t)

	t.Run("deletes all credentials", func(t *testing.T) {
		// Store all types
		kc.SetAccessToken("access")
		kc.SetRefreshToken("refresh")
		kc.SetAPIKey("api-key")
		kc.SetUserEmail("user@test.com")
		kc.SetTokenPair(&jwt.TokenPair{
			AccessToken:  "pair-access",
			RefreshToken: "pair-refresh",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		})

		// Delete all
		err := kc.DeleteAll()
		require.NoError(t, err)

		// All should be gone
		_, err = kc.GetAccessToken()
		assert.Error(t, err)

		_, err = kc.GetRefreshToken()
		assert.Error(t, err)

		_, err = kc.GetAPIKey()
		assert.Error(t, err)

		_, err = kc.GetUserEmail()
		assert.Error(t, err)

		_, err = kc.GetTokenPair()
		assert.Error(t, err)
	})
}

func TestExists(t *testing.T) {
	kc := getTestKeychain(t)

	t.Run("returns true when key exists", func(t *testing.T) {
		kc.SetAccessToken("test-token")

		exists := kc.Exists(keychain.AccessTokenKey)
		assert.True(t, exists)
	})

	t.Run("returns false when key does not exist", func(t *testing.T) {
		exists := kc.Exists("non-existent-key")
		assert.False(t, exists)
	})
}

func TestCompleteWorkflow(t *testing.T) {
	kc := getTestKeychain(t)

	t.Run("complete authentication workflow", func(t *testing.T) {
		// 1. User authenticates, receive tokens
		tokens := &jwt.TokenPair{
			AccessToken:  "workflow-access-token",
			RefreshToken: "workflow-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}

		// 2. Store tokens
		err := kc.SetTokenPair(tokens)
		require.NoError(t, err)

		// 3. Store user email
		err = kc.SetUserEmail("workflow@example.com")
		require.NoError(t, err)

		// 4. Store API key
		err = kc.SetAPIKey("workflow-api-key")
		require.NoError(t, err)

		// 5. Retrieve everything
		retrievedTokens, err := kc.GetTokenPair()
		require.NoError(t, err)
		assert.Equal(t, tokens.AccessToken, retrievedTokens.AccessToken)

		retrievedEmail, err := kc.GetUserEmail()
		require.NoError(t, err)
		assert.Equal(t, "workflow@example.com", retrievedEmail)

		retrievedKey, err := kc.GetAPIKey()
		require.NoError(t, err)
		assert.Equal(t, "workflow-api-key", retrievedKey)

		// 6. User logs out, delete all
		err = kc.DeleteAll()
		require.NoError(t, err)

		// 7. Verify all deleted
		_, err = kc.GetTokenPair()
		assert.Error(t, err)

		_, err = kc.GetUserEmail()
		assert.Error(t, err)

		_, err = kc.GetAPIKey()
		assert.Error(t, err)
	})
}
