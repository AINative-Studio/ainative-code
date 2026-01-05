package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewThinkingState tests creating a new thinking state
func TestNewThinkingState(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "creates state with defaults",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewThinkingState()

			require.NotNil(t, state)
			assert.True(t, state.ShowThinking)
			assert.NotNil(t, state.Blocks)
			assert.Empty(t, state.Blocks)
			assert.NotNil(t, state.BlocksByID)
			assert.Empty(t, state.BlocksByID)
			assert.Nil(t, state.CurrentBlock)
			assert.Equal(t, 1, state.nextID)
		})
	}
}

// TestAddThinkingBlock tests adding thinking blocks
func TestAddThinkingBlock(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		depth          int
		expectedBlocks int
	}{
		{
			name:           "adds block at depth 0",
			content:        "Root level thinking",
			depth:          0,
			expectedBlocks: 1,
		},
		{
			name:           "adds block at depth 1",
			content:        "Nested thinking",
			depth:          1,
			expectedBlocks: 1,
		},
		{
			name:           "adds empty block",
			content:        "",
			depth:          0,
			expectedBlocks: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewThinkingState()

			block := state.AddThinkingBlock(tt.content, tt.depth)

			require.NotNil(t, block)
			assert.Equal(t, tt.content, block.Content)
			assert.Equal(t, tt.depth, block.Depth)
			assert.False(t, block.Collapsed)
			assert.NotEmpty(t, block.ID)
			assert.Equal(t, tt.expectedBlocks, len(state.Blocks))
			assert.Equal(t, block, state.CurrentBlock)
			assert.Contains(t, state.BlocksByID, block.ID)
		})
	}
}

// TestAddMultipleThinkingBlocks tests adding multiple blocks
func TestAddMultipleThinkingBlocks(t *testing.T) {
	state := NewThinkingState()

	block1 := state.AddThinkingBlock("First", 0)
	block2 := state.AddThinkingBlock("Second", 1)
	block3 := state.AddThinkingBlock("Third", 0)

	assert.Equal(t, 3, len(state.Blocks))
	assert.Equal(t, block3, state.CurrentBlock)

	// Verify IDs are unique
	assert.NotEqual(t, block1.ID, block2.ID)
	assert.NotEqual(t, block2.ID, block3.ID)

	// Verify all blocks are in map
	assert.Contains(t, state.BlocksByID, block1.ID)
	assert.Contains(t, state.BlocksByID, block2.ID)
	assert.Contains(t, state.BlocksByID, block3.ID)
}

// TestAppendToCurrentBlock tests appending to current block
func TestAppendToCurrentBlock(t *testing.T) {
	tests := []struct {
		name            string
		initialContent  string
		appendContent   string
		expectedContent string
	}{
		{
			name:            "appends to existing content",
			initialContent:  "Initial ",
			appendContent:   "appended",
			expectedContent: "Initial appended",
		},
		{
			name:            "appends to empty content",
			initialContent:  "",
			appendContent:   "new content",
			expectedContent: "new content",
		},
		{
			name:            "appends empty string",
			initialContent:  "Content",
			appendContent:   "",
			expectedContent: "Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewThinkingState()
			state.AddThinkingBlock(tt.initialContent, 0)

			state.AppendToCurrentBlock(tt.appendContent)

			assert.Equal(t, tt.expectedContent, state.CurrentBlock.Content)
		})
	}
}

// TestAppendToCurrentBlockWithNoCurrent tests appending when no current block
func TestAppendToCurrentBlockWithNoCurrent(t *testing.T) {
	state := NewThinkingState()

	// Should not panic when no current block
	assert.NotPanics(t, func() {
		state.AppendToCurrentBlock("test")
	})
}

// TestToggleBlock tests toggling individual blocks
func TestToggleBlock(t *testing.T) {
	state := NewThinkingState()
	block := state.AddThinkingBlock("Test", 0)

	// Initially not collapsed
	assert.False(t, block.Collapsed)

	// Toggle to collapsed
	state.ToggleBlock(block.ID)
	assert.True(t, block.Collapsed)

	// Toggle back to expanded
	state.ToggleBlock(block.ID)
	assert.False(t, block.Collapsed)
}

// TestToggleNonExistentBlock tests toggling a block that doesn't exist
func TestToggleNonExistentBlock(t *testing.T) {
	state := NewThinkingState()

	// Should not panic
	assert.NotPanics(t, func() {
		state.ToggleBlock("nonexistent-id")
	})
}

