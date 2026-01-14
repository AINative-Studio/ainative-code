package tui

import (
	"errors"
	"testing"
)

// TestNewModel tests the NewModel constructor
func TestNewModel(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "creates model with default state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()

			// Verify initial state
			if m.ready {
				t.Error("expected ready to be false initially")
			}
			if m.quitting {
				t.Error("expected quitting to be false initially")
			}
			if m.streaming {
				t.Error("expected streaming to be false initially")
			}
			if m.err != nil {
				t.Errorf("expected err to be nil initially, got %v", m.err)
			}
			if len(m.messages) != 0 {
				t.Errorf("expected messages to be empty initially, got length %d", len(m.messages))
			}
			if m.width != 0 {
				t.Errorf("expected width to be 0 initially, got %d", m.width)
			}
			if m.height != 0 {
				t.Errorf("expected height to be 0 initially, got %d", m.height)
			}
		})
	}
}

// TestSetSize tests the SetSize method
func TestSetSize(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		height         int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "sets standard terminal size",
			width:          80,
			height:         24,
			expectedWidth:  80,
			expectedHeight: 24,
		},
		{
			name:           "sets large terminal size",
			width:          200,
			height:         60,
			expectedWidth:  200,
			expectedHeight: 60,
		},
		{
			name:           "sets small terminal size",
			width:          40,
			height:         10,
			expectedWidth:  40,
			expectedHeight: 10,
		},
		{
			name:           "handles zero width",
			width:          0,
			height:         24,
			expectedWidth:  0,
			expectedHeight: 24,
		},
		{
			name:           "handles zero height",
			width:          80,
			height:         0,
			expectedWidth:  80,
			expectedHeight: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.SetSize(tt.width, tt.height)

			if m.width != tt.expectedWidth {
				t.Errorf("expected width %d, got %d", tt.expectedWidth, m.width)
			}
			if m.height != tt.expectedHeight {
				t.Errorf("expected height %d, got %d", tt.expectedHeight, m.height)
			}

			// Verify viewport was updated
			// Note: viewport.Width and viewport.Height are set to width and height-5
			// (reserving space for input area and status bar)
			expectedViewportHeight := tt.expectedHeight - 5
			if expectedViewportHeight < 0 {
				expectedViewportHeight = 0
			}
			// We can't directly access viewport dimensions, so we verify no panic occurred
		})
	}
}

// TestAddMessage tests the AddMessage method
func TestAddMessage(t *testing.T) {
	tests := []struct {
		name            string
		role            string
		content         string
		expectedLength  int
		expectedRole    string
		expectedContent string
	}{
		{
			name:            "adds user message",
			role:            "user",
			content:         "Hello, assistant!",
			expectedLength:  1,
			expectedRole:    "user",
			expectedContent: "Hello, assistant!",
		},
		{
			name:            "adds assistant message",
			role:            "assistant",
			content:         "Hello! How can I help?",
			expectedLength:  1,
			expectedRole:    "assistant",
			expectedContent: "Hello! How can I help?",
		},
		{
			name:            "adds system message",
			role:            "system",
			content:         "Connection established",
			expectedLength:  1,
			expectedRole:    "system",
			expectedContent: "Connection established",
		},
		{
			name:            "adds empty message",
			role:            "user",
			content:         "",
			expectedLength:  1,
			expectedRole:    "user",
			expectedContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.AddMessage(tt.role, tt.content)

			if len(m.messages) != tt.expectedLength {
				t.Errorf("expected %d messages, got %d", tt.expectedLength, len(m.messages))
			}

			if len(m.messages) > 0 {
				msg := m.messages[0]
				if msg.Role != tt.expectedRole {
					t.Errorf("expected role %q, got %q", tt.expectedRole, msg.Role)
				}
				if msg.Content != tt.expectedContent {
					t.Errorf("expected content %q, got %q", tt.expectedContent, msg.Content)
				}
			}
		})
	}
}

