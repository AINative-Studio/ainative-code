package dialogs

import (
	"github.com/charmbracelet/lipgloss"
)

// Dialog styles following AINative branding
var (
	// Dialog container styles
	DialogContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#8B5CF6")). // AINative purple
				Padding(1, 2).
				Background(lipgloss.Color("#1A1A2E")). // Dark background
				Foreground(lipgloss.Color("#E0E0E0"))   // Light text

	// Dialog title styles
	DialogTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#8B5CF6")). // AINative purple
				Align(lipgloss.Center).
				Width(40)

	// Dialog description styles
	DialogDescriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#B0B0B0")). // Muted text
				Align(lipgloss.Left).
				MarginTop(1).
				MarginBottom(1)

	// Button styles
	ButtonActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#8B5CF6")). // AINative purple
				Padding(0, 2).
				Margin(0, 1)

	ButtonInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#808080")).
				Background(lipgloss.Color("#2A2A3E")).
				Padding(0, 2).
				Margin(0, 1)

	// Input field styles
	InputFieldStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#8B5CF6")).
			Padding(0, 1).
			Width(36)

	InputFieldFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#A78BFA")). // Lighter purple when focused
				Padding(0, 1).
				Width(36)

	// List item styles for SelectDialog
	ListItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0E0E0")).
			Padding(0, 2)

	ListItemSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#8B5CF6")).
				Padding(0, 2)

	ListItemHoverStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A78BFA")).
				Padding(0, 2)

	// Error message styles
	ErrorTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")). // Red
			Bold(true).
			MarginTop(1)

	// Help text styles
	HelpTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")). // Gray
			Italic(true).
			MarginTop(1).
			Align(lipgloss.Center)

	// Success message styles
	SuccessTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")). // Green
				Bold(true)

	// Warning message styles
	WarningTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F59E0B")). // Orange
				Bold(true)
)

// RenderBackdrop creates a semi-transparent backdrop (deprecated - use BackdropRenderer)
func RenderBackdrop(width, height int) string {
	// Use the new BackdropRenderer with dark backdrop
	renderer := NewBackdropRenderer(width, height, DarkBackdrop)
	return renderer.Render()
}

// RenderDialogBox renders a dialog box at the center
func RenderDialogBox(content string, width int) string {
	style := DialogContainerStyle.Width(width)
	return style.Render(content)
}
