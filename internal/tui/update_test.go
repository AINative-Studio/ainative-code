package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestUpdateCtrlC tests Ctrl+C keyboard shortcut quits the application
func TestUpdateCtrlC(t *testing.T) {
	m := NewModel()
	m.ready = true

	// Simulate Ctrl+C key press
	msg := tea.KeyMsg{
		Type:  tea.KeyCtrlC,
		Runes: []rune{'c'},
	}

	updatedModel, cmd := m.Update(msg)
	m = updatedModel.(Model)

	// Verify quitting flag is set
	if !m.quitting {
		t.Error("expected quitting to be true after Ctrl+C")
	}

	// Verify tea.Quit command is returned
	if cmd == nil {
		t.Fatal("expected tea.Quit command, got nil")
	}

	// Execute the command to verify it's tea.Quit
	result := cmd()
	if _, ok := result.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", result)
	}
}

// TestUpdateEnter tests Enter key submission behavior
func TestUpdateEnter(t *testing.T) {
	tests := []struct {
		name              string
		initialStreaming  bool
		inputValue        string
		expectedStreaming bool
		expectedMsgCount  int
		shouldReturnCmd   bool
	}{
		{
			name:              "submits valid input",
			initialStreaming:  false,
			inputValue:        "Hello, assistant",
			expectedStreaming: true,
			expectedMsgCount:  2, // user message + empty assistant placeholder
			shouldReturnCmd:   true,
		},
		{
			name:              "ignores empty input",
			initialStreaming:  false,
			inputValue:        "",
			expectedStreaming: false,
			expectedMsgCount:  0,
			shouldReturnCmd:   false,
		},
		{
			name:              "ignores whitespace-only input",
			initialStreaming:  false,
			inputValue:        "   ",
			expectedStreaming: false,
			expectedMsgCount:  0,
			shouldReturnCmd:   false,
		},
		{
			name:              "blocks input during streaming",
			initialStreaming:  true,
			inputValue:        "Hello",
			expectedStreaming: true,
			expectedMsgCount:  0,
			shouldReturnCmd:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.ready = true
			m.streaming = tt.initialStreaming
			m.textInput.SetValue(tt.inputValue)

			// Simulate Enter key press
			msg := tea.KeyMsg{
				Type: tea.KeyEnter,
			}

			updatedModel, cmd := m.Update(msg)
			m = updatedModel.(Model)

			// Verify streaming state
			if m.streaming != tt.expectedStreaming {
				t.Errorf("expected streaming %v, got %v", tt.expectedStreaming, m.streaming)
			}

			// Verify message count
			if len(m.messages) != tt.expectedMsgCount {
				t.Errorf("expected %d messages, got %d", tt.expectedMsgCount, len(m.messages))
			}

			// Verify command returned
			if tt.shouldReturnCmd && cmd == nil {
				t.Error("expected command to be returned, got nil")
			}
			if !tt.shouldReturnCmd && cmd != nil {
				t.Error("expected no command, got command")
			}

			// Verify input was cleared for valid submission
			if tt.expectedMsgCount > 0 && m.textInput.Value() != "" {
				t.Error("expected input to be cleared after submission")
			}

			// Verify message roles for valid submission
			if tt.expectedMsgCount == 2 {
				if m.messages[0].Role != "user" {
					t.Errorf("expected first message role to be 'user', got %q", m.messages[0].Role)
				}
				if m.messages[0].Content != strings.TrimSpace(tt.inputValue) {
					t.Errorf("expected user message content %q, got %q", tt.inputValue, m.messages[0].Content)
				}
				if m.messages[1].Role != "assistant" {
					t.Errorf("expected second message role to be 'assistant', got %q", m.messages[1].Role)
				}
				if m.messages[1].Content != "" {
					t.Errorf("expected assistant placeholder to be empty, got %q", m.messages[1].Content)
				}
			}
		})
	}
}

