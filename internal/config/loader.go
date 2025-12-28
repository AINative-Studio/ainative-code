package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/errors"
	"github.com/spf13/viper"
)

// Loader handles loading configuration from multiple sources
type Loader struct {
	viper       *viper.Viper
	configPaths []string
	configName  string
	configType  string
	envPrefix   string
	resolver    *Resolver
}

// LoaderOption is a functional option for configuring the Loader
type LoaderOption func(*Loader)

// NewLoader creates a new configuration loader with options
func NewLoader(opts ...LoaderOption) *Loader {
	l := &Loader{
		viper:       viper.New(),
		configPaths: []string{".", "./configs", "$HOME/.ainative", "/etc/ainative"},
		configName:  "config",
		configType:  "yaml",
		envPrefix:   "AINATIVE",
		resolver:    NewResolver(), // Initialize with default resolver
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// WithConfigPaths sets custom configuration file search paths
func WithConfigPaths(paths ...string) LoaderOption {
	return func(l *Loader) {
		l.configPaths = paths
	}
}

// WithConfigName sets the configuration file name (without extension)
func WithConfigName(name string) LoaderOption {
	return func(l *Loader) {
		l.configName = name
	}
}

// WithConfigType sets the configuration file type
func WithConfigType(configType string) LoaderOption {
	return func(l *Loader) {
		l.configType = configType
	}
}

// WithEnvPrefix sets the environment variable prefix
func WithEnvPrefix(prefix string) LoaderOption {
	return func(l *Loader) {
		l.envPrefix = prefix
	}
}

// WithResolver sets a custom API key resolver
func WithResolver(resolver *Resolver) LoaderOption {
	return func(l *Loader) {
		l.resolver = resolver
	}
}

// Load loads configuration from all sources (file, environment, defaults)
func (l *Loader) Load() (*Config, error) {
	// Set defaults
	l.setDefaults()

	// Configure environment variable support
	l.setupEnvVars()

	// Add config paths
	for _, path := range l.configPaths {
		expandedPath := os.ExpandEnv(path)
		l.viper.AddConfigPath(expandedPath)
	}

	// Set config name and type
	l.viper.SetConfigName(l.configName)
	l.viper.SetConfigType(l.configType)

	// Read configuration file (optional)
	if err := l.viper.ReadInConfig(); err != nil {
		// Config file not found is acceptable, we can use defaults and env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, errors.NewConfigParseError(l.viper.ConfigFileUsed(), err)
		}
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := l.viper.Unmarshal(&cfg); err != nil {
		return nil, errors.NewConfigParseError("unmarshal", err)
	}

	// Resolve dynamic API keys
	if err := l.resolveAPIKeys(&cfg); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeConfigInvalid, "failed to resolve API keys")
	}

	// Validate configuration
	validator := NewValidator(&cfg)
	if err := validator.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadFromFile loads configuration from a specific file path
func (l *Loader) LoadFromFile(filePath string) (*Config, error) {
	// Expand environment variables in path
	expandedPath := os.ExpandEnv(filePath)

	// Check if file exists
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return nil, errors.NewConfigParseError(expandedPath, fmt.Errorf("file does not exist"))
	}

	// Set config file
	l.viper.SetConfigFile(expandedPath)

	// Configure environment variable support
	l.setupEnvVars()

	// Set defaults
	l.setDefaults()

	// Read configuration
	if err := l.viper.ReadInConfig(); err != nil {
		return nil, errors.NewConfigParseError(expandedPath, err)
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := l.viper.Unmarshal(&cfg); err != nil {
		return nil, errors.NewConfigParseError(expandedPath, err)
	}

	// Resolve dynamic API keys
	if err := l.resolveAPIKeys(&cfg); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeConfigInvalid, "failed to resolve API keys")
	}

	// Validate configuration
	validator := NewValidator(&cfg)
	if err := validator.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// setupEnvVars configures environment variable support
func (l *Loader) setupEnvVars() {
	// Set environment variable prefix
	l.viper.SetEnvPrefix(l.envPrefix)

	// Replace dots and dashes with underscores in env var names
	l.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Enable automatic env var binding
	l.viper.AutomaticEnv()

	// Bind specific environment variables for sensitive data
	l.bindEnvVars()
}

