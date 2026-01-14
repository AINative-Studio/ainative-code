package tui

import (
	"testing"
)

// TestHelpGolden_Compact tests the compact help view for small terminals
func TestHelpGolden_Compact(t *testing.T) {
	state := NewHelpState()
	state.Show()
	state.FullHelp = false

	view := state.RenderHelp(50, 15)
	CheckGolden(t, "help_compact", view)
}

// TestHelpGolden_CompactVerySmall tests compact help view for very small terminals
func TestHelpGolden_CompactVerySmall(t *testing.T) {
	state := NewHelpState()
	state.Show()
	state.FullHelp = false

	// Very small terminal - should trigger minimal compact mode
	view := state.RenderHelp(35, 8)
	CheckGolden(t, "help_compact_small", view)
}

// TestHelpGolden_Categorized tests the categorized help view (normal mode)
func TestHelpGolden_Categorized(t *testing.T) {
	state := NewHelpState()
	state.Show()
	state.FullHelp = false

	view := state.RenderHelp(80, 30)
	CheckGolden(t, "help_categorized", view)
}

// TestHelpGolden_Full tests the full help view with all details
func TestHelpGolden_Full(t *testing.T) {
	state := NewHelpState()
	state.Show()
	state.FullHelp = true

	view := state.RenderHelp(100, 40)
	CheckGolden(t, "help_full", view)
}

// TestHelpGolden_FullWide tests full help view on wide terminal
func TestHelpGolden_FullWide(t *testing.T) {
	state := NewHelpState()
	state.Show()
	state.FullHelp = true

	view := state.RenderHelp(120, 45)
	CheckGolden(t, "help_full_wide", view)
}

// TestHelpGolden_Hidden tests that hidden help returns empty string
func TestHelpGolden_Hidden(t *testing.T) {
	state := NewHelpState()
	state.Hide()

	view := state.RenderHelp(80, 30)

	// Hidden help should return empty string
	if view != "" {
		t.Errorf("Expected empty string for hidden help, got: %s", view)
	}
}

// TestHelpGolden_InlineHelp tests the inline help hints
func TestHelpGolden_InlineHelp(t *testing.T) {
	hints := []string{"Enter: send", "↑/↓: scroll", "?: help"}
	view := RenderInlineHelp(hints)
	CheckGolden(t, "help_inline", view)
}

// TestHelpGolden_ContextualChat tests contextual help for chat mode
func TestHelpGolden_ContextualChat(t *testing.T) {
	hints := GetContextualHelp("chat", false)
	view := RenderInlineHelp(hints)
	CheckGolden(t, "help_contextual_chat", view)
}

// TestHelpGolden_ContextualHelpMode tests contextual help for help mode
func TestHelpGolden_ContextualHelpMode(t *testing.T) {
	hints := GetContextualHelp("help", false)
	view := RenderInlineHelp(hints)
	CheckGolden(t, "help_contextual_help_mode", view)
}

// TestHelpGolden_ContextualStreaming tests contextual help during streaming
func TestHelpGolden_ContextualStreaming(t *testing.T) {
	hints := GetContextualHelp("chat", true)
	view := RenderInlineHelp(hints)
	CheckGolden(t, "help_contextual_streaming", view)
}
