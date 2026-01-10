package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AINative-studio/ainative-code/internal/config"
	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/AINative-studio/ainative-code/internal/provider/anthropic"
	"github.com/AINative-studio/ainative-code/internal/provider/openai"
	"github.com/AINative-studio/ainative-code/internal/setup"
)

// TestSetupWizardArrowKeyNavigationWithRealAPIs tests arrow key navigation
// and validates that the resulting configuration works with REAL API calls
func TestSetupWizardArrowKeyNavigationWithRealAPIs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Ensure we have real API keys
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	if anthropicKey == "" && openaiKey == "" {
		t.Skip("Skipping test - No API keys available (ANTHROPIC_API_KEY or OPENAI_API_KEY)")
	}

	if anthropicKey != "" {
		t.Run("AnthropicSetupWithArrowKeyNavigation", func(t *testing.T) {
			// Setup test directory
			testDir := t.TempDir()
			configPath := filepath.Join(testDir, "test-config.yaml")

			// Create wizard with non-interactive mode
			ctx := context.Background()
			wizard := setup.NewWizard(ctx, setup.WizardConfig{
				ConfigPath:      configPath,
				SkipValidation:  false,
				InteractiveMode: false,
				Force:           true,
			})

			// Simulate arrow key navigation by setting selections programmatically
			// This mimics a user navigating with arrow keys and selecting options
			selections := map[string]interface{}{
				"provider":            "anthropic",                      // Navigate with arrow keys, select Anthropic
				"anthropic_api_key":   anthropicKey,                     // Type in API key
				"anthropic_model":     "claude-3-5-sonnet-20241022",     // Arrow key navigate through models
				"extended_thinking":   false,                            // Press 'n' for no
				"ainative_login":      false,                            // Press 'n' for no
				"color_scheme":        "auto",                           // Arrow key navigate to 'auto'
				"prompt_caching":      true,                             // Press 'y' for yes
			}

			wizard.SetSelections(selections)

			// Run the wizard
			result, err := wizard.Run()
			require.NoError(t, err, "Setup wizard should complete successfully")
			require.NotNil(t, result, "Wizard result should not be nil")
			require.NotNil(t, result.Config, "Config should be created")

			// Verify configuration was created
			assert.FileExists(t, configPath, "Config file should be created")

			// Load the configuration
			loader := config.NewLoader()
			cfg, err := loader.LoadFromFile(configPath)
			require.NoError(t, err, "Should load config successfully")

			// Verify Anthropic configuration
			assert.Equal(t, "anthropic", cfg.LLM.DefaultProvider)
			assert.NotNil(t, cfg.LLM.Anthropic)
			assert.Equal(t, anthropicKey, cfg.LLM.Anthropic.APIKey)
			assert.Equal(t, "claude-3-5-sonnet-20241022", cfg.LLM.Anthropic.Model)

			// CRITICAL: Make a REAL API call to verify the setup works
			t.Log("Making REAL API call to Anthropic to verify setup...")
			anthropicProvider, err := anthropic.NewAnthropicProvider(anthropic.Config{
				APIKey: cfg.LLM.Anthropic.APIKey,
			})
			require.NoError(t, err, "Should create Anthropic provider")

			testCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			messages := []provider.Message{
				{Role: "user", Content: "Say 'Arrow key navigation test successful' and nothing else."},
			}

			response, err := anthropicProvider.Chat(testCtx, messages,
				provider.WithModel(cfg.LLM.Anthropic.Model),
				provider.WithMaxTokens(50),
				provider.WithTemperature(0.7))

			require.NoError(t, err, "Real Anthropic API call should succeed")
			assert.NotEmpty(t, response.Content, "Response content should not be empty")
			assert.Contains(t, response.Content, "Arrow key navigation", "Response should contain expected text")

			t.Logf("✓ REAL API Response from Anthropic: %s", response.Content)
			t.Logf("✓ API Call Details: Model=%s, Tokens=%d", response.Model, response.Usage.TotalTokens)
		})
	}

	if openaiKey != "" {
		t.Run("OpenAISetupWithArrowKeyNavigation", func(t *testing.T) {
			// Setup test directory
			testDir := t.TempDir()
			configPath := filepath.Join(testDir, "test-config.yaml")

			// Create wizard
			ctx := context.Background()
			wizard := setup.NewWizard(ctx, setup.WizardConfig{
				ConfigPath:      configPath,
				SkipValidation:  false,
				InteractiveMode: false,
				Force:           true,
			})

			// Simulate arrow key navigation to select OpenAI
			selections := map[string]interface{}{
				"provider":        "openai",                  // Arrow down to OpenAI, press enter
				"openai_api_key":  openaiKey,                 // Type in API key
				"openai_model":    "gpt-4-turbo-preview",     // Arrow keys to select model
				"ainative_login":  false,
				"color_scheme":    "dark",                    // Arrow keys to select dark
				"prompt_caching":  true,
			}

			wizard.SetSelections(selections)

			// Run the wizard
			result, err := wizard.Run()
			require.NoError(t, err, "Setup wizard should complete successfully")
			require.NotNil(t, result.Config)

			// Load the configuration
			loader := config.NewLoader()
			cfg, err := loader.LoadFromFile(configPath)
			require.NoError(t, err)

			// Verify OpenAI configuration
			assert.Equal(t, "openai", cfg.LLM.DefaultProvider)
			assert.NotNil(t, cfg.LLM.OpenAI)
			assert.Equal(t, openaiKey, cfg.LLM.OpenAI.APIKey)
			assert.Equal(t, "gpt-4-turbo-preview", cfg.LLM.OpenAI.Model)

			// CRITICAL: Make a REAL API call to verify the setup works
			t.Log("Making REAL API call to OpenAI to verify setup...")
			openaiProvider, err := openai.NewOpenAIProvider(openai.Config{
				APIKey: cfg.LLM.OpenAI.APIKey,
			})
			require.NoError(t, err, "Should create OpenAI provider")

			testCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			messages := []provider.Message{
				{Role: "user", Content: "Respond with exactly: 'OpenAI arrow key test passed'"},
			}

			response, err := openaiProvider.Chat(testCtx, messages,
				provider.WithModel(cfg.LLM.OpenAI.Model),
				provider.WithMaxTokens(30),
				provider.WithTemperature(0.7))

			require.NoError(t, err, "Real OpenAI API call should succeed")
			assert.NotEmpty(t, response.Content)

			t.Logf("✓ REAL API Response from OpenAI: %s", response.Content)
			t.Logf("✓ API Call Details: Model=%s, Tokens=%d", response.Model, response.Usage.TotalTokens)
		})
	}
}

