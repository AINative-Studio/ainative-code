package tui

import (
	"testing"
	"time"
)

// TestThinkingGolden_SingleBlockExpanded tests a single expanded thinking block
func TestThinkingGolden_SingleBlockExpanded(t *testing.T) {
	state := NewThinkingState()
	block := state.AddThinkingBlock("Let me analyze this problem step by step.\n\nFirst, I'll consider the requirements:\n- The function needs to handle edge cases\n- Performance is important\n- Code should be maintainable", 0)
	block.Timestamp = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	config := DefaultThinkingConfig()
	config.ShowTimestamps = false // Disable timestamps for deterministic output

	view := RenderThinkingBlock(block, config)
	CheckGolden(t, "thinking_single_expanded", view)
}

// TestThinkingGolden_SingleBlockCollapsed tests a single collapsed thinking block
func TestThinkingGolden_SingleBlockCollapsed(t *testing.T) {
	state := NewThinkingState()
	block := state.AddThinkingBlock("Let me analyze this problem step by step.\n\nFirst, I'll consider the requirements:\n- The function needs to handle edge cases\n- Performance is important\n- Code should be maintainable", 0)
	block.Collapsed = true
	block.Timestamp = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	config := DefaultThinkingConfig()
	config.ShowTimestamps = false

	view := RenderThinkingBlock(block, config)
	CheckGolden(t, "thinking_single_collapsed", view)
}

// TestThinkingGolden_MultipleBlocks tests multiple thinking blocks
func TestThinkingGolden_MultipleBlocks(t *testing.T) {
	state := NewThinkingState()

	// Add multiple blocks at different depths
	block1 := state.AddThinkingBlock("Initial analysis of the problem", 0)
	block1.Timestamp = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	block2 := state.AddThinkingBlock("Considering approach A: iterative solution", 1)
	block2.Timestamp = time.Date(2024, 1, 1, 12, 0, 5, 0, time.UTC)

	block3 := state.AddThinkingBlock("Considering approach B: recursive solution", 1)
	block3.Timestamp = time.Date(2024, 1, 1, 12, 0, 10, 0, time.UTC)

	config := DefaultThinkingConfig()
	config.ShowTimestamps = false

	view := RenderAllThinkingBlocks(state, config)
	CheckGolden(t, "thinking_multiple", view)
}

// TestThinkingGolden_NestedDepth tests thinking blocks with varying depth
func TestThinkingGolden_NestedDepth(t *testing.T) {
	state := NewThinkingState()

	// Create nested thinking structure
	block1 := state.AddThinkingBlock("Top-level analysis", 0)
	block2 := state.AddThinkingBlock("Nested consideration", 1)
	block3 := state.AddThinkingBlock("Deeply nested thought", 2)
	block4 := state.AddThinkingBlock("Even deeper analysis", 3)

	for _, block := range []*ThinkingBlock{block1, block2, block3, block4} {
		block.Timestamp = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	}

	config := DefaultThinkingConfig()
	config.ShowTimestamps = false

	view := RenderAllThinkingBlocks(state, config)
	CheckGolden(t, "thinking_nested_depth", view)
}

// TestThinkingGolden_WithCodeBlock tests thinking with code syntax
func TestThinkingGolden_WithCodeBlock(t *testing.T) {
	state := NewThinkingState()
	block := state.AddThinkingBlock("Here's a solution in Go:\n\n```go\nfunc Add(a, b int) int {\n    return a + b\n}\n```\n\nThis function is simple and efficient.", 0)
	block.Timestamp = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	config := DefaultThinkingConfig()
	config.ShowTimestamps = false
	config.SyntaxHighlighting = true

	view := RenderThinkingBlock(block, config)
	CheckGolden(t, "thinking_with_code", view)
}

// TestThinkingGolden_MixedCollapsed tests mixed collapsed and expanded blocks
func TestThinkingGolden_MixedCollapsed(t *testing.T) {
	state := NewThinkingState()

	block1 := state.AddThinkingBlock("First thought process", 0)
	block2 := state.AddThinkingBlock("Second thought process - this one is collapsed", 0)
	block2.Collapsed = true
	block3 := state.AddThinkingBlock("Third thought process", 0)

	for _, block := range []*ThinkingBlock{block1, block2, block3} {
		block.Timestamp = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	}

	config := DefaultThinkingConfig()
	config.ShowTimestamps = false

	view := RenderAllThinkingBlocks(state, config)
	CheckGolden(t, "thinking_mixed_collapsed", view)
}

// TestThinkingGolden_Header tests the thinking header rendering
func TestThinkingGolden_Header(t *testing.T) {
	state := NewThinkingState()

	// Add a few blocks for summary
	state.AddThinkingBlock("First block", 0)
	block2 := state.AddThinkingBlock("Second block", 0)
	block2.Collapsed = true
	state.AddThinkingBlock("Third block", 0)

	view := RenderThinkingHeader(state)
	CheckGolden(t, "thinking_header", view)
}

// TestThinkingGolden_ToggleHint tests the thinking toggle hint
func TestThinkingGolden_ToggleHintShow(t *testing.T) {
	view := RenderThinkingToggleHint(true)
	CheckGolden(t, "thinking_toggle_hint_show", view)
}

// TestThinkingGolden_ToggleHintHide tests the thinking toggle hint for hiding
func TestThinkingGolden_ToggleHintHide(t *testing.T) {
	view := RenderThinkingToggleHint(false)
	CheckGolden(t, "thinking_toggle_hint_hide", view)
}

// TestThinkingGolden_Summary tests the thinking summary rendering
func TestThinkingGolden_Summary(t *testing.T) {
	state := NewThinkingState()

	// Add blocks with varying content
	state.AddThinkingBlock("Short block", 0)
	block2 := state.AddThinkingBlock("Medium block with\nmultiple lines\nof content", 0)
	block2.Collapsed = true
	state.AddThinkingBlock("Another block", 0)

	view := RenderThinkingSummary(state)
	CheckGolden(t, "thinking_summary", view)
}

// TestThinkingGolden_AnimationIndicator tests the thinking animation indicator
func TestThinkingGolden_AnimationIndicator(t *testing.T) {
	// Test thinking indicator at tick 0
	view := ThinkingIndicator(0)
	CheckGolden(t, "thinking_indicator_tick0", view)
}

// TestThinkingGolden_AnimationIndicatorTick2 tests thinking indicator at different tick
func TestThinkingGolden_AnimationIndicatorTick2(t *testing.T) {
	// Test thinking indicator at tick 2
	view := ThinkingIndicator(2)
	CheckGolden(t, "thinking_indicator_tick2", view)
}

// TestThinkingGolden_Empty tests empty thinking state
func TestThinkingGolden_Empty(t *testing.T) {
	state := NewThinkingState()
	config := DefaultThinkingConfig()

	view := RenderAllThinkingBlocks(state, config)

	// Empty state should return empty string
	if view != "" {
		t.Errorf("Expected empty string for empty thinking state, got: %s", view)
	}
}

// TestThinkingGolden_Hidden tests hidden thinking blocks
func TestThinkingGolden_Hidden(t *testing.T) {
	state := NewThinkingState()
	state.AddThinkingBlock("This block should not be visible", 0)
	state.ShowThinking = false

	config := DefaultThinkingConfig()
	view := RenderAllThinkingBlocks(state, config)

	// Hidden thinking should return empty string
	if view != "" {
		t.Errorf("Expected empty string for hidden thinking, got: %s", view)
	}
}
