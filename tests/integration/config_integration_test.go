package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigLoadingFromFile tests loading configuration from a file
func TestConfigLoadingFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
app:
  name: test-app
  environment: production

llm:
  default_provider: anthropic
  anthropic:
    api_key: sk-ant-test-key
    model: claude-3-5-sonnet-20241022

platform:
  authentication:
    method: api_key
    api_key: test-platform-key

logging:
  level: debug
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load configuration
	loader := config.NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify loaded values
	assert.Equal(t, "test-app", cfg.App.Name)
	assert.Equal(t, "production", cfg.App.Environment)
	assert.Equal(t, "anthropic", cfg.LLM.DefaultProvider)
	assert.Equal(t, "sk-ant-test-key", cfg.LLM.Anthropic.APIKey)
	assert.Equal(t, "claude-3-5-sonnet-20241022", cfg.LLM.Anthropic.Model)
	assert.Equal(t, "debug", cfg.Logging.Level)
}

// TestConfigEnvironmentOverride tests that environment variables override file config
func TestConfigEnvironmentOverride(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
app:
  name: file-app
  environment: development

llm:
  default_provider: anthropic
  anthropic:
    api_key: file-api-key
    model: claude-3-sonnet

platform:
  authentication:
    method: api_key
    api_key: file-platform-key

logging:
  level: info
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables
	t.Setenv("AINATIVE_APP_NAME", "env-app")
	t.Setenv("AINATIVE_APP_ENVIRONMENT", "production")
	t.Setenv("AINATIVE_LLM_ANTHROPIC_API_KEY", "env-api-key")
	t.Setenv("AINATIVE_LOGGING_LEVEL", "debug")

	// Load configuration
	loader := config.NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Environment variables should override file values
	assert.Equal(t, "env-app", cfg.App.Name)
	assert.Equal(t, "production", cfg.App.Environment)
	assert.Equal(t, "env-api-key", cfg.LLM.Anthropic.APIKey)
	assert.Equal(t, "debug", cfg.Logging.Level)

	// Values not in env should use file values
	assert.Equal(t, "claude-3-sonnet", cfg.LLM.Anthropic.Model)
}

