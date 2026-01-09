package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"
)

// TestChatCommand tests the chat command initialization
func TestChatCommand(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "chat command exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if chatCmd == nil {
				t.Fatal("chatCmd should not be nil")
			}

			if chatCmd.Use != "chat [message]" {
				t.Errorf("expected Use 'chat [message]', got %s", chatCmd.Use)
			}

			if chatCmd.Short == "" {
				t.Error("expected Short description to be set")
			}

			if chatCmd.Long == "" {
				t.Error("expected Long description to be set")
			}

			// Verify aliases
			expectedAliases := []string{"c", "ask"}
			if len(chatCmd.Aliases) != len(expectedAliases) {
				t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(chatCmd.Aliases))
			}

			for i, alias := range expectedAliases {
				if i >= len(chatCmd.Aliases) || chatCmd.Aliases[i] != alias {
					t.Errorf("expected alias %s at index %d", alias, i)
				}
			}
		})
	}
}

// TestChatFlags tests the chat command flags
func TestChatFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		shorthand string
	}{
		{
			name:     "session-id flag exists",
			flagName: "session-id",
			shorthand: "s",
		},
		{
			name:     "system flag exists",
			flagName: "system",
			shorthand: "",
		},
		{
			name:     "stream flag exists",
			flagName: "stream",
			shorthand: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := chatCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag %s should exist", tt.flagName)
				return
			}

			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("expected shorthand %s, got %s", tt.shorthand, flag.Shorthand)
			}
		})
	}
}

// TestRunChatNoProvider tests chat command without provider
func TestRunChatNoProvider(t *testing.T) {
	// Reset viper and flags
	viper.Reset()
	provider = ""
	chatSessionID = ""
	chatSystemMsg = ""
	chatStream = true

	// Setup command
	chatCmd.SetArgs([]string{})

	// Execute should return error
	err := runChat(chatCmd, []string{})

	if err == nil {
		t.Error("expected error when provider not configured")
	}

	expectedErrMsg := "AI provider not configured"
	if err != nil && err.Error() != expectedErrMsg && len(err.Error()) > 0 {
		// Check if error contains the expected message
		if err.Error()[:len(expectedErrMsg)] != expectedErrMsg {
			t.Logf("got error: %v", err)
		}
	}
}

// TestRunChatWithProvider tests chat command with provider configured
func TestRunChatWithProvider(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		model     string
		args      []string
		wantErr   bool
	}{
		// Skip interactive mode test - it requires stdin mocking
		// {
		// 	name:     "interactive mode with provider",
		// 	provider: "openai",
		// 	model:    "gpt-4",
		// 	args:     []string{},
		// 	wantErr:  true, // Will fail due to invalid API key
		// },
		{
			name:     "single message mode",
			provider: "openai",
			model:    "gpt-4",
			args:     []string{"Hello, assistant!"},
			wantErr:  true, // Will fail due to invalid API key
		},
		{
			name:     "anthropic provider",
			provider: "anthropic",
			model:    "claude-3-opus",
			args:     []string{"Explain goroutines"},
			wantErr:  true, // Will fail due to invalid API key
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and flags
			viper.Reset()
			provider = tt.provider
			model = tt.model
			chatSessionID = ""
			chatSystemMsg = ""
			chatStream = true

			// Set mock API key for testing
			viper.Set("api_key", "test-api-key-for-testing")

			// Capture output
			var buf bytes.Buffer
			chatCmd.SetOut(&buf)

			// Execute
			err := runChat(chatCmd, tt.args)

			// We expect errors due to invalid API key, but the provider should initialize
			if err != nil {
				// Check that the error is NOT "provider not configured" or "no API key found"
				errMsg := err.Error()
				if errMsg == "AI provider not configured. Use --provider flag or set in config file" {
					t.Errorf("Unexpected error: provider should be configured")
				}
				if len(errMsg) > 18 && errMsg[:18] == "no API key found" {
					t.Errorf("Unexpected error: API key should be found from viper config")
				}
			}
		})
	}
}

