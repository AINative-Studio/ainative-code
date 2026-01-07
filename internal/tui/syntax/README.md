# Syntax Highlighting Package

A high-performance syntax highlighting package for the AINative-Code TUI, providing language-specific code block highlighting for 30+ programming languages.

## Quick Start

### Basic Usage

```go
import "github.com/AINative-studio/ainative-code/internal/tui/syntax"

// Create highlighter with default config
h := syntax.NewHighlighter(syntax.DefaultConfig())

// Highlight a code block
code := `func main() {
    fmt.Println("Hello, World!")
}`
result, err := h.HighlightCode(code, "go")

// Or highlight markdown with multiple code blocks
markdown := `Here's some Go code:

` + "```go" + `
func main() {
    fmt.Println("Hello")
}
` + "```" + `

And some Python:

` + "```python" + `
print("Hello")
` + "```" + `
`

highlighted := h.HighlightMarkdown(markdown)
```

### Configuration

```go
// Use AINative branded theme
h := syntax.NewHighlighter(syntax.AINativeConfig())

// Custom configuration
config := syntax.HighlighterConfig{
    Theme:              "dracula",
    Enable:             true,
    FallbackToPlain:    true,
    MaxCodeBlockLines:  1000,
    UseTerminal256:     true,
    UseTrueColor:       false,
}
h := syntax.NewHighlighter(config)
```

## Supported Languages

### Core Languages (Required)
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

### Additional Languages
- Ruby
- PHP
- Swift
- Kotlin
- Scala
- Bash/Shell
- PowerShell
- HTML
- CSS/SCSS/Sass
- Markdown
- Dockerfile
- Makefile
- And 10+ more...

## Features

- **Fast**: Sub-millisecond highlighting for typical code blocks
- **Memory Efficient**: ~26KB allocation for small blocks
- **Comprehensive**: 30+ language support via Chroma
- **Configurable**: Multiple themes and options
- **Robust**: 90.7% test coverage
- **Terminal-Optimized**: 256-color and true-color support

## API Reference

### Functions

#### `NewHighlighter(config HighlighterConfig) *Highlighter`
Creates a new syntax highlighter with the given configuration.

#### `ParseCodeBlocks(text string) []CodeBlock`
Extracts all code blocks from markdown text.

#### `NormalizeLanguage(lang string) string`
Normalizes language identifiers (e.g., "js" → "javascript").

#### `IsLanguageSupported(language string) bool`
Checks if a language is supported by the highlighter.

#### `SupportedLanguages() []string`
Returns a list of commonly used supported languages.

### Methods

#### `(*Highlighter) HighlightCode(code, language string) (string, error)`
Applies syntax highlighting to a code snippet.

#### `(*Highlighter) HighlightMarkdown(text string) string`
Processes markdown text and highlights all code blocks.

### Types

#### `CodeBlock`
```go
type CodeBlock struct {
    Language string  // Language identifier
    Code     string  // Code content
    Raw      string  // Original markdown block
}
```

#### `HighlighterConfig`
```go
type HighlighterConfig struct {
    Theme              string
    Enable             bool
    FallbackToPlain    bool
    MaxCodeBlockLines  int
    UseTerminal256     bool
    UseTrueColor       bool
}
```

## Performance

### Benchmarks (Apple M3)

| Operation | Time | Memory | Throughput |
|-----------|------|--------|------------|
| Parse code blocks | 3.3 µs | 4 KB | 300K ops/sec |
| Highlight small (10 lines) | 112 µs | 26 KB | 8.9K ops/sec |
| Highlight medium (100 lines) | 216 µs | 55 KB | 4.6K ops/sec |
| Highlight large (500 lines) | 681 µs | 1 MB | 1.4K ops/sec |

### Optimization Features

- **Line Limit**: Blocks >1000 lines use simplified rendering
- **Lazy Processing**: Only highlights visible blocks
- **Efficient Parsing**: Single-pass regex matching
- **Memory Pooling**: Reuse string builders

