package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.NotNil(t, cfg, "DefaultConfig should not return nil")

	// Verify all top-level sections are initialized
	assert.NotZero(t, cfg.App, "App config should be initialized")
	assert.NotZero(t, cfg.LLM, "LLM config should be initialized")
	assert.NotZero(t, cfg.Platform, "Platform config should be initialized")
	assert.NotZero(t, cfg.Services, "Services config should be initialized")
	assert.NotZero(t, cfg.Tools, "Tools config should be initialized")
	assert.NotZero(t, cfg.Performance, "Performance config should be initialized")
	assert.NotZero(t, cfg.Logging, "Logging config should be initialized")
	assert.NotZero(t, cfg.Security, "Security config should be initialized")
}

func TestDefaultAppConfig(t *testing.T) {
	cfg := DefaultAppConfig()

	assert.Equal(t, "ainative-code", cfg.Name)
	assert.Equal(t, "0.1.0", cfg.Version)
	assert.Equal(t, "development", cfg.Environment)
	assert.False(t, cfg.Debug)
}

func TestDefaultLLMConfig(t *testing.T) {
	cfg := DefaultLLMConfig()

	assert.Equal(t, "anthropic", cfg.DefaultProvider)
	assert.NotNil(t, cfg.Anthropic, "Anthropic should be configured by default")
	assert.Nil(t, cfg.OpenAI, "OpenAI should not be configured by default")
	assert.Nil(t, cfg.Google, "Google should not be configured by default")
	assert.Nil(t, cfg.Bedrock, "Bedrock should not be configured by default")
	assert.Nil(t, cfg.Azure, "Azure should not be configured by default")
	assert.Nil(t, cfg.Ollama, "Ollama should not be configured by default")
	assert.NotNil(t, cfg.Fallback, "Fallback should be configured")
}

func TestDefaultAnthropicConfig(t *testing.T) {
	cfg := DefaultAnthropicConfig()

	assert.Equal(t, "", cfg.APIKey, "API key should be empty by default")
	assert.Equal(t, "claude-3-5-sonnet-20241022", cfg.Model)
	assert.Equal(t, 4096, cfg.MaxTokens)
	assert.Equal(t, 0.7, cfg.Temperature)
	assert.Equal(t, 1.0, cfg.TopP)
	assert.Equal(t, 0, cfg.TopK)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.RetryAttempts)
	assert.Equal(t, "2023-06-01", cfg.APIVersion)
	assert.NotNil(t, cfg.ExtendedThinking)
	assert.NotNil(t, cfg.Retry)
}

func TestDefaultOpenAIConfig(t *testing.T) {
	cfg := DefaultOpenAIConfig()

	assert.Equal(t, "", cfg.APIKey)
	assert.Equal(t, "gpt-4-turbo-preview", cfg.Model)
	assert.Equal(t, 4096, cfg.MaxTokens)
	assert.Equal(t, 0.7, cfg.Temperature)
	assert.Equal(t, 1.0, cfg.TopP)
	assert.Equal(t, 0.0, cfg.FrequencyPenalty)
	assert.Equal(t, 0.0, cfg.PresencePenalty)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.RetryAttempts)
	assert.NotNil(t, cfg.Retry)
}

func TestDefaultGoogleConfig(t *testing.T) {
	cfg := DefaultGoogleConfig()

	assert.Equal(t, "", cfg.APIKey)
	assert.Equal(t, "gemini-pro", cfg.Model)
	assert.Equal(t, "us-central1", cfg.Location)
	assert.Equal(t, 4096, cfg.MaxTokens)
	assert.Equal(t, 0.7, cfg.Temperature)
	assert.Equal(t, 1.0, cfg.TopP)
	assert.Equal(t, 40, cfg.TopK)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.RetryAttempts)
}

func TestDefaultBedrockConfig(t *testing.T) {
	cfg := DefaultBedrockConfig()

	assert.Equal(t, "us-east-1", cfg.Region)
	assert.Equal(t, "anthropic.claude-3-sonnet-20240229-v1:0", cfg.Model)
	assert.Equal(t, "default", cfg.Profile)
	assert.Equal(t, 4096, cfg.MaxTokens)
	assert.Equal(t, 0.7, cfg.Temperature)
	assert.Equal(t, 1.0, cfg.TopP)
	assert.Equal(t, 60*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.RetryAttempts)
}

