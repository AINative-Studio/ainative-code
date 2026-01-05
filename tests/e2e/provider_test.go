package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMultiProviderSwitching tests switching between different AI providers
func TestMultiProviderSwitching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("switch from OpenAI to Anthropic", func(t *testing.T) {
		// Initialize with OpenAI
		h.RunCommand("config", "init")
		result := h.RunCommand("config", "get", "provider")
		h.AssertSuccess(result, "should get default provider")

		// Switch to Anthropic
		result = h.RunCommand("config", "set", "provider", "anthropic")
		h.AssertSuccess(result, "should set anthropic provider")
		h.AssertStdoutContains(result, "anthropic", "should confirm anthropic")

		// Verify the switch
		result = h.RunCommand("config", "get", "provider")
		h.AssertSuccess(result, "should get updated provider")
		h.AssertStdoutContains(result, "anthropic", "should show anthropic")
	})

	t.Run("switch from Anthropic to Ollama", func(t *testing.T) {
		h.RunCommand("config", "init")

		// Set to Anthropic first
		h.RunCommand("config", "set", "provider", "anthropic")

		// Switch to Ollama
		result := h.RunCommand("config", "set", "provider", "ollama")
		h.AssertSuccess(result, "should set ollama provider")

		// Verify
		result = h.RunCommand("config", "get", "provider")
		h.AssertStdoutContains(result, "ollama", "should show ollama")
	})

	t.Run("provider flag overrides config file", func(t *testing.T) {
		// Set config to OpenAI
		h.RunCommand("config", "init")
		h.RunCommand("config", "set", "provider", "openai")

		// Use Anthropic via flag
		result := h.RunCommand("--provider", "anthropic", "chat", "test")
		h.AssertSuccess(result, "provider flag should work")

		// Verify config is unchanged
		result = h.RunCommand("config", "get", "provider")
		h.AssertStdoutContains(result, "openai", "config should remain openai")
	})

	t.Run("model flag works with provider switch", func(t *testing.T) {
		h.RunCommand("config", "init")

		result := h.RunCommand("--provider", "anthropic", "--model", "claude-3-opus", "chat", "test")
		h.AssertSuccess(result, "provider and model flags should work together")
	})
}

// TestProviderConfiguration tests provider-specific configuration
func TestProviderConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	providers := []string{"openai", "anthropic", "ollama"}

	for _, provider := range providers {
		t.Run("configure "+provider, func(t *testing.T) {
			result := h.RunCommand("config", "set", "provider", provider)
			h.AssertSuccess(result, "should set %s provider", provider)

			// Validate configuration
			result = h.RunCommand("config", "validate")
			h.AssertSuccess(result, "%s config should be valid", provider)
			h.AssertStdoutContains(result, provider, "should show provider in validation")
		})
	}
}

// TestProviderValidation tests provider validation
func TestProviderValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("valid providers pass validation", func(t *testing.T) {
		validProviders := []string{"openai", "anthropic", "ollama"}
		for _, provider := range validProviders {
			h.RunCommand("config", "set", "provider", provider)
			result := h.RunCommand("config", "validate")
			h.AssertSuccess(result, "%s should be valid", provider)
		}
	})

	t.Run("invalid provider fails validation", func(t *testing.T) {
		h.RunCommand("config", "set", "provider", "invalid-provider")
		result := h.RunCommand("config", "validate")
		h.AssertFailure(result, "invalid provider should fail validation")
		h.AssertStderrContains(result, "invalid provider", "should show error")
	})
}

// TestProviderSwitchingWorkflow tests complete provider switching workflow
func TestProviderSwitchingWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("complete provider switching workflow", func(t *testing.T) {
		// Initialize with default (OpenAI)
		result := h.RunCommand("config", "init")
		h.AssertSuccess(result, "init should succeed")

		// Use OpenAI
		result = h.RunCommand("chat", "test with openai")
		h.AssertSuccess(result, "chat with openai should work")

		// Switch to Anthropic
		result = h.RunCommand("config", "set", "provider", "anthropic")
		h.AssertSuccess(result, "switch to anthropic should work")

		// Validate new config
		result = h.RunCommand("config", "validate")
		h.AssertSuccess(result, "anthropic config should be valid")

		// Use Anthropic
		result = h.RunCommand("chat", "test with anthropic")
		h.AssertSuccess(result, "chat with anthropic should work")

		// Switch to Ollama
		result = h.RunCommand("config", "set", "provider", "ollama")
		h.AssertSuccess(result, "switch to ollama should work")

		// Use Ollama
		result = h.RunCommand("chat", "test with ollama")
		h.AssertSuccess(result, "chat with ollama should work")
	})
}

