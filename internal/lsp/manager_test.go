package lsp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ManagerTestSuite is the test suite for LSP manager
type ManagerTestSuite struct {
	suite.Suite
	manager *Manager
}

func (s *ManagerTestSuite) SetupTest() {
	s.manager = NewManager()
}

func (s *ManagerTestSuite) TearDownTest() {
	if s.manager != nil {
		s.manager.Close()
	}
}

func (s *ManagerTestSuite) TestManagerRegisterLanguage() {
	config := DefaultConfig("go")

	err := s.manager.RegisterLanguage(config)
	s.Assert().NoError(err)

	// Verify registration
	registered := s.manager.GetRegisteredLanguages()
	s.Assert().Contains(registered, "go")
}

func (s *ManagerTestSuite) TestManagerRegisterDuplicateLanguage() {
	config := DefaultConfig("go")

	err := s.manager.RegisterLanguage(config)
	s.Require().NoError(err)

	// Try to register again
	err = s.manager.RegisterLanguage(config)
	s.Assert().Error(err)
}

func (s *ManagerTestSuite) TestManagerUnregisterLanguage() {
	config := DefaultConfig("go")

	err := s.manager.RegisterLanguage(config)
	s.Require().NoError(err)

	err = s.manager.UnregisterLanguage("go")
	s.Assert().NoError(err)

	registered := s.manager.GetRegisteredLanguages()
	s.Assert().NotContains(registered, "go")
}

func (s *ManagerTestSuite) TestManagerUnregisterNonexistentLanguage() {
	err := s.manager.UnregisterLanguage("nonexistent")
	s.Assert().Error(err)
}

func (s *ManagerTestSuite) TestManagerGetClient() {
	// Register language with mock config
	config := &LanguageServerConfig{
		Language:       "go",
		Command:        "mock-lsp",
		InitTimeout:    5 * time.Second,
		RequestTimeout: 2 * time.Second,
		Env:            make(map[string]string),
	}

	err := s.manager.RegisterLanguage(config)
	s.Require().NoError(err)

	// Get client (should create it)
	client, err := s.manager.GetClient("go")
	s.Assert().NoError(err)
	s.Assert().NotNil(client)

	// Get client again (should return same instance)
	client2, err := s.manager.GetClient("go")
	s.Assert().NoError(err)
	s.Assert().Equal(client, client2)
}

func (s *ManagerTestSuite) TestManagerGetClientUnregistered() {
	_, err := s.manager.GetClient("nonexistent")
	s.Assert().Error(err)
}

func (s *ManagerTestSuite) TestManagerInitializeClient() {
	// Create mock client without starting process
	config := &LanguageServerConfig{
		Language:       "go",
		Command:        "mock-lsp",
		InitTimeout:    5 * time.Second,
		RequestTimeout: 2 * time.Second,
		Env:            make(map[string]string),
	}

	err := s.manager.RegisterLanguage(config)
	s.Require().NoError(err)

	// We can't fully test initialization without a real server
	// but we can verify the manager tracks the client
	client, err := s.manager.GetClient("go")
	s.Assert().NoError(err)
	s.Assert().NotNil(client)
}

func (s *ManagerTestSuite) TestManagerCloseClient() {
	config := &LanguageServerConfig{
		Language:       "go",
		Command:        "mock-lsp",
		InitTimeout:    5 * time.Second,
		RequestTimeout: 2 * time.Second,
		Env:            make(map[string]string),
	}

	err := s.manager.RegisterLanguage(config)
	s.Require().NoError(err)

	client, err := s.manager.GetClient("go")
	s.Require().NoError(err)
	s.Require().NotNil(client)

	err = s.manager.CloseClient("go")
	s.Assert().NoError(err)

	// Getting client again should create a new instance
	client2, err := s.manager.GetClient("go")
	s.Assert().NoError(err)
	s.Assert().NotEqual(client, client2)
}

func (s *ManagerTestSuite) TestManagerCloseNonexistentClient() {
	err := s.manager.CloseClient("nonexistent")
	s.Assert().NoError(err) // Should not error, just no-op
}

func (s *ManagerTestSuite) TestManagerCloseAll() {
	// Register multiple languages
	for _, lang := range []string{"go", "python", "typescript"} {
		config := DefaultConfig(lang)
		err := s.manager.RegisterLanguage(config)
		s.Require().NoError(err)

		_, err = s.manager.GetClient(lang)
		s.Require().NoError(err)
	}

	// Close all
	s.manager.Close()

	// Verify all clients are closed
	s.manager.mu.RLock()
	s.Assert().Len(s.manager.clients, 0)
	s.manager.mu.RUnlock()
}

func (s *ManagerTestSuite) TestManagerGetClientCount() {
	// Initially no clients
	count := s.manager.GetClientCount()
	s.Assert().Equal(0, count)

	// Register and get clients
	for _, lang := range []string{"go", "python"} {
		config := DefaultConfig(lang)
		err := s.manager.RegisterLanguage(config)
		s.Require().NoError(err)

		_, err = s.manager.GetClient(lang)
		s.Require().NoError(err)
	}

	count = s.manager.GetClientCount()
	s.Assert().Equal(2, count)
}

