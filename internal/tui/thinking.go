package tui

import (
	"strings"
	"time"
)

// ThinkingBlock represents a collapsible thinking content block
type ThinkingBlock struct {
	ID        string    // Unique identifier
	Content   string    // Thinking content
	Depth     int       // Nesting depth (0 = root)
	Collapsed bool      // Whether the block is collapsed
	Timestamp time.Time // When the thinking occurred
	Children  []*ThinkingBlock // Nested thinking blocks
}

// ThinkingState manages the state of all thinking blocks
type ThinkingState struct {
	Blocks        []*ThinkingBlock // All thinking blocks in order
	ShowThinking  bool             // Whether to show thinking blocks
	CurrentBlock  *ThinkingBlock   // Currently active/focused block
	BlocksByID    map[string]*ThinkingBlock // Quick lookup by ID
	nextID        int              // Counter for generating IDs
}

// NewThinkingState creates a new thinking state manager
func NewThinkingState() *ThinkingState {
	return &ThinkingState{
		Blocks:       make([]*ThinkingBlock, 0),
		ShowThinking: true, // Show thinking by default
		BlocksByID:   make(map[string]*ThinkingBlock),
		nextID:       1,
	}
}

// AddThinkingBlock creates and adds a new thinking block
func (ts *ThinkingState) AddThinkingBlock(content string, depth int) *ThinkingBlock {
	block := &ThinkingBlock{
		ID:        ts.generateID(),
		Content:   content,
		Depth:     depth,
		Collapsed: false, // Start expanded
		Timestamp: time.Now(),
		Children:  make([]*ThinkingBlock, 0),
	}

	ts.Blocks = append(ts.Blocks, block)
	ts.BlocksByID[block.ID] = block
	ts.CurrentBlock = block

	return block
}

// AppendToCurrentBlock appends content to the currently active block
func (ts *ThinkingState) AppendToCurrentBlock(content string) {
	if ts.CurrentBlock != nil {
		ts.CurrentBlock.Content += content
	}
}

// ToggleBlock toggles the collapsed state of a block by ID
func (ts *ThinkingState) ToggleBlock(blockID string) {
	if block, exists := ts.BlocksByID[blockID]; exists {
		block.Collapsed = !block.Collapsed
	}
}

// ToggleAllBlocks toggles all blocks to collapsed/expanded
func (ts *ThinkingState) ToggleAllBlocks(collapsed bool) {
	for _, block := range ts.Blocks {
		block.Collapsed = collapsed
	}
}

// CollapseAll collapses all thinking blocks
func (ts *ThinkingState) CollapseAll() {
	ts.ToggleAllBlocks(true)
}

// ExpandAll expands all thinking blocks
func (ts *ThinkingState) ExpandAll() {
	ts.ToggleAllBlocks(false)
}

// ToggleDisplay toggles whether thinking blocks are shown
func (ts *ThinkingState) ToggleDisplay() {
	ts.ShowThinking = !ts.ShowThinking
}

// ClearBlocks removes all thinking blocks
func (ts *ThinkingState) ClearBlocks() {
	ts.Blocks = make([]*ThinkingBlock, 0)
	ts.BlocksByID = make(map[string]*ThinkingBlock)
	ts.CurrentBlock = nil
	ts.nextID = 1
}

// GetVisibleBlocks returns blocks that should be displayed
func (ts *ThinkingState) GetVisibleBlocks() []*ThinkingBlock {
	if !ts.ShowThinking {
		return []*ThinkingBlock{}
	}
	return ts.Blocks
}

// generateID generates a unique ID for a thinking block
func (ts *ThinkingState) generateID() string {
	id := ts.nextID
	ts.nextID++
	return "thinking-" + string(rune('0'+id))
}

// IsCollapsed returns whether a block is collapsed
func (tb *ThinkingBlock) IsCollapsed() bool {
	return tb.Collapsed
}

// GetPreview returns a preview of the thinking content (first line)
func (tb *ThinkingBlock) GetPreview(maxLength int) string {
	lines := strings.Split(tb.Content, "\n")
	if len(lines) == 0 {
		return ""
	}

	preview := lines[0]
	if len(preview) > maxLength {
		preview = preview[:maxLength] + "..."
	}

	return preview
}

// GetLineCount returns the number of lines in the thinking content
func (tb *ThinkingBlock) GetLineCount() int {
	if tb.Content == "" {
		return 0
	}
	return len(strings.Split(tb.Content, "\n"))
}

// HasChildren returns whether the block has nested children
func (tb *ThinkingBlock) HasChildren() bool {
	return len(tb.Children) > 0
}

// AddChild adds a child thinking block
func (tb *ThinkingBlock) AddChild(child *ThinkingBlock) {
	tb.Children = append(tb.Children, child)
}

// GetDepthIndicator returns a visual indicator for the depth level
func (tb *ThinkingBlock) GetDepthIndicator() string {
	if tb.Depth == 0 {
		return ""
	}

	indicator := ""
	for i := 0; i < tb.Depth; i++ {
		indicator += "  " // Two spaces per depth level
	}

	return indicator
}

// ThinkingConfig holds configuration for thinking visualization
type ThinkingConfig struct {
	ShowThinking       bool // Whether to show thinking blocks
	CollapseByDefault  bool // Whether new blocks start collapsed
	MaxPreviewLength   int  // Maximum length of preview text
	SyntaxHighlighting bool // Whether to apply syntax highlighting
	ShowTimestamps     bool // Whether to show timestamps
	ShowDepthIndicator bool // Whether to show depth indicators
}

// DefaultThinkingConfig returns default configuration
func DefaultThinkingConfig() ThinkingConfig {
	return ThinkingConfig{
		ShowThinking:       true,
		CollapseByDefault:  false,
		MaxPreviewLength:   80,
		SyntaxHighlighting: true,
		ShowTimestamps:     false,
		ShowDepthIndicator: true,
	}
}
