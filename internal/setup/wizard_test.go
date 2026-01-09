package setup

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/AINative-studio/ainative-code/internal/config"
)

func TestNewWizard(t *testing.T) {
	ctx := context.Background()
	cfg := WizardConfig{
		InteractiveMode: false,
		SkipValidation:  true,
	}

	wizard := NewWizard(ctx, cfg)

	assert.NotNil(t, wizard)
	assert.NotNil(t, wizard.userSelections)
	assert.NotNil(t, wizard.result)
	assert.Equal(t, cfg, wizard.config)
}

func TestCheckFirstRun(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// First run - no marker file
	firstRun := CheckFirstRun()
	assert.True(t, firstRun, "Should detect first run when marker doesn't exist")

	// Create marker file
	markerPath := filepath.Join(tempDir, ".ainative-code-initialized")
	err := os.WriteFile(markerPath, []byte("test"), 0644)
	require.NoError(t, err)

	// Not first run - marker exists
	firstRun = CheckFirstRun()
	assert.False(t, firstRun, "Should not detect first run when marker exists")
}

func TestBuildConfiguration_Anthropic(t *testing.T) {
	ctx := context.Background()
	wizard := NewWizard(ctx, WizardConfig{
		InteractiveMode: false,
		SkipValidation:  true,
	})

	// Set Anthropic selections
	wizard.userSelections = map[string]interface{}{
		"provider":          "anthropic",
		"anthropic_api_key": "sk-ant-test123",
		"anthropic_model":   "claude-3-5-sonnet-20241022",
		"extended_thinking": true,
	}

	err := wizard.buildConfiguration()
	require.NoError(t, err)

	cfg := wizard.result.Config
	assert.Equal(t, "anthropic", cfg.LLM.DefaultProvider)
	assert.NotNil(t, cfg.LLM.Anthropic)
	assert.Equal(t, "sk-ant-test123", cfg.LLM.Anthropic.APIKey)
	assert.Equal(t, "claude-3-5-sonnet-20241022", cfg.LLM.Anthropic.Model)
	assert.NotNil(t, cfg.LLM.Anthropic.ExtendedThinking)
	assert.True(t, cfg.LLM.Anthropic.ExtendedThinking.Enabled)
}

func TestBuildConfiguration_OpenAI(t *testing.T) {
	ctx := context.Background()
	wizard := NewWizard(ctx, WizardConfig{
		InteractiveMode: false,
		SkipValidation:  true,
	})

	// Set OpenAI selections
	wizard.userSelections = map[string]interface{}{
		"provider":        "openai",
		"openai_api_key":  "sk-test123",
		"openai_model":    "gpt-4-turbo-preview",
	}

	err := wizard.buildConfiguration()
	require.NoError(t, err)

	cfg := wizard.result.Config
	assert.Equal(t, "openai", cfg.LLM.DefaultProvider)
	assert.NotNil(t, cfg.LLM.OpenAI)
	assert.Equal(t, "sk-test123", cfg.LLM.OpenAI.APIKey)
	assert.Equal(t, "gpt-4-turbo-preview", cfg.LLM.OpenAI.Model)
}

func TestBuildConfiguration_Google(t *testing.T) {
	ctx := context.Background()
	wizard := NewWizard(ctx, WizardConfig{
		InteractiveMode: false,
		SkipValidation:  true,
	})

	// Set Google selections
	wizard.userSelections = map[string]interface{}{
		"provider":       "google",
		"google_api_key": "test-google-key",
		"google_model":   "gemini-pro",
	}

	err := wizard.buildConfiguration()
	require.NoError(t, err)

	cfg := wizard.result.Config
	assert.Equal(t, "google", cfg.LLM.DefaultProvider)
	assert.NotNil(t, cfg.LLM.Google)
	assert.Equal(t, "test-google-key", cfg.LLM.Google.APIKey)
	assert.Equal(t, "gemini-pro", cfg.LLM.Google.Model)
}

