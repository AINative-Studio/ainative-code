package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/errors"
)

// Validator provides configuration validation functionality
type Validator struct {
	config *Config
	errs   []error
}

// NewValidator creates a new configuration validator
func NewValidator(cfg *Config) *Validator {
	return &Validator{
		config: cfg,
		errs:   make([]error, 0),
	}
}

// Validate performs comprehensive validation of the configuration
func (v *Validator) Validate() error {
	v.validateApp()
	v.validateLLM()
	v.validatePlatform()
	v.validateServices()
	v.validateTools()
	v.validatePerformance()
	v.validateLogging()
	v.validateSecurity()

	if len(v.errs) > 0 {
		return v.createValidationError()
	}

	return nil
}

// validateApp validates application configuration
func (v *Validator) validateApp() {
	if v.config.App.Name == "" {
		v.addError("app.name", "application name is required")
	}

	if v.config.App.Environment == "" {
		v.config.App.Environment = "development" // default
	}

	validEnvs := []string{"development", "staging", "production"}
	if !v.isValidEnum(v.config.App.Environment, validEnvs) {
		v.addError("app.environment", fmt.Sprintf("must be one of: %s", strings.Join(validEnvs, ", ")))
	}
}

// validateLLM validates LLM provider configurations
func (v *Validator) validateLLM() {
	if v.config.LLM.DefaultProvider == "" {
		v.addError("llm.default_provider", "default LLM provider must be specified")
		return
	}

	validProviders := []string{"anthropic", "openai", "google", "bedrock", "azure", "ollama"}
	if !v.isValidEnum(v.config.LLM.DefaultProvider, validProviders) {
		v.addError("llm.default_provider", fmt.Sprintf("must be one of: %s", strings.Join(validProviders, ", ")))
	}

	// Validate the default provider is configured
	switch v.config.LLM.DefaultProvider {
	case "anthropic":
		if v.config.LLM.Anthropic == nil {
			v.addError("llm.anthropic", "default provider 'anthropic' is not configured")
		} else {
			v.validateAnthropic()
		}
	case "openai":
		if v.config.LLM.OpenAI == nil {
			v.addError("llm.openai", "default provider 'openai' is not configured")
		} else {
			v.validateOpenAI()
		}
	case "google":
		if v.config.LLM.Google == nil {
			v.addError("llm.google", "default provider 'google' is not configured")
		} else {
			v.validateGoogle()
		}
	case "bedrock":
		if v.config.LLM.Bedrock == nil {
			v.addError("llm.bedrock", "default provider 'bedrock' is not configured")
		} else {
			v.validateBedrock()
		}
	case "azure":
		if v.config.LLM.Azure == nil {
			v.addError("llm.azure", "default provider 'azure' is not configured")
		} else {
			v.validateAzure()
		}
	case "ollama":
		if v.config.LLM.Ollama == nil {
			v.addError("llm.ollama", "default provider 'ollama' is not configured")
		} else {
			v.validateOllama()
		}
	case "meta_llama", "meta":
		if v.config.LLM.MetaLlama == nil {
			v.addError("llm.meta_llama", "default provider 'meta_llama' is not configured")
		} else {
			v.validateMetaLlama()
		}
	}

	// Validate fallback configuration if enabled
	if v.config.LLM.Fallback != nil && v.config.LLM.Fallback.Enabled {
		v.validateFallback()
	}
}

// validateAnthropic validates Anthropic configuration
func (v *Validator) validateAnthropic() {
	cfg := v.config.LLM.Anthropic

	if cfg.APIKey == "" {
		v.addError("llm.anthropic.api_key", "Anthropic API key is required")
	}

	if cfg.Model == "" {
		cfg.Model = "claude-3-5-sonnet-20241022" // default
	}

	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 4096 // default
	}

	if cfg.Temperature < 0 || cfg.Temperature > 1 {
		v.addError("llm.anthropic.temperature", "must be between 0 and 1")
	}

	if cfg.TopP < 0 || cfg.TopP > 1 {
		v.addError("llm.anthropic.top_p", "must be between 0 and 1")
	}

	if cfg.APIVersion == "" {
		cfg.APIVersion = "2023-06-01" // default
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30000000000 // 30 seconds default
	}

	if cfg.RetryAttempts < 0 {
		cfg.RetryAttempts = 3 // default
	}

	if cfg.BaseURL != "" && !v.isValidURL(cfg.BaseURL) {
		v.addError("llm.anthropic.base_url", "must be a valid URL")
	}

	// Validate extended thinking configuration if present
	if cfg.ExtendedThinking != nil {
		if err := ValidateExtendedThinkingConfig(cfg.ExtendedThinking); err != nil {
			v.errs = append(v.errs, err)
		}
	}
}

