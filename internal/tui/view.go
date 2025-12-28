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
		return "Thanks for using ainative-code! Goodbye!\n"
	}

	// Handle not ready state
	if !m.ready {
		return "Initializing TUI...\n"
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

	// Add separator line
	separator := strings.Repeat("─", m.width)
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(separator))
	sb.WriteString("\n")

	// Add prompt and input field
	prompt := inputPromptStyle.Render("►")
	sb.WriteString(prompt)
	sb.WriteString(" ")

	if m.streaming {
		// Show disabled state during streaming
		disabledInput := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(m.textInput.Placeholder)
		sb.WriteString(disabledInput)
	} else {
		// Show active input field
		sb.WriteString(m.textInput.View())
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
		leftSection = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Render("● Ready")
	}

	// Right section: Help hint
	rightSection = helpHintStyle.Render("Press ? for help")

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
