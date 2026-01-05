package config

import (
	"time"
)

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		App:         DefaultAppConfig(),
		LLM:         DefaultLLMConfig(),
		Platform:    DefaultPlatformConfig(),
		Services:    DefaultServicesConfig(),
		Tools:       DefaultToolsConfig(),
		Performance: DefaultPerformanceConfig(),
		Logging:     DefaultLoggingConfig(),
		Security:    DefaultSecurityConfig(),
	}
}

// DefaultAppConfig returns default application configuration
func DefaultAppConfig() AppConfig {
	return AppConfig{
		Name:        "ainative-code",
		Version:     "0.1.0",
		Environment: "development",
		Debug:       false,
	}
}

// DefaultLLMConfig returns default LLM configuration with Anthropic as default provider
func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		DefaultProvider: "anthropic",
		Anthropic:       DefaultAnthropicConfig(),
		OpenAI:          nil, // Only configure when needed
		Google:          nil,
		Bedrock:         nil,
		Azure:           nil,
		Ollama:          nil,
		Fallback:        DefaultFallbackConfig(),
	}
}

// DefaultAnthropicConfig returns default Anthropic configuration
func DefaultAnthropicConfig() *AnthropicConfig {
	return &AnthropicConfig{
		APIKey:        "", // Must be provided by user
		Model:         "claude-3-5-sonnet-20241022",
		MaxTokens:     4096,
		Temperature:   0.7,
		TopP:          1.0,
		TopK:          0,
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		BaseURL:       "",
		APIVersion:    "2023-06-01",
		ExtendedThinking: &ExtendedThinkingConfig{
			Enabled:    false,
			AutoExpand: false,
			MaxDepth:   3,
		},
		Retry: DefaultRetryConfig(),
	}
}

// DefaultOpenAIConfig returns default OpenAI configuration
func DefaultOpenAIConfig() *OpenAIConfig {
	return &OpenAIConfig{
		APIKey:           "", // Must be provided by user
		Model:            "gpt-4-turbo-preview",
		Organization:     "",
		MaxTokens:        4096,
		Temperature:      0.7,
		TopP:             1.0,
		FrequencyPenalty: 0.0,
		PresencePenalty:  0.0,
		Timeout:          30 * time.Second,
		RetryAttempts:    3,
		BaseURL:          "",
		Retry:            DefaultRetryConfig(),
	}
}

// DefaultGoogleConfig returns default Google (Gemini) configuration
func DefaultGoogleConfig() *GoogleConfig {
	return &GoogleConfig{
		APIKey:        "", // Must be provided by user
		Model:         "gemini-pro",
		ProjectID:     "",
		Location:      "us-central1",
		MaxTokens:     4096,
		Temperature:   0.7,
		TopP:          1.0,
		TopK:          40,
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
	}
}

// DefaultBedrockConfig returns default AWS Bedrock configuration
func DefaultBedrockConfig() *BedrockConfig {
	return &BedrockConfig{
		Region:          "us-east-1",
		Model:           "anthropic.claude-3-sonnet-20240229-v1:0",
		AccessKeyID:     "",
		SecretAccessKey: "",
		SessionToken:    "",
		Profile:         "default",
		MaxTokens:       4096,
		Temperature:     0.7,
		TopP:            1.0,
		Timeout:         60 * time.Second,
		RetryAttempts:   3,
	}
}

// DefaultAzureConfig returns default Azure OpenAI configuration
func DefaultAzureConfig() *AzureConfig {
	return &AzureConfig{
		APIKey:         "", // Must be provided by user
		Endpoint:       "", // Must be provided by user
		DeploymentName: "", // Must be provided by user
		APIVersion:     "2023-05-15",
		MaxTokens:      4096,
		Temperature:    0.7,
		TopP:           1.0,
		Timeout:        30 * time.Second,
		RetryAttempts:  3,
	}
}

// DefaultOllamaConfig returns default Ollama configuration
func DefaultOllamaConfig() *OllamaConfig {
	return &OllamaConfig{
		BaseURL:       "http://localhost:11434",
		Model:         "", // Must be provided by user
		MaxTokens:     4096,
		Temperature:   0.7,
		TopP:          1.0,
		TopK:          40,
		Timeout:       120 * time.Second,
		RetryAttempts: 1,
		KeepAlive:     "5m",
	}
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:            3,
		InitialDelay:           1 * time.Second,
		MaxDelay:               30 * time.Second,
		Multiplier:             2.0,
		EnableJitter:           true,
		EnableAPIKeyResolution: false,
		EnableTokenReduction:   true,
		TokenReductionPercent:  20,
		EnableTimeoutIncrease:  true,
		TimeoutIncreasePercent: 50,
	}
}

// DefaultFallbackConfig returns default fallback configuration
func DefaultFallbackConfig() *FallbackConfig {
	return &FallbackConfig{
		Enabled:    false,
		Providers:  []string{},
		MaxRetries: 2,
		RetryDelay: 1 * time.Second,
	}
}

// DefaultPlatformConfig returns default platform configuration
func DefaultPlatformConfig() PlatformConfig {
	return PlatformConfig{
		Authentication: DefaultAuthConfig(),
		Organization:   DefaultOrgConfig(),
	}
}

// DefaultAuthConfig returns default authentication configuration
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Method:       "api_key",
		APIKey:       "",
		Token:        "",
		RefreshToken: "",
		ClientID:     "",
		ClientSecret: "",
		TokenURL:     "",
		Scopes:       []string{},
		Timeout:      10 * time.Second,
	}
}

