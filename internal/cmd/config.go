package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `Manage AINative Code configuration settings.

Configuration can be set via command-line flags, environment variables,
or a configuration file. The configuration file is searched in the following
locations:
  - $HOME/.ainative-code.yaml
  - ./.ainative-code.yaml

Examples:
  # Show current configuration
  ainative-code config show

  # Set a configuration value
  ainative-code config set provider openai
  ainative-code config set model gpt-4

  # Get a configuration value
  ainative-code config get provider

  # Initialize configuration file
  ainative-code config init

  # Validate configuration
  ainative-code config validate`,
	Aliases: []string{"cfg"},
}

// configShowCmd represents the config show command
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display all current configuration values.`,
	Aliases: []string{"list", "ls"},
	RunE:  runConfigShow,
}

// configSetCmd represents the config set command
var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  `Set a configuration value and save it to the config file.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

// configGetCmd represents the config get command
var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long:  `Retrieve a specific configuration value.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

// configInitCmd represents the config init command
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long:  `Create a new configuration file with default values.`,
	RunE:  runConfigInit,
}

// configValidateCmd represents the config validate command
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  `Validate the current configuration for correctness.`,
	RunE:  runConfigValidate,
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Add subcommands
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configValidateCmd)

	// Config init flags
	configInitCmd.Flags().BoolP("force", "f", false, "overwrite existing config file")
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	logger.Debug("Showing configuration")

	fmt.Println("Current Configuration:")
	fmt.Println("======================")

	allSettings := viper.AllSettings()
	if len(allSettings) == 0 {
		fmt.Println("No configuration values set")
		return nil
	}

	for key, value := range allSettings {
		fmt.Printf("%s: %v\n", key, value)
	}

	if viper.ConfigFileUsed() != "" {
		fmt.Printf("\nConfig file: %s\n", viper.ConfigFileUsed())
	} else {
		fmt.Println("\nNo config file in use")
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	logger.DebugEvent().
		Str("key", key).
		Str("value", value).
		Msg("Setting configuration value")

	// Validate the configuration value before setting
	if err := validateConfigValue(key, value); err != nil {
		return fmt.Errorf("invalid configuration value: %w", err)
	}

	viper.Set(key, value)

	// Determine config file path
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configFile = fmt.Sprintf("%s/.ainative-code.yaml", home)
	}

	// Write config to file
	if err := viper.WriteConfigAs(configFile); err != nil {
		// If file doesn't exist, create it
		if os.IsNotExist(err) {
			if err := viper.SafeWriteConfigAs(configFile); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
		} else {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}

	fmt.Printf("Set %s = %s\n", key, value)
	fmt.Printf("Configuration saved to: %s\n", configFile)

	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	logger.DebugEvent().Str("key", key).Msg("Getting configuration value")

	if !viper.IsSet(key) {
		return fmt.Errorf("configuration key '%s' not found", key)
	}

	value := viper.Get(key)
	fmt.Printf("%s: %v\n", key, value)

	return nil
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	logger.DebugEvent().Bool("force", force).Msg("Initializing configuration")

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configFile := fmt.Sprintf("%s/.ainative-code.yaml", home)

	// Check if file exists
	if _, err := os.Stat(configFile); err == nil && !force {
		return fmt.Errorf("config file already exists: %s (use --force to overwrite)", configFile)
	}

	// Set default values
	viper.Set("provider", "openai")
	viper.Set("model", "gpt-4")
	viper.Set("verbose", false)
	viper.Set("database.path", fmt.Sprintf("%s/.ainative-code/data.db", home))
	viper.Set("session.auto_save", true)

	// Write config file
	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Configuration file created: %s\n", configFile)
	fmt.Println("\nDefault settings:")
	fmt.Println("  provider: openai")
	fmt.Println("  model: gpt-4")
	fmt.Println("  verbose: false")

	return nil
}

// validateConfigValue validates a configuration key-value pair
func validateConfigValue(key, value string) error {
	// Maximum length for any config value
	const maxValueLength = 1000

	// Check value length
	if len(value) > maxValueLength {
		return fmt.Errorf("value exceeds maximum length of %d characters", maxValueLength)
	}

	// Validate specific keys
	switch key {
	case "provider":
		validProviders := []string{"openai", "anthropic", "ollama", "azure", "google", "bedrock"}
		if !isValidEnumValue(value, validProviders) {
			return fmt.Errorf("invalid provider '%s'. Valid providers: %v", value, validProviders)
		}

	case "model":
		// Basic model name validation - no empty, reasonable length
		if value == "" {
			return fmt.Errorf("model name cannot be empty")
		}
		if len(value) > 100 {
			return fmt.Errorf("model name exceeds maximum length of 100 characters")
		}

	case "temperature":
		// Temperature should be a number between 0 and 2
		// Note: This is a string at this point, but we can validate format
		if value == "" {
			return fmt.Errorf("temperature cannot be empty")
		}

	case "max_tokens":
		// Max tokens should be a positive number
		if value == "" {
			return fmt.Errorf("max_tokens cannot be empty")
		}

	case "api_key":
		// API keys should not be empty and have minimum length
		if value == "" {
			return fmt.Errorf("api_key cannot be empty")
		}
		if len(value) < 10 {
			return fmt.Errorf("api_key appears to be too short (minimum 10 characters)")
		}

	case "endpoint", "base_url":
		// URLs should not be empty
		if value == "" {
			return fmt.Errorf("%s cannot be empty", key)
		}

	case "verbose":
		// Boolean values
		validBools := []string{"true", "false", "1", "0", "yes", "no"}
		if !isValidEnumValue(value, validBools) {
			return fmt.Errorf("invalid boolean value '%s'. Valid values: %v", value, validBools)
		}
	}

	return nil
}

// isValidEnumValue checks if a value is in a list of valid values
func isValidEnumValue(value string, validValues []string) bool {
	for _, valid := range validValues {
		if value == valid {
			return true
		}
	}
	return false
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	logger.Debug("Validating configuration")

	fmt.Println("Validating configuration...")

	// Check required settings
	requiredSettings := []string{"provider"}
	missingSettings := []string{}

	for _, setting := range requiredSettings {
		if !viper.IsSet(setting) {
			missingSettings = append(missingSettings, setting)
		}
	}

	if len(missingSettings) > 0 {
		fmt.Println("\nValidation failed!")
		fmt.Println("Missing required settings:")
		for _, setting := range missingSettings {
			fmt.Printf("  - %s\n", setting)
		}
		return fmt.Errorf("configuration validation failed")
	}

	// Validate provider value
	provider := viper.GetString("provider")
	validProviders := []string{"openai", "anthropic", "ollama"}
	isValidProvider := false
	for _, vp := range validProviders {
		if provider == vp {
			isValidProvider = true
			break
		}
	}

	if !isValidProvider {
		return fmt.Errorf("invalid provider '%s'. Valid providers: %v", provider, validProviders)
	}

	fmt.Println("\nConfiguration is valid!")
	fmt.Printf("Provider: %s\n", provider)
	if viper.IsSet("model") {
		fmt.Printf("Model: %s\n", viper.GetString("model"))
	}

	return nil
}
