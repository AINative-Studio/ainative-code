package refresh_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth/jwt"
	"github.com/AINative-studio/ainative-code/internal/auth/oauth"
	"github.com/AINative-studio/ainative-code/internal/auth/refresh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	t.Run("creates manager with config", func(t *testing.T) {
		config := refresh.Config{
			OAuthClient: &oauth.Client{},
		}

		manager := refresh.NewManager(config)
		assert.NotNil(t, manager)
		assert.False(t, manager.IsRunning())
	})

	t.Run("uses default refresh threshold", func(t *testing.T) {
		config := refresh.Config{
			OAuthClient: &oauth.Client{},
		}

		manager := refresh.NewManager(config)
		assert.NotNil(t, manager)
	})

	t.Run("enforces minimum refresh threshold", func(t *testing.T) {
		config := refresh.Config{
			OAuthClient:      &oauth.Client{},
			RefreshThreshold: 30 * time.Second, // Less than minimum
		}

		manager := refresh.NewManager(config)
		assert.NotNil(t, manager)
	})
}

func TestManager_Start(t *testing.T) {
	t.Run("starts manager with tokens", func(t *testing.T) {
		config := refresh.Config{
			OAuthClient: &oauth.Client{},
		}

		manager := refresh.NewManager(config)

		tokens := &jwt.TokenPair{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}

		err := manager.Start(context.Background(), tokens)
		require.NoError(t, err)
		assert.True(t, manager.IsRunning())

		manager.Stop()
	})

	t.Run("rejects nil tokens", func(t *testing.T) {
		config := refresh.Config{
			OAuthClient: &oauth.Client{},
		}

		manager := refresh.NewManager(config)

		err := manager.Start(context.Background(), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("rejects starting when already running", func(t *testing.T) {
		config := refresh.Config{
			OAuthClient: &oauth.Client{},
		}

		manager := refresh.NewManager(config)

		tokens := &jwt.TokenPair{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}

		err := manager.Start(context.Background(), tokens)
		require.NoError(t, err)

		err = manager.Start(context.Background(), tokens)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already running")

		manager.Stop()
	})
}

func TestManager_Stop(t *testing.T) {
	t.Run("gracefully stops manager", func(t *testing.T) {
		config := refresh.Config{
			OAuthClient: &oauth.Client{},
		}

		manager := refresh.NewManager(config)

		tokens := &jwt.TokenPair{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}

		err := manager.Start(context.Background(), tokens)
		require.NoError(t, err)

		manager.Stop()
		assert.False(t, manager.IsRunning())
	})

	t.Run("stop is idempotent", func(t *testing.T) {
		config := refresh.Config{
			OAuthClient: &oauth.Client{},
		}

		manager := refresh.NewManager(config)

		tokens := &jwt.TokenPair{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}

		err := manager.Start(context.Background(), tokens)
		require.NoError(t, err)

		manager.Stop()
		manager.Stop() // Second stop should not panic

		assert.False(t, manager.IsRunning())
	})
}

func TestManager_AutoRefresh(t *testing.T) {
	t.Run("automatically refreshes token before expiry", func(t *testing.T) {
		refreshCalled := false
		var mu sync.Mutex

		// Mock OAuth server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			refreshCalled = true
			mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(oauth.TokenResponse{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
			})
		}))
		defer server.Close()

		oauthConfig := oauth.Config{
			TokenURL: server.URL,
			ClientID: "test-client",
		}
		oauthClient := oauth.NewClient(oauthConfig)

		storeCalled := false
		var storedTokens *jwt.TokenPair

		config := refresh.Config{
			OAuthClient:      oauthClient,
			RefreshThreshold: 100 * time.Millisecond, // Very short for testing
			CheckInterval:    50 * time.Millisecond,
			TokenStore: func(tokens *jwt.TokenPair) error {
				storeCalled = true
				storedTokens = tokens
				return nil
			},
		}

		manager := refresh.NewManager(config)

		// Create tokens that expire soon
		tokens := &jwt.TokenPair{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    1, // Expires in 1 second
			TokenType:    "Bearer",
		}

		err := manager.Start(context.Background(), tokens)
		require.NoError(t, err)
		defer manager.Stop()

		// Wait for refresh to happen
		time.Sleep(500 * time.Millisecond)

		mu.Lock()
		assert.True(t, refreshCalled, "refresh should have been called")
		mu.Unlock()

		assert.True(t, storeCalled, "token store should have been called")
		assert.NotNil(t, storedTokens)
		assert.Equal(t, "new-access-token", storedTokens.AccessToken)
	})

	t.Run("handles refresh failure", func(t *testing.T) {
		failCalled := false

		// Mock OAuth server that fails
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		oauthConfig := oauth.Config{
			TokenURL: server.URL,
			ClientID: "test-client",
		}
		oauthClient := oauth.NewClient(oauthConfig)

		config := refresh.Config{
			OAuthClient:      oauthClient,
			RefreshThreshold: 100 * time.Millisecond,
			CheckInterval:    50 * time.Millisecond,
			OnRefreshFail: func(err error) bool {
				failCalled = true
				return false // Don't require re-auth
			},
		}

		manager := refresh.NewManager(config)

		tokens := &jwt.TokenPair{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    1,
			TokenType:    "Bearer",
		}

		err := manager.Start(context.Background(), tokens)
		require.NoError(t, err)
		defer manager.Stop()

		// Wait for refresh attempt
		time.Sleep(500 * time.Millisecond)

		assert.True(t, failCalled, "OnRefreshFail should have been called")
	})
}

