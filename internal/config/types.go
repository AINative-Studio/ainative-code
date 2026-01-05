package config

import (
	"time"
)

// Config represents the complete application configuration
type Config struct {
	// Application settings
	App AppConfig `mapstructure:"app" yaml:"app"`

	// LLM Provider configurations
	LLM LLMConfig `mapstructure:"llm" yaml:"llm"`

	// AINative platform configuration
	Platform PlatformConfig `mapstructure:"platform" yaml:"platform"`

	// Service endpoints
	Services ServicesConfig `mapstructure:"services" yaml:"services"`

	// Tool configurations
	Tools ToolsConfig `mapstructure:"tools" yaml:"tools"`

	// Performance settings
	Performance PerformanceConfig `mapstructure:"performance" yaml:"performance"`

	// Logging configuration
	Logging LoggingConfig `mapstructure:"logging" yaml:"logging"`

	// Security settings
	Security SecurityConfig `mapstructure:"security" yaml:"security"`
}

// AppConfig contains general application settings
type AppConfig struct {
	Name        string `mapstructure:"name" yaml:"name"`
	Version     string `mapstructure:"version" yaml:"version"`
	Environment string `mapstructure:"environment" yaml:"environment"` // development, staging, production
	Debug       bool   `mapstructure:"debug" yaml:"debug"`
}

// LLMConfig contains all LLM provider configurations
type LLMConfig struct {
	DefaultProvider string              `mapstructure:"default_provider" yaml:"default_provider"`
	Anthropic       *AnthropicConfig    `mapstructure:"anthropic,omitempty" yaml:"anthropic,omitempty"`
	OpenAI          *OpenAIConfig       `mapstructure:"openai,omitempty" yaml:"openai,omitempty"`
	Google          *GoogleConfig       `mapstructure:"google,omitempty" yaml:"google,omitempty"`
	Bedrock         *BedrockConfig      `mapstructure:"bedrock,omitempty" yaml:"bedrock,omitempty"`
	Azure           *AzureConfig        `mapstructure:"azure,omitempty" yaml:"azure,omitempty"`
	Ollama          *OllamaConfig       `mapstructure:"ollama,omitempty" yaml:"ollama,omitempty"`
	Fallback        *FallbackConfig     `mapstructure:"fallback,omitempty" yaml:"fallback,omitempty"`
}

// AnthropicConfig contains Anthropic Claude configuration
type AnthropicConfig struct {
	APIKey          string        `mapstructure:"api_key" yaml:"api_key"`
	Model           string        `mapstructure:"model" yaml:"model"`
	MaxTokens       int           `mapstructure:"max_tokens" yaml:"max_tokens"`
	Temperature     float64       `mapstructure:"temperature" yaml:"temperature"`
	TopP            float64       `mapstructure:"top_p" yaml:"top_p"`
	TopK            int           `mapstructure:"top_k" yaml:"top_k"`
	Timeout         time.Duration `mapstructure:"timeout" yaml:"timeout"`
	RetryAttempts   int           `mapstructure:"retry_attempts" yaml:"retry_attempts"`
	BaseURL         string        `mapstructure:"base_url,omitempty" yaml:"base_url,omitempty"`
	APIVersion      string        `mapstructure:"api_version" yaml:"api_version"`
	ExtendedThinking *ExtendedThinkingConfig `mapstructure:"extended_thinking,omitempty" yaml:"extended_thinking,omitempty"`
	Retry           *RetryConfig  `mapstructure:"retry,omitempty" yaml:"retry,omitempty"`
}

// ExtendedThinkingConfig contains extended thinking visualization settings
type ExtendedThinkingConfig struct {
	Enabled    bool `mapstructure:"enabled" yaml:"enabled"`
	AutoExpand bool `mapstructure:"auto_expand" yaml:"auto_expand"`
	MaxDepth   int  `mapstructure:"max_depth" yaml:"max_depth"`
}

// RetryConfig contains advanced retry and error recovery configuration for LLM providers
type RetryConfig struct {
	// Basic retry settings (backward compatible)
	MaxAttempts  int           `mapstructure:"max_attempts" yaml:"max_attempts"`
	InitialDelay time.Duration `mapstructure:"initial_delay" yaml:"initial_delay"`
	MaxDelay     time.Duration `mapstructure:"max_delay" yaml:"max_delay"`
	Multiplier   float64       `mapstructure:"multiplier" yaml:"multiplier"`

	// Advanced recovery settings
	EnableJitter           bool `mapstructure:"enable_jitter" yaml:"enable_jitter"`
	EnableAPIKeyResolution bool `mapstructure:"enable_api_key_resolution" yaml:"enable_api_key_resolution"`
	EnableTokenReduction   bool `mapstructure:"enable_token_reduction" yaml:"enable_token_reduction"`
	TokenReductionPercent  int  `mapstructure:"token_reduction_percent" yaml:"token_reduction_percent"`
	EnableTimeoutIncrease  bool `mapstructure:"enable_timeout_increase" yaml:"enable_timeout_increase"`
	TimeoutIncreasePercent int  `mapstructure:"timeout_increase_percent" yaml:"timeout_increase_percent"`
}