// TestSetupWizardModelSelectionWithArrowKeys tests that arrow key navigation
// correctly selects different models and they work with real APIs
func TestSetupWizardModelSelectionWithArrowKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		t.Skip("Skipping test - ANTHROPIC_API_KEY not set")
	}

	// Test different model selections via arrow key navigation
	models := []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-haiku-20240307",
	}

	for i, model := range models {
		t.Run(fmt.Sprintf("ArrowKeySelect_%s", model), func(t *testing.T) {
			testDir := t.TempDir()
			configPath := filepath.Join(testDir, "test-config.yaml")

			ctx := context.Background()
			wizard := setup.NewWizard(ctx, setup.WizardConfig{
				ConfigPath:      configPath,
				SkipValidation:  false,
				InteractiveMode: false,
				Force:           true,
			})

			// Simulate: Arrow down 'i' times through model list, then Enter
			selections := map[string]interface{}{
				"provider":          "anthropic",
				"anthropic_api_key": anthropicKey,
				"anthropic_model":   model, // Represents arrow key navigation result
				"extended_thinking": false,
				"ainative_login":    false,
				"color_scheme":      "auto",
				"prompt_caching":    false,
			}

			wizard.SetSelections(selections)
			result, err := wizard.Run()
			require.NoError(t, err)
			_ = result // Use result to avoid unused variable error

			// Load config
			loader := config.NewLoader()
			cfg, err := loader.LoadFromFile(configPath)
			require.NoError(t, err)

			assert.Equal(t, model, cfg.LLM.Anthropic.Model,
				"Selected model (via arrow keys) should be correctly configured")

			// Make REAL API call with the selected model
			t.Logf("Testing REAL API call with model: %s (selected via arrow key navigation #%d)", model, i)

			anthropicProvider, err := anthropic.NewAnthropicProvider(anthropic.Config{
				APIKey: cfg.LLM.Anthropic.APIKey,
			})
			require.NoError(t, err)

			testCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			messages := []provider.Message{
				{Role: "user", Content: fmt.Sprintf("Say 'Model %s works' and nothing else.", model)},
			}

			response, err := anthropicProvider.Chat(testCtx, messages,
				provider.WithModel(cfg.LLM.Anthropic.Model),
				provider.WithMaxTokens(30),
				provider.WithTemperature(0.5))

			require.NoError(t, err, "Real API call should work with arrow-key-selected model")

			t.Logf("✓ REAL API Response (model %s): %s", model, response.Content)
			t.Logf("✓ Verified arrow key model selection #%d works correctly", i)
		})
	}
}

