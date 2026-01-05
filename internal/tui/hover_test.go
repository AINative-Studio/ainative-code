package tui

import (
	"context"
	"testing"

	"github.com/AINative-studio/ainative-code/pkg/lsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestHover(t *testing.T) {
	t.Run("requests hover information for position", func(t *testing.T) {
		model := NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()
		_ = model.lspClient.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		params := lsp.HoverParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: lsp.Position{
				Line:      30,
				Character: 10,
			},
		}

		hover, err := RequestHover(ctx, &model, params)
		require.NoError(t, err)
		assert.NotNil(t, hover)
	})

	t.Run("returns error when LSP not enabled", func(t *testing.T) {
		model := NewModel()
		ctx := context.Background()

		params := lsp.HoverParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: lsp.Position{
				Line:      30,
				Character: 10,
			},
		}

		_, err := RequestHover(ctx, &model, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LSP not enabled")
	})

	t.Run("returns nil for position without symbol", func(t *testing.T) {
		model := NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()
		_ = model.lspClient.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		params := lsp.HoverParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: lsp.Position{
				Line:      0,
				Character: 0,
			},
		}

		hover, err := RequestHover(ctx, &model, params)
		require.NoError(t, err)
		assert.Nil(t, hover)
	})
}

func TestRenderHover(t *testing.T) {
	t.Run("renders hover popup with markdown content", func(t *testing.T) {
		model := NewModel()
		model.hoverInfo = &lsp.Hover{
			Contents: lsp.MarkupContent{
				Kind: "markdown",
				Value: "```go\ntype Model struct\n```\n\n" +
					"Model represents the TUI application state",
			},
		}
		model.showHover = true

		rendered := RenderHover(&model)
		assert.NotEmpty(t, rendered)
		assert.Contains(t, rendered, "Model")
		assert.Contains(t, rendered, "struct")
	})

	t.Run("renders hover popup with plaintext content", func(t *testing.T) {
		model := NewModel()
		model.hoverInfo = &lsp.Hover{
			Contents: lsp.MarkupContent{
				Kind:  "plaintext",
				Value: "This is a simple text description",
			},
		}
		model.showHover = true

		rendered := RenderHover(&model)
		assert.NotEmpty(t, rendered)
		assert.Contains(t, rendered, "simple text description")
	})

	t.Run("returns empty string when no hover info", func(t *testing.T) {
		model := NewModel()
		model.hoverInfo = nil
		model.showHover = false

		rendered := RenderHover(&model)
		assert.Empty(t, rendered)
	})

	t.Run("wraps long lines properly", func(t *testing.T) {
		model := NewModel()
		longText := ""
		for i := 0; i < 10; i++ {
			longText += "This is a very long line that should be wrapped. "
		}
		model.hoverInfo = &lsp.Hover{
			Contents: lsp.MarkupContent{
				Kind:  "plaintext",
				Value: longText,
			},
		}
		model.showHover = true

		rendered := RenderHover(&model)
		assert.NotEmpty(t, rendered)
		// Check that content is wrapped
		assert.Contains(t, rendered, "wrapped")
	})
}

func TestFormatHoverContent(t *testing.T) {
	t.Run("formats markdown content", func(t *testing.T) {
		content := lsp.MarkupContent{
			Kind: "markdown",
			Value: "```go\nfunc main() {}\n```\n\n" +
				"**Description:**\nThis is the main function",
		}

		formatted := FormatHoverContent(content)
		assert.NotEmpty(t, formatted)
		assert.Contains(t, formatted, "func main()")
	})

	t.Run("formats plaintext content", func(t *testing.T) {
		content := lsp.MarkupContent{
			Kind:  "plaintext",
			Value: "Simple description text",
		}

		formatted := FormatHoverContent(content)
		assert.NotEmpty(t, formatted)
		assert.Contains(t, formatted, "Simple description text")
	})

	t.Run("handles empty content", func(t *testing.T) {
		content := lsp.MarkupContent{
			Kind:  "markdown",
			Value: "",
		}

		formatted := FormatHoverContent(content)
		assert.Empty(t, formatted)
	})
}

func TestExtractCodeBlocks(t *testing.T) {
	t.Run("extracts code blocks from markdown", func(t *testing.T) {
		markdown := "```go\nfunc test() {}\n```\n\nSome text\n\n```python\nprint('hello')\n```"

		blocks := ExtractCodeBlocks(markdown)
		assert.Len(t, blocks, 2)
		assert.Equal(t, "go", blocks[0].Language)
		assert.Contains(t, blocks[0].Code, "func test()")
		assert.Equal(t, "python", blocks[1].Language)
		assert.Contains(t, blocks[1].Code, "print")
	})

	t.Run("handles markdown without code blocks", func(t *testing.T) {
		markdown := "Just some regular text without code blocks"

		blocks := ExtractCodeBlocks(markdown)
		assert.Empty(t, blocks)
	})

	t.Run("handles malformed code blocks", func(t *testing.T) {
		markdown := "```\nCode without language\n```"

		blocks := ExtractCodeBlocks(markdown)
		assert.Len(t, blocks, 1)
		assert.Equal(t, "", blocks[0].Language)
	})
}

func TestTriggerHover(t *testing.T) {
	t.Run("triggers hover at position", func(t *testing.T) {
		model := NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()
		_ = model.lspClient.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err := TriggerHover(ctx, &model, documentURI, 30, 10)
		require.NoError(t, err)

		assert.True(t, model.showHover)
		assert.NotNil(t, model.hoverInfo)
	})

	t.Run("returns error when LSP not enabled", func(t *testing.T) {
		model := NewModel()
		ctx := context.Background()

		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err := TriggerHover(ctx, &model, documentURI, 30, 10)
		assert.Error(t, err)
	})

	t.Run("clears hover when no info at position", func(t *testing.T) {
		model := NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()
		_ = model.lspClient.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err := TriggerHover(ctx, &model, documentURI, 0, 0)
		require.NoError(t, err)

		assert.False(t, model.showHover)
		assert.Nil(t, model.hoverInfo)
	})
}
