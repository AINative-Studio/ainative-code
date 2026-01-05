package lsp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Manager manages multiple language server clients
type Manager struct {
	configs  map[string]*LanguageServerConfig
	clients  map[string]*Client
	restarts map[string]int
	mu       sync.RWMutex

	// Health monitoring
	healthChecks map[string]*time.Ticker
	healthMu     sync.RWMutex

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewManager creates a new language server manager
func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		configs:      make(map[string]*LanguageServerConfig),
		clients:      make(map[string]*Client),
		restarts:     make(map[string]int),
		healthChecks: make(map[string]*time.Ticker),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// RegisterLanguage registers a language server configuration
func (m *Manager) RegisterLanguage(config *LanguageServerConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.configs[config.Language]; exists {
		return &JSONRPCError{
			Code:    InternalError,
			Message: fmt.Sprintf("language %s already registered", config.Language),
		}
	}

	m.configs[config.Language] = config.Clone()
	m.restarts[config.Language] = 0

	// Start health check if enabled
	if config.HealthCheckInterval > 0 {
		m.startHealthCheck(config.Language)
	}

	return nil
}

// UnregisterLanguage unregisters a language server
func (m *Manager) UnregisterLanguage(language string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.configs[language]; !exists {
		return &JSONRPCError{
			Code:    InternalError,
			Message: fmt.Sprintf("language %s not registered", language),
		}
	}

	// Close client if running
	if client, exists := m.clients[language]; exists {
		client.Close()
		delete(m.clients, language)
	}

	// Stop health check
	m.stopHealthCheck(language)

	delete(m.configs, language)
	delete(m.restarts, language)

	return nil
}

// GetClient returns or creates a client for the specified language
func (m *Manager) GetClient(language string) (*Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if language is registered
	config, exists := m.configs[language]
	if !exists {
		return nil, &JSONRPCError{
			Code:    InternalError,
			Message: fmt.Sprintf("language %s not registered", language),
		}
	}

	// Return existing client if available
	if client, exists := m.clients[language]; exists {
		return client, nil
	}

	// Create new client
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}

	m.clients[language] = client
	return client, nil
}

// InitializeClient initializes a client for the specified language
func (m *Manager) InitializeClient(ctx context.Context, language string, rootURI string) (*InitializeResult, error) {
	client, err := m.GetClient(language)
	if err != nil {
		return nil, err
	}

	// Start the language server if not already started
	m.mu.RLock()
	config := m.configs[language]
	m.mu.RUnlock()

	// Only start if we have a real command (not mock)
	if client.cmd == nil && config.Command != "mock-lsp" {
		if err := client.Start(); err != nil {
			return nil, err
		}
	}

	// Initialize
	rootURIPtr := &rootURI
	result, err := client.Initialize(ctx, rootURIPtr, config.InitializationOptions)
	if err != nil {
		// If initialization fails and auto-restart is enabled, try to restart
		if config.AutoRestart && m.shouldRestart(language) {
			m.incrementRestartCount(language)
			client.Close()

			// Remove from clients map so next GetClient creates a new one
			m.mu.Lock()
			delete(m.clients, language)
			m.mu.Unlock()

			return nil, fmt.Errorf("initialization failed, will retry: %w", err)
		}

		return nil, err
	}

	// Send initialized notification
	if err := client.Initialized(ctx); err != nil {
		return nil, err
	}

	// Reset restart count on successful initialization
	m.resetRestartCount(language)

	return result, nil
}

// CloseClient closes the client for the specified language
func (m *Manager) CloseClient(language string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[language]
	if !exists {
		return nil // No-op if client doesn't exist
	}

	// Try graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client.Shutdown(ctx)
	client.Exit(ctx)
	client.Close()

	delete(m.clients, language)
	return nil
}

// Close closes all clients and stops the manager
func (m *Manager) Close() {
	m.cancel()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all clients
	for language, client := range m.clients {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		client.Shutdown(ctx)
		client.Exit(ctx)
		client.Close()
		cancel()

		delete(m.clients, language)
	}

	// Stop all health checks
	for language := range m.healthChecks {
		m.stopHealthCheck(language)
	}

	m.wg.Wait()
}