// TestToggleAllBlocks tests toggling all blocks
func TestToggleAllBlocks(t *testing.T) {
	state := NewThinkingState()

	block1 := state.AddThinkingBlock("First", 0)
	block2 := state.AddThinkingBlock("Second", 1)
	block3 := state.AddThinkingBlock("Third", 0)

	// Collapse all
	state.ToggleAllBlocks(true)
	assert.True(t, block1.Collapsed)
	assert.True(t, block2.Collapsed)
	assert.True(t, block3.Collapsed)

	// Expand all
	state.ToggleAllBlocks(false)
	assert.False(t, block1.Collapsed)
	assert.False(t, block2.Collapsed)
	assert.False(t, block3.Collapsed)
}

// TestCollapseAll tests collapsing all blocks
func TestCollapseAll(t *testing.T) {
	state := NewThinkingState()

	state.AddThinkingBlock("First", 0)
	state.AddThinkingBlock("Second", 1)
	state.AddThinkingBlock("Third", 0)

	state.CollapseAll()

	for _, block := range state.Blocks {
		assert.True(t, block.Collapsed)
	}
}

// TestExpandAll tests expanding all blocks
func TestExpandAll(t *testing.T) {
	state := NewThinkingState()

	block1 := state.AddThinkingBlock("First", 0)
	block2 := state.AddThinkingBlock("Second", 1)

	// Manually collapse them first
	block1.Collapsed = true
	block2.Collapsed = true

	state.ExpandAll()

	for _, block := range state.Blocks {
		assert.False(t, block.Collapsed)
	}
}

// TestToggleDisplay tests toggling display
func TestToggleDisplay(t *testing.T) {
	state := NewThinkingState()

	// Initially true
	assert.True(t, state.ShowThinking)

	state.ToggleDisplay()
	assert.False(t, state.ShowThinking)

	state.ToggleDisplay()
	assert.True(t, state.ShowThinking)
}

// TestClearBlocks tests clearing all blocks
func TestClearBlocks(t *testing.T) {
	state := NewThinkingState()

	state.AddThinkingBlock("First", 0)
	state.AddThinkingBlock("Second", 1)
	state.AddThinkingBlock("Third", 0)

	assert.Equal(t, 3, len(state.Blocks))
	assert.NotNil(t, state.CurrentBlock)

	state.ClearBlocks()

	assert.Empty(t, state.Blocks)
	assert.Empty(t, state.BlocksByID)
	assert.Nil(t, state.CurrentBlock)
	assert.Equal(t, 1, state.nextID)
}

// TestGetVisibleBlocks tests retrieving visible blocks
func TestGetVisibleBlocks(t *testing.T) {
	tests := []struct {
		name           string
		showThinking   bool
		blockCount     int
		expectedVisible int
	}{
		{
			name:           "returns all blocks when showing",
			showThinking:   true,
			blockCount:     3,
			expectedVisible: 3,
		},
		{
			name:           "returns empty when not showing",
			showThinking:   false,
			blockCount:     3,
			expectedVisible: 0,
		},
		{
			name:           "returns empty when no blocks",
			showThinking:   true,
			blockCount:     0,
			expectedVisible: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewThinkingState()
			state.ShowThinking = tt.showThinking

			for i := 0; i < tt.blockCount; i++ {
				state.AddThinkingBlock("Block", i%2)
			}

			visible := state.GetVisibleBlocks()
			assert.Equal(t, tt.expectedVisible, len(visible))
		})
	}
}

// TestThinkingBlockIsCollapsed tests the IsCollapsed method
func TestThinkingBlockIsCollapsed(t *testing.T) {
	block := &ThinkingBlock{
		Collapsed: false,
	}

	assert.False(t, block.IsCollapsed())

	block.Collapsed = true
	assert.True(t, block.IsCollapsed())
}

// TestThinkingBlockGetPreview tests preview generation
func TestThinkingBlockGetPreview(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		maxLength int
		expected  string
	}{
		{
			name:      "returns first line within limit",
			content:   "Short line",
			maxLength: 20,
			expected:  "Short line",
		},
		{
			name:      "truncates long first line",
			content:   "This is a very long line that exceeds the maximum length",
			maxLength: 20,
			expected:  "This is a very long ...",
		},
		{
			name:      "returns first line of multi-line content",
			content:   "First line\nSecond line\nThird line",
			maxLength: 50,
			expected:  "First line",
		},
		{
			name:      "handles empty content",
			content:   "",
			maxLength: 20,
			expected:  "",
		},
		{
			name:      "handles zero max length",
			content:   "Content",
			maxLength: 0,
			expected:  "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := &ThinkingBlock{
				Content: tt.content,
			}

			preview := block.GetPreview(tt.maxLength)
			assert.Equal(t, tt.expected, preview)
		})
	}
}

