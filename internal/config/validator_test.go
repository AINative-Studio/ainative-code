package config

import (
	"testing"
	"time"
)

func TestValidateApp(t *testing.T) {
	tests := []struct {
		name    string
		config  AppConfig
		wantErr bool
	}{
		{
			name: "valid app config",
			config: AppConfig{
				Name:        "test-app",
				Version:     "1.0.0",
				Environment: "development",
				Debug:       true,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: AppConfig{
				Version:     "1.0.0",
				Environment: "development",
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			config: AppConfig{
				Name:        "test-app",
				Environment: "invalid",
			},
			wantErr: true,
		},
		{
			name: "empty environment defaults to development",
			config: AppConfig{
				Name: "test-app",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{App: tt.config}
			validator := NewValidator(cfg)
			validator.validateApp()

			hasErrors := len(validator.errs) > 0
			if hasErrors != tt.wantErr {
				t.Errorf("validateApp() error = %v, wantErr %v", hasErrors, tt.wantErr)
			}
		})
	}
}

func TestValidateAnthropic(t *testing.T) {
	tests := []struct {
		name    string
		config  *AnthropicConfig
		wantErr bool
	}{
		{
			name: "valid anthropic config",
			config: &AnthropicConfig{
				APIKey:      "sk-ant-test",
				Model:       "claude-3-5-sonnet-20241022",
				MaxTokens:   4096,
				Temperature: 0.7,
				TopP:        1.0,
				Timeout:     30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: &AnthropicConfig{
				Model: "claude-3-5-sonnet-20241022",
			},
			wantErr: true,
		},
		{
			name: "invalid temperature",
			config: &AnthropicConfig{
				APIKey:      "sk-ant-test",
				Temperature: 1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid top_p",
			config: &AnthropicConfig{
				APIKey: "sk-ant-test",
				TopP:   1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid base URL",
			config: &AnthropicConfig{
				APIKey:  "sk-ant-test",
				BaseURL: "not-a-url",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				LLM: LLMConfig{
					DefaultProvider: "anthropic",
					Anthropic:       tt.config,
				},
			}
			validator := NewValidator(cfg)
			validator.validateAnthropic()

			hasErrors := len(validator.errs) > 0
			if hasErrors != tt.wantErr {
				t.Errorf("validateAnthropic() error = %v, wantErr %v, errors: %v", hasErrors, tt.wantErr, validator.errs)
			}
		})
	}
}

func TestValidateOpenAI(t *testing.T) {
	tests := []struct {
		name    string
		config  *OpenAIConfig
		wantErr bool
	}{
		{
			name: "valid openai config",
			config: &OpenAIConfig{
				APIKey:      "sk-test",
				Model:       "gpt-4",
				MaxTokens:   4096,
				Temperature: 0.7,
				TopP:        1.0,
				Timeout:     30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: &OpenAIConfig{
				Model: "gpt-4",
			},
			wantErr: true,
		},
		{
			name: "invalid temperature",
			config: &OpenAIConfig{
				APIKey:      "sk-test",
				Temperature: 2.5,
			},
			wantErr: true,
		},
		{
			name: "invalid frequency penalty",
			config: &OpenAIConfig{
				APIKey:           "sk-test",
				FrequencyPenalty: -3.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				LLM: LLMConfig{
					DefaultProvider: "openai",
					OpenAI:          tt.config,
				},
			}
			validator := NewValidator(cfg)
			validator.validateOpenAI()

			hasErrors := len(validator.errs) > 0
			if hasErrors != tt.wantErr {
				t.Errorf("validateOpenAI() error = %v, wantErr %v", hasErrors, tt.wantErr)
			}
		})
	}
}

func TestValidateZeroDB(t *testing.T) {
	tests := []struct {
		name    string
		config  *ZeroDBConfig
		wantErr bool
	}{
		{
			name: "valid zerodb config",
			config: &ZeroDBConfig{
				Enabled:         true,
				Endpoint:        "postgresql://localhost:5432",
				Database:        "testdb",
				MaxConnections:  10,
				IdleConnections: 2,
				Timeout:         5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing endpoint",
			config: &ZeroDBConfig{
				Enabled:  true,
				Database: "testdb",
			},
			wantErr: true,
		},
		{
			name: "missing database",
			config: &ZeroDBConfig{
				Enabled:  true,
				Endpoint: "postgresql://localhost:5432",
			},
			wantErr: true,
		},
		{
			name: "idle connections exceed max",
			config: &ZeroDBConfig{
				Enabled:         true,
				Endpoint:        "postgresql://localhost:5432",
				Database:        "testdb",
				MaxConnections:  5,
				IdleConnections: 10,
			},
			wantErr: true,
		},
		{
			name: "invalid ssl mode",
			config: &ZeroDBConfig{
				Enabled:  true,
				Endpoint: "postgresql://localhost:5432",
				Database: "testdb",
				SSLMode:  "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Services: ServicesConfig{
					ZeroDB: tt.config,
				},
			}
			validator := NewValidator(cfg)
			validator.validateZeroDB()

			hasErrors := len(validator.errs) > 0
			if hasErrors != tt.wantErr {
				t.Errorf("validateZeroDB() error = %v, wantErr %v, errors: %v", hasErrors, tt.wantErr, validator.errs)
			}
		})
	}
}

func TestValidateAuthentication(t *testing.T) {
	tests := []struct {
		name    string
		config  AuthConfig
		wantErr bool
	}{
		{
			name: "valid API key auth",
			config: AuthConfig{
				Method: "api_key",
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name: "valid JWT auth",
			config: AuthConfig{
				Method: "jwt",
				Token:  "test-token",
			},
			wantErr: false,
		},
		{
			name: "valid OAuth2 auth",
			config: AuthConfig{
				Method:       "oauth2",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				TokenURL:     "https://auth.example.com/token",
			},
			wantErr: false,
		},
		{
			name: "invalid auth method",
			config: AuthConfig{
				Method: "invalid",
			},
			wantErr: true,
		},
		{
			name: "API key auth missing key",
			config: AuthConfig{
				Method: "api_key",
			},
			wantErr: true,
		},
		{
			name: "JWT auth missing token",
			config: AuthConfig{
				Method: "jwt",
			},
			wantErr: true,
		},
		{
			name: "OAuth2 missing client ID",
			config: AuthConfig{
				Method:       "oauth2",
				ClientSecret: "secret",
				TokenURL:     "https://auth.example.com/token",
			},
			wantErr: true,
		},
		{
			name: "OAuth2 invalid token URL",
			config: AuthConfig{
				Method:       "oauth2",
				ClientID:     "client-id",
				ClientSecret: "secret",
				TokenURL:     "not-a-url",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Platform: PlatformConfig{
					Authentication: tt.config,
				},
			}
			validator := NewValidator(cfg)
			validator.validateAuthentication()

			hasErrors := len(validator.errs) > 0
			if hasErrors != tt.wantErr {
				t.Errorf("validateAuthentication() error = %v, wantErr %v, errors: %v", hasErrors, tt.wantErr, validator.errs)
			}
		})
	}
}

func TestValidateFallback(t *testing.T) {
	tests := []struct {
		name    string
		config  *FallbackConfig
		wantErr bool
	}{
		{
			name: "valid fallback config",
			config: &FallbackConfig{
				Enabled:    true,
				Providers:  []string{"anthropic", "openai"},
				MaxRetries: 2,
				RetryDelay: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty providers list",
			config: &FallbackConfig{
				Enabled:   true,
				Providers: []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid provider",
			config: &FallbackConfig{
				Enabled:   true,
				Providers: []string{"invalid-provider"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				LLM: LLMConfig{
					Fallback: tt.config,
				},
			}
			validator := NewValidator(cfg)
			validator.validateFallback()

			hasErrors := len(validator.errs) > 0
			if hasErrors != tt.wantErr {
				t.Errorf("validateFallback() error = %v, wantErr %v", hasErrors, tt.wantErr)
			}
		})
	}
}

func TestValidateCache(t *testing.T) {
	tests := []struct {
		name    string
		config  CacheConfig
		wantErr bool
	}{
		{
			name: "valid memory cache",
			config: CacheConfig{
				Enabled: true,
				Type:    "memory",
				TTL:     1 * time.Hour,
				MaxSize: 100,
			},
			wantErr: false,
		},
		{
			name: "valid redis cache",
			config: CacheConfig{
				Enabled:  true,
				Type:     "redis",
				RedisURL: "redis://localhost:6379",
			},
			wantErr: false,
		},
		{
			name: "redis cache missing URL",
			config: CacheConfig{
				Enabled: true,
				Type:    "redis",
			},
			wantErr: true,
		},
		{
			name: "invalid cache type",
			config: CacheConfig{
				Enabled: true,
				Type:    "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Performance: PerformanceConfig{
					Cache: tt.config,
				},
			}
			validator := NewValidator(cfg)
			validator.validateCache()

			hasErrors := len(validator.errs) > 0
			if hasErrors != tt.wantErr {
				t.Errorf("validateCache() error = %v, wantErr %v", hasErrors, tt.wantErr)
			}
		})
	}
}

func TestValidateLogging(t *testing.T) {
	tests := []struct {
		name    string
		config  LoggingConfig
		wantErr bool
	}{
		{
			name: "valid console logging",
			config: LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
			wantErr: false,
		},
		{
			name: "valid file logging",
			config: LoggingConfig{
				Level:    "debug",
				Format:   "console",
				Output:   "file",
				FilePath: "/var/log/app.log",
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: LoggingConfig{
				Level:  "invalid",
				Format: "json",
				Output: "stdout",
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			config: LoggingConfig{
				Level:  "info",
				Format: "invalid",
				Output: "stdout",
			},
			wantErr: true,
		},
		{
			name: "file output missing path",
			config: LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "file",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Logging: tt.config,
			}
			validator := NewValidator(cfg)
			validator.validateLogging()

			hasErrors := len(validator.errs) > 0
			if hasErrors != tt.wantErr {
				t.Errorf("validateLogging() error = %v, wantErr %v", hasErrors, tt.wantErr)
			}
		})
	}
}

func TestValidateSecurity(t *testing.T) {
	tests := []struct {
		name    string
		config  SecurityConfig
		wantErr bool
	}{
		{
			name: "valid security config",
			config: SecurityConfig{
				EncryptConfig: false,
				TLSEnabled:    false,
			},
			wantErr: false,
		},
		{
			name: "encryption enabled with key",
			config: SecurityConfig{
				EncryptConfig: true,
				EncryptionKey: "this-is-a-valid-32-character-key-here!!!",
			},
			wantErr: false,
		},
		{
			name: "encryption enabled without key",
			config: SecurityConfig{
				EncryptConfig: true,
			},
			wantErr: true,
		},
		{
			name: "encryption key too short",
			config: SecurityConfig{
				EncryptConfig: true,
				EncryptionKey: "short-key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Security: tt.config,
			}
			validator := NewValidator(cfg)
			validator.validateSecurity()

			hasErrors := len(validator.errs) > 0
			if hasErrors != tt.wantErr {
				t.Errorf("validateSecurity() error = %v, wantErr %v", hasErrors, tt.wantErr)
			}
		})
	}
}

func TestValidate_Complete(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "minimal valid config",
			config: &Config{
				App: AppConfig{
					Name:        "test-app",
					Environment: "development",
				},
				LLM: LLMConfig{
					DefaultProvider: "anthropic",
					Anthropic: &AnthropicConfig{
						APIKey: "sk-ant-test",
					},
				},
				Platform: PlatformConfig{
					Authentication: AuthConfig{
						Method: "api_key",
						APIKey: "test-key",
					},
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
					Output: "stdout",
				},
			},
			wantErr: false,
		},
		{
			name: "config with multiple validation errors",
			config: &Config{
				App: AppConfig{
					Environment: "invalid",
				},
				LLM: LLMConfig{
					DefaultProvider: "invalid-provider",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(tt.config)
			err := validator.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	validator := &Validator{}

	tests := []struct {
		url     string
		valid   bool
	}{
		{"https://api.anthropic.com", true},
		{"http://localhost:8080", true},
		{"postgresql://localhost:5432", true},
		{"redis://localhost:6379", true},
		{"not-a-url", false},
		{"", false},
		{"http://", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := validator.isValidURL(tt.url)
			if result != tt.valid {
				t.Errorf("isValidURL(%s) = %v, want %v", tt.url, result, tt.valid)
			}
		})
	}
}

func TestIsValidEnum(t *testing.T) {
	validator := &Validator{}

	validValues := []string{"development", "staging", "production"}

	tests := []struct {
		value string
		valid bool
	}{
		{"development", true},
		{"staging", true},
		{"production", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := validator.isValidEnum(tt.value, validValues)
			if result != tt.valid {
				t.Errorf("isValidEnum(%s) = %v, want %v", tt.value, result, tt.valid)
			}
		})
	}
}
