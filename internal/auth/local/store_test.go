package local

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestStore creates a temporary SQLite database for testing.
func setupTestStore(t *testing.T) (*Store, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "local-auth-test-*")
	require.NoError(t, err)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Create store
	store, err := NewStore(dbPath)
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

func TestNewStore(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	assert.NotNil(t, store)
	assert.NotNil(t, store.db)
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid registration",
			email:       "test@example.com",
			password:    "secure-password-123",
			expectError: false,
		},
		{
			name:        "empty email",
			email:       "",
			password:    "secure-password-123",
			expectError: true,
			errorMsg:    "email cannot be empty",
		},
		{
			name:        "empty password",
			email:       "test@example.com",
			password:    "",
			expectError: true,
			errorMsg:    "password cannot be empty",
		},
		{
			name:        "duplicate email",
			email:       "test@example.com",
			password:    "another-password",
			expectError: true,
			errorMsg:    "failed to create user",
		},
	}

	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Register first user for duplicate test
	err := store.Register("test@example.com", "secure-password-123")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip first test since we already registered
			if tt.name == "valid registration" {
				return
			}

			err := store.Register(tt.email, tt.password)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthenticate(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Register a test user
	email := "auth@example.com"
	password := "correct-password"
	err := store.Register(email, password)
	require.NoError(t, err)

	t.Run("valid credentials", func(t *testing.T) {
		session, err := store.Authenticate(email, password)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Greater(t, session.ID, int64(0))
		assert.Greater(t, session.UserID, int64(0))
		assert.NotEmpty(t, session.AccessToken)
		assert.NotEmpty(t, session.RefreshToken)
		assert.True(t, session.ExpiresAt.After(time.Now()))
		assert.Equal(t, LocalTokenDuration, time.Until(session.ExpiresAt).Round(time.Second))
	})

	t.Run("invalid email", func(t *testing.T) {
		session, err := store.Authenticate("wrong@example.com", password)
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("invalid password", func(t *testing.T) {
		session, err := store.Authenticate(email, "wrong-password")
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "invalid credentials")
	})
}

func TestValidateToken(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Register and authenticate to get a valid token
	email := "validate@example.com"
	password := "test-password"
	err := store.Register(email, password)
	require.NoError(t, err)

	session, err := store.Authenticate(email, password)
	require.NoError(t, err)

	t.Run("valid token", func(t *testing.T) {
		userID, err := store.ValidateToken(session.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, session.UserID, userID)
	})

	t.Run("invalid token", func(t *testing.T) {
		userID, err := store.ValidateToken("invalid-token")
		assert.Error(t, err)
		assert.Equal(t, int64(0), userID)
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("expired token", func(t *testing.T) {
		// Create a session with past expiration
		now := time.Now()
		expiredTime := now.Add(-1 * time.Hour)

		_, err := store.db.Exec(
			"INSERT INTO sessions (user_id, access_token, refresh_token, expires_at, created_at) VALUES (?, ?, ?, ?, ?)",
			session.UserID, "expired-token", "expired-refresh", expiredTime, now,
		)
		require.NoError(t, err)

		userID, err := store.ValidateToken("expired-token")
		assert.Error(t, err)
		assert.Equal(t, int64(0), userID)
		assert.Contains(t, err.Error(), "token expired")
	})
}

func TestRefreshSession(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Register and authenticate to get tokens
	email := "refresh@example.com"
	password := "test-password"
	err := store.Register(email, password)
	require.NoError(t, err)

	originalSession, err := store.Authenticate(email, password)
	require.NoError(t, err)

	t.Run("valid refresh token", func(t *testing.T) {
		newSession, err := store.RefreshSession(originalSession.RefreshToken)
		assert.NoError(t, err)
		assert.NotNil(t, newSession)
		assert.NotEqual(t, originalSession.ID, newSession.ID)
		assert.Equal(t, originalSession.UserID, newSession.UserID)
		assert.NotEqual(t, originalSession.AccessToken, newSession.AccessToken)
		assert.NotEqual(t, originalSession.RefreshToken, newSession.RefreshToken)

		// Old token should be invalid now
		_, err = store.ValidateToken(originalSession.AccessToken)
		assert.Error(t, err)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		newSession, err := store.RefreshSession("invalid-refresh-token")
		assert.Error(t, err)
		assert.Nil(t, newSession)
		assert.Contains(t, err.Error(), "invalid refresh token")
	})
}

func TestGetUser(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Register a user
	email := "getuser@example.com"
	password := "test-password"
	err := store.Register(email, password)
	require.NoError(t, err)

	// Get the user ID by authenticating
	session, err := store.Authenticate(email, password)
	require.NoError(t, err)

	t.Run("existing user", func(t *testing.T) {
		user, err := store.GetUser(session.UserID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, session.UserID, user.ID)
		assert.Equal(t, email, user.Email)
		assert.NotEmpty(t, user.PasswordHash)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("non-existent user", func(t *testing.T) {
		user, err := store.GetUser(99999)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestDeleteSession(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Register and authenticate
	email := "delete@example.com"
	password := "test-password"
	err := store.Register(email, password)
	require.NoError(t, err)

	session, err := store.Authenticate(email, password)
	require.NoError(t, err)

	t.Run("delete existing session", func(t *testing.T) {
		err := store.DeleteSession(session.AccessToken)
		assert.NoError(t, err)

		// Token should now be invalid
		_, err = store.ValidateToken(session.AccessToken)
		assert.Error(t, err)
	})

	t.Run("delete non-existent session", func(t *testing.T) {
		err := store.DeleteSession("non-existent-token")
		assert.NoError(t, err) // Should not error
	})
}

func TestDeleteAllSessions(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Register and create multiple sessions
	email := "deleteall@example.com"
	password := "test-password"
	err := store.Register(email, password)
	require.NoError(t, err)

	session1, err := store.Authenticate(email, password)
	require.NoError(t, err)

	session2, err := store.Authenticate(email, password)
	require.NoError(t, err)

	t.Run("delete all user sessions", func(t *testing.T) {
		err := store.DeleteAllSessions(session1.UserID)
		assert.NoError(t, err)

		// Both tokens should now be invalid
		_, err = store.ValidateToken(session1.AccessToken)
		assert.Error(t, err)

		_, err = store.ValidateToken(session2.AccessToken)
		assert.Error(t, err)
	})

	t.Run("delete sessions for non-existent user", func(t *testing.T) {
		err := store.DeleteAllSessions(99999)
		assert.NoError(t, err) // Should not error
	})
}

func TestSessionToTokenPair(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Register and authenticate
	email := "tokenpair@example.com"
	password := "test-password"
	err := store.Register(email, password)
	require.NoError(t, err)

	session, err := store.Authenticate(email, password)
	require.NoError(t, err)

	tokenPair := session.ToTokenPair()
	assert.NotNil(t, tokenPair)
	assert.Equal(t, session.AccessToken, tokenPair.AccessToken)
	assert.Equal(t, session.RefreshToken, tokenPair.RefreshToken)
	assert.Equal(t, "Bearer", tokenPair.TokenType)
	assert.Greater(t, tokenPair.ExpiresIn, int64(0))
}

func TestPasswordHashing(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	email := "hash@example.com"
	password := "my-secure-password"

	// Register user
	err := store.Register(email, password)
	require.NoError(t, err)

	// Get user
	session, err := store.Authenticate(email, password)
	require.NoError(t, err)

	user, err := store.GetUser(session.UserID)
	require.NoError(t, err)

	t.Run("password is hashed", func(t *testing.T) {
		assert.NotEqual(t, password, user.PasswordHash)
		assert.Contains(t, user.PasswordHash, "$2a$") // bcrypt prefix
	})

	t.Run("different passwords produce different hashes", func(t *testing.T) {
		email2 := "hash2@example.com"
		err := store.Register(email2, password)
		require.NoError(t, err)

		session2, err := store.Authenticate(email2, password)
		require.NoError(t, err)

		user2, err := store.GetUser(session2.UserID)
		require.NoError(t, err)

		// Same password should produce different hashes due to salt
		assert.NotEqual(t, user.PasswordHash, user2.PasswordHash)
	})
}

func TestGenerateToken(t *testing.T) {
	token1, err := generateToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := generateToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Tokens should be unique
	assert.NotEqual(t, token1, token2)

	// Tokens should be base64 URL-encoded
	assert.NotContains(t, token1, "+")
	assert.NotContains(t, token1, "/")
}

func TestConcurrentAccess(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Register multiple users concurrently
	const numUsers = 10

	done := make(chan bool, numUsers)
	for i := 0; i < numUsers; i++ {
		go func(n int) {
			email := "concurrent" + string(rune('0'+n)) + "@example.com"
			password := "password" + string(rune('0'+n))

			err := store.Register(email, password)
			assert.NoError(t, err)

			session, err := store.Authenticate(email, password)
			assert.NoError(t, err)
			assert.NotNil(t, session)

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numUsers; i++ {
		<-done
	}
}

func TestStoreClose(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "local-auth-test-close-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Create store
	store, err := NewStore(dbPath)
	require.NoError(t, err)

	// Close the store
	err = store.Close()
	assert.NoError(t, err)

	// Operations should fail after close
	_, err = store.Authenticate("test@example.com", "password")
	assert.Error(t, err)
}
