package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvironmentVariableSupport tests that environment variables are properly bound
func TestEnvironmentVariableSupport(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected map[string]interface{}
	}{
		{
			name: "basic provider and model",
			envVars: map[string]string{
				"AINATIVE_CODE_PROVIDER": "openai",
				"AINATIVE_CODE_MODEL":    "gpt-4",
			},
			expected: map[string]interface{}{
				"provider": "openai",
				"model":    "gpt-4",
			},
		},
		{
			name: "api key",
			envVars: map[string]string{
				"AINATIVE_CODE_API_KEY": "sk-test-key-123",
			},
			expected: map[string]interface{}{
				"api_key": "sk-test-key-123",
			},
		},
		{
			name: "verbose flag",
			envVars: map[string]string{
				"AINATIVE_CODE_VERBOSE": "true",
			},
			expected: map[string]interface{}{
				"verbose": true,
			},
		},
		{
			name: "multiple config values",
			envVars: map[string]string{
				"AINATIVE_CODE_PROVIDER": "anthropic",
				"AINATIVE_CODE_MODEL":    "claude-3-5-sonnet-20241022",
				"AINATIVE_CODE_VERBOSE":  "false",
			},
			expected: map[string]interface{}{
				"provider": "anthropic",
				"model":    "claude-3-5-sonnet-20241022",
				"verbose":  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper for each test
			v := viper.New()

			// Set environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Configure viper with AINATIVE_CODE prefix
			v.SetEnvPrefix("AINATIVE_CODE")
			v.AutomaticEnv()

			// Bind environment variables
			for key := range tt.expected {
				err := v.BindEnv(key)
				require.NoError(t, err)
			}

			// Verify values
			for key, expectedValue := range tt.expected {
				var actualValue interface{}
				// Use appropriate getter based on expected type
				switch expectedValue.(type) {
				case bool:
					actualValue = v.GetBool(key)
				case int:
					actualValue = v.GetInt(key)
				case string:
					actualValue = v.GetString(key)
				default:
					actualValue = v.Get(key)
				}
				assert.Equal(t, expectedValue, actualValue, "key: %s", key)
			}
		})
	}
}

// TestEnvironmentVariableKeyReplacer tests underscore replacement for dotted keys
func TestEnvironmentVariableKeyReplacer(t *testing.T) {
	tests := []struct {
		name    string
		envVar  string
		envVal  string
		viperKey string
		expected interface{}
	}{
		{
			name:     "dotted key - llm.anthropic.api_key",
			envVar:   "AINATIVE_CODE_LLM_ANTHROPIC_API_KEY",
			envVal:   "sk-ant-test",
			viperKey: "llm.anthropic.api_key",
			expected: "sk-ant-test",
		},
		{
			name:     "dotted key - llm.openai.model",
			envVar:   "AINATIVE_CODE_LLM_OPENAI_MODEL",
			envVal:   "gpt-4-turbo",
			viperKey: "llm.openai.model",
			expected: "gpt-4-turbo",
		},
		{
			name:     "dashed key - max-tokens",
			envVar:   "AINATIVE_CODE_MAX_TOKENS",
			envVal:   "4096",
			viperKey: "max-tokens",
			expected: "4096",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper for each test
			v := viper.New()

			// Set environment variable
			t.Setenv(tt.envVar, tt.envVal)

			// Configure viper with AINATIVE_CODE prefix and key replacer
			v.SetEnvPrefix("AINATIVE_CODE")
			v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
			v.AutomaticEnv()

			// Bind the key
			err := v.BindEnv(tt.viperKey)
			require.NoError(t, err)

			// Verify value
			actualValue := v.Get(tt.viperKey)
			assert.Equal(t, tt.expected, actualValue)
		})
	}
}

