package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestErrMsg tests the errMsg message type
func TestErrMsg(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedError string
	}{
		{
			name:          "creates errMsg with simple error",
			err:           errors.New("test error"),
			expectedError: "test error",
		},
		{
			name:          "creates errMsg with complex error",
			err:           errors.New("API error: connection timeout"),
			expectedError: "API error: connection timeout",
		},
		{
			name:          "creates errMsg with wrapped error",
			err:           errors.New("outer error: inner error"),
			expectedError: "outer error: inner error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := errMsg{err: tt.err}

			// Verify message type
			var _ tea.Msg = msg

			// Verify error content
			if msg.err.Error() != tt.expectedError {
				t.Errorf("expected error %q, got %q", tt.expectedError, msg.err.Error())
			}
		})
	}
}

// TestReadyMsg tests the readyMsg message type
func TestReadyMsg(t *testing.T) {
	msg := readyMsg{}

	// Verify message type implements tea.Msg
	var _ tea.Msg = msg

	// readyMsg is a signal type with no fields to test
	// Just verify it can be created
}

// TestStreamChunkMsg tests the streamChunkMsg message type
func TestStreamChunkMsg(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedContent string
	}{
		{
			name:            "creates streamChunkMsg with text content",
			content:         "Hello, world!",
			expectedContent: "Hello, world!",
		},
		{
			name:            "creates streamChunkMsg with empty content",
			content:         "",
			expectedContent: "",
		},
		{
			name:            "creates streamChunkMsg with whitespace content",
			content:         "   ",
			expectedContent: "   ",
		},
		{
			name:            "creates streamChunkMsg with multiline content",
			content:         "Line 1\nLine 2\nLine 3",
			expectedContent: "Line 1\nLine 2\nLine 3",
		},
		{
			name:            "creates streamChunkMsg with special characters",
			content:         "Special: @#$%^&*()",
			expectedContent: "Special: @#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := streamChunkMsg{content: tt.content}

			// Verify message type
			var _ tea.Msg = msg

			// Verify content
			if msg.content != tt.expectedContent {
				t.Errorf("expected content %q, got %q", tt.expectedContent, msg.content)
			}
		})
	}
}

// TestStreamDoneMsg tests the streamDoneMsg message type
func TestStreamDoneMsg(t *testing.T) {
	msg := streamDoneMsg{}

	// Verify message type implements tea.Msg
	var _ tea.Msg = msg

	// streamDoneMsg is a signal type with no fields to test
	// Just verify it can be created
}

// TestUserInputMsg tests the userInputMsg message type
func TestUserInputMsg(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedInput string
	}{
		{
			name:          "creates userInputMsg with text input",
			input:         "What is the weather?",
			expectedInput: "What is the weather?",
		},
		{
			name:          "creates userInputMsg with empty input",
			input:         "",
			expectedInput: "",
		},
		{
			name:          "creates userInputMsg with whitespace input",
			input:         "   trimmed   ",
			expectedInput: "   trimmed   ",
		},
		{
			name:          "creates userInputMsg with multiline input",
			input:         "Line 1\nLine 2",
			expectedInput: "Line 1\nLine 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := userInputMsg{input: tt.input}

			// Verify message type
			var _ tea.Msg = msg

			// Verify input
			if msg.input != tt.expectedInput {
				t.Errorf("expected input %q, got %q", tt.expectedInput, msg.input)
			}
		})
	}
}

// TestSendError tests the SendError helper function
func TestSendError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedError string
	}{
		{
			name:          "sends simple error",
			err:           errors.New("test error"),
			expectedError: "test error",
		},
		{
			name:          "sends complex error",
			err:           errors.New("API error: timeout"),
			expectedError: "API error: timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := SendError(tt.err)

			// Execute command
			msg := cmd()

			// Verify message type
			errMessage, ok := msg.(errMsg)
			if !ok {
				t.Fatalf("expected errMsg, got %T", msg)
			}

			// Verify error content
			if errMessage.err.Error() != tt.expectedError {
				t.Errorf("expected error %q, got %q", tt.expectedError, errMessage.err.Error())
			}
		})
	}
}

// TestSendReady tests the SendReady helper function
func TestSendReady(t *testing.T) {
	cmd := SendReady()

	// Execute command
	msg := cmd()

	// Verify message type
	_, ok := msg.(readyMsg)
	if !ok {
		t.Fatalf("expected readyMsg, got %T", msg)
	}
}

