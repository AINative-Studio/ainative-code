package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
)

// TestAuthLoginWithBackendClient tests auth login using backend.Client
func TestAuthLoginWithBackendClient(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		password       string
		serverResponse int
		serverBody     string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "successful login",
			email:          "test@example.com",
			password:       "password123",
			serverResponse: http.StatusOK,
			serverBody:     `{"access_token":"token123","refresh_token":"refresh123","token_type":"Bearer","user":{"id":"user1","email":"test@example.com"}}`,
			expectError:    false,
		},
		{
			name:           "invalid credentials",
			email:          "test@example.com",
			password:       "wrongpassword",
			serverResponse: http.StatusUnauthorized,
			serverBody:     `{"error":"invalid credentials"}`,
			expectError:    true,
			errorContains:  "unauthorized",
		},
		{
			name:           "server error",
			email:          "test@example.com",
			password:       "password123",
			serverResponse: http.StatusInternalServerError,
			serverBody:     `{"error":"internal server error"}`,
			expectError:    true,
			errorContains:  "server error",
		},
		{
			name:           "empty email",
			email:          "",
			password:       "password123",
			serverResponse: http.StatusBadRequest,
			serverBody:     `{"error":"email required"}`,
			expectError:    true,
		},
		{
			name:           "empty password",
			email:          "test@example.com",
			password:       "",
			serverResponse: http.StatusBadRequest,
			serverBody:     `{"error":"password required"}`,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/auth/login" {
					w.WriteHeader(tt.serverResponse)
					w.Write([]byte(tt.serverBody))
				}
			}))
			defer server.Close()

			// Reset viper
			viper.Reset()
			viper.Set("backend_url", server.URL)

			// Create login command
			cmd := newAuthLoginBackendCmd()
			cmd.SetArgs([]string{"--email", tt.email, "--password", tt.password})

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Execute
			err := cmd.Execute()

			// Verify
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
			if tt.errorContains != "" && err != nil {
				if !containsString(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got: %v", tt.errorContains, err)
				}
			}

			// If successful, verify config was saved
			if !tt.expectError {
				accessToken := viper.GetString("access_token")
				if accessToken == "" {
					t.Error("expected access_token to be saved")
				}
				userEmail := viper.GetString("user_email")
				if userEmail != tt.email {
					t.Errorf("expected user_email to be %s, got %s", tt.email, userEmail)
				}
			}
		})
	}
}

// TestAuthLoginSavesTokens tests that login saves tokens to config
func TestAuthLoginSavesTokens(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token":"token123","refresh_token":"refresh123","token_type":"Bearer","user":{"id":"user1","email":"test@example.com"}}`))
		}
	}))
	defer server.Close()

	// Reset viper
	viper.Reset()
	viper.Set("backend_url", server.URL)

	// Create and execute login command
	cmd := newAuthLoginBackendCmd()
	cmd.SetArgs([]string{"--email", "test@example.com", "--password", "password123"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Verify tokens are saved
	if viper.GetString("access_token") != "token123" {
		t.Errorf("expected access_token 'token123', got %s", viper.GetString("access_token"))
	}
	if viper.GetString("refresh_token") != "refresh123" {
		t.Errorf("expected refresh_token 'refresh123', got %s", viper.GetString("refresh_token"))
	}
	if viper.GetString("user_email") != "test@example.com" {
		t.Errorf("expected user_email 'test@example.com', got %s", viper.GetString("user_email"))
	}
}

// TestAuthLogoutWithBackendClient tests auth logout using backend.Client
func TestAuthLogoutWithBackendClient(t *testing.T) {
	tests := []struct {
		name           string
		hasToken       bool
		serverResponse int
		expectError    bool
	}{
		{
			name:           "successful logout with token",
			hasToken:       true,
			serverResponse: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "logout without token",
			hasToken:       false,
			serverResponse: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "server error during logout",
			hasToken:       true,
			serverResponse: http.StatusInternalServerError,
			expectError:    false, // Should still clear local tokens
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/auth/logout" {
					w.WriteHeader(tt.serverResponse)
				}
			}))
			defer server.Close()

			// Reset viper and set tokens if needed
			viper.Reset()
			viper.Set("backend_url", server.URL)
			if tt.hasToken {
				viper.Set("access_token", "token123")
				viper.Set("refresh_token", "refresh123")
				viper.Set("user_email", "test@example.com")
			}

			// Create logout command
			cmd := newAuthLogoutBackendCmd()

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Execute
			err := cmd.Execute()

			// Verify
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			// Verify tokens are cleared
			if viper.GetString("access_token") != "" {
				t.Error("expected access_token to be cleared")
			}
			if viper.GetString("refresh_token") != "" {
				t.Error("expected refresh_token to be cleared")
			}
			if viper.GetString("user_email") != "" {
				t.Error("expected user_email to be cleared")
			}
		})
	}
}

// TestAuthLogoutClearsTokens tests that logout clears all tokens
func TestAuthLogoutClearsTokens(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/logout" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Reset viper and set tokens
	viper.Reset()
	viper.Set("backend_url", server.URL)
	viper.Set("access_token", "token123")
	viper.Set("refresh_token", "refresh123")
	viper.Set("user_email", "test@example.com")

	// Create and execute logout command
	cmd := newAuthLogoutBackendCmd()
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// Verify all tokens are cleared
	if viper.GetString("access_token") != "" {
		t.Error("expected access_token to be cleared")
	}
	if viper.GetString("refresh_token") != "" {
		t.Error("expected refresh_token to be cleared")
	}
	if viper.GetString("user_email") != "" {
		t.Error("expected user_email to be cleared")
	}
}

// TestAuthRefreshToken tests token refresh functionality
func TestAuthRefreshToken(t *testing.T) {
	tests := []struct {
		name            string
		hasRefreshToken bool
		serverResponse  int
		serverBody      string
		expectError     bool
		errorContains   string
	}{
		{
			name:            "successful token refresh",
			hasRefreshToken: true,
			serverResponse:  http.StatusOK,
			serverBody:      `{"access_token":"new_token","refresh_token":"new_refresh","token_type":"Bearer"}`,
			expectError:     false,
		},
		{
			name:            "no refresh token",
			hasRefreshToken: false,
			serverResponse:  http.StatusUnauthorized,
			expectError:     true,
			errorContains:   "no refresh token found",
		},
		{
			name:            "invalid refresh token",
			hasRefreshToken: true,
			serverResponse:  http.StatusUnauthorized,
			serverBody:      `{"error":"invalid refresh token"}`,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/auth/refresh" {
					w.WriteHeader(tt.serverResponse)
					if tt.serverBody != "" {
						w.Write([]byte(tt.serverBody))
					}
				}
			}))
			defer server.Close()

			// Reset viper
			viper.Reset()
			viper.Set("backend_url", server.URL)
			if tt.hasRefreshToken {
				viper.Set("refresh_token", "refresh123")
			} else {
				viper.Set("refresh_token", "")
			}

			// Create refresh command
			cmd := newAuthRefreshBackendCmd()

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Execute
			err := cmd.Execute()

			// Verify
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
			if tt.errorContains != "" && err != nil {
				if !containsString(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got: %v", tt.errorContains, err)
				}
			}

			// If successful, verify new tokens are saved
			if !tt.expectError {
				if viper.GetString("access_token") != "new_token" {
					t.Error("expected new access_token to be saved")
				}
			}
		})
	}
}

// Helper functions are now implemented in auth.go