// TestProviderEnvironmentConfig tests environment variable configuration
func TestProviderEnvironmentConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize config first to avoid provider not configured error
	h.RunCommand("config", "init")

	t.Run("provider can be set via environment", func(t *testing.T) {
		env := map[string]string{
			"AINATIVE_PROVIDER": "anthropic",
		}
		result := h.RunCommandWithEnv(env, "chat", "test")
		h.AssertSuccess(result, "chat with env provider should work")
	})

	t.Run("model can be set via environment", func(t *testing.T) {
		env := map[string]string{
			"AINATIVE_PROVIDER": "openai",
			"AINATIVE_MODEL":    "gpt-4",
		}
		result := h.RunCommandWithEnv(env, "chat", "test")
		h.AssertSuccess(result, "chat with env model should work")
	})
}

// TestProviderFlagOverrides tests flag precedence
func TestProviderFlagOverrides(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("flag overrides config file", func(t *testing.T) {
		// Set config to openai
		h.RunCommand("config", "set", "provider", "openai")

		// Use anthropic via flag
		result := h.RunCommand("--provider", "anthropic", "chat", "test")
		h.AssertSuccess(result, "flag should override config")
	})

	t.Run("flag overrides environment", func(t *testing.T) {
		env := map[string]string{
			"AINATIVE_PROVIDER": "openai",
		}
		// Use anthropic via flag
		result := h.RunCommandWithEnv(env, "--provider", "anthropic", "chat", "test")
		h.AssertSuccess(result, "flag should override environment")
	})
}

// TestProviderErrorMessages tests provider-specific error messages
func TestProviderErrorMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("missing provider shows helpful error", func(t *testing.T) {
		result := h.RunCommand("chat", "test")
		h.AssertFailure(result, "should fail without provider")
		h.AssertStderrContains(result, "provider", "error should mention provider")
	})
}

// TestProviderConfigPersistence tests that provider config persists
func TestProviderConfigPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("provider config persists across commands", func(t *testing.T) {
		// Set provider
		h.RunCommand("config", "init")
		h.RunCommand("config", "set", "provider", "anthropic")

		// Verify it persists
		result := h.RunCommand("config", "get", "provider")
		h.AssertStdoutContains(result, "anthropic", "provider should persist")

		// Use in another command
		result = h.RunCommand("chat", "test")
		h.AssertSuccess(result, "provider should be available")

		// Verify still persists
		result = h.RunCommand("config", "get", "provider")
		h.AssertStdoutContains(result, "anthropic", "provider should still be set")
	})
}

// TestProviderWithCustomEndpoints tests custom endpoint configuration
func TestProviderWithCustomEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("ollama with custom endpoint", func(t *testing.T) {
		h.RunCommand("config", "set", "provider", "ollama")
		h.RunCommand("config", "set", "ollama.endpoint", "http://localhost:11434")

		result := h.RunCommand("chat", "test")
		h.AssertSuccess(result, "ollama with custom endpoint should work")
	})
}

// TestProviderWithSessionContext tests provider with session
func TestProviderWithSessionContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("provider persists within session", func(t *testing.T) {
		h.RunCommand("config", "set", "provider", "anthropic")

		// Create session with provider
		result := h.RunCommand("chat", "-s", "provider-session", "test")
		h.AssertSuccess(result, "chat with session should work")

		// Continue session
		result = h.RunCommand("chat", "-s", "provider-session", "another message")
		h.AssertSuccess(result, "session continuation should work")
	})
}

// TestProviderPerformance tests provider performance characteristics
func TestProviderPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()
	h.SetTimeout(60 * 1000000000) // 60 seconds for performance tests

	h.RunCommand("config", "init")

	t.Run("provider switching is fast", func(t *testing.T) {
		providers := []string{"openai", "anthropic", "ollama"}
		for _, provider := range providers {
			result := h.RunCommand("config", "set", "provider", provider)
			h.AssertSuccess(result, "switching to %s should be fast", provider)
			assert.Less(t, result.Duration.Milliseconds(), int64(1000),
				"provider switch should complete in < 1s")
		}
	})
}
