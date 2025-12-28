package tui

import tea "github.com/charmbracelet/bubbletea"

// Message types for Bubble Tea event handling

// errMsg represents an error event
type errMsg struct {
	err error
}

// Error returns the error
func (e errMsg) Error() string {
	return e.err.Error()
}

// readyMsg signals that the TUI is ready to display
type readyMsg struct{}

// streamChunkMsg represents a chunk of streamed content
type streamChunkMsg struct {
	content string
}

// streamDoneMsg signals the end of a streaming response
type streamDoneMsg struct{}

// userInputMsg represents user input submission
type userInputMsg struct {
	input string
}

// windowSizeMsg is handled by Bubble Tea automatically via tea.WindowSizeMsg

// Helper functions to create commands that return messages

// SendError creates a command that sends an error message
func SendError(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg{err: err}
	}
}

// SendReady creates a command that sends a ready message
func SendReady() tea.Cmd {
	return func() tea.Msg {
		return readyMsg{}
	}
}

// SendStreamChunk creates a command that sends a stream chunk
func SendStreamChunk(content string) tea.Cmd {
	return func() tea.Msg {
		return streamChunkMsg{content: content}
	}
}

// SendStreamDone creates a command that sends a stream done message
func SendStreamDone() tea.Cmd {
	return func() tea.Msg {
		return streamDoneMsg{}
	}
}

// SendUserInput creates a command that sends user input
func SendUserInput(input string) tea.Cmd {
	return func() tea.Msg {
		return userInputMsg{input: input}
	}
}
