package tui

import (
	"context"
	"testing"

	"github.com/AINative-studio/ainative-code/pkg/lsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestDefinition(t *testing.T) {
	t.Run("requests definition for symbol", func(t *testing.T) {
		model := NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()
		_ = model.lspClient.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		params := lsp.DefinitionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: lsp.Position{
				Line:      30,
				Character: 10,
			},
		}

		locations, err := RequestDefinition(ctx, &model, params)
		require.NoError(t, err)
		assert.NotNil(t, locations)
	})

	t.Run("returns error when LSP not enabled", func(t *testing.T) {
		model := NewModel()
		ctx := context.Background()

		params := lsp.DefinitionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: lsp.Position{
				Line:      30,
				Character: 10,
			},
		}

		_, err := RequestDefinition(ctx, &model, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LSP not enabled")
	})
}

func TestRequestReferences(t *testing.T) {
	t.Run("requests references for symbol", func(t *testing.T) {
		model := NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()
		_ = model.lspClient.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		params := lsp.ReferencesParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: lsp.Position{
				Line:      30,
				Character: 10,
			},
			Context: lsp.ReferencesContext{
				IncludeDeclaration: true,
			},
		}

		locations, err := RequestReferences(ctx, &model, params)
		require.NoError(t, err)
		assert.NotNil(t, locations)
	})

	t.Run("returns error when LSP not enabled", func(t *testing.T) {
		model := NewModel()
		ctx := context.Background()

		params := lsp.ReferencesParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: lsp.Position{
				Line:      30,
				Character: 10,
			},
			Context: lsp.ReferencesContext{
				IncludeDeclaration: false,
			},
		}

		_, err := RequestReferences(ctx, &model, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LSP not enabled")
	})
}

func TestRenderNavigation(t *testing.T) {
	t.Run("renders navigation results", func(t *testing.T) {
		model := NewModel()
		model.navigationResult = []lsp.Location{
			{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
				Range: lsp.Range{
					Start: lsp.Position{Line: 9, Character: 5},
					End:   lsp.Position{Line: 9, Character: 10},
				},
			},
			{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
				Range: lsp.Range{
					Start: lsp.Position{Line: 37, Character: 9},
					End:   lsp.Position{Line: 37, Character: 14},
				},
			},
		}
		model.showNavigation = true

		rendered := RenderNavigation(&model)
		assert.NotEmpty(t, rendered)
		assert.Contains(t, rendered, "model.go")
		assert.Contains(t, rendered, "9")
		assert.Contains(t, rendered, "37")
	})

	t.Run("returns empty string when no results", func(t *testing.T) {
		model := NewModel()
		model.navigationResult = []lsp.Location{}
		model.showNavigation = false

		rendered := RenderNavigation(&model)
		assert.Empty(t, rendered)
	})

	t.Run("handles many results with pagination", func(t *testing.T) {
		model := NewModel()
		locations := make([]lsp.Location, 50)
		for i := 0; i < 50; i++ {
			locations[i] = lsp.Location{
				URI: "file:///test.go",
				Range: lsp.Range{
					Start: lsp.Position{Line: i, Character: 0},
					End:   lsp.Position{Line: i, Character: 5},
				},
			}
		}
		model.navigationResult = locations
		model.showNavigation = true

		rendered := RenderNavigation(&model)
		assert.NotEmpty(t, rendered)
		// Should indicate there are more results
		assert.Contains(t, rendered, "50")
	})
}

func TestFormatLocation(t *testing.T) {
	t.Run("formats location with file and line", func(t *testing.T) {
		location := lsp.Location{
			URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			Range: lsp.Range{
				Start: lsp.Position{Line: 42, Character: 10},
				End:   lsp.Position{Line: 42, Character: 20},
			},
		}

		formatted := FormatLocation(location)
		assert.NotEmpty(t, formatted)
		assert.Contains(t, formatted, "model.go")
		assert.Contains(t, formatted, "43") // Line numbers are 1-indexed in display
	})

	t.Run("handles URI without file scheme", func(t *testing.T) {
		location := lsp.Location{
			URI: "/Users/aideveloper/AINative-Code/internal/tui/model.go",
			Range: lsp.Range{
				Start: lsp.Position{Line: 10, Character: 5},
				End:   lsp.Position{Line: 10, Character: 15},
			},
		}

		formatted := FormatLocation(location)
		assert.NotEmpty(t, formatted)
		assert.Contains(t, formatted, "model.go")
	})
}

func TestExtractFileName(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected string
	}{
		{
			name:     "file URI with file scheme",
			uri:      "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			expected: "model.go",
		},
		{
			name:     "absolute path without scheme",
			uri:      "/Users/aideveloper/AINative-Code/internal/tui/model.go",
			expected: "model.go",
		},
		{
			name:     "relative path",
			uri:      "internal/tui/model.go",
			expected: "model.go",
		},
		{
			name:     "file name only",
			uri:      "model.go",
			expected: "model.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFileName(tt.uri)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFilePath(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected string
	}{
		{
			name:     "file URI with file scheme",
			uri:      "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			expected: "/Users/aideveloper/AINative-Code/internal/tui/model.go",
		},
		{
			name:     "absolute path without scheme",
			uri:      "/Users/aideveloper/AINative-Code/internal/tui/model.go",
			expected: "/Users/aideveloper/AINative-Code/internal/tui/model.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFilePath(tt.uri)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGotoDefinition(t *testing.T) {
	t.Run("navigates to definition", func(t *testing.T) {
		model := NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()
		_ = model.lspClient.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err := GotoDefinition(ctx, &model, documentURI, 30, 10)
		require.NoError(t, err)

		assert.True(t, model.showNavigation)
		assert.NotEmpty(t, model.navigationResult)
	})

	t.Run("returns error when LSP not enabled", func(t *testing.T) {
		model := NewModel()
		ctx := context.Background()

		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err := GotoDefinition(ctx, &model, documentURI, 30, 10)
		assert.Error(t, err)
	})

	t.Run("clears navigation when no definition found", func(t *testing.T) {
		model := NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()
		_ = model.lspClient.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err := GotoDefinition(ctx, &model, documentURI, 0, 0)
		require.NoError(t, err)

		assert.False(t, model.showNavigation)
	})
}

func TestFindReferences(t *testing.T) {
	t.Run("finds references for symbol", func(t *testing.T) {
		model := NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()
		_ = model.lspClient.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err := FindReferences(ctx, &model, documentURI, 30, 10, true)
		require.NoError(t, err)

		assert.True(t, model.showNavigation)
		assert.NotEmpty(t, model.navigationResult)
	})

	t.Run("returns error when LSP not enabled", func(t *testing.T) {
		model := NewModel()
		ctx := context.Background()

		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err := FindReferences(ctx, &model, documentURI, 30, 10, false)
		assert.Error(t, err)
	})

	t.Run("clears navigation when no references found", func(t *testing.T) {
		model := NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()
		_ = model.lspClient.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err := FindReferences(ctx, &model, documentURI, 0, 0, false)
		require.NoError(t, err)

		assert.False(t, model.showNavigation)
	})
}
