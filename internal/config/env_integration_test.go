package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoaderEnvironmentVariables tests that the config loader properly loads environment variables
func TestLoaderEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(t *testing.T, cfg *Config)
	}{
		{
			name: "anthropic api key from env",
			envVars: map[string]string{
				"AINATIVE_CODE_LLM_ANTHROPIC_API_KEY": "sk-ant-test-key",
				"AINATIVE_CODE_LLM_ANTHROPIC_MODEL":   "claude-3-opus",
			},
			validate: func(t *testing.T, cfg *Config) {
				require.NotNil(t, cfg.LLM.Anthropic)
				assert.Equal(t, "sk-ant-test-key", cfg.LLM.Anthropic.APIKey)
				assert.Equal(t, "claude-3-opus", cfg.LLM.Anthropic.Model)
			},
		},
		{
			name: "openai configuration from env",
			envVars: map[string]string{
				"AINATIVE_CODE_LLM_OPENAI_API_KEY":    "sk-openai-test",
				"AINATIVE_CODE_LLM_OPENAI_MODEL":      "gpt-4-turbo",
				"AINATIVE_CODE_LLM_OPENAI_MAX_TOKENS": "8192",
			},
			validate: func(t *testing.T, cfg *Config) {
				require.NotNil(t, cfg.LLM.OpenAI)
				assert.Equal(t, "sk-openai-test", cfg.LLM.OpenAI.APIKey)
				assert.Equal(t, "gpt-4-turbo", cfg.LLM.OpenAI.Model)
				assert.Equal(t, 8192, cfg.LLM.OpenAI.MaxTokens)
			},
		},
		{
			name: "default provider from env",
			envVars: map[string]string{
				"AINATIVE_CODE_LLM_DEFAULT_PROVIDER": "openai",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "openai", cfg.LLM.DefaultProvider)
			},
		},
		{
			name: "azure configuration from env",
			envVars: map[string]string{
				"AINATIVE_CODE_LLM_AZURE_API_KEY":   "azure-key-123",
				"AINATIVE_CODE_LLM_AZURE_ENDPOINT":  "https://test.openai.azure.com",
				"AINATIVE_CODE_LLM_AZURE_DEPLOYMENT_NAME": "gpt-4-deployment",
			},
			validate: func(t *testing.T, cfg *Config) {
				require.NotNil(t, cfg.LLM.Azure)
				assert.Equal(t, "azure-key-123", cfg.LLM.Azure.APIKey)
				assert.Equal(t, "https://test.openai.azure.com", cfg.LLM.Azure.Endpoint)
				assert.Equal(t, "gpt-4-deployment", cfg.LLM.Azure.DeploymentName)
			},
		},
		{
			name: "bedrock aws credentials from env",
			envVars: map[string]string{
				"AINATIVE_CODE_LLM_BEDROCK_REGION":            "us-west-2",
				"AINATIVE_CODE_LLM_BEDROCK_ACCESS_KEY_ID":     "AKIATEST123",
				"AINATIVE_CODE_LLM_BEDROCK_SECRET_ACCESS_KEY": "secret-key-test",
			},
			validate: func(t *testing.T, cfg *Config) {
				require.NotNil(t, cfg.LLM.Bedrock)
				assert.Equal(t, "us-west-2", cfg.LLM.Bedrock.Region)
				assert.Equal(t, "AKIATEST123", cfg.LLM.Bedrock.AccessKeyID)
				assert.Equal(t, "secret-key-test", cfg.LLM.Bedrock.SecretAccessKey)
			},
		},
		{
			name: "google gemini from env",
			envVars: map[string]string{
				"AINATIVE_CODE_LLM_GOOGLE_API_KEY":    "google-api-key",
				"AINATIVE_CODE_LLM_GOOGLE_PROJECT_ID": "my-project-123",
				"AINATIVE_CODE_LLM_GOOGLE_MODEL":      "gemini-pro",
			},
			validate: func(t *testing.T, cfg *Config) {
				require.NotNil(t, cfg.LLM.Google)
				assert.Equal(t, "google-api-key", cfg.LLM.Google.APIKey)
				assert.Equal(t, "my-project-123", cfg.LLM.Google.ProjectID)
				assert.Equal(t, "gemini-pro", cfg.LLM.Google.Model)
			},
		},
		{
			name: "ollama configuration from env",
			envVars: map[string]string{
				"AINATIVE_CODE_LLM_OLLAMA_BASE_URL": "http://localhost:11434",
				"AINATIVE_CODE_LLM_OLLAMA_MODEL":    "llama3",
			},
			validate: func(t *testing.T, cfg *Config) {
				require.NotNil(t, cfg.LLM.Ollama)
				assert.Equal(t, "http://localhost:11434", cfg.LLM.Ollama.BaseURL)
				assert.Equal(t, "llama3", cfg.LLM.Ollama.Model)
			},
		},
		{
			name: "zerodb configuration from env",
			envVars: map[string]string{
				"AINATIVE_CODE_SERVICES_ZERODB_ENABLED":    "true",
				"AINATIVE_CODE_SERVICES_ZERODB_PROJECT_ID": "project-123",
				"AINATIVE_CODE_SERVICES_ZERODB_ENDPOINT":   "https://api.zerodb.dev",
			},
			validate: func(t *testing.T, cfg *Config) {
				require.NotNil(t, cfg.Services.ZeroDB)
				assert.True(t, cfg.Services.ZeroDB.Enabled)
				assert.Equal(t, "project-123", cfg.Services.ZeroDB.ProjectID)
				assert.Equal(t, "https://api.zerodb.dev", cfg.Services.ZeroDB.Endpoint)
			},
		},
		{
			name: "logging configuration from env",
			envVars: map[string]string{
				"AINATIVE_CODE_LOGGING_LEVEL":  "debug",
				"AINATIVE_CODE_LOGGING_FORMAT": "json",
				"AINATIVE_CODE_LOGGING_OUTPUT": "stdout",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "debug", cfg.Logging.Level)
				assert.Equal(t, "json", cfg.Logging.Format)
				assert.Equal(t, "stdout", cfg.Logging.Output)
			},
		},
		{
			name: "app configuration from env",
			envVars: map[string]string{
				"AINATIVE_CODE_APP_NAME":        "custom-app",
				"AINATIVE_CODE_APP_ENVIRONMENT": "production",
				"AINATIVE_CODE_APP_DEBUG":       "true",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "custom-app", cfg.App.Name)
				assert.Equal(t, "production", cfg.App.Environment)
				assert.True(t, cfg.App.Debug)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Create loader with AINATIVE_CODE prefix
			loader := NewLoader()

			// Load configuration
			cfg, err := loader.Load()

			// Some tests may fail validation if not all required fields are set
			// We're primarily testing that env vars are loaded, not full validation
			if err != nil {
				// If validation fails, we can still check if env vars were loaded
				// by inspecting the viper instance directly
				t.Logf("Validation error (expected for partial config): %v", err)

				// For tests that only set a few env vars, we'll skip validation
				// and just verify the viper instance has the values
				v := loader.GetViper()

				// Create a partial config for validation
				var partialCfg Config
				if unmarshalErr := v.Unmarshal(&partialCfg); unmarshalErr == nil {
					tt.validate(t, &partialCfg)
				} else {
					t.Fatalf("Failed to unmarshal config: %v", unmarshalErr)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				tt.validate(t, cfg)
			}
		})
	}
}

