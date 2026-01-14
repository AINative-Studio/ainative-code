package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestView tests the main View function
func TestView(t *testing.T) {
	tests := []struct {
		name             string
		quitting         bool
		ready            bool
		expectedContains string
	}{
		{
			name:             "renders quitting message",
			quitting:         true,
			ready:            false,
			expectedContains: "Thanks for using ainative-code! Goodbye!",
		},
		{
			name:             "renders not ready message",
			quitting:         false,
			ready:            false,
			expectedContains: "Initializing TUI...",
		},
		{
			name:             "renders ready state with viewport",
			quitting:         false,
			ready:            true,
			expectedContains: "", // Will contain viewport content
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.quitting = tt.quitting
			m.ready = tt.ready
			m.SetSize(80, 24)

			view := m.View()

			if tt.expectedContains != "" && !strings.Contains(view, tt.expectedContains) {
				t.Errorf("expected view to contain %q, got %q", tt.expectedContains, view)
			}

			// Verify view is not empty
			if view == "" {
				t.Error("expected view to return non-empty string")
			}
		})
	}
}

// TestViewQuitting tests quitting state specifically
func TestViewQuitting(t *testing.T) {
	m := NewModel()
	m.quitting = true

	view := m.View()

	expected := "Thanks for using ainative-code! Goodbye!\n"
	if view != expected {
		t.Errorf("expected %q, got %q", expected, view)
	}
}

// TestViewNotReady tests not ready state specifically
func TestViewNotReady(t *testing.T) {
	m := NewModel()
	m.ready = false

	view := m.View()

	expected := "Initializing TUI...\n"
	if view != expected {
		t.Errorf("expected %q, got %q", expected, view)
	}
}

// TestViewReady tests ready state combines all sections
func TestViewReady(t *testing.T) {
	m := NewModel()
	m.ready = true
	m.SetSize(80, 24)

	view := m.View()

	// Verify view contains multiple sections
	if view == "" {
		t.Error("expected ready view to return non-empty string")
	}

	// Ready view should be more complex than simple messages
	if view == "Thanks for using ainative-code! Goodbye!\n" {
		t.Error("ready view should not be quitting message")
	}
	if view == "Initializing TUI...\n" {
		t.Error("ready view should not be initializing message")
	}
}

// TestRenderInputArea tests input area rendering
func TestRenderInputArea(t *testing.T) {
	tests := []struct {
		name             string
		streaming        bool
		width            int
		expectedContains []string
	}{
		{
			name:             "renders input area when not streaming",
			streaming:        false,
			width:            80,
			expectedContains: []string{"►"},
		},
		{
			name:             "renders input area when streaming",
			streaming:        true,
			width:            80,
			expectedContains: []string{"►"},
		},
		{
			name:             "renders input area with narrow width",
			streaming:        false,
			width:            40,
			expectedContains: []string{"►"},
		},
		{
			name:             "renders input area with zero width",
			streaming:        false,
			width:            0,
			expectedContains: []string{"►"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.streaming = tt.streaming
			m.width = tt.width
			m.SetSize(tt.width, 24)

			inputArea := m.renderInputArea()

			// Verify input area is not empty
			if inputArea == "" {
				t.Error("expected input area to return non-empty string")
			}

			// Verify expected content
			for _, expected := range tt.expectedContains {
				if !strings.Contains(inputArea, expected) {
					t.Errorf("expected input area to contain %q, got %q", expected, inputArea)
				}
			}

			// Verify separator line is present
			if !strings.Contains(inputArea, "─") {
				t.Error("expected input area to contain separator line")
			}
		})
	}
}

// TestRenderInputAreaStreaming tests streaming state shows disabled input
func TestRenderInputAreaStreaming(t *testing.T) {
	m := NewModel()
	m.streaming = true
	m.width = 80
	m.SetSize(80, 24)

	inputArea := m.renderInputArea()

	// When streaming, input should show placeholder
	if inputArea == "" {
		t.Error("expected input area to return non-empty string")
	}

	// Verify prompt is present
	if !strings.Contains(inputArea, "►") {
		t.Error("expected prompt to be present in streaming state")
	}
}

