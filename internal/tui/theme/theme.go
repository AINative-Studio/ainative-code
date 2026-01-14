package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines all colors, styles, and visual properties for the TUI
type Theme struct {
	Name    string
	Colors  ColorPalette
	Styles  StyleSet
	Borders BorderSet
	Spacing SpacingSet
}

// ColorPalette contains all semantic colors used throughout the application
type ColorPalette struct {
	// Base colors
	Background lipgloss.Color
	Foreground lipgloss.Color

	// Semantic colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color

	// Status colors
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
	Info    lipgloss.Color

	// UI element colors
	Border     lipgloss.Color
	Selection  lipgloss.Color
	Cursor     lipgloss.Color
	Highlight  lipgloss.Color
	Muted      lipgloss.Color
	Disabled   lipgloss.Color

	// Component-specific
	StatusBar      lipgloss.Color
	DialogBackdrop lipgloss.Color
	ButtonActive   lipgloss.Color
	ButtonInactive lipgloss.Color
	InputBorder    lipgloss.Color
	InputFocus     lipgloss.Color

	// Code syntax highlighting
	CodeKeyword  lipgloss.Color
	CodeString   lipgloss.Color
	CodeComment  lipgloss.Color
	CodeFunction lipgloss.Color
	CodeNumber   lipgloss.Color
	CodeType     lipgloss.Color
	CodeVariable lipgloss.Color
	CodeOperator lipgloss.Color

	// Thinking blocks
	ThinkingBorder     lipgloss.Color
	ThinkingBackground lipgloss.Color
	ThinkingText       lipgloss.Color
	ThinkingHeader     lipgloss.Color

	// Help system
	HelpTitle      lipgloss.Color
	HelpCategory   lipgloss.Color
	HelpKey        lipgloss.Color
	HelpDesc       lipgloss.Color
	HelpHint       lipgloss.Color
}

// StyleSet contains pre-built lipgloss.Style objects for common UI elements
type StyleSet struct {
	// Text styles
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Body     lipgloss.Style
	Code     lipgloss.Style
	Muted    lipgloss.Style
	Bold     lipgloss.Style
	Italic   lipgloss.Style

	// Button styles
	Button        lipgloss.Style
	ButtonFocused lipgloss.Style
	ButtonActive  lipgloss.Style

	// Status styles
	StatusBar lipgloss.Style
	Success   lipgloss.Style
	Warning   lipgloss.Style
	Error     lipgloss.Style
	Info      lipgloss.Style

	// Dialog styles
	Dialog           lipgloss.Style
	DialogTitle      lipgloss.Style
	DialogDesc       lipgloss.Style
	DialogBackdrop   lipgloss.Style
	InputField       lipgloss.Style
	InputFieldFocus  lipgloss.Style

	// List styles
	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style
	ListItemHover    lipgloss.Style

	// Thinking styles
	ThinkingBlock       lipgloss.Style
	ThinkingHeader      lipgloss.Style
	ThinkingCollapsed   lipgloss.Style
	ThinkingExpanded    lipgloss.Style

	// Help styles
	HelpBox      lipgloss.Style
	HelpTitle    lipgloss.Style
	HelpCategory lipgloss.Style
	HelpKey      lipgloss.Style
	HelpDesc     lipgloss.Style
}

// BorderSet contains border styles for different contexts
type BorderSet struct {
	Normal  lipgloss.Border
	Rounded lipgloss.Border
	Thick   lipgloss.Border
	Double  lipgloss.Border
	Hidden  lipgloss.Border
}

// SpacingSet contains spacing values for consistent layout
type SpacingSet struct {
	None   int
	Small  int
	Medium int
	Large  int
	XLarge int
}

// NewTheme creates a new theme with the given name and color palette
func NewTheme(name string, colors ColorPalette) *Theme {
	borders := buildBorderSet()
	spacing := buildSpacingSet()
	styles := buildStyleSet(colors, borders, spacing)

	return &Theme{
		Name:    name,
		Colors:  colors,
		Styles:  styles,
		Borders: borders,
		Spacing: spacing,
	}
}

// buildBorderSet creates the standard border set
func buildBorderSet() BorderSet {
	return BorderSet{
		Normal:  lipgloss.NormalBorder(),
		Rounded: lipgloss.RoundedBorder(),
		Thick:   lipgloss.ThickBorder(),
		Double:  lipgloss.DoubleBorder(),
		Hidden:  lipgloss.HiddenBorder(),
	}
}

// buildSpacingSet creates the standard spacing set
func buildSpacingSet() SpacingSet {
	return SpacingSet{
		None:   0,
		Small:  1,
		Medium: 2,
		Large:  4,
		XLarge: 8,
	}
}