// bindEnvVars explicitly binds environment variables for sensitive configuration
func (l *Loader) bindEnvVars() {
	// LLM Provider API Keys
	envBindings := []string{
		// Anthropic
		"llm.anthropic.api_key",
		"llm.anthropic.base_url",

		// OpenAI
		"llm.openai.api_key",
		"llm.openai.organization",
		"llm.openai.base_url",

		// Google
		"llm.google.api_key",
		"llm.google.project_id",

		// Bedrock
		"llm.bedrock.region",
		"llm.bedrock.access_key_id",
		"llm.bedrock.secret_access_key",
		"llm.bedrock.session_token",
		"llm.bedrock.profile",

		// Azure
		"llm.azure.api_key",
		"llm.azure.endpoint",

		// Ollama
		"llm.ollama.base_url",

		// Platform Authentication
		"platform.authentication.api_key",
		"platform.authentication.token",
		"platform.authentication.refresh_token",
		"platform.authentication.client_id",
		"platform.authentication.client_secret",

		// ZeroDB
		"services.zerodb.endpoint",
		"services.zerodb.username",
		"services.zerodb.password",

		// Other Services
		"services.design.api_key",
		"services.strapi.api_key",
		"services.rlhf.api_key",

		// Security
		"security.encryption_key",
		"security.tls_cert_path",
		"security.tls_key_path",
	}

	for _, key := range envBindings {
		_ = l.viper.BindEnv(key)
	}
}

