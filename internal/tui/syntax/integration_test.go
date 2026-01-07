package syntax

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationFullWorkflow tests the complete syntax highlighting workflow
func TestIntegrationFullWorkflow(t *testing.T) {
	t.Run("complete workflow with multiple languages", func(t *testing.T) {
		h := NewHighlighter(AINativeConfig())

		markdown := `
Here's some Go code:

` + "```go" + `
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
` + "```" + `

And some Python:

` + "```python" + `
def greet(name):
	"""Greet someone by name."""
	print(f"Hello, {name}!")

greet("World")
` + "```" + `

And some JavaScript:

` + "```javascript" + `
function greet(name) {
	console.log('Hello, ' + name + '!');
}

greet("World");
` + "```" + `

Plain text remains unchanged.
`

		result := h.HighlightMarkdown(markdown)

		// Verify result is not empty and has been processed
		assert.NotEmpty(t, result)
		assert.NotEqual(t, markdown, result, "Result should be different from input")

		// Check that language labels are present
		assert.Contains(t, result, "go", "Should contain Go label")
		assert.Contains(t, result, "python", "Should contain Python label")
		assert.Contains(t, result, "javascript", "Should contain JavaScript label")

		// Check that plain text is preserved
		assert.Contains(t, result, "Plain text remains unchanged")
	})

	t.Run("real-world assistant response", func(t *testing.T) {
		h := NewHighlighter(DefaultConfig())

		// Simulate a typical assistant response with code
		response := `I can help you create a REST API endpoint. Here's an example:

` + "```go" + `
package api

import (
	"encoding/json"
	"net/http"
)

type User struct {
	ID   int
	Name string
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	user := User{
		ID:   1,
		Name: "Alice",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
` + "```" + `

To use this, register the handler like:

` + "```go" + `
http.HandleFunc("/api/user", GetUserHandler)
http.ListenAndServe(":8080", nil)
` + "```" + `

This creates a simple API that returns user data as JSON.`

		result := h.HighlightMarkdown(response)

		assert.NotEmpty(t, result)
		// The response should contain code with highlighting
		assert.Contains(t, result, "package")
		assert.Contains(t, result, "func")
		assert.Contains(t, result, "http")
	})

	t.Run("mixed content with inline code", func(t *testing.T) {
		h := NewHighlighter(DefaultConfig())

		markdown := "To install the package, run: `npm install react`\n\nThen create a component:\n\n" +
			"```jsx\n" +
			"import React from 'react';\n\n" +
			"export default function Hello() {\n" +
			"\treturn <h1>Hello, World!</h1>;\n" +
			"}\n" +
			"```\n\n" +
			"Use `import` to include it in your app."

		result := h.HighlightMarkdown(markdown)

		assert.NotEmpty(t, result)
		assert.Contains(t, result, "npm install react")
		assert.Contains(t, result, "import")
	})
}

// TestIntegrationPerformance tests performance with realistic data
func TestIntegrationPerformance(t *testing.T) {
	h := NewHighlighter(DefaultConfig())

	t.Run("moderate code block", func(t *testing.T) {
		// 100 lines of Go code
		var sb strings.Builder
		sb.WriteString("```go\n")
		sb.WriteString("package main\n\nimport \"fmt\"\n\n")
		for i := 0; i < 100; i++ {
			sb.WriteString("func Example")
			sb.WriteString(string(rune('A' + (i % 26))))
			sb.WriteString("() {\n")
			sb.WriteString("\tfmt.Println(\"Example ")
			sb.WriteString(string(rune('A' + (i % 26))))
			sb.WriteString("\")\n")
			sb.WriteString("}\n\n")
		}
		sb.WriteString("```")

		markdown := sb.String()
		result := h.HighlightMarkdown(markdown)

		assert.NotEmpty(t, result)
		// Should complete without hanging
	})

	t.Run("multiple code blocks", func(t *testing.T) {
		var sb strings.Builder

		// Add 10 different code blocks
		languages := []string{"go", "python", "javascript", "rust", "java"}
		for _, lang := range languages {
			sb.WriteString("\n```")
			sb.WriteString(lang)
			sb.WriteString("\n")
			sb.WriteString("// Sample code in ")
			sb.WriteString(lang)
			sb.WriteString("\n")
			for i := 0; i < 20; i++ {
				sb.WriteString("function test")
				sb.WriteString(string(rune('0' + (i % 10))))
				sb.WriteString("() { return true; }\n")
			}
			sb.WriteString("```\n\n")
		}

		markdown := sb.String()
		result := h.HighlightMarkdown(markdown)

		assert.NotEmpty(t, result)
		// All languages should be processed
		for _, lang := range languages {
			assert.Contains(t, result, lang)
		}
	})
}

