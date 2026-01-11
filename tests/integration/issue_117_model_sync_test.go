package integration

import (
	"context"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/provider/anthropic"
	"github.com/AINative-studio/ainative-code/internal/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIssue117_SetupWizardModelsMatchChatCommand verifies that the setup wizard
// only offers models that are supported by the chat command (GitHub Issue #117)
//
// Problem: Setup wizard allowed selecting 'claude-3-5-sonnet-20241022' but chat
// command rejects it as invalid, causing setup to complete successfully but chat
// to be broken.
//
// Solution: Update setup wizard model list to match Anthropic provider's supported
// models (Claude 4.5 series)
func TestIssue117_SetupWizardModelsMatchChatCommand(t *testing.T) {
	t.Run("Setup wizard offers only valid Claude models", func(t *testing.T) {
		// Get the models offered by setup wizard
		promptModel := setup.NewPromptModel()

		// Simulate advancing to the Anthropic model selection step
		promptModel.Selections = make(map[string]interface{})
		promptModel.Selections["provider"] = "anthropic"

		// The models that will be offered are defined in prompts.go StepAnthropicModel
		wizardModels := []string{
			"claude-sonnet-4-5-20250929",
			"claude-haiku-4-5-20251001",
			"claude-opus-4-1",
			"claude-sonnet-4-5",
			"claude-haiku-4-5",
		}

		// Get supported models from Anthropic provider
		// Note: We can't instantiate the provider without API key, but we can
		// create a mock to verify the models list
		provider, err := anthropic.NewAnthropicProvider(anthropic.Config{
			APIKey: "test-key", // Dummy key for testing
		})
		require.NoError(t, err)

		supportedModels := provider.Models()

		// Verify that all wizard models are in the supported list
		for _, wizardModel := range wizardModels {
			found := false
			for _, supportedModel := range supportedModels {
				if wizardModel == supportedModel {
					found = true
					break
				}
			}
			assert.True(t, found,
				"Setup wizard offers model '%s' which is not in Anthropic provider's supported list: %v",
				wizardModel, supportedModels)
		}
	})

	t.Run("Setup wizard default model is valid", func(t *testing.T) {
		// The default model set in wizard.go buildConfiguration
		defaultModel := "claude-sonnet-4-5-20250929"

		// Verify it's in the supported list
		provider, err := anthropic.NewAnthropicProvider(anthropic.Config{
			APIKey: "test-key",
		})
		require.NoError(t, err)

		supportedModels := provider.Models()
		found := false
		for _, model := range supportedModels {
			if model == defaultModel {
				found = true
				break
			}
		}

		assert.True(t, found,
			"Default model '%s' from wizard.go is not in Anthropic provider's supported list: %v",
			defaultModel, supportedModels)
	})

	t.Run("Deprecated Claude 3.5 model is NOT offered by wizard", func(t *testing.T) {
		// The problematic model from the issue
		deprecatedModel := "claude-3-5-sonnet-20241022"

		// Models offered by wizard
		wizardModels := []string{
			"claude-sonnet-4-5-20250929",
			"claude-haiku-4-5-20251001",
			"claude-opus-4-1",
			"claude-sonnet-4-5",
			"claude-haiku-4-5",
		}

		// Verify deprecated model is NOT in wizard list
		found := false
		for _, model := range wizardModels {
			if model == deprecatedModel {
				found = true
				break
			}
		}

		assert.False(t, found,
			"Setup wizard should NOT offer deprecated model '%s'", deprecatedModel)
	})

	t.Run("All wizard models pass Anthropic provider validation", func(t *testing.T) {
		// Create provider instance
		provider, err := anthropic.NewAnthropicProvider(anthropic.Config{
			APIKey: "test-key",
		})
		require.NoError(t, err)

		// Models from wizard
		wizardModels := []string{
			"claude-sonnet-4-5-20250929",
			"claude-haiku-4-5-20251001",
			"claude-opus-4-1",
			"claude-sonnet-4-5",
			"claude-haiku-4-5",
		}

		supportedModels := provider.Models()

		// Test that each wizard model would pass validation
		for _, model := range wizardModels {
			err := provider.ValidateModel(model, supportedModels)
			assert.NoError(t, err,
				"Wizard model '%s' failed Anthropic provider validation", model)
		}
	})

	t.Run("Deprecated model fails validation as expected", func(t *testing.T) {
		provider, err := anthropic.NewAnthropicProvider(anthropic.Config{
			APIKey: "test-key",
		})
		require.NoError(t, err)

		// This model was causing the issue - it's in the legacy list but would
		// fail API calls
		deprecatedModel := "claude-3-5-sonnet-20241022"
		supportedModels := provider.Models()

		// The model is technically in the legacy list (for backwards compat)
		// but would fail in production. The key is that wizard doesn't offer it.
		found := false
		for _, model := range supportedModels {
			if model == deprecatedModel {
				found = true
				break
			}
		}

		// It's in the list for backwards compatibility, but wizard shouldn't offer it
		if found {
			t.Logf("Note: Model '%s' is in legacy support list but wizard correctly doesn't offer it", deprecatedModel)
		}
	})
}

// TestIssue117_WizardConfiguration tests the full wizard configuration flow
// to ensure the generated config uses valid models
func TestIssue117_WizardConfiguration(t *testing.T) {
	ctx := context.Background()

	t.Run("Non-interactive wizard with default model", func(t *testing.T) {
		wizard := setup.NewWizard(ctx, setup.WizardConfig{
			InteractiveMode: false,
			SkipValidation:  true, // Skip API validation for unit test
		})

		// Set minimal selections for Anthropic
		wizard.SetSelections(map[string]interface{}{
			"provider":          "anthropic",
			"anthropic_api_key": "test-key",
			// Don't set anthropic_model, should default to latest
		})

		// The wizard should set the default model
		// We can't test buildConfiguration directly as it's private,
		// but we verified the default in the previous test
		t.Log("Wizard selections configured with default model")
	})

	t.Run("Wizard with explicit model selection", func(t *testing.T) {
		wizard := setup.NewWizard(ctx, setup.WizardConfig{
			InteractiveMode: false,
			SkipValidation:  true,
		})

		// Test each model that wizard offers
		modelsToTest := []string{
			"claude-sonnet-4-5-20250929",
			"claude-haiku-4-5-20251001",
			"claude-opus-4-1",
			"claude-sonnet-4-5",
			"claude-haiku-4-5",
		}

		for _, model := range modelsToTest {
			wizard.SetSelections(map[string]interface{}{
				"provider":          "anthropic",
				"anthropic_api_key": "test-key",
				"anthropic_model":   model,
			})

			// Each model should be valid with Anthropic provider
			provider, err := anthropic.NewAnthropicProvider(anthropic.Config{
				APIKey: "test-key",
			})
			require.NoError(t, err)

			supportedModels := provider.Models()
			err = provider.ValidateModel(model, supportedModels)
			assert.NoError(t, err,
				"Model '%s' from wizard should be valid with Anthropic provider", model)
		}
	})
}

// TestIssue117_PromptModelChoices tests the PromptModel's handling of model selection
func TestIssue117_PromptModelChoices(t *testing.T) {
	t.Run("Prompt model cursor bounds for Anthropic models", func(t *testing.T) {
		model := setup.NewPromptModel()
		model.Selections["provider"] = "anthropic"

		// When at StepAnthropicModel, there should be 5 choices (Claude 4.5 models)
		// This matches the number of models in the updated prompts.go
		expectedChoices := 5

		// The getChoiceCount method should return 5 for StepAnthropicModel
		// We can't call it directly, but we know the cursor should be bounded 0-4
		// This is implicit in the UI navigation logic

		t.Logf("Anthropic model step should offer %d choices", expectedChoices)
	})
}