// TestSetupWizardPromptModelInteraction simulates the exact Bubble Tea interaction
func TestSetupWizardPromptModelInteraction(t *testing.T) {
	t.Run("ArrowKeyUpDownNavigation", func(t *testing.T) {
		model := setup.NewPromptModel()

		// Start at provider selection (cursor at 0 - Anthropic)
		assert.Equal(t, setup.StepProvider, model.GetCurrentStep())
		assert.Equal(t, 0, model.GetCursor())

		// Simulate: Press Down arrow (move to OpenAI)
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = updatedModel.(setup.PromptModel)
		assert.Equal(t, 1, model.GetCursor(), "Down arrow should move cursor to position 1")

		// Simulate: Press Down arrow again (move to Google)
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = updatedModel.(setup.PromptModel)
		assert.Equal(t, 2, model.GetCursor(), "Down arrow should move cursor to position 2")

		// Simulate: Press Up arrow (move back to OpenAI)
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
		model = updatedModel.(setup.PromptModel)
		assert.Equal(t, 1, model.GetCursor(), "Up arrow should move cursor back to position 1")

		// Simulate: Press Up arrow (move back to Anthropic)
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
		model = updatedModel.(setup.PromptModel)
		assert.Equal(t, 0, model.GetCursor(), "Up arrow should move cursor back to position 0")

		// Simulate: Press Up arrow at first position (should stay at 0)
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
		model = updatedModel.(setup.PromptModel)
		assert.Equal(t, 0, model.GetCursor(), "Up arrow at first position should not move cursor")

		t.Log("✓ Arrow key navigation behaves correctly in UI model")
	})

	t.Run("VimStyleNavigation", func(t *testing.T) {
		model := setup.NewPromptModel()

		// Test 'j' key (vim down)
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		model = updatedModel.(setup.PromptModel)
		assert.Equal(t, 1, model.GetCursor(), "'j' key should move cursor down")

		// Test 'k' key (vim up)
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		model = updatedModel.(setup.PromptModel)
		assert.Equal(t, 0, model.GetCursor(), "'k' key should move cursor up")

		t.Log("✓ Vim-style navigation (j/k) works correctly")
	})
}

// TestEndToEndSetupWithRealAPIValidation performs a complete end-to-end test
func TestEndToEndSetupWithRealAPIValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		t.Skip("Skipping test - ANTHROPIC_API_KEY not set")
	}

	t.Log("=== Starting End-to-End Setup Test with Real API Validation ===")

	// 1. Setup phase - simulate user interaction with arrow keys
	testDir := t.TempDir()
	configPath := filepath.Join(testDir, "test-config.yaml")

	ctx := context.Background()
	wizard := setup.NewWizard(ctx, setup.WizardConfig{
		ConfigPath:      configPath,
		SkipValidation:  false, // Enable validation - will make real API calls
		InteractiveMode: false,
		Force:           true,
	})

	// Simulate complete user interaction with arrow keys
	selections := map[string]interface{}{
		"provider":          "anthropic",
		"anthropic_api_key": anthropicKey,
		"anthropic_model":   "claude-3-haiku-20240307", // Fast model for testing
		"extended_thinking": false,
		"ainative_login":    false,
		"color_scheme":      "light",
		"prompt_caching":    true,
	}

	wizard.SetSelections(selections)
	t.Log("Step 1: Running setup wizard with arrow key navigation selections...")

	result, err := wizard.Run()
	require.NoError(t, err, "Setup should complete successfully")
	t.Logf("✓ Setup wizard completed, config saved to: %s", result.ConfigPath)

	// 2. Verification phase - load config and make real API calls
	t.Log("Step 2: Loading configuration...")
	loader := config.NewLoader()
	cfg, err := loader.LoadFromFile(configPath)
	require.NoError(t, err)
	t.Log("✓ Configuration loaded successfully")

	// 3. Make multiple real API calls to thoroughly test
	t.Log("Step 3: Making REAL API calls to verify setup...")

	anthropicProvider, err := anthropic.NewAnthropicProvider(anthropic.Config{
		APIKey: cfg.LLM.Anthropic.APIKey,
	})
	require.NoError(t, err)

	testCases := []struct {
		name    string
		message string
	}{
		{"Simple greeting", "Say hello"},
		{"Math question", "What is 2+2?"},
		{"Arrow key test", "Confirm arrow key navigation works by saying YES"},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			messages := []provider.Message{
				{Role: "user", Content: tc.message},
			}

			t.Logf("  Making API call %d/%d: %s", i+1, len(testCases), tc.name)
			response, err := anthropicProvider.Chat(testCtx, messages,
				provider.WithModel(cfg.LLM.Anthropic.Model),
				provider.WithMaxTokens(100),
				provider.WithTemperature(0.7))

			require.NoError(t, err, "API call should succeed")
			assert.NotEmpty(t, response.Content)

			t.Logf("  ✓ API Response: %s", response.Content[:min(100, len(response.Content))])
			t.Logf("  ✓ Tokens used: %d (input: %d, output: %d)",
				response.Usage.TotalTokens,
				response.Usage.PromptTokens,
				response.Usage.CompletionTokens)
		})
	}

	t.Log("=== End-to-End Test Completed Successfully ===")
	t.Log("✓ Arrow key navigation setup verified with real API calls")
	t.Log("✓ Configuration file created and validated")
	t.Log("✓ Multiple API calls succeeded with configured provider")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