// TestSendStreamChunk tests the SendStreamChunk helper function
func TestSendStreamChunk(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedContent string
	}{
		{
			name:            "sends text chunk",
			content:         "Hello",
			expectedContent: "Hello",
		},
		{
			name:            "sends empty chunk",
			content:         "",
			expectedContent: "",
		},
		{
			name:            "sends chunk with special characters",
			content:         "Special: !@#$%",
			expectedContent: "Special: !@#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := SendStreamChunk(tt.content)

			// Execute command
			msg := cmd()

			// Verify message type
			chunkMessage, ok := msg.(streamChunkMsg)
			if !ok {
				t.Fatalf("expected streamChunkMsg, got %T", msg)
			}

			// Verify content
			if chunkMessage.content != tt.expectedContent {
				t.Errorf("expected content %q, got %q", tt.expectedContent, chunkMessage.content)
			}
		})
	}
}

// TestSendStreamDone tests the SendStreamDone helper function
func TestSendStreamDone(t *testing.T) {
	cmd := SendStreamDone()

	// Execute command
	msg := cmd()

	// Verify message type
	_, ok := msg.(streamDoneMsg)
	if !ok {
		t.Fatalf("expected streamDoneMsg, got %T", msg)
	}
}

// TestSendUserInput tests the SendUserInput helper function
func TestSendUserInput(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedInput string
	}{
		{
			name:          "sends user input",
			input:         "Hello, assistant",
			expectedInput: "Hello, assistant",
		},
		{
			name:          "sends empty input",
			input:         "",
			expectedInput: "",
		},
		{
			name:          "sends input with spaces",
			input:         "  test  ",
			expectedInput: "  test  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := SendUserInput(tt.input)

			// Execute command
			msg := cmd()

			// Verify message type
			inputMessage, ok := msg.(userInputMsg)
			if !ok {
				t.Fatalf("expected userInputMsg, got %T", msg)
			}

			// Verify input
			if inputMessage.input != tt.expectedInput {
				t.Errorf("expected input %q, got %q", tt.expectedInput, inputMessage.input)
			}
		})
	}
}

// TestMessageTypesIntegration tests interaction between different message types
func TestMessageTypesIntegration(t *testing.T) {
	// Test that all message types can be created and used together
	messages := []tea.Msg{
		errMsg{err: errors.New("test error")},
		readyMsg{},
		streamChunkMsg{content: "chunk"},
		streamDoneMsg{},
		userInputMsg{input: "input"},
	}

	// Verify all messages implement tea.Msg interface
	for i, msg := range messages {
		var _ tea.Msg = msg
		if msg == nil {
			t.Errorf("message %d is nil", i)
		}
	}
}

// TestHelperFunctionsReturnValidCommands tests that all helper functions return valid tea.Cmd
func TestHelperFunctionsReturnValidCommands(t *testing.T) {
	commands := []tea.Cmd{
		SendError(errors.New("test")),
		SendReady(),
		SendStreamChunk("chunk"),
		SendStreamDone(),
		SendUserInput("input"),
	}

	// Verify all commands are not nil and can be executed
	for i, cmd := range commands {
		if cmd == nil {
			t.Errorf("command %d is nil", i)
			continue
		}

		// Execute command and verify it returns a message
		msg := cmd()
		if msg == nil {
			t.Errorf("command %d returned nil message", i)
		}
	}
}

// Benchmark tests for performance validation

// BenchmarkSendError benchmarks error message creation
func BenchmarkSendError(b *testing.B) {
	err := errors.New("test error")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := SendError(err)
		_ = cmd()
	}
}

// BenchmarkSendStreamChunk benchmarks stream chunk message creation
func BenchmarkSendStreamChunk(b *testing.B) {
	content := "test content"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := SendStreamChunk(content)
		_ = cmd()
	}
}

// BenchmarkSendUserInput benchmarks user input message creation
func BenchmarkSendUserInput(b *testing.B) {
	input := "test input"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := SendUserInput(input)
		_ = cmd()
	}
}

// BenchmarkMessageTypeCreation benchmarks creation of all message types
func BenchmarkMessageTypeCreation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = errMsg{err: errors.New("error")}
		_ = readyMsg{}
		_ = streamChunkMsg{content: "content"}
		_ = streamDoneMsg{}
		_ = userInputMsg{input: "input"}
	}
}