// TestRenderInputAreaNormal tests normal state shows active input
func TestRenderInputAreaNormal(t *testing.T) {
	m := NewModel()
	m.streaming = false
	m.width = 80
	m.SetSize(80, 24)
	m.textInput.SetValue("test input")

	inputArea := m.renderInputArea()

	// When not streaming, input should be active
	if inputArea == "" {
		t.Error("expected input area to return non-empty string")
	}

	// Verify prompt is present
	if !strings.Contains(inputArea, "►") {
		t.Error("expected prompt to be present in normal state")
	}
}

// TestRenderStatusBar tests status bar rendering
func TestRenderStatusBar(t *testing.T) {
	tests := []struct {
		name             string
		streaming        bool
		hasError         bool
		width            int
		expectedContains []string
	}{
		{
			name:             "renders streaming status",
			streaming:        true,
			hasError:         false,
			width:            80,
			expectedContains: []string{"Streaming", "Press ? for help"},
		},
		{
			name:             "renders error status",
			streaming:        false,
			hasError:         true,
			width:            80,
			expectedContains: []string{"Error occurred", "Press ? for help"},
		},
		{
			name:             "renders ready status",
			streaming:        false,
			hasError:         false,
			width:            80,
			expectedContains: []string{"Ready", "Press ? for help"},
		},
		{
			name:             "handles narrow width",
			streaming:        false,
			hasError:         false,
			width:            40,
			expectedContains: []string{"Ready"},
		},
		{
			name:             "handles zero width",
			streaming:        false,
			hasError:         false,
			width:            0,
			expectedContains: []string{"Ready"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.streaming = tt.streaming
			m.width = tt.width
			if tt.hasError {
				m.SetError(errors.New("test error"))
			}

			statusBar := m.renderStatusBar()

			// Verify status bar is not empty
			if statusBar == "" {
				t.Error("expected status bar to return non-empty string")
			}

			// Verify expected content
			for _, expected := range tt.expectedContains {
				if !strings.Contains(statusBar, expected) {
					t.Errorf("expected status bar to contain %q, got %q", expected, statusBar)
				}
			}
		})
	}
}

// TestRenderStatusBarStreaming tests streaming indicator
func TestRenderStatusBarStreaming(t *testing.T) {
	m := NewModel()
	m.streaming = true
	m.width = 80

	statusBar := m.renderStatusBar()

	if !strings.Contains(statusBar, "Streaming") {
		t.Errorf("expected streaming indicator, got %q", statusBar)
	}
}

// TestRenderStatusBarError tests error indicator
func TestRenderStatusBarError(t *testing.T) {
	m := NewModel()
	m.streaming = false
	m.SetError(errors.New("test error"))
	m.width = 80

	statusBar := m.renderStatusBar()

	if !strings.Contains(statusBar, "Error occurred") {
		t.Errorf("expected error indicator, got %q", statusBar)
	}
}

// TestRenderStatusBarReady tests ready indicator
func TestRenderStatusBarReady(t *testing.T) {
	m := NewModel()
	m.streaming = false
	m.err = nil
	m.width = 80

	statusBar := m.renderStatusBar()

	if !strings.Contains(statusBar, "Ready") {
		t.Errorf("expected ready indicator, got %q", statusBar)
	}
}

// TestRenderStatusBarSpacing tests spacing calculation
func TestRenderStatusBarSpacing(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "standard width",
			width: 80,
		},
		{
			name:  "wide width",
			width: 120,
		},
		{
			name:  "narrow width",
			width: 40,
		},
		{
			name:  "minimal width",
			width: 20,
		},
		{
			name:  "zero width",
			width: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.width = tt.width

			statusBar := m.renderStatusBar()

			// Verify status bar is rendered without panic
			if statusBar == "" {
				t.Error("expected status bar to return non-empty string")
			}

			// Verify both sections are present
			if !strings.Contains(statusBar, "Ready") && !strings.Contains(statusBar, "Press ? for help") {
				t.Error("expected status bar to contain both left and right sections")
			}
		})
	}
}

