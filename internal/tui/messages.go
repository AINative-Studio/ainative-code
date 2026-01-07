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

// thinkingChunkMsg represents a chunk of thinking content
type thinkingChunkMsg struct {
	content string
	depth   int
}

// thinkingDoneMsg signals the end of a thinking block
type thinkingDoneMsg struct{}

// toggleThinkingMsg signals to toggle thinking display
type toggleThinkingMsg struct{}

// collapseAllThinkingMsg signals to collapse all thinking blocks
type collapseAllThinkingMsg struct{}

// expandAllThinkingMsg signals to expand all thinking blocks
type expandAllThinkingMsg struct{}

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

// SendThinkingChunk creates a command that sends a thinking chunk
func SendThinkingChunk(content string, depth int) tea.Cmd {
	return func() tea.Msg {
		return thinkingChunkMsg{content: content, depth: depth}
	}
}

// SendThinkingDone creates a command that sends a thinking done message
func SendThinkingDone() tea.Cmd {
	return func() tea.Msg {
		return thinkingDoneMsg{}
	}
}

// SendToggleThinking creates a command that toggles thinking display
func SendToggleThinking() tea.Cmd {
	return func() tea.Msg {
		return toggleThinkingMsg{}
	}
}

// SendCollapseAllThinking creates a command to collapse all thinking
func SendCollapseAllThinking() tea.Cmd {
	return func() tea.Msg {
		return collapseAllThinkingMsg{}
	}
}

// SendExpandAllThinking creates a command to expand all thinking
func SendExpandAllThinking() tea.Cmd {
	return func() tea.Msg {
		return expandAllThinkingMsg{}
	}
}

// RLHF-related messages (TASK-064)

// feedbackPromptMsg signals to show the feedback prompt
type feedbackPromptMsg struct {
	interactionID string
}

// feedbackSubmittedMsg signals that feedback was submitted
type feedbackSubmittedMsg struct {
	interactionID string
	rating        int
	comment       string
}

// implicitFeedbackMsg records an implicit feedback action
type implicitFeedbackMsg struct {
	action string
}

// SendFeedbackPrompt creates a command to show feedback prompt
func SendFeedbackPrompt(interactionID string) tea.Cmd {
	return func() tea.Msg {
		return feedbackPromptMsg{interactionID: interactionID}
	}
}

// SendFeedbackSubmitted creates a command for submitted feedback
func SendFeedbackSubmitted(interactionID string, rating int, comment string) tea.Cmd {
	return func() tea.Msg {
		return feedbackSubmittedMsg{
			interactionID: interactionID,
			rating:        rating,
			comment:       comment,
		}
	}
}

// SendImplicitFeedback creates a command to record implicit feedback
func SendImplicitFeedback(action string) tea.Cmd {
	return func() tea.Msg {
		return implicitFeedbackMsg{action: action}
	}
}