// validateOpenAI validates OpenAI configuration
func (v *Validator) validateOpenAI() {
	cfg := v.config.LLM.OpenAI

	if cfg.APIKey == "" {
		v.addError("llm.openai.api_key", "OpenAI API key is required")
	}

	if cfg.Model == "" {
		cfg.Model = "gpt-4-turbo-preview" // default
	}

	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 4096 // default
	}

	if cfg.Temperature < 0 || cfg.Temperature > 2 {
		v.addError("llm.openai.temperature", "must be between 0 and 2")
	}

	if cfg.TopP < 0 || cfg.TopP > 1 {
		v.addError("llm.openai.top_p", "must be between 0 and 1")
	}

	if cfg.FrequencyPenalty < -2 || cfg.FrequencyPenalty > 2 {
		v.addError("llm.openai.frequency_penalty", "must be between -2 and 2")
	}

	if cfg.PresencePenalty < -2 || cfg.PresencePenalty > 2 {
		v.addError("llm.openai.presence_penalty", "must be between -2 and 2")
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30000000000 // 30 seconds default
	}

	if cfg.RetryAttempts < 0 {
		cfg.RetryAttempts = 3 // default
	}

	if cfg.BaseURL != "" && !v.isValidURL(cfg.BaseURL) {
		v.addError("llm.openai.base_url", "must be a valid URL")
	}
}

// validateGoogle validates Google (Gemini) configuration
func (v *Validator) validateGoogle() {
	cfg := v.config.LLM.Google

	if cfg.APIKey == "" {
		v.addError("llm.google.api_key", "Google API key is required")
	}

	if cfg.Model == "" {
		cfg.Model = "gemini-pro" // default
	}

	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 4096 // default
	}

	if cfg.Temperature < 0 || cfg.Temperature > 1 {
		v.addError("llm.google.temperature", "must be between 0 and 1")
	}

	if cfg.TopP < 0 || cfg.TopP > 1 {
		v.addError("llm.google.top_p", "must be between 0 and 1")
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30000000000 // 30 seconds default
	}

	if cfg.RetryAttempts < 0 {
		cfg.RetryAttempts = 3 // default
	}
}

// validateBedrock validates AWS Bedrock configuration
func (v *Validator) validateBedrock() {
	cfg := v.config.LLM.Bedrock

	if cfg.Region == "" {
		v.addError("llm.bedrock.region", "AWS region is required")
	}

	if cfg.Model == "" {
		cfg.Model = "anthropic.claude-3-sonnet-20240229-v1:0" // default
	}

	// Either use credentials or profile, not both
	hasCredentials := cfg.AccessKeyID != "" && cfg.SecretAccessKey != ""
	hasProfile := cfg.Profile != ""

	if !hasCredentials && !hasProfile {
		v.addError("llm.bedrock", "either credentials (access_key_id and secret_access_key) or profile must be provided")
	}

	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 4096 // default
	}

	if cfg.Temperature < 0 || cfg.Temperature > 1 {
		v.addError("llm.bedrock.temperature", "must be between 0 and 1")
	}

	if cfg.TopP < 0 || cfg.TopP > 1 {
		v.addError("llm.bedrock.top_p", "must be between 0 and 1")
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 60000000000 // 60 seconds default (Bedrock can be slower)
	}

	if cfg.RetryAttempts < 0 {
		cfg.RetryAttempts = 3 // default
	}
}

// validateAzure validates Azure OpenAI configuration
func (v *Validator) validateAzure() {
	cfg := v.config.LLM.Azure

	if cfg.APIKey == "" {
		v.addError("llm.azure.api_key", "Azure API key is required")
	}

	if cfg.Endpoint == "" {
		v.addError("llm.azure.endpoint", "Azure endpoint is required")
	} else if !v.isValidURL(cfg.Endpoint) {
		v.addError("llm.azure.endpoint", "must be a valid URL")
	}

	if cfg.DeploymentName == "" {
		v.addError("llm.azure.deployment_name", "deployment name is required")
	}

	if cfg.APIVersion == "" {
		cfg.APIVersion = "2023-05-15" // default
	}

	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 4096 // default
	}

	if cfg.Temperature < 0 || cfg.Temperature > 2 {
		v.addError("llm.azure.temperature", "must be between 0 and 2")
	}

	if cfg.TopP < 0 || cfg.TopP > 1 {
		v.addError("llm.azure.top_p", "must be between 0 and 1")
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30000000000 // 30 seconds default
	}

	if cfg.RetryAttempts < 0 {
		cfg.RetryAttempts = 3 // default
	}
}