// DefaultOrgConfig returns default organization configuration
func DefaultOrgConfig() OrgConfig {
	return OrgConfig{
		ID:        "",
		Name:      "",
		Workspace: "default",
	}
}

// DefaultServicesConfig returns default services configuration
func DefaultServicesConfig() ServicesConfig {
	return ServicesConfig{
		ZeroDB: DefaultZeroDBConfig(),
		Design: DefaultDesignConfig(),
		Strapi: DefaultStrapiConfig(),
		RLHF:   DefaultRLHFConfig(),
	}
}

// DefaultZeroDBConfig returns default ZeroDB configuration
func DefaultZeroDBConfig() *ZeroDBConfig {
	return &ZeroDBConfig{
		Enabled:         false,
		Endpoint:        "",
		Database:        "ainative_code",
		Username:        "",
		Password:        "",
		SSL:             false,
		SSLMode:         "disable",
		MaxConnections:  10,
		IdleConnections: 2,
		ConnMaxLifetime: 1 * time.Hour,
		Timeout:         5 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      1 * time.Second,
	}
}

// DefaultDesignConfig returns default Design service configuration
func DefaultDesignConfig() *DesignConfig {
	return &DesignConfig{
		Enabled:       false,
		Endpoint:      "",
		APIKey:        "",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
	}
}

// DefaultStrapiConfig returns default Strapi configuration
func DefaultStrapiConfig() *StrapiConfig {
	return &StrapiConfig{
		Enabled:       false,
		Endpoint:      "",
		APIKey:        "",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
	}
}

// DefaultRLHFConfig returns default RLHF service configuration
func DefaultRLHFConfig() *RLHFConfig {
	return &RLHFConfig{
		Enabled:       false,
		Endpoint:      "",
		APIKey:        "",
		Timeout:       60 * time.Second,
		RetryAttempts: 3,
		ModelID:       "",
	}
}

// DefaultToolsConfig returns default tools configuration
func DefaultToolsConfig() ToolsConfig {
	return ToolsConfig{
		FileSystem:   DefaultFileSystemToolConfig(),
		Terminal:     DefaultTerminalToolConfig(),
		Browser:      DefaultBrowserToolConfig(),
		CodeAnalysis: DefaultCodeAnalysisToolConfig(),
	}
}

// DefaultFileSystemToolConfig returns default filesystem tool configuration
func DefaultFileSystemToolConfig() *FileSystemToolConfig {
	return &FileSystemToolConfig{
		Enabled:           false,
		AllowedPaths:      []string{},
		BlockedPaths:      []string{},
		MaxFileSize:       104857600, // 100MB
		AllowedExtensions: []string{},
	}
}

// DefaultTerminalToolConfig returns default terminal tool configuration
func DefaultTerminalToolConfig() *TerminalToolConfig {
	return &TerminalToolConfig{
		Enabled:         false,
		AllowedCommands: []string{},
		BlockedCommands: []string{},
		Timeout:         5 * time.Minute,
		WorkingDir:      "",
	}
}

// DefaultBrowserToolConfig returns default browser tool configuration
func DefaultBrowserToolConfig() *BrowserToolConfig {
	return &BrowserToolConfig{
		Enabled:   false,
		Headless:  true,
		Timeout:   30 * time.Second,
		UserAgent: "AINative-Code/0.1.0",
	}
}

// DefaultCodeAnalysisToolConfig returns default code analysis tool configuration
func DefaultCodeAnalysisToolConfig() *CodeAnalysisToolConfig {
	return &CodeAnalysisToolConfig{
		Enabled:      false,
		Languages:    []string{"go", "python", "javascript", "typescript", "java"},
		MaxFileSize:  10485760, // 10MB
		IncludeTests: true,
	}
}

// DefaultPerformanceConfig returns default performance configuration
func DefaultPerformanceConfig() PerformanceConfig {
	return PerformanceConfig{
		Cache:          DefaultCacheConfig(),
		RateLimit:      DefaultRateLimitConfig(),
		Concurrency:    DefaultConcurrencyConfig(),
		CircuitBreaker: DefaultCircuitBreakerConfig(),
	}
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled:       false,
		Type:          "memory",
		TTL:           1 * time.Hour,
		MaxSize:       100, // 100MB
		RedisURL:      "",
		MemcachedURL:  "",
	}
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:           false,
		RequestsPerMinute: 60,
		BurstSize:         10,
		TimeWindow:        1 * time.Minute,
		PerUser:           false,
		PerEndpoint:       false,
		Storage:           "memory",
		RedisURL:          "",
		EndpointLimits:    map[string]int{},
		SkipPaths:         []string{},
		IPAllowlist:       []string{},
		IPBlocklist:       []string{},
	}
}

// DefaultConcurrencyConfig returns default concurrency configuration
func DefaultConcurrencyConfig() ConcurrencyConfig {
	return ConcurrencyConfig{
		MaxWorkers:    10,
		MaxQueueSize:  100,
		WorkerTimeout: 5 * time.Minute,
	}
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Enabled:          false,
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          60 * time.Second,
		ResetTimeout:     30 * time.Second,
	}
}

// DefaultLoggingConfig returns default logging configuration
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:         "info",
		Format:        "json",
		Output:        "stdout",
		FilePath:      "",
		MaxSize:       100, // 100MB
		MaxBackups:    3,
		MaxAge:        7, // 7 days
		Compress:      true,
		SensitiveKeys: []string{"api_key", "password", "token", "secret"},
	}
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		EncryptConfig:  false,
		EncryptionKey:  "",
		AllowedOrigins: []string{},
		TLSEnabled:     false,
		TLSCertPath:    "",
		TLSKeyPath:     "",
		SecretRotation: 0,
	}
}
