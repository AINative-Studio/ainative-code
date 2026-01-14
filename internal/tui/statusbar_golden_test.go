package tui

import (
	"testing"
	"time"
)

// TestStatusBarGolden_Compact tests the compact status bar for narrow terminals (width < 40)
func TestStatusBarGolden_Compact(t *testing.T) {
	state := NewStatusBarState()
	state.Provider = "Anthropic"
	state.Model = "Claude-3-Opus"
	state.SetMode("chat")
	state.SessionStart = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	state.ShowKeyHints = true

	view := state.RenderStatusBar(35, false, false)
	CheckGolden(t, "statusbar_compact", view)
}

// TestStatusBarGolden_CompactWithSpinner tests compact status bar with streaming indicator
func TestStatusBarGolden_CompactWithSpinner(t *testing.T) {
	state := NewStatusBarState()
	state.Provider = "Anthropic"
	state.Model = "Claude-3-Opus"
	state.SetMode("chat")
	state.AnimationTick = 0
	state.ShowKeyHints = true

	view := state.RenderStatusBar(35, true, false)
	CheckGolden(t, "statusbar_compact_streaming", view)
}

// TestStatusBarGolden_Basic tests the basic status bar (width 40-80)
func TestStatusBarGolden_Basic(t *testing.T) {
	state := NewStatusBarState()
	state.Provider = "Anthropic"
	state.Model = "Claude-3-Opus"
	state.SetMode("chat")
	state.SessionStart = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	state.ShowKeyHints = true

	view := state.RenderStatusBar(60, false, false)
	CheckGolden(t, "statusbar_basic", view)
}

// TestStatusBarGolden_BasicWithTokens tests basic status bar with token information
func TestStatusBarGolden_BasicWithTokens(t *testing.T) {
	state := NewStatusBarState()
	state.Provider = "Anthropic"
	state.Model = "Claude-3-Opus"
	state.SetTokens(5000, 100000)
	state.SetMode("chat")
	state.SessionStart = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	state.ShowKeyHints = true

	view := state.RenderStatusBar(60, false, false)
	CheckGolden(t, "statusbar_basic_tokens", view)
}

// TestStatusBarGolden_Full tests the full status bar (width 80-100)
func TestStatusBarGolden_Full(t *testing.T) {
	state := NewStatusBarState()
	state.Provider = "Anthropic"
	state.Model = "Claude-3-Opus"
	state.SetTokens(15000, 100000)
	state.SetMode("chat")
	state.SessionStart = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	state.ShowKeyHints = true

	view := state.RenderStatusBar(85, false, false)
	CheckGolden(t, "statusbar_full", view)
}

// TestStatusBarGolden_FullStreaming tests full status bar with streaming indicator
func TestStatusBarGolden_FullStreaming(t *testing.T) {
	state := NewStatusBarState()
	state.Provider = "Anthropic"
	state.Model = "Claude-3-Opus"
	state.SetTokens(15000, 100000)
	state.SetMode("chat")
	state.AnimationTick = 0
	state.SessionStart = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	state.ShowKeyHints = true

	view := state.RenderStatusBar(85, true, false)
	CheckGolden(t, "statusbar_full_streaming", view)
}

// TestStatusBarGolden_Extended tests the extended status bar (width 100+)
func TestStatusBarGolden_Extended(t *testing.T) {
	state := NewStatusBarState()
	state.Provider = "Anthropic"
	state.Model = "Claude-3-Opus"
	state.SetTokens(15000, 100000)
	state.SetMode("chat")
	// Use a fixed time far in the past to ensure consistent session duration
	state.SessionStart = time.Now().Add(-30 * time.Minute)
	state.ShowKeyHints = true

	view := state.RenderStatusBar(120, false, false)
	CheckGolden(t, "statusbar_extended", view)
}

// TestStatusBarGolden_Error tests status bar with error state
func TestStatusBarGolden_Error(t *testing.T) {
	state := NewStatusBarState()
	state.Provider = "Anthropic"
	state.Model = "Claude-3-Opus"
	state.SetMode("chat")
	state.SessionStart = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	state.ShowKeyHints = true

	view := state.RenderStatusBar(85, false, true)
	CheckGolden(t, "statusbar_error", view)
}

// TestStatusBarGolden_Disconnected tests status bar with disconnected state
func TestStatusBarGolden_Disconnected(t *testing.T) {
	state := NewStatusBarState()
	state.Provider = "Anthropic"
	state.Model = "Claude-3-Opus"
	state.SetConnectionStatus(false)
	state.SetMode("chat")
	state.SessionStart = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	state.ShowKeyHints = true

	view := state.RenderStatusBar(85, false, false)
	CheckGolden(t, "statusbar_disconnected", view)
}

// TestStatusBarGolden_HelpMode tests status bar in help mode
func TestStatusBarGolden_HelpMode(t *testing.T) {
	state := NewStatusBarState()
	state.Provider = "Anthropic"
	state.Model = "Claude-3-Opus"
	state.SetMode("help")
	state.SessionStart = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	state.ShowKeyHints = true

	view := state.RenderStatusBar(85, false, false)
	CheckGolden(t, "statusbar_help_mode", view)
}

// TestStatusBarGolden_CustomMessage tests status bar with custom message
func TestStatusBarGolden_CustomMessage(t *testing.T) {
	state := NewStatusBarState()
	state.Provider = "Anthropic"
	state.Model = "Claude-3-Opus"
	state.SetMode("chat")
	state.SetCustomMessage("Analyzing code...")
	state.SessionStart = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	state.ShowKeyHints = true

	view := state.RenderStatusBar(85, false, false)
	CheckGolden(t, "statusbar_custom_message", view)
}

// TestStatusBarGolden_HighTokenUsage tests status bar with high token usage (warning state)
func TestStatusBarGolden_HighTokenUsage(t *testing.T) {
	state := NewStatusBarState()
	state.Provider = "Anthropic"
	state.Model = "Claude-3-Opus"
	state.SetTokens(85000, 100000) // 85% usage - should show yellow warning
	state.SetMode("chat")
	state.SessionStart = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	state.ShowKeyHints = true

	view := state.RenderStatusBar(85, false, false)
	CheckGolden(t, "statusbar_high_tokens", view)
}

// TestStatusBarGolden_CriticalTokenUsage tests status bar with critical token usage (error state)
func TestStatusBarGolden_CriticalTokenUsage(t *testing.T) {
	state := NewStatusBarState()
	state.Provider = "Anthropic"
	state.Model = "Claude-3-Opus"
	state.SetTokens(97000, 100000) // 97% usage - should show red critical
	state.SetMode("chat")
	state.SessionStart = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	state.ShowKeyHints = true

	view := state.RenderStatusBar(85, false, false)
	CheckGolden(t, "statusbar_critical_tokens", view)
}