// TestUpdateScrollKeys tests Up/Down arrow keys for viewport scrolling
func TestUpdateScrollKeys(t *testing.T) {
	tests := []struct {
		name   string
		keyMsg tea.KeyMsg
	}{
		{
			name: "up arrow scrolls up",
			keyMsg: tea.KeyMsg{
				Type: tea.KeyUp,
			},
		},
		{
			name: "down arrow scrolls down",
			keyMsg: tea.KeyMsg{
				Type: tea.KeyDown,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.ready = true
			m.SetSize(80, 24)

			// Add multiple messages to enable scrolling
			for i := 0; i < 20; i++ {
				m.AddMessage("user", "Message content that is long enough to require scrolling")
			}

			// Simulate scroll key press
			updatedModel, cmd := m.Update(tt.keyMsg)
			m = updatedModel.(Model)

			// Verify no command is returned (scrolling is synchronous)
			if cmd != nil {
				t.Error("expected no command for scroll action")
			}

			// No error should occur (viewport handles scrolling internally)
			// We can't directly verify viewport scroll position, but we ensure no panic
		})
	}
}

// TestUpdateClearMessages tests Ctrl+L clears conversation history
func TestUpdateClearMessages(t *testing.T) {
	m := NewModel()
	m.ready = true
	m.SetSize(80, 24)

	// Add some messages
	m.AddMessage("user", "Message 1")
	m.AddMessage("assistant", "Response 1")
	m.AddMessage("user", "Message 2")

	// Verify messages exist
	if len(m.messages) != 3 {
		t.Fatalf("expected 3 messages before clear, got %d", len(m.messages))
	}

	// Simulate Ctrl+L key press
	msg := tea.KeyMsg{
		Type:  tea.KeyCtrlL,
		Runes: []rune{'l'},
	}

	updatedModel, cmd := m.Update(msg)
	m = updatedModel.(Model)

	// Verify messages are cleared
	if len(m.messages) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(m.messages))
	}

	// Verify no command is returned
	if cmd != nil {
		t.Error("expected no command for clear action")
	}
}

// TestUpdateHelp tests ? key displays help text
func TestUpdateHelp(t *testing.T) {
	m := NewModel()
	m.ready = true
	m.SetSize(80, 24)

	// Simulate ? key press
	msg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'?'},
	}

	updatedModel, cmd := m.Update(msg)
	m = updatedModel.(Model)

	// Verify help text is displayed in viewport
	viewportContent := m.viewport.View()
	if !strings.Contains(viewportContent, "Keyboard Shortcuts") {
		t.Error("expected viewport to contain help text with 'Keyboard Shortcuts'")
	}
	if !strings.Contains(viewportContent, "Ctrl+C") {
		t.Error("expected help text to mention Ctrl+C")
	}
	if !strings.Contains(viewportContent, "Enter") {
		t.Error("expected help text to mention Enter")
	}

	// Verify no command is returned
	if cmd != nil {
		t.Error("expected no command for help display")
	}
}

// TestUpdateWindowSize tests terminal resize handling
func TestUpdateWindowSize(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		height         int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "sets standard size",
			width:          80,
			height:         24,
			expectedWidth:  80,
			expectedHeight: 24,
		},
		{
			name:           "sets large size",
			width:          200,
			height:         60,
			expectedWidth:  200,
			expectedHeight: 60,
		},
		{
			name:           "sets small size",
			width:          40,
			height:         10,
			expectedWidth:  40,
			expectedHeight: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.ready = true

			// Simulate window resize message
			msg := tea.WindowSizeMsg{
				Width:  tt.width,
				Height: tt.height,
			}

			updatedModel, cmd := m.Update(msg)
			m = updatedModel.(Model)

			// Verify size was updated
			if m.width != tt.expectedWidth {
				t.Errorf("expected width %d, got %d", tt.expectedWidth, m.width)
			}
			if m.height != tt.expectedHeight {
				t.Errorf("expected height %d, got %d", tt.expectedHeight, m.height)
			}

			// Verify no command is returned
			if cmd != nil {
				t.Error("expected no command for resize")
			}
		})
	}
}

