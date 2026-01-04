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
		{
			name:     "interactive mode with provider",
			provider: "openai",
			model:    "gpt-4",
			args:     []string{},
			wantErr:  false,
		},
		{
			name:     "single message mode",
			provider: "openai",
			model:    "gpt-4",
			args:     []string{"Hello, assistant!"},
			wantErr:  false,
		},
		{
			name:     "anthropic provider",
			provider: "anthropic",
			model:    "claude-3-opus",
			args:     []string{"Explain goroutines"},
			wantErr:  false,
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

			// Capture output
			var buf bytes.Buffer
			chatCmd.SetOut(&buf)

			// Execute
			err := runChat(chatCmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runChat() error = %v, wantErr %v", err, tt.wantErr)
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
			wantErr:  false,
		},
		{
			name:     "complex message",
			message:  "Explain how to implement a binary search tree in Go",
			provider: "anthropic",
			wantErr:  false,
		},
		{
			name:     "message with special characters",
			message:  "What is the difference between `make` and `new` in Go?",
			provider: "openai",
			wantErr:  false,
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

			// Capture output
			var buf bytes.Buffer
			chatCmd.SetOut(&buf)

			// Execute
			err := runChat(chatCmd, []string{tt.message})

			if (err != nil) != tt.wantErr {
				t.Errorf("runChat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRunChatWithSessionID tests chat with session ID
func TestRunChatWithSessionID(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		provider  string
		wantErr   bool
	}{
		{
			name:      "with valid session ID",
			sessionID: "session-123",
			provider:  "openai",
			wantErr:   false,
		},
		{
			name:      "with UUID session ID",
			sessionID: "550e8400-e29b-41d4-a716-446655440000",
			provider:  "anthropic",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and flags
			viper.Reset()
			provider = tt.provider
			chatSessionID = tt.sessionID
			chatSystemMsg = ""
			chatStream = true

			// Capture output
			var buf bytes.Buffer
			chatCmd.SetOut(&buf)

			// Execute
			err := runChat(chatCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runChat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRunChatWithSystemMessage tests chat with custom system message
func TestRunChatWithSystemMessage(t *testing.T) {
	tests := []struct {
		name      string
		systemMsg string
		provider  string
		wantErr   bool
	}{
		{
			name:      "with custom system message",
			systemMsg: "You are a helpful Go programming assistant",
			provider:  "openai",
			wantErr:   false,
		},
		{
			name:      "with empty system message",
			systemMsg: "",
			provider:  "anthropic",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and flags
			viper.Reset()
			provider = tt.provider
			chatSessionID = ""
			chatSystemMsg = tt.systemMsg
			chatStream = true

			// Capture output
			var buf bytes.Buffer
			chatCmd.SetOut(&buf)

			// Execute
			err := runChat(chatCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runChat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRunChatStreamingOptions tests streaming flag
func TestRunChatStreamingOptions(t *testing.T) {
	tests := []struct {
		name     string
		stream   bool
		provider string
		wantErr  bool
	}{
		{
			name:     "with streaming enabled",
			stream:   true,
			provider: "openai",
			wantErr:  false,
		},
		{
			name:     "with streaming disabled",
			stream:   false,
			provider: "openai",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and flags
			viper.Reset()
			provider = tt.provider
			chatSessionID = ""
			chatSystemMsg = ""
			chatStream = tt.stream

			// Capture output
			var buf bytes.Buffer
			chatCmd.SetOut(&buf)

			// Execute
			err := runChat(chatCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runChat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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

	// Capture output
	var buf bytes.Buffer
	chatCmd.SetOut(&buf)

	// Test interactive mode
	err := runChat(chatCmd, []string{})
	if err != nil {
		t.Errorf("interactive mode failed: %v", err)
	}

	// Reset for single message mode
	buf.Reset()

	// Test single message mode
	err = runChat(chatCmd, []string{"Test message"})
	if err != nil {
		t.Errorf("single message mode failed: %v", err)
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
