# LSP Code Intelligence Keyboard Shortcuts

This document describes the keyboard shortcuts for LSP (Language Server Protocol) code intelligence features in the AINative-Code TUI.

## Overview

The LSP integration provides intelligent code assistance including auto-completion, hover information, and code navigation directly within the TUI chat interface.

## Status Indicators

### LSP Connection Status
Located in the status bar at the bottom of the screen:

- `[LSP: ●]` (Green) - Connected and ready
- `[LSP: ○]` (Yellow) - Connecting...
- `[LSP: ✗]` (Red) - Connection error
- `[LSP: -]` (Gray) - Disconnected

## Auto-Completion

### Triggering Completion
Completion suggestions appear automatically as you type in the input field. You can also manually trigger completion:

- **Trigger:** Type in the input field (debounced after 500ms)
- **Manual Trigger:** `Ctrl+Space` (when implemented)

### Navigating Completion Items
When the completion popup is displayed:

- **Next Item:** `Tab` or `Down Arrow (↓)`
- **Previous Item:** `Shift+Tab` or `Up Arrow (↑)`
- **Select & Insert:** `Enter`
- **Cancel:** `Esc`

### Completion Popup Display
- Shows up to 10 items at once
- Displays icon indicating completion kind (function, variable, struct, etc.)
- Shows detail information (type signature)
- Scroll indicator shows position in full list

### Completion Item Icons
- `ƒ` - Function
- `m` - Method
- `v` - Variable
- `c` - Constant
- `f` - Field
- `C` - Class
- `I` - Interface
- `S` - Struct
- `M` - Module
- `k` - Keyword
- `E` - Enum

## Hover Information

### Triggering Hover
Display type information and documentation for symbols in code:

- **Show Hover:** `Ctrl+K`
- **Close Hover:** `Esc`

### Hover Popup Display
- Shows up to 60 characters wide
- Maximum 20 lines of content
- Displays type signature with syntax highlighting
- Shows documentation if available
- Supports markdown formatting:
  - Code blocks with language syntax
  - **Bold** text
  - *Italic* text
  - Automatic text wrapping

### When to Use Hover
- Understanding function signatures
- Reading struct field types
- Viewing variable types
- Reading inline documentation
- Exploring API interfaces

## Code Navigation

### Go to Definition
Jump to the definition of a symbol:

- **Keyboard:** `Ctrl+]`
- **Alternative:** `Ctrl+Click` (when mouse support enabled)

### Find References
Find all references to a symbol:

- **Keyboard:** `Ctrl+Shift+F`

### Navigation Results Display
When navigation results are shown:

- Results grouped by file
- Shows file name with line and column numbers
- Format: `filename.go:line:column`
- Maximum 20 results displayed
- Scroll indicator for more results
- **Close Results:** `Esc`

### Navigation Example
```
Navigation Results (showing 5 of 12)
─────────────────────────────────────
model.go
  Line 9:5
  Line 37:9
  Line 48:15

view.go
  Line 82:20
  Line 134:10
─────────────────────────────────────
Press Esc to close
```

## Performance Features

### Request Debouncing
- Completion requests are debounced (500ms default)
- Prevents excessive LSP server queries
- Configurable timeout: 5 seconds default

### Caching
- Hover information is cached
- Completion results are cached per position
- Cache size: 100 entries (LRU eviction)
- Automatic cache invalidation on file changes

### Async Operations
- All LSP requests are asynchronous
- UI remains responsive during queries
- Loading indicators for long-running operations
- Request cancellation on new input

## Configuration

### LSP Client Configuration
The LSP client can be configured with custom settings:

```go
config := lsp.Config{
    CompletionDebounce: 500 * time.Millisecond,  // Completion trigger delay
    RequestTimeout:     5 * time.Second,          // Max request duration
    EnableCache:        true,                     // Enable result caching
    CacheSize:          100,                      // Number of cached entries
    MaxConcurrentReqs:  10,                       // Max parallel requests
}
```

### Enabling LSP in TUI
Create a model with LSP enabled:

```go
model := tui.NewModelWithLSP("/path/to/workspace")
```

## Troubleshooting

### LSP Not Connecting
- Check workspace path is valid
- Ensure gopls is installed: `go install golang.org/x/tools/gopls@latest`
- Verify LSP status indicator in status bar
- Check for error messages in the status bar

### Completion Not Working
- Verify LSP status shows connected (green `●`)
- Ensure you're typing in a Go source file context
- Check completion debounce hasn't cancelled request
- Verify position is in a valid completion context

### Navigation Not Finding Results
- Ensure symbol is under cursor position
- Verify file has been indexed by LSP server
- Check that symbol has a definition in workspace
- Large workspaces may take time to index

### Performance Issues
- Reduce cache size if memory is constrained
- Increase debounce delay to reduce request frequency
- Check network latency if using remote LSP server
- Monitor LSP server CPU/memory usage

## Best Practices

### Efficient Completion Usage
1. Let auto-completion trigger naturally while typing
2. Use filtering to narrow results quickly
3. Learn common completion icons for faster selection
4. Use Enter to accept, Esc to dismiss

### Effective Hover Usage
1. Hover over unfamiliar function names
2. Check type signatures before calling functions
3. Read documentation without leaving the TUI
4. Use for API exploration and discovery

### Smart Navigation
1. Use Go to Definition to understand implementations
2. Find References to see usage patterns
3. Navigate file-by-file from results
4. Close popups when done to free screen space

## Integration Examples

### Typical Workflow
1. Start typing function name → Completion appears
2. Select from completion → Function inserted
3. Hover over parameter → See type info
4. Ctrl+] on function → Go to definition
5. Ctrl+Shift+F → Find all usages
6. Review results → Navigate to relevant code

### Code Exploration
1. Open file in discussion
2. Use hover to understand types
3. Follow definitions to implementation
4. Find references to see usage
5. Build mental model of codebase

## Future Enhancements

Planned features for future releases:
- Signature help (parameter hints)
- Code actions (refactoring)
- Inline diagnostics (errors/warnings)
- Symbol search (workspace-wide)
- Document formatting
- Rename refactoring
- Import organization

## Support

For issues or feature requests related to LSP integration:
- Check LSP server logs
- Verify gopls version compatibility
- Report bugs with LSP status information
- Include workspace size and file count

## See Also

- [Main Keyboard Shortcuts](./KEYBOARD_SHORTCUTS.md)
- [TUI User Guide](./TUI_GUIDE.md)
- [LSP Specification](https://microsoft.github.io/language-server-protocol/)
- [gopls Documentation](https://github.com/golang/tools/tree/master/gopls)
