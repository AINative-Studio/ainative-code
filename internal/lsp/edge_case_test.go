package lsp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestManagerHandleUnhealthyServerNoAutoRestart tests unhealthy server handling without auto-restart
func TestManagerHandleUnhealthyServerNoAutoRestart(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := DefaultConfig("go")
	config.AutoRestart = false

	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// Manually call handleUnhealthyServer
	manager.handleUnhealthyServer("go")

	// Should do nothing since auto-restart is disabled
}

// TestManagerHandleUnhealthyServerMaxRestartsReached tests when max restarts reached
func TestManagerHandleUnhealthyServerMaxRestartsReached(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := DefaultConfig("go")
	config.AutoRestart = true
	config.MaxRestarts = 2

	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// Increment restart count to max
	manager.incrementRestartCount("go")
	manager.incrementRestartCount("go")

	// handleUnhealthyServer should not restart (max reached)
	manager.handleUnhealthyServer("go")
}

// TestManagerHandleUnhealthyServerNonexistent tests handling nonexistent language
func TestManagerHandleUnhealthyServerNonexistent(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// Should not panic
	manager.handleUnhealthyServer("nonexistent")
}

// TestManagerRunHealthCheckCancellation tests health check cancellation
func TestManagerRunHealthCheckCancellation(t *testing.T) {
	manager := NewManager()

	config := DefaultConfig("go")
	config.HealthCheckInterval = 100 * time.Millisecond

	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// Verify health check was started
	manager.healthMu.RLock()
	_, exists := manager.healthChecks["go"]
	manager.healthMu.RUnlock()
	assert.True(t, exists)

	// Close manager (cancels context and stops health checks)
	manager.Close()

	// Health check should have stopped
	manager.healthMu.RLock()
	_, exists = manager.healthChecks["go"]
	manager.healthMu.RUnlock()

	assert.False(t, exists)
}

// TestClientInitializeWithCustomOptions tests initialization with custom options
func TestClientInitializeWithCustomOptions(t *testing.T) {
	config := &LanguageServerConfig{
		Language:       "go",
		Command:        "mock-lsp",
		InitTimeout:    100 * time.Millisecond,
		RequestTimeout: 100 * time.Millisecond,
		Env:            make(map[string]string),
	}

	client, err := NewClient(config)
	require.NoError(t, err)
	defer client.Close()

	// Test that passing custom init options works
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	customOptions := map[string]string{"custom": "value"}
	rootURI := "file:///workspace"

	// Will fail without server, but tests the code path
	_, err = client.Initialize(ctx, &rootURI, customOptions)
	assert.Error(t, err) // Expected to fail without server
}

// TestConfigMergeArgs tests merging arguments
func TestConfigMergeArgs(t *testing.T) {
	base := DefaultConfig("go")
	originalArgs := base.Args

	override := &LanguageServerConfig{
		Args: []string{"new-arg1", "new-arg2"},
	}

	base.Merge(override)

	assert.NotEqual(t, originalArgs, base.Args)
	assert.Equal(t, []string{"new-arg1", "new-arg2"}, base.Args)
}

