package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/AINative-studio/ainative-code/internal/config"
)

// TestIssue96_SetupToChatFlow tests the complete flow from setup wizard to chat command
// This reproduces and validates the fix for issue #96
func TestIssue96_SetupToChatFlow(t *testing.T) {
	// Create a temporary directory for test configs
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".ainative-code.yaml")

	t.Run("setup wizard creates nested config and chat reads it correctly", func(t *testing.T) {
		// Step 1: Simulate setup wizard creating config file
		// This is what the setup wizard in wizard.go does
		setupConfig := &config.Config{
			App: config.AppConfig{
				Name:        "ainative-code",
				Version:     "0.1.7",
				Environment: "development",
				Debug:       false,
			},
			LLM: config.LLMConfig{
				DefaultProvider: "anthropic",
				Anthropic: &config.AnthropicConfig{
					APIKey:      "sk-ant-test-key-123",
					Model:       "claude-3-5-sonnet-20241022",
					MaxTokens:   4096,
					Temperature: 0.7,
				},
			},
		}

		// Write config to file (simulating setup wizard)
		data, err := yaml.Marshal(setupConfig)
		require.NoError(t, err)
		err = os.WriteFile(configPath, data, 0600)
		require.NoError(t, err)

		// Step 2: Simulate chat command loading the config
		// Reset viper to ensure clean state
		viper.Reset()

		// Load the config file
		viper.SetConfigFile(configPath)
		err = viper.ReadInConfig()
		require.NoError(t, err, "Chat command should be able to read config file")

		// Step 3: Test GetProvider() reads from nested config
		providerName := GetProvider()
		assert.Equal(t, "anthropic", providerName, "GetProvider() should read from llm.default_provider")

		// Step 4: Test GetModel() reads from nested config
		modelName := GetModel()
		assert.Equal(t, "claude-3-5-sonnet-20241022", modelName, "GetModel() should read from llm.anthropic.model")

		// Step 5: Test getAPIKey() reads from nested config
		apiKey, err := getAPIKey("anthropic")
		require.NoError(t, err, "getAPIKey() should find key in nested config")
		assert.Equal(t, "sk-ant-test-key-123", apiKey)
	})

	t.Run("backward compatibility with flat config", func(t *testing.T) {
		// Create a flat config (old format)
		flatConfigPath := filepath.Join(tmpDir, "flat-config.yaml")
		flatConfig := map[string]interface{}{
			"provider": "openai",
			"model":    "gpt-4",
			"api_key":  "sk-openai-test-key",
		}

		data, err := yaml.Marshal(flatConfig)
		require.NoError(t, err)
		err = os.WriteFile(flatConfigPath, data, 0600)
		require.NoError(t, err)

		// Reset and load flat config
		viper.Reset()
		viper.SetConfigFile(flatConfigPath)
		err = viper.ReadInConfig()
		require.NoError(t, err)

		// Test backward compatibility
		providerName := GetProvider()
		assert.Equal(t, "openai", providerName, "Should read from flat 'provider' field")

		modelName := GetModel()
		assert.Equal(t, "gpt-4", modelName, "Should read from flat 'model' field")

		apiKey, err := getAPIKey("openai")
		require.NoError(t, err)
		assert.Equal(t, "sk-openai-test-key", apiKey, "Should read from flat 'api_key' field")
	})

	t.Run("nested config takes priority over flat config", func(t *testing.T) {
		// Create a config with both nested and flat fields
		mixedConfigPath := filepath.Join(tmpDir, "mixed-config.yaml")

		viper.Reset()
		viper.SetConfigFile(mixedConfigPath)

		// Manually set both nested and flat values
		viper.Set("llm.default_provider", "anthropic")
		viper.Set("llm.anthropic.model", "claude-3-5-sonnet-20241022")
		viper.Set("llm.anthropic.api_key", "sk-ant-nested-key")
		viper.Set("provider", "openai") // This should be ignored
		viper.Set("model", "gpt-3.5")    // This should be ignored
		viper.Set("api_key", "sk-flat-key") // This should be ignored

		// Test that nested config takes priority
		providerName := GetProvider()
		assert.Equal(t, "anthropic", providerName, "Nested config should take priority")

		modelName := GetModel()
		assert.Equal(t, "claude-3-5-sonnet-20241022", modelName, "Nested model should take priority")

		apiKey, err := getAPIKey("anthropic")
		require.NoError(t, err)
		assert.Equal(t, "sk-ant-nested-key", apiKey, "Nested API key should take priority")
	})

	t.Run("all supported providers work with nested config", func(t *testing.T) {
		providers := []struct {
			name      string
			configKey string
			apiKey    string
			model     string
		}{
			{"anthropic", "llm.anthropic", "sk-ant-test", "claude-3-5-sonnet-20241022"},
			{"openai", "llm.openai", "sk-openai-test", "gpt-4"},
			{"google", "llm.google", "google-api-key", "gemini-pro"},
			{"meta_llama", "llm.meta_llama", "meta-api-key", "llama-4"},
		}

		for _, tc := range providers {
			t.Run(tc.name, func(t *testing.T) {
				viper.Reset()
				viper.Set("llm.default_provider", tc.name)
				viper.Set(tc.configKey+".api_key", tc.apiKey)
				viper.Set(tc.configKey+".model", tc.model)

				providerName := GetProvider()
				assert.Equal(t, tc.name, providerName)

				modelName := GetModel()
				assert.Equal(t, tc.model, modelName)

				apiKey, err := getAPIKey(tc.name)
				require.NoError(t, err)
				assert.Equal(t, tc.apiKey, apiKey)
			})
		}
	})

	t.Run("error when provider configured but no API key", func(t *testing.T) {
		viper.Reset()
		viper.Set("llm.default_provider", "anthropic")
		viper.Set("llm.anthropic.model", "claude-3-5-sonnet-20241022")
		// No API key set

		providerName := GetProvider()
		assert.Equal(t, "anthropic", providerName, "Provider should be configured")

		_, err := getAPIKey("anthropic")
		require.Error(t, err, "Should error when API key is missing")
		assert.Contains(t, err.Error(), "no API key found")
	})

	t.Run("environment variables take precedence over config", func(t *testing.T) {
		// Set environment variable
		os.Setenv("ANTHROPIC_API_KEY", "sk-ant-from-env")
		defer os.Unsetenv("ANTHROPIC_API_KEY")

		viper.Reset()
		viper.Set("llm.default_provider", "anthropic")
		viper.Set("llm.anthropic.api_key", "sk-ant-from-config")

		apiKey, err := getAPIKey("anthropic")
		require.NoError(t, err)
		assert.Equal(t, "sk-ant-from-env", apiKey, "Environment variable should take precedence")
	})
}

