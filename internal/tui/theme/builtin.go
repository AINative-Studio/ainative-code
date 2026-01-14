package theme

import "github.com/charmbracelet/lipgloss"

// AINativeTheme returns the branded AINative purple theme (DEFAULT)
// This theme showcases the AINative brand with purple accents and dark background
func AINativeTheme() *Theme {
	colors := ColorPalette{
		// Base colors - Deep dark background with light foreground
		Background: lipgloss.Color("#0f0f1a"),
		Foreground: lipgloss.Color("#e0e0ff"),

		// Semantic colors - AINative purple palette
		Primary:   lipgloss.Color("#8b5cf6"), // AINative purple
		Secondary: lipgloss.Color("#a78bfa"), // Light purple
		Accent:    lipgloss.Color("#c4b5fd"), // Lighter purple accent

		// Status colors
		Success: lipgloss.Color("#10b981"), // Green
		Warning: lipgloss.Color("#f59e0b"), // Amber
		Error:   lipgloss.Color("#ef4444"), // Red
		Info:    lipgloss.Color("#3b82f6"), // Blue

		// UI element colors
		Border:    lipgloss.Color("#8b5cf6"), // Purple borders
		Selection: lipgloss.Color("#8b5cf6"), // Purple selection
		Cursor:    lipgloss.Color("#a78bfa"), // Light purple cursor
		Highlight: lipgloss.Color("#c4b5fd"), // Lighter purple highlight
		Muted:     lipgloss.Color("#9ca3af"), // Gray
		Disabled:  lipgloss.Color("#6b7280"), // Dark gray

		// Component-specific
		StatusBar:      lipgloss.Color("#1a1a2e"), // Slightly lighter than background
		DialogBackdrop: lipgloss.Color("#00000099"), // Semi-transparent black
		ButtonActive:   lipgloss.Color("#8b5cf6"), // Purple
		ButtonInactive: lipgloss.Color("#2a2a3e"), // Dark gray
		InputBorder:    lipgloss.Color("#8b5cf6"), // Purple
		InputFocus:     lipgloss.Color("#a78bfa"), // Light purple

		// Code syntax highlighting - Purple theme
		CodeKeyword:  lipgloss.Color("#c678dd"), // Purple
		CodeString:   lipgloss.Color("#98c379"), // Green
		CodeComment:  lipgloss.Color("#5c6370"), // Gray
		CodeFunction: lipgloss.Color("#61afef"), // Blue
		CodeNumber:   lipgloss.Color("#d19a66"), // Orange
		CodeType:     lipgloss.Color("#e5c07b"), // Yellow
		CodeVariable: lipgloss.Color("#e06c75"), // Red
		CodeOperator: lipgloss.Color("#56b6c2"), // Cyan

		// Thinking blocks
		ThinkingBorder:     lipgloss.Color("#8b5cf6"), // Purple
		ThinkingBackground: lipgloss.Color("#1a1a2e"), // Slightly lighter
		ThinkingText:       lipgloss.Color("#d0d0e0"), // Muted white
		ThinkingHeader:     lipgloss.Color("#a78bfa"), // Light purple

		// Help system
		HelpTitle:    lipgloss.Color("#8b5cf6"), // Purple
		HelpCategory: lipgloss.Color("#a78bfa"), // Light purple
		HelpKey:      lipgloss.Color("#10b981"), // Green
		HelpDesc:     lipgloss.Color("#e0e0ff"), // Light foreground
		HelpHint:     lipgloss.Color("#6b7280"), // Dark gray
	}

	return NewTheme("AINative", colors)
}

// DarkTheme returns the default dark theme
// Classic dark theme with blue accents, inspired by popular terminals
func DarkTheme() *Theme {
	colors := ColorPalette{
		// Base colors - Terminal-style dark
		Background: lipgloss.Color("#1a1b26"), // Tokyo Night background
		Foreground: lipgloss.Color("#c0caf5"), // Light blue-white

		// Semantic colors - Blue palette
		Primary:   lipgloss.Color("#7aa2f7"), // Blue
		Secondary: lipgloss.Color("#9d7cd8"), // Purple
		Accent:    lipgloss.Color("#bb9af7"), // Light purple

		// Status colors
		Success: lipgloss.Color("#9ece6a"), // Green
		Warning: lipgloss.Color("#e0af68"), // Orange
		Error:   lipgloss.Color("#f7768e"), // Red
		Info:    lipgloss.Color("#7dcfff"), // Cyan

		// UI element colors
		Border:    lipgloss.Color("#3b4261"), // Muted blue-gray
		Selection: lipgloss.Color("#364a82"), // Dark blue
		Cursor:    lipgloss.Color("#7aa2f7"), // Blue
		Highlight: lipgloss.Color("#bb9af7"), // Purple
		Muted:     lipgloss.Color("#565f89"), // Gray
		Disabled:  lipgloss.Color("#3b4261"), // Dark gray

		// Component-specific
		StatusBar:      lipgloss.Color("#16161e"), // Darker background
		DialogBackdrop: lipgloss.Color("#00000099"), // Semi-transparent
		ButtonActive:   lipgloss.Color("#7aa2f7"), // Blue
		ButtonInactive: lipgloss.Color("#292e42"), // Dark gray
		InputBorder:    lipgloss.Color("#3b4261"), // Gray
		InputFocus:     lipgloss.Color("#7aa2f7"), // Blue

		// Code syntax highlighting - Tokyo Night colors
		CodeKeyword:  lipgloss.Color("#bb9af7"), // Purple
		CodeString:   lipgloss.Color("#9ece6a"), // Green
		CodeComment:  lipgloss.Color("#565f89"), // Gray
		CodeFunction: lipgloss.Color("#7aa2f7"), // Blue
		CodeNumber:   lipgloss.Color("#ff9e64"), // Orange
		CodeType:     lipgloss.Color("#2ac3de"), // Cyan
		CodeVariable: lipgloss.Color("#f7768e"), // Red
		CodeOperator: lipgloss.Color("#89ddff"), // Light blue

		// Thinking blocks
		ThinkingBorder:     lipgloss.Color("#3b4261"), // Gray
		ThinkingBackground: lipgloss.Color("#16161e"), // Dark
		ThinkingText:       lipgloss.Color("#a9b1d6"), // Muted blue-white
		ThinkingHeader:     lipgloss.Color("#7aa2f7"), // Blue

		// Help system
		HelpTitle:    lipgloss.Color("#7aa2f7"), // Blue
		HelpCategory: lipgloss.Color("#bb9af7"), // Purple
		HelpKey:      lipgloss.Color("#9ece6a"), // Green
		HelpDesc:     lipgloss.Color("#c0caf5"), // Foreground
		HelpHint:     lipgloss.Color("#565f89"), // Gray
	}

	return NewTheme("Dark", colors)
}