// buildStyleSet creates all pre-built styles based on the color palette
func buildStyleSet(colors ColorPalette, borders BorderSet, spacing SpacingSet) StyleSet {
	return StyleSet{
		// Text styles
		Title: lipgloss.NewStyle().
			Foreground(colors.Primary).
			Bold(true).
			Padding(0, spacing.Small),

		Subtitle: lipgloss.NewStyle().
			Foreground(colors.Secondary).
			Bold(true),

		Body: lipgloss.NewStyle().
			Foreground(colors.Foreground),

		Code: lipgloss.NewStyle().
			Foreground(colors.CodeKeyword).
			Background(colors.ThinkingBackground).
			Padding(0, spacing.Small),

		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),

		Bold: lipgloss.NewStyle().
			Foreground(colors.Foreground).
			Bold(true),

		Italic: lipgloss.NewStyle().
			Foreground(colors.Foreground).
			Italic(true),

		// Button styles
		Button: lipgloss.NewStyle().
			Foreground(colors.ButtonInactive).
			Background(colors.Background).
			Padding(0, spacing.Medium).
			Margin(0, spacing.Small),

		ButtonFocused: lipgloss.NewStyle().
			Foreground(colors.Foreground).
			Background(colors.ButtonInactive).
			Padding(0, spacing.Medium).
			Margin(0, spacing.Small),

		ButtonActive: lipgloss.NewStyle().
			Foreground(colors.Background).
			Background(colors.ButtonActive).
			Bold(true).
			Padding(0, spacing.Medium).
			Margin(0, spacing.Small),

		// Status styles
		StatusBar: lipgloss.NewStyle().
			Foreground(colors.Foreground).
			Background(colors.StatusBar).
			Bold(true),

		Success: lipgloss.NewStyle().
			Foreground(colors.Success).
			Bold(true),

		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true),

		Error: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),

		Info: lipgloss.NewStyle().
			Foreground(colors.Info).
			Bold(true),

		// Dialog styles
		Dialog: lipgloss.NewStyle().
			Border(borders.Rounded).
			BorderForeground(colors.Border).
			Background(colors.Background).
			Foreground(colors.Foreground).
			Padding(spacing.Small, spacing.Medium),

		DialogTitle: lipgloss.NewStyle().
			Foreground(colors.Primary).
			Bold(true).
			Align(lipgloss.Center),

		DialogDesc: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Align(lipgloss.Left).
			MarginTop(spacing.Small).
			MarginBottom(spacing.Small),

		DialogBackdrop: lipgloss.NewStyle().
			Background(colors.DialogBackdrop),

		InputField: lipgloss.NewStyle().
			Border(borders.Normal).
			BorderForeground(colors.InputBorder).
			Padding(0, spacing.Small),

		InputFieldFocus: lipgloss.NewStyle().
			Border(borders.Normal).
			BorderForeground(colors.InputFocus).
			Padding(0, spacing.Small),

		// List styles
		ListItem: lipgloss.NewStyle().
			Foreground(colors.Foreground).
			Padding(0, spacing.Medium),

		ListItemSelected: lipgloss.NewStyle().
			Foreground(colors.Background).
			Background(colors.Selection).
			Bold(true).
			Padding(0, spacing.Medium),

		ListItemHover: lipgloss.NewStyle().
			Foreground(colors.Accent).
			Padding(0, spacing.Medium),

		// Thinking styles
		ThinkingBlock: lipgloss.NewStyle().
			Border(borders.Rounded).
			BorderForeground(colors.ThinkingBorder).
			Background(colors.ThinkingBackground).
			Foreground(colors.ThinkingText).
			Padding(spacing.Small, spacing.Medium).
			MarginTop(spacing.Small),

		ThinkingHeader: lipgloss.NewStyle().
			Foreground(colors.ThinkingHeader).
			Bold(true),

		ThinkingCollapsed: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),

		ThinkingExpanded: lipgloss.NewStyle().
			Foreground(colors.ThinkingText),

		// Help styles
		HelpBox: lipgloss.NewStyle().
			Border(borders.Rounded).
			BorderForeground(colors.Border).
			Background(colors.Background).
			Padding(spacing.Small, spacing.Medium),

		HelpTitle: lipgloss.NewStyle().
			Foreground(colors.HelpTitle).
			Bold(true).
			Align(lipgloss.Center),

		HelpCategory: lipgloss.NewStyle().
			Foreground(colors.HelpCategory).
			Bold(true).
			Underline(true),

		HelpKey: lipgloss.NewStyle().
			Foreground(colors.HelpKey).
			Bold(true),

		HelpDesc: lipgloss.NewStyle().
			Foreground(colors.HelpDesc),
	}
}

// Clone creates a deep copy of the theme
func (t *Theme) Clone() *Theme {
	return &Theme{
		Name:    t.Name,
		Colors:  t.Colors,
		Styles:  t.Styles,
		Borders: t.Borders,
		Spacing: t.Spacing,
	}
}

// GetName returns the theme name
func (t *Theme) GetName() string {
	return t.Name
}

// GetColor returns a color from the palette by name
func (t *Theme) GetColor(name string) lipgloss.Color {
	switch name {
	case "background":
		return t.Colors.Background
	case "foreground":
		return t.Colors.Foreground
	case "primary":
		return t.Colors.Primary
	case "secondary":
		return t.Colors.Secondary
	case "accent":
		return t.Colors.Accent
	case "success":
		return t.Colors.Success
	case "warning":
		return t.Colors.Warning
	case "error":
		return t.Colors.Error
	case "info":
		return t.Colors.Info
	case "border":
		return t.Colors.Border
	case "selection":
		return t.Colors.Selection
	default:
		return t.Colors.Foreground
	}
}

// Validate checks if the theme is valid and complete
func (t *Theme) Validate() error {
	if t.Name == "" {
		return ErrInvalidTheme{Reason: "theme name cannot be empty"}
	}

	// Check required colors are set
	if t.Colors.Background == "" {
		return ErrInvalidTheme{Reason: "background color not set"}
	}
	if t.Colors.Foreground == "" {
		return ErrInvalidTheme{Reason: "foreground color not set"}
	}
	if t.Colors.Primary == "" {
		return ErrInvalidTheme{Reason: "primary color not set"}
	}

	return nil
}

// ErrInvalidTheme represents an invalid theme error
type ErrInvalidTheme struct {
	Reason string
}

func (e ErrInvalidTheme) Error() string {
	return "invalid theme: " + e.Reason
}