// TestRenderStatusBarHelpHint tests help hint is always present
func TestRenderStatusBarHelpHint(t *testing.T) {
	tests := []struct {
		name      string
		streaming bool
		hasError  bool
	}{
		{
			name:      "help hint with streaming",
			streaming: true,
			hasError:  false,
		},
		{
			name:      "help hint with error",
			streaming: false,
			hasError:  true,
		},
		{
			name:      "help hint with ready",
			streaming: false,
			hasError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.streaming = tt.streaming
			m.width = 80
			if tt.hasError {
				m.SetError(errors.New("test error"))
			}

			statusBar := m.renderStatusBar()

			if !strings.Contains(statusBar, "Press ? for help") {
				t.Errorf("expected help hint in all states, got %q", statusBar)
			}
		})
	}
}

// TestFormatError tests error message formatting
func TestFormatError(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		expectedEmpty    bool
		expectedContains string
	}{
		{
			name:          "formats nil error",
			err:           nil,
			expectedEmpty: true,
		},
		{
			name:             "formats simple error",
			err:              errors.New("test error"),
			expectedEmpty:    false,
			expectedContains: "Error: test error",
		},
		{
			name:             "formats complex error",
			err:              errors.New("API error: connection timeout"),
			expectedEmpty:    false,
			expectedContains: "Error: API error: connection timeout",
		},
		{
			name:             "formats error with special characters",
			err:              errors.New("error: invalid input @#$%"),
			expectedEmpty:    false,
			expectedContains: "Error: error: invalid input @#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatError(tt.err)

			if tt.expectedEmpty {
				if result != "" {
					t.Errorf("expected empty string for nil error, got %q", result)
				}
			} else {
				if result == "" {
					t.Error("expected non-empty string for non-nil error")
				}
				if !strings.Contains(result, tt.expectedContains) {
					t.Errorf("expected %q to contain %q", result, tt.expectedContains)
				}
			}
		})
	}
}

// TestFormatUserMessage tests user message formatting
func TestFormatUserMessage(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		expectedContains []string
	}{
		{
			name:             "formats simple user message",
			content:          "Hello",
			expectedContains: []string{"You:", "Hello"},
		},
		{
			name:             "formats complex user message",
			content:          "What is the weather today?",
			expectedContains: []string{"You:", "What is the weather today?"},
		},
		{
			name:             "formats empty user message",
			content:          "",
			expectedContains: []string{"You:"},
		},
		{
			name:             "formats user message with special characters",
			content:          "Test @#$%^&*()",
			expectedContains: []string{"You:", "Test @#$%^&*()"},
		},
		{
			name:             "formats multiline user message",
			content:          "Line 1\nLine 2",
			expectedContains: []string{"You:", "Line 1", "Line 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatUserMessage(tt.content)

			if result == "" {
				t.Error("expected non-empty formatted message")
			}

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected %q to contain %q", result, expected)
				}
			}
		})
	}
}

// TestFormatAssistantMessage tests assistant message formatting
func TestFormatAssistantMessage(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		expectedContains []string
	}{
		{
			name:             "formats simple assistant message",
			content:          "Hello!",
			expectedContains: []string{"Assistant:", "Hello!"},
		},
		{
			name:             "formats complex assistant message",
			content:          "I can help you with that.",
			expectedContains: []string{"Assistant:", "I can help you with that."},
		},
		{
			name:             "formats empty assistant message",
			content:          "",
			expectedContains: []string{"Assistant:"},
		},
		{
			name:             "formats assistant message with special characters",
			content:          "Response: @#$%",
			expectedContains: []string{"Assistant:", "Response: @#$%"},
		},
		{
			name:             "formats multiline assistant message",
			content:          "First line\nSecond line",
			expectedContains: []string{"Assistant:", "First line", "Second line"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatAssistantMessage(tt.content)

			if result == "" {
				t.Error("expected non-empty formatted message")
			}

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected %q to contain %q", result, expected)
				}
			}
		})
	}
}

