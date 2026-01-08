package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AINative-studio/ainative-code/internal/database"
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
