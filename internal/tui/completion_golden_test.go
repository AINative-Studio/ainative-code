package tui

import (
	"testing"

	"github.com/AINative-studio/ainative-code/pkg/lsp"
)

// TestCompletionGolden_Empty tests empty completion popup
func TestCompletionGolden_Empty(t *testing.T) {
	model := NewModel()
	model.completionItems = []lsp.CompletionItem{}
	model.showCompletion = false

	view := RenderCompletion(&model)

	// Empty completion should return empty string
	if view != "" {
		t.Errorf("Expected empty string for empty completion, got: %s", view)
	}
}

// TestCompletionGolden_Hidden tests hidden completion popup
func TestCompletionGolden_Hidden(t *testing.T) {
	model := NewModel()
	model.completionItems = []lsp.CompletionItem{
		{
			Label:  "SampleFunction",
			Kind:   lsp.CompletionItemKindFunction,
			Detail: "func SampleFunction()",
		},
	}
	model.showCompletion = false

	view := RenderCompletion(&model)

	// Hidden completion should return empty string
	if view != "" {
		t.Errorf("Expected empty string for hidden completion, got: %s", view)
	}
}

// TestCompletionGolden_SingleItem tests completion popup with single item
func TestCompletionGolden_SingleItem(t *testing.T) {
	model := NewModel()
	model.completionItems = []lsp.CompletionItem{
		{
			Label:         "Calculate",
			Kind:          lsp.CompletionItemKindFunction,
			Detail:        "func Calculate(a, b int) int",
			Documentation: "Calculates the sum of two integers",
		},
	}
	model.showCompletion = true
	model.completionIndex = 0

	view := RenderCompletion(&model)
	CheckGolden(t, "completion_single_item", view)
}

// TestCompletionGolden_MultipleItems tests completion popup with multiple items
func TestCompletionGolden_MultipleItems(t *testing.T) {
	model := NewModel()
	model.completionItems = []lsp.CompletionItem{
		{
			Label:  "Model",
			Kind:   lsp.CompletionItemKindStruct,
			Detail: "type Model struct",
		},
		{
			Label:  "Message",
			Kind:   lsp.CompletionItemKindStruct,
			Detail: "type Message struct",
		},
		{
			Label:  "NewModel",
			Kind:   lsp.CompletionItemKindFunction,
			Detail: "func NewModel() Model",
		},
		{
			Label:  "RenderCompletion",
			Kind:   lsp.CompletionItemKindFunction,
			Detail: "func RenderCompletion(m *Model) string",
		},
	}
	model.showCompletion = true
	model.completionIndex = 0

	view := RenderCompletion(&model)
	CheckGolden(t, "completion_multiple_items", view)
}

// TestCompletionGolden_SelectedItem tests completion with non-first item selected
func TestCompletionGolden_SelectedItem(t *testing.T) {
	model := NewModel()
	model.completionItems = []lsp.CompletionItem{
		{
			Label:  "Model",
			Kind:   lsp.CompletionItemKindStruct,
			Detail: "type Model struct",
		},
		{
			Label:  "NewModel",
			Kind:   lsp.CompletionItemKindFunction,
			Detail: "func NewModel() Model",
		},
		{
			Label:  "Update",
			Kind:   lsp.CompletionItemKindMethod,
			Detail: "func (m *Model) Update(msg tea.Msg)",
		},
	}
	model.showCompletion = true
	model.completionIndex = 1 // Select second item

	view := RenderCompletion(&model)
	CheckGolden(t, "completion_selected_item", view)
}

// TestCompletionGolden_VariousKinds tests completion items of various kinds
func TestCompletionGolden_VariousKinds(t *testing.T) {
	model := NewModel()
	model.completionItems = []lsp.CompletionItem{
		{
			Label: "MyFunction",
			Kind:  lsp.CompletionItemKindFunction,
		},
		{
			Label: "MyMethod",
			Kind:  lsp.CompletionItemKindMethod,
		},
		{
			Label: "myVariable",
			Kind:  lsp.CompletionItemKindVariable,
		},
		{
			Label: "MY_CONSTANT",
			Kind:  lsp.CompletionItemKindConstant,
		},
		{
			Label: "MyClass",
			Kind:  lsp.CompletionItemKindClass,
		},
		{
			Label: "MyInterface",
			Kind:  lsp.CompletionItemKindInterface,
		},
		{
			Label: "MyStruct",
			Kind:  lsp.CompletionItemKindStruct,
		},
		{
			Label: "if",
			Kind:  lsp.CompletionItemKindKeyword,
		},
	}
	model.showCompletion = true
	model.completionIndex = 0

	view := RenderCompletion(&model)
	CheckGolden(t, "completion_various_kinds", view)
}

