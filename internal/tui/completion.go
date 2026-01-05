package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/AINative-studio/ainative-code/pkg/lsp"
	"github.com/charmbracelet/lipgloss"
)

const (
	maxCompletionItems = 10
	completionWidth    = 50
)

var (
	completionBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1).
				Width(completionWidth)

	completionItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	completionSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("62")).
				Bold(true)

	completionKindStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("105")).
				Bold(true)

	completionDetailStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Italic(true)
)

// RequestCompletion requests completion items from LSP
func RequestCompletion(ctx context.Context, m *Model, params lsp.CompletionParams) ([]lsp.CompletionItem, error) {
	if !m.IsLSPEnabled() {
		return nil, fmt.Errorf("LSP not enabled")
	}

	items, err := m.lspClient.GetCompletion(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get completion: %w", err)
	}

	return items, nil
}

// RenderCompletion renders the completion popup
func RenderCompletion(m *Model) string {
	if !m.showCompletion || len(m.completionItems) == 0 {
		return ""
	}

	var sb strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("105")).
		Bold(true)
	sb.WriteString(headerStyle.Render("Completions"))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", completionWidth-4))
	sb.WriteString("\n")

	// Determine which items to show
	startIdx := 0
	endIdx := len(m.completionItems)
	if endIdx > maxCompletionItems {
		// Calculate window around selected item
		halfWindow := maxCompletionItems / 2
		startIdx = m.completionIndex - halfWindow
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + maxCompletionItems
		if endIdx > len(m.completionItems) {
			endIdx = len(m.completionItems)
			startIdx = endIdx - maxCompletionItems
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	// Render items
	for i := startIdx; i < endIdx; i++ {
		item := m.completionItems[i]
		icon := GetCompletionKindIcon(item.Kind)

		var line strings.Builder

		// Add icon and kind indicator
		line.WriteString(completionKindStyle.Render(icon))
		line.WriteString(" ")

		// Add label
		label := item.Label
		if len(label) > 30 {
			label = label[:27] + "..."
		}

		if i == m.completionIndex {
			line.WriteString(completionSelectedStyle.Render(label))
		} else {
			line.WriteString(completionItemStyle.Render(label))
		}

		// Add detail if available
		if item.Detail != "" {
			detail := item.Detail
			if len(detail) > 20 {
				detail = detail[:17] + "..."
			}
			line.WriteString(" ")
			line.WriteString(completionDetailStyle.Render(detail))
		}

		sb.WriteString(line.String())
		sb.WriteString("\n")
	}

	// Show scroll indicator if needed
	if len(m.completionItems) > maxCompletionItems {
		scrollInfo := fmt.Sprintf("%d/%d", m.completionIndex+1, len(m.completionItems))
		scrollStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true)
		sb.WriteString(strings.Repeat("─", completionWidth-4))
		sb.WriteString("\n")
		sb.WriteString(scrollStyle.Render(scrollInfo))
	}

	// Show hint
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Italic(true)
	sb.WriteString("\n")
	sb.WriteString(hintStyle.Render("Tab/↑↓ to navigate • Enter to select • Esc to cancel"))

	return completionBoxStyle.Render(sb.String())
}

// InsertCompletion inserts the selected completion into the input
func InsertCompletion(m *Model) {
	if !m.showCompletion || len(m.completionItems) == 0 {
		return
	}

	selected := m.GetSelectedCompletion()
	if selected == nil {
		return
	}

	// Get the text to insert
	insertText := selected.InsertText
	if insertText == "" {
		insertText = selected.Label
	}

	// Get current input value
	currentValue := m.textInput.Value()

	// Find the word being completed
	// For simplicity, we'll replace from the last word boundary
	lastSpace := strings.LastIndex(currentValue, " ")
	if lastSpace == -1 {
		// No spaces, replace entire input
		m.textInput.SetValue(insertText)
	} else {
		// Replace from last space
		prefix := currentValue[:lastSpace+1]
		m.textInput.SetValue(prefix + insertText)
	}

	// Clear completion popup
	m.ClearCompletion()
}

// GetCompletionKindIcon returns an icon for a completion item kind
func GetCompletionKindIcon(kind lsp.CompletionItemKind) string {
	switch kind {
	case lsp.CompletionItemKindFunction:
		return "ƒ"
	case lsp.CompletionItemKindMethod:
		return "m"
	case lsp.CompletionItemKindVariable:
		return "v"
	case lsp.CompletionItemKindConstant:
		return "c"
	case lsp.CompletionItemKindField:
		return "f"
	case lsp.CompletionItemKindClass:
		return "C"
	case lsp.CompletionItemKindInterface:
		return "I"
	case lsp.CompletionItemKindStruct:
		return "S"
	case lsp.CompletionItemKindModule:
		return "M"
	case lsp.CompletionItemKindProperty:
		return "p"
	case lsp.CompletionItemKindKeyword:
		return "k"
	case lsp.CompletionItemKindSnippet:
		return "s"
	case lsp.CompletionItemKindEnum:
		return "E"
	case lsp.CompletionItemKindEnumMember:
		return "e"
	default:
		return "•"
	}
}

// FilterCompletionItems filters completion items by prefix
func FilterCompletionItems(items []lsp.CompletionItem, prefix string) []lsp.CompletionItem {
	if prefix == "" {
		return items
	}

	prefix = strings.ToLower(prefix)
	filtered := make([]lsp.CompletionItem, 0, len(items))

	for _, item := range items {
		label := strings.ToLower(item.Label)
		filterText := strings.ToLower(item.FilterText)

		if strings.HasPrefix(label, prefix) || strings.HasPrefix(filterText, prefix) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// SortCompletionItems sorts completion items by relevance
func SortCompletionItems(items []lsp.CompletionItem) []lsp.CompletionItem {
	sorted := make([]lsp.CompletionItem, len(items))
	copy(sorted, items)

	sort.SliceStable(sorted, func(i, j int) bool {
		// If both have sort text, use it
		if sorted[i].SortText != "" && sorted[j].SortText != "" {
			return sorted[i].SortText < sorted[j].SortText
		}

		// If one has sort text, prioritize it
		if sorted[i].SortText != "" {
			return true
		}
		if sorted[j].SortText != "" {
			return false
		}

		// Fall back to alphabetical by label
		return sorted[i].Label < sorted[j].Label
	})

	return sorted
}

// TriggerCompletion triggers completion at the current cursor position
func TriggerCompletion(ctx context.Context, m *Model, documentURI string, line, char int) error {
	if !m.IsLSPEnabled() {
		return fmt.Errorf("LSP not enabled")
	}

	params := lsp.CompletionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: documentURI,
		},
		Position: lsp.Position{
			Line:      line,
			Character: char,
		},
	}

	items, err := RequestCompletion(ctx, m, params)
	if err != nil {
		return err
	}

	// Sort and set items
	sorted := SortCompletionItems(items)
	m.SetCompletionItems(sorted)

	return nil
}
