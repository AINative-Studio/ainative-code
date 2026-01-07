package syntax

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

// CodeBlock represents a parsed code block from markdown
type CodeBlock struct {
	Language string
	Code     string
	Raw      string // Original markdown block including backticks
}

// HighlighterConfig configures the syntax highlighter
type HighlighterConfig struct {
	Theme              string // Chroma theme name (e.g., "monokai", "dracula", "github")
	Enable             bool   // Whether syntax highlighting is enabled
	FallbackToPlain    bool   // Use plain text for unsupported languages
	MaxCodeBlockLines  int    // Maximum lines before switching to simpler highlighting
	UseTerminal256     bool   // Use 256-color terminal mode
	UseTrueColor       bool   // Use true-color (24-bit) terminal mode
}

// DefaultConfig returns a default highlighter configuration
func DefaultConfig() HighlighterConfig {
	return HighlighterConfig{
		Theme:              "monokai",
		Enable:             true,
		FallbackToPlain:    true,
		MaxCodeBlockLines:  1000,
		UseTerminal256:     true,
		UseTrueColor:       false, // Conservative default for compatibility
	}
}

// AINativeConfig returns a configuration with AINative branding colors
func AINativeConfig() HighlighterConfig {
	return HighlighterConfig{
		Theme:              "dracula", // Purple-ish theme matching AINative branding
		Enable:             true,
		FallbackToPlain:    true,
		MaxCodeBlockLines:  1000,
		UseTerminal256:     true,
		UseTrueColor:       false,
	}
}

// Highlighter provides syntax highlighting for code blocks
type Highlighter struct {
	config    HighlighterConfig
	formatter chroma.Formatter
	style     *chroma.Style
}

// NewHighlighter creates a new syntax highlighter with the given configuration
func NewHighlighter(config HighlighterConfig) *Highlighter {
	// Get the style
	style := styles.Get(config.Theme)
	if style == nil {
		// Fallback to default style if theme not found
		style = styles.Fallback
	}

	// Create the appropriate formatter based on config
	var formatter chroma.Formatter
	if config.UseTrueColor {
		formatter = formatters.Get("terminal16m")
	} else if config.UseTerminal256 {
		formatter = formatters.Get("terminal256")
	} else {
		formatter = formatters.Get("terminal")
	}

	if formatter == nil {
		// Fallback to basic terminal formatter
		formatter = formatters.Get("terminal")
	}

	return &Highlighter{
		config:    config,
		formatter: formatter,
		style:     style,
	}
}

// ParseCodeBlocks extracts code blocks from markdown text
func ParseCodeBlocks(text string) []CodeBlock {
	// Pattern to match code blocks: ```language\ncode\n```
	codeBlockPattern := regexp.MustCompile("(?s)```([\\w+#-]*)\\n(.*?)```")
	matches := codeBlockPattern.FindAllStringSubmatch(text, -1)

	blocks := make([]CodeBlock, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 3 {
			language := strings.ToLower(strings.TrimSpace(match[1]))
			code := match[2]
			raw := match[0]

			blocks = append(blocks, CodeBlock{
				Language: language,
				Code:     code,
				Raw:      raw,
			})
		}
	}

	return blocks
}

// HighlightCode applies syntax highlighting to a code snippet
func (h *Highlighter) HighlightCode(code, language string) (string, error) {
	if !h.config.Enable {
		return code, nil
	}

	// Check line count for performance
	lineCount := strings.Count(code, "\n") + 1
	if lineCount > h.config.MaxCodeBlockLines {
		// For very large blocks, return plain text with styling
		return h.renderPlainCodeBlock(code, language), nil
	}

	// Normalize language identifier
	language = NormalizeLanguage(language)

	// Get lexer for the language
	lexer := lexers.Get(language)
	if lexer == nil {
		// Try to analyze the code to guess the language
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		// Fallback to plain text if language not supported
		if h.config.FallbackToPlain {
			return h.renderPlainCodeBlock(code, language), nil
		}
		return code, nil
	}

	// Ensure lexer is not nil
	lexer = chroma.Coalesce(lexer)

	// Tokenize the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code, err
	}

	// Format with syntax highlighting
	var buf bytes.Buffer
	err = h.formatter.Format(&buf, h.style, iterator)
	if err != nil {
		return code, err
	}

	return buf.String(), nil
}