// TestIntegrationEdgeCases tests edge cases in integration
func TestIntegrationEdgeCases(t *testing.T) {
	h := NewHighlighter(DefaultConfig())

	t.Run("nested code blocks in markdown", func(t *testing.T) {
		markdown := "Here's how to write markdown:\n\n```markdown\n# Title\n\nSome text with `inline code`\n```"

		result := h.HighlightMarkdown(markdown)
		assert.NotEmpty(t, result)
	})

	t.Run("code block with special characters", func(t *testing.T) {
		markdown := "```bash\necho \"Use backticks for command substitution\"\n```"

		result := h.HighlightMarkdown(markdown)
		assert.NotEmpty(t, result)
	})

	t.Run("empty code block", func(t *testing.T) {
		markdown := "```go\n```"

		result := h.HighlightMarkdown(markdown)
		assert.NotEmpty(t, result)
	})

	t.Run("code block with only whitespace", func(t *testing.T) {
		markdown := "```python\n   \n\t\n   \n```"

		result := h.HighlightMarkdown(markdown)
		assert.NotEmpty(t, result)
	})

	t.Run("mixed valid and invalid code blocks", func(t *testing.T) {
		markdown := `
Valid code:
` + "```go" + `
func main() {}
` + "```" + `

Invalid (unclosed):
` + "```python" + `
def test():
    pass

Regular text continues...
`

		result := h.HighlightMarkdown(markdown)
		assert.NotEmpty(t, result)
		// Should process the valid block and leave invalid parts alone
	})
}

// TestIntegrationLanguageCoverage ensures all required languages work
func TestIntegrationLanguageCoverage(t *testing.T) {
	h := NewHighlighter(DefaultConfig())

	requiredLanguages := map[string]string{
		"go":         "package main\nfunc main() {}",
		"python":     "def main():\n    pass",
		"javascript": "function main() {}",
		"typescript": "function main(): void {}",
		"rust":       "fn main() {}",
		"java":       "public class Main { public static void main(String[] args) {} }",
		"cpp":        "#include <iostream>\nint main() { return 0; }",
		"sql":        "SELECT * FROM users;",
		"yaml":       "name: test\nversion: 1.0",
		"json":       "{\"name\": \"test\"}",
	}

	for lang, code := range requiredLanguages {
		t.Run(lang, func(t *testing.T) {
			result, err := h.HighlightCode(code, lang)
			require.NoError(t, err, "Should highlight %s without error", lang)
			assert.NotEmpty(t, result, "Result should not be empty for %s", lang)
			assert.True(t, len(result) > 0, "Should have content for %s", lang)
		})
	}
}

// TestIntegrationConfigurationOptions tests different configuration options
func TestIntegrationConfigurationOptions(t *testing.T) {
	code := "```go\nfunc main() {}\n```"

	t.Run("with highlighting enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.Enable = true
		h := NewHighlighter(config)

		result := h.HighlightMarkdown(code)
		assert.NotEqual(t, code, result)
	})

	t.Run("with highlighting disabled", func(t *testing.T) {
		config := DefaultConfig()
		config.Enable = false
		h := NewHighlighter(config)

		result := h.HighlightMarkdown(code)
		assert.Equal(t, code, result)
	})

	t.Run("with ainative theme", func(t *testing.T) {
		h := NewHighlighter(AINativeConfig())

		result := h.HighlightMarkdown(code)
		assert.NotEmpty(t, result)
		assert.NotEqual(t, code, result)
	})

	t.Run("with different max lines", func(t *testing.T) {
		config := DefaultConfig()
		config.MaxCodeBlockLines = 10

		h := NewHighlighter(config)

		// Create code with more than 10 lines
		var sb strings.Builder
		sb.WriteString("```go\n")
		for i := 0; i < 20; i++ {
			sb.WriteString("func test() {}\n")
		}
		sb.WriteString("```")

		largeCode := sb.String()
		result := h.HighlightMarkdown(largeCode)

		assert.NotEmpty(t, result)
		// Should still process but may use simpler highlighting
	})
}

// TestIntegrationRealWorldExamples tests with real-world code examples
func TestIntegrationRealWorldExamples(t *testing.T) {
	h := NewHighlighter(AINativeConfig())

	t.Run("docker compose file", func(t *testing.T) {
		markdown := `Here's a docker-compose.yml:

` + "```yaml" + `
version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
    volumes:
      - ./html:/usr/share/nginx/html
  db:
    image: postgres:13
    environment:
      POSTGRES_PASSWORD: secret
` + "```"

		result := h.HighlightMarkdown(markdown)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "version")
		assert.Contains(t, result, "services")
	})

	t.Run("sql migration", func(t *testing.T) {
		markdown := `Database migration:

` + "```sql" + `
CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	username VARCHAR(50) UNIQUE NOT NULL,
	email VARCHAR(255) NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
` + "```"

		result := h.HighlightMarkdown(markdown)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "CREATE")
		assert.Contains(t, result, "TABLE")
	})

	t.Run("json config", func(t *testing.T) {
		markdown := `Configuration file:

` + "```json" + `
{
	"name": "ainative-code",
	"version": "1.0.0",
	"dependencies": {
		"chroma": "^2.0.0",
		"lipgloss": "^1.0.0"
	},
	"scripts": {
		"test": "go test ./...",
		"build": "go build"
	}
}
` + "```"

		result := h.HighlightMarkdown(markdown)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "name")
		assert.Contains(t, result, "version")
	})
}