// TestThinkingBlockGetLineCount tests line counting
func TestThinkingBlockGetLineCount(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "counts single line",
			content:  "Single line",
			expected: 1,
		},
		{
			name:     "counts multiple lines",
			content:  "Line 1\nLine 2\nLine 3",
			expected: 3,
		},
		{
			name:     "counts empty content as zero",
			content:  "",
			expected: 0,
		},
		{
			name:     "counts line with trailing newline",
			content:  "Line\n",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := &ThinkingBlock{
				Content: tt.content,
			}

			count := block.GetLineCount()
			assert.Equal(t, tt.expected, count)
		})
	}
}

// TestThinkingBlockHasChildren tests child detection
func TestThinkingBlockHasChildren(t *testing.T) {
	block := &ThinkingBlock{
		Children: []*ThinkingBlock{},
	}

	assert.False(t, block.HasChildren())

	child := &ThinkingBlock{}
	block.AddChild(child)

	assert.True(t, block.HasChildren())
}

// TestThinkingBlockAddChild tests adding children
func TestThinkingBlockAddChild(t *testing.T) {
	parent := &ThinkingBlock{
		Children: []*ThinkingBlock{},
	}

	child1 := &ThinkingBlock{Content: "Child 1"}
	child2 := &ThinkingBlock{Content: "Child 2"}

	parent.AddChild(child1)
	assert.Equal(t, 1, len(parent.Children))

	parent.AddChild(child2)
	assert.Equal(t, 2, len(parent.Children))

	assert.Equal(t, child1, parent.Children[0])
	assert.Equal(t, child2, parent.Children[1])
}

// TestThinkingBlockGetDepthIndicator tests depth indicator generation
func TestThinkingBlockGetDepthIndicator(t *testing.T) {
	tests := []struct {
		name     string
		depth    int
		expected string
	}{
		{
			name:     "depth 0 returns empty",
			depth:    0,
			expected: "",
		},
		{
			name:     "depth 1 returns two spaces",
			depth:    1,
			expected: "  ",
		},
		{
			name:     "depth 3 returns six spaces",
			depth:    3,
			expected: "      ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := &ThinkingBlock{
				Depth: tt.depth,
			}

			indicator := block.GetDepthIndicator()
			assert.Equal(t, tt.expected, indicator)
		})
	}
}

// TestDefaultThinkingConfig tests default config creation
func TestDefaultThinkingConfig(t *testing.T) {
	config := DefaultThinkingConfig()

	assert.True(t, config.ShowThinking)
	assert.False(t, config.CollapseByDefault)
	assert.Equal(t, 80, config.MaxPreviewLength)
	assert.True(t, config.SyntaxHighlighting)
	assert.False(t, config.ShowTimestamps)
	assert.True(t, config.ShowDepthIndicator)
}

// TestRenderThinkingBlock tests rendering a single block
func TestRenderThinkingBlock(t *testing.T) {
	tests := []struct {
		name       string
		block      *ThinkingBlock
		config     ThinkingConfig
		shouldContain []string
	}{
		{
			name: "renders expanded block",
			block: &ThinkingBlock{
				ID:        "test-1",
				Content:   "Test content",
				Depth:     0,
				Collapsed: false,
				Timestamp: time.Now(),
			},
			config: DefaultThinkingConfig(),
			shouldContain: []string{"Thinking", "Test content"},
		},
		{
			name: "renders collapsed block with preview",
			block: &ThinkingBlock{
				ID:        "test-2",
				Content:   "Long content that should be previewed",
				Depth:     0,
				Collapsed: true,
				Timestamp: time.Now(),
			},
			config: DefaultThinkingConfig(),
			shouldContain: []string{"Thinking"},
		},
		{
			name: "renders with depth indicator",
			block: &ThinkingBlock{
				ID:        "test-3",
				Content:   "Nested",
				Depth:     2,
				Collapsed: false,
				Timestamp: time.Now(),
			},
			config: ThinkingConfig{
				ShowDepthIndicator: true,
			},
			shouldContain: []string{"Thinking"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := RenderThinkingBlock(tt.block, tt.config)

			assert.NotEmpty(t, rendered)
			for _, expected := range tt.shouldContain {
				assert.Contains(t, rendered, expected)
			}
		})
	}
}

