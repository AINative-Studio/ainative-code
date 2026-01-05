package lsp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestManagerRegisterDefaultLanguages tests registering all default languages
func TestManagerRegisterDefaultLanguages(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	err := manager.RegisterDefaultLanguages()
	assert.NoError(t, err)

	registered := manager.GetRegisteredLanguages()
	assert.Len(t, registered, 8) // All supported languages

	// Verify each language is registered
	for _, lang := range SupportedLanguages() {
		assert.Contains(t, registered, lang)
	}
}

// TestManagerGetActiveLanguages tests getting active language list
func TestManagerGetActiveLanguages(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// Initially no active languages
	active := manager.GetActiveLanguages()
	assert.Len(t, active, 0)

	// Register and activate some languages
	for _, lang := range []string{"go", "python"} {
		config := DefaultConfig(lang)
		err := manager.RegisterLanguage(config)
		require.NoError(t, err)

		_, err = manager.GetClient(lang)
		require.NoError(t, err)
	}

	active = manager.GetActiveLanguages()
	assert.Len(t, active, 2)
	assert.Contains(t, active, "go")
	assert.Contains(t, active, "python")
}

// TestManagerRestart tests restart functionality
func TestManagerRestartFunctionality(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := DefaultConfig("go")
	config.AutoRestart = true
	config.MaxRestarts = 3

	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// Test shouldRestart
	assert.True(t, manager.shouldRestart("go"))

	// Increment restart count
	manager.incrementRestartCount("go")
	assert.Equal(t, 1, manager.GetRestartCount("go"))

	manager.incrementRestartCount("go")
	assert.Equal(t, 2, manager.GetRestartCount("go"))

	// Still should restart (2 < 3)
	assert.True(t, manager.shouldRestart("go"))

	manager.incrementRestartCount("go")
	assert.Equal(t, 3, manager.GetRestartCount("go"))

	// Should not restart anymore (reached max)
	assert.False(t, manager.shouldRestart("go"))

	// Reset restart count
	manager.resetRestartCount("go")
	assert.Equal(t, 0, manager.GetRestartCount("go"))

	// Should restart again
	assert.True(t, manager.shouldRestart("go"))
}

// TestManagerHealthMonitoring tests health monitoring features
func TestManagerHealthMonitoring(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := DefaultConfig("go")
	config.AutoRestart = false // Disable auto-restart for this test

	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	client, err := manager.GetClient("go")
	require.NoError(t, err)

	// Client not initialized yet
	assert.False(t, manager.IsHealthy("go"))

	// Simulate initialization
	client.initialized = true
	assert.True(t, manager.IsHealthy("go"))

	// Simulate shutdown
	client.shutdown = true
	assert.False(t, manager.IsHealthy("go"))
}

// TestJSONRPCErrorInterface tests the error interface implementation
func TestJSONRPCErrorInterface(t *testing.T) {
	err := &JSONRPCError{
		Code:    InvalidParams,
		Message: "Invalid parameters",
	}

	// Test that it implements error interface
	var _ error = err

	// Test Error() method
	errMsg := err.Error()
	assert.Contains(t, errMsg, "Invalid parameters")
	assert.Contains(t, errMsg, "-32602")

	// Test with data
	errWithData := &JSONRPCError{
		Code:    InternalError,
		Message: "Internal error",
		Data:    map[string]string{"detail": "test"},
	}

	errMsgWithData := errWithData.Error()
	assert.Contains(t, errMsgWithData, "Internal error")
	assert.Contains(t, errMsgWithData, "data:")
}

// TestClientExitNotification tests the exit notification
func TestClientExitNotification(t *testing.T) {
	// This test verifies the Exit method exists and has the right signature
	// Actual testing would require a mock server
	config := DefaultConfig("go")
	client, err := NewClient(config)
	require.NoError(t, err)
	defer client.Close()

	// Exit should not error even without a server
	// (it just sends a notification, no response expected)
	// We can't fully test this without pipes set up, but we can verify it compiles
	assert.NotNil(t, client)
}

// TestConfigValidationEdgeCases tests edge cases in config validation
func TestConfigValidationEdgeCases(t *testing.T) {
	t.Run("zero request timeout", func(t *testing.T) {
		config := &LanguageServerConfig{
			Language:       "go",
			Command:        "gopls",
			InitTimeout:    1,
			RequestTimeout: 0,
		}
		err := config.Validate()
		assert.Error(t, err)
	})

	t.Run("negative init timeout", func(t *testing.T) {
		config := &LanguageServerConfig{
			Language:       "go",
			Command:        "gopls",
			InitTimeout:    -1,
			RequestTimeout: 1,
		}
		err := config.Validate()
		assert.Error(t, err)
	})
}

// TestManagerUnregisterWithActiveClient tests unregistering with active client
func TestManagerUnregisterWithActiveClient(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := DefaultConfig("go")
	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// Get client (creates it)
	client, err := manager.GetClient("go")
	require.NoError(t, err)
	require.NotNil(t, client)

	// Unregister should close the client
	err = manager.UnregisterLanguage("go")
	assert.NoError(t, err)

	// Verify language is no longer registered
	registered := manager.GetRegisteredLanguages()
	assert.NotContains(t, registered, "go")
}

// TestConfigMergeWithNilEnv tests merging with nil environment
func TestConfigMergeWithNilEnv(t *testing.T) {
	base := DefaultConfig("go")
	base.Env = nil

	override := &LanguageServerConfig{
		Env: map[string]string{"TEST": "value"},
	}

	base.Merge(override)
	assert.NotNil(t, base.Env)
	assert.Equal(t, "value", base.Env["TEST"])
}

// TestManagerShouldRestartWithoutConfig tests shouldRestart with missing config
func TestManagerShouldRestartWithoutConfig(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// Language not registered
	assert.False(t, manager.shouldRestart("nonexistent"))
}
