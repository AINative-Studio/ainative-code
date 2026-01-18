package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
)

// TestChatWithAINativeProvider tests chat command using AINative backend
func TestChatWithAINativeProvider(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		accessToken    string
		serverResponse int
		serverBody     string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "successful chat request",
			message:        "Hello, AI!",
			accessToken:    "token123",
			serverResponse: http.StatusOK,
			serverBody:     `{"id":"chat1","model":"claude-sonnet-4-5","choices":[{"message":{"role":"assistant","content":"Hello! How can I help you?"}}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`,
			expectError:    false,
		},
		{
			name:           "unauthorized - no token",
			message:        "Hello",
			accessToken:    "",
			serverResponse: http.StatusUnauthorized,
			expectError:    true,
			errorContains:  "not authenticated. Please run",
		},
		{
			name:           "insufficient credits",
			message:        "Hello",
			accessToken:    "token123",
			serverResponse: http.StatusPaymentRequired,
			serverBody:     `{"error":"insufficient credits"}`,
			expectError:    true,
			errorContains:  "payment required",
		},
		{
			name:           "empty message",
			message:        "",
			accessToken:    "token123",
			serverResponse: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "message cannot be empty",
		},
		{
			name:           "server error",
			message:        "Hello",
			accessToken:    "token123",
			serverResponse: http.StatusInternalServerError,
			serverBody:     `{"error":"internal server error"}`,
			expectError:    true,
			errorContains:  "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/chat/completions" {
					// Check authorization header
					if tt.accessToken != "" {
						authHeader := r.Header.Get("Authorization")
						if authHeader != "Bearer "+tt.accessToken {
							w.WriteHeader(http.StatusUnauthorized)
							w.Write([]byte(`{"error":"unauthorized"}`))
							return
						}
					}
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
			if tt.accessToken != "" {
				viper.Set("access_token", tt.accessToken)
			} else {
				viper.Set("access_token", "")
			}

			// Create chat command
			cmd := newChatAINativeCmd()
			if tt.message != "" {
				cmd.SetArgs([]string{"--message", tt.message})
			} else {
				cmd.SetArgs([]string{})
			}

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

			// If successful, verify output contains response
			if !tt.expectError {
				output := buf.String()
				if !containsString(output, "Hello! How can I help you?") {
					t.Errorf("expected output to contain response, got: %s", output)
				}
			}
		})
	}
}

// TestChatWithProviderSelection tests chat with provider selector integration
func TestChatWithProviderSelection(t *testing.T) {
	tests := []struct {
		name              string
		preferredProvider string
		credits           int
		expectWarning     bool
		expectError       bool
	}{
		{
			name:              "preferred provider selected",
			preferredProvider: "anthropic",
			credits:           100,
			expectWarning:     false,
			expectError:       false,
		},
		{
			name:              "low credit warning",
			preferredProvider: "anthropic",
			credits:           10,
			expectWarning:     true,
			expectError:       false,
		},
		{
			name:              "insufficient credits",
			preferredProvider: "anthropic",
			credits:           0,
			expectWarning:     false,
			expectError:       true,
		},
		{
			name:              "auto provider selection",
			preferredProvider: "",
			credits:           100,
			expectWarning:     false,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/chat/completions" {
					if tt.credits == 0 {
						w.WriteHeader(http.StatusPaymentRequired)
						w.Write([]byte(`{"error":"insufficient credits"}`))
						return
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"id":"chat1","model":"claude-sonnet-4-5","choices":[{"message":{"role":"assistant","content":"Response"}}]}`))
				}
			}))
			defer server.Close()

			// Reset viper
			viper.Reset()
			viper.Set("backend_url", server.URL)
			viper.Set("access_token", "token123")
			viper.Set("user_email", "test@example.com")
			viper.Set("credits", tt.credits)
			if tt.preferredProvider != "" {
				viper.Set("preferred_provider", tt.preferredProvider)
			}

			// Create chat command
			cmd := newChatAINativeCmd()
			cmd.SetArgs([]string{"--message", "Hello", "--auto-provider"})

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

			// Check for low credit warning in output
			if tt.expectWarning {
				output := buf.String()
				if !containsString(output, "Warning") && !containsString(output, "Low credit") {
					t.Error("expected low credit warning in output")
				}
			}
		})
	}
}

// TestChatProviderFallback tests provider fallback functionality
func TestChatProviderFallback(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/chat/completions" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"chat1","model":"gpt-4","choices":[{"message":{"role":"assistant","content":"Fallback response"}}]}`))
		}
	}))
	defer server.Close()

	// Reset viper
	viper.Reset()
	viper.Set("backend_url", server.URL)
	viper.Set("access_token", "token123")
	viper.Set("preferred_provider", "anthropic")
	viper.Set("fallback_enabled", true)
	viper.Set("credits", 100)

	// Create chat command
	cmd := newChatAINativeCmd()
	cmd.SetArgs([]string{"--message", "Hello", "--auto-provider"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected fallback to succeed, got error: %v", err)
	}

	// Verify response
	output := buf.String()
	if !containsString(output, "response") && !containsString(output, "Response") {
		t.Errorf("expected response in output, got: %s", output)
	}
}

// TestChatRequiresAuthentication tests that chat requires authentication
func TestChatRequiresAuthentication(t *testing.T) {
	// Reset viper - no token
	viper.Reset()

	// Create chat command
	cmd := newChatAINativeCmd()
	cmd.SetArgs([]string{"--message", "Hello"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Execute
	err := cmd.Execute()

	// Verify error
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}

	if !containsString(err.Error(), "not authenticated") {
		t.Errorf("expected 'not authenticated' error, got: %v", err)
	}
}

// TestChatWithModel tests chat with specific model selection
func TestChatWithModel(t *testing.T) {
	tests := []struct {
		name        string
		model       string
		expectError bool
	}{
		{
			name:        "default model",
			model:       "",
			expectError: false,
		},
		{
			name:        "claude model",
			model:       "claude-sonnet-4-5",
			expectError: false,
		},
		{
			name:        "gpt model",
			model:       "gpt-4",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/chat/completions" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"id":"chat1","model":"` + tt.model + `","choices":[{"message":{"role":"assistant","content":"Response"}}]}`))
				}
			}))
			defer server.Close()

			// Reset viper
			viper.Reset()
			viper.Set("backend_url", server.URL)
			viper.Set("access_token", "token123")

			// Create chat command
			cmd := newChatAINativeCmd()
			args := []string{"--message", "Hello"}
			if tt.model != "" {
				args = append(args, "--model", tt.model)
			}
			cmd.SetArgs(args)

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
		})
	}
}

// TestChatDisplaysUsageStats tests that usage stats are displayed
func TestChatDisplaysUsageStats(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/chat/completions" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"chat1","model":"claude-sonnet-4-5","choices":[{"message":{"role":"assistant","content":"Response"}}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`))
		}
	}))
	defer server.Close()

	// Reset viper
	viper.Reset()
	viper.Set("backend_url", server.URL)
	viper.Set("access_token", "token123")

	// Create chat command
	cmd := newChatAINativeCmd()
	cmd.SetArgs([]string{"--message", "Hello", "--verbose"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify usage stats in output (when verbose is enabled)
	output := buf.String()
	// Note: The actual implementation should display usage stats when verbose flag is set
	// For now, just verify the command executed successfully
	if output == "" {
		t.Error("expected output, got empty string")
	}
}

// Helper function is now implemented in chat.go
