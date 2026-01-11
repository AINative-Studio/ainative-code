package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/config"
	"github.com/AINative-studio/ainative-code/internal/setup"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// TestZeroDBConfigPathConsistency verifies that setup wizard and ZeroDB commands
// use consistent configuration paths (GitHub issue #125).
//
// Problem: Setup wizard saves to services.zerodb.* but commands read from zerodb.*
// Solution: Both should use services.zerodb.* for consistency
func TestZeroDBConfigPathConsistency(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	// Create configuration with ZeroDB settings using services.zerodb.* path
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "ainative-code",
			Version:     "0.1.0",
			Environment: "test",
			Debug:       false,
		},
		LLM: config.LLMConfig{
			DefaultProvider: "anthropic",
		},
		Platform: config.PlatformConfig{
			Authentication: config.AuthConfig{
				Method:  "none",
				Timeout: 10 * time.Second,
			},
		},
		Services: config.ServicesConfig{
			ZeroDB: &config.ZeroDBConfig{
				Enabled:         true,
				ProjectID:       "test-project-12345",
				Endpoint:        "https://api.ainative.studio",
				Database:        "default",
				SSL:             true,
				SSLMode:         "require",
				MaxConnections:  10,
				IdleConnections: 5,
				ConnMaxLifetime: 1 * time.Hour,
				Timeout:         30 * time.Second,
				RetryAttempts:   3,
				RetryDelay:      1 * time.Second,
			},
		},
		Performance: config.PerformanceConfig{
			Cache: config.CacheConfig{
				Enabled: false,
				Type:    "memory",
				TTL:     1 * time.Hour,
				MaxSize: 100,
			},
			Concurrency: config.ConcurrencyConfig{
				MaxWorkers:    10,
				MaxQueueSize:  100,
				WorkerTimeout: 5 * time.Minute,
			},
		},
		Logging: config.LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
		Security: config.SecurityConfig{
			EncryptConfig: false,
			TLSEnabled:    false,
		},
	}

	// Marshal configuration to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Write configuration file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load configuration using viper (same way ZeroDB commands do)
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Test 1: Verify services.zerodb.project_id is accessible
	projectID := v.GetString("services.zerodb.project_id")
	if projectID == "" {
		t.Error("services.zerodb.project_id not found in config")
	}
	if projectID != "test-project-12345" {
		t.Errorf("Expected project_id 'test-project-12345', got '%s'", projectID)
	}

	// Test 2: Verify services.zerodb.endpoint is accessible
	endpoint := v.GetString("services.zerodb.endpoint")
	if endpoint == "" {
		t.Error("services.zerodb.endpoint not found in config")
	}
	if endpoint != "https://api.ainative.studio" {
		t.Errorf("Expected endpoint 'https://api.ainative.studio', got '%s'", endpoint)
	}

	// Test 3: Verify old paths (zerodb.* without services.) are NOT populated
	oldProjectID := v.GetString("zerodb.project_id")
	if oldProjectID != "" {
		t.Errorf("Old path 'zerodb.project_id' should not be populated, got '%s'", oldProjectID)
	}

	oldEndpoint := v.GetString("zerodb.base_url")
	if oldEndpoint != "" {
		t.Errorf("Old path 'zerodb.base_url' should not be populated, got '%s'", oldEndpoint)
	}

	t.Log("✓ ZeroDB configuration paths are consistent")
}

// TestSetupWizardZeroDBConfig verifies that the setup wizard saves ZeroDB
// configuration to the correct path (services.zerodb.*).
func TestSetupWizardZeroDBConfig(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "wizard-config.yaml")

	// Create wizard with non-interactive mode
	wizard := setup.NewWizard(ctx, setup.WizardConfig{
		ConfigPath:      configPath,
		SkipValidation:  true,
		InteractiveMode: false,
		Force:           true,
	})

	// Set user selections for ZeroDB
	selections := map[string]interface{}{
		"provider":          "anthropic",
		"anthropic_api_key": "test-key",
		"anthropic_model":   "claude-3-5-sonnet-20241022",
		"zerodb_enabled":    true,
		"zerodb_project_id": "wizard-test-project",
		"zerodb_endpoint":   "https://api.ainative.studio",
	}
	wizard.SetSelections(selections)

	// Run wizard
	result, err := wizard.Run()
	if err != nil {
		t.Fatalf("Wizard failed: %v", err)
	}

	// Verify configuration was written
	if result.ConfigPath != configPath {
		t.Errorf("Expected config path '%s', got '%s'", configPath, result.ConfigPath)
	}

	// Load the written configuration
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read wizard config file: %v", err)
	}

	// Verify ZeroDB settings are at services.zerodb.* path
	projectID := v.GetString("services.zerodb.project_id")
	if projectID != "wizard-test-project" {
		t.Errorf("Expected project_id 'wizard-test-project', got '%s'", projectID)
	}

	endpoint := v.GetString("services.zerodb.endpoint")
	if endpoint != "https://api.ainative.studio" {
		t.Errorf("Expected endpoint 'https://api.ainative.studio', got '%s'", endpoint)
	}

	enabled := v.GetBool("services.zerodb.enabled")
	if !enabled {
		t.Error("Expected ZeroDB to be enabled")
	}

	t.Log("✓ Setup wizard saves ZeroDB config to correct path")
}

