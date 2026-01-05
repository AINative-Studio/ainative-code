package tui

import (
	"context"
	"testing"

	"github.com/AINative-studio/ainative-code/pkg/lsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestCompletion(t *testing.T) {
	t.Run("requests completion for current position", func(t *testing.T) {
		model := NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()
		_ = model.lspClient.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		params := lsp.CompletionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: lsp.Position{
				Line:      10,
				Character: 5,
			},
		}

		items, err := RequestCompletion(ctx, &model, params)
		require.NoError(t, err)
		assert.NotNil(t, items)
	})

	t.Run("returns error when LSP not enabled", func(t *testing.T) {
		model := NewModel()
		ctx := context.Background()

		params := lsp.CompletionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: lsp.Position{
				Line:      10,
				Character: 5,
			},
		}

		_, err := RequestCompletion(ctx, &model, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LSP not enabled")
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		model := NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx, cancel := context.WithCancel(context.Background())
		_ = model.lspClient.Initialize(ctx, "/Users/aideveloper/AINative-Code")
		cancel() // Cancel immediately

		params := lsp.CompletionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: lsp.Position{
				Line:      10,
				Character: 5,
			},
		}

		_, err := RequestCompletion(ctx, &model, params)
		assert.Error(t, err)
	})
}

func TestRenderCompletion(t *testing.T) {
	t.Run("renders completion popup with items", func(t *testing.T) {
		model := NewModel()
		model.completionItems = []lsp.CompletionItem{
			{
				Label:         "Model",
				Kind:          lsp.CompletionItemKindStruct,
				Detail:        "type Model struct",
				Documentation: "Model represents the TUI application state",
			},
			{
				Label:         "NewModel",
				Kind:          lsp.CompletionItemKindFunction,
				Detail:        "func NewModel() Model",
				Documentation: "NewModel creates a new TUI model",
			},
		}
		model.showCompletion = true
		model.completionIndex = 0

		rendered := RenderCompletion(&model)
		assert.NotEmpty(t, rendered)
		assert.Contains(t, rendered, "Model")
		assert.Contains(t, rendered, "NewModel")
	})

	t.Run("returns empty string when no completion items", func(t *testing.T) {
		model := NewModel()
		model.completionItems = []lsp.CompletionItem{}
		model.showCompletion = false

		rendered := RenderCompletion(&model)
		assert.Empty(t, rendered)
	})

	t.Run("highlights selected item", func(t *testing.T) {
		model := NewModel()
		model.completionItems = []lsp.CompletionItem{
			{Label: "Item1", Kind: lsp.CompletionItemKindFunction},
			{Label: "Item2", Kind: lsp.CompletionItemKindFunction},
		}
		model.showCompletion = true
		model.completionIndex = 1

		rendered := RenderCompletion(&model)
		assert.NotEmpty(t, rendered)
		// Second item should be highlighted
		assert.Contains(t, rendered, "Item2")
	})

	t.Run("truncates long completion lists", func(t *testing.T) {
		model := NewModel()
		items := make([]lsp.CompletionItem, 20)
		for i := 0; i < 20; i++ {
			items[i] = lsp.CompletionItem{
				Label: "Item",
				Kind:  lsp.CompletionItemKindFunction,
			}
		}
		model.completionItems = items
		model.showCompletion = true
		model.completionIndex = 0

		rendered := RenderCompletion(&model)
		assert.NotEmpty(t, rendered)
		// Should only show max items (e.g., 10)
	})
}

