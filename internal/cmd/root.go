package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

var (
	cfgFile    string
	provider   string
	model      string
	verbose    bool
	skipSetup  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ainative-code",
	Short: "AINative Code - AI-powered development tool",
	Long: `AINative Code is a comprehensive AI-powered development tool that provides:
- Interactive AI chat mode for code assistance
- Session management for conversation history
- ZeroDB integration for secure data storage
- Design token management for UI consistency
- Strapi CMS integration for content management
- RLHF feedback collection for model improvement`,
	Version: "0.1.0",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Configure logger based on verbose flag
		if verbose {
			logger.SetLevel("debug")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ainative-code.yaml)")
	rootCmd.PersistentFlags().StringVar(&provider, "provider", "", "AI provider (openai, anthropic, ollama)")
	rootCmd.PersistentFlags().StringVar(&model, "model", "", "AI model to use")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&skipSetup, "skip-setup", false, "skip first-time setup check")

	// Bind flags to viper
	viper.BindPFlag("provider", rootCmd.PersistentFlags().Lookup("provider"))
	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Bind environment variables
	bindEnvironmentVariables()
}

// bindEnvironmentVariables binds all supported environment variables
func bindEnvironmentVariables() {
	// Basic configuration
	viper.BindEnv("provider")
	viper.BindEnv("model")
	viper.BindEnv("verbose")
	viper.BindEnv("api_key")

	// Note: Viper will automatically look for AINATIVE_CODE_PROVIDER,
	// AINATIVE_CODE_MODEL, etc. after SetEnvPrefix is called in initConfig
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Set environment variable prefix for AINATIVE_CODE_*
	viper.SetEnvPrefix("AINATIVE_CODE")

	// Replace dots and dashes with underscores for environment variables
	// e.g., AINATIVE_CODE_API_KEY for "api_key" config
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Enable automatic environment variable binding
	viper.AutomaticEnv()

	if cfgFile != "" {
		// Use config file from the flag - validate it exists
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			logger.ErrorEvent().Str("file", cfgFile).Msg("Config file not found")
			fmt.Fprintf(os.Stderr, "Error: config file not found: %s\n", cfgFile)
			os.Exit(1)
		}
		viper.SetConfigFile(cfgFile)
	} else {
		// Config priority (like Crush):
		// 1. ./.ainative-code.yaml (project-local, hidden)
		// 2. ./ainative-code.yaml (project-local, visible)
		// 3. ~/.ainative-code.yaml (global)

		home, err := os.UserHomeDir()
		if err != nil {
			logger.ErrorEvent().Err(err).Msg("Failed to get home directory")
			os.Exit(1)
		}

		// Try files in priority order
		configFiles := []string{
			"./.ainative-code.yaml",
			"./ainative-code.yaml",
			filepath.Join(home, ".ainative-code.yaml"),
		}

		for _, configPath := range configFiles {
			if _, err := os.Stat(configPath); err == nil {
				viper.SetConfigFile(configPath)
				break
			}
		}
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			logger.DebugEvent().Str("file", viper.ConfigFileUsed()).Msg("Using config file")
		}
	} else {
		// Handle different error types appropriately
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			// Config file not found is acceptable - we'll use defaults
			if verbose {
				logger.DebugEvent().Msg("No config file found, using defaults")
			}
		case *os.PathError:
			// Permission or file access errors
			logger.WarnEvent().
				Err(err).
				Str("file", cfgFile).
				Msg("Cannot read config file due to permission or access error")
			fmt.Fprintf(os.Stderr, "Warning: Cannot read config file: %v\n", err)
		default:
			// YAML parse errors or other config errors
			logger.WarnEvent().
				Err(err).
				Str("file", viper.ConfigFileUsed()).
				Msg("Error parsing config file")
			fmt.Fprintf(os.Stderr, "Warning: Error reading config file: %v\n", err)
			fmt.Fprintf(os.Stderr, "Using default configuration instead.\n")
		}
	}
}

// GetProvider returns the configured AI provider
func GetProvider() string {
	if provider != "" {
		return provider
	}
	return viper.GetString("provider")
}

// GetModel returns the configured AI model
func GetModel() string {
	if model != "" {
		return model
	}
	return viper.GetString("model")
}

// GetVerbose returns the verbose flag value
func GetVerbose() bool {
	return verbose || viper.GetBool("verbose")
}

// NewRootCmd returns a new root command instance for testing/benchmarking
func NewRootCmd() *cobra.Command {
	return rootCmd
}

// NewChatCmd returns the chat command for testing/benchmarking
func NewChatCmd() *cobra.Command {
	return chatCmd
}

// NewSessionCmd returns the session command for testing/benchmarking
func NewSessionCmd() *cobra.Command {
	return sessionCmd
}

// NewConfigCmd returns the config command for testing/benchmarking
func NewConfigCmd() *cobra.Command {
	return configCmd
}
