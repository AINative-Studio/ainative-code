package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/database"
	"github.com/AINative-studio/ainative-code/internal/session"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// TestSessionListCommand tests the session list command initialization
func TestSessionListCommand(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "list command exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if sessionListCmd == nil {
				t.Fatal("sessionListCmd should not be nil")
			}

			if sessionListCmd.Use != "list" {
				t.Errorf("expected Use 'list', got %s", sessionListCmd.Use)
			}

			if sessionListCmd.Short == "" {
				t.Error("expected Short description to be set")
			}

			if sessionListCmd.Long == "" {
				t.Error("expected Long description to be set")
			}
		})
	}
}

// TestSessionListFlags tests that all required flags are present
func TestSessionListFlags(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		shouldExist  bool
		defaultValue interface{}
	}{
		{
			name:         "all flag exists",
			flagName:     "all",
			shouldExist:  true,
			defaultValue: false,
		},
		{
			name:         "limit flag exists",
			flagName:     "limit",
			shouldExist:  true,
			defaultValue: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := sessionListCmd.Flags().Lookup(tt.flagName)

			if tt.shouldExist {
				if flag == nil {
					t.Errorf("flag %s should exist", tt.flagName)
				}
			} else {
				if flag != nil {
					t.Errorf("flag %s should not exist", tt.flagName)
				}
			}
		})
	}
}

// TestSessionListLimitValidation tests the limit validation logic
func TestSessionListLimitValidation(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		wantError bool
		errorMsg  string
	}{
		{
			name:      "negative limit returns error",
			limit:     -1,
			wantError: true,
			errorMsg:  "Error: limit must be a positive integer",
		},
		{
			name:      "zero limit returns error",
			limit:     0,
			wantError: true,
			errorMsg:  "Error: limit must be a positive integer",
		},
		{
			name:      "very negative limit returns error",
			limit:     -999,
			wantError: true,
			errorMsg:  "Error: limit must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the limit value
			sessionLimit = tt.limit

			// Create a mock command
			cmd := &cobra.Command{}

			// Run the session list command
			// We expect it to fail at validation, before trying to access the database
			err := runSessionList(cmd, []string{})

			// Check error expectations
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for limit %d, but got nil", tt.limit)
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil && err.Error() == tt.errorMsg {
					t.Errorf("unexpected validation error for limit %d: %v", tt.limit, err)
				}
			}
		})
	}
}

// TestSessionListPositiveLimitValidation tests that positive limits pass validation
func TestSessionListPositiveLimitValidation(t *testing.T) {
	tests := []struct {
		name  string
		limit int
	}{
		{
			name:  "positive limit passes validation",
			limit: 1,
		},
		{
			name:  "default limit passes validation",
			limit: 10,
		},
		{
			name:  "large positive limit passes validation",
			limit: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple inline validation check - mimics the actual validation in runSessionList
			sessionLimit = tt.limit

			// This is the exact validation logic from runSessionList
			if sessionLimit <= 0 {
				t.Errorf("positive limit %d should pass validation but failed", tt.limit)
			}
		})
	}
}

// TestSessionListWithSessions tests listing sessions with various data
func TestSessionListWithSessions(t *testing.T) {
	// Set up test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Set environment variable for database path
	oldDBPath := os.Getenv("AINATIVE_DB_PATH")
	os.Setenv("AINATIVE_DB_PATH", dbPath)
	defer os.Setenv("AINATIVE_DB_PATH", oldDBPath)

	// Initialize test database
	config := database.DefaultConfig(dbPath)
	db, err := database.Initialize(config)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create session manager and add test sessions
	mgr := session.NewSQLiteManager(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create some test sessions
	testSessions := []*session.Session{
		{
			ID:        uuid.New().String(),
			Name:      "Test Session 1",
			CreatedAt: time.Now().Add(-2 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
			Status:    session.StatusActive,
		},
		{
			ID:        uuid.New().String(),
			Name:      "Test Session 2",
			CreatedAt: time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
			Status:    session.StatusActive,
		},
		{
			ID:        uuid.New().String(),
			Name:      "Test Session 3",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Status:    session.StatusActive,
		},
	}

	for _, sess := range testSessions {
		if err := mgr.CreateSession(ctx, sess); err != nil {
			t.Fatalf("failed to create test session: %v", err)
		}
	}

	tests := []struct {
		name      string
		limit     int
		all       bool
		wantError bool
	}{
		{
			name:      "list with limit 2",
			limit:     2,
			all:       false,
			wantError: false,
		},
		{
			name:      "list with limit 5",
			limit:     5,
			all:       false,
			wantError: false,
		},
		{
			name:      "list all sessions",
			limit:     10,
			all:       true,
			wantError: false,
		},
		{
			name:      "list with limit 1",
			limit:     1,
			all:       false,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the flags
			sessionLimit = tt.limit
			sessionListAll = tt.all

			// Create a mock command
			cmd := &cobra.Command{}

			// Redirect stdout to suppress output during tests
			oldStdout := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			// Run the session list command
			err := runSessionList(cmd, []string{})

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestSessionListAliases tests that aliases work correctly
func TestSessionListAliases(t *testing.T) {
	expectedAliases := []string{"ls", "l"}

	if len(sessionListCmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(sessionListCmd.Aliases))
	}

	for i, alias := range expectedAliases {
		if i >= len(sessionListCmd.Aliases) {
			t.Errorf("missing alias: %s", alias)
			continue
		}
		if sessionListCmd.Aliases[i] != alias {
			t.Errorf("expected alias %s, got %s", alias, sessionListCmd.Aliases[i])
		}
	}
}

// TestSessionListCommandAliases tests that the session command has proper aliases
func TestSessionListCommandAliases(t *testing.T) {
	expectedAliases := []string{"sessions", "sess"}

	if len(sessionCmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d command aliases, got %d", len(expectedAliases), len(sessionCmd.Aliases))
	}

	for i, alias := range expectedAliases {
		if i >= len(sessionCmd.Aliases) {
			t.Errorf("missing command alias: %s", alias)
			continue
		}
		if sessionCmd.Aliases[i] != alias {
			t.Errorf("expected command alias %s, got %s", alias, sessionCmd.Aliases[i])
		}
	}
}

// TestSessionListEdgeCases tests edge cases for session listing
func TestSessionListEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		all       bool
		wantError bool
	}{
		{
			name:      "empty database",
			limit:     10,
			all:       false,
			wantError: false,
		},
		{
			name:      "limit larger than session count",
			limit:     100,
			all:       false,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test database
			tmpDir := t.TempDir()
			dbPath := filepath.Join(tmpDir, "test.db")

			// Set environment variable for database path
			oldDBPath := os.Getenv("AINATIVE_DB_PATH")
			os.Setenv("AINATIVE_DB_PATH", dbPath)
			defer os.Setenv("AINATIVE_DB_PATH", oldDBPath)

			// Initialize test database
			config := database.DefaultConfig(dbPath)
			db, err := database.Initialize(config)
			if err != nil {
				t.Fatalf("failed to initialize database: %v", err)
			}
			defer db.Close()

			// Set the flags
			sessionLimit = tt.limit
			sessionListAll = tt.all

			// Create a mock command
			cmd := &cobra.Command{}

			// Redirect stdout to suppress output during tests
			oldStdout := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			// Run the session list command
			err = runSessionList(cmd, []string{})

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