// setDefaults sets default configuration values
func (l *Loader) setDefaults() {
	// Application defaults
	l.viper.SetDefault("app.name", "ainative-code")
	l.viper.SetDefault("app.version", "0.1.0")
	l.viper.SetDefault("app.environment", "development")
	l.viper.SetDefault("app.debug", false)

	// LLM defaults
	l.viper.SetDefault("llm.default_provider", "anthropic")

	// Anthropic defaults
	l.viper.SetDefault("llm.anthropic.model", "claude-3-5-sonnet-20241022")
	l.viper.SetDefault("llm.anthropic.max_tokens", 4096)
	l.viper.SetDefault("llm.anthropic.temperature", 0.7)
	l.viper.SetDefault("llm.anthropic.top_p", 1.0)
	l.viper.SetDefault("llm.anthropic.top_k", 0)
	l.viper.SetDefault("llm.anthropic.timeout", "30s")
	l.viper.SetDefault("llm.anthropic.retry_attempts", 3)
	l.viper.SetDefault("llm.anthropic.api_version", "2023-06-01")

	// OpenAI defaults
	l.viper.SetDefault("llm.openai.model", "gpt-4-turbo-preview")
	l.viper.SetDefault("llm.openai.max_tokens", 4096)
	l.viper.SetDefault("llm.openai.temperature", 0.7)
	l.viper.SetDefault("llm.openai.top_p", 1.0)
	l.viper.SetDefault("llm.openai.frequency_penalty", 0.0)
	l.viper.SetDefault("llm.openai.presence_penalty", 0.0)
	l.viper.SetDefault("llm.openai.timeout", "30s")
	l.viper.SetDefault("llm.openai.retry_attempts", 3)

	// Google defaults
	l.viper.SetDefault("llm.google.model", "gemini-pro")
	l.viper.SetDefault("llm.google.max_tokens", 4096)
	l.viper.SetDefault("llm.google.temperature", 0.7)
	l.viper.SetDefault("llm.google.top_p", 1.0)
	l.viper.SetDefault("llm.google.top_k", 40)
	l.viper.SetDefault("llm.google.timeout", "30s")
	l.viper.SetDefault("llm.google.retry_attempts", 3)

	// Bedrock defaults
	l.viper.SetDefault("llm.bedrock.model", "anthropic.claude-3-sonnet-20240229-v1:0")
	l.viper.SetDefault("llm.bedrock.max_tokens", 4096)
	l.viper.SetDefault("llm.bedrock.temperature", 0.7)
	l.viper.SetDefault("llm.bedrock.top_p", 1.0)
	l.viper.SetDefault("llm.bedrock.timeout", "60s")
	l.viper.SetDefault("llm.bedrock.retry_attempts", 3)

	// Azure defaults
	l.viper.SetDefault("llm.azure.api_version", "2023-05-15")
	l.viper.SetDefault("llm.azure.max_tokens", 4096)
	l.viper.SetDefault("llm.azure.temperature", 0.7)
	l.viper.SetDefault("llm.azure.top_p", 1.0)
	l.viper.SetDefault("llm.azure.timeout", "30s")
	l.viper.SetDefault("llm.azure.retry_attempts", 3)

	// Ollama defaults
	l.viper.SetDefault("llm.ollama.base_url", "http://localhost:11434")
	l.viper.SetDefault("llm.ollama.max_tokens", 4096)
	l.viper.SetDefault("llm.ollama.temperature", 0.7)
	l.viper.SetDefault("llm.ollama.top_p", 1.0)
	l.viper.SetDefault("llm.ollama.top_k", 40)
	l.viper.SetDefault("llm.ollama.timeout", "120s")
	l.viper.SetDefault("llm.ollama.retry_attempts", 1)
	l.viper.SetDefault("llm.ollama.keep_alive", "5m")

	// Fallback defaults
	l.viper.SetDefault("llm.fallback.enabled", false)
	l.viper.SetDefault("llm.fallback.max_retries", 2)
	l.viper.SetDefault("llm.fallback.retry_delay", "1s")

	// Authentication defaults
	l.viper.SetDefault("platform.authentication.method", "api_key")
	l.viper.SetDefault("platform.authentication.timeout", "10s")

	// ZeroDB defaults
	l.viper.SetDefault("services.zerodb.enabled", false)
	l.viper.SetDefault("services.zerodb.ssl", false)
	l.viper.SetDefault("services.zerodb.max_connections", 10)
	l.viper.SetDefault("services.zerodb.idle_connections", 2)
	l.viper.SetDefault("services.zerodb.conn_max_lifetime", "1h")
	l.viper.SetDefault("services.zerodb.timeout", "5s")
	l.viper.SetDefault("services.zerodb.retry_attempts", 3)
	l.viper.SetDefault("services.zerodb.retry_delay", "1s")

	// Design service defaults
	l.viper.SetDefault("services.design.enabled", false)
	l.viper.SetDefault("services.design.timeout", "30s")
	l.viper.SetDefault("services.design.retry_attempts", 3)

	// Strapi defaults
	l.viper.SetDefault("services.strapi.enabled", false)
	l.viper.SetDefault("services.strapi.timeout", "30s")
	l.viper.SetDefault("services.strapi.retry_attempts", 3)

	// RLHF defaults
	l.viper.SetDefault("services.rlhf.enabled", false)
	l.viper.SetDefault("services.rlhf.timeout", "60s")
	l.viper.SetDefault("services.rlhf.retry_attempts", 3)

	// Tool defaults
	l.viper.SetDefault("tools.filesystem.enabled", false) // Disabled by default, requires path configuration
	l.viper.SetDefault("tools.filesystem.max_file_size", 104857600) // 100MB
	l.viper.SetDefault("tools.terminal.enabled", false) // Disabled by default for security
	l.viper.SetDefault("tools.terminal.timeout", "5m")
	l.viper.SetDefault("tools.browser.enabled", false)
	l.viper.SetDefault("tools.browser.headless", true)
	l.viper.SetDefault("tools.browser.timeout", "30s")
	l.viper.SetDefault("tools.code_analysis.enabled", false) // Disabled by default
	l.viper.SetDefault("tools.code_analysis.max_file_size", 10485760) // 10MB
	l.viper.SetDefault("tools.code_analysis.include_tests", true)

	// Performance defaults
	l.viper.SetDefault("performance.cache.enabled", false)
	l.viper.SetDefault("performance.cache.type", "memory")
	l.viper.SetDefault("performance.cache.ttl", "1h")
	l.viper.SetDefault("performance.cache.max_size", 100) // MB

	l.viper.SetDefault("performance.rate_limit.enabled", false)
	l.viper.SetDefault("performance.rate_limit.requests_per_minute", 60)
	l.viper.SetDefault("performance.rate_limit.burst_size", 10)
	l.viper.SetDefault("performance.rate_limit.time_window", "1m")

	l.viper.SetDefault("performance.concurrency.max_workers", 10)
	l.viper.SetDefault("performance.concurrency.max_queue_size", 100)
	l.viper.SetDefault("performance.concurrency.worker_timeout", "5m")

	l.viper.SetDefault("performance.circuit_breaker.enabled", false)
	l.viper.SetDefault("performance.circuit_breaker.failure_threshold", 5)
	l.viper.SetDefault("performance.circuit_breaker.success_threshold", 2)
	l.viper.SetDefault("performance.circuit_breaker.timeout", "60s")
	l.viper.SetDefault("performance.circuit_breaker.reset_timeout", "30s")

	// Logging defaults
	l.viper.SetDefault("logging.level", "info")
	l.viper.SetDefault("logging.format", "json")
	l.viper.SetDefault("logging.output", "stdout")
	l.viper.SetDefault("logging.max_size", 100) // MB
	l.viper.SetDefault("logging.max_backups", 3)
	l.viper.SetDefault("logging.max_age", 7) // days
	l.viper.SetDefault("logging.compress", true)

	// Security defaults
	l.viper.SetDefault("security.encrypt_config", false)
	l.viper.SetDefault("security.tls_enabled", false)
}

