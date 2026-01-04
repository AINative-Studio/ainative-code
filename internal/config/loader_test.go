package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLoader(t *testing.T) {
	tests := []struct {
		name string
		opts []LoaderOption
		want *Loader
	}{
		{
			name: "default loader",
			opts: nil,
			want: &Loader{
				configName: "config",
				configType: "yaml",
				envPrefix:  "AINATIVE",
			},
		},
		{
			name: "custom config name",
			opts: []LoaderOption{WithConfigName("custom")},
			want: &Loader{
				configName: "custom",
				configType: "yaml",
				envPrefix:  "AINATIVE",
			},
		},
		{
			name: "custom config type",
			opts: []LoaderOption{WithConfigType("json")},
			want: &Loader{
				configName: "config",
				configType: "json",
				envPrefix:  "AINATIVE",
			},
		},
		{
			name: "custom env prefix",
			opts: []LoaderOption{WithEnvPrefix("MYAPP")},
			want: &Loader{
				configName: "config",
				configType: "yaml",
				envPrefix:  "MYAPP",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLoader(tt.opts...)

			if got.configName != tt.want.configName {
				t.Errorf("configName = %v, want %v", got.configName, tt.want.configName)
			}
			if got.configType != tt.want.configType {
				t.Errorf("configType = %v, want %v", got.configType, tt.want.configType)
			}
			if got.envPrefix != tt.want.envPrefix {
				t.Errorf("envPrefix = %v, want %v", got.envPrefix, tt.want.envPrefix)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
app:
  name: test-app
  environment: development

llm:
  default_provider: anthropic
  anthropic:
    api_key: sk-ant-test
    model: claude-3-5-sonnet-20241022

platform:
  authentication:
    method: api_key
    api_key: test-key

logging:
  level: info
  format: json
  output: stdout
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if cfg.App.Name != "test-app" {
		t.Errorf("App.Name = %v, want %v", cfg.App.Name, "test-app")
	}

	if cfg.LLM.DefaultProvider != "anthropic" {
		t.Errorf("LLM.DefaultProvider = %v, want %v", cfg.LLM.DefaultProvider, "anthropic")
	}

	if cfg.LLM.Anthropic == nil {
		t.Error("LLM.Anthropic is nil")
	} else {
		if cfg.LLM.Anthropic.APIKey != "sk-ant-test" {
			t.Errorf("LLM.Anthropic.APIKey = %v, want %v", cfg.LLM.Anthropic.APIKey, "sk-ant-test")
		}
	}
}

func TestLoadFromFile_NonExistent(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadFromFile("/nonexistent/path/config.yaml")

	if err == nil {
		t.Error("LoadFromFile() expected error for non-existent file, got nil")
	}
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidContent := `
app:
  name: test
  invalid yaml structure
    broken: indentation
`

	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	loader := NewLoader()
	_, err := loader.LoadFromFile(configPath)

	if err == nil {
		t.Error("LoadFromFile() expected error for invalid YAML, got nil")
	}
}

func TestLoadFromFile_ValidationErrors(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-config.yaml")

	// Config with validation errors
	configContent := `
app:
  environment: invalid-env

llm:
  default_provider: invalid-provider
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	loader := NewLoader()
	_, err := loader.LoadFromFile(configPath)

	if err == nil {
		t.Error("LoadFromFile() expected validation error, got nil")
	}
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"AINATIVE_APP_NAME":                    "env-app",
		"AINATIVE_APP_ENVIRONMENT":             "production",
		"AINATIVE_LLM_DEFAULT_PROVIDER":        "openai",
		"AINATIVE_LLM_OPENAI_API_KEY":          "sk-test-env",
		"AINATIVE_PLATFORM_AUTHENTICATION_METHOD": "api_key",
		"AINATIVE_PLATFORM_AUTHENTICATION_API_KEY": "env-key",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	loader := NewLoader()
	cfg, err := loader.Load()

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Name != "env-app" {
		t.Errorf("App.Name = %v, want %v", cfg.App.Name, "env-app")
	}

	if cfg.App.Environment != "production" {
		t.Errorf("App.Environment = %v, want %v", cfg.App.Environment, "production")
	}

	if cfg.LLM.DefaultProvider != "openai" {
		t.Errorf("LLM.DefaultProvider = %v, want %v", cfg.LLM.DefaultProvider, "openai")
	}
}

func TestLoad_Defaults(t *testing.T) {
	// Clear any environment variables
	os.Clearenv()

	// Set only required fields via env
	os.Setenv("AINATIVE_APP_NAME", "test-app")
	os.Setenv("AINATIVE_LLM_ANTHROPIC_API_KEY", "sk-ant-test")
	os.Setenv("AINATIVE_PLATFORM_AUTHENTICATION_API_KEY", "test-key")
	defer func() {
		os.Unsetenv("AINATIVE_APP_NAME")
		os.Unsetenv("AINATIVE_LLM_ANTHROPIC_API_KEY")
		os.Unsetenv("AINATIVE_PLATFORM_AUTHENTICATION_API_KEY")
	}()

	loader := NewLoader()
	cfg, err := loader.Load()

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check defaults are applied
	if cfg.App.Environment != "development" {
		t.Errorf("App.Environment = %v, want %v (default)", cfg.App.Environment, "development")
	}

	if cfg.LLM.DefaultProvider != "anthropic" {
		t.Errorf("LLM.DefaultProvider = %v, want %v (default)", cfg.LLM.DefaultProvider, "anthropic")
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %v, want %v (default)", cfg.Logging.Level, "info")
	}

	if cfg.Logging.Format != "json" {
		t.Errorf("Logging.Format = %v, want %v (default)", cfg.Logging.Format, "json")
	}
}

func TestLoad_FileAndEnvPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "precedence.yaml")

	// File config
	configContent := `
app:
  name: file-app
  environment: development

llm:
  default_provider: anthropic
  anthropic:
    api_key: sk-ant-file

platform:
  authentication:
    method: api_key
    api_key: file-key

logging:
  level: debug
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Environment variable should override file
	os.Setenv("AINATIVE_APP_NAME", "env-app")
	os.Setenv("AINATIVE_LOGGING_LEVEL", "error")
	defer func() {
		os.Unsetenv("AINATIVE_APP_NAME")
		os.Unsetenv("AINATIVE_LOGGING_LEVEL")
	}()

	loader := NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	// Environment variable should win
	if cfg.App.Name != "env-app" {
		t.Errorf("App.Name = %v, want %v (from env)", cfg.App.Name, "env-app")
	}

	if cfg.Logging.Level != "error" {
		t.Errorf("Logging.Level = %v, want %v (from env)", cfg.Logging.Level, "error")
	}

	// File values should be used where env not set
	if cfg.LLM.Anthropic.APIKey != "sk-ant-file" {
		t.Errorf("LLM.Anthropic.APIKey = %v, want %v (from file)", cfg.LLM.Anthropic.APIKey, "sk-ant-file")
	}
}

func TestWithConfigPaths(t *testing.T) {
	customPaths := []string{"/custom/path1", "/custom/path2"}
	loader := NewLoader(WithConfigPaths(customPaths...))

	if len(loader.configPaths) != len(customPaths) {
		t.Errorf("len(configPaths) = %v, want %v", len(loader.configPaths), len(customPaths))
	}

	for i, path := range loader.configPaths {
		if path != customPaths[i] {
			t.Errorf("configPaths[%d] = %v, want %v", i, path, customPaths[i])
		}
	}
}

func TestGetConfigFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")

	configContent := `
app:
  name: test-app

llm:
  default_provider: anthropic
  anthropic:
    api_key: sk-ant-test

platform:
  authentication:
    method: api_key
    api_key: test-key
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	loader := NewLoader()
	_, err := loader.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	usedPath := loader.GetConfigFilePath()
	if usedPath != configPath {
		t.Errorf("GetConfigFilePath() = %v, want %v", usedPath, configPath)
	}
}

func TestLoad_CompleteConfiguration(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "complete.yaml")

	// Complete configuration
	configContent := `
app:
  name: complete-test
  version: 1.0.0
  environment: production
  debug: false

llm:
  default_provider: anthropic
  anthropic:
    api_key: sk-ant-complete
    model: claude-3-5-sonnet-20241022
    max_tokens: 8192
    temperature: 0.8
    top_p: 0.95
    timeout: 45s
    retry_attempts: 5

  fallback:
    enabled: true
    providers:
      - anthropic
      - openai
    max_retries: 3
    retry_delay: 2s

platform:
  authentication:
    method: jwt
    token: test-jwt-token
    timeout: 15s

  organization:
    id: org-123
    name: Test Org
    workspace: production

services:
  zerodb:
    enabled: true
    endpoint: postgresql://localhost:5432
    database: testdb
    username: testuser
    password: testpass
    ssl: true
    ssl_mode: require
    max_connections: 20
    idle_connections: 5
    timeout: 10s

tools:
  filesystem:
    enabled: false  # Disabled to avoid path validation

  terminal:
    enabled: true
    timeout: 10m

performance:
  cache:
    enabled: true
    type: memory
    ttl: 2h
    max_size: 200

  rate_limit:
    enabled: true
    requests_per_minute: 120
    burst_size: 20

logging:
  level: warn
  format: json
  output: stdout

security:
  encrypt_config: false
  tls_enabled: false
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.LoadFromFile(configPath)

	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	// Verify all sections loaded correctly
	if cfg.App.Name != "complete-test" {
		t.Errorf("App.Name = %v, want complete-test", cfg.App.Name)
	}

	if cfg.LLM.Anthropic.MaxTokens != 8192 {
		t.Errorf("LLM.Anthropic.MaxTokens = %v, want 8192", cfg.LLM.Anthropic.MaxTokens)
	}

	if !cfg.LLM.Fallback.Enabled {
		t.Error("LLM.Fallback.Enabled should be true")
	}

	if cfg.Platform.Authentication.Method != "jwt" {
		t.Errorf("Platform.Authentication.Method = %v, want jwt", cfg.Platform.Authentication.Method)
	}

	if !cfg.Services.ZeroDB.Enabled {
		t.Error("Services.ZeroDB.Enabled should be true")
	}

	if cfg.Services.ZeroDB.MaxConnections != 20 {
		t.Errorf("Services.ZeroDB.MaxConnections = %v, want 20", cfg.Services.ZeroDB.MaxConnections)
	}

	if !cfg.Performance.Cache.Enabled {
		t.Error("Performance.Cache.Enabled should be true")
	}

	if cfg.Logging.Level != "warn" {
		t.Errorf("Logging.Level = %v, want warn", cfg.Logging.Level)
	}
}

func TestWithResolver(t *testing.T) {
	resolver := NewResolver()
	loader := NewLoader(WithResolver(resolver))

	if loader.resolver == nil {
		t.Error("WithResolver() should set resolver")
	}
}

func TestGetViper(t *testing.T) {
	loader := NewLoader()

	// Load a basic config first
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")
	configContent := `
app:
  name: test-app

llm:
  default_provider: anthropic
  anthropic:
    api_key: sk-ant-test

platform:
  authentication:
    method: api_key
    api_key: test-key
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	if _, err := loader.LoadFromFile(configPath); err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	viper := loader.GetViper()
	if viper == nil {
		t.Error("GetViper() should return viper instance")
	}

	// Verify we can access config through viper
	if viper.GetString("app.name") != "test-app" {
		t.Errorf("viper.GetString(\"app.name\") = %v, want test-app", viper.GetString("app.name"))
	}
}

func TestWriteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.yaml")
	outputPath := filepath.Join(tmpDir, "output.yaml")

	inputContent := `
app:
  name: test-app
  environment: development

llm:
  default_provider: anthropic
  anthropic:
    api_key: sk-ant-test

platform:
  authentication:
    method: api_key
    api_key: test-key
`

	if err := os.WriteFile(inputPath, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.LoadFromFile(inputPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if err := WriteConfig(cfg, outputPath); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("WriteConfig() should create output file")
	}

	// Verify we can load the written config
	loader2 := NewLoader()
	cfg2, err := loader2.LoadFromFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load written config: %v", err)
	}

	if cfg2.App.Name != "test-app" {
		t.Errorf("Written config App.Name = %v, want test-app", cfg2.App.Name)
	}
}
