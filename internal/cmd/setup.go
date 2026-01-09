package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/AINative-studio/ainative-code/internal/setup"
)

var (
	setupSkipValidation bool
	setupConfigPath     string
	setupForce          bool
	setupNonInteractive bool
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run first-time setup wizard",
	Long: `Interactive first-time setup wizard for AINative Code.

This wizard will guide you through:
  - Selecting your preferred LLM provider (Anthropic, OpenAI, Google, Ollama)
  - Configuring API keys and credentials
  - Setting up optional AINative platform integration
  - Customizing your development experience

The wizard creates a configuration file at ~/.ainative-code.yaml
and a marker file at ~/.ainative-code-initialized to track first-time setup.

Examples:
  # Run interactive setup wizard
  ainative-code setup

  # Run setup with custom config path
  ainative-code setup --config ~/my-config.yaml

  # Force re-run setup (overwrites existing config)
  ainative-code setup --force

  # Skip API key validation (faster setup)
  ainative-code setup --skip-validation

Advanced Usage:
  # Skip setup wizard (for CI/CD or advanced users)
  ainative-code --skip-setup chat

  # Non-interactive mode (requires environment variables)
  ainative-code setup --non-interactive`,
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.Flags().BoolVar(&setupSkipValidation, "skip-validation", false, "skip API key validation")
	setupCmd.Flags().StringVar(&setupConfigPath, "config", "", "custom config file path")
	setupCmd.Flags().BoolVarP(&setupForce, "force", "f", false, "force re-run setup and overwrite existing config")
	setupCmd.Flags().BoolVar(&setupNonInteractive, "non-interactive", false, "run in non-interactive mode (uses env vars)")
}

func runSetup(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
	defer cancel()

	logger.InfoEvent().Msg("Starting setup wizard")

	// Check if already initialized - verify BOTH marker AND config file exist
	if !setupForce {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configPath := filepath.Join(homeDir, ".ainative-code.yaml")
			markerPath := filepath.Join(homeDir, ".ainative-code-initialized")

			// Only skip setup if BOTH marker AND config file exist
			_, markerErr := os.Stat(markerPath)
			_, configErr := os.Stat(configPath)

			if markerErr == nil && configErr == nil {
				return handleAlreadyInitialized(cmd)
			}
		}
	}

	// Configure wizard
	wizardConfig := setup.WizardConfig{
		ConfigPath:      setupConfigPath,
		SkipValidation:  setupSkipValidation,
		InteractiveMode: !setupNonInteractive,
		Force:           setupForce,
	}

	// Create and run wizard
	wizard := setup.NewWizard(ctx, wizardConfig)
	result, err := wizard.Run()
	if err != nil {
		logger.ErrorEvent().Err(err).Msg("Setup wizard failed")
		return fmt.Errorf("setup failed: %w", err)
	}

	// Log successful setup
	logger.InfoEvent().
		Str("config_path", result.ConfigPath).
		Bool("validation_passed", result.ValidationPass).
		Msg("Setup completed successfully")

	return nil
}

// handleAlreadyInitialized handles the case where setup has already been run
func handleAlreadyInitialized(cmd *cobra.Command) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := setupConfigPath
	if configPath == "" {
		configPath = fmt.Sprintf("%s/.ainative-code.yaml", homeDir)
	}

	fmt.Println("AINative Code is already configured!")
	fmt.Printf("\nConfiguration file: %s\n", configPath)
	fmt.Println("\nWhat would you like to do?")
	fmt.Println("  1. View current configuration: ainative-code config show")
	fmt.Println("  2. Edit configuration manually: edit ~/.ainative-code.yaml")
	fmt.Println("  3. Re-run setup wizard: ainative-code setup --force")
	fmt.Println("  4. Start using the CLI: ainative-code chat")
	fmt.Println()

	return nil
}