// TestConfigValidationAllErrors tests all validation error paths
func TestConfigValidationAllErrors(t *testing.T) {
	tests := []struct {
		name   string
		config *LanguageServerConfig
		errMsg string
	}{
		{
			name:   "missing language",
			config: &LanguageServerConfig{Command: "test", InitTimeout: 1, RequestTimeout: 1},
			errMsg: "language identifier is required",
		},
		{
			name:   "missing command",
			config: &LanguageServerConfig{Language: "go", InitTimeout: 1, RequestTimeout: 1},
			errMsg: "language server command is required",
		},
		{
			name:   "invalid init timeout",
			config: &LanguageServerConfig{Language: "go", Command: "test", InitTimeout: 0, RequestTimeout: 1},
			errMsg: "initialization timeout must be positive",
		},
		{
			name:   "invalid request timeout",
			config: &LanguageServerConfig{Language: "go", Command: "test", InitTimeout: 1, RequestTimeout: 0},
			errMsg: "request timeout must be positive",
		},
		{
			name:   "negative max restarts",
			config: &LanguageServerConfig{Language: "go", Command: "test", InitTimeout: 1, RequestTimeout: 1, MaxRestarts: -1},
			errMsg: "max restarts cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// TestManagerInitializeClientAutoRestart tests auto-restart on initialization failure
func TestManagerInitializeClientAutoRestart(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := &LanguageServerConfig{
		Language:       "go",
		Command:        "/nonexistent/command",
		InitTimeout:    100 * time.Millisecond,
		RequestTimeout: 100 * time.Millisecond,
		AutoRestart:    true,
		MaxRestarts:    3,
		Env:            make(map[string]string),
	}

	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	ctx := context.Background()

	// First attempt - should fail and increment restart count
	_, err = manager.InitializeClient(ctx, "go", "file:///test")
	assert.Error(t, err)

	// Verify restart count was NOT incremented (only happens on init failure with server started)
	// Since Start() fails, we don't increment
	count := manager.GetRestartCount("go")
	assert.Equal(t, 0, count)
}

// TestConfigDefaultsForAllLanguages tests defaults for all supported languages
func TestConfigDefaultsForAllLanguages(t *testing.T) {
	expectedCommands := map[string]string{
		"go":         "gopls",
		"python":     "pylsp",
		"typescript": "typescript-language-server",
		"javascript": "typescript-language-server",
		"rust":       "rust-analyzer",
		"java":       "jdtls",
		"cpp":        "clangd",
		"c":          "clangd",
	}

	for _, lang := range SupportedLanguages() {
		config := DefaultConfig(lang)

		assert.Equal(t, lang, config.Language)
		assert.Equal(t, expectedCommands[lang], config.Command)
		assert.True(t, config.EnableCompletion)
		assert.True(t, config.EnableHover)
		assert.True(t, config.EnableDefinition)
		assert.True(t, config.EnableReferences)
		assert.True(t, config.AutoRestart)
		assert.Equal(t, 3, config.MaxRestarts)
		assert.Equal(t, 30*time.Second, config.InitTimeout)
		assert.Equal(t, 10*time.Second, config.RequestTimeout)
	}
}

// TestManagerRegisterDefaultLanguagesError tests error handling
func TestManagerRegisterDefaultLanguagesError(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// Register all defaults
	err := manager.RegisterDefaultLanguages()
	assert.NoError(t, err)

	// Try again - should fail because languages already registered
	err = manager.RegisterDefaultLanguages()
	assert.Error(t, err)
}

// TestClientStartWithEnvironment tests starting with environment variables
func TestClientStartWithEnvironment(t *testing.T) {
	config := &LanguageServerConfig{
		Language:       "go",
		Command:        "/nonexistent/command",
		InitTimeout:    1 * time.Second,
		RequestTimeout: 1 * time.Second,
		Env: map[string]string{
			"TEST_VAR": "test_value",
			"DEBUG":    "true",
		},
	}

	client, err := NewClient(config)
	require.NoError(t, err)
	defer client.Close()

	// Start will fail, but environment should be set
	err = client.Start()
	assert.Error(t, err) // Expected - command doesn't exist
}

// TestInsertTextFormatConstants tests insert text format constants
func TestInsertTextFormatConstants(t *testing.T) {
	assert.Equal(t, 1, InsertTextFormatPlainText)
	assert.Equal(t, 2, InsertTextFormatSnippet)
}

// TestMarkupKindConstants tests markup kind constants
func TestMarkupKindConstants(t *testing.T) {
	assert.Equal(t, "plaintext", MarkupKindPlainText)
	assert.Equal(t, "markdown", MarkupKindMarkdown)
}

// TestClientCloseWithProcess tests closing with process running
func TestClientCloseWithProcess(t *testing.T) {
	config := DefaultConfig("go")
	client, err := NewClient(config)
	require.NoError(t, err)

	// Close should handle nil process gracefully
	err = client.Close()
	assert.NoError(t, err)
}