// LightTheme returns a clean light theme for daytime use
// Professional light theme with good contrast and readability
func LightTheme() *Theme {
	colors := ColorPalette{
		// Base colors - Clean white background
		Background: lipgloss.Color("#ffffff"),
		Foreground: lipgloss.Color("#24292f"), // Almost black

		// Semantic colors - Vibrant but professional
		Primary:   lipgloss.Color("#0969da"), // GitHub blue
		Secondary: lipgloss.Color("#8250df"), // Purple
		Accent:    lipgloss.Color("#bf3989"), // Magenta

		// Status colors
		Success: lipgloss.Color("#1a7f37"), // Green
		Warning: lipgloss.Color("#bf8700"), // Amber
		Error:   lipgloss.Color("#cf222e"), // Red
		Info:    lipgloss.Color("#0969da"), // Blue

		// UI element colors
		Border:    lipgloss.Color("#d0d7de"), // Light gray
		Selection: lipgloss.Color("#ddf4ff"), // Light blue
		Cursor:    lipgloss.Color("#0969da"), // Blue
		Highlight: lipgloss.Color("#fff8c5"), // Light yellow
		Muted:     lipgloss.Color("#656d76"), // Dark gray
		Disabled:  lipgloss.Color("#8c959f"), // Gray

		// Component-specific
		StatusBar:      lipgloss.Color("#f6f8fa"), // Light gray
		DialogBackdrop: lipgloss.Color("#00000040"), // Semi-transparent
		ButtonActive:   lipgloss.Color("#0969da"), // Blue
		ButtonInactive: lipgloss.Color("#eaeef2"), // Light gray
		InputBorder:    lipgloss.Color("#d0d7de"), // Light gray
		InputFocus:     lipgloss.Color("#0969da"), // Blue

		// Code syntax highlighting - GitHub light colors
		CodeKeyword:  lipgloss.Color("#cf222e"), // Red
		CodeString:   lipgloss.Color("#0a3069"), // Dark blue
		CodeComment:  lipgloss.Color("#6e7781"), // Gray
		CodeFunction: lipgloss.Color("#8250df"), // Purple
		CodeNumber:   lipgloss.Color("#0550ae"), // Blue
		CodeType:     lipgloss.Color("#953800"), // Brown
		CodeVariable: lipgloss.Color("#24292f"), // Black
		CodeOperator: lipgloss.Color("#cf222e"), // Red

		// Thinking blocks
		ThinkingBorder:     lipgloss.Color("#d0d7de"), // Light gray
		ThinkingBackground: lipgloss.Color("#f6f8fa"), // Very light gray
		ThinkingText:       lipgloss.Color("#24292f"), // Dark text
		ThinkingHeader:     lipgloss.Color("#0969da"), // Blue

		// Help system
		HelpTitle:    lipgloss.Color("#0969da"), // Blue
		HelpCategory: lipgloss.Color("#8250df"), // Purple
		HelpKey:      lipgloss.Color("#1a7f37"), // Green
		HelpDesc:     lipgloss.Color("#24292f"), // Foreground
		HelpHint:     lipgloss.Color("#656d76"), // Dark gray
	}

	return NewTheme("Light", colors)
}

// GetAllBuiltinThemes returns all built-in themes
func GetAllBuiltinThemes() []*Theme {
	return []*Theme{
		AINativeTheme(), // Default first
		DarkTheme(),
		LightTheme(),
	}
}

// RegisterBuiltinThemes registers all built-in themes with a theme manager
func RegisterBuiltinThemes(manager *ThemeManager) error {
	themes := GetAllBuiltinThemes()

	for _, theme := range themes {
		if err := manager.RegisterTheme(theme); err != nil {
			return err
		}
	}

	return nil
}