func TestBuildConfiguration_Ollama(t *testing.T) {
	ctx := context.Background()
	wizard := NewWizard(ctx, WizardConfig{
		InteractiveMode: false,
		SkipValidation:  true,
	})

	// Set Ollama selections
	wizard.userSelections = map[string]interface{}{
		"provider":     "ollama",
		"ollama_url":   "http://localhost:11434",
		"ollama_model": "llama2",
	}

	err := wizard.buildConfiguration()
	require.NoError(t, err)

	cfg := wizard.result.Config
	assert.Equal(t, "ollama", cfg.LLM.DefaultProvider)
	assert.NotNil(t, cfg.LLM.Ollama)
	assert.Equal(t, "http://localhost:11434", cfg.LLM.Ollama.BaseURL)
	assert.Equal(t, "llama2", cfg.LLM.Ollama.Model)
}

func TestBuildConfiguration_WithAINativePlatform(t *testing.T) {
	ctx := context.Background()
	wizard := NewWizard(ctx, WizardConfig{
		InteractiveMode: false,
		SkipValidation:  true,
	})

	// Set selections with AINative platform
	wizard.userSelections = map[string]interface{}{
		"provider":          "anthropic",
		"anthropic_api_key": "sk-ant-test123",
		"anthropic_model":   "claude-3-5-sonnet-20241022",
		"ainative_login":    true,
		"ainative_api_key":  "ainative-test-key",
	}

	err := wizard.buildConfiguration()
	require.NoError(t, err)

	cfg := wizard.result.Config
	assert.Equal(t, "ainative-test-key", cfg.Platform.Authentication.APIKey)
}

func TestWriteConfiguration(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	ctx := context.Background()
	wizard := NewWizard(ctx, WizardConfig{
		ConfigPath:      configPath,
		InteractiveMode: false,
		SkipValidation:  true,
	})

	// Build a simple configuration
	wizard.userSelections = map[string]interface{}{
		"provider":          "anthropic",
		"anthropic_api_key": "sk-ant-test123",
		"anthropic_model":   "claude-3-5-sonnet-20241022",
	}

	err := wizard.buildConfiguration()
	require.NoError(t, err)

	// Write configuration
	err = wizard.writeConfiguration()
	require.NoError(t, err)

	// Verify file exists
	assert.FileExists(t, configPath)

	// Verify content
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var cfg config.Config
	err = yaml.Unmarshal(data, &cfg)
	require.NoError(t, err)

	assert.Equal(t, "anthropic", cfg.LLM.DefaultProvider)
	assert.NotNil(t, cfg.LLM.Anthropic)
}

func TestCreateMarker(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	ctx := context.Background()
	wizard := NewWizard(ctx, WizardConfig{
		InteractiveMode: false,
		SkipValidation:  true,
	})

	err := wizard.createMarker()
	require.NoError(t, err)

	markerPath := filepath.Join(tempDir, ".ainative-code-initialized")
	assert.FileExists(t, markerPath)
	assert.True(t, wizard.result.MarkerCreated)
}

func TestSetSelections(t *testing.T) {
	ctx := context.Background()
	wizard := NewWizard(ctx, WizardConfig{
		InteractiveMode: false,
	})

	selections := map[string]interface{}{
		"provider":          "anthropic",
		"anthropic_api_key": "sk-ant-test",
	}

	wizard.SetSelections(selections)
	assert.Equal(t, selections, wizard.userSelections)
}

func TestBuildConfiguration_DefaultValues(t *testing.T) {
	ctx := context.Background()
	wizard := NewWizard(ctx, WizardConfig{
		InteractiveMode: false,
		SkipValidation:  true,
	})

	// Minimal selections
	wizard.userSelections = map[string]interface{}{
		"provider":          "anthropic",
		"anthropic_api_key": "sk-ant-test",
	}

	err := wizard.buildConfiguration()
	require.NoError(t, err)

	cfg := wizard.result.Config

	// Check default app config
	assert.Equal(t, "ainative-code", cfg.App.Name)
	assert.Equal(t, "0.1.0", cfg.App.Version)
	assert.Equal(t, "development", cfg.App.Environment)

	// Check default logging config
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)

	// Check Anthropic model default
	assert.Equal(t, "claude-3-5-sonnet-20241022", cfg.LLM.Anthropic.Model)
	assert.Equal(t, 4096, cfg.LLM.Anthropic.MaxTokens)
	assert.Equal(t, 0.7, cfg.LLM.Anthropic.Temperature)
}