// TestConfigPrecedenceOrder tests the complete precedence order
func TestConfigPrecedenceOrder(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	// File config (lowest priority after defaults)
	configContent := `
app:
  name: file-app
  version: 1.0.0
  environment: development
  debug: false

llm:
  default_provider: anthropic
  anthropic:
    api_key: file-api-key
    temperature: 0.5

platform:
  authentication:
    method: api_key
    api_key: file-platform-key
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Environment variables (higher priority than file)
	t.Setenv("AINATIVE_APP_NAME", "env-app")
	t.Setenv("AINATIVE_LLM_ANTHROPIC_TEMPERATURE", "0.8")

	// Load configuration
	loader := config.NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Test precedence: env > file > defaults
	assert.Equal(t, "env-app", cfg.App.Name)                                      // from env
	assert.Equal(t, "1.0.0", cfg.App.Version)                                     // from file
	assert.Equal(t, "development", cfg.App.Environment)                           // from file
	assert.Equal(t, 0.8, cfg.LLM.Anthropic.Temperature)                           // from env
	assert.Equal(t, "claude-3-5-sonnet-20241022", cfg.LLM.Anthropic.Model)        // from defaults
	assert.Equal(t, 4096, cfg.LLM.Anthropic.MaxTokens)                            // from defaults
}

// TestConfigWithMultipleProviders tests configuration with multiple LLM providers
func TestConfigWithMultipleProviders(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
llm:
  default_provider: anthropic

  anthropic:
    api_key: sk-ant-test
    model: claude-3-5-sonnet-20241022

  openai:
    api_key: sk-openai-test
    model: gpt-4

  google:
    api_key: google-api-key
    model: gemini-pro

  bedrock:
    region: us-east-1
    model: anthropic.claude-3-sonnet-20240229-v1:0
    profile: default

  ollama:
    base_url: http://localhost:11434
    model: llama2

  fallback:
    enabled: true
    providers:
      - anthropic
      - openai
      - ollama

platform:
  authentication:
    method: api_key
    api_key: test-platform-key
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := config.NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify all providers are configured
	assert.Equal(t, "anthropic", cfg.LLM.DefaultProvider)

	require.NotNil(t, cfg.LLM.Anthropic)
	assert.Equal(t, "sk-ant-test", cfg.LLM.Anthropic.APIKey)

	require.NotNil(t, cfg.LLM.OpenAI)
	assert.Equal(t, "sk-openai-test", cfg.LLM.OpenAI.APIKey)

	require.NotNil(t, cfg.LLM.Google)
	assert.Equal(t, "google-api-key", cfg.LLM.Google.APIKey)

	require.NotNil(t, cfg.LLM.Bedrock)
	assert.Equal(t, "us-east-1", cfg.LLM.Bedrock.Region)

	require.NotNil(t, cfg.LLM.Ollama)
	assert.Equal(t, "llama2", cfg.LLM.Ollama.Model)

	require.NotNil(t, cfg.LLM.Fallback)
	assert.True(t, cfg.LLM.Fallback.Enabled)
	assert.Equal(t, []string{"anthropic", "openai", "ollama"}, cfg.LLM.Fallback.Providers)
}

// TestConfigWithAllServices tests configuration with all AINative services
func TestConfigWithAllServices(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
llm:
  default_provider: anthropic
  anthropic:
    api_key: sk-ant-test

platform:
  authentication:
    method: api_key
    api_key: platform-api-key
  organization:
    id: org-123
    name: Test Org
    workspace: production

services:
  zerodb:
    enabled: true
    project_id: proj-abc123
    endpoint: postgresql://localhost:5432
    database: testdb
    username: testuser
    password: testpass
    ssl: true
    ssl_mode: require

  design:
    enabled: true
    endpoint: https://design.ainative.studio/api
    api_key: design-api-key

  strapi:
    enabled: true
    endpoint: https://strapi.example.com
    api_key: strapi-api-key

  rlhf:
    enabled: true
    endpoint: https://rlhf.ainative.studio
    api_key: rlhf-api-key
    model_id: model-123
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := config.NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify platform configuration
	assert.Equal(t, "api_key", cfg.Platform.Authentication.Method)
	assert.Equal(t, "platform-api-key", cfg.Platform.Authentication.APIKey)
	assert.Equal(t, "org-123", cfg.Platform.Organization.ID)
	assert.Equal(t, "Test Org", cfg.Platform.Organization.Name)

	// Verify ZeroDB configuration
	require.NotNil(t, cfg.Services.ZeroDB)
	assert.True(t, cfg.Services.ZeroDB.Enabled)
	assert.Equal(t, "proj-abc123", cfg.Services.ZeroDB.ProjectID)
	assert.Equal(t, "testdb", cfg.Services.ZeroDB.Database)
	assert.True(t, cfg.Services.ZeroDB.SSL)

	// Verify Design service
	require.NotNil(t, cfg.Services.Design)
	assert.True(t, cfg.Services.Design.Enabled)
	assert.Equal(t, "design-api-key", cfg.Services.Design.APIKey)

	// Verify Strapi service
	require.NotNil(t, cfg.Services.Strapi)
	assert.True(t, cfg.Services.Strapi.Enabled)
	assert.Equal(t, "strapi-api-key", cfg.Services.Strapi.APIKey)

	// Verify RLHF service
	require.NotNil(t, cfg.Services.RLHF)
	assert.True(t, cfg.Services.RLHF.Enabled)
	assert.Equal(t, "rlhf-api-key", cfg.Services.RLHF.APIKey)
	assert.Equal(t, "model-123", cfg.Services.RLHF.ModelID)
}

// TestConfigValidationErrors tests that invalid configurations are rejected
func TestConfigValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectedErr string
	}{
		{
			name: "missing_api_key",
			config: `
llm:
  default_provider: anthropic
  anthropic:
    model: claude-3-sonnet
`,
			expectedErr: "api_key",
		},
		{
			name: "invalid_temperature",
			config: `
llm:
  default_provider: anthropic
  anthropic:
    api_key: sk-ant-test
    temperature: 1.5
`,
			expectedErr: "temperature",
		},
		{
			name: "invalid_environment",
			config: `
app:
  environment: invalid-env

llm:
  default_provider: anthropic
  anthropic:
    api_key: sk-ant-test
`,
			expectedErr: "environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "test-config.yaml")

			err := os.WriteFile(configPath, []byte(tt.config), 0644)
			require.NoError(t, err)

			loader := config.NewLoader()
			_, err = loader.LoadFromFile(configPath)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// TestMinimalConfig tests that minimal configuration works
func TestMinimalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minimal-config.yaml")

	// Minimal config - only required fields
	configContent := `
llm:
  default_provider: anthropic
  anthropic:
    api_key: sk-ant-test

platform:
  authentication:
    method: api_key
    api_key: test-platform-key
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := config.NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify required field
	assert.Equal(t, "sk-ant-test", cfg.LLM.Anthropic.APIKey)

	// Verify defaults are applied
	assert.Equal(t, "claude-3-5-sonnet-20241022", cfg.LLM.Anthropic.Model)
	assert.Equal(t, 4096, cfg.LLM.Anthropic.MaxTokens)
	assert.Equal(t, "ainative-code", cfg.App.Name)
	assert.Equal(t, "info", cfg.Logging.Level)
}

// TestConfigFileNotFound tests behavior when config file doesn't exist
func TestConfigFileNotFound(t *testing.T) {
	loader := config.NewLoader()
	_, err := loader.LoadFromFile("/non/existent/path/config.yaml")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file does not exist")
}