// OpenAIConfig contains OpenAI configuration
type OpenAIConfig struct {
	APIKey          string        `mapstructure:"api_key" yaml:"api_key"`
	Model           string        `mapstructure:"model" yaml:"model"`
	Organization    string        `mapstructure:"organization,omitempty" yaml:"organization,omitempty"`
	MaxTokens       int           `mapstructure:"max_tokens" yaml:"max_tokens"`
	Temperature     float64       `mapstructure:"temperature" yaml:"temperature"`
	TopP            float64       `mapstructure:"top_p" yaml:"top_p"`
	FrequencyPenalty float64      `mapstructure:"frequency_penalty" yaml:"frequency_penalty"`
	PresencePenalty float64       `mapstructure:"presence_penalty" yaml:"presence_penalty"`
	Timeout         time.Duration `mapstructure:"timeout" yaml:"timeout"`
	RetryAttempts   int           `mapstructure:"retry_attempts" yaml:"retry_attempts"`
	BaseURL         string        `mapstructure:"base_url,omitempty" yaml:"base_url,omitempty"`
	Retry           *RetryConfig  `mapstructure:"retry,omitempty" yaml:"retry,omitempty"`
}

// GoogleConfig contains Google (Gemini) configuration
type GoogleConfig struct {
	APIKey          string        `mapstructure:"api_key" yaml:"api_key"`
	Model           string        `mapstructure:"model" yaml:"model"`
	ProjectID       string        `mapstructure:"project_id,omitempty" yaml:"project_id,omitempty"`
	Location        string        `mapstructure:"location,omitempty" yaml:"location,omitempty"`
	MaxTokens       int           `mapstructure:"max_tokens" yaml:"max_tokens"`
	Temperature     float64       `mapstructure:"temperature" yaml:"temperature"`
	TopP            float64       `mapstructure:"top_p" yaml:"top_p"`
	TopK            int           `mapstructure:"top_k" yaml:"top_k"`
	Timeout         time.Duration `mapstructure:"timeout" yaml:"timeout"`
	RetryAttempts   int           `mapstructure:"retry_attempts" yaml:"retry_attempts"`
}

// BedrockConfig contains AWS Bedrock configuration
type BedrockConfig struct {
	Region          string        `mapstructure:"region" yaml:"region"`
	Model           string        `mapstructure:"model" yaml:"model"`
	AccessKeyID     string        `mapstructure:"access_key_id,omitempty" yaml:"access_key_id,omitempty"`
	SecretAccessKey string        `mapstructure:"secret_access_key,omitempty" yaml:"secret_access_key,omitempty"`
	SessionToken    string        `mapstructure:"session_token,omitempty" yaml:"session_token,omitempty"`
	Profile         string        `mapstructure:"profile,omitempty" yaml:"profile,omitempty"`
	MaxTokens       int           `mapstructure:"max_tokens" yaml:"max_tokens"`
	Temperature     float64       `mapstructure:"temperature" yaml:"temperature"`
	TopP            float64       `mapstructure:"top_p" yaml:"top_p"`
	Timeout         time.Duration `mapstructure:"timeout" yaml:"timeout"`
	RetryAttempts   int           `mapstructure:"retry_attempts" yaml:"retry_attempts"`
}

// AzureConfig contains Azure OpenAI configuration
type AzureConfig struct {
	APIKey          string        `mapstructure:"api_key" yaml:"api_key"`
	Endpoint        string        `mapstructure:"endpoint" yaml:"endpoint"`
	DeploymentName  string        `mapstructure:"deployment_name" yaml:"deployment_name"`
	APIVersion      string        `mapstructure:"api_version" yaml:"api_version"`
	MaxTokens       int           `mapstructure:"max_tokens" yaml:"max_tokens"`
	Temperature     float64       `mapstructure:"temperature" yaml:"temperature"`
	TopP            float64       `mapstructure:"top_p" yaml:"top_p"`
	Timeout         time.Duration `mapstructure:"timeout" yaml:"timeout"`
	RetryAttempts   int           `mapstructure:"retry_attempts" yaml:"retry_attempts"`
}

