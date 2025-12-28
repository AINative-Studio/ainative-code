package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles state transitions based on incoming messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle keyboard shortcuts
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
			// Ctrl+C: Quit the application
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// Enter: Submit user input
			if m.streaming {
				// Don't allow input submission while streaming
				return m, nil
			}

			input := strings.TrimSpace(m.textInput.Value())
			if input == "" {
				// Don't submit empty input
				return m, nil
			}

			// Clear the input field
			m.textInput.SetValue("")

			// Add user message to conversation
			m.messages = append(m.messages, Message{
				Role:    "user",
				Content: input,
			})

			// Update viewport content and scroll to bottom
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()

			// Start streaming state
			m.streaming = true

			// Add placeholder for assistant response
			m.messages = append(m.messages, Message{
				Role:    "assistant",
				Content: "",
			})

			// Return command to process user input
			return m, SendUserInput(input)

		case key.Matches(msg, key.NewBinding(key.WithKeys("up"))):
			// Up: Scroll viewport up
			m.viewport.LineUp(1)
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("down"))):
			// Down: Scroll viewport down
			m.viewport.LineDown(1)
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+l"))):
			// Ctrl+L: Clear messages
			m.messages = []Message{}
			m.viewport.SetContent("")
			m.viewport.GotoTop()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("?"))):
			// ?: Toggle help display
			// For now, add help message to viewport
			helpText := m.renderHelp()
			m.viewport.SetContent(helpText)
			m.viewport.GotoTop()
			return m, nil

		default:
			// Update text input for other key presses
			if !m.streaming {
				m.textInput, cmd = m.textInput.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case tea.WindowSizeMsg:
		// Handle terminal resize
		m.SetSize(msg.Width, msg.Height)

		// Update viewport content after resize
		if m.ready {
			m.viewport.SetContent(m.renderMessages())
		}

		return m, nil

	case errMsg:
		// Handle error message
		m.err = msg.err
		m.streaming = false

		// Update viewport to show error
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

		return m, nil

	case readyMsg:
		// Handle ready signal
		m.ready = true
		return m, nil

	case streamChunkMsg:
		// Handle streaming content chunk
		if len(m.messages) > 0 {
			// Append to the last message (should be assistant's response)
			lastIdx := len(m.messages) - 1
			if m.messages[lastIdx].Role == "assistant" {
				m.messages[lastIdx].Content += msg.content

				// Update viewport content
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()
			}
		}
		return m, nil

	case streamDoneMsg:
		// Handle end of streaming
		m.streaming = false

		// Update viewport content
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

		return m, nil

	case userInputMsg:
		// This message type is used to signal that user input should be processed
		// The actual processing happens in the command layer (commands.go)
		// This is just a placeholder for the command response
		return m, nil

	default:
		// Update viewport for other messages
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// renderMessages creates the viewport content from messages
func (m *Model) renderMessages() string {
	var sb strings.Builder

	for i, msg := range m.messages {
		if i > 0 {
			sb.WriteString("\n\n")
		}

		// Format based on role
		switch msg.Role {
		case "user":
			sb.WriteString("You: ")
			sb.WriteString(msg.Content)
		case "assistant":
			sb.WriteString("Assistant: ")
			sb.WriteString(msg.Content)
		case "system":
			sb.WriteString("System: ")
			sb.WriteString(msg.Content)
		default:
			sb.WriteString(msg.Content)
		}
	}

	// Show error if present
	if m.err != nil {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("Error: ")
		sb.WriteString(m.err.Error())
	}

	return sb.String()
}

// renderHelp creates help text with keyboard shortcuts
func (m *Model) renderHelp() string {
	return `Keyboard Shortcuts:
━━━━━━━━━━━━━━━━━━

Enter       Submit your message
Ctrl+C      Quit the application
Up/Down     Scroll through messages
Ctrl+L      Clear conversation history
?           Show this help

Press any key to return to the conversation.`
}