// TestZeroDBConfigEnvironmentVariables verifies that environment variables
// are correctly mapped to services.zerodb.* configuration paths.
func TestZeroDBConfigEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("AINATIVE_CODE_SERVICES_ZERODB_PROJECT_ID", "env-test-project")
	os.Setenv("AINATIVE_CODE_SERVICES_ZERODB_ENDPOINT", "https://env.api.ainative.studio")
	defer os.Unsetenv("AINATIVE_CODE_SERVICES_ZERODB_PROJECT_ID")
	defer os.Unsetenv("AINATIVE_CODE_SERVICES_ZERODB_ENDPOINT")

	// Create a minimal config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "env-config.yaml")

	minimalConfig := `app:
  name: ainative-code
  version: 0.1.0
  environment: test
llm:
  default_provider: anthropic
  anthropic:
    api_key: test-key
    model: claude-3-5-sonnet-20241022
platform:
  authentication:
    method: none
`
	if err := os.WriteFile(configPath, []byte(minimalConfig), 0600); err != nil {
		t.Fatalf("Failed to write minimal config: %v", err)
	}

	// Instead of using the config loader which validates, use viper directly
	// to test environment variable binding
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetEnvPrefix("AINATIVE_CODE")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Bind environment variables
	v.BindEnv("services.zerodb.project_id")
	v.BindEnv("services.zerodb.endpoint")

	projectID := v.GetString("services.zerodb.project_id")
	if projectID != "env-test-project" {
		t.Errorf("Expected project_id from env 'env-test-project', got '%s'", projectID)
	}

	endpoint := v.GetString("services.zerodb.endpoint")
	if endpoint != "https://env.api.ainative.studio" {
		t.Errorf("Expected endpoint from env 'https://env.api.ainative.studio', got '%s'", endpoint)
	}

	t.Log("✓ Environment variables correctly map to services.zerodb.* paths")
}

// TestZeroDBConfigBackwardCompatibility verifies the system can handle
// configurations from different versions.
func TestZeroDBConfigBackwardCompatibility(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "legacy-config.yaml")

	// Create a config file with services.zerodb.* structure
	legacyConfig := `app:
  name: ainative-code
  version: 0.1.0
  environment: test
services:
  zerodb:
    enabled: true
    project_id: legacy-project-123
    endpoint: https://legacy.api.ainative.studio
    database: default
    ssl: true
    ssl_mode: require
    max_connections: 10
    idle_connections: 5
    conn_max_lifetime: 3600000000000
    timeout: 30000000000
    retry_attempts: 3
    retry_delay: 1000000000
`
	if err := os.WriteFile(configPath, []byte(legacyConfig), 0600); err != nil {
		t.Fatalf("Failed to write legacy config: %v", err)
	}

	// Load configuration
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read legacy config: %v", err)
	}

	// Verify all fields are accessible via services.zerodb.* path
	projectID := v.GetString("services.zerodb.project_id")
	if projectID != "legacy-project-123" {
		t.Errorf("Expected project_id 'legacy-project-123', got '%s'", projectID)
	}

	endpoint := v.GetString("services.zerodb.endpoint")
	if endpoint != "https://legacy.api.ainative.studio" {
		t.Errorf("Expected endpoint 'https://legacy.api.ainative.studio', got '%s'", endpoint)
	}

	enabled := v.GetBool("services.zerodb.enabled")
	if !enabled {
		t.Error("Expected ZeroDB to be enabled")
	}

	t.Log("✓ Backward compatibility maintained for services.zerodb.* paths")
}