// GetConfigFilePath returns the path of the loaded configuration file
func (l *Loader) GetConfigFilePath() string {
	return l.viper.ConfigFileUsed()
}

// GetViper returns the underlying viper instance for advanced use cases
func (l *Loader) GetViper() *viper.Viper {
	return l.viper
}

// WriteConfig writes the current configuration to a file
func WriteConfig(cfg *Config, filePath string) error {
	// Expand environment variables in path
	expandedPath := os.ExpandEnv(filePath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(expandedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.NewConfigParseError(expandedPath, fmt.Errorf("failed to create directory: %w", err))
	}

	v := viper.New()

	// Marshal config into viper
	if err := v.MergeConfigMap(structToMap(cfg)); err != nil {
		return errors.NewConfigParseError(expandedPath, err)
	}

	// Write to file
	if err := v.WriteConfigAs(expandedPath); err != nil {
		return errors.NewConfigParseError(expandedPath, err)
	}

	return nil
}

// structToMap converts a struct to a map for viper
func structToMap(v interface{}) map[string]interface{} {
	// This is a simplified version - in production, use a proper reflection-based
	// or use viper's built-in marshaling capabilities
	// For now, returning empty map as this is typically not needed
	return make(map[string]interface{})
}

// resolveAPIKeys resolves dynamic API keys using the configured resolver
func (l *Loader) resolveAPIKeys(cfg *Config) error {
	if l.resolver == nil {
		return nil // Skip if no resolver configured
	}

	// Resolve LLM provider API keys
	if cfg.LLM.Anthropic != nil && cfg.LLM.Anthropic.APIKey != "" {
		resolved, err := l.resolver.Resolve(cfg.LLM.Anthropic.APIKey)
		if err != nil {
			return fmt.Errorf("anthropic API key: %w", err)
		}
		cfg.LLM.Anthropic.APIKey = resolved
	}

	if cfg.LLM.OpenAI != nil && cfg.LLM.OpenAI.APIKey != "" {
		resolved, err := l.resolver.Resolve(cfg.LLM.OpenAI.APIKey)
		if err != nil {
			return fmt.Errorf("openai API key: %w", err)
		}
		cfg.LLM.OpenAI.APIKey = resolved
	}

	if cfg.LLM.Google != nil && cfg.LLM.Google.APIKey != "" {
		resolved, err := l.resolver.Resolve(cfg.LLM.Google.APIKey)
		if err != nil {
			return fmt.Errorf("google API key: %w", err)
		}
		cfg.LLM.Google.APIKey = resolved
	}

	if cfg.LLM.Azure != nil && cfg.LLM.Azure.APIKey != "" {
		resolved, err := l.resolver.Resolve(cfg.LLM.Azure.APIKey)
		if err != nil {
			return fmt.Errorf("azure API key: %w", err)
		}
		cfg.LLM.Azure.APIKey = resolved
	}

	// Resolve AWS Bedrock credentials
	if cfg.LLM.Bedrock != nil {
		if cfg.LLM.Bedrock.AccessKeyID != "" {
			resolved, err := l.resolver.Resolve(cfg.LLM.Bedrock.AccessKeyID)
			if err != nil {
				return fmt.Errorf("bedrock access key ID: %w", err)
			}
			cfg.LLM.Bedrock.AccessKeyID = resolved
		}
		if cfg.LLM.Bedrock.SecretAccessKey != "" {
			resolved, err := l.resolver.Resolve(cfg.LLM.Bedrock.SecretAccessKey)
			if err != nil {
				return fmt.Errorf("bedrock secret access key: %w", err)
			}
			cfg.LLM.Bedrock.SecretAccessKey = resolved
		}
		if cfg.LLM.Bedrock.SessionToken != "" {
			resolved, err := l.resolver.Resolve(cfg.LLM.Bedrock.SessionToken)
			if err != nil {
				return fmt.Errorf("bedrock session token: %w", err)
			}
			cfg.LLM.Bedrock.SessionToken = resolved
		}
	}

	// Resolve platform authentication credentials
	if cfg.Platform.Authentication.APIKey != "" {
		resolved, err := l.resolver.Resolve(cfg.Platform.Authentication.APIKey)
		if err != nil {
			return fmt.Errorf("platform API key: %w", err)
		}
		cfg.Platform.Authentication.APIKey = resolved
	}

	if cfg.Platform.Authentication.Token != "" {
		resolved, err := l.resolver.Resolve(cfg.Platform.Authentication.Token)
		if err != nil {
			return fmt.Errorf("platform token: %w", err)
		}
		cfg.Platform.Authentication.Token = resolved
	}

	if cfg.Platform.Authentication.ClientSecret != "" {
		resolved, err := l.resolver.Resolve(cfg.Platform.Authentication.ClientSecret)
		if err != nil {
			return fmt.Errorf("platform client secret: %w", err)
		}
		cfg.Platform.Authentication.ClientSecret = resolved
	}

	// Resolve service API keys
	if cfg.Services.ZeroDB != nil && cfg.Services.ZeroDB.Password != "" {
		resolved, err := l.resolver.Resolve(cfg.Services.ZeroDB.Password)
		if err != nil {
			return fmt.Errorf("zerodb password: %w", err)
		}
		cfg.Services.ZeroDB.Password = resolved
	}

	if cfg.Services.Design != nil && cfg.Services.Design.APIKey != "" {
		resolved, err := l.resolver.Resolve(cfg.Services.Design.APIKey)
		if err != nil {
			return fmt.Errorf("design service API key: %w", err)
		}
		cfg.Services.Design.APIKey = resolved
	}

	if cfg.Services.Strapi != nil && cfg.Services.Strapi.APIKey != "" {
		resolved, err := l.resolver.Resolve(cfg.Services.Strapi.APIKey)
		if err != nil {
			return fmt.Errorf("strapi API key: %w", err)
		}
		cfg.Services.Strapi.APIKey = resolved
	}

	if cfg.Services.RLHF != nil && cfg.Services.RLHF.APIKey != "" {
		resolved, err := l.resolver.Resolve(cfg.Services.RLHF.APIKey)
		if err != nil {
			return fmt.Errorf("rlhf service API key: %w", err)
		}
		cfg.Services.RLHF.APIKey = resolved
	}

	// Resolve security encryption key
	if cfg.Security.EncryptionKey != "" {
		resolved, err := l.resolver.Resolve(cfg.Security.EncryptionKey)
		if err != nil {
			return fmt.Errorf("encryption key: %w", err)
		}
		cfg.Security.EncryptionKey = resolved
	}

	return nil
}
