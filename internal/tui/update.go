package tui

import (
	"strings"

	"github.com/AINative-studio/ainative-code/internal/tui/theme"
	"github.com/AINative-studio/ainative-code/internal/tui/toast"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles state transitions based on incoming messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// Always update toast manager first (it handles its own messages)
	if m.toastManager != nil {
		toastModel, toastCmd := m.toastManager.Update(msg)
		m.toastManager = toastModel.(*toast.ToastManager)
		if toastCmd != nil {
			cmds = append(cmds, toastCmd)
		}
	}

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

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			// Up/k: Scroll viewport up (smooth scrolling)
			m.viewport.LineUp(1)
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			// Down/j: Scroll viewport down (smooth scrolling)
			m.viewport.LineDown(1)
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("pgup", "ctrl+u"))):
			// PageUp/Ctrl+U: Scroll half page up
			m.viewport.HalfViewUp()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("pgdown", "ctrl+d"))):
			// PageDown/Ctrl+D: Scroll half page down
			m.viewport.HalfViewDown()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("home", "g"))):
			// Home/g: Jump to top
			m.viewport.GotoTop()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("end", "G"))):
			// End/G: Jump to bottom
			m.viewport.GotoBottom()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+l"))):
			// Ctrl+L: Clear messages
			m.messages = []Message{}
			m.viewport.SetContent("")
			m.viewport.GotoTop()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("?"))):
			// ?: Toggle help display
			helpText := m.renderHelp()
			m.viewport.SetContent(helpText)
			m.viewport.GotoTop()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+h"))):
			// Ctrl+H: Toggle compact/full help
			helpText := m.renderHelp()
			m.viewport.SetContent(helpText)
			m.viewport.GotoTop()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+r"))):
			// Ctrl+R: Refresh display
			m.viewport.SetContent(m.renderMessages())
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("t"))):
			// t: Toggle thinking display
			m.ToggleThinkingDisplay()
			m.viewport.SetContent(m.renderMessages())
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("e"))):
			// e: Expand all thinking blocks
			m.ExpandAllThinking()
			m.viewport.SetContent(m.renderMessages())
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
			// c: Collapse all thinking blocks
			m.CollapseAllThinking()
			m.viewport.SetContent(m.renderMessages())
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+t"))):
			// Ctrl+T: Cycle through themes
			err := m.CycleTheme()
			if err == nil {
				// Update viewport with new theme colors
				m.viewport.SetContent(m.renderMessages())
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			// ESC: Close help or cancel input
			if strings.Contains(m.viewport.View(), "Keyboard Shortcuts") {
				// If we're showing help, go back to messages
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()
			} else {
				// Clear current input
				m.textInput.SetValue("")
			}
			return m, nil

		default:
			// Update text input for other key presses
			if !m.streaming {
				m.textInput, cmd = m.textInput.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case tea.WindowSizeMsg:
		// Handle terminal resize with improved responsive behavior
		m.SetSize(msg.Width, msg.Height)

		// Update viewport content after resize
		if m.ready {
			m.viewport.SetContent(m.renderMessages())
		}

		return m, nil

	case tea.MouseMsg:
		// Handle mouse events for scrolling
		switch msg.Type {
		case tea.MouseWheelUp:
			m.viewport.LineUp(3)
		case tea.MouseWheelDown:
			m.viewport.LineDown(3)
		}
		return m, nil

	case errMsg:
		// Handle error message with improved error display
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
		// Handle streaming content chunk with auto-scroll
		if len(m.messages) > 0 {
			// Append to the last message (should be assistant's response)
			lastIdx := len(m.messages) - 1
			if m.messages[lastIdx].Role == "assistant" {
				m.messages[lastIdx].Content += msg.content

				// Update viewport content
				m.viewport.SetContent(m.renderMessages())

				// Auto-scroll to bottom during streaming
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

	case thinkingChunkMsg:
		// Handle thinking content chunk
		if m.thinkingState.CurrentBlock == nil {
			// Create new thinking block if none exists
			m.AddThinking(msg.content, msg.depth)
		} else {
			// Append to current block
			m.AppendThinking(msg.content)
		}

		// Update viewport content
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

		return m, nil

	case thinkingDoneMsg:
		// Handle end of thinking block
		// Mark current block as complete (ready for new block)
		if m.thinkingState.CurrentBlock != nil {
			m.thinkingState.CurrentBlock = nil
		}

		// Update viewport content
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

		return m, nil

	case toggleThinkingMsg:
		// Toggle thinking display
		m.ToggleThinkingDisplay()
		m.viewport.SetContent(m.renderMessages())
		return m, nil

	case collapseAllThinkingMsg:
		// Collapse all thinking blocks
		m.CollapseAllThinking()
		m.viewport.SetContent(m.renderMessages())
		return m, nil

	case expandAllThinkingMsg:
		// Expand all thinking blocks
		m.ExpandAllThinking()
		m.viewport.SetContent(m.renderMessages())
		return m, nil

	case theme.SwitchThemeMsg:
		// Switch to specific theme
		err := m.SwitchTheme(msg.ThemeName)
		if err == nil {
			// Update viewport with new theme
			m.viewport.SetContent(m.renderMessages())
		}
		return m, nil

	case theme.CycleThemeMsg:
		// Cycle to next theme
		err := m.CycleTheme()
		if err == nil {
			// Update viewport with new theme
			m.viewport.SetContent(m.renderMessages())
		}
		return m, nil

	case theme.ThemeChangeMsg:
		// Theme changed, update all components
		m.viewport.SetContent(m.renderMessages())
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
		var content string
		if m.syntaxEnabled && m.syntaxHighlighter != nil {
			// Apply syntax highlighting to the content
			content = m.syntaxHighlighter.HighlightMarkdown(msg.Content)
		} else {
			content = msg.Content
		}

		switch msg.Role {
		case "user":
			sb.WriteString("You: ")
			sb.WriteString(content)
		case "assistant":
			sb.WriteString("Assistant: ")
			sb.WriteString(content)
		case "system":
			sb.WriteString("System: ")
			sb.WriteString(content)
		default:
			sb.WriteString(content)
		}
	}

	// Render thinking blocks if any exist
	if len(m.thinkingState.Blocks) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}

		// Add thinking header
		thinkingHeader := RenderThinkingHeader(m.thinkingState)
		if thinkingHeader != "" {
			sb.WriteString(thinkingHeader)
			sb.WriteString("\n")
		}

		// Add all thinking blocks
		thinkingContent := RenderAllThinkingBlocks(m.thinkingState, m.thinkingConfig)
		if thinkingContent != "" {
			sb.WriteString(thinkingContent)
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

Navigation:
  ↑/↓, k/j       Scroll up/down
  PgUp/PgDn      Page up/down
  Ctrl+U/Ctrl+D  Half page up/down
  Home/g         Jump to top
  End/G          Jump to bottom

Editing:
  Enter          Send message
  Ctrl+L         Clear conversation
  ESC            Cancel input/close help

View Options:
  t              Toggle thinking display
  e              Expand all thinking blocks
  c              Collapse all thinking blocks
  Ctrl+R         Refresh display

Help & System:
  ?              Show/hide this help
  Ctrl+H         Toggle help view
  Ctrl+C         Quit application

Press any key to return to the conversation.`
}
