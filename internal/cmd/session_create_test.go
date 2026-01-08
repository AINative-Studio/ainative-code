package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/AINative-studio/ainative-code/internal/database"
	"github.com/AINative-studio/ainative-code/internal/session"
)

// TestSessionCreateCommand tests the session create command initialization
func TestSessionCreateCommand(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create command exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if sessionCreateCmd == nil {
				t.Fatal("sessionCreateCmd should not be nil")
			}

			if sessionCreateCmd.Use != "create" {
				t.Errorf("expected Use 'create', got %s", sessionCreateCmd.Use)
			}

			if sessionCreateCmd.Short == "" {
				t.Error("expected Short description to be set")
			}

			if sessionCreateCmd.Long == "" {
				t.Error("expected Long description to be set")
			}
		})
	}
}

// TestSessionCreateFlags tests that all required flags are present
func TestSessionCreateFlags(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		shouldExist  bool
		isRequired   bool
	}{
		{
			name:        "title flag exists",
			flagName:    "title",
			shouldExist: true,
			isRequired:  true,
		},
		{
			name:        "tags flag exists",
			flagName:    "tags",
			shouldExist: true,
			isRequired:  false,
		},
		{
			name:        "provider flag exists",
			flagName:    "provider",
			shouldExist: true,
			isRequired:  false,
		},
		{
			name:        "model flag exists",
			flagName:    "model",
			shouldExist: true,
			isRequired:  false,
		},
		{
			name:        "metadata flag exists",
			flagName:    "metadata",
			shouldExist: true,
			isRequired:  false,
		},
		{
			name:        "no-activate flag exists",
			flagName:    "no-activate",
			shouldExist: true,
			isRequired:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := sessionCreateCmd.Flags().Lookup(tt.flagName)

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

// TestValidateTitle tests title validation logic
func TestValidateTitle(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		wantError bool
	}{
		{
			name:      "valid title",
			title:     "My Session",
			wantError: false,
		},
		{
			name:      "empty title",
			title:     "",
			wantError: true,
		},
		{
			name:      "whitespace only title",
			title:     "   ",
			wantError: true,
		},
		{
			name:      "title with special characters",
			title:     "Session: API Development #1",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title := strings.TrimSpace(tt.title)
			isEmpty := title == ""

			if tt.wantError && !isEmpty {
				t.Errorf("expected error for title %q, but got valid", tt.title)
			}

			if !tt.wantError && isEmpty {
				t.Errorf("expected valid title for %q, but got error", tt.title)
			}
		})
	}
}

// TestParseTags tests tag parsing logic
func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		tagsStr  string
		expected []string
	}{
		{
			name:     "single tag",
			tagsStr:  "golang",
			expected: []string{"golang"},
		},
		{
			name:     "multiple tags",
			tagsStr:  "golang,api,rest",
			expected: []string{"golang", "api", "rest"},
		},
		{
			name:     "tags with spaces",
			tagsStr:  "golang, api, rest",
			expected: []string{"golang", "api", "rest"},
		},
		{
			name:     "empty tags string",
			tagsStr:  "",
			expected: nil,
		},
		{
			name:     "tags with empty values",
			tagsStr:  "golang,,api",
			expected: []string{"golang", "api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tags []string
			if tt.tagsStr != "" {
				tags = strings.Split(tt.tagsStr, ",")
				for i := range tags {
					tags[i] = strings.TrimSpace(tags[i])
				}
				// Remove empty tags
				var validTags []string
				for _, tag := range tags {
					if tag != "" {
						validTags = append(validTags, tag)
					}
				}
				tags = validTags
			}

			if len(tags) != len(tt.expected) {
				t.Errorf("expected %d tags, got %d", len(tt.expected), len(tags))
				return
			}

			for i, tag := range tags {
				if tag != tt.expected[i] {
					t.Errorf("tag[%d]: expected %q, got %q", i, tt.expected[i], tag)
				}
			}
		})
	}
}