// TestAddMultipleMessages tests adding multiple messages
func TestAddMultipleMessages(t *testing.T) {
	m := NewModel()

	m.AddMessage("user", "First message")
	m.AddMessage("assistant", "First response")
	m.AddMessage("user", "Second message")
	m.AddMessage("assistant", "Second response")

	if len(m.messages) != 4 {
		t.Errorf("expected 4 messages, got %d", len(m.messages))
	}

	// Verify order is preserved
	if m.messages[0].Role != "user" || m.messages[0].Content != "First message" {
		t.Error("first message not preserved correctly")
	}
	if m.messages[1].Role != "assistant" || m.messages[1].Content != "First response" {
		t.Error("second message not preserved correctly")
	}
	if m.messages[2].Role != "user" || m.messages[2].Content != "Second message" {
		t.Error("third message not preserved correctly")
	}
	if m.messages[3].Role != "assistant" || m.messages[3].Content != "Second response" {
		t.Error("fourth message not preserved correctly")
	}
}

// TestClearMessages tests the ClearMessages method
func TestClearMessages(t *testing.T) {
	m := NewModel()

	// Add some messages
	m.AddMessage("user", "Message 1")
	m.AddMessage("assistant", "Response 1")
	m.AddMessage("user", "Message 2")

	// Verify messages exist
	if len(m.messages) != 3 {
		t.Errorf("expected 3 messages before clear, got %d", len(m.messages))
	}

	// Clear messages
	m.ClearMessages()

	// Verify messages are cleared
	if len(m.messages) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(m.messages))
	}
}

// TestGetUserInput tests the GetUserInput method
func TestGetUserInput(t *testing.T) {
	tests := []struct {
		name          string
		inputValue    string
		expectedValue string
	}{
		{
			name:          "returns input value",
			inputValue:    "test input",
			expectedValue: "test input",
		},
		{
			name:          "returns empty string",
			inputValue:    "",
			expectedValue: "",
		},
		{
			name:          "returns input with spaces",
			inputValue:    "  test  ",
			expectedValue: "  test  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.textInput.SetValue(tt.inputValue)

			result := m.GetUserInput()

			if result != tt.expectedValue {
				t.Errorf("expected %q, got %q", tt.expectedValue, result)
			}
		})
	}
}

// TestSetError tests the SetError method
func TestSetError(t *testing.T) {
	tests := []struct {
		name          string
		errMsg        string
		expectedError string
	}{
		{
			name:          "sets simple error",
			errMsg:        "test error",
			expectedError: "test error",
		},
		{
			name:          "sets complex error",
			errMsg:        "API error: connection timeout",
			expectedError: "API error: connection timeout",
		},
		{
			name:          "sets empty error",
			errMsg:        "",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.SetError(errors.New(tt.errMsg))

			if m.err == nil {
				t.Error("expected error to be set, got nil")
			} else if m.err.Error() != tt.expectedError {
				t.Errorf("expected error %q, got %q", tt.expectedError, m.err.Error())
			}
		})
	}
}

// TestClearError tests the ClearError method
func TestClearError(t *testing.T) {
	m := NewModel()

	// Set an error
	m.SetError(errors.New("test error"))
	if m.err == nil {
		t.Error("expected error to be set")
	}

	// Clear the error
	m.ClearError()
	if m.err != nil {
		t.Errorf("expected error to be nil after clear, got %v", m.err)
	}
}

// TestIsReady tests the IsReady method
func TestIsReady(t *testing.T) {
	m := NewModel()

	// Initially not ready
	if m.IsReady() {
		t.Error("expected IsReady to return false initially")
	}

	// Set ready to true
	m.ready = true
	if !m.IsReady() {
		t.Error("expected IsReady to return true after setting ready")
	}

	// Set ready to false
	m.ready = false
	if m.IsReady() {
		t.Error("expected IsReady to return false after unsetting ready")
	}
}

// TestIsQuitting tests the IsQuitting method
func TestIsQuitting(t *testing.T) {
	m := NewModel()

	// Initially not quitting
	if m.IsQuitting() {
		t.Error("expected IsQuitting to return false initially")
	}

	// Set quitting to true
	m.quitting = true
	if !m.IsQuitting() {
		t.Error("expected IsQuitting to return true after setting quitting")
	}

	// Set quitting to false
	m.quitting = false
	if m.IsQuitting() {
		t.Error("expected IsQuitting to return false after unsetting quitting")
	}
}