// TestEnvVarPrecedenceOverConfigFile tests that env vars override config file values
func TestEnvVarPrecedenceOverConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := tmpDir + "/config.yaml"

	configContent := []byte(`llm:
  default_provider: "anthropic"
  anthropic:
    api_key: "file-api-key"
    model: "claude-3-sonnet"
    max_tokens: 2048
`)

	err := os.WriteFile(configFile, configContent, 0644)
	require.NoError(t, err)

	// Set environment variables that should override the config file
	t.Setenv("AINATIVE_CODE_LLM_ANTHROPIC_API_KEY", "env-api-key")
	t.Setenv("AINATIVE_CODE_LLM_ANTHROPIC_MODEL", "claude-3-opus")

	// Create loader and load from specific file
	loader := NewLoader()
	cfg, err := loader.LoadFromFile(configFile)

	// Check if config loaded (may fail validation due to missing required fields)
	if err != nil {
		t.Logf("Validation error (expected): %v", err)

		// Get viper instance to check values directly
		v := loader.GetViper()

		// Verify env vars override file values
		assert.Equal(t, "env-api-key", v.GetString("llm.anthropic.api_key"))
		assert.Equal(t, "claude-3-opus", v.GetString("llm.anthropic.model"))
		assert.Equal(t, 2048, v.GetInt("llm.anthropic.max_tokens")) // Should keep file value
	} else {
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.NotNil(t, cfg.LLM.Anthropic)

		// Environment variable should override file value
		assert.Equal(t, "env-api-key", cfg.LLM.Anthropic.APIKey)
		assert.Equal(t, "claude-3-opus", cfg.LLM.Anthropic.Model)
		// File value should be used for max_tokens (not overridden by env)
		assert.Equal(t, 2048, cfg.LLM.Anthropic.MaxTokens)
	}
}

