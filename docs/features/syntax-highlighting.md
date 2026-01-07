# Syntax Highlighting for TUI

## Overview

The AINative-Code TUI now features comprehensive syntax highlighting for code blocks in chat messages. This enhancement improves code readability and provides a better developer experience when reviewing code examples from the AI assistant.

## Features

### 1. Code Block Detection

The system automatically detects markdown code blocks with the following format:

```
```language
code here
```
```

Supported features:
- Language markers (e.g., \`\`\`go, \`\`\`python)
- Code blocks without language markers (fallback to plain display)
- Multiple code blocks in a single message
- Inline code with backticks

### 2. Language Support

The following programming languages are supported with full syntax highlighting:

**Primary Languages:**
- Go
- Python
- JavaScript
- TypeScript
- Rust
- Java
- C++
- SQL
- YAML
- JSON

**Additional Languages:**
- Ruby
- PHP
- Swift
- Kotlin
- Scala
- Bash/Shell
- PowerShell
- PostgreSQL/MySQL
- HTML
- CSS/SCSS/Sass
- Markdown
- Dockerfile
- Makefile
- Protocol Buffers
- GraphQL
- Regex

**Total:** 30+ languages supported

### 3. Intelligent Fallback

For unsupported or unrecognized languages:
- Code is displayed in a styled code block
- Maintains readability with monospace font
- Background highlighting for visual distinction
- Language label shown if provided

### 4. AINative Branding

The syntax highlighting uses a custom color scheme that:
- Matches the AINative brand identity
- Uses Dracula theme for rich, vibrant colors
- Works well in both light and dark terminal themes
- Provides purple accents matching the TUI's thinking blocks

### 5. Performance Optimization

**Large Code Block Handling:**
- Maximum 1000 lines with full syntax highlighting
- Blocks exceeding limit use simplified rendering
- Prevents UI blocking on very large files

**Benchmark Results (Apple M3):**
- Small code block (10 lines): ~112µs
- Medium code block (100 lines): ~216µs
- Large code block (500 lines): ~681µs

**Memory Efficiency:**
- Small blocks: ~26KB allocation
- Large blocks: ~1MB allocation
- Minimal GC pressure

### 6. Configuration Options

The highlighter supports multiple configuration options:

```go
config := syntax.HighlighterConfig{
    Theme:              "dracula",     // Chroma theme name
    Enable:             true,           // Enable/disable highlighting
    FallbackToPlain:    true,           // Use plain text for unsupported languages
    MaxCodeBlockLines:  1000,           // Max lines before simplified rendering
    UseTerminal256:     true,           // Use 256-color mode
    UseTrueColor:       false,          // Use 24-bit true color (if supported)
}
```

**Pre-configured Options:**
- `DefaultConfig()`: Standard configuration with monokai theme
- `AINativeConfig()`: Branded configuration with dracula theme

## Usage

### Automatic Integration

Syntax highlighting is automatically enabled in the TUI. Code blocks in assistant responses are highlighted with no user action required.

### Example Output

When the assistant responds with:

```
Here's a simple Go function:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```
```

The code will be displayed with:
- Color-coded keywords (package, import, func)
- Highlighted strings ("Hello, World!")
- Distinguished function names (main, Println)
- Background shading for the code block
- Language label ("go") at the top

### Toggling Highlighting

Programmatic control (for future enhancements):

```go
// Enable syntax highlighting
model.EnableSyntaxHighlighting()

// Disable syntax highlighting
model.DisableSyntaxHighlighting()

// Check status
enabled := model.IsSyntaxHighlightingEnabled()
```

## Implementation Details

### Architecture

The syntax highlighting system consists of:

1. **Parser (`ParseCodeBlocks`)**: Extracts code blocks from markdown
2. **Normalizer (`NormalizeLanguage`)**: Maps language aliases to standard names
3. **Highlighter (`HighlightCode`)**: Applies syntax highlighting using Chroma
4. **Renderer (`HighlightMarkdown`)**: Processes full markdown with multiple blocks

### Technology Stack

- **Library**: [Chroma v2](https://github.com/alecthomas/chroma) - Pure Go syntax highlighter
- **Formatters**: Terminal256, Terminal16m (true color)
- **Styles**: Dracula (AINative), Monokai, GitHub, and 20+ others available
- **UI Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss) for terminal styling

### Code Location

```
internal/tui/syntax/
├── highlighter.go          # Core highlighting logic
├── highlighter_test.go     # Unit tests
└── integration_test.go     # Integration tests

internal/tui/
├── model.go               # Highlighter integration
└── update.go              # Message rendering with highlighting
```

## Testing

### Test Coverage

- **Coverage**: 90.7% of statements
- **Unit Tests**: 15+ test cases
- **Integration Tests**: 30+ scenarios
- **Benchmark Tests**: Performance validation

### Running Tests

```bash
# Run all syntax highlighting tests
go test ./internal/tui/syntax/...

# Run with coverage
go test -cover ./internal/tui/syntax/...

# Run benchmarks
go test -bench=. ./internal/tui/syntax/...
```

## Performance Characteristics

### Memory Usage

| Code Block Size | Memory Allocation | Allocations |
|-----------------|-------------------|-------------|
| 10 lines        | ~26 KB           | 373         |
| 100 lines       | ~55 KB           | 690         |
| 500 lines       | ~1 MB            | 2097        |

### Throughput

| Operation              | Time (µs) | Ops/sec   |
|------------------------|-----------|-----------|
| Parse code blocks      | 3.3       | 303,000   |
| Highlight small block  | 112       | 8,900     |
| Highlight medium block | 216       | 4,600     |
| Highlight large block  | 681       | 1,400     |

## Future Enhancements

Potential improvements for future releases:

1. **Custom Themes**: Allow users to define custom color schemes
2. **Keyboard Shortcuts**: Toggle highlighting with hotkey
3. **Line Numbers**: Optional line numbering for code blocks
4. **Copy Support**: Easy copy-to-clipboard for code blocks
5. **Diff Highlighting**: Show code differences in reviews
6. **Collapsible Blocks**: Fold large code blocks
7. **Export Options**: Save highlighted code as HTML/image

## Troubleshooting

### Common Issues

**Issue**: Colors don't appear correctly
- **Solution**: Ensure your terminal supports 256-color mode
- **Check**: Run `echo $TERM` - should show `xterm-256color` or similar

**Issue**: Performance slow with large files
- **Solution**: Code blocks over 1000 lines use simplified rendering
- **Workaround**: Split large files into smaller blocks

**Issue**: Language not recognized
- **Solution**: Check supported languages list
- **Workaround**: Specify language explicitly in code fence

### Debug Mode

To disable syntax highlighting for debugging:

```go
model.DisableSyntaxHighlighting()
```

## References

- [Chroma Documentation](https://github.com/alecthomas/chroma)
- [Supported Languages](https://github.com/alecthomas/chroma#supported-languages)
- [Available Themes](https://xyproto.github.io/splash/docs/all.html)
- [Lipgloss Styling](https://github.com/charmbracelet/lipgloss)

## Related Tasks

- **TASK-022**: Implement Syntax Highlighting (This feature)
- **TASK-021**: TUI Message Display (Dependency)
- Issue #15: Syntax Highlighting for TUI