// validateOllama validates Ollama configuration
func (v *Validator) validateOllama() {
	cfg := v.config.LLM.Ollama

	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:11434" // default
	} else if !v.isValidURL(cfg.BaseURL) {
		v.addError("llm.ollama.base_url", "must be a valid URL")
	}

	if cfg.Model == "" {
		v.addError("llm.ollama.model", "model name is required")
	}

	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 4096 // default
	}

	if cfg.Temperature < 0 || cfg.Temperature > 1 {
		v.addError("llm.ollama.temperature", "must be between 0 and 1")
	}

	if cfg.TopP < 0 || cfg.TopP > 1 {
		v.addError("llm.ollama.top_p", "must be between 0 and 1")
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 120000000000 // 120 seconds default (local models can be slower)
	}

	if cfg.RetryAttempts < 0 {
		cfg.RetryAttempts = 1 // default (less retries for local)
	}

	if cfg.KeepAlive == "" {
		cfg.KeepAlive = "5m" // default
	}
}

// validateMetaLlama validates Meta Llama configuration
func (v *Validator) validateMetaLlama() {
	cfg := v.config.LLM.MetaLlama

	if cfg.APIKey == "" {
		v.addError("llm.meta_llama.api_key", "Meta Llama API key is required")
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.llama.com/compat/v1" // default
	} else if !v.isValidURL(cfg.BaseURL) {
		v.addError("llm.meta_llama.base_url", "must be a valid URL")
	}

	if cfg.Model == "" {
		cfg.Model = "Llama-4-Maverick-17B-128E-Instruct-FP8" // default
	}

	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 4096 // default
	}

	if cfg.Temperature < 0 || cfg.Temperature > 2 {
		v.addError("llm.meta_llama.temperature", "must be between 0 and 2")
	}

	if cfg.TopP < 0 || cfg.TopP > 1 {
		v.addError("llm.meta_llama.top_p", "must be between 0 and 1")
	}

	if cfg.PresencePenalty < -2.0 || cfg.PresencePenalty > 2.0 {
		v.addError("llm.meta_llama.presence_penalty", "must be between -2 and 2")
	}

	if cfg.FrequencyPenalty < -2.0 || cfg.FrequencyPenalty > 2.0 {
		v.addError("llm.meta_llama.frequency_penalty", "must be between -2 and 2")
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 60000000000 // 60 seconds default
	}

	if cfg.RetryAttempts < 0 {
		cfg.RetryAttempts = 3 // default
	}
}

// validateFallback validates fallback configuration
func (v *Validator) validateFallback() {
	cfg := v.config.LLM.Fallback

	if len(cfg.Providers) == 0 {
		v.addError("llm.fallback.providers", "at least one fallback provider must be specified")
	}

	validProviders := []string{"anthropic", "openai", "google", "bedrock", "azure", "ollama"}
	for _, provider := range cfg.Providers {
		if !v.isValidEnum(provider, validProviders) {
			v.addError("llm.fallback.providers", fmt.Sprintf("invalid provider '%s', must be one of: %s", provider, strings.Join(validProviders, ", ")))
		}
	}

	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 2 // default
	}

	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = 1000000000 // 1 second default
	}
}

// validatePlatform validates platform configuration
func (v *Validator) validatePlatform() {
	v.validateAuthentication()
}