func TestDefaultAzureConfig(t *testing.T) {
	cfg := DefaultAzureConfig()

	assert.Equal(t, "", cfg.APIKey)
	assert.Equal(t, "", cfg.Endpoint)
	assert.Equal(t, "", cfg.DeploymentName)
	assert.Equal(t, "2023-05-15", cfg.APIVersion)
	assert.Equal(t, 4096, cfg.MaxTokens)
	assert.Equal(t, 0.7, cfg.Temperature)
	assert.Equal(t, 1.0, cfg.TopP)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.RetryAttempts)
}

func TestDefaultOllamaConfig(t *testing.T) {
	cfg := DefaultOllamaConfig()

	assert.Equal(t, "http://localhost:11434", cfg.BaseURL)
	assert.Equal(t, "", cfg.Model)
	assert.Equal(t, 4096, cfg.MaxTokens)
	assert.Equal(t, 0.7, cfg.Temperature)
	assert.Equal(t, 1.0, cfg.TopP)
	assert.Equal(t, 40, cfg.TopK)
	assert.Equal(t, 120*time.Second, cfg.Timeout)
	assert.Equal(t, 1, cfg.RetryAttempts)
	assert.Equal(t, "5m", cfg.KeepAlive)
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	assert.Equal(t, 3, cfg.MaxAttempts)
	assert.Equal(t, 1*time.Second, cfg.InitialDelay)
	assert.Equal(t, 30*time.Second, cfg.MaxDelay)
	assert.Equal(t, 2.0, cfg.Multiplier)
	assert.True(t, cfg.EnableJitter)
	assert.False(t, cfg.EnableAPIKeyResolution)
	assert.True(t, cfg.EnableTokenReduction)
	assert.Equal(t, 20, cfg.TokenReductionPercent)
	assert.True(t, cfg.EnableTimeoutIncrease)
	assert.Equal(t, 50, cfg.TimeoutIncreasePercent)
}

func TestDefaultFallbackConfig(t *testing.T) {
	cfg := DefaultFallbackConfig()

	assert.False(t, cfg.Enabled)
	assert.Empty(t, cfg.Providers)
	assert.Equal(t, 2, cfg.MaxRetries)
	assert.Equal(t, 1*time.Second, cfg.RetryDelay)
}

func TestDefaultPlatformConfig(t *testing.T) {
	cfg := DefaultPlatformConfig()

	assert.NotZero(t, cfg.Authentication)
	assert.NotZero(t, cfg.Organization)
}

func TestDefaultAuthConfig(t *testing.T) {
	cfg := DefaultAuthConfig()

	assert.Equal(t, "api_key", cfg.Method)
	assert.Empty(t, cfg.APIKey)
	assert.Empty(t, cfg.Token)
	assert.Empty(t, cfg.RefreshToken)
	assert.Empty(t, cfg.ClientID)
	assert.Empty(t, cfg.ClientSecret)
	assert.Empty(t, cfg.TokenURL)
	assert.Empty(t, cfg.Scopes)
	assert.Equal(t, 10*time.Second, cfg.Timeout)
}

func TestDefaultOrgConfig(t *testing.T) {
	cfg := DefaultOrgConfig()

	assert.Empty(t, cfg.ID)
	assert.Empty(t, cfg.Name)
	assert.Equal(t, "default", cfg.Workspace)
}

func TestDefaultServicesConfig(t *testing.T) {
	cfg := DefaultServicesConfig()

	assert.NotNil(t, cfg.ZeroDB)
	assert.NotNil(t, cfg.Design)
	assert.NotNil(t, cfg.Strapi)
	assert.NotNil(t, cfg.RLHF)
}