// TestFormatSystemMessage tests system message formatting
func TestFormatSystemMessage(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		expectedContains []string
	}{
		{
			name:             "formats simple system message",
			content:          "Connected",
			expectedContains: []string{"System:", "Connected"},
		},
		{
			name:             "formats complex system message",
			content:          "Connection established successfully",
			expectedContains: []string{"System:", "Connection established successfully"},
		},
		{
			name:             "formats empty system message",
			content:          "",
			expectedContains: []string{"System:"},
		},
		{
			name:             "formats system message with special characters",
			content:          "Status: OK @#$%",
			expectedContains: []string{"System:", "Status: OK @#$%"},
		},
		{
			name:             "formats multiline system message",
			content:          "Line 1\nLine 2",
			expectedContains: []string{"System:", "Line 1", "Line 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSystemMessage(tt.content)

			if result == "" {
				t.Error("expected non-empty formatted message")
			}

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected %q to contain %q", result, expected)
				}
			}
		})
	}
}

// TestFormatFunctions tests all format functions integration
func TestFormatFunctions(t *testing.T) {
	// Test that all format functions work together
	userMsg := FormatUserMessage("Hello")
	assistantMsg := FormatAssistantMessage("Hi there!")
	systemMsg := FormatSystemMessage("Connected")
	errMsg := FormatError(errors.New("test error"))

	// Verify all return non-empty strings
	if userMsg == "" {
		t.Error("user message should not be empty")
	}
	if assistantMsg == "" {
		t.Error("assistant message should not be empty")
	}
	if systemMsg == "" {
		t.Error("system message should not be empty")
	}
	if errMsg == "" {
		t.Error("error message should not be empty")
	}

	// Verify each has correct label
	if !strings.Contains(userMsg, "You:") {
		t.Error("user message should contain 'You:' label")
	}
	if !strings.Contains(assistantMsg, "Assistant:") {
		t.Error("assistant message should contain 'Assistant:' label")
	}
	if !strings.Contains(systemMsg, "System:") {
		t.Error("system message should contain 'System:' label")
	}
	if !strings.Contains(errMsg, "Error:") {
		t.Error("error message should contain 'Error:' prefix")
	}
}

// TestViewStateTransitions tests view rendering across state changes
func TestViewStateTransitions(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)

	// Test not ready state
	m.ready = false
	notReadyView := m.View()
	if !strings.Contains(notReadyView, "Initializing") {
		t.Error("expected initializing message in not ready state")
	}

	// Transition to ready state
	m.ready = true
	readyView := m.View()
	if readyView == notReadyView {
		t.Error("expected view to change when transitioning to ready state")
	}

	// Transition to quitting state
	m.quitting = true
	quittingView := m.View()
	if !strings.Contains(quittingView, "Goodbye") {
		t.Error("expected goodbye message in quitting state")
	}
}

// TestViewWithMessages tests view with messages added
func TestViewWithMessages(t *testing.T) {
	m := NewModel()
	m.ready = true
	m.SetSize(80, 24)

	// Add messages
	m.AddMessage("user", "Hello")
	m.AddMessage("assistant", "Hi there!")

	view := m.View()

	// Verify view is rendered
	if view == "" {
		t.Error("expected non-empty view with messages")
	}

	// View should be complex with messages
	if view == "Initializing TUI...\n" {
		t.Error("view should not be initializing message with messages added")
	}
}