// validateAuthentication validates authentication configuration
func (v *Validator) validateAuthentication() {
	cfg := v.config.Platform.Authentication

	if cfg.Method == "" {
		cfg.Method = "api_key" // default
	}

	validMethods := []string{"jwt", "api_key", "oauth2"}
	if !v.isValidEnum(cfg.Method, validMethods) {
		v.addError("platform.authentication.method", fmt.Sprintf("must be one of: %s", strings.Join(validMethods, ", ")))
	}

	switch cfg.Method {
	case "api_key":
		if cfg.APIKey == "" {
			v.addError("platform.authentication.api_key", "API key is required when method is 'api_key'")
		}
	case "jwt":
		if cfg.Token == "" {
			v.addError("platform.authentication.token", "token is required when method is 'jwt'")
		}
	case "oauth2":
		if cfg.ClientID == "" {
			v.addError("platform.authentication.client_id", "client_id is required when method is 'oauth2'")
		}
		if cfg.ClientSecret == "" {
			v.addError("platform.authentication.client_secret", "client_secret is required when method is 'oauth2'")
		}
		if cfg.TokenURL == "" {
			v.addError("platform.authentication.token_url", "token_url is required when method is 'oauth2'")
		} else if !v.isValidURL(cfg.TokenURL) {
			v.addError("platform.authentication.token_url", "must be a valid URL")
		}
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 10000000000 // 10 seconds default
	}
}

// validateServices validates service configurations
func (v *Validator) validateServices() {
	if v.config.Services.ZeroDB != nil && v.config.Services.ZeroDB.Enabled {
		v.validateZeroDB()
	}

	if v.config.Services.Design != nil && v.config.Services.Design.Enabled {
		v.validateDesign()
	}

	if v.config.Services.Strapi != nil && v.config.Services.Strapi.Enabled {
		v.validateStrapi()
	}

	if v.config.Services.RLHF != nil && v.config.Services.RLHF.Enabled {
		v.validateRLHF()
	}
}

// validateZeroDB validates ZeroDB configuration
func (v *Validator) validateZeroDB() {
	cfg := v.config.Services.ZeroDB

	if cfg.Endpoint == "" {
		v.addError("services.zerodb.endpoint", "endpoint is required")
	}

	if cfg.Database == "" {
		v.addError("services.zerodb.database", "database name is required")
	}

	if cfg.MaxConnections <= 0 {
		cfg.MaxConnections = 10 // default
	}

	if cfg.IdleConnections < 0 {
		cfg.IdleConnections = 2 // default
	}

	if cfg.IdleConnections > cfg.MaxConnections {
		v.addError("services.zerodb.idle_connections", "cannot exceed max_connections")
	}

	if cfg.ConnMaxLifetime <= 0 {
		cfg.ConnMaxLifetime = 3600000000000 // 1 hour default
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 5000000000 // 5 seconds default
	}

	if cfg.RetryAttempts < 0 {
		cfg.RetryAttempts = 3 // default
	}

	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = 1000000000 // 1 second default
	}

	if cfg.SSLMode != "" {
		validSSLModes := []string{"disable", "require", "verify-ca", "verify-full"}
		if !v.isValidEnum(cfg.SSLMode, validSSLModes) {
			v.addError("services.zerodb.ssl_mode", fmt.Sprintf("must be one of: %s", strings.Join(validSSLModes, ", ")))
		}
	}
}

// validateDesign validates Design service configuration
func (v *Validator) validateDesign() {
	cfg := v.config.Services.Design

	if cfg.Endpoint == "" {
		v.addError("services.design.endpoint", "endpoint is required")
	} else if !v.isValidURL(cfg.Endpoint) {
		v.addError("services.design.endpoint", "must be a valid URL")
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30000000000 // 30 seconds default
	}

	if cfg.RetryAttempts < 0 {
		cfg.RetryAttempts = 3 // default
	}
}

// validateStrapi validates Strapi configuration
func (v *Validator) validateStrapi() {
	cfg := v.config.Services.Strapi

	if cfg.Endpoint == "" {
		v.addError("services.strapi.endpoint", "endpoint is required")
	} else if !v.isValidURL(cfg.Endpoint) {
		v.addError("services.strapi.endpoint", "must be a valid URL")
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30000000000 // 30 seconds default
	}

	if cfg.RetryAttempts < 0 {
		cfg.RetryAttempts = 3 // default
	}
}

// validateRLHF validates RLHF service configuration
func (v *Validator) validateRLHF() {
	cfg := v.config.Services.RLHF

	if cfg.Endpoint == "" {
		v.addError("services.rlhf.endpoint", "endpoint is required")
	} else if !v.isValidURL(cfg.Endpoint) {
		v.addError("services.rlhf.endpoint", "must be a valid URL")
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 60000000000 // 60 seconds default
	}

	if cfg.RetryAttempts < 0 {
		cfg.RetryAttempts = 3 // default
	}
}