func TestBuildConfiguration_NoExtendedThinking(t *testing.T) {
	ctx := context.Background()
	wizard := NewWizard(ctx, WizardConfig{
		InteractiveMode: false,
		SkipValidation:  true,
	})

	// Set selections without extended thinking
	wizard.userSelections = map[string]interface{}{
		"provider":          "anthropic",
		"anthropic_api_key": "sk-ant-test",
		"extended_thinking": false,
	}

	err := wizard.buildConfiguration()
	require.NoError(t, err)

	cfg := wizard.result.Config
	assert.Nil(t, cfg.LLM.Anthropic.ExtendedThinking)
}

func TestCheckAlreadyInitialized(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	ctx := context.Background()
	wizard := NewWizard(ctx, WizardConfig{
		InteractiveMode: false,
	})

	// Not initialized yet
	isInitialized := wizard.checkAlreadyInitialized()
	assert.False(t, isInitialized)

	// Create marker and config
	markerPath := filepath.Join(tempDir, ".ainative-code-initialized")
	err := os.WriteFile(markerPath, []byte("test"), 0644)
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, ".ainative-code.yaml")
	err = os.WriteFile(configPath, []byte("test: config"), 0644)
	require.NoError(t, err)

	// Now should be initialized
	isInitialized = wizard.checkAlreadyInitialized()
	assert.True(t, isInitialized)
	assert.True(t, wizard.result.SkippedSetup)
	assert.Equal(t, configPath, wizard.result.ConfigPath)
}

func TestForceFlag_BypassesInitializedCheck(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create marker and config to simulate already initialized state
	markerPath := filepath.Join(tempDir, ".ainative-code-initialized")
	err := os.WriteFile(markerPath, []byte("test"), 0644)
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, ".ainative-code.yaml")
	err = os.WriteFile(configPath, []byte("test: config"), 0644)
	require.NoError(t, err)

	ctx := context.Background()

	// Test WITHOUT Force flag - should skip setup
	wizardNoForce := NewWizard(ctx, WizardConfig{
		InteractiveMode: false,
		SkipValidation:  true,
		Force:           false,
	})

	wizardNoForce.userSelections = map[string]interface{}{
		"provider":          "anthropic",
		"anthropic_api_key": "sk-ant-test",
	}

	// Run should return early when already initialized
	result, err := wizardNoForce.Run()
	require.NoError(t, err)
	assert.True(t, result.SkippedSetup, "Should skip setup when Force=false and already initialized")

	// Test WITH Force flag - should run wizard
	wizardWithForce := NewWizard(ctx, WizardConfig{
		InteractiveMode: false,
		SkipValidation:  true,
		Force:           true,
		ConfigPath:      filepath.Join(tempDir, "forced-config.yaml"),
	})

	wizardWithForce.userSelections = map[string]interface{}{
		"provider":          "anthropic",
		"anthropic_api_key": "sk-ant-forced",
	}

	// Run should complete wizard even when already initialized
	result, err = wizardWithForce.Run()
	require.NoError(t, err)
	assert.False(t, result.SkippedSetup, "Should NOT skip setup when Force=true")
	assert.True(t, result.MarkerCreated, "Should create marker when Force=true")
	assert.NotEmpty(t, result.ConfigPath, "Should write config when Force=true")

	// Verify the forced config was actually written
	assert.FileExists(t, result.ConfigPath)
}

func TestForceFlag_WorksOnFreshInstall(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	ctx := context.Background()

	// Test Force flag on fresh install (no existing config)
	wizard := NewWizard(ctx, WizardConfig{
		InteractiveMode: false,
		SkipValidation:  true,
		Force:           true,
		ConfigPath:      filepath.Join(tempDir, "fresh-config.yaml"),
	})

	wizard.userSelections = map[string]interface{}{
		"provider":          "anthropic",
		"anthropic_api_key": "sk-ant-fresh",
	}

	// Run should complete normally
	result, err := wizard.Run()
	require.NoError(t, err)
	assert.False(t, result.SkippedSetup, "Should not skip setup")
	assert.True(t, result.MarkerCreated, "Should create marker")
	assert.NotEmpty(t, result.ConfigPath, "Should write config")
	assert.FileExists(t, result.ConfigPath)
}