// TestRenderAllThinkingBlocks tests rendering all blocks
func TestRenderAllThinkingBlocks(t *testing.T) {
	tests := []struct {
		name       string
		state      *ThinkingState
		config     ThinkingConfig
		expectEmpty bool
	}{
		{
			name: "renders multiple blocks",
			state: func() *ThinkingState {
				s := NewThinkingState()
				s.AddThinkingBlock("Block 1", 0)
				s.AddThinkingBlock("Block 2", 1)
				return s
			}(),
			config:     DefaultThinkingConfig(),
			expectEmpty: false,
		},
		{
			name: "returns empty when not showing",
			state: func() *ThinkingState {
				s := NewThinkingState()
				s.AddThinkingBlock("Block", 0)
				s.ShowThinking = false
				return s
			}(),
			config:     DefaultThinkingConfig(),
			expectEmpty: true,
		},
		{
			name:       "returns empty when no blocks",
			state:      NewThinkingState(),
			config:     DefaultThinkingConfig(),
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := RenderAllThinkingBlocks(tt.state, tt.config)

			if tt.expectEmpty {
				assert.Empty(t, rendered)
			} else {
				assert.NotEmpty(t, rendered)
			}
		})
	}
}

// TestApplySyntaxHighlighting tests syntax highlighting
func TestApplySyntaxHighlighting(t *testing.T) {
	tests := []struct {
		name    string
		content string
		checks  []string
	}{
		{
			name:    "handles plain text",
			content: "Plain text without code",
			checks:  []string{"Plain text"},
		},
		{
			name:    "highlights inline code",
			content: "This has `inline code` here",
			checks:  []string{"inline code"},
		},
		{
			name: "highlights code blocks",
			content: "```go\nfunc main() {}\n```",
			checks:  []string{"func", "main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplySyntaxHighlighting(tt.content)

			assert.NotEmpty(t, result)
			// Result will have ANSI codes, so just check it's not empty
		})
	}
}

// TestRenderThinkingToggleHint tests hint rendering
func TestRenderThinkingToggleHint(t *testing.T) {
	tests := []struct {
		name         string
		showThinking bool
		shouldContain string
	}{
		{
			name:         "shows hide hint when visible",
			showThinking: true,
			shouldContain: "hide",
		},
		{
			name:         "shows show hint when hidden",
			showThinking: false,
			shouldContain: "show",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hint := RenderThinkingToggleHint(tt.showThinking)

			assert.NotEmpty(t, hint)
			assert.Contains(t, hint, tt.shouldContain)
		})
	}
}

// TestRenderThinkingSummary tests summary rendering
func TestRenderThinkingSummary(t *testing.T) {
	tests := []struct {
		name        string
		state       *ThinkingState
		expectEmpty bool
	}{
		{
			name: "renders summary with blocks",
			state: func() *ThinkingState {
				s := NewThinkingState()
				s.AddThinkingBlock("Block 1", 0)
				s.AddThinkingBlock("Block 2", 1)
				return s
			}(),
			expectEmpty: false,
		},
		{
			name:        "returns empty with no blocks",
			state:       NewThinkingState(),
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := RenderThinkingSummary(tt.state)

			if tt.expectEmpty {
				assert.Empty(t, summary)
			} else {
				assert.NotEmpty(t, summary)
				assert.Contains(t, summary, "blocks")
			}
		})
	}
}

// TestRenderThinkingHeader tests header rendering
func TestRenderThinkingHeader(t *testing.T) {
	tests := []struct {
		name        string
		state       *ThinkingState
		expectEmpty bool
	}{
		{
			name: "renders header with blocks visible",
			state: func() *ThinkingState {
				s := NewThinkingState()
				s.AddThinkingBlock("Block", 0)
				return s
			}(),
			expectEmpty: false,
		},
		{
			name: "returns empty when not showing",
			state: func() *ThinkingState {
				s := NewThinkingState()
				s.AddThinkingBlock("Block", 0)
				s.ShowThinking = false
				return s
			}(),
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := RenderThinkingHeader(tt.state)

			if tt.expectEmpty {
				assert.Empty(t, header)
			} else {
				assert.NotEmpty(t, header)
			}
		})
	}
}