func (s *ManagerTestSuite) TestManagerHealthCheck() {
	config := &LanguageServerConfig{
		Language:       "go",
		Command:        "mock-lsp",
		InitTimeout:    5 * time.Second,
		RequestTimeout: 2 * time.Second,
		Env:            make(map[string]string),
	}

	err := s.manager.RegisterLanguage(config)
	s.Require().NoError(err)

	client, err := s.manager.GetClient("go")
	s.Require().NoError(err)
	s.Require().NotNil(client)

	// Simulate initialization for health check
	client.initialized = true

	// Check health
	healthy := s.manager.IsHealthy("go")
	s.Assert().True(healthy) // Client exists and is initialized
}

func (s *ManagerTestSuite) TestManagerHealthCheckNonexistent() {
	healthy := s.manager.IsHealthy("nonexistent")
	s.Assert().False(healthy)
}

func (s *ManagerTestSuite) TestManagerRegisterMultipleLanguages() {
	languages := []string{"go", "python", "typescript", "rust"}

	for _, lang := range languages {
		config := DefaultConfig(lang)
		err := s.manager.RegisterLanguage(config)
		s.Assert().NoError(err)
	}

	registered := s.manager.GetRegisteredLanguages()
	s.Assert().Len(registered, len(languages))

	for _, lang := range languages {
		s.Assert().Contains(registered, lang)
	}
}

func (s *ManagerTestSuite) TestManagerConcurrentAccess() {
	// Test concurrent registration
	languages := []string{"go", "python", "typescript", "rust"}

	done := make(chan bool)
	for _, lang := range languages {
		lang := lang // capture variable
		go func() {
			config := DefaultConfig(lang)
			s.manager.RegisterLanguage(config)
			done <- true
		}()
	}

	for range languages {
		<-done
	}

	// Test concurrent client access
	for _, lang := range languages {
		lang := lang // capture variable
		go func() {
			s.manager.GetClient(lang)
			done <- true
		}()
	}

	for range languages {
		<-done
	}

	// Verify all registered
	registered := s.manager.GetRegisteredLanguages()
	s.Assert().Len(registered, len(languages))
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

// Unit tests for manager helper functions
func TestNewManager(t *testing.T) {
	manager := NewManager()
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.configs)
	assert.NotNil(t, manager.clients)
}

func TestManagerGetConfig(t *testing.T) {
	manager := NewManager()

	// Register a language
	config := DefaultConfig("go")
	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// Get config
	retrieved := manager.GetConfig("go")
	require.NotNil(t, retrieved)
	assert.Equal(t, "go", retrieved.Language)
	assert.Equal(t, "gopls", retrieved.Command)

	// Get non-existent config
	retrieved = manager.GetConfig("nonexistent")
	assert.Nil(t, retrieved)
}

func TestManagerUpdateConfig(t *testing.T) {
	manager := NewManager()

	// Register a language
	config := DefaultConfig("go")
	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// Update config
	newConfig := DefaultConfig("go")
	newConfig.RequestTimeout = 20 * time.Second

	err = manager.UpdateConfig("go", newConfig)
	assert.NoError(t, err)

	// Verify update
	retrieved := manager.GetConfig("go")
	assert.Equal(t, 20*time.Second, retrieved.RequestTimeout)
}

func TestManagerUpdateConfigNonexistent(t *testing.T) {
	manager := NewManager()

	config := DefaultConfig("go")
	err := manager.UpdateConfig("nonexistent", config)
	assert.Error(t, err)
}

func TestManagerAutoRestart(t *testing.T) {
	manager := NewManager()

	config := &LanguageServerConfig{
		Language:       "go",
		Command:        "mock-lsp",
		InitTimeout:    5 * time.Second,
		RequestTimeout: 2 * time.Second,
		AutoRestart:    true,
		MaxRestarts:    3,
		Env:            make(map[string]string),
	}

	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// Get restart count (should be 0 initially)
	count := manager.GetRestartCount("go")
	assert.Equal(t, 0, count)
}

func TestManagerWithInitialization(t *testing.T) {
	manager := NewManager()

	// Register language
	config := &LanguageServerConfig{
		Language:       "go",
		Command:        "mock-lsp",
		InitTimeout:    5 * time.Second,
		RequestTimeout: 2 * time.Second,
		Env:            make(map[string]string),
	}

	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// Create context
	ctx := context.Background()

	// Try to initialize (will fail without real server, but tests the path)
	_, err = manager.InitializeClient(ctx, "go", "file:///workspace")
	assert.Error(t, err) // Expected to fail without real server
}

func TestManagerGetOrCreateClient(t *testing.T) {
	manager := NewManager()

	config := &LanguageServerConfig{
		Language:       "go",
		Command:        "mock-lsp",
		InitTimeout:    5 * time.Second,
		RequestTimeout: 2 * time.Second,
		Env:            make(map[string]string),
	}

	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// First call should create client
	client1, err := manager.GetClient("go")
	require.NoError(t, err)
	require.NotNil(t, client1)

	// Second call should return same client
	client2, err := manager.GetClient("go")
	require.NoError(t, err)
	assert.Equal(t, client1, client2)
}

func TestManagerRegistrationValidation(t *testing.T) {
	manager := NewManager()

	// Invalid config (missing language)
	config := &LanguageServerConfig{
		Command:        "gopls",
		InitTimeout:    5 * time.Second,
		RequestTimeout: 2 * time.Second,
	}

	err := manager.RegisterLanguage(config)
	assert.Error(t, err)
}