// validateTools validates tool configurations
func (v *Validator) validateTools() {
	if v.config.Tools.FileSystem != nil && v.config.Tools.FileSystem.Enabled {
		v.validateFileSystemTool()
	}

	if v.config.Tools.Terminal != nil && v.config.Tools.Terminal.Enabled {
		v.validateTerminalTool()
	}

	if v.config.Tools.Browser != nil && v.config.Tools.Browser.Enabled {
		v.validateBrowserTool()
	}

	if v.config.Tools.CodeAnalysis != nil && v.config.Tools.CodeAnalysis.Enabled {
		v.validateCodeAnalysisTool()
	}
}

// validateFileSystemTool validates filesystem tool configuration
func (v *Validator) validateFileSystemTool() {
	cfg := v.config.Tools.FileSystem

	if len(cfg.AllowedPaths) == 0 {
		v.addError("tools.filesystem.allowed_paths", "at least one allowed path must be specified when filesystem tool is enabled")
		return
	}

	// Validate paths are absolute and exist (warning only for non-existent paths)
	for _, path := range cfg.AllowedPaths {
		if !filepath.IsAbs(path) {
			v.addError("tools.filesystem.allowed_paths", fmt.Sprintf("path '%s' must be absolute", path))
		}
		// Only check existence for paths that are absolute
		if filepath.IsAbs(path) {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				// This is a warning, not an error - path might be created later
				v.addError("tools.filesystem.allowed_paths", fmt.Sprintf("path '%s' does not exist", path))
			}
		}
	}

	if cfg.MaxFileSize <= 0 {
		cfg.MaxFileSize = 104857600 // 100MB default
	}
}

// validateTerminalTool validates terminal tool configuration
func (v *Validator) validateTerminalTool() {
	cfg := v.config.Tools.Terminal

	if cfg.Timeout <= 0 {
		cfg.Timeout = 300000000000 // 5 minutes default
	}

	if cfg.WorkingDir != "" {
		if !filepath.IsAbs(cfg.WorkingDir) {
			v.addError("tools.terminal.working_dir", "must be an absolute path")
		}
		if _, err := os.Stat(cfg.WorkingDir); os.IsNotExist(err) {
			v.addError("tools.terminal.working_dir", "directory does not exist")
		}
	}
}

// validateBrowserTool validates browser tool configuration
func (v *Validator) validateBrowserTool() {
	cfg := v.config.Tools.Browser

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30000000000 // 30 seconds default
	}
}

// validateCodeAnalysisTool validates code analysis tool configuration
func (v *Validator) validateCodeAnalysisTool() {
	cfg := v.config.Tools.CodeAnalysis

	if len(cfg.Languages) == 0 {
		// Default to common languages
		cfg.Languages = []string{"go", "python", "javascript", "typescript", "java"}
	}

	if cfg.MaxFileSize <= 0 {
		cfg.MaxFileSize = 10485760 // 10MB default
	}
}

// validatePerformance validates performance configurations
func (v *Validator) validatePerformance() {
	v.validateCache()
	v.validateRateLimit()
	v.validateConcurrency()
	v.validateCircuitBreaker()
}

// validateCache validates cache configuration
func (v *Validator) validateCache() {
	cfg := v.config.Performance.Cache

	if !cfg.Enabled {
		return
	}

	if cfg.Type == "" {
		cfg.Type = "memory" // default
	}

	validTypes := []string{"memory", "redis", "memcached"}
	if !v.isValidEnum(cfg.Type, validTypes) {
		v.addError("performance.cache.type", fmt.Sprintf("must be one of: %s", strings.Join(validTypes, ", ")))
	}

	if cfg.Type == "redis" && cfg.RedisURL == "" {
		v.addError("performance.cache.redis_url", "redis_url is required when type is 'redis'")
	}

	if cfg.Type == "memcached" && cfg.MemcachedURL == "" {
		v.addError("performance.cache.memcached_url", "memcached_url is required when type is 'memcached'")
	}

	if cfg.TTL <= 0 {
		cfg.TTL = 3600000000000 // 1 hour default
	}

	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 100 // 100MB default
	}
}

// validateRateLimit validates rate limit configuration
func (v *Validator) validateRateLimit() {
	cfg := v.config.Performance.RateLimit

	if !cfg.Enabled {
		return
	}

	if cfg.RequestsPerMinute <= 0 {
		cfg.RequestsPerMinute = 60 // default
	}

	if cfg.BurstSize <= 0 {
		cfg.BurstSize = 10 // default
	}

	if cfg.TimeWindow <= 0 {
		cfg.TimeWindow = 60000000000 // 1 minute default
	}
}

