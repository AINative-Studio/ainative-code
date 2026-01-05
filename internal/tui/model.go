package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

// Model represents the TUI application state
type Model struct {
	viewport       viewport.Model
	textInput      textinput.Model
	messages       []Message
	thinkingState  *ThinkingState
	thinkingConfig ThinkingConfig
	width          int
	height         int
	ready          bool
	quitting       bool
	streaming      bool
	err            error
}

// Message represents a chat message
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

// NewModel creates a new TUI model
func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Type your message..."
	ti.Focus()
	ti.CharLimit = 0
	ti.Width = 50

	return Model{
		textInput:      ti,
		messages:       []Message{},
		thinkingState:  NewThinkingState(),
		thinkingConfig: DefaultThinkingConfig(),
		ready:          false,
		quitting:       false,
		streaming:      false,
	}
}

// SetSize updates the model dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	if !m.ready {
		// Initialize viewport with proper dimensions
		// Reserve space for input area (3 lines) and status bar (1 line)
		viewportHeight := height - 4
		if viewportHeight < 1 {
			viewportHeight = 1
		}

		m.viewport = viewport.New(width, viewportHeight)
		m.viewport.YPosition = 0
		m.ready = true
	} else {
		// Update existing viewport dimensions
		viewportHeight := height - 4
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		m.viewport.Width = width
		m.viewport.Height = viewportHeight
	}

	// Update text input width
	m.textInput.Width = width - 4
}

// AddMessage adds a new message to the conversation
func (m *Model) AddMessage(role, content string) {
	m.messages = append(m.messages, Message{
		Role:    role,
		Content: content,
	})
}

// ClearMessages removes all messages
func (m *Model) ClearMessages() {
	m.messages = []Message{}
}

// GetUserInput returns and clears the current input
func (m *Model) GetUserInput() string {
	input := m.textInput.Value()
	m.textInput.SetValue("")
	return input
}

// SetError sets an error state
func (m *Model) SetError(err error) {
	m.err = err
}

// ClearError clears the error state
func (m *Model) ClearError() {
	m.err = nil
}

// IsReady returns whether the TUI is ready to display
func (m *Model) IsReady() bool {
	return m.ready
}

// IsQuitting returns whether the TUI is quitting
func (m *Model) IsQuitting() bool {
	return m.quitting
}

// IsStreaming returns whether a response is being streamed
func (m *Model) IsStreaming() bool {
	return m.streaming
}

// SetStreaming sets the streaming state
func (m *Model) SetStreaming(streaming bool) {
	m.streaming = streaming
}

// SetQuitting sets the quitting state
func (m *Model) SetQuitting(quitting bool) {
	m.quitting = quitting
}

// Thinking-related methods

// ToggleThinkingDisplay toggles the display of thinking blocks
func (m *Model) ToggleThinkingDisplay() {
	m.thinkingState.ToggleDisplay()
}

// AddThinking adds a new thinking block
func (m *Model) AddThinking(content string, depth int) {
	m.thinkingState.AddThinkingBlock(content, depth)
}

// AppendThinking appends content to the current thinking block
func (m *Model) AppendThinking(content string) {
	m.thinkingState.AppendToCurrentBlock(content)
}

// CollapseAllThinking collapses all thinking blocks
func (m *Model) CollapseAllThinking() {
	m.thinkingState.CollapseAll()
}

// ExpandAllThinking expands all thinking blocks
func (m *Model) ExpandAllThinking() {
	m.thinkingState.ExpandAll()
}

// ClearThinking removes all thinking blocks
func (m *Model) ClearThinking() {
	m.thinkingState.ClearBlocks()
}

// IsThinkingVisible returns whether thinking blocks are visible
func (m *Model) IsThinkingVisible() bool {
	return m.thinkingState.ShowThinking
}

// GetThinkingState returns the thinking state
func (m *Model) GetThinkingState() *ThinkingState {
	return m.thinkingState
}

// GetThinkingConfig returns the thinking configuration
func (m *Model) GetThinkingConfig() ThinkingConfig {
	return m.thinkingConfig
}

// SetThinkingConfig sets the thinking configuration
func (m *Model) SetThinkingConfig(config ThinkingConfig) {
	m.thinkingConfig = config
}
