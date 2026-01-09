package cmd

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetAPIKey tests the API key retrieval logic
func TestGetAPIKey(t *testing.T) {
	// Save original env and restore after test
	originalEnv := make(map[string]string)
	envVars := []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "META_LLAMA_API_KEY", "GOOGLE_API_KEY", "AINATIVE_CODE_API_KEY"}
	for _, env := range envVars {
		originalEnv[env] = os.Getenv(env)
		os.Unsetenv(env)
	}
	defer func() {
		for env, val := range originalEnv {
			if val != "" {
				os.Setenv(env, val)
			}
		}
	}()

	// Reset viper
	viper.Reset()

	t.Run("provider-specific environment variable", func(t *testing.T) {
		os.Setenv("ANTHROPIC_API_KEY", "test-anthropic-key")
		defer os.Unsetenv("ANTHROPIC_API_KEY")

		key, err := getAPIKey("anthropic")
		require.NoError(t, err)
		assert.Equal(t, "test-anthropic-key", key)
	})

	t.Run("nested config key", func(t *testing.T) {
		viper.Reset()
		viper.Set("llm.openai.api_key", "test-openai-config-key")

		key, err := getAPIKey("openai")
		require.NoError(t, err)
		assert.Equal(t, "test-openai-config-key", key)
	})

	t.Run("generic api_key fallback", func(t *testing.T) {
		viper.Reset()
		viper.Set("api_key", "test-generic-key")

		key, err := getAPIKey("anthropic")
		require.NoError(t, err)
		assert.Equal(t, "test-generic-key", key)
	})

	t.Run("ollama no key needed", func(t *testing.T) {
		viper.Reset()

		key, err := getAPIKey("ollama")
		require.NoError(t, err)
		assert.Equal(t, "", key)
	})

	t.Run("missing api key error", func(t *testing.T) {
		viper.Reset()

		_, err := getAPIKey("openai")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no API key found")
	})
}

// TestInitializeProvider tests provider initialization
func TestInitializeProvider(t *testing.T) {
	// This test requires actual provider implementations, so we'll test the basic flow
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("unsupported provider", func(t *testing.T) {
		_, err := initializeProvider(ctx, "unsupported", "test-model")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported provider")
	})

	t.Run("ollama provider without API key", func(t *testing.T) {
		viper.Reset()
		viper.Set("llm.ollama.base_url", "http://localhost:11434")

		// This will fail if Ollama is not running, which is expected
		// We're just testing that the initialization logic works
		_, err := initializeProvider(ctx, "ollama", "llama2")
		// Error is OK here since Ollama might not be running
		// We're testing the initialization flow, not the actual connection
		if err != nil {
			t.Logf("Expected: Ollama initialization attempted (error: %v)", err)
		}
	})

	t.Run("provider with missing API key", func(t *testing.T) {
		viper.Reset()
		// Clear all env vars
		for _, env := range []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "AINATIVE_CODE_API_KEY"} {
			os.Unsetenv(env)
		}

		_, err := initializeProvider(ctx, "openai", "gpt-4")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no API key found")
	})
}

// TestProviderConfigurationFlow tests the full configuration flow
func TestProviderConfigurationFlow(t *testing.T) {
	// Reset environment
	viper.Reset()
	for _, env := range []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "AINATIVE_CODE_API_KEY"} {
		os.Unsetenv(env)
	}

	t.Run("full config with nested structure", func(t *testing.T) {
		viper.Reset()

		// Simulate a config file structure
		viper.Set("llm.default_provider", "anthropic")
		viper.Set("llm.anthropic.api_key", "test-claude-key")
		viper.Set("llm.anthropic.model", "claude-3-5-sonnet-20241022")

		// Test API key retrieval
		key, err := getAPIKey("anthropic")
		require.NoError(t, err)
		assert.Equal(t, "test-claude-key", key)

		// Test default provider
		defaultProvider := viper.GetString("llm.default_provider")
		assert.Equal(t, "anthropic", defaultProvider)
	})

	t.Run("flat config for backward compatibility", func(t *testing.T) {
		viper.Reset()

		// Simulate old flat config structure
		viper.Set("provider", "openai")
		viper.Set("model", "gpt-4")
		viper.Set("api_key", "test-openai-key")

		// Test API key retrieval (should use fallback)
		key, err := getAPIKey("openai")
		require.NoError(t, err)
		assert.Equal(t, "test-openai-key", key)

		// Test provider/model reading (from root.go GetProvider/GetModel)
		provider := viper.GetString("provider")
		model := viper.GetString("model")
		assert.Equal(t, "openai", provider)
		assert.Equal(t, "gpt-4", model)
	})
}
