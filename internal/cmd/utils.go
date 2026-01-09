package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/AINative-studio/ainative-code/internal/auth/keychain"
	"github.com/AINative-studio/ainative-code/internal/database"
	"github.com/AINative-studio/ainative-code/internal/logger"
	llmprovider "github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/AINative-studio/ainative-code/internal/provider/anthropic"
	"github.com/AINative-studio/ainative-code/internal/provider/openai"
	"github.com/AINative-studio/ainative-code/internal/provider/meta"
)

// outputAsJSON outputs data as formatted JSON
func outputAsJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// getDatabase returns a database connection with default configuration
func getDatabase() (*database.DB, error) {
	dbPath := getDatabasePath()

	// Initialize database with default config
	config := database.DefaultConfig(dbPath)
	db, err := database.Initialize(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

// getDatabasePath returns the database file path
func getDatabasePath() string {
	// Get database path from environment or use default
	dbPath := os.Getenv("AINATIVE_DB_PATH")
	if dbPath == "" {
		// Use default path in user's home directory
		homeDir, _ := os.UserHomeDir()
		dbPath = filepath.Join(homeDir, ".ainative", "ainative.db")
	}
	return dbPath
}

// getAPIKey retrieves the API key for the specified provider
// It checks in this order:
// 1. Provider-specific environment variable (e.g., OPENAI_API_KEY)
// 2. Generic AINATIVE_CODE_API_KEY environment variable
// 3. Viper configuration (api_key)
// 4. System keychain
func getAPIKey(providerName string) (string, error) {
	// Check provider-specific environment variable
	providerEnvKey := ""
	switch providerName {
	case "openai":
		providerEnvKey = "OPENAI_API_KEY"
	case "anthropic":
		providerEnvKey = "ANTHROPIC_API_KEY"
	case "meta_llama", "meta":
		providerEnvKey = "META_LLAMA_API_KEY"
	case "ollama":
		// Ollama typically doesn't require an API key for local instances
		return "", nil
	}

	if providerEnvKey != "" {
		if key := os.Getenv(providerEnvKey); key != "" {
			logger.DebugEvent().
				Str("provider", providerName).
				Str("source", "provider_env").
				Msg("Using API key from provider-specific environment variable")
			return key, nil
		}
	}

	// Check generic environment variable
	if key := viper.GetString("api_key"); key != "" {
		logger.DebugEvent().
			Str("provider", providerName).
			Str("source", "viper_config").
			Msg("Using API key from configuration")
		return key, nil
	}

	// Check keychain
	kc := keychain.Get()
	if apiKey, err := kc.GetAPIKey(); err == nil && apiKey != "" {
		logger.DebugEvent().
			Str("provider", providerName).
			Str("source", "keychain").
			Msg("Using API key from keychain")
		return apiKey, nil
	}

	return "", fmt.Errorf("no API key found for provider %s. Set %s or AINATIVE_CODE_API_KEY environment variable, or run 'ainative-code setup'", providerName, providerEnvKey)
}

// initializeProvider creates and initializes an AI provider based on the provider name
func initializeProvider(ctx context.Context, providerName, modelName string) (llmprovider.Provider, error) {
	logger.DebugEvent().
		Str("provider", providerName).
		Str("model", modelName).
		Msg("Initializing AI provider")

	// Get API key for the provider
	apiKey, err := getAPIKey(providerName)
	if err != nil {
		return nil, err
	}

	// Initialize the appropriate provider
	switch providerName {
	case "openai":
		return openai.NewOpenAIProvider(openai.Config{
			APIKey: apiKey,
			Logger: nil, // Use default logger
		})

	case "anthropic":
		return anthropic.NewAnthropicProvider(anthropic.Config{
			APIKey: apiKey,
			Logger: nil, // Use default logger
		})

	case "meta_llama", "meta":
		return meta.NewMetaProvider(&meta.Config{
			APIKey: apiKey,
		})

	default:
		return nil, fmt.Errorf("unsupported provider: %s. Supported providers: openai, anthropic, meta_llama", providerName)
	}
}