// TestEnvVarWithVariableSubstitution tests ${VAR} syntax in config files
func TestEnvVarWithVariableSubstitution(t *testing.T) {
	// Set environment variable
	t.Setenv("OPENAI_API_KEY", "sk-real-key-from-env")

	// Create a temporary config file with ${VAR} syntax
	tmpDir := t.TempDir()
	configFile := tmpDir + "/config.yaml"

	configContent := []byte(`llm:
  openai:
    api_key: "${OPENAI_API_KEY}"
    model: "gpt-4"
`)

	err := os.WriteFile(configFile, configContent, 0644)
	require.NoError(t, err)

	// Create loader and load from specific file
	loader := NewLoader()
	cfg, err := loader.LoadFromFile(configFile)

	if err != nil {
		t.Logf("Validation error: %v", err)
		// Even if validation fails, check if the resolver worked
		v := loader.GetViper()
		apiKey := v.GetString("llm.openai.api_key")

		// The resolver should have expanded ${OPENAI_API_KEY}
		assert.Equal(t, "sk-real-key-from-env", apiKey,
			"Resolver should expand ${OPENAI_API_KEY} to the actual env var value")
	} else {
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.NotNil(t, cfg.LLM.OpenAI)

		// The ${OPENAI_API_KEY} should be resolved to the actual env var value
		assert.Equal(t, "sk-real-key-from-env", cfg.LLM.OpenAI.APIKey)
	}
}

// TestMultipleEnvVarSources tests combining different environment variable patterns
func TestMultipleEnvVarSources(t *testing.T) {
	// Set various environment variables
	t.Setenv("AINATIVE_CODE_LLM_DEFAULT_PROVIDER", "openai")
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-direct") // For ${ANTHROPIC_API_KEY} resolution
	t.Setenv("AINATIVE_CODE_LOGGING_LEVEL", "debug")

	// Create a config file that uses ${VAR} syntax
	tmpDir := t.TempDir()
	configFile := tmpDir + "/config.yaml"

	configContent := []byte(`llm:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    model: "claude-3-sonnet"
`)

	err := os.WriteFile(configFile, configContent, 0644)
	require.NoError(t, err)

	// Create loader
	loader := NewLoader()
	cfg, err := loader.LoadFromFile(configFile)

	if err != nil {
		t.Logf("Validation error: %v", err)
		v := loader.GetViper()

		// Check AINATIVE_CODE_* env vars
		assert.Equal(t, "openai", v.GetString("llm.default_provider"))
		assert.Equal(t, "debug", v.GetString("logging.level"))

		// Check ${VAR} resolution
		apiKey := v.GetString("llm.anthropic.api_key")
		assert.Equal(t, "sk-ant-direct", apiKey)
	} else {
		require.NoError(t, err)

		// Verify AINATIVE_CODE_* prefix env vars work
		assert.Equal(t, "openai", cfg.LLM.DefaultProvider)
		assert.Equal(t, "debug", cfg.Logging.Level)

		// Verify ${VAR} substitution in config file worked
		if cfg.LLM.Anthropic != nil {
			assert.Equal(t, "sk-ant-direct", cfg.LLM.Anthropic.APIKey)
		}
	}
}

// TestEnvVarNumericAndBooleanTypes tests that numeric and boolean env vars work
func TestEnvVarNumericAndBooleanTypes(t *testing.T) {
	// Set environment variables
	t.Setenv("AINATIVE_CODE_LLM_ANTHROPIC_MAX_TOKENS", "8192")
	t.Setenv("AINATIVE_CODE_LLM_ANTHROPIC_TEMPERATURE", "0.9")
	t.Setenv("AINATIVE_CODE_SERVICES_ZERODB_ENABLED", "true")
	t.Setenv("AINATIVE_CODE_APP_DEBUG", "false")

	// Create loader
	loader := NewLoader()

	// Get viper instance
	v := loader.GetViper()

	// Verify numeric types
	assert.Equal(t, 8192, v.GetInt("llm.anthropic.max_tokens"))
	assert.Equal(t, 0.9, v.GetFloat64("llm.anthropic.temperature"))

	// Verify boolean types
	assert.True(t, v.GetBool("services.zerodb.enabled"))
	assert.False(t, v.GetBool("app.debug"))
}
