package syntax

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []CodeBlock
	}{
		{
			name: "single code block with language",
			input: "Some text\n```go\nfunc main() {}\n```\nMore text",
			expected: []CodeBlock{
				{
					Language: "go",
					Code:     "func main() {}\n",
					Raw:      "```go\nfunc main() {}\n```",
				},
			},
		},
		{
			name: "code block without language",
			input: "```\nplain text\n```",
			expected: []CodeBlock{
				{
					Language: "",
					Code:     "plain text\n",
					Raw:      "```\nplain text\n```",
				},
			},
		},
		{
			name: "multiple code blocks",
			input: "```python\nprint('hello')\n```\n\nSome text\n\n```javascript\nconsole.log('hi')\n```",
			expected: []CodeBlock{
				{
					Language: "python",
					Code:     "print('hello')\n",
					Raw:      "```python\nprint('hello')\n```",
				},
				{
					Language: "javascript",
					Code:     "console.log('hi')\n",
					Raw:      "```javascript\nconsole.log('hi')\n```",
				},
			},
		},
		{
			name:     "no code blocks",
			input:    "Just plain text without code blocks",
			expected: []CodeBlock{},
		},
		{
			name: "code block with special characters in language",
			input: "```c++\nint main() { return 0; }\n```",
			expected: []CodeBlock{
				{
					Language: "c++",
					Code:     "int main() { return 0; }\n",
					Raw:      "```c++\nint main() { return 0; }\n```",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := ParseCodeBlocks(tt.input)
			assert.Equal(t, len(tt.expected), len(blocks), "Number of code blocks should match")

			for i, expected := range tt.expected {
				if i < len(blocks) {
					assert.Equal(t, expected.Language, blocks[i].Language, "Language should match")
					assert.Equal(t, expected.Code, blocks[i].Code, "Code should match")
					assert.Equal(t, expected.Raw, blocks[i].Raw, "Raw block should match")
				}
			}
		})
	}
}

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"golang", "go"},
		{"Go", "go"},
		{"GO", "go"},
		{"js", "javascript"},
		{"JavaScript", "javascript"},
		{"ts", "typescript"},
		{"py", "python"},
		{"Python", "python"},
		{"rb", "ruby"},
		{"rs", "rust"},
		{"cpp", "cpp"},
		{"c++", "cpp"},
		{"C++", "cpp"},
		{"sh", "bash"},
		{"shell", "bash"},
		{"yml", "yaml"},
		{"YAML", "yaml"},
		{"unknown", "unknown"}, // Unknown languages pass through
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeLanguage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsLanguageSupported(t *testing.T) {
	tests := []struct {
		language string
		expected bool
	}{
		{"go", true},
		{"python", true},
		{"javascript", true},
		{"typescript", true},
		{"rust", true},
		{"java", true},
		{"cpp", true},
		{"c++", true},
		{"sql", true},
		{"yaml", true},
		{"json", true},
		{"golang", true}, // Should work after normalization
		{"js", true},     // Should work after normalization
		{"unknownlang123", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			result := IsLanguageSupported(tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewHighlighter(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := DefaultConfig()
		h := NewHighlighter(config)

		assert.NotNil(t, h)
		assert.NotNil(t, h.formatter)
		assert.NotNil(t, h.style)
		assert.True(t, h.config.Enable)
	})

	t.Run("ainative config", func(t *testing.T) {
		config := AINativeConfig()
		h := NewHighlighter(config)

		assert.NotNil(t, h)
		assert.Equal(t, "dracula", h.config.Theme)
		assert.True(t, h.config.Enable)
	})

	t.Run("custom config", func(t *testing.T) {
		config := HighlighterConfig{
			Theme:              "github",
			Enable:             true,
			FallbackToPlain:    true,
			MaxCodeBlockLines:  500,
			UseTerminal256:     true,
			UseTrueColor:       false,
		}
		h := NewHighlighter(config)

		assert.NotNil(t, h)
		assert.Equal(t, 500, h.config.MaxCodeBlockLines)
	})
}

func TestHighlightCode(t *testing.T) {
	h := NewHighlighter(DefaultConfig())

	tests := []struct {
		name     string
		code     string
		language string
		wantErr  bool
	}{
		{
			name:     "go code",
			code:     "func main() {\n\tfmt.Println(\"Hello\")\n}",
			language: "go",
			wantErr:  false,
		},
		{
			name:     "python code",
			code:     "def hello():\n    print('Hello')",
			language: "python",
			wantErr:  false,
		},
		{
			name:     "javascript code",
			code:     "function hello() {\n  console.log('Hello');\n}",
			language: "javascript",
			wantErr:  false,
		},
		{
			name:     "typescript code",
			code:     "const greeting: string = 'Hello';\nconsole.log(greeting);",
			language: "typescript",
			wantErr:  false,
		},
		{
			name:     "rust code",
			code:     "fn main() {\n    println!(\"Hello\");\n}",
			language: "rust",
			wantErr:  false,
		},
		{
			name:     "java code",
			code:     "public class Hello {\n    public static void main(String[] args) {\n        System.out.println(\"Hello\");\n    }\n}",
			language: "java",
			wantErr:  false,
		},
		{
			name:     "cpp code",
			code:     "#include <iostream>\nint main() {\n    std::cout << \"Hello\";\n    return 0;\n}",
			language: "cpp",
			wantErr:  false,
		},
		{
			name:     "sql code",
			code:     "SELECT * FROM users WHERE id = 1;",
			language: "sql",
			wantErr:  false,
		},
		{
			name:     "yaml code",
			code:     "name: test\nversion: 1.0\nservices:\n  - api\n  - web",
			language: "yaml",
			wantErr:  false,
		},
		{
			name:     "json code",
			code:     "{\n  \"name\": \"test\",\n  \"version\": \"1.0\"\n}",
			language: "json",
			wantErr:  false,
		},
		{
			name:     "unsupported language fallback",
			code:     "some random code",
			language: "unknownlang",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.HighlightCode(tt.code, tt.language)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
				// Result should not be empty and should contain some content
				assert.True(t, len(result) > 0)
			}
		})
	}
}