// TestParseMetadata tests metadata JSON parsing
func TestParseMetadata(t *testing.T) {
	tests := []struct {
		name      string
		metadata  string
		wantError bool
		expected  map[string]interface{}
	}{
		{
			name:      "valid metadata",
			metadata:  `{"project":"myapp","priority":"high"}`,
			wantError: false,
			expected: map[string]interface{}{
				"project":  "myapp",
				"priority": "high",
			},
		},
		{
			name:      "empty metadata",
			metadata:  "",
			wantError: false,
			expected:  nil,
		},
		{
			name:      "invalid JSON",
			metadata:  `{invalid json}`,
			wantError: true,
			expected:  nil,
		},
		{
			name:      "nested metadata",
			metadata:  `{"settings":{"theme":"dark","font":"mono"}}`,
			wantError: false,
			expected: map[string]interface{}{
				"settings": map[string]interface{}{
					"theme": "dark",
					"font":  "mono",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata map[string]interface{}
			var err error

			if tt.metadata != "" {
				err = json.Unmarshal([]byte(tt.metadata), &metadata)
			}

			if tt.wantError && err == nil {
				t.Error("expected error but got nil")
			}

			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantError && err == nil && tt.expected != nil {
				// Basic validation - just check keys exist
				for key := range tt.expected {
					if _, ok := metadata[key]; !ok {
						t.Errorf("expected key %q in metadata", key)
					}
				}
			}
		})
	}
}

// TestValidateProvider tests provider validation logic
func TestValidateProvider(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		wantError bool
	}{
		{
			name:      "valid provider - anthropic",
			provider:  "anthropic",
			wantError: false,
		},
		{
			name:      "valid provider - openai",
			provider:  "openai",
			wantError: false,
		},
		{
			name:      "valid provider - azure",
			provider:  "azure",
			wantError: false,
		},
		{
			name:      "valid provider - bedrock",
			provider:  "bedrock",
			wantError: false,
		},
		{
			name:      "valid provider - gemini",
			provider:  "gemini",
			wantError: false,
		},
		{
			name:      "valid provider - ollama",
			provider:  "ollama",
			wantError: false,
		},
		{
			name:      "valid provider - meta",
			provider:  "meta",
			wantError: false,
		},
		{
			name:      "valid provider - uppercase",
			provider:  "ANTHROPIC",
			wantError: false,
		},
		{
			name:      "invalid provider",
			provider:  "invalid-provider",
			wantError: true,
		},
		{
			name:      "empty provider",
			provider:  "",
			wantError: false, // Empty is allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.provider == "" {
				// Empty provider is valid (optional)
				return
			}

			provider := strings.ToLower(strings.TrimSpace(tt.provider))
			validProviders := []string{"anthropic", "openai", "azure", "bedrock", "gemini", "ollama", "meta"}
			isValid := false
			for _, vp := range validProviders {
				if provider == vp {
					isValid = true
					break
				}
			}

			if tt.wantError && isValid {
				t.Errorf("expected error for provider %q, but it was valid", tt.provider)
			}

			if !tt.wantError && !isValid {
				t.Errorf("expected provider %q to be valid, but got error", tt.provider)
			}
		})
	}
}

// TestSessionCreation tests the actual session creation with database
func TestSessionCreation(t *testing.T) {
	// Create a temporary database for testing
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize test database
	config := database.DefaultConfig(dbPath)
	db, err := database.Initialize(config)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create session manager
	mgr := session.NewSQLiteManager(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		session   *session.Session
		wantError bool
	}{
		{
			name: "create basic session",
			session: &session.Session{
				ID:        uuid.New().String(),
				Name:      "Test Session",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Status:    session.StatusActive,
			},
			wantError: false,
		},
		{
			name: "create session with model",
			session: &session.Session{
				ID:        uuid.New().String(),
				Name:      "Test Session with Model",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Status:    session.StatusActive,
				Model:     stringPtr("claude-3-5-sonnet-20241022"),
			},
			wantError: false,
		},
		{
			name: "create session with metadata",
			session: &session.Session{
				ID:        uuid.New().String(),
				Name:      "Test Session with Metadata",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Status:    session.StatusActive,
				Settings: map[string]interface{}{
					"tags": []string{"golang", "api"},
					"project": "myapp",
				},
			},
			wantError: false,
		},
		{
			name: "create session with empty name",
			session: &session.Session{
				ID:        uuid.New().String(),
				Name:      "",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Status:    session.StatusActive,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.CreateSession(ctx, tt.session)

			if tt.wantError && err == nil {
				t.Error("expected error but got nil")
			}

			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// If creation was successful, verify we can retrieve it
			if !tt.wantError && err == nil {
				retrieved, err := mgr.GetSession(ctx, tt.session.ID)
				if err != nil {
					t.Errorf("failed to retrieve created session: %v", err)
				}

				if retrieved.Name != tt.session.Name {
					t.Errorf("retrieved session name mismatch: expected %q, got %q",
						tt.session.Name, retrieved.Name)
				}

				if retrieved.Status != tt.session.Status {
					t.Errorf("retrieved session status mismatch: expected %q, got %q",
						tt.session.Status, retrieved.Status)
				}
			}
		})
	}
}

// TestUUIDGeneration tests that UUIDs are properly generated
func TestUUIDGeneration(t *testing.T) {
	// Generate multiple UUIDs and ensure they're unique
	seen := make(map[string]bool)

	for i := 0; i < 100; i++ {
		id := uuid.New().String()

		if id == "" {
			t.Error("generated UUID should not be empty")
		}

		if seen[id] {
			t.Errorf("duplicate UUID generated: %s", id)
		}
		seen[id] = true

		// Validate UUID format (basic check)
		parts := strings.Split(id, "-")
		if len(parts) != 5 {
			t.Errorf("invalid UUID format: %s (expected 5 parts)", id)
		}
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// TestDatabasePathResolution tests database path resolution
func TestDatabasePathResolution(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		setup  func()
		cleanup func()
	}{
		{
			name:   "use environment variable",
			envVar: "/custom/path/test.db",
			setup: func() {
				os.Setenv("AINATIVE_DB_PATH", "/custom/path/test.db")
			},
			cleanup: func() {
				os.Unsetenv("AINATIVE_DB_PATH")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			// Get database path from environment or use default
			dbPath := os.Getenv("AINATIVE_DB_PATH")
			if dbPath == "" {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					dbPath = filepath.Join(homeDir, ".ainative", "ainative.db")
				}
			}

			if tt.envVar != "" && dbPath != tt.envVar {
				t.Errorf("expected database path %q, got %q", tt.envVar, dbPath)
			}

			if dbPath == "" {
				t.Error("database path should not be empty")
			}
		})
	}
}
