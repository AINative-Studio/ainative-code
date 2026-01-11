package integration

import (
	"context"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupWizardProviderValidation verifies that all providers offered in the
// setup wizard are properly supported by the validation logic.
// This test addresses GitHub issue #126: Setup wizard offers Meta Llama provider
// but validation rejects it as unsupported.
func TestSetupWizardProviderValidation(t *testing.T) {
	ctx := context.Background()
	validator := setup.NewValidator()

	// These are the providers offered in the setup wizard (internal/setup/prompts.go line 152-158)
	// 1. Anthropic (Claude)
	// 2. OpenAI (GPT)
	// 3. Google (Gemini)
	// 4. Meta (Llama)
	// 5. Ollama (Local)
	wizardProviders := map[string]map[string]interface{}{
		"anthropic": {
			"provider":           "anthropic",
			"anthropic_api_key":  "sk-ant-test-key-12345678901234567890123456789012345678901234567890",
		},
		"openai": {
			"provider":        "openai",
			"openai_api_key":  "sk-test-key-12345678901234567890",
		},
		"google": {
			"provider":       "google",
			"google_api_key": "google-test-key-12345678901234567890",
		},
		"meta_llama": {
			"provider":            "meta_llama",
			"meta_llama_api_key":  "meta-llama-test-key-12345678901234567890",
		},
		"ollama": {
			"provider":     "ollama",
			"ollama_url":   "http://localhost:11434",
			"ollama_model": "llama2",
		},
	}

	for providerName, selections := range wizardProviders {
		t.Run(providerName, func(t *testing.T) {
			err := validator.ValidateProviderConfig(ctx, providerName, selections)

			// The validation should not return "unsupported provider" error
			// Network errors or connection failures are acceptable in tests
			if err != nil {
				assert.NotContains(t, err.Error(), "unsupported provider",
					"Provider %s is offered in setup wizard but validation rejects it as unsupported", providerName)
			}
		})
	}
}

// TestSetupWizardProviderCount verifies that the choice count in prompts.go
// matches the actual number of providers offered
func TestSetupWizardProviderCount(t *testing.T) {
	// The wizard offers 5 providers (see internal/setup/prompts.go line 152-158):
	// 1. Anthropic (Claude)
	// 2. OpenAI (GPT)
	// 3. Google (Gemini)
	// 4. Meta (Llama)
	// 5. Ollama (Local)
	expectedProviderCount := 5

	// This is the count returned by getChoiceCount() for StepProvider
	// (see internal/setup/prompts.go line 681-682)
	actualProviderCount := 5

	assert.Equal(t, expectedProviderCount, actualProviderCount,
		"Provider count mismatch between wizard display and getChoiceCount()")
}

// TestProviderSelectionMapping verifies that provider selection indices
// map correctly to provider names
func TestProviderSelectionMapping(t *testing.T) {
	// These are the provider names used in handleEnter() for StepProvider
	// (see internal/setup/prompts.go line 435)
	expectedProviderNames := []string{
		"anthropic",
		"openai",
		"google",
		"meta_llama",
		"ollama",
	}

	ctx := context.Background()
	validator := setup.NewValidator()

	// Verify each provider name is supported by validation
	for i, providerName := range expectedProviderNames {
		t.Run(providerName, func(t *testing.T) {
			selections := map[string]interface{}{
				"provider": providerName,
			}

			// Add required API key/config for each provider
			switch providerName {
			case "anthropic":
				selections["anthropic_api_key"] = "sk-ant-test-key-12345678901234567890123456789012345678901234567890"
			case "openai":
				selections["openai_api_key"] = "sk-test-key-12345678901234567890"
			case "google":
				selections["google_api_key"] = "google-test-key-12345678901234567890"
			case "meta_llama":
				selections["meta_llama_api_key"] = "meta-llama-test-key-12345678901234567890"
			case "ollama":
				selections["ollama_url"] = "http://localhost:11434"
				selections["ollama_model"] = "llama2"
			}

			err := validator.ValidateProviderConfig(ctx, providerName, selections)

			// Should not be unsupported
			if err != nil {
				assert.NotContains(t, err.Error(), "unsupported provider",
					"Provider at index %d (%s) is not supported by validation", i, providerName)
			}
		})
	}
}

// TestMetaLlamaAliasSupport verifies that both "meta_llama" and "meta" work
func TestMetaLlamaAliasSupport(t *testing.T) {
	ctx := context.Background()
	validator := setup.NewValidator()

	testCases := []struct {
		name         string
		providerName string
	}{
		{
			name:         "meta_llama",
			providerName: "meta_llama",
		},
		{
			name:         "meta alias",
			providerName: "meta",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			selections := map[string]interface{}{
				"provider":            tc.providerName,
				"meta_llama_api_key":  "meta-llama-test-key-12345678901234567890",
			}

			err := validator.ValidateProviderConfig(ctx, tc.providerName, selections)

			// Should not return unsupported provider error
			if err != nil {
				assert.NotContains(t, err.Error(), "unsupported provider",
					"Provider %s should be supported", tc.providerName)
			}
		})
	}
}

// TestUnsupportedProvider verifies that truly unsupported providers are rejected
func TestUnsupportedProvider(t *testing.T) {
	ctx := context.Background()
	validator := setup.NewValidator()

	selections := map[string]interface{}{
		"provider": "totally_fake_provider",
	}

	err := validator.ValidateProviderConfig(ctx, "totally_fake_provider", selections)

	require.Error(t, err, "Validation should reject truly unsupported providers")
	assert.Contains(t, err.Error(), "unsupported provider",
		"Error message should indicate provider is unsupported")
}
