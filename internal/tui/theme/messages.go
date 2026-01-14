package theme

import tea "github.com/charmbracelet/bubbletea"

// ThemeChangeMsg is sent when the theme changes
// Components should listen for this message and rebuild their styles
type ThemeChangeMsg struct {
	OldTheme *Theme
	NewTheme *Theme
}

// SwitchThemeMsg triggers a theme switch by name
type SwitchThemeMsg struct {
	ThemeName string
}

// CycleThemeMsg triggers cycling to the next theme
type CycleThemeMsg struct{}

// ThemeListMsg returns the list of available themes
type ThemeListMsg struct {
	Themes []string
}

// ThemeErrorMsg indicates a theme-related error
type ThemeErrorMsg struct {
	Err error
}

// SwitchTheme creates a command to switch to a specific theme
func SwitchTheme(name string) tea.Cmd {
	return func() tea.Msg {
		return SwitchThemeMsg{ThemeName: name}
	}
}

// CycleTheme creates a command to cycle to the next theme
func CycleTheme() tea.Cmd {
	return func() tea.Msg {
		return CycleThemeMsg{}
	}
}

// GetThemeList creates a command to get the list of available themes
func GetThemeList(manager *ThemeManager) tea.Cmd {
	return func() tea.Msg {
		return ThemeListMsg{
			Themes: manager.ListThemes(),
		}
	}
}

// NotifyThemeChange creates a theme change message
func NotifyThemeChange(oldTheme, newTheme *Theme) tea.Msg {
	return ThemeChangeMsg{
		OldTheme: oldTheme,
		NewTheme: newTheme,
	}
}

// ThemeError creates a theme error message
func ThemeError(err error) tea.Msg {
	return ThemeErrorMsg{Err: err}
}