// GetRegisteredLanguages returns a list of registered languages
func (m *Manager) GetRegisteredLanguages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	languages := make([]string, 0, len(m.configs))
	for lang := range m.configs {
		languages = append(languages, lang)
	}

	return languages
}

// GetClientCount returns the number of active clients
func (m *Manager) GetClientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.clients)
}

// GetConfig returns the configuration for a language
func (m *Manager) GetConfig(language string) *LanguageServerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if config, exists := m.configs[language]; exists {
		return config.Clone()
	}

	return nil
}

// UpdateConfig updates the configuration for a language
func (m *Manager) UpdateConfig(language string, config *LanguageServerConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.configs[language]; !exists {
		return &JSONRPCError{
			Code:    InternalError,
			Message: fmt.Sprintf("language %s not registered", language),
		}
	}

	m.configs[language] = config.Clone()
	return nil
}

// IsHealthy checks if a language server is healthy
func (m *Manager) IsHealthy(language string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[language]
	if !exists {
		return false
	}

	// For now, we consider a client healthy if it exists and is initialized
	return client.isInitialized() && !client.IsShutdown()
}

// GetRestartCount returns the restart count for a language
func (m *Manager) GetRestartCount(language string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.restarts[language]
}

// startHealthCheck starts periodic health checks for a language
func (m *Manager) startHealthCheck(language string) {
	m.healthMu.Lock()
	defer m.healthMu.Unlock()

	// Don't start if already running
	if _, exists := m.healthChecks[language]; exists {
		return
	}

	config := m.configs[language]
	if config.HealthCheckInterval <= 0 {
		return
	}

	ticker := time.NewTicker(config.HealthCheckInterval)
	m.healthChecks[language] = ticker

	m.wg.Add(1)
	go m.runHealthCheck(language, ticker)
}

// stopHealthCheck stops health checks for a language
func (m *Manager) stopHealthCheck(language string) {
	m.healthMu.Lock()
	defer m.healthMu.Unlock()

	ticker, exists := m.healthChecks[language]
	if !exists {
		return
	}

	ticker.Stop()
	delete(m.healthChecks, language)
}

// runHealthCheck runs periodic health checks
func (m *Manager) runHealthCheck(language string, ticker *time.Ticker) {
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if !m.IsHealthy(language) {
				m.handleUnhealthyServer(language)
			}
		}
	}
}

// handleUnhealthyServer handles an unhealthy server
func (m *Manager) handleUnhealthyServer(language string) {
	m.mu.RLock()
	config, exists := m.configs[language]
	m.mu.RUnlock()

	if !exists || !config.AutoRestart {
		return
	}

	if !m.shouldRestart(language) {
		return
	}

	// Close the unhealthy client
	m.CloseClient(language)
	m.incrementRestartCount(language)

	// The client will be recreated on next GetClient call
}

// shouldRestart checks if a server should be restarted
func (m *Manager) shouldRestart(language string) bool {
	m.mu.RLock()
	config := m.configs[language]
	restartCount := m.restarts[language]
	m.mu.RUnlock()

	if config == nil || !config.AutoRestart {
		return false
	}

	return restartCount < config.MaxRestarts
}

// incrementRestartCount increments the restart count
func (m *Manager) incrementRestartCount(language string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.restarts[language]++
}

// resetRestartCount resets the restart count
func (m *Manager) resetRestartCount(language string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.restarts[language] = 0
}

// RegisterDefaultLanguages registers all supported languages with default configurations
func (m *Manager) RegisterDefaultLanguages() error {
	for _, language := range SupportedLanguages() {
		config := DefaultConfig(language)
		if err := m.RegisterLanguage(config); err != nil {
			return err
		}
	}
	return nil
}

// GetActiveLanguages returns a list of languages with active clients
func (m *Manager) GetActiveLanguages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	languages := make([]string, 0, len(m.clients))
	for lang := range m.clients {
		languages = append(languages, lang)
	}

	return languages
}