// TestCompletionGolden_LongLabels tests completion with long labels that need truncation
func TestCompletionGolden_LongLabels(t *testing.T) {
	model := NewModel()
	model.completionItems = []lsp.CompletionItem{
		{
			Label:  "VeryLongFunctionNameThatExceedsTheMaximumDisplayWidth",
			Kind:   lsp.CompletionItemKindFunction,
			Detail: "This is a very long detail that should also be truncated",
		},
		{
			Label:  "AnotherExtremelyLongMethodNameForTesting",
			Kind:   lsp.CompletionItemKindMethod,
			Detail: "Another lengthy detail description",
		},
		{
			Label:  "ShortFunc",
			Kind:   lsp.CompletionItemKindFunction,
			Detail: "Brief",
		},
	}
	model.showCompletion = true
	model.completionIndex = 0

	view := RenderCompletion(&model)
	CheckGolden(t, "completion_long_labels", view)
}

// TestCompletionGolden_ManyItems tests completion with more items than can fit (scrolling)
func TestCompletionGolden_ManyItems(t *testing.T) {
	model := NewModel()
	model.completionItems = []lsp.CompletionItem{
		{Label: "Item01", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item02", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item03", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item04", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item05", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item06", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item07", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item08", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item09", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item10", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item11", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item12", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item13", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item14", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item15", Kind: lsp.CompletionItemKindFunction},
	}
	model.showCompletion = true
	model.completionIndex = 0

	view := RenderCompletion(&model)
	CheckGolden(t, "completion_many_items", view)
}

// TestCompletionGolden_ManyItemsMiddleSelected tests scrolling with middle item selected
func TestCompletionGolden_ManyItemsMiddleSelected(t *testing.T) {
	model := NewModel()
	model.completionItems = []lsp.CompletionItem{
		{Label: "Item01", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item02", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item03", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item04", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item05", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item06", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item07", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item08", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item09", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item10", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item11", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item12", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item13", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item14", Kind: lsp.CompletionItemKindFunction},
		{Label: "Item15", Kind: lsp.CompletionItemKindFunction},
	}
	model.showCompletion = true
	model.completionIndex = 7 // Middle item

	view := RenderCompletion(&model)
	CheckGolden(t, "completion_many_items_middle", view)
}

// TestCompletionGolden_WithDetailAndDoc tests completion with both detail and documentation
func TestCompletionGolden_WithDetailAndDoc(t *testing.T) {
	model := NewModel()
	model.completionItems = []lsp.CompletionItem{
		{
			Label:         "ProcessData",
			Kind:          lsp.CompletionItemKindFunction,
			Detail:        "func ProcessData(input []byte) error",
			Documentation: "ProcessData handles raw data and returns any errors",
		},
		{
			Label:         "ValidateInput",
			Kind:          lsp.CompletionItemKindFunction,
			Detail:        "func ValidateInput(s string) bool",
			Documentation: "ValidateInput checks if input string is valid",
		},
	}
	model.showCompletion = true
	model.completionIndex = 0

	view := RenderCompletion(&model)
	CheckGolden(t, "completion_with_detail", view)
}

// TestCompletionGolden_KindIcons tests that each completion kind has the correct icon
func TestCompletionGolden_KindIcons(t *testing.T) {
	// Test individual icons
	tests := []struct {
		name string
		kind lsp.CompletionItemKind
	}{
		{"Function", lsp.CompletionItemKindFunction},
		{"Method", lsp.CompletionItemKindMethod},
		{"Variable", lsp.CompletionItemKindVariable},
		{"Constant", lsp.CompletionItemKindConstant},
		{"Field", lsp.CompletionItemKindField},
		{"Class", lsp.CompletionItemKindClass},
		{"Interface", lsp.CompletionItemKindInterface},
		{"Struct", lsp.CompletionItemKindStruct},
		{"Module", lsp.CompletionItemKindModule},
		{"Property", lsp.CompletionItemKindProperty},
		{"Keyword", lsp.CompletionItemKindKeyword},
		{"Snippet", lsp.CompletionItemKindSnippet},
		{"Enum", lsp.CompletionItemKindEnum},
		{"EnumMember", lsp.CompletionItemKindEnumMember},
	}

	for _, tt := range tests {
		icon := GetCompletionKindIcon(tt.kind)
		if icon == "" {
			t.Errorf("%s: expected non-empty icon, got empty string", tt.name)
		}
	}
}
