package refresh

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth/jwt"
	"github.com/AINative-studio/ainative-code/internal/auth/oauth"
)

const (
	// DefaultRefreshThreshold is the time before expiration to trigger refresh
	DefaultRefreshThreshold = 5 * time.Minute

	// DefaultCheckInterval is how often to check token expiration
	DefaultCheckInterval = 1 * time.Minute

	// MinRefreshThreshold is the minimum allowed refresh threshold
	MinRefreshThreshold = 1 * time.Minute
)

// Config represents the configuration for the refresh manager.
type Config struct {
	// OAuthClient is the OAuth client for token refresh
	OAuthClient *oauth.Client

	// TokenStore is called to store refreshed tokens
	TokenStore TokenStoreFunc

	// OnRefreshFail is called when refresh fails
	OnRefreshFail RefreshFailFunc

	// RefreshThreshold is the time before expiry to trigger refresh
	// Default: 5 minutes
	RefreshThreshold time.Duration

	// CheckInterval is how often to check expiration
	// Default: 1 minute
	CheckInterval time.Duration
}

// TokenStoreFunc is called to store refreshed tokens.
type TokenStoreFunc func(tokens *jwt.TokenPair) error

// RefreshFailFunc is called when token refresh fails.
// It should return true if the manager should attempt re-authentication.
type RefreshFailFunc func(err error) bool

// Manager manages automatic token refresh.
type Manager struct {
	config        Config
	mu            sync.RWMutex
	tokens        *jwt.TokenPair
	expiresAt     time.Time
	stopChan      chan struct{}
	stoppedChan   chan struct{}
	running       bool
	lastRefreshAt time.Time
}

// NewManager creates a new token refresh manager.
func NewManager(config Config) *Manager {
	if config.RefreshThreshold == 0 {
		config.RefreshThreshold = DefaultRefreshThreshold
	}

	if config.RefreshThreshold < MinRefreshThreshold {
		config.RefreshThreshold = MinRefreshThreshold
	}

	if config.CheckInterval == 0 {
		config.CheckInterval = DefaultCheckInterval
	}

	return &Manager{
		config:      config,
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}
}

// Start begins monitoring and refreshing tokens.
func (m *Manager) Start(ctx context.Context, tokens *jwt.TokenPair) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("manager already running")
	}

	if tokens == nil {
		return fmt.Errorf("tokens cannot be nil")
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)

	m.tokens = tokens
	m.expiresAt = expiresAt
	m.running = true

	// Start background goroutine
	go m.refreshLoop(ctx)

	return nil
}

// Stop gracefully stops the refresh manager.
func (m *Manager) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}

	close(m.stopChan)
	m.running = false
	m.mu.Unlock()

	// Wait for goroutine to finish
	<-m.stoppedChan
}

// IsRunning returns whether the manager is currently running.
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetTokens returns the current tokens.
func (m *Manager) GetTokens() *jwt.TokenPair {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tokens
}

// GetExpiresAt returns the token expiration time.
func (m *Manager) GetExpiresAt() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.expiresAt
}

// GetLastRefreshAt returns the time of the last successful refresh.
func (m *Manager) GetLastRefreshAt() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastRefreshAt
}

// refreshLoop is the background goroutine that monitors and refreshes tokens.
func (m *Manager) refreshLoop(ctx context.Context) {
	defer close(m.stoppedChan)

	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.checkAndRefresh(ctx); err != nil {
				// Log error but continue running
				fmt.Printf("token refresh check failed: %v\n", err)
			}
		}
	}
}

// checkAndRefresh checks if refresh is needed and performs it.
func (m *Manager) checkAndRefresh(ctx context.Context) error {
	m.mu.RLock()
	expiresAt := m.expiresAt
	refreshToken := ""
	if m.tokens != nil {
		refreshToken = m.tokens.RefreshToken
	}
	m.mu.RUnlock()

	// Check if refresh is needed
	now := time.Now()
	refreshAt := expiresAt.Add(-m.config.RefreshThreshold)

	if now.Before(refreshAt) {
		// Not time to refresh yet
		return nil
	}

	// Time to refresh
	return m.performRefresh(ctx, refreshToken)
}

// performRefresh executes the token refresh.
func (m *Manager) performRefresh(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// Call OAuth client to refresh
	newTokens, err := m.config.OAuthClient.RefreshToken(ctx, refreshToken)
	if err != nil {
		// Handle refresh failure
		if m.config.OnRefreshFail != nil {
			shouldReauth := m.config.OnRefreshFail(err)
			if shouldReauth {
				return fmt.Errorf("refresh failed, re-authentication required: %w", err)
			}
		}
		return fmt.Errorf("token refresh failed: %w", err)
	}

	// Update tokens
	m.mu.Lock()
	m.tokens = newTokens
	m.expiresAt = time.Now().Add(time.Duration(newTokens.ExpiresIn) * time.Second)
	m.lastRefreshAt = time.Now()
	m.mu.Unlock()

	// Store new tokens
	if m.config.TokenStore != nil {
		if err := m.config.TokenStore(newTokens); err != nil {
			return fmt.Errorf("failed to store refreshed tokens: %w", err)
		}
	}

	fmt.Printf("tokens refreshed successfully, expires at: %s\n", m.expiresAt.Format(time.RFC3339))

	return nil
}

// ForceRefresh immediately refreshes the tokens regardless of expiration time.
func (m *Manager) ForceRefresh(ctx context.Context) error {
	m.mu.RLock()
	refreshToken := ""
	if m.tokens != nil {
		refreshToken = m.tokens.RefreshToken
	}
	m.mu.RUnlock()

	if refreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	return m.performRefresh(ctx, refreshToken)
}

// UpdateTokens manually updates the tokens (useful after re-authentication).
func (m *Manager) UpdateTokens(tokens *jwt.TokenPair) error {
	if tokens == nil {
		return fmt.Errorf("tokens cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokens = tokens
	m.expiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)

	return nil
}

// GetRefreshStatus returns the current refresh status.
func (m *Manager) GetRefreshStatus() *RefreshStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	refreshAt := m.expiresAt.Add(-m.config.RefreshThreshold)

	return &RefreshStatus{
		IsRunning:       m.running,
		ExpiresAt:       m.expiresAt,
		RefreshAt:       refreshAt,
		LastRefreshAt:   m.lastRefreshAt,
		TimeUntilExpiry: m.expiresAt.Sub(now),
		TimeUntilRefresh: refreshAt.Sub(now),
		NeedsRefresh:    now.After(refreshAt),
	}
}

// RefreshStatus represents the current refresh status.
type RefreshStatus struct {
	IsRunning        bool
	ExpiresAt        time.Time
	RefreshAt        time.Time
	LastRefreshAt    time.Time
	TimeUntilExpiry  time.Duration
	TimeUntilRefresh time.Duration
	NeedsRefresh     bool
}