// OllamaConfig contains Ollama (local LLM) configuration
type OllamaConfig struct {
	BaseURL         string        `mapstructure:"base_url" yaml:"base_url"`
	Model           string        `mapstructure:"model" yaml:"model"`
	MaxTokens       int           `mapstructure:"max_tokens" yaml:"max_tokens"`
	Temperature     float64       `mapstructure:"temperature" yaml:"temperature"`
	TopP            float64       `mapstructure:"top_p" yaml:"top_p"`
	TopK            int           `mapstructure:"top_k" yaml:"top_k"`
	Timeout         time.Duration `mapstructure:"timeout" yaml:"timeout"`
	RetryAttempts   int           `mapstructure:"retry_attempts" yaml:"retry_attempts"`
	KeepAlive       string        `mapstructure:"keep_alive" yaml:"keep_alive"`
}

// FallbackConfig defines fallback provider configuration
type FallbackConfig struct {
	Enabled       bool     `mapstructure:"enabled" yaml:"enabled"`
	Providers     []string `mapstructure:"providers" yaml:"providers"` // ordered list of fallback providers
	MaxRetries    int      `mapstructure:"max_retries" yaml:"max_retries"`
	RetryDelay    time.Duration `mapstructure:"retry_delay" yaml:"retry_delay"`
}

// PlatformConfig contains AINative platform settings
type PlatformConfig struct {
	Authentication AuthConfig `mapstructure:"authentication" yaml:"authentication"`
	Organization   OrgConfig  `mapstructure:"organization,omitempty" yaml:"organization,omitempty"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Method       string        `mapstructure:"method" yaml:"method"` // jwt, api_key, oauth2
	APIKey       string        `mapstructure:"api_key,omitempty" yaml:"api_key,omitempty"`
	Token        string        `mapstructure:"token,omitempty" yaml:"token,omitempty"`
	RefreshToken string        `mapstructure:"refresh_token,omitempty" yaml:"refresh_token,omitempty"`
	ClientID     string        `mapstructure:"client_id,omitempty" yaml:"client_id,omitempty"`
	ClientSecret string        `mapstructure:"client_secret,omitempty" yaml:"client_secret,omitempty"`
	TokenURL     string        `mapstructure:"token_url,omitempty" yaml:"token_url,omitempty"`
	Scopes       []string      `mapstructure:"scopes,omitempty" yaml:"scopes,omitempty"`
	Timeout      time.Duration `mapstructure:"timeout" yaml:"timeout"`
}

// OrgConfig contains organization settings
type OrgConfig struct {
	ID          string `mapstructure:"id" yaml:"id"`
	Name        string `mapstructure:"name,omitempty" yaml:"name,omitempty"`
	Workspace   string `mapstructure:"workspace,omitempty" yaml:"workspace,omitempty"`
}

// ServicesConfig contains service endpoint configurations
type ServicesConfig struct {
	ZeroDB *ZeroDBConfig `mapstructure:"zerodb,omitempty" yaml:"zerodb,omitempty"`
	Design *DesignConfig `mapstructure:"design,omitempty" yaml:"design,omitempty"`
	Strapi *StrapiConfig `mapstructure:"strapi,omitempty" yaml:"strapi,omitempty"`
	RLHF   *RLHFConfig   `mapstructure:"rlhf,omitempty" yaml:"rlhf,omitempty"`
}

// ZeroDBConfig contains ZeroDB connection settings
type ZeroDBConfig struct {
	Enabled         bool          `mapstructure:"enabled" yaml:"enabled"`
	Endpoint        string        `mapstructure:"endpoint" yaml:"endpoint"`
	Database        string        `mapstructure:"database" yaml:"database"`
	Username        string        `mapstructure:"username,omitempty" yaml:"username,omitempty"`
	Password        string        `mapstructure:"password,omitempty" yaml:"password,omitempty"`
	SSL             bool          `mapstructure:"ssl" yaml:"ssl"`
	SSLMode         string        `mapstructure:"ssl_mode,omitempty" yaml:"ssl_mode,omitempty"`
	MaxConnections  int           `mapstructure:"max_connections" yaml:"max_connections"`
	IdleConnections int           `mapstructure:"idle_connections" yaml:"idle_connections"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	Timeout         time.Duration `mapstructure:"timeout" yaml:"timeout"`
	RetryAttempts   int           `mapstructure:"retry_attempts" yaml:"retry_attempts"`
	RetryDelay      time.Duration `mapstructure:"retry_delay" yaml:"retry_delay"`
}

