// Package setup provides an interactive first-time setup wizard for AINative Code.
//
// The setup wizard guides new users through the initial configuration process,
// including:
//
//   - LLM provider selection (Anthropic, OpenAI, Google Gemini, Ollama)
//   - API key input and validation
//   - Model selection with recommendations
//   - Optional AINative platform integration
//   - Extended thinking configuration
//   - Color scheme and UI preferences
//
// The wizard creates a configuration file at ~/.ainative-code.yaml and an
// initialization marker at ~/.ainative-code-initialized to track first-time setup.
//
// # Interactive Mode
//
// By default, the wizard runs in interactive mode with a rich TUI powered by
// Bubble Tea. Users can navigate through steps using arrow keys and make
// selections interactively:
//
//	wizard := setup.NewWizard(ctx, setup.WizardConfig{
//	    InteractiveMode: true,
//	    SkipValidation:  false,
//	})
//	result, err := wizard.Run()
//
// # Non-Interactive Mode
//
// For CI/CD environments or automated setups, the wizard supports non-interactive
// mode using environment variables:
//
//	export AINATIVE_PROVIDER=anthropic
//	export AINATIVE_ANTHROPIC_API_KEY=sk-ant-...
//	export AINATIVE_ANTHROPIC_MODEL=claude-3-5-sonnet-20241022
//	ainative-code setup --non-interactive
//
// # Validation
//
// The wizard validates API keys and credentials by default:
//
//   - Anthropic: Tests API key format and authentication
//   - OpenAI: Validates key format and tests API access
//   - Google: Checks API key validity
//   - Ollama: Verifies server connectivity and model availability
//
// Validation can be skipped for faster setup:
//
//	ainative-code setup --skip-validation
//
// # First-Run Detection
//
// The wizard automatically detects first-time runs by checking for the
// initialization marker file. Users can force re-run setup:
//
//	ainative-code setup --force
//
// # Configuration Structure
//
// The generated configuration includes:
//
//   - App metadata (name, version, environment)
//   - LLM provider settings with defaults
//   - Platform authentication (optional)
//   - Performance tuning (cache, concurrency)
//   - Logging configuration
//   - Security settings
//
// # Provider-Specific Configuration
//
// Anthropic (Claude):
//   - API key validation (must start with "sk-ant-")
//   - Model selection (Sonnet 3.5, Opus 3, etc.)
//   - Extended thinking mode
//   - Prompt caching preferences
//
// OpenAI (GPT):
//   - API key validation (must start with "sk-")
//   - Model selection (GPT-4 Turbo, GPT-4, GPT-3.5)
//   - Organization ID (optional)
//
// Google (Gemini):
//   - API key validation
//   - Model selection (Gemini Pro, Gemini Pro Vision)
//   - Project ID and location (optional)
//
// Ollama (Local):
//   - Server URL (default: http://localhost:11434)
//   - Model name (must be pre-downloaded)
//   - Connection validation
//
// # Error Handling
//
// The wizard handles errors gracefully:
//
//   - Invalid API keys: Clear error messages with help links
//   - Network failures: Distinguishes from validation failures
//   - User cancellation: Clean exit without partial state
//   - File system errors: Directory creation and permissions
//
// # Testing
//
// The package includes comprehensive tests for:
//
//   - Configuration building for all providers
//   - API key validation
//   - File operations (writing config, creating markers)
//   - First-run detection
//   - Non-interactive mode
//   - Error cases
//
// Example usage in tests:
//
//	wizard := setup.NewWizard(ctx, setup.WizardConfig{
//	    InteractiveMode: false,
//	    SkipValidation:  true,
//	})
//	wizard.SetSelections(map[string]interface{}{
//	    "provider":          "anthropic",
//	    "anthropic_api_key": "sk-ant-test",
//	})
//	result, err := wizard.Run()
package setup