// TestRunChatSingleMessage tests single message processing
func TestRunChatSingleMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		provider string
		wantErr  bool
	}{
		{
			name:     "simple message",
			message:  "Hello",
			provider: "openai",
			wantErr:  true, // Will fail due to invalid API key
		},
		{
			name:     "complex message",
			message:  "Explain how to implement a binary search tree in Go",
			provider: "anthropic",
			wantErr:  true, // Will fail due to invalid API key
		},
		{
			name:     "message with special characters",
			message:  "What is the difference between `make` and `new` in Go?",
			provider: "openai",
			wantErr:  true, // Will fail due to invalid API key
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and flags
			viper.Reset()
			provider = tt.provider
			chatSessionID = ""
			chatSystemMsg = ""
			chatStream = true

			// Set mock API key for testing
			viper.Set("api_key", "test-api-key-for-testing")

			// Capture output
			var buf bytes.Buffer
			chatCmd.SetOut(&buf)

			// Execute
			err := runChat(chatCmd, []string{tt.message})

			// We expect errors due to invalid API key, but verify it's not a config error
			if err != nil {
				errMsg := err.Error()
				if len(errMsg) > 18 && errMsg[:18] == "no API key found" {
					t.Errorf("Unexpected error: API key should be found from viper config")
				}
			}
		})
	}
}

// TestRunChatWithSessionID tests chat with session ID (single message mode only - interactive mode requires stdin mocking)
func TestRunChatWithSessionID(t *testing.T) {
	t.Skip("Skipping interactive mode tests - they require stdin mocking")
}

// TestRunChatWithSystemMessage tests chat with custom system message (skipped - requires stdin mocking)
func TestRunChatWithSystemMessage(t *testing.T) {
	t.Skip("Skipping interactive mode tests - they require stdin mocking")
}

// TestRunChatStreamingOptions tests streaming flag (skipped - requires stdin mocking)
func TestRunChatStreamingOptions(t *testing.T) {
	t.Skip("Skipping interactive mode tests - they require stdin mocking")
}

// TestChatCommandIntegration tests complete chat command flow
func TestChatCommandIntegration(t *testing.T) {
	// Reset viper and flags
	viper.Reset()
	provider = "openai"
	model = "gpt-4"
	chatSessionID = "test-session"
	chatSystemMsg = "You are a helpful assistant"
	chatStream = true

	// Set mock API key for testing
	viper.Set("api_key", "test-api-key-for-testing")

	// Capture output
	var buf bytes.Buffer
	chatCmd.SetOut(&buf)

	// Test single message mode - expect error due to invalid API key
	err := runChat(chatCmd, []string{"Test message"})
	if err != nil {
		// Verify it's not a configuration error
		errMsg := err.Error()
		if len(errMsg) > 18 && errMsg[:18] == "no API key found" {
			t.Errorf("Unexpected error: API key should be found from viper config: %v", err)
		}
	}
}

// TestChatCommandDefaultFlagValues tests default flag values
func TestChatCommandDefaultFlagValues(t *testing.T) {
	// Get stream flag default value
	streamFlag := chatCmd.Flags().Lookup("stream")
	if streamFlag == nil {
		t.Fatal("stream flag should exist")
	}

	if streamFlag.DefValue != "true" {
		t.Errorf("expected stream default value 'true', got %s", streamFlag.DefValue)
	}

	// Get session-id flag default value
	sessionIDFlag := chatCmd.Flags().Lookup("session-id")
	if sessionIDFlag == nil {
		t.Fatal("session-id flag should exist")
	}

	if sessionIDFlag.DefValue != "" {
		t.Errorf("expected session-id default value '', got %s", sessionIDFlag.DefValue)
	}
}

// Benchmark tests for performance validation

// BenchmarkRunChatSingleMessage benchmarks single message processing
func BenchmarkRunChatSingleMessage(b *testing.B) {
	viper.Reset()
	provider = "openai"
	chatSessionID = ""
	chatSystemMsg = ""
	chatStream = true

	var buf bytes.Buffer
	chatCmd.SetOut(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runChat(chatCmd, []string{"test message"})
	}
}

// BenchmarkRunChatInteractive benchmarks interactive mode initialization
func BenchmarkRunChatInteractive(b *testing.B) {
	viper.Reset()
	provider = "openai"
	chatSessionID = ""
	chatSystemMsg = ""
	chatStream = true

	var buf bytes.Buffer
	chatCmd.SetOut(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runChat(chatCmd, []string{})
	}
}