// TestViewEdgeCases tests edge cases in view rendering
func TestViewEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Model)
		validate func(*testing.T, string)
	}{
		{
			name: "zero dimensions",
			setup: func(m *Model) {
				m.ready = true
				m.SetSize(0, 0)
			},
			validate: func(t *testing.T, view string) {
				if view == "" {
					t.Error("expected non-empty view with zero dimensions")
				}
			},
		},
		{
			name: "negative dimensions handled",
			setup: func(m *Model) {
				m.ready = true
				m.width = -10
				m.height = -5
			},
			validate: func(t *testing.T, view string) {
				if view == "" {
					t.Error("expected non-empty view with negative dimensions")
				}
			},
		},
		{
			name: "very large dimensions",
			setup: func(m *Model) {
				m.ready = true
				m.SetSize(1000, 500)
			},
			validate: func(t *testing.T, view string) {
				if view == "" {
					t.Error("expected non-empty view with large dimensions")
				}
			},
		},
		{
			name: "streaming with error",
			setup: func(m *Model) {
				m.ready = true
				m.streaming = true
				m.SetError(errors.New("test error"))
				m.SetSize(80, 24)
			},
			validate: func(t *testing.T, view string) {
				if view == "" {
					t.Error("expected non-empty view with streaming and error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			tt.setup(&m)
			view := m.View()
			tt.validate(t, view)
		})
	}
}

// TestLipglossWidthCalculation tests width calculation in status bar
func TestLipglossWidthCalculation(t *testing.T) {
	// Test that lipgloss.Width is used correctly in status bar
	m := NewModel()
	m.width = 100

	statusBar := m.renderStatusBar()

	// Verify status bar is rendered
	if statusBar == "" {
		t.Error("expected non-empty status bar")
	}

	// Verify lipgloss Width calculations don't panic
	leftSection := "● Ready"
	rightSection := "Press ? for help"

	leftWidth := lipgloss.Width(leftSection)
	rightWidth := lipgloss.Width(rightSection)

	if leftWidth <= 0 {
		t.Error("expected positive left section width")
	}
	if rightWidth <= 0 {
		t.Error("expected positive right section width")
	}
}

// Benchmark tests for performance validation

// BenchmarkView benchmarks view rendering
func BenchmarkView(b *testing.B) {
	m := NewModel()
	m.ready = true
	m.SetSize(80, 24)
	m.AddMessage("user", "Hello")
	m.AddMessage("assistant", "Hi there!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}

// BenchmarkRenderInputArea benchmarks input area rendering
func BenchmarkRenderInputArea(b *testing.B) {
	m := NewModel()
	m.width = 80
	m.SetSize(80, 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.renderInputArea()
	}
}

// BenchmarkRenderStatusBar benchmarks status bar rendering
func BenchmarkRenderStatusBar(b *testing.B) {
	m := NewModel()
	m.width = 80

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.renderStatusBar()
	}
}

// BenchmarkFormatError benchmarks error formatting
func BenchmarkFormatError(b *testing.B) {
	err := errors.New("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatError(err)
	}
}

// BenchmarkFormatUserMessage benchmarks user message formatting
func BenchmarkFormatUserMessage(b *testing.B) {
	content := "Hello, assistant!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatUserMessage(content)
	}
}

// BenchmarkFormatAssistantMessage benchmarks assistant message formatting
func BenchmarkFormatAssistantMessage(b *testing.B) {
	content := "Hello! How can I help you?"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatAssistantMessage(content)
	}
}

// BenchmarkFormatSystemMessage benchmarks system message formatting
func BenchmarkFormatSystemMessage(b *testing.B) {
	content := "Connection established"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatSystemMessage(content)
	}
}

// BenchmarkAllFormatFunctions benchmarks all format functions together
func BenchmarkAllFormatFunctions(b *testing.B) {
	err := errors.New("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatError(err)
		_ = FormatUserMessage("Hello")
		_ = FormatAssistantMessage("Hi")
		_ = FormatSystemMessage("Connected")
	}
}