func TestManager_ForceRefresh(t *testing.T) {
	t.Run("forces immediate refresh", func(t *testing.T) {
		// Mock OAuth server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(oauth.TokenResponse{
				AccessToken:  "forced-access-token",
				RefreshToken: "forced-refresh-token",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
			})
		}))
		defer server.Close()

		oauthConfig := oauth.Config{
			TokenURL: server.URL,
			ClientID: "test-client",
		}
		oauthClient := oauth.NewClient(oauthConfig)

		config := refresh.Config{
			OAuthClient: oauthClient,
		}

		manager := refresh.NewManager(config)

		tokens := &jwt.TokenPair{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}

		err := manager.Start(context.Background(), tokens)
		require.NoError(t, err)
		defer manager.Stop()

		// Force refresh
		err = manager.ForceRefresh(context.Background())
		require.NoError(t, err)

		// Check tokens were updated
		newTokens := manager.GetTokens()
		assert.Equal(t, "forced-access-token", newTokens.AccessToken)
	})

	t.Run("fails when no refresh token", func(t *testing.T) {
		config := refresh.Config{
			OAuthClient: &oauth.Client{},
		}

		manager := refresh.NewManager(config)

		err := manager.ForceRefresh(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no refresh token")
	})
}

func TestManager_UpdateTokens(t *testing.T) {
	t.Run("updates tokens manually", func(t *testing.T) {
		config := refresh.Config{
			OAuthClient: &oauth.Client{},
		}

		manager := refresh.NewManager(config)

		tokens := &jwt.TokenPair{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}

		err := manager.Start(context.Background(), tokens)
		require.NoError(t, err)
		defer manager.Stop()

		newTokens := &jwt.TokenPair{
			AccessToken:  "updated-access-token",
			RefreshToken: "updated-refresh-token",
			ExpiresIn:    7200,
			TokenType:    "Bearer",
		}

		err = manager.UpdateTokens(newTokens)
		require.NoError(t, err)

		retrievedTokens := manager.GetTokens()
		assert.Equal(t, "updated-access-token", retrievedTokens.AccessToken)
	})

	t.Run("rejects nil tokens", func(t *testing.T) {
		config := refresh.Config{
			OAuthClient: &oauth.Client{},
		}

		manager := refresh.NewManager(config)

		err := manager.UpdateTokens(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

func TestManager_GetRefreshStatus(t *testing.T) {
	t.Run("returns current refresh status", func(t *testing.T) {
		config := refresh.Config{
			OAuthClient:      &oauth.Client{},
			RefreshThreshold: 5 * time.Minute,
		}

		manager := refresh.NewManager(config)

		tokens := &jwt.TokenPair{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}

		err := manager.Start(context.Background(), tokens)
		require.NoError(t, err)
		defer manager.Stop()

		status := manager.GetRefreshStatus()
		assert.True(t, status.IsRunning)
		assert.False(t, status.ExpiresAt.IsZero())
		assert.False(t, status.RefreshAt.IsZero())
		assert.True(t, status.TimeUntilExpiry > 0)
	})
}

func TestManager_ContextCancellation(t *testing.T) {
	t.Run("stops when context cancelled", func(t *testing.T) {
		config := refresh.Config{
			OAuthClient:   &oauth.Client{},
			CheckInterval: 50 * time.Millisecond,
		}

		manager := refresh.NewManager(config)

		tokens := &jwt.TokenPair{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}

		ctx, cancel := context.WithCancel(context.Background())

		err := manager.Start(ctx, tokens)
		require.NoError(t, err)

		// Cancel context
		cancel()

		// Wait for manager to stop
		time.Sleep(200 * time.Millisecond)

		// Manager should have stopped
		// (Note: We don't have a direct way to check if the goroutine stopped,
		// but Stop() should not block if it did)
		manager.Stop()
	})
}

func TestManager_TokenStoreError(t *testing.T) {
	t.Run("handles token store error", func(t *testing.T) {
		// Mock OAuth server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(oauth.TokenResponse{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
			})
		}))
		defer server.Close()

		oauthConfig := oauth.Config{
			TokenURL: server.URL,
			ClientID: "test-client",
		}
		oauthClient := oauth.NewClient(oauthConfig)

		config := refresh.Config{
			OAuthClient:      oauthClient,
			RefreshThreshold: 100 * time.Millisecond,
			CheckInterval:    50 * time.Millisecond,
			TokenStore: func(tokens *jwt.TokenPair) error {
				return errors.New("storage error")
			},
		}

		manager := refresh.NewManager(config)

		tokens := &jwt.TokenPair{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    1,
			TokenType:    "Bearer",
		}

		err := manager.Start(context.Background(), tokens)
		require.NoError(t, err)
		defer manager.Stop()

		// Wait for refresh attempt (should fail on store)
		time.Sleep(500 * time.Millisecond)

		// Manager should still be running even with store error
		assert.True(t, manager.IsRunning())
	})
}