// DesignConfig contains AINative Design service settings
type DesignConfig struct {
	Enabled       bool          `mapstructure:"enabled" yaml:"enabled"`
	Endpoint      string        `mapstructure:"endpoint" yaml:"endpoint"`
	APIKey        string        `mapstructure:"api_key,omitempty" yaml:"api_key,omitempty"`
	Timeout       time.Duration `mapstructure:"timeout" yaml:"timeout"`
	RetryAttempts int           `mapstructure:"retry_attempts" yaml:"retry_attempts"`
}

// StrapiConfig contains Strapi CMS settings
type StrapiConfig struct {
	Enabled       bool          `mapstructure:"enabled" yaml:"enabled"`
	Endpoint      string        `mapstructure:"endpoint" yaml:"endpoint"`
	APIKey        string        `mapstructure:"api_key,omitempty" yaml:"api_key,omitempty"`
	Timeout       time.Duration `mapstructure:"timeout" yaml:"timeout"`
	RetryAttempts int           `mapstructure:"retry_attempts" yaml:"retry_attempts"`
}

// RLHFConfig contains RLHF service settings
type RLHFConfig struct {
	Enabled       bool          `mapstructure:"enabled" yaml:"enabled"`
	Endpoint      string        `mapstructure:"endpoint" yaml:"endpoint"`
	APIKey        string        `mapstructure:"api_key,omitempty" yaml:"api_key,omitempty"`
	Timeout       time.Duration `mapstructure:"timeout" yaml:"timeout"`
	RetryAttempts int           `mapstructure:"retry_attempts" yaml:"retry_attempts"`
	ModelID       string        `mapstructure:"model_id,omitempty" yaml:"model_id,omitempty"`
}

// ToolsConfig contains tool-specific configurations
type ToolsConfig struct {
	FileSystem   *FileSystemToolConfig   `mapstructure:"filesystem,omitempty" yaml:"filesystem,omitempty"`
	Terminal     *TerminalToolConfig     `mapstructure:"terminal,omitempty" yaml:"terminal,omitempty"`
	Browser      *BrowserToolConfig      `mapstructure:"browser,omitempty" yaml:"browser,omitempty"`
	CodeAnalysis *CodeAnalysisToolConfig `mapstructure:"code_analysis,omitempty" yaml:"code_analysis,omitempty"`
}

// FileSystemToolConfig contains filesystem tool settings
type FileSystemToolConfig struct {
	Enabled        bool     `mapstructure:"enabled" yaml:"enabled"`
	AllowedPaths   []string `mapstructure:"allowed_paths" yaml:"allowed_paths"`
	BlockedPaths   []string `mapstructure:"blocked_paths" yaml:"blocked_paths"`
	MaxFileSize    int64    `mapstructure:"max_file_size" yaml:"max_file_size"` // bytes
	AllowedExtensions []string `mapstructure:"allowed_extensions,omitempty" yaml:"allowed_extensions,omitempty"`
}

// TerminalToolConfig contains terminal tool settings
type TerminalToolConfig struct {
	Enabled         bool     `mapstructure:"enabled" yaml:"enabled"`
	AllowedCommands []string `mapstructure:"allowed_commands" yaml:"allowed_commands"`
	BlockedCommands []string `mapstructure:"blocked_commands" yaml:"blocked_commands"`
	Timeout         time.Duration `mapstructure:"timeout" yaml:"timeout"`
	WorkingDir      string   `mapstructure:"working_dir,omitempty" yaml:"working_dir,omitempty"`
}

// BrowserToolConfig contains browser automation tool settings
type BrowserToolConfig struct {
	Enabled    bool          `mapstructure:"enabled" yaml:"enabled"`
	Headless   bool          `mapstructure:"headless" yaml:"headless"`
	Timeout    time.Duration `mapstructure:"timeout" yaml:"timeout"`
	UserAgent  string        `mapstructure:"user_agent,omitempty" yaml:"user_agent,omitempty"`
}

// CodeAnalysisToolConfig contains code analysis tool settings
type CodeAnalysisToolConfig struct {
	Enabled        bool     `mapstructure:"enabled" yaml:"enabled"`
	Languages      []string `mapstructure:"languages" yaml:"languages"`
	MaxFileSize    int64    `mapstructure:"max_file_size" yaml:"max_file_size"`
	IncludeTests   bool     `mapstructure:"include_tests" yaml:"include_tests"`
}

// PerformanceConfig contains performance-related settings
type PerformanceConfig struct {
	Cache         CacheConfig         `mapstructure:"cache" yaml:"cache"`
	RateLimit     RateLimitConfig     `mapstructure:"rate_limit" yaml:"rate_limit"`
	Concurrency   ConcurrencyConfig   `mapstructure:"concurrency" yaml:"concurrency"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker" yaml:"circuit_breaker"`
}

