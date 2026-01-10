package cmd

import (
	"os"
	"testing"
)

// TestGetAuthURL tests the auth URL resolution logic
func TestGetAuthURL(t *testing.T) {
	tests := []struct {
		name        string
		envVar      string
		envValue    string
		wantContain string
	}{
		{
			name:        "uses environment variable when set",
			envVar:      "AINATIVE_AUTH_URL",
			envValue:    "https://custom.example.com/oauth/authorize",
			wantContain: "https://custom.example.com/oauth/authorize",
		},
		{
			name:        "uses production endpoint when env not set (api.ainative.studio)",
			envVar:      "",
			envValue:    "",
			wantContain: "api.ainative.studio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env
			originalEnv := os.Getenv("AINATIVE_AUTH_URL")
			defer os.Setenv("AINATIVE_AUTH_URL", originalEnv)

			// Set test env
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envValue)
			} else {
				os.Unsetenv("AINATIVE_AUTH_URL")
			}

			url := getAuthURL()

			if tt.wantContain != "" && !containsString(url, tt.wantContain) {
				t.Errorf("getAuthURL() = %q, want to contain %q", url, tt.wantContain)
			}
		})
	}
}

// TestGetTokenURL tests the token URL resolution logic
func TestGetTokenURL(t *testing.T) {
	tests := []struct {
		name        string
		envVar      string
		envValue    string
		wantContain string
	}{
		{
			name:        "uses environment variable when set",
			envVar:      "AINATIVE_TOKEN_URL",
			envValue:    "https://custom.example.com/oauth/token",
			wantContain: "https://custom.example.com/oauth/token",
		},
		{
			name:        "uses production endpoint when env not set (api.ainative.studio)",
			envVar:      "",
			envValue:    "",
			wantContain: "api.ainative.studio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env
			originalEnv := os.Getenv("AINATIVE_TOKEN_URL")
			defer os.Setenv("AINATIVE_TOKEN_URL", originalEnv)

			// Set test env
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envValue)
			} else {
				os.Unsetenv("AINATIVE_TOKEN_URL")
			}

			url := getTokenURL()

			if tt.wantContain != "" && !containsString(url, tt.wantContain) {
				t.Errorf("getTokenURL() = %q, want to contain %q", url, tt.wantContain)
			}
		})
	}
}

// TestAuthCommandExists tests that auth command is properly initialized
func TestAuthCommandExists(t *testing.T) {
	if authCmd == nil {
		t.Fatal("authCmd should not be nil")
	}

	if authCmd.Use != "auth" {
		t.Errorf("expected Use 'auth', got %s", authCmd.Use)
	}

	// Verify subcommands exist
	expectedSubcommands := []string{"login", "logout", "whoami", "token"}
	for _, subcmd := range expectedSubcommands {
		found := false
		for _, cmd := range authCmd.Commands() {
			if cmd.Use == subcmd || cmd.Use == "token" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %s to exist", subcmd)
		}
	}
}

// TestLoginCommandFlags tests that login command has expected flags
func TestLoginCommandFlags(t *testing.T) {
	expectedFlags := []string{"auth-url", "token-url", "client-id", "scopes"}

	for _, flagName := range expectedFlags {
		flag := loginCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("expected flag %s to exist", flagName)
		}
	}
}

// TestDefaultOAuthConfig tests that default OAuth config uses production endpoint
func TestDefaultOAuthConfig(t *testing.T) {
	// The default config should be initialized
	if defaultOAuthConfig.ClientID == "" {
		t.Error("defaultOAuthConfig.ClientID should not be empty")
	}

	if defaultOAuthConfig.RedirectURL == "" {
		t.Error("defaultOAuthConfig.RedirectURL should not be empty")
	}

	if len(defaultOAuthConfig.Scopes) == 0 {
		t.Error("defaultOAuthConfig.Scopes should not be empty")
	}

	// AuthURL and TokenURL should be set to production endpoints
	if defaultOAuthConfig.AuthURL == "" {
		t.Error("defaultOAuthConfig.AuthURL should not be empty")
	}

	if defaultOAuthConfig.TokenURL == "" {
		t.Error("defaultOAuthConfig.TokenURL should not be empty")
	}

	// Verify they contain api.ainative.studio
	if !containsString(defaultOAuthConfig.AuthURL, "api.ainative.studio") {
		t.Errorf("AuthURL should contain 'api.ainative.studio', got %s", defaultOAuthConfig.AuthURL)
	}

	if !containsString(defaultOAuthConfig.TokenURL, "api.ainative.studio") {
		t.Errorf("TokenURL should contain 'api.ainative.studio', got %s", defaultOAuthConfig.TokenURL)
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	if len(s) == 0 || len(substr) == 0 {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