// TestConfigPrecedence tests that flags > env vars > config file precedence is maintained
func TestConfigPrecedence(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := tmpDir + "/test-config.yaml"

	configContent := []byte(`provider: "config-file-provider"
model: "config-file-model"
verbose: false
`)
	err := os.WriteFile(configFile, configContent, 0644)
	require.NoError(t, err)

	tests := []struct {
		name          string
		configFile    string
		envVars       map[string]string
		flagValues    map[string]string
		expectedKey   string
		expectedValue string
		description   string
	}{
		{
			name:          "env var overrides config file",
			configFile:    configFile,
			envVars:       map[string]string{"AINATIVE_CODE_PROVIDER": "env-provider"},
			flagValues:    map[string]string{},
			expectedKey:   "provider",
			expectedValue: "env-provider",
			description:   "Environment variable should override config file value",
		},
		{
			name:          "flag overrides env var",
			configFile:    configFile,
			envVars:       map[string]string{"AINATIVE_CODE_MODEL": "env-model"},
			flagValues:    map[string]string{"model": "flag-model"},
			expectedKey:   "model",
			expectedValue: "flag-model",
			description:   "Flag should override environment variable",
		},
		{
			name:          "flag overrides config file",
			configFile:    configFile,
			envVars:       map[string]string{},
			flagValues:    map[string]string{"provider": "flag-provider"},
			expectedKey:   "provider",
			expectedValue: "flag-provider",
			description:   "Flag should override config file value",
		},
		{
			name:          "config file used when no env var or flag",
			configFile:    configFile,
			envVars:       map[string]string{},
			flagValues:    map[string]string{},
			expectedKey:   "verbose",
			expectedValue: "false",
			description:   "Config file value should be used when no env var or flag is set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper for each test
			v := viper.New()

			// Set environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Configure viper
			v.SetEnvPrefix("AINATIVE_CODE")
			v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
			v.AutomaticEnv()

			// Set config file
			if tt.configFile != "" {
				v.SetConfigFile(tt.configFile)
				err := v.ReadInConfig()
				require.NoError(t, err)
			}

			// Bind environment variables
			v.BindEnv("provider")
			v.BindEnv("model")
			v.BindEnv("verbose")

			// Set flag values (simulating command-line flags)
			for key, value := range tt.flagValues {
				v.Set(key, value)
			}

			// Verify the expected value
			actualValue := v.GetString(tt.expectedKey)
			assert.Equal(t, tt.expectedValue, actualValue, tt.description)
		})
	}
}

// TestConfigShowWithEnvVars tests that config show command displays env var values
func TestConfigShowWithEnvVars(t *testing.T) {
	// Set environment variables
	t.Setenv("AINATIVE_CODE_PROVIDER", "anthropic")
	t.Setenv("AINATIVE_CODE_MODEL", "claude-3-opus")
	t.Setenv("AINATIVE_CODE_VERBOSE", "true")

	// Reset viper
	v := viper.New()

	// Configure viper
	v.SetEnvPrefix("AINATIVE_CODE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// Bind environment variables
	v.BindEnv("provider")
	v.BindEnv("model")
	v.BindEnv("verbose")

	// Verify environment variables are accessible
	assert.Equal(t, "anthropic", v.GetString("provider"))
	assert.Equal(t, "claude-3-opus", v.GetString("model"))
	assert.True(t, v.GetBool("verbose"))

	// Verify AllSettings includes env var values
	settings := v.AllSettings()
	assert.Equal(t, "anthropic", settings["provider"])
	assert.Equal(t, "claude-3-opus", settings["model"])
	// Note: AllSettings returns the raw value from env, which may be a string
	// Use GetBool for proper type conversion
	assert.NotNil(t, settings["verbose"])
}

// TestNoEnvVarsSet tests behavior when no environment variables are set
func TestNoEnvVarsSet(t *testing.T) {
	// Reset viper
	v := viper.New()

	// Configure viper
	v.SetEnvPrefix("AINATIVE_CODE")
	v.AutomaticEnv()

	// Bind environment variables
	v.BindEnv("provider")
	v.BindEnv("model")

	// Verify that unset env vars return empty/zero values
	assert.Equal(t, "", v.GetString("provider"))
	assert.Equal(t, "", v.GetString("model"))
}

// TestEnvVarCaseSensitivity tests that env var names are case-insensitive
func TestEnvVarCaseSensitivity(t *testing.T) {
	// Environment variables are case-sensitive in Unix, but Viper handles this
	t.Setenv("AINATIVE_CODE_PROVIDER", "test-provider")

	v := viper.New()
	v.SetEnvPrefix("AINATIVE_CODE")
	v.AutomaticEnv()
	v.BindEnv("provider")

	// Verify value is accessible
	assert.Equal(t, "test-provider", v.GetString("provider"))
	assert.Equal(t, "test-provider", v.GetString("PROVIDER")) // Viper is case-insensitive
}