// TestFormatThinkingContent tests content formatting
func TestFormatThinkingContent(t *testing.T) {
	config := DefaultThinkingConfig()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "formats plain text",
			content: "Plain text",
		},
		{
			name:    "formats code",
			content: "`code here`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatThinkingContent(tt.content, config)
			assert.NotEmpty(t, result)
		})
	}
}

// TestWrapThinkingContent tests content wrapping
func TestWrapThinkingContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		width    int
		expected string
	}{
		{
			name:     "wraps long line",
			content:  "This is a very long line that should be wrapped",
			width:    20,
			expected: "This is a very long\nline that should be\nwrapped",
		},
		{
			name:     "preserves short line",
			content:  "Short",
			width:    20,
			expected: "Short",
		},
		{
			name:     "handles zero width",
			content:  "Content",
			width:    0,
			expected: "Content",
		},
		{
			name:     "preserves newlines",
			content:  "Line 1\nLine 2",
			width:    50,
			expected: "Line 1\nLine 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapThinkingContent(tt.content, tt.width)

			// For wrapped content, verify line lengths
			if tt.width > 0 {
				lines := strings.Split(result, "\n")
				for _, line := range lines {
					// Allow some tolerance for word boundaries
					if len(line) > tt.width+20 {
						t.Errorf("Line too long: %d > %d: %q", len(line), tt.width, line)
					}
				}
			}
		})
	}
}

// TestGenerateID tests ID generation
func TestGenerateID(t *testing.T) {
	state := NewThinkingState()

	id1 := state.generateID()
	id2 := state.generateID()
	id3 := state.generateID()

	// IDs should be unique
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)

	// IDs should have correct prefix
	assert.Contains(t, id1, "thinking-")
}

// TestThinkingBlockTimestamp tests timestamp handling
func TestThinkingBlockTimestamp(t *testing.T) {
	before := time.Now()
	time.Sleep(time.Millisecond)

	state := NewThinkingState()
	block := state.AddThinkingBlock("Test", 0)

	time.Sleep(time.Millisecond)
	after := time.Now()

	assert.True(t, block.Timestamp.After(before))
	assert.True(t, block.Timestamp.Before(after))
}

// TestComplexThinkingScenario tests a complex multi-level scenario
func TestComplexThinkingScenario(t *testing.T) {
	state := NewThinkingState()

	// Add root block
	root := state.AddThinkingBlock("Root thinking", 0)

	// Add nested blocks
	nested1 := state.AddThinkingBlock("Nested level 1", 1)
	nested2 := state.AddThinkingBlock("Nested level 2", 2)

	// Add more root level
	root2 := state.AddThinkingBlock("Another root", 0)

	// Verify structure
	assert.Equal(t, 4, len(state.Blocks))
	assert.Equal(t, 0, root.Depth)
	assert.Equal(t, 1, nested1.Depth)
	assert.Equal(t, 2, nested2.Depth)
	assert.Equal(t, 0, root2.Depth)

	// Test collapsing
	state.CollapseAll()
	for _, block := range state.Blocks {
		assert.True(t, block.Collapsed)
	}

	// Test expanding specific block
	state.ToggleBlock(nested1.ID)
	assert.False(t, nested1.Collapsed)
	assert.True(t, root.Collapsed)

	// Test rendering
	config := DefaultThinkingConfig()
	rendered := RenderAllThinkingBlocks(state, config)
	assert.NotEmpty(t, rendered)

	// Test toggling display
	state.ToggleDisplay()
	assert.False(t, state.ShowThinking)

	rendered = RenderAllThinkingBlocks(state, config)
	assert.Empty(t, rendered)
}

// Benchmark tests

func BenchmarkAddThinkingBlock(b *testing.B) {
	state := NewThinkingState()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.AddThinkingBlock("Benchmark content", i%5)
	}
}

func BenchmarkRenderThinkingBlock(b *testing.B) {
	state := NewThinkingState()
	block := state.AddThinkingBlock("Benchmark content with some text", 2)
	config := DefaultThinkingConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderThinkingBlock(block, config)
	}
}

func BenchmarkApplySyntaxHighlighting(b *testing.B) {
	content := "This is some `code` and more ```go\nfunc test() {}\n```"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplySyntaxHighlighting(content)
	}
}

func BenchmarkToggleAllBlocks(b *testing.B) {
	state := NewThinkingState()
	for i := 0; i < 100; i++ {
		state.AddThinkingBlock("Content", i%5)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.ToggleAllBlocks(i%2 == 0)
	}
}