// TestIssue96_RootCause tests the specific root cause that was fixed
func TestIssue96_RootCause(t *testing.T) {
	t.Run("GetProvider reads from llm.default_provider", func(t *testing.T) {
		viper.Reset()
		viper.Set("llm.default_provider", "anthropic")

		provider := GetProvider()
		assert.Equal(t, "anthropic", provider, "GetProvider() must read from llm.default_provider")
	})

	t.Run("GetModel reads from provider-specific config", func(t *testing.T) {
		viper.Reset()
		viper.Set("llm.default_provider", "anthropic")
		viper.Set("llm.anthropic.model", "claude-3-5-sonnet-20241022")

		model := GetModel()
		assert.Equal(t, "claude-3-5-sonnet-20241022", model, "GetModel() must read from llm.anthropic.model")
	})

	t.Run("getAPIKey reads from nested config", func(t *testing.T) {
		viper.Reset()
		viper.Set("llm.anthropic.api_key", "sk-ant-nested")

		apiKey, err := getAPIKey("anthropic")
		require.NoError(t, err)
		assert.Equal(t, "sk-ant-nested", apiKey, "getAPIKey() must read from llm.anthropic.api_key")
	})
}

// TestIssue96_ChatCommandFlow tests the actual chat command flow
func TestIssue96_ChatCommandFlow(t *testing.T) {
	// Skip if no API key available (this is an integration test)
	if os.Getenv("ANTHROPIC_API_KEY") == "" && os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration test - no ANTHROPIC_API_KEY set")
	}

	t.Run("chat command can read provider after setup", func(t *testing.T) {
		viper.Reset()

		// Simulate setup wizard output
		viper.Set("llm.default_provider", "anthropic")
		viper.Set("llm.anthropic.model", "claude-3-5-sonnet-20241022")

		// Use environment variable for API key (safer for tests)
		if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
			viper.Set("llm.anthropic.api_key", apiKey)
		}

		// This is what chat.go does
		providerName := GetProvider()
		modelName := GetModel()

		assert.NotEmpty(t, providerName, "Provider should be configured")
		assert.NotEmpty(t, modelName, "Model should be configured")

		// Verify we can get API key
		if os.Getenv("ANTHROPIC_API_KEY") != "" {
			apiKey, err := getAPIKey(providerName)
			require.NoError(t, err, "Should be able to get API key")
			assert.NotEmpty(t, apiKey)
		}
	})
}