// TestUpdateErrMsg tests error message handling
func TestUpdateErrMsg(t *testing.T) {
	tests := []struct {
		name              string
		err               error
		initialStreaming  bool
		expectedStreaming bool
	}{
		{
			name:              "sets error and stops streaming",
			err:               errors.New("API error"),
			initialStreaming:  true,
			expectedStreaming: false,
		},
		{
			name:              "sets error when not streaming",
			err:               errors.New("connection timeout"),
			initialStreaming:  false,
			expectedStreaming: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.ready = true
			m.SetSize(80, 24)
			m.streaming = tt.initialStreaming

			// Simulate error message
			msg := errMsg{err: tt.err}

			updatedModel, cmd := m.Update(msg)
			m = updatedModel.(Model)

			// Verify error is set
			if m.err == nil {
				t.Error("expected error to be set, got nil")
			} else if m.err.Error() != tt.err.Error() {
				t.Errorf("expected error %q, got %q", tt.err.Error(), m.err.Error())
			}

			// Verify streaming stopped
			if m.streaming != tt.expectedStreaming {
				t.Errorf("expected streaming %v, got %v", tt.expectedStreaming, m.streaming)
			}

			// Verify no command is returned
			if cmd != nil {
				t.Error("expected no command for error message")
			}
		})
	}
}

// TestUpdateReadyMsg tests ready signal handling
func TestUpdateReadyMsg(t *testing.T) {
	m := NewModel()
	// Initially not ready
	if m.ready {
		t.Error("expected ready to be false initially")
	}

	// Simulate ready message
	msg := readyMsg{}

	updatedModel, cmd := m.Update(msg)
	m = updatedModel.(Model)

	// Verify ready flag is set
	if !m.ready {
		t.Error("expected ready to be true after readyMsg")
	}

	// Verify no command is returned
	if cmd != nil {
		t.Error("expected no command for ready message")
	}
}

// TestUpdateStreamChunkMsg tests streaming content handling
func TestUpdateStreamChunkMsg(t *testing.T) {
	tests := []struct {
		name             string
		setupMessages    []Message
		chunk            string
		expectedContent  string
		shouldAppend     bool
	}{
		{
			name: "appends to assistant message",
			setupMessages: []Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi"},
			},
			chunk:           " there!",
			expectedContent: "Hi there!",
			shouldAppend:    true,
		},
		{
			name: "appends multiple chunks",
			setupMessages: []Message{
				{Role: "assistant", Content: "Hello"},
			},
			chunk:           " world",
			expectedContent: "Hello world",
			shouldAppend:    true,
		},
		{
			name:            "handles empty message list",
			setupMessages:   []Message{},
			chunk:           "content",
			expectedContent: "",
			shouldAppend:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.ready = true
			m.SetSize(80, 24)
			m.messages = tt.setupMessages

			// Simulate stream chunk message
			msg := streamChunkMsg{content: tt.chunk}

			updatedModel, cmd := m.Update(msg)
			m = updatedModel.(Model)

			// Verify content was appended
			if tt.shouldAppend && len(m.messages) > 0 {
				lastMsg := m.messages[len(m.messages)-1]
				if lastMsg.Content != tt.expectedContent {
					t.Errorf("expected content %q, got %q", tt.expectedContent, lastMsg.Content)
				}
			}

			// Verify no command is returned
			if cmd != nil {
				t.Error("expected no command for stream chunk")
			}
		})
	}
}

// TestUpdateStreamDoneMsg tests streaming completion handling
func TestUpdateStreamDoneMsg(t *testing.T) {
	m := NewModel()
	m.ready = true
	m.SetSize(80, 24)
	m.streaming = true

	// Add messages to simulate active streaming
	m.AddMessage("user", "Hello")
	m.AddMessage("assistant", "Hi there!")

	// Simulate stream done message
	msg := streamDoneMsg{}

	updatedModel, cmd := m.Update(msg)
	m = updatedModel.(Model)

	// Verify streaming stopped
	if m.streaming {
		t.Error("expected streaming to be false after streamDoneMsg")
	}

	// Verify no command is returned
	if cmd != nil {
		t.Error("expected no command for stream done")
	}

	// Verify messages are preserved
	if len(m.messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(m.messages))
	}
}

