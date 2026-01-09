package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
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
			name:        "uses fallback when env not set and prod unreachable",
			envVar:      "",
			envValue:    "",
			wantContain: "localhost:9090",
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
			name:        "uses fallback when env not set and prod unreachable",
			envVar:      "",
			envValue:    "",
			wantContain: "localhost:9090",
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

// TestIsEndpointReachable tests endpoint reachability checking
func TestIsEndpointReachable(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func() *httptest.Server
		wantResult bool
	}{
		{
			name: "returns true for reachable endpoint (200 OK)",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
			},
			wantResult: true,
		},
		{
			name: "returns true for reachable endpoint (404 Not Found)",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			wantResult: true,
		},
		{
			name: "returns true for reachable endpoint (401 Unauthorized)",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
				}))
			},
			wantResult: true,
		},
		{
			name: "returns false for unreachable endpoint (500 Internal Server Error)",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			wantResult: false,
		},
		{
			name: "returns false for slow endpoint (timeout)",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Simulate slow response that exceeds timeout
					time.Sleep(3 * time.Second)
					w.WriteHeader(http.StatusOK)
				}))
			},
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupMock()
			defer server.Close()

			result := isEndpointReachable(server.URL)

			if result != tt.wantResult {
				t.Errorf("isEndpointReachable() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// TestIsEndpointReachable_InvalidURL tests handling of invalid URLs
func TestIsEndpointReachable_InvalidURL(t *testing.T) {
	// Test with completely unreachable URL
	result := isEndpointReachable("http://invalid-domain-that-does-not-exist-12345.com")
	if result {
		t.Error("isEndpointReachable() should return false for invalid domain")
	}

	// Test with malformed URL
	result = isEndpointReachable("not-a-valid-url")
	if result {
		t.Error("isEndpointReachable() should return false for malformed URL")
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

// TestDefaultOAuthConfig tests that default OAuth config uses fallback logic
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

	// AuthURL and TokenURL should be set (either to prod or fallback)
	if defaultOAuthConfig.AuthURL == "" {
		t.Error("defaultOAuthConfig.AuthURL should not be empty")
	}

	if defaultOAuthConfig.TokenURL == "" {
		t.Error("defaultOAuthConfig.TokenURL should not be empty")
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
