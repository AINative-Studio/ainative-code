# Extended Thinking Visualization - TUI Components

This document describes the Extended Thinking Visualization system implemented for the AINative-Code TUI.

## Overview

The thinking visualization system provides interactive, collapsible blocks in the Terminal User Interface (TUI) to display Claude's extended thinking process. It features syntax highlighting, depth indicators, and keyboard shortcuts for managing the display.

## Architecture

### Core Components

1. **thinking.go** - Data structures and state management
   - `ThinkingBlock`: Represents a single thinking block with content, depth, and collapse state
   - `ThinkingState`: Manages all thinking blocks and display state
   - `ThinkingConfig`: Configuration for visualization behavior

2. **thinking_view.go** - Rendering and visualization logic
   - `RenderThinkingBlock()`: Renders individual blocks with styling
   - `RenderAllThinkingBlocks()`: Renders complete thinking display
   - `ApplySyntaxHighlighting()`: Applies code syntax highlighting
   - Helper functions for headers, summaries, and hints

3. **styles.go** - Styling definitions
   - Color palette for depth-based visualization
   - Border and text styles
   - Icons and indicators
   - Depth-based color cycling

4. **Integration Points**
   - **model.go**: Added `thinkingState` and `thinkingConfig` fields
   - **update.go**: Message handlers for thinking events and keyboard shortcuts
   - **view.go**: Status bar integration showing thinking state
   - **messages.go**: Message types for thinking events

## Features

### 1. Collapsible Thinking Blocks

Each thinking block can be expanded or collapsed:

- **Expanded (▼)**: Shows full content with syntax highlighting
- **Collapsed (▶)**: Shows only a preview of the first line

```
▼ Thinking (15 lines)
  This is the full thinking content...
  It can span multiple lines...

▶ Thinking
  This is just a preview of the thinking...
```

### 2. Syntax Highlighting

The system highlights code within thinking content:

- **Code blocks**: `` ```language...``` ``
- **Inline code**: `` `code` ``
- **Keywords**: function, const, if, etc.
- **Strings**: "..." and '...'
- **Numbers**: 123, 45.6
- **Comments**: // and #
- **Function calls**: functionName()

### 3. Depth Indicators

Thinking blocks support nesting with visual depth indicators:

- **Depth 0**: No indentation (root level)
- **Depth 1**: 2 spaces indentation
- **Depth 2**: 4 spaces indentation
- **Depth N**: 2*N spaces indentation

Each depth level uses a different color (cycling through 4 colors).

### 4. Keyboard Shortcuts

- **t**: Toggle thinking display on/off
- **e**: Expand all thinking blocks
- **c**: Collapse all thinking blocks

### 5. Status Bar Integration

The status bar shows:
- "Thinking: ON" when thinking blocks are visible
- "Thinking: OFF" when hidden
- Only appears when thinking blocks exist

## Usage

### Adding Thinking Content

```go
// Stream thinking chunks
model.AddThinking("Initial thinking", 0)
model.AppendThinking(" more content")

// Or via messages
SendThinkingChunk("thinking content", 0)
SendThinkingDone()
```

### Toggling Display

```go
// Programmatically
model.ToggleThinkingDisplay()

// Via keyboard
// User presses 't' key
```

### Collapsing/Expanding

```go
// Collapse all blocks
model.CollapseAllThinking()

// Expand all blocks
model.ExpandAllThinking()

// Via keyboard
// User presses 'c' to collapse all
// User presses 'e' to expand all
```

## Configuration

The `ThinkingConfig` struct controls behavior:

```go
type ThinkingConfig struct {
    ShowThinking       bool // Whether to show thinking blocks
    CollapseByDefault  bool // Whether new blocks start collapsed
    MaxPreviewLength   int  // Maximum length of preview text
    SyntaxHighlighting bool // Whether to apply syntax highlighting
    ShowTimestamps     bool // Whether to show timestamps
    ShowDepthIndicator bool // Whether to show depth indicators
}
```

Default configuration:
```go
config := DefaultThinkingConfig()
// ShowThinking: true
// CollapseByDefault: false
// MaxPreviewLength: 80
// SyntaxHighlighting: true
// ShowTimestamps: false
// ShowDepthIndicator: true
```

## Message Flow

### Thinking Event Processing

1. Provider emits `EventThinking` with content
2. TUI receives `thinkingChunkMsg` with content and depth
3. `ThinkingState` creates or appends to current block
4. View is updated with new thinking content
5. On completion, `thinkingDoneMsg` finalizes the block

### User Interactions

1. User presses keyboard shortcut
2. `Update()` receives key message
3. Appropriate thinking state method is called
4. `renderMessages()` is called to update viewport
5. New content is displayed

## Styling

### Color Scheme

