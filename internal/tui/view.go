package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color scheme and styles
var (
	// Border styles
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))

	// Status bar styles
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("235"))

	streamingIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)

	helpHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// Input area styles
	inputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("12")).
				Bold(true)

	// Error styles
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	// Loading/Quitting styles
	centeredTextStyle = lipgloss.NewStyle().
				Align(lipgloss.Center)
)

// View renders the complete TUI interface
func (m Model) View() string {
	// Handle quitting state
	if m.quitting {
		quitStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true).
			Align(lipgloss.Center)
		return quitStyle.Render("Thanks for using ainative-code! Goodbye!\n")
	}

	// Handle not ready state
	if !m.ready {
		loadingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("14"))
		return loadingStyle.Render("Initializing TUI...\n")
	}

	var sb strings.Builder

	// 1. Viewport area (messages display)
	viewportContent := m.viewport.View()
	sb.WriteString(viewportContent)
	sb.WriteString("\n")

	// 2. Input area (text input field)
	inputSection := m.renderInputArea()
	sb.WriteString(inputSection)
	sb.WriteString("\n")

	// 3. Status bar (streaming indicator, help hint)
	statusBar := m.renderStatusBar()
	sb.WriteString(statusBar)

	return sb.String()
}

// renderInputArea creates the input section with prompt and text field
func (m *Model) renderInputArea() string {
	var sb strings.Builder

	// Add separator line with improved styling
	separator := strings.Repeat("─", m.width)
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sb.WriteString(separatorStyle.Render(separator))
	sb.WriteString("\n")

	// Add prompt and input field
	prompt := inputPromptStyle.Render("►")
	sb.WriteString(prompt)
	sb.WriteString(" ")

	if m.streaming {
		// Show disabled state during streaming with optional animation
		disabledStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
		sb.WriteString(disabledStyle.Render(m.textInput.Placeholder))
	} else {
		// Show active input field
		sb.WriteString(m.textInput.View())
	}

	// Add input hint for small terminals
	if m.width < 80 && !m.streaming {
		hintStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")).
			Italic(true)
		sb.WriteString(" ")
		sb.WriteString(hintStyle.Render("(Enter to send)"))
	}

	return sb.String()
}

// renderStatusBar creates the status bar with streaming indicator and help hint
func (m *Model) renderStatusBar() string {
	var leftSection, rightSection string

	// Left section: Streaming indicator or ready status
	if m.streaming {
		leftSection = streamingIndicatorStyle.Render("● Streaming...")
	} else if m.err != nil {
		leftSection = errorStyle.Render("✗ Error occurred")
	} else {
		readyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))
		leftSection = readyStyle.Render("● Ready")
	}

	// Right section: Help hint and thinking status
	rightParts := []string{}

	// Add thinking status if there are thinking blocks
	if len(m.thinkingState.Blocks) > 0 {
		thinkingStatus := ""
		if m.thinkingState.ShowThinking {
			thinkingStatus = "Thinking: ON"
		} else {
			thinkingStatus = "Thinking: OFF"
		}
		rightParts = append(rightParts, helpHintStyle.Render(thinkingStatus))
	}

	// Add scroll indicator for large content
	scrollIndicator := m.renderScrollIndicator()
	if scrollIndicator != "" {
		rightParts = append(rightParts, scrollIndicator)
	}

	rightParts = append(rightParts, helpHintStyle.Render("Press ? for help"))
	rightSection = strings.Join(rightParts, " | ")

	// Calculate spacing
	leftWidth := lipgloss.Width(leftSection)
	rightWidth := lipgloss.Width(rightSection)
	spacingWidth := m.width - leftWidth - rightWidth
	if spacingWidth < 0 {
		spacingWidth = 0
	}
	spacing := strings.Repeat(" ", spacingWidth)

	// Combine sections
	statusContent := leftSection + spacing + rightSection
	return statusBarStyle.Render(statusContent)
}

// renderScrollIndicator renders a scroll position indicator
func (m *Model) renderScrollIndicator() string {
	if m.viewport.TotalLineCount() == 0 {
		return ""
	}

	// Only show scroll indicator if there's content to scroll
	if m.viewport.TotalLineCount() <= m.viewport.Height {
		return ""
	}

	// Calculate scroll percentage
	scrollPercent := 0
	if m.viewport.TotalLineCount() > 0 {
		scrollPercent = int(float64(m.viewport.YOffset) / float64(m.viewport.TotalLineCount()) * 100)
	}

	indicatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Italic(true)

	var indicator string
	if scrollPercent == 0 {
		indicator = "↑ Top"
	} else if scrollPercent >= 100 {
		indicator = "↓ Bottom"
	} else {
		indicator = fmt.Sprintf("↕ %d%%", scrollPercent)
	}

	return indicatorStyle.Render(indicator)
}

// FormatError formats an error message for display
func FormatError(err error) string {
	if err == nil {
		return ""
	}
	return errorStyle.Render(fmt.Sprintf("Error: %s", err.Error()))
}

// FormatUserMessage formats a user message for display
func FormatUserMessage(content string) string {
	userStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true)

	label := userStyle.Render("You:")
	return fmt.Sprintf("%s %s", label, content)
}

// FormatAssistantMessage formats an assistant message for display
func FormatAssistantMessage(content string) string {
	assistantStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	label := assistantStyle.Render("Assistant:")
	return fmt.Sprintf("%s %s", label, content)
}

// FormatSystemMessage formats a system message for display
func FormatSystemMessage(content string) string {
	systemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Bold(true)

	label := systemStyle.Render("System:")
	return fmt.Sprintf("%s %s", label, content)
}

// renderCompactView renders a simplified view for very small terminals
func (m *Model) renderCompactView() string {
	if m.width < 40 || m.height < 10 {
		var sb strings.Builder

		// Show only the last message or a placeholder
		if len(m.messages) > 0 {
			lastMsg := m.messages[len(m.messages)-1]
			var formatted string
			switch lastMsg.Role {
			case "user":
				formatted = FormatUserMessage(lastMsg.Content)
			case "assistant":
				formatted = FormatAssistantMessage(lastMsg.Content)
			case "system":
				formatted = FormatSystemMessage(lastMsg.Content)
			default:
				formatted = lastMsg.Content
			}
			sb.WriteString(formatted)
		} else {
			placeholderStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("242")).
				Italic(true)
			sb.WriteString(placeholderStyle.Render("No messages yet"))
		}

		sb.WriteString("\n")

		// Minimal status
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		if m.streaming {
			sb.WriteString(statusStyle.Render("● Streaming..."))
		} else {
			sb.WriteString(statusStyle.Render("► Ready"))
		}

		return sb.String()
	}

	return m.View()
}