// TestEnvironmentVariableMapping tests all environment variable mappings
func TestEnvironmentVariableMapping(t *testing.T) {
	// Set various environment variables
	envVars := map[string]string{
		"AINATIVE_APP_NAME":                              "env-test-app",
		"AINATIVE_APP_ENVIRONMENT":                       "production",
		"AINATIVE_LLM_DEFAULT_PROVIDER":                  "openai",
		"AINATIVE_LLM_ANTHROPIC_API_KEY":                 "sk-ant-env",
		"AINATIVE_LLM_OPENAI_API_KEY":                    "sk-openai-env",
		"AINATIVE_LLM_GOOGLE_API_KEY":                    "google-env",
		"AINATIVE_LLM_BEDROCK_REGION":                    "eu-west-1",
		"AINATIVE_LLM_OLLAMA_BASE_URL":                   "http://192.168.1.100:11434",
		"AINATIVE_PLATFORM_AUTHENTICATION_API_KEY":       "platform-env-key",
		"AINATIVE_SERVICES_ZERODB_PROJECT_ID":            "proj-env-123",
		"AINATIVE_SERVICES_ZERODB_ENDPOINT":              "postgresql://env-host:5432",
		"AINATIVE_SERVICES_DESIGN_API_KEY":               "design-env-key",
		"AINATIVE_SERVICES_STRAPI_API_KEY":               "strapi-env-key",
		"AINATIVE_SERVICES_RLHF_API_KEY":                 "rlhf-env-key",
		"AINATIVE_LOGGING_LEVEL":                         "debug",
		"AINATIVE_SECURITY_ENCRYPT_CONFIG":               "true",
		"AINATIVE_SECURITY_ENCRYPTION_KEY":               "this-is-a-32-char-encryption-key-test",
	}

	for key, value := range envVars {
		t.Setenv(key, value)
	}

	// Create minimal config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
llm:
  default_provider: anthropic
  anthropic:
    api_key: file-key
  openai:
    api_key: file-openai-key
  google:
    api_key: file-google-key
  bedrock:
    region: us-east-1
    profile: default
  ollama:
    base_url: http://localhost:11434
    model: llama2

platform:
  authentication:
    method: api_key
    api_key: file-platform-key

services:
  zerodb:
    enabled: true
    endpoint: postgresql://localhost:5432
    database: testdb
  design:
    enabled: true
    endpoint: https://design.example.com
  strapi:
    enabled: true
    endpoint: https://strapi.example.com
  rlhf:
    enabled: true
    endpoint: https://rlhf.example.com
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := config.NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify environment variables override file config
	assert.Equal(t, "env-test-app", cfg.App.Name)
	assert.Equal(t, "production", cfg.App.Environment)
	assert.Equal(t, "openai", cfg.LLM.DefaultProvider)
	assert.Equal(t, "sk-ant-env", cfg.LLM.Anthropic.APIKey)
	assert.Equal(t, "sk-openai-env", cfg.LLM.OpenAI.APIKey)
	assert.Equal(t, "google-env", cfg.LLM.Google.APIKey)
	assert.Equal(t, "eu-west-1", cfg.LLM.Bedrock.Region)
	assert.Equal(t, "http://192.168.1.100:11434", cfg.LLM.Ollama.BaseURL)
	assert.Equal(t, "platform-env-key", cfg.Platform.Authentication.APIKey)
	assert.Equal(t, "proj-env-123", cfg.Services.ZeroDB.ProjectID)
	assert.Equal(t, "postgresql://env-host:5432", cfg.Services.ZeroDB.Endpoint)
	assert.Equal(t, "design-env-key", cfg.Services.Design.APIKey)
	assert.Equal(t, "strapi-env-key", cfg.Services.Strapi.APIKey)
	assert.Equal(t, "rlhf-env-key", cfg.Services.RLHF.APIKey)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.True(t, cfg.Security.EncryptConfig)
}

// TestConfigWithPerformanceSettings tests performance configuration
func TestConfigWithPerformanceSettings(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
llm:
  default_provider: anthropic
  anthropic:
    api_key: sk-ant-test

platform:
  authentication:
    method: api_key
    api_key: test-platform-key

performance:
  cache:
    enabled: true
    type: redis
    redis_url: redis://localhost:6379/0
    ttl: 30m

  rate_limit:
    enabled: true
    requests_per_minute: 120
    burst_size: 20

  concurrency:
    max_workers: 20
    max_queue_size: 200

  circuit_breaker:
    enabled: true
    failure_threshold: 10
    success_threshold: 3
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := config.NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify performance settings
	assert.True(t, cfg.Performance.Cache.Enabled)
	assert.Equal(t, "redis", cfg.Performance.Cache.Type)
	assert.Equal(t, "redis://localhost:6379/0", cfg.Performance.Cache.RedisURL)

	assert.True(t, cfg.Performance.RateLimit.Enabled)
	assert.Equal(t, 120, cfg.Performance.RateLimit.RequestsPerMinute)

	assert.Equal(t, 20, cfg.Performance.Concurrency.MaxWorkers)
	assert.Equal(t, 200, cfg.Performance.Concurrency.MaxQueueSize)

	assert.True(t, cfg.Performance.CircuitBreaker.Enabled)
	assert.Equal(t, 10, cfg.Performance.CircuitBreaker.FailureThreshold)
}