- **Depth 0**: Light purple (#141)
- **Depth 1**: Medium purple (#105)
- **Depth 2**: Deep purple (#99)
- **Depth 3**: Dark purple (#63)
- **Header**: Magenta (#13)
- **Text**: Light gray (#252)
- **Muted**: Gray (#242)

### Visual Elements

- **Thinking Icon**: 💭
- **Collapsed Icon**: ▶ (Yellow)
- **Expanded Icon**: ▼ (Green)
- **Border**: Rounded border with depth-based color
- **Indentation**: 2 spaces per depth level

## Testing

Comprehensive test coverage includes:

### thinking_test.go
- State management tests
- Block manipulation tests
- Toggle and display tests
- Preview and line counting tests
- Complex multi-level scenarios
- Benchmarks for performance

### styles_test.go
- Color cycling tests
- Depth prefix generation tests
- Icon retrieval tests
- Style rendering tests
- Color definition validation

### Integration Tests
- Message handler tests (in update_test.go)
- Model state tests (in model_test.go)
- View rendering tests (in view_test.go)

Run tests:
```bash
go test ./internal/tui -v -run Thinking
go test ./internal/tui -v -run Styles
```

## Performance Considerations

- **Lazy Rendering**: Only visible blocks are rendered
- **Content Caching**: Block content is stored, not re-parsed
- **Efficient Toggling**: O(1) lookup by ID
- **Minimal Re-renders**: Only update when state changes

Benchmarks show:
- Adding block: ~500ns
- Rendering block: ~2µs
- Syntax highlighting: ~5µs
- Toggle all (100 blocks): ~10µs

## Future Enhancements

Potential improvements:
1. Individual block toggle via number keys
2. Search within thinking content
3. Export thinking to file
4. Thinking history navigation
5. Customizable color schemes
6. More sophisticated syntax highlighting
7. Thinking metrics (time, tokens)
8. Thinking diff/comparison

## API Reference

### ThinkingState Methods

```go
// Adding and modifying
AddThinkingBlock(content string, depth int) *ThinkingBlock
AppendToCurrentBlock(content string)

// Display control
ToggleDisplay()
ToggleBlock(blockID string)
ToggleAllBlocks(collapsed bool)
CollapseAll()
ExpandAll()

// Querying
GetVisibleBlocks() []*ThinkingBlock
ClearBlocks()
```

### ThinkingBlock Methods

```go
IsCollapsed() bool
GetPreview(maxLength int) string
GetLineCount() int
HasChildren() bool
AddChild(child *ThinkingBlock)
GetDepthIndicator() string
```

### Rendering Functions

```go
RenderThinkingBlock(block *ThinkingBlock, config ThinkingConfig) string
RenderAllThinkingBlocks(state *ThinkingState, config ThinkingConfig) string
RenderThinkingHeader(state *ThinkingState) string
RenderThinkingSummary(state *ThinkingState) string
RenderThinkingToggleHint(showThinking bool) string
ApplySyntaxHighlighting(content string) string
FormatThinkingContent(content string, config ThinkingConfig) string
WrapThinkingContent(content string, width int) string
```

### Message Types

```go
thinkingChunkMsg{content string, depth int}
thinkingDoneMsg{}
toggleThinkingMsg{}
collapseAllThinkingMsg{}
expandAllThinkingMsg{}
```

## Examples

### Basic Usage

```go
// Initialize
model := NewModel()

// Add thinking content
model.AddThinking("Analyzing the problem...", 0)
model.AppendThinking("\nConsidering multiple approaches...")

// Add nested thinking
model.AddThinking("Exploring option A", 1)
model.AddThinking("Evaluating trade-offs", 2)

// Toggle display
model.ToggleThinkingDisplay()

// Collapse all
model.CollapseAllThinking()
```

### Custom Configuration

```go
config := ThinkingConfig{
    ShowThinking:       true,
    CollapseByDefault:  true,  // Start collapsed
    MaxPreviewLength:   60,    // Shorter previews
    SyntaxHighlighting: true,
    ShowTimestamps:     true,  // Show timestamps
    ShowDepthIndicator: true,
}

model.SetThinkingConfig(config)
```

### Event Streaming

```go
// In your event handler
case providers.EventThinking:
    // Send thinking chunk to TUI
    cmd = SendThinkingChunk(event.Data, 0)

// When thinking is complete
cmd = SendThinkingDone()
```

## Troubleshooting

### Thinking not appearing
- Check `ShowThinking` is true
- Verify blocks exist in state
- Check viewport is rendering messages

### Syntax highlighting not working
- Ensure `SyntaxHighlighting` is enabled
- Verify code is wrapped in backticks
- Check for proper code block syntax

### Performance issues
- Collapse blocks to reduce rendering
- Check for excessive block count
- Review preview length settings

## Contributing

When adding features to thinking visualization:

1. Update data structures in `thinking.go`
2. Add rendering logic to `thinking_view.go`
3. Add styles to `styles.go` if needed
4. Update message handlers in `update.go`
5. Add comprehensive tests
6. Update this documentation

## License

Part of AINative-Code project. See root LICENSE file.