func TestHighlightCodeDisabled(t *testing.T) {
	config := DefaultConfig()
	config.Enable = false
	h := NewHighlighter(config)

	code := "func main() {}"
	result, err := h.HighlightCode(code, "go")

	assert.NoError(t, err)
	assert.Equal(t, code, result, "Should return original code when disabled")
}

func TestHighlightCodeLargeBlock(t *testing.T) {
	h := NewHighlighter(DefaultConfig())

	// Create a large code block exceeding MaxCodeBlockLines
	var sb strings.Builder
	for i := 0; i < 1500; i++ {
		sb.WriteString("func example() {}\n")
	}
	largeCode := sb.String()

	result, err := h.HighlightCode(largeCode, "go")

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// Should still return styled content, just without full syntax highlighting
	// The result will be wrapped in styling, so just check it's not empty
	assert.True(t, len(result) > len(largeCode)/2, "Result should contain substantial content")
}

func TestHighlightMarkdown(t *testing.T) {
	h := NewHighlighter(DefaultConfig())

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "markdown with go code",
			input: "Here's some Go code:\n```go\nfunc main() {}\n```\nThat's it!",
			contains: []string{
				"func", "main",
			},
		},
		{
			name:  "markdown with multiple code blocks",
			input: "Python:\n```python\nprint('hi')\n```\n\nJavaScript:\n```javascript\nconsole.log('hi');\n```",
			contains: []string{
				"print", "console",
			},
		},
		{
			name:     "markdown without code blocks",
			input:    "Just plain text here",
			contains: []string{"Just plain text here"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.HighlightMarkdown(tt.input)
			assert.NotEmpty(t, result)

			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestHighlightMarkdownDisabled(t *testing.T) {
	config := DefaultConfig()
	config.Enable = false
	h := NewHighlighter(config)

	input := "```go\nfunc main() {}\n```"
	result := h.HighlightMarkdown(input)

	assert.Equal(t, input, result, "Should return original markdown when disabled")
}

func TestSupportedLanguages(t *testing.T) {
	languages := SupportedLanguages()

	assert.NotEmpty(t, languages)
	assert.Contains(t, languages, "go")
	assert.Contains(t, languages, "python")
	assert.Contains(t, languages, "javascript")
	assert.Contains(t, languages, "typescript")
	assert.Contains(t, languages, "rust")
	assert.Contains(t, languages, "java")
	assert.Contains(t, languages, "cpp")
	assert.Contains(t, languages, "sql")
	assert.Contains(t, languages, "yaml")
	assert.Contains(t, languages, "json")

	// Should have at least 10 languages
	assert.GreaterOrEqual(t, len(languages), 10)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.Enable)
	assert.True(t, config.FallbackToPlain)
	assert.Equal(t, 1000, config.MaxCodeBlockLines)
	assert.True(t, config.UseTerminal256)
	assert.False(t, config.UseTrueColor)
	assert.NotEmpty(t, config.Theme)
}

func TestAINativeConfig(t *testing.T) {
	config := AINativeConfig()

	assert.True(t, config.Enable)
	assert.Equal(t, "dracula", config.Theme)
	assert.True(t, config.FallbackToPlain)
	assert.True(t, config.UseTerminal256)
}

// Benchmark tests

func BenchmarkParseCodeBlocks(b *testing.B) {
	input := "```go\nfunc main() {}\n```\n\n```python\nprint('hi')\n```\n\n```javascript\nconsole.log('hi');\n```"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseCodeBlocks(input)
	}
}

