package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette for thinking visualization
var (
	// Thinking block colors (depth-based)
	ThinkingColor0 = lipgloss.Color("141") // Light purple
	ThinkingColor1 = lipgloss.Color("105") // Medium purple
	ThinkingColor2 = lipgloss.Color("99")  // Deep purple
	ThinkingColor3 = lipgloss.Color("63")  // Dark purple

	// UI element colors
	ThinkingBorderColor   = lipgloss.Color("240") // Gray border
	ThinkingHeaderColor   = lipgloss.Color("13")  // Magenta
	ThinkingTextColor     = lipgloss.Color("252") // Light gray
	ThinkingMutedColor    = lipgloss.Color("242") // Muted gray
	CollapsedIconColor    = lipgloss.Color("11")  // Yellow
	ExpandedIconColor     = lipgloss.Color("10")  // Green

	// Code highlighting colors
	CodeKeywordColor   = lipgloss.Color("12")  // Blue
	CodeStringColor    = lipgloss.Color("10")  // Green
	CodeCommentColor   = lipgloss.Color("241") // Gray
	CodeNumberColor    = lipgloss.Color("11")  // Yellow
	CodeFunctionColor  = lipgloss.Color("14")  // Cyan
)

// Thinking block styles

// ThinkingHeaderStyle styles the thinking block header
var ThinkingHeaderStyle = lipgloss.NewStyle().
	Foreground(ThinkingHeaderColor).
	Bold(true)

// ThinkingContentStyle styles the thinking block content
var ThinkingContentStyle = lipgloss.NewStyle().
	Foreground(ThinkingTextColor).
	PaddingLeft(2)

// ThinkingBorderStyle creates a border style for thinking blocks
func ThinkingBorderStyle(depth int) lipgloss.Style {
	// Cycle through colors based on depth
	colors := []lipgloss.Color{ThinkingColor0, ThinkingColor1, ThinkingColor2, ThinkingColor3}
	color := colors[depth%len(colors)]

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 1)
}

// ThinkingIndentStyle creates indentation for nested thinking
func ThinkingIndentStyle(depth int) lipgloss.Style {
	indent := depth * 2
	return lipgloss.NewStyle().
		PaddingLeft(indent)
}

// CollapsedIndicatorStyle styles the collapsed state indicator
var CollapsedIndicatorStyle = lipgloss.NewStyle().
	Foreground(CollapsedIconColor).
	Bold(true)

// ExpandedIndicatorStyle styles the expanded state indicator
var ExpandedIndicatorStyle = lipgloss.NewStyle().
	Foreground(ExpandedIconColor).
	Bold(true)

// ThinkingLabelStyle styles the "Thinking" label
var ThinkingLabelStyle = lipgloss.NewStyle().
	Foreground(ThinkingMutedColor).
	Italic(true)

// Code block styles

// CodeBlockStyle styles code blocks within thinking content
var CodeBlockStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("235")).
	Foreground(lipgloss.Color("252")).
	Padding(1).
	MarginTop(1).
	MarginBottom(1)

// InlineCodeStyle styles inline code snippets
var InlineCodeStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("236")).
	Foreground(CodeFunctionColor).
	Padding(0, 1)

// Helper functions for depth-based styling

// GetDepthColor returns the color for a specific depth level
func GetDepthColor(depth int) lipgloss.Color {
	colors := []lipgloss.Color{ThinkingColor0, ThinkingColor1, ThinkingColor2, ThinkingColor3}
	return colors[depth%len(colors)]
}

// GetDepthPrefix returns the tree-style prefix for a depth level
func GetDepthPrefix(depth int, isLast bool) string {
	if depth == 0 {
		return ""
	}

	prefix := ""
	for i := 0; i < depth-1; i++ {
		prefix += "│  "
	}

	if isLast {
		prefix += "└─ "
	} else {
		prefix += "├─ "
	}

	return prefix
}

// GetCollapsedIcon returns the icon for collapsed state
func GetCollapsedIcon() string {
	return "▶"
}

// GetExpandedIcon returns the icon for expanded state
func GetExpandedIcon() string {
	return "▼"
}

// GetThinkingIcon returns the icon for thinking blocks
func GetThinkingIcon() string {
	return "💭"
}