// CacheConfig contains caching settings
type CacheConfig struct {
	Enabled        bool          `mapstructure:"enabled" yaml:"enabled"`
	Type           string        `mapstructure:"type" yaml:"type"` // memory, redis, memcached
	TTL            time.Duration `mapstructure:"ttl" yaml:"ttl"`
	MaxSize        int64         `mapstructure:"max_size" yaml:"max_size"` // MB
	RedisURL       string        `mapstructure:"redis_url,omitempty" yaml:"redis_url,omitempty"`
	MemcachedURL   string        `mapstructure:"memcached_url,omitempty" yaml:"memcached_url,omitempty"`
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Enabled           bool              `mapstructure:"enabled" yaml:"enabled"`
	RequestsPerMinute int               `mapstructure:"requests_per_minute" yaml:"requests_per_minute"`
	BurstSize         int               `mapstructure:"burst_size" yaml:"burst_size"`
	TimeWindow        time.Duration     `mapstructure:"time_window" yaml:"time_window"`
	PerUser           bool              `mapstructure:"per_user" yaml:"per_user"`
	PerEndpoint       bool              `mapstructure:"per_endpoint" yaml:"per_endpoint"`
	Storage           string            `mapstructure:"storage" yaml:"storage"` // memory, redis
	RedisURL          string            `mapstructure:"redis_url,omitempty" yaml:"redis_url,omitempty"`
	EndpointLimits    map[string]int    `mapstructure:"endpoint_limits,omitempty" yaml:"endpoint_limits,omitempty"`
	SkipPaths         []string          `mapstructure:"skip_paths,omitempty" yaml:"skip_paths,omitempty"`
	IPAllowlist       []string          `mapstructure:"ip_allowlist,omitempty" yaml:"ip_allowlist,omitempty"`
	IPBlocklist       []string          `mapstructure:"ip_blocklist,omitempty" yaml:"ip_blocklist,omitempty"`
}

// ConcurrencyConfig contains concurrency settings
type ConcurrencyConfig struct {
	MaxWorkers      int `mapstructure:"max_workers" yaml:"max_workers"`
	MaxQueueSize    int `mapstructure:"max_queue_size" yaml:"max_queue_size"`
	WorkerTimeout   time.Duration `mapstructure:"worker_timeout" yaml:"worker_timeout"`
}

// CircuitBreakerConfig contains circuit breaker settings
type CircuitBreakerConfig struct {
	Enabled           bool          `mapstructure:"enabled" yaml:"enabled"`
	FailureThreshold  int           `mapstructure:"failure_threshold" yaml:"failure_threshold"`
	SuccessThreshold  int           `mapstructure:"success_threshold" yaml:"success_threshold"`
	Timeout           time.Duration `mapstructure:"timeout" yaml:"timeout"`
	ResetTimeout      time.Duration `mapstructure:"reset_timeout" yaml:"reset_timeout"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level         string `mapstructure:"level" yaml:"level"` // debug, info, warn, error
	Format        string `mapstructure:"format" yaml:"format"` // json, console
	Output        string `mapstructure:"output" yaml:"output"` // stdout, file
	FilePath      string `mapstructure:"file_path,omitempty" yaml:"file_path,omitempty"`
	MaxSize       int    `mapstructure:"max_size" yaml:"max_size"` // MB
	MaxBackups    int    `mapstructure:"max_backups" yaml:"max_backups"`
	MaxAge        int    `mapstructure:"max_age" yaml:"max_age"` // days
	Compress      bool   `mapstructure:"compress" yaml:"compress"`
	SensitiveKeys []string `mapstructure:"sensitive_keys,omitempty" yaml:"sensitive_keys,omitempty"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	EncryptConfig    bool     `mapstructure:"encrypt_config" yaml:"encrypt_config"`
	EncryptionKey    string   `mapstructure:"encryption_key,omitempty" yaml:"encryption_key,omitempty"`
	AllowedOrigins   []string `mapstructure:"allowed_origins" yaml:"allowed_origins"`
	TLSEnabled       bool     `mapstructure:"tls_enabled" yaml:"tls_enabled"`
	TLSCertPath      string   `mapstructure:"tls_cert_path,omitempty" yaml:"tls_cert_path,omitempty"`
	TLSKeyPath       string   `mapstructure:"tls_key_path,omitempty" yaml:"tls_key_path,omitempty"`
	SecretRotation   time.Duration `mapstructure:"secret_rotation,omitempty" yaml:"secret_rotation,omitempty"`
}