func TestInsertCompletion(t *testing.T) {
	t.Run("inserts selected completion into input", func(t *testing.T) {
		model := NewModel()
		model.textInput.SetValue("M")
		model.completionItems = []lsp.CompletionItem{
			{
				Label:      "Model",
				Kind:       lsp.CompletionItemKindStruct,
				InsertText: "Model",
			},
		}
		model.completionIndex = 0
		model.showCompletion = true

		InsertCompletion(&model)

		value := model.textInput.Value()
		assert.Contains(t, value, "Model")
		assert.False(t, model.showCompletion)
	})

	t.Run("does nothing when no completion selected", func(t *testing.T) {
		model := NewModel()
		model.textInput.SetValue("test")
		model.completionItems = []lsp.CompletionItem{}
		model.showCompletion = false

		InsertCompletion(&model)

		assert.Equal(t, "test", model.textInput.Value())
	})

	t.Run("replaces partial word with completion", func(t *testing.T) {
		model := NewModel()
		model.textInput.SetValue("New")
		model.completionItems = []lsp.CompletionItem{
			{
				Label:      "NewModel",
				InsertText: "NewModel",
			},
		}
		model.completionIndex = 0
		model.showCompletion = true

		InsertCompletion(&model)

		value := model.textInput.Value()
		assert.Contains(t, value, "NewModel")
	})
}

func TestGetCompletionKindIcon(t *testing.T) {
	tests := []struct {
		name     string
		kind     lsp.CompletionItemKind
		expected string
	}{
		{"Function", lsp.CompletionItemKindFunction, ""},
		{"Method", lsp.CompletionItemKindMethod, ""},
		{"Variable", lsp.CompletionItemKindVariable, ""},
		{"Struct", lsp.CompletionItemKindStruct, ""},
		{"Interface", lsp.CompletionItemKindInterface, ""},
		{"Keyword", lsp.CompletionItemKindKeyword, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := GetCompletionKindIcon(tt.kind)
			assert.NotEmpty(t, icon)
		})
	}
}

func TestFilterCompletionItems(t *testing.T) {
	t.Run("filters items by prefix", func(t *testing.T) {
		items := []lsp.CompletionItem{
			{Label: "Model"},
			{Label: "NewModel"},
			{Label: "Message"},
			{Label: "GetMessage"},
		}

		filtered := FilterCompletionItems(items, "Mod")
		assert.Len(t, filtered, 2)
		assert.Equal(t, "Model", filtered[0].Label)
		assert.Equal(t, "NewModel", filtered[1].Label)
	})

	t.Run("returns all items for empty prefix", func(t *testing.T) {
		items := []lsp.CompletionItem{
			{Label: "Model"},
			{Label: "Message"},
		}

		filtered := FilterCompletionItems(items, "")
		assert.Len(t, filtered, 2)
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		items := []lsp.CompletionItem{
			{Label: "Model"},
			{Label: "Message"},
		}

		filtered := FilterCompletionItems(items, "Xyz")
		assert.Empty(t, filtered)
	})

	t.Run("case insensitive filtering", func(t *testing.T) {
		items := []lsp.CompletionItem{
			{Label: "Model"},
			{Label: "NewModel"},
		}

		filtered := FilterCompletionItems(items, "mod")
		assert.Len(t, filtered, 2)
	})
}

func TestSortCompletionItems(t *testing.T) {
	t.Run("sorts items by relevance", func(t *testing.T) {
		items := []lsp.CompletionItem{
			{Label: "zebra", SortText: "3"},
			{Label: "apple", SortText: "1"},
			{Label: "banana", SortText: "2"},
		}

		sorted := SortCompletionItems(items)
		assert.Equal(t, "apple", sorted[0].Label)
		assert.Equal(t, "banana", sorted[1].Label)
		assert.Equal(t, "zebra", sorted[2].Label)
	})

	t.Run("falls back to alphabetical when no sort text", func(t *testing.T) {
		items := []lsp.CompletionItem{
			{Label: "zebra"},
			{Label: "apple"},
			{Label: "banana"},
		}

		sorted := SortCompletionItems(items)
		assert.Equal(t, "apple", sorted[0].Label)
		assert.Equal(t, "banana", sorted[1].Label)
		assert.Equal(t, "zebra", sorted[2].Label)
	})
}