func TestDefaultZeroDBConfig(t *testing.T) {
	cfg := DefaultZeroDBConfig()

	assert.False(t, cfg.Enabled)
	assert.Empty(t, cfg.Endpoint)
	assert.Equal(t, "ainative_code", cfg.Database)
	assert.Empty(t, cfg.Username)
	assert.Empty(t, cfg.Password)
	assert.False(t, cfg.SSL)
	assert.Equal(t, "disable", cfg.SSLMode)
	assert.Equal(t, 10, cfg.MaxConnections)
	assert.Equal(t, 2, cfg.IdleConnections)
	assert.Equal(t, 1*time.Hour, cfg.ConnMaxLifetime)
	assert.Equal(t, 5*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.RetryAttempts)
	assert.Equal(t, 1*time.Second, cfg.RetryDelay)
}

func TestDefaultDesignConfig(t *testing.T) {
	cfg := DefaultDesignConfig()

	assert.False(t, cfg.Enabled)
	assert.Empty(t, cfg.Endpoint)
	assert.Empty(t, cfg.APIKey)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.RetryAttempts)
}

func TestDefaultStrapiConfig(t *testing.T) {
	cfg := DefaultStrapiConfig()

	assert.False(t, cfg.Enabled)
	assert.Empty(t, cfg.Endpoint)
	assert.Empty(t, cfg.APIKey)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.RetryAttempts)
}

func TestDefaultRLHFConfig(t *testing.T) {
	cfg := DefaultRLHFConfig()

	assert.False(t, cfg.Enabled)
	assert.Empty(t, cfg.Endpoint)
	assert.Empty(t, cfg.APIKey)
	assert.Equal(t, 60*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.RetryAttempts)
	assert.Empty(t, cfg.ModelID)
}

func TestDefaultToolsConfig(t *testing.T) {
	cfg := DefaultToolsConfig()

	assert.NotNil(t, cfg.FileSystem)
	assert.NotNil(t, cfg.Terminal)
	assert.NotNil(t, cfg.Browser)
	assert.NotNil(t, cfg.CodeAnalysis)
}

func TestDefaultFileSystemToolConfig(t *testing.T) {
	cfg := DefaultFileSystemToolConfig()

	assert.False(t, cfg.Enabled)
	assert.Empty(t, cfg.AllowedPaths)
	assert.Empty(t, cfg.BlockedPaths)
	assert.Equal(t, int64(104857600), cfg.MaxFileSize) // 100MB
	assert.Empty(t, cfg.AllowedExtensions)
}

func TestDefaultTerminalToolConfig(t *testing.T) {
	cfg := DefaultTerminalToolConfig()

	assert.False(t, cfg.Enabled)
	assert.Empty(t, cfg.AllowedCommands)
	assert.Empty(t, cfg.BlockedCommands)
	assert.Equal(t, 5*time.Minute, cfg.Timeout)
	assert.Empty(t, cfg.WorkingDir)
}

func TestDefaultBrowserToolConfig(t *testing.T) {
	cfg := DefaultBrowserToolConfig()

	assert.False(t, cfg.Enabled)
	assert.True(t, cfg.Headless)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, "AINative-Code/0.1.0", cfg.UserAgent)
}

func TestDefaultCodeAnalysisToolConfig(t *testing.T) {
	cfg := DefaultCodeAnalysisToolConfig()

	assert.False(t, cfg.Enabled)
	assert.Equal(t, []string{"go", "python", "javascript", "typescript", "java"}, cfg.Languages)
	assert.Equal(t, int64(10485760), cfg.MaxFileSize) // 10MB
	assert.True(t, cfg.IncludeTests)
}

func TestDefaultPerformanceConfig(t *testing.T) {
	cfg := DefaultPerformanceConfig()

	assert.NotZero(t, cfg.Cache)
	assert.NotZero(t, cfg.RateLimit)
	assert.NotZero(t, cfg.Concurrency)
	assert.NotZero(t, cfg.CircuitBreaker)
}

func TestDefaultCacheConfig(t *testing.T) {
	cfg := DefaultCacheConfig()

	assert.False(t, cfg.Enabled)
	assert.Equal(t, "memory", cfg.Type)
	assert.Equal(t, 1*time.Hour, cfg.TTL)
	assert.Equal(t, int64(100), cfg.MaxSize)
	assert.Empty(t, cfg.RedisURL)
	assert.Empty(t, cfg.MemcachedURL)
}