// HighlightMarkdown processes markdown text and highlights all code blocks
func (h *Highlighter) HighlightMarkdown(text string) string {
	if !h.config.Enable {
		return text
	}

	// Parse all code blocks
	blocks := ParseCodeBlocks(text)
	if len(blocks) == 0 {
		return text
	}

	// Replace each code block with its highlighted version
	result := text
	for _, block := range blocks {
		highlighted, err := h.HighlightCode(block.Code, block.Language)
		if err != nil {
			// On error, keep the original block
			continue
		}

		// Create a styled code block with header
		styledBlock := h.renderStyledCodeBlock(highlighted, block.Language)

		// Replace the original block with the highlighted version
		result = strings.Replace(result, block.Raw, styledBlock, 1)
	}

	return result
}

// renderStyledCodeBlock wraps highlighted code in a styled container
func (h *Highlighter) renderStyledCodeBlock(highlightedCode, language string) string {
	var sb strings.Builder

	// Add language label if present
	if language != "" {
		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("141")). // Purple matching thinking blocks
			Bold(true).
			Background(lipgloss.Color("235"))
		label := labelStyle.Render(" " + language + " ")
		sb.WriteString(label)
		sb.WriteString("\n")
	}

	// Add the highlighted code with background
	codeStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Padding(1).
		MarginTop(0).
		MarginBottom(1)

	sb.WriteString(codeStyle.Render(highlightedCode))

	return sb.String()
}

// renderPlainCodeBlock renders a code block without syntax highlighting
func (h *Highlighter) renderPlainCodeBlock(code, language string) string {
	var sb strings.Builder

	// Add language label if present
	if language != "" {
		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")).
			Italic(true).
			Background(lipgloss.Color("235"))
		label := labelStyle.Render(" " + language + " ")
		sb.WriteString(label)
		sb.WriteString("\n")
	}

	// Add plain code with background
	codeStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("252")).
		Padding(1).
		MarginTop(0).
		MarginBottom(1)

	sb.WriteString(codeStyle.Render(code))

	return sb.String()
}

// NormalizeLanguage normalizes language identifiers to Chroma lexer names
func NormalizeLanguage(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))

	// Map common aliases to Chroma lexer names
	langMap := map[string]string{
		"golang":     "go",
		"js":         "javascript",
		"ts":         "typescript",
		"py":         "python",
		"rb":         "ruby",
		"rs":         "rust",
		"cpp":        "cpp",
		"c++":        "cpp",
		"cxx":        "cpp",
		"cc":         "cpp",
		"sh":         "bash",
		"shell":      "bash",
		"yml":        "yaml",
		"dockerfile": "docker",
		"makefile":   "make",
		"md":         "markdown",
	}

	if normalized, ok := langMap[lang]; ok {
		return normalized
	}

	return lang
}

// SupportedLanguages returns a list of commonly used supported languages
func SupportedLanguages() []string {
	return []string{
		"go", "python", "javascript", "typescript", "rust",
		"java", "c", "cpp", "c++", "csharp",
		"ruby", "php", "swift", "kotlin", "scala",
		"bash", "shell", "powershell",
		"sql", "postgresql", "mysql",
		"html", "css", "scss", "sass",
		"json", "yaml", "toml", "xml",
		"markdown", "dockerfile", "makefile",
		"proto", "graphql", "regex",
	}
}

// IsLanguageSupported checks if a language is supported by the highlighter
func IsLanguageSupported(language string) bool {
	language = NormalizeLanguage(language)
	lexer := lexers.Get(language)
	return lexer != nil
}