// validateConcurrency validates concurrency configuration
func (v *Validator) validateConcurrency() {
	cfg := v.config.Performance.Concurrency

	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = 10 // default
	}

	if cfg.MaxQueueSize <= 0 {
		cfg.MaxQueueSize = 100 // default
	}

	if cfg.WorkerTimeout <= 0 {
		cfg.WorkerTimeout = 300000000000 // 5 minutes default
	}
}

// validateCircuitBreaker validates circuit breaker configuration
func (v *Validator) validateCircuitBreaker() {
	cfg := v.config.Performance.CircuitBreaker

	if !cfg.Enabled {
		return
	}

	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5 // default
	}

	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = 2 // default
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 60000000000 // 60 seconds default
	}

	if cfg.ResetTimeout <= 0 {
		cfg.ResetTimeout = 30000000000 // 30 seconds default
	}
}

// validateLogging validates logging configuration
func (v *Validator) validateLogging() {
	cfg := v.config.Logging

	if cfg.Level == "" {
		cfg.Level = "info" // default
	}

	validLevels := []string{"debug", "info", "warn", "error"}
	if !v.isValidEnum(cfg.Level, validLevels) {
		v.addError("logging.level", fmt.Sprintf("must be one of: %s", strings.Join(validLevels, ", ")))
	}

	if cfg.Format == "" {
		cfg.Format = "json" // default
	}

	validFormats := []string{"json", "console"}
	if !v.isValidEnum(cfg.Format, validFormats) {
		v.addError("logging.format", fmt.Sprintf("must be one of: %s", strings.Join(validFormats, ", ")))
	}

	if cfg.Output == "" {
		cfg.Output = "stdout" // default
	}

	validOutputs := []string{"stdout", "file"}
	if !v.isValidEnum(cfg.Output, validOutputs) {
		v.addError("logging.output", fmt.Sprintf("must be one of: %s", strings.Join(validOutputs, ", ")))
	}

	if cfg.Output == "file" && cfg.FilePath == "" {
		v.addError("logging.file_path", "file_path is required when output is 'file'")
	}

	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 100 // 100MB default
	}

	if cfg.MaxBackups < 0 {
		cfg.MaxBackups = 3 // default
	}

	if cfg.MaxAge <= 0 {
		cfg.MaxAge = 7 // 7 days default
	}
}

// validateSecurity validates security configuration
func (v *Validator) validateSecurity() {
	cfg := v.config.Security

	if cfg.EncryptConfig && cfg.EncryptionKey == "" {
		v.addError("security.encryption_key", "encryption_key is required when encrypt_config is true")
	}

	// Only validate key length if encryption is enabled
	if cfg.EncryptConfig && cfg.EncryptionKey != "" && len(cfg.EncryptionKey) < 32 {
		v.addError("security.encryption_key", "must be at least 32 characters for AES-256")
	}

	if cfg.TLSEnabled {
		if cfg.TLSCertPath == "" {
			v.addError("security.tls_cert_path", "tls_cert_path is required when tls_enabled is true")
		}
		if cfg.TLSKeyPath == "" {
			v.addError("security.tls_key_path", "tls_key_path is required when tls_enabled is true")
		}

		// Validate cert and key files exist
		if cfg.TLSCertPath != "" {
			if _, err := os.Stat(cfg.TLSCertPath); os.IsNotExist(err) {
				v.addError("security.tls_cert_path", "certificate file does not exist")
			}
		}
		if cfg.TLSKeyPath != "" {
			if _, err := os.Stat(cfg.TLSKeyPath); os.IsNotExist(err) {
				v.addError("security.tls_key_path", "key file does not exist")
			}
		}
	}
}

// Helper methods

func (v *Validator) addError(key, message string) {
	v.errs = append(v.errs, errors.NewConfigValidationError(key, message))
}

func (v *Validator) isValidEnum(value string, validValues []string) bool {
	for _, valid := range validValues {
		if value == valid {
			return true
		}
	}
	return false
}

func (v *Validator) isValidURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

func (v *Validator) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func (v *Validator) createValidationError() error {
	var messages []string
	for _, err := range v.errs {
		messages = append(messages, err.Error())
	}
	return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(messages, "\n  - "))
}
