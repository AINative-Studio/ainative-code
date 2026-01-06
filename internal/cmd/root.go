package cmd

import (
	"os"

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
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			logger.ErrorEvent().Err(err).Msg("Failed to get home directory")
			os.Exit(1)
		}

		// Search config in home directory with name ".ainative-code" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ainative-code")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			logger.DebugEvent().Str("file", viper.ConfigFileUsed()).Msg("Using config file")
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