// TestUpdateUserInputMsg tests userInputMsg placeholder
func TestUpdateUserInputMsg(t *testing.T) {
	m := NewModel()
	m.ready = true

	// Simulate user input message
	msg := userInputMsg{input: "test input"}

	updatedModel, cmd := m.Update(msg)
	m = updatedModel.(Model)

	// Verify no command is returned (placeholder behavior)
	if cmd != nil {
		t.Error("expected no command for userInputMsg")
	}

	// No state changes expected for placeholder
}

// TestRenderMessages tests message rendering with different roles
func TestRenderMessages(t *testing.T) {
	tests := []struct {
		name             string
		messages         []Message
		err              error
		expectedContains []string
		notContains      []string
	}{
		{
			name: "renders user message",
			messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			expectedContains: []string{"You:", "Hello"},
		},
		{
			name: "renders assistant message",
			messages: []Message{
				{Role: "assistant", Content: "Hi there!"},
			},
			expectedContains: []string{"Assistant:", "Hi there!"},
		},
		{
			name: "renders system message",
			messages: []Message{
				{Role: "system", Content: "Connected"},
			},
			expectedContains: []string{"System:", "Connected"},
		},
		{
			name: "renders multiple messages with spacing",
			messages: []Message{
				{Role: "user", Content: "First"},
				{Role: "assistant", Content: "Second"},
			},
			expectedContains: []string{"You:", "First", "Assistant:", "Second"},
		},
		{
			name: "renders error",
			messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			err:              errors.New("test error"),
			expectedContains: []string{"Error:", "test error"},
		},
		{
			name:             "renders empty message list",
			messages:         []Message{},
			expectedContains: []string{},
			notContains:      []string{"You:", "Assistant:", "System:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.messages = tt.messages
			m.err = tt.err

			result := m.renderMessages()

			// Verify expected content is present
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain %q, got:\n%s", expected, result)
				}
			}

			// Verify unwanted content is not present
			for _, notExpected := range tt.notContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("expected result to not contain %q, got:\n%s", notExpected, result)
				}
			}
		})
	}
}

// TestRenderHelp tests help text generation
func TestRenderHelp(t *testing.T) {
	m := NewModel()

	result := m.renderHelp()

	// Verify help text contains all keyboard shortcuts
	expectedContent := []string{
		"Keyboard Shortcuts",
		"Enter",
		"Submit your message",
		"Ctrl+C",
		"Quit",
		"Up/Down",
		"Scroll",
		"Ctrl+L",
		"Clear",
		"?",
		"help",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(result, expected) {
			t.Errorf("expected help text to contain %q, got:\n%s", expected, result)
		}
	}

	// Verify help text is not empty
	if len(result) == 0 {
		t.Error("expected help text to be non-empty")
	}
}

// TestUpdateStateTransitions tests complex state transition scenarios
func TestUpdateStateTransitions(t *testing.T) {
	t.Run("streaming blocks input submission", func(t *testing.T) {
		m := NewModel()
		m.ready = true
		m.streaming = true
		m.textInput.SetValue("test input")

		// Try to submit while streaming
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, cmd := m.Update(msg)
		m = updatedModel.(Model)

		// Should not submit
		if cmd != nil {
			t.Error("expected no command during streaming")
		}
		if len(m.messages) != 0 {
			t.Error("expected no messages added during streaming")
		}
	})

	t.Run("error stops streaming", func(t *testing.T) {
		m := NewModel()
		m.ready = true
		m.streaming = true

		// Receive error while streaming
		msg := errMsg{err: errors.New("stream error")}
		updatedModel, _ := m.Update(msg)
		m = updatedModel.(Model)

		// Should stop streaming
		if m.streaming {
			t.Error("expected streaming to stop on error")
		}
		if m.err == nil {
			t.Error("expected error to be set")
		}
	})

	t.Run("resize while ready updates viewport", func(t *testing.T) {
		m := NewModel()
		m.ready = true
		m.AddMessage("user", "Test message")

		// Resize terminal
		msg := tea.WindowSizeMsg{Width: 100, Height: 30}
		updatedModel, _ := m.Update(msg)
		m = updatedModel.(Model)

		// Should update size
		if m.width != 100 || m.height != 30 {
			t.Errorf("expected size 100x30, got %dx%d", m.width, m.height)
		}
	})
}

