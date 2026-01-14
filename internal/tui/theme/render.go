package theme

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// RenderHelpers provides theme-aware rendering utilities
type RenderHelpers struct {
	theme *Theme
}

// NewRenderHelpers creates a new render helpers instance
func NewRenderHelpers(theme *Theme) *RenderHelpers {
	return &RenderHelpers{theme: theme}
}

// GetTheme returns the current theme
func (r *RenderHelpers) GetTheme() *Theme {
	return r.theme
}

// SetTheme updates the theme
func (r *RenderHelpers) SetTheme(theme *Theme) {
	r.theme = theme
}

// FormatError formats an error message using theme colors
func (r *RenderHelpers) FormatError(err error) string {
	if err == nil {
		return ""
	}
	return r.theme.Styles.Error.Render(fmt.Sprintf("Error: %s", err.Error()))
}

// FormatSuccess formats a success message using theme colors
func (r *RenderHelpers) FormatSuccess(message string) string {
	return r.theme.Styles.Success.Render(message)
}

// FormatWarning formats a warning message using theme colors
func (r *RenderHelpers) FormatWarning(message string) string {
	return r.theme.Styles.Warning.Render(message)
}

// FormatInfo formats an info message using theme colors
func (r *RenderHelpers) FormatInfo(message string) string {
	return r.theme.Styles.Info.Render(message)
}

// FormatUserMessage formats a user message using theme colors
func (r *RenderHelpers) FormatUserMessage(content string) string {
	userStyle := lipgloss.NewStyle().
		Foreground(r.theme.Colors.Primary).
		Bold(true)
	label := userStyle.Render("You:")
	return fmt.Sprintf("%s %s", label, content)
}

// FormatAssistantMessage formats an assistant message using theme colors
func (r *RenderHelpers) FormatAssistantMessage(content string) string {
	assistantStyle := lipgloss.NewStyle().
		Foreground(r.theme.Colors.Success).
		Bold(true)
	label := assistantStyle.Render("Assistant:")
	return fmt.Sprintf("%s %s", label, content)
}

// FormatSystemMessage formats a system message using theme colors
func (r *RenderHelpers) FormatSystemMessage(content string) string {
	systemStyle := lipgloss.NewStyle().
		Foreground(r.theme.Colors.Warning).
		Bold(true)
	label := systemStyle.Render("System:")
	return fmt.Sprintf("%s %s", label, content)
}

// BorderStyle returns a border style using theme colors
func (r *RenderHelpers) BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(r.theme.Borders.Rounded).
		BorderForeground(r.theme.Colors.Border)
}

// StatusBarStyle returns a status bar style using theme colors
func (r *RenderHelpers) StatusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Foreground).
		Background(r.theme.Colors.StatusBar)
}

// StreamingIndicatorStyle returns a streaming indicator style
func (r *RenderHelpers) StreamingIndicatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Success).
		Bold(true)
}

// HelpHintStyle returns a help hint style using theme colors
func (r *RenderHelpers) HelpHintStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Muted)
}

// InputPromptStyle returns an input prompt style using theme colors
func (r *RenderHelpers) InputPromptStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Primary).
		Bold(true)
}

// ErrorStyle returns an error style using theme colors
func (r *RenderHelpers) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Error).
		Bold(true)
}

// SeparatorStyle returns a separator style using theme colors
func (r *RenderHelpers) SeparatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Border)
}

// DisabledStyle returns a disabled style using theme colors
func (r *RenderHelpers) DisabledStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Disabled)
}

// ReadyStyle returns a ready indicator style
func (r *RenderHelpers) ReadyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Success)
}

// CenteredTextStyle returns a centered text style
func (r *RenderHelpers) CenteredTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center)
}

// QuitStyle returns a quit message style
func (r *RenderHelpers) QuitStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Success).
		Bold(true).
		Align(lipgloss.Center)
}

// LoadingStyle returns a loading message style
func (r *RenderHelpers) LoadingStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Info)
}

// ScrollIndicatorStyle returns a scroll indicator style
func (r *RenderHelpers) ScrollIndicatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Muted).
		Italic(true)
}

// PlaceholderStyle returns a placeholder text style
func (r *RenderHelpers) PlaceholderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Muted).
		Italic(true)
}

// LSPConnectedStyle returns an LSP connected indicator style
func (r *RenderHelpers) LSPConnectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Success)
}

// LSPConnectingStyle returns an LSP connecting indicator style
func (r *RenderHelpers) LSPConnectingStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Warning)
}

// LSPErrorStyle returns an LSP error indicator style
func (r *RenderHelpers) LSPErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Error)
}

// LSPDisconnectedStyle returns an LSP disconnected indicator style
func (r *RenderHelpers) LSPDisconnectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Disabled)
}

// ThemeIndicatorStyle returns a theme name indicator style
func (r *RenderHelpers) ThemeIndicatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(r.theme.Colors.Accent).
		Italic(true)
}

// FormatThemeIndicator formats the current theme name for display
func (r *RenderHelpers) FormatThemeIndicator() string {
	return r.ThemeIndicatorStyle().Render(fmt.Sprintf("[%s]", r.theme.Name))
}