## Testing

### Run Tests
```bash
# All tests
go test ./internal/tui/syntax/...

# With coverage
go test -cover ./internal/tui/syntax/...

# Benchmarks
go test -bench=. ./internal/tui/syntax/...

# Verbose
go test -v ./internal/tui/syntax/...
```

### Coverage
```
coverage: 90.7% of statements
```

### Test Categories
- Unit tests: Code parsing, normalization, highlighting
- Integration tests: Full workflow, real-world examples
- Benchmark tests: Performance validation
- Edge case tests: Unicode, special characters, large blocks

## Examples

### Example 1: Highlight Go Code
```go
h := syntax.NewHighlighter(syntax.DefaultConfig())

goCode := `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
`

result, err := h.HighlightCode(goCode, "go")
if err != nil {
    log.Fatal(err)
}

fmt.Println(result)
```

### Example 2: Process Markdown
```go
h := syntax.NewHighlighter(syntax.AINativeConfig())

markdown := `# My Document

Here's a Python example:

` + "```python" + `
def greet(name):
    print(f"Hello, {name}!")

greet("World")
` + "```" + `
`

highlighted := h.HighlightMarkdown(markdown)
fmt.Println(highlighted)
```

### Example 3: Check Language Support
```go
if syntax.IsLanguageSupported("rust") {
    fmt.Println("Rust is supported!")
}

languages := syntax.SupportedLanguages()
fmt.Printf("Supported: %v\n", languages)
```

### Example 4: Custom Theme
```go
config := syntax.DefaultConfig()
config.Theme = "github"  // Use GitHub theme instead of Monokai

h := syntax.NewHighlighter(config)
```

## Integration with TUI

The syntax highlighting is automatically integrated into the TUI's message rendering:

```go
// In internal/tui/model.go
type Model struct {
    // ...
    syntaxHighlighter *syntax.Highlighter
    syntaxEnabled     bool
}

// In internal/tui/update.go
func (m *Model) renderMessages() string {
    // ...
    if m.syntaxEnabled && m.syntaxHighlighter != nil {
        content = m.syntaxHighlighter.HighlightMarkdown(msg.Content)
    }
    // ...
}
```

### Control Methods
```go
// Enable/disable highlighting
m.EnableSyntaxHighlighting()
m.DisableSyntaxHighlighting()

// Check status
enabled := m.IsSyntaxHighlightingEnabled()

// Get highlighter instance
h := m.GetSyntaxHighlighter()
```

## Dependencies

- `github.com/alecthomas/chroma/v2` - Syntax highlighting engine
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/dlclark/regexp2` - Extended regex (chroma dependency)

## License

Part of AINative-Code. See main project license.

## Contributing

When adding support for new languages:

1. Check if Chroma supports it: [Chroma Languages](https://github.com/alecthomas/chroma#supported-languages)
2. Add language alias to `NormalizeLanguage()` if needed
3. Add test case to `TestIntegrationLanguageCoverage`
4. Update documentation with new language

## Troubleshooting

**Issue**: Colors don't appear
- Check terminal supports 256 colors: `echo $TERM`
- Try: `export TERM=xterm-256color`

**Issue**: Slow performance
- Check code block size (limit is 1000 lines)
- Consider disabling for very large blocks

**Issue**: Language not recognized
- Check `SupportedLanguages()` list
- Try language alias (e.g., "js" for "javascript")
- Verify code fence syntax: \`\`\`language

## Resources

- [Chroma Documentation](https://github.com/alecthomas/chroma)
- [Available Themes](https://xyproto.github.io/splash/docs/all.html)
- [Terminal Colors Guide](https://misc.flogisoft.com/bash/tip_colors_and_formatting)

---

**Version**: 1.0.0
**Status**: Production Ready
**Coverage**: 90.7%
