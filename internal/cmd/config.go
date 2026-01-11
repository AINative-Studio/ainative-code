package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	Long: `Display all current configuration values.

By default, sensitive values (API keys, tokens, passwords, secrets) are masked
for security. Use the --show-secrets flag to display full values when needed.

Examples:
  # Show configuration with masked secrets
  ainative-code config show

  # Show configuration with full secrets (use with caution)
  ainative-code config show --show-secrets`,
	Aliases: []string{"list", "ls"},
	RunE:    runConfigShow,
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

	// Config show flags
	configShowCmd.Flags().BoolP("show-secrets", "s", false, "show sensitive values (API keys, tokens, passwords) in plain text")

	// Config init flags
	configInitCmd.Flags().BoolP("force", "f", false, "overwrite existing config file")
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	logger.Debug("Showing configuration")

	// Check if user wants to show secrets
	showSecrets, _ := cmd.Flags().GetBool("show-secrets")

	fmt.Println("Current Configuration:")
	fmt.Println("======================")

	allSettings := viper.AllSettings()
	if len(allSettings) == 0 {
		fmt.Println("No configuration values set")
		return nil
	}

	// Mask sensitive data unless --show-secrets flag is set
	displaySettings := allSettings
	if !showSecrets {
		displaySettings = maskSensitiveData(allSettings).(map[string]interface{})
		fmt.Println("(Sensitive values are masked. Use --show-secrets to display full values)")
		fmt.Println()
	}

	// Format and display the configuration
	output := formatConfigOutput(displaySettings, 0)
	fmt.Print(output)

	if viper.ConfigFileUsed() != "" {
		fmt.Printf("\nConfig file: %s\n", viper.ConfigFileUsed())
	} else {
		fmt.Println("\nNo config file in use")
	}

	// Show security warning if secrets are displayed
	if showSecrets {
		fmt.Println("\nWARNING: Sensitive values are displayed in plain text!")
		fmt.Println("Ensure this output is not shared or logged in insecure locations.")
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

	// Validate the key name before proceeding
	if err := validateConfigKey(key); err != nil {
		return fmt.Errorf("invalid configuration key: %w", err)
	}

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

	// Check if key exists in configuration
	// viper.IsSet returns false for keys with empty string values (Issue #101)
	// So we need to check both IsSet and if the value is explicitly an empty string
	value := viper.Get(key)

	// If viper.Get returns nil, the key truly doesn't exist
	if !viper.IsSet(key) && value == nil {
		return fmt.Errorf("configuration key '%s' not found", key)
	}

	// Handle empty string values explicitly
	if strValue, ok := value.(string); ok && strValue == "" {
		fmt.Printf("%s: (empty)\n", key)
		return nil
	}

	// For nil values from empty config entries
	if value == nil {
		fmt.Printf("%s: (not set)\n", key)
		return nil
	}

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

// validateConfigKey validates a configuration key name
func validateConfigKey(key string) error {
	// Check for empty key
	if key == "" {
		return fmt.Errorf("key name cannot be empty")
	}

	// Check for whitespace-only key
	if len(strings.TrimSpace(key)) == 0 {
		return fmt.Errorf("key name cannot be whitespace only")
	}

	// Maximum length for key name
	const maxKeyLength = 100

	// Check key length
	if len(key) > maxKeyLength {
		return fmt.Errorf("key name exceeds maximum length of %d characters", maxKeyLength)
	}

	// Check if key contains only valid characters (alphanumeric, dots, underscores, hyphens)
	// This prevents issues with config file parsing
	for i, ch := range key {
		if !((ch >= 'a' && ch <= 'z') ||
		     (ch >= 'A' && ch <= 'Z') ||
		     (ch >= '0' && ch <= '9') ||
		     ch == '.' || ch == '_' || ch == '-') {
			return fmt.Errorf("key name contains invalid character '%c' at position %d. Valid characters: a-z, A-Z, 0-9, dot (.), underscore (_), hyphen (-)", ch, i)
		}
	}

	// Key must not start or end with a dot (prevents config parsing issues)
	if strings.HasPrefix(key, ".") {
		return fmt.Errorf("key name cannot start with a dot")
	}
	if strings.HasSuffix(key, ".") {
		return fmt.Errorf("key name cannot end with a dot")
	}

	// Key must not contain consecutive dots (prevents config parsing issues)
	if strings.Contains(key, "..") {
		return fmt.Errorf("key name cannot contain consecutive dots")
	}

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

	// Check for provider in the correct location
	// Priority: llm.default_provider (new structure) > provider (legacy)
	var provider string
	var providerLocation string

	if viper.IsSet("llm.default_provider") {
		provider = viper.GetString("llm.default_provider")
		providerLocation = "llm.default_provider"
	} else if viper.IsSet("provider") {
		provider = viper.GetString("provider")
		providerLocation = "provider"
	}

	// Check if provider is set and not empty
	if provider == "" {
		fmt.Println("\nValidation failed!")
		fmt.Println("Missing required settings:")
		fmt.Println("  - Either 'llm.default_provider' or 'provider' must be set")
		fmt.Println("\nRecommended: Use 'llm.default_provider' in your config file")
		fmt.Println("Example:")
		fmt.Println("  llm:")
		fmt.Println("    default_provider: anthropic")
		return fmt.Errorf("configuration validation failed: no provider configured")
	}

	// Validate provider value
	validProviders := []string{"openai", "anthropic", "ollama", "google", "bedrock", "azure", "meta_llama", "meta"}
	isValidProvider := false
	for _, vp := range validProviders {
		if provider == vp {
			isValidProvider = true
			break
		}
	}

	if !isValidProvider {
		fmt.Println("\nValidation failed!")
		fmt.Printf("Invalid provider: '%s'\n", provider)
		fmt.Printf("Valid providers: %v\n", validProviders)
		return fmt.Errorf("configuration validation failed: invalid provider '%s'", provider)
	}

	// Success message
	fmt.Println("\nConfiguration is valid!")
	fmt.Printf("Provider: %s (from %s)\n", provider, providerLocation)

	// Show additional info if available
	if viper.IsSet("llm." + provider + ".model") {
		fmt.Printf("Model: %s\n", viper.GetString("llm."+provider+".model"))
	} else if viper.IsSet("model") {
		fmt.Printf("Model: %s\n", viper.GetString("model"))
	}

	// Warn if using legacy structure
	if providerLocation == "provider" {
		fmt.Println("\nNote: You are using the legacy 'provider' field.")
		fmt.Println("Consider migrating to the new structure:")
		fmt.Println("  llm:")
		fmt.Println("    default_provider: " + provider)
	}

	return nil
}