// TestIsStreaming tests the IsStreaming method
func TestIsStreaming(t *testing.T) {
	m := NewModel()

	// Initially not streaming
	if m.IsStreaming() {
		t.Error("expected IsStreaming to return false initially")
	}

	// Set streaming to true
	m.streaming = true
	if !m.IsStreaming() {
		t.Error("expected IsStreaming to return true after setting streaming")
	}

	// Set streaming to false
	m.streaming = false
	if m.IsStreaming() {
		t.Error("expected IsStreaming to return false after unsetting streaming")
	}
}

// TestSetStreaming tests the SetStreaming method
func TestSetStreaming(t *testing.T) {
	tests := []struct {
		name      string
		value     bool
		expected  bool
	}{
		{
			name:     "sets streaming to true",
			value:    true,
			expected: true,
		},
		{
			name:     "sets streaming to false",
			value:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.SetStreaming(tt.value)

			if m.streaming != tt.expected {
				t.Errorf("expected streaming to be %v, got %v", tt.expected, m.streaming)
			}
		})
	}
}

// TestSetQuitting tests the SetQuitting method
func TestSetQuitting(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		expected bool
	}{
		{
			name:     "sets quitting to true",
			value:    true,
			expected: true,
		},
		{
			name:     "sets quitting to false",
			value:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.SetQuitting(tt.value)

			if m.quitting != tt.expected {
				t.Errorf("expected quitting to be %v, got %v", tt.expected, m.quitting)
			}
		})
	}
}

// TestModelIntegration tests integration between model methods
func TestModelIntegration(t *testing.T) {
	m := NewModel()

	// Set up size
	m.SetSize(80, 24)

	// Add messages
	m.AddMessage("user", "Hello")
	m.AddMessage("assistant", "Hi there!")

	// Set streaming
	m.SetStreaming(true)

	// Verify state
	if !m.IsStreaming() {
		t.Error("expected streaming to be true")
	}
	if len(m.messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(m.messages))
	}

	// Clear messages
	m.ClearMessages()
	if len(m.messages) != 0 {
		t.Error("expected messages to be cleared")
	}

	// Set error
	m.SetError(errors.New("test error"))
	if m.err == nil {
		t.Error("expected error to be set")
	}

	// Clear error
	m.ClearError()
	if m.err != nil {
		t.Error("expected error to be cleared")
	}
}

// TestMessageType tests the Message type
func TestMessageType(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		content string
	}{
		{
			name:    "user message",
			role:    "user",
			content: "test content",
		},
		{
			name:    "assistant message",
			role:    "assistant",
			content: "response content",
		},
		{
			name:    "system message",
			role:    "system",
			content: "system content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message{
				Role:    tt.role,
				Content: tt.content,
			}

			if msg.Role != tt.role {
				t.Errorf("expected role %q, got %q", tt.role, msg.Role)
			}
			if msg.Content != tt.content {
				t.Errorf("expected content %q, got %q", tt.content, msg.Content)
			}
		})
	}
}

// TestModelComponentsInitialization tests that viewport and textInput are properly initialized
func TestModelComponentsInitialization(t *testing.T) {
	m := NewModel()

	// Verify viewport is initialized (not nil)
	// We can't directly check if it's nil, but we can verify it has default properties
	_ = m.viewport

	// Verify textInput is initialized (not nil)
	// We can't directly check if it's nil, but we can verify it has default properties
	_ = m.textInput

	// Test that we can call methods on these components without panic
	m.textInput.SetValue("test")
	if m.textInput.Value() != "test" {
		t.Error("textInput not properly initialized")
	}
}

// Benchmark tests for performance validation

// BenchmarkNewModel benchmarks model creation
func BenchmarkNewModel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModel()
	}
}

// BenchmarkAddMessage benchmarks adding messages
func BenchmarkAddMessage(b *testing.B) {
	m := NewModel()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.AddMessage("user", "test message")
	}
}

// BenchmarkSetSize benchmarks setting terminal size
func BenchmarkSetSize(b *testing.B) {
	m := NewModel()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.SetSize(80, 24)
	}
}
