package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

// TestGetDepthColor tests depth color cycling
func TestGetDepthColor(t *testing.T) {
	tests := []struct {
		name     string
		depth    int
		expected lipgloss.Color
	}{
		{
			name:     "depth 0",
			depth:    0,
			expected: ThinkingColor0,
		},
		{
			name:     "depth 1",
			depth:    1,
			expected: ThinkingColor1,
		},
		{
			name:     "depth 2",
			depth:    2,
			expected: ThinkingColor2,
		},
		{
			name:     "depth 3",
			depth:    3,
			expected: ThinkingColor3,
		},
		{
			name:     "depth 4 cycles to 0",
			depth:    4,
			expected: ThinkingColor0,
		},
		{
			name:     "depth 5 cycles to 1",
			depth:    5,
			expected: ThinkingColor1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := GetDepthColor(tt.depth)
			assert.Equal(t, tt.expected, color)
		})
	}
}

// TestGetDepthPrefix tests tree-style prefix generation
func TestGetDepthPrefix(t *testing.T) {
	tests := []struct {
		name     string
		depth    int
		isLast   bool
		expected string
	}{
		{
			name:     "depth 0",
			depth:    0,
			isLast:   false,
			expected: "",
		},
		{
			name:     "depth 1 not last",
			depth:    1,
			isLast:   false,
			expected: "├─ ",
		},
		{
			name:     "depth 1 last",
			depth:    1,
			isLast:   true,
			expected: "└─ ",
		},
		{
			name:     "depth 2 not last",
			depth:    2,
			isLast:   false,
			expected: "│  ├─ ",
		},
		{
			name:     "depth 2 last",
			depth:    2,
			isLast:   true,
			expected: "│  └─ ",
		},
		{
			name:     "depth 3 not last",
			depth:    3,
			isLast:   false,
			expected: "│  │  ├─ ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := GetDepthPrefix(tt.depth, tt.isLast)
			assert.Equal(t, tt.expected, prefix)
		})
	}
}

// TestGetCollapsedIcon tests collapsed icon
func TestGetCollapsedIcon(t *testing.T) {
	icon := GetCollapsedIcon()
	assert.NotEmpty(t, icon)
	assert.Equal(t, "▶", icon)
}

// TestGetExpandedIcon tests expanded icon
func TestGetExpandedIcon(t *testing.T) {
	icon := GetExpandedIcon()
	assert.NotEmpty(t, icon)
	assert.Equal(t, "▼", icon)
}

// TestGetThinkingIcon tests thinking icon
func TestGetThinkingIcon(t *testing.T) {
	icon := GetThinkingIcon()
	assert.NotEmpty(t, icon)
	assert.Equal(t, "💭", icon)
}

// TestThinkingBorderStyle tests border style creation
func TestThinkingBorderStyle(t *testing.T) {
	tests := []struct {
		name  string
		depth int
	}{
		{
			name:  "depth 0",
			depth: 0,
		},
		{
			name:  "depth 1",
			depth: 1,
		},
		{
			name:  "depth 5",
			depth: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := ThinkingBorderStyle(tt.depth)
			assert.NotNil(t, style)

			// Verify it can render something
			rendered := style.Render("test")
			assert.NotEmpty(t, rendered)
		})
	}
}

// TestThinkingIndentStyle tests indent style creation
func TestThinkingIndentStyle(t *testing.T) {
	tests := []struct {
		name  string
		depth int
	}{
		{
			name:  "depth 0",
			depth: 0,
		},
		{
			name:  "depth 1",
			depth: 1,
		},
		{
			name:  "depth 3",
			depth: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := ThinkingIndentStyle(tt.depth)
			assert.NotNil(t, style)

			// Verify it can render something
			rendered := style.Render("test")
			assert.NotEmpty(t, rendered)
		})
	}
}

// TestStylesRenderWithoutPanic tests that all styles can render without panicking
func TestStylesRenderWithoutPanic(t *testing.T) {
	testContent := "Test content"

	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"ThinkingHeaderStyle", ThinkingHeaderStyle},
		{"ThinkingContentStyle", ThinkingContentStyle},
		{"CollapsedIndicatorStyle", CollapsedIndicatorStyle},
		{"ExpandedIndicatorStyle", ExpandedIndicatorStyle},
		{"ThinkingLabelStyle", ThinkingLabelStyle},
		{"CodeBlockStyle", CodeBlockStyle},
		{"InlineCodeStyle", InlineCodeStyle},
	}

	for _, tt := range styles {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				rendered := tt.style.Render(testContent)
				assert.NotEmpty(t, rendered)
			})
		})
	}
}

// TestColorDefinitions tests that all colors are defined
func TestColorDefinitions(t *testing.T) {
	colors := []struct {
		name  string
		color lipgloss.Color
	}{
		{"ThinkingColor0", ThinkingColor0},
		{"ThinkingColor1", ThinkingColor1},
		{"ThinkingColor2", ThinkingColor2},
		{"ThinkingColor3", ThinkingColor3},
		{"ThinkingBorderColor", ThinkingBorderColor},
		{"ThinkingHeaderColor", ThinkingHeaderColor},
		{"ThinkingTextColor", ThinkingTextColor},
		{"ThinkingMutedColor", ThinkingMutedColor},
		{"CollapsedIconColor", CollapsedIconColor},
		{"ExpandedIconColor", ExpandedIconColor},
		{"CodeKeywordColor", CodeKeywordColor},
		{"CodeStringColor", CodeStringColor},
		{"CodeCommentColor", CodeCommentColor},
		{"CodeNumberColor", CodeNumberColor},
		{"CodeFunctionColor", CodeFunctionColor},
	}

	for _, tt := range colors {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, string(tt.color))
		})
	}
}