func TestDefaultRateLimitConfig(t *testing.T) {
	cfg := DefaultRateLimitConfig()

	assert.False(t, cfg.Enabled)
	assert.Equal(t, 60, cfg.RequestsPerMinute)
	assert.Equal(t, 10, cfg.BurstSize)
	assert.Equal(t, 1*time.Minute, cfg.TimeWindow)
	assert.False(t, cfg.PerUser)
	assert.False(t, cfg.PerEndpoint)
	assert.Equal(t, "memory", cfg.Storage)
	assert.Empty(t, cfg.RedisURL)
	assert.NotNil(t, cfg.EndpointLimits)
	assert.Empty(t, cfg.SkipPaths)
	assert.Empty(t, cfg.IPAllowlist)
	assert.Empty(t, cfg.IPBlocklist)
}

func TestDefaultConcurrencyConfig(t *testing.T) {
	cfg := DefaultConcurrencyConfig()

	assert.Equal(t, 10, cfg.MaxWorkers)
	assert.Equal(t, 100, cfg.MaxQueueSize)
	assert.Equal(t, 5*time.Minute, cfg.WorkerTimeout)
}

func TestDefaultCircuitBreakerConfig(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()

	assert.False(t, cfg.Enabled)
	assert.Equal(t, 5, cfg.FailureThreshold)
	assert.Equal(t, 2, cfg.SuccessThreshold)
	assert.Equal(t, 60*time.Second, cfg.Timeout)
	assert.Equal(t, 30*time.Second, cfg.ResetTimeout)
}

func TestDefaultLoggingConfig(t *testing.T) {
	cfg := DefaultLoggingConfig()

	assert.Equal(t, "info", cfg.Level)
	assert.Equal(t, "json", cfg.Format)
	assert.Equal(t, "stdout", cfg.Output)
	assert.Empty(t, cfg.FilePath)
	assert.Equal(t, 100, cfg.MaxSize)
	assert.Equal(t, 3, cfg.MaxBackups)
	assert.Equal(t, 7, cfg.MaxAge)
	assert.True(t, cfg.Compress)
	assert.Equal(t, []string{"api_key", "password", "token", "secret"}, cfg.SensitiveKeys)
}

func TestDefaultSecurityConfig(t *testing.T) {
	cfg := DefaultSecurityConfig()

	assert.False(t, cfg.EncryptConfig)
	assert.Empty(t, cfg.EncryptionKey)
	assert.Empty(t, cfg.AllowedOrigins)
	assert.False(t, cfg.TLSEnabled)
	assert.Empty(t, cfg.TLSCertPath)
	assert.Empty(t, cfg.TLSKeyPath)
	assert.Equal(t, time.Duration(0), cfg.SecretRotation)
}

// Integration tests to ensure defaults work with validation

func TestDefaultConfig_PassesValidation(t *testing.T) {
	cfg := DefaultConfig()

	// Set required fields that must be provided by user
	cfg.LLM.Anthropic.APIKey = "sk-ant-test-key-for-validation"
	cfg.Platform.Authentication.APIKey = "test-platform-api-key"

	validator := NewValidator(cfg)
	err := validator.Validate()

	// Should pass validation with only API key provided
	assert.NoError(t, err, "Default config should pass validation when required API keys are provided")
}

func TestDefaultConfig_CustomizationPreservesDefaults(t *testing.T) {
	cfg := DefaultConfig()

	// Customize one field
	cfg.App.Name = "custom-app"

	// Other defaults should be preserved
	assert.Equal(t, "0.1.0", cfg.App.Version)
	assert.Equal(t, "development", cfg.App.Environment)
	assert.Equal(t, "anthropic", cfg.LLM.DefaultProvider)
}

func TestDefaultConfig_AllProvidersHaveSensibleDefaults(t *testing.T) {
	providers := []struct {
		name   string
		config interface{}
	}{
		{"Anthropic", DefaultAnthropicConfig()},
		{"OpenAI", DefaultOpenAIConfig()},
		{"Google", DefaultGoogleConfig()},
		{"Bedrock", DefaultBedrockConfig()},
		{"Azure", DefaultAzureConfig()},
		{"Ollama", DefaultOllamaConfig()},
	}

	for _, p := range providers {
		t.Run(p.name, func(t *testing.T) {
			assert.NotNil(t, p.config, "%s config should not be nil", p.name)
		})
	}
}