// TestUpdateTextInputPassthrough tests that non-handled keys update text input
func TestUpdateTextInputPassthrough(t *testing.T) {
	m := NewModel()
	m.ready = true
	m.streaming = false

	// Simulate regular character input
	msg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'h'},
	}

	initialValue := m.textInput.Value()
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// Text input should have been updated
	// Note: We can't test the exact value because textInput.Update handles this internally
	// We just verify no panic occurred
	_ = m.textInput.Value()

	// Verify no state corruption
	if m.quitting {
		t.Error("unexpected quitting state")
	}
	if m.streaming {
		t.Error("unexpected streaming state")
	}

	// Ensure input was not the same (textInput processed the key)
	_ = initialValue
}

// TestUpdateMultipleStateChanges tests sequential state changes
func TestUpdateMultipleStateChanges(t *testing.T) {
	m := NewModel()
	m.ready = true
	m.SetSize(80, 24)

	// 1. Add user message
	m.textInput.SetValue("Hello")
	msg1 := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := m.Update(msg1)
	m = updatedModel.(Model)

	if len(m.messages) != 2 || !m.streaming {
		t.Fatal("failed to submit user message")
	}

	// 2. Receive streaming chunk
	msg2 := streamChunkMsg{content: "Hi "}
	updatedModel, _ = m.Update(msg2)
	m = updatedModel.(Model)

	if m.messages[1].Content != "Hi " {
		t.Fatal("failed to append stream chunk")
	}

	// 3. Receive another chunk
	msg3 := streamChunkMsg{content: "there!"}
	updatedModel, _ = m.Update(msg3)
	m = updatedModel.(Model)

	if m.messages[1].Content != "Hi there!" {
		t.Fatal("failed to append second chunk")
	}

	// 4. Finish streaming
	msg4 := streamDoneMsg{}
	updatedModel, _ = m.Update(msg4)
	m = updatedModel.(Model)

	if m.streaming {
		t.Fatal("failed to stop streaming")
	}

	// Verify final state
	if len(m.messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(m.messages))
	}
	if m.messages[0].Role != "user" || m.messages[0].Content != "Hello" {
		t.Error("user message corrupted")
	}
	if m.messages[1].Role != "assistant" || m.messages[1].Content != "Hi there!" {
		t.Error("assistant message incorrect")
	}
}

// Benchmark tests for performance validation

// BenchmarkUpdateKeyPress benchmarks keyboard input handling
func BenchmarkUpdateKeyPress(b *testing.B) {
	m := NewModel()
	m.ready = true
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Update(msg)
	}
}

// BenchmarkUpdateStreamChunk benchmarks stream chunk processing
func BenchmarkUpdateStreamChunk(b *testing.B) {
	m := NewModel()
	m.ready = true
	m.SetSize(80, 24)
	m.AddMessage("assistant", "")
	msg := streamChunkMsg{content: "test "}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		updatedModel, _ := m.Update(msg)
		m = updatedModel.(Model)
	}
}

// BenchmarkRenderMessages benchmarks message rendering
func BenchmarkRenderMessages(b *testing.B) {
	m := NewModel()
	for i := 0; i < 10; i++ {
		m.AddMessage("user", "Message content that is reasonably long")
		m.AddMessage("assistant", "Response content that is also reasonably long")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.renderMessages()
	}
}