func BenchmarkHighlightCode(b *testing.B) {
	h := NewHighlighter(DefaultConfig())
	code := "func main() {\n\tfmt.Println(\"Hello, World!\")\n}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.HighlightCode(code, "go")
	}
}

func BenchmarkHighlightMarkdown(b *testing.B) {
	h := NewHighlighter(DefaultConfig())
	markdown := "Here's some code:\n```go\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n```\n\nAnd more:\n```python\nprint('hello')\n```"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.HighlightMarkdown(markdown)
	}
}

func BenchmarkHighlightLargeCode(b *testing.B) {
	h := NewHighlighter(DefaultConfig())

	// Create a 500-line Go file
	var sb strings.Builder
	for i := 0; i < 500; i++ {
		sb.WriteString("func example")
		sb.WriteString(string(rune(i)))
		sb.WriteString("() {\n\tfmt.Println(\"Line ")
		sb.WriteString(string(rune(i)))
		sb.WriteString("\")\n}\n\n")
	}
	code := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.HighlightCode(code, "go")
	}
}

func TestEdgeCases(t *testing.T) {
	h := NewHighlighter(DefaultConfig())

	t.Run("empty code", func(t *testing.T) {
		result, err := h.HighlightCode("", "go")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("empty language", func(t *testing.T) {
		result, err := h.HighlightCode("some code", "")
		assert.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("empty markdown", func(t *testing.T) {
		result := h.HighlightMarkdown("")
		assert.Equal(t, "", result)
	})

	t.Run("malformed code block", func(t *testing.T) {
		input := "```go\nfunc main() {"
		result := h.HighlightMarkdown(input)
		// Should not crash, just return original
		assert.NotEmpty(t, result)
	})

	t.Run("code with unicode", func(t *testing.T) {
		code := "// こんにちは世界\nfunc main() {\n\tfmt.Println(\"Hello, 世界\")\n}"
		result, err := h.HighlightCode(code, "go")
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("code with special characters", func(t *testing.T) {
		code := "SELECT * FROM users WHERE name LIKE '%test%';"
		result, err := h.HighlightCode(code, "sql")
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}
