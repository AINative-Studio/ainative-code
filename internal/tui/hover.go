package tui

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/tui/components"
	"github.com/AINative-studio/ainative-code/pkg/lsp"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	hoverWidth  = 60
	hoverHeight = 20
)

var (
	hoverBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("105")).
			Padding(1, 2).
			Width(hoverWidth)

	hoverHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("105")).
				Bold(true)

	hoverCodeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("120")).
			Background(lipgloss.Color("235"))

	hoverTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	hoverHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true)
)

// CodeBlock represents a code block extracted from markdown
type CodeBlock struct {
	Language string
	Code     string
}

// RequestHover requests hover information from LSP
func RequestHover(ctx context.Context, m *Model, params lsp.HoverParams) (*lsp.Hover, error) {
	if !m.IsLSPEnabled() {
		return nil, fmt.Errorf("LSP not enabled")
	}

	hover, err := m.lspClient.GetHover(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get hover info: %w", err)
	}

	return hover, nil
}

// RenderHover renders the hover information popup
func RenderHover(m *Model) string {
	if !m.showHover || m.hoverInfo == nil {
		return ""
	}

	var sb strings.Builder

	// Header
	sb.WriteString(hoverHeaderStyle.Render("Type Information"))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", hoverWidth-4))
	sb.WriteString("\n\n")

	// Format and render content
	formatted := FormatHoverContent(m.hoverInfo.Contents)
	sb.WriteString(formatted)

	// Add hint
	sb.WriteString("\n\n")
	sb.WriteString(strings.Repeat("─", hoverWidth-4))
	sb.WriteString("\n")
	sb.WriteString(hoverHintStyle.Render("Press Esc to close"))

	return hoverBoxStyle.Render(sb.String())
}

// FormatHoverContent formats hover content based on its type
func FormatHoverContent(content lsp.MarkupContent) string {
	if content.Value == "" {
		return ""
	}

	switch content.Kind {
	case "markdown":
		return formatMarkdown(content.Value)
	case "plaintext":
		return wrapText(content.Value, hoverWidth-8)
	default:
		return wrapText(content.Value, hoverWidth-8)
	}
}

// formatMarkdown formats markdown content for display
func formatMarkdown(markdown string) string {
	var sb strings.Builder

	// Extract code blocks
	codeBlocks := ExtractCodeBlocks(markdown)

	// Remove code blocks from markdown temporarily
	text := markdown
	for _, block := range codeBlocks {
		blockText := fmt.Sprintf("```%s\n%s\n```", block.Language, block.Code)
		text = strings.Replace(text, blockText, "{{CODE_BLOCK}}", 1)
	}

	// Split by lines and process
	lines := strings.Split(text, "\n")
	codeBlockIndex := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "{{CODE_BLOCK}}" {
			// Insert code block
			if codeBlockIndex < len(codeBlocks) {
				block := codeBlocks[codeBlockIndex]
				sb.WriteString(hoverCodeStyle.Render(block.Code))
				sb.WriteString("\n")
				codeBlockIndex++
			}
			continue
		}

		if trimmed == "" {
			sb.WriteString("\n")
			continue
		}

		// Format bold text (**text**)
		boldRegex := regexp.MustCompile(`\*\*([^*]+)\*\*`)
		line = boldRegex.ReplaceAllStringFunc(line, func(match string) string {
			text := strings.Trim(match, "*")
			boldStyle := lipgloss.NewStyle().Bold(true)
			return boldStyle.Render(text)
		})

		// Format italic text (*text*)
		italicRegex := regexp.MustCompile(`\*([^*]+)\*`)
		line = italicRegex.ReplaceAllStringFunc(line, func(match string) string {
			text := strings.Trim(match, "*")
			italicStyle := lipgloss.NewStyle().Italic(true)
			return italicStyle.Render(text)
		})

		// Wrap and render
		wrapped := wrapText(line, hoverWidth-8)
		sb.WriteString(hoverTextStyle.Render(wrapped))
		sb.WriteString("\n")
	}

	return sb.String()
}

// ExtractCodeBlocks extracts code blocks from markdown
func ExtractCodeBlocks(markdown string) []CodeBlock {
	blocks := []CodeBlock{}

	// Regex to match code blocks: ```lang\ncode\n```
	codeBlockRegex := regexp.MustCompile("```([a-zA-Z0-9]*)\n([\\s\\S]*?)```")
	matches := codeBlockRegex.FindAllStringSubmatch(markdown, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			blocks = append(blocks, CodeBlock{
				Language: match[1],
				Code:     strings.TrimSpace(match[2]),
			})
		}
	}

	return blocks
}

// wrapText wraps text at the specified width
func wrapText(text string, width int) string {
	if width <= 0 {
		width = 60
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		// If adding this word would exceed width, start new line
		if currentLine.Len() > 0 && currentLine.Len()+len(word)+1 > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}

		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	// Add the last line
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}

// TriggerHover triggers hover information at the current position
func TriggerHover(ctx context.Context, m *Model, documentURI string, line, char int) error {
	if !m.IsLSPEnabled() {
		return fmt.Errorf("LSP not enabled")
	}

	params := lsp.HoverParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: documentURI,
		},
		Position: lsp.Position{
			Line:      line,
			Character: char,
		},
	}

	hover, err := RequestHover(ctx, m, params)
	if err != nil {
		return err
	}

	if hover == nil {
		// No hover information at this position
		m.ClearHover()
		return nil
	}

	m.SetHoverInfo(hover)
	return nil
}

// HoverPopup is a wrapper that makes Model's hover functionality
// implement the PopupComponent interface. This is NON-BREAKING - it only
// adds new methods and doesn't modify existing behavior.
type HoverPopup struct {
	model *Model
	*components.PopupAdapter
}

// NewHoverPopup creates a new hover popup wrapper around a Model.
func NewHoverPopup(m *Model) *HoverPopup {
	adapter := components.NewPopupAdapter()
	adapter.SetSize(hoverWidth, hoverHeight)
	return &HoverPopup{
		model:        m,
		PopupAdapter: adapter,
	}
}

// Init implements Component interface.
func (h *HoverPopup) Init() tea.Cmd {
	return h.PopupAdapter.Init()
}

// Update implements Component interface.
func (h *HoverPopup) Update(msg tea.Msg) (components.Component, tea.Cmd) {
	return h, nil
}

// View implements Component interface.
func (h *HoverPopup) View() string {
	return h.RenderPopup()
}

// RenderPopup implements PopupComponent interface.
func (h *HoverPopup) RenderPopup() string {
	return RenderHover(h.model)
}

// IsVisible implements Stateful interface.
func (h *HoverPopup) IsVisible() bool {
	return h.model.GetShowHover()
}

// Show implements Stateful interface.
func (h *HoverPopup) Show() {
	// This would be controlled by the model's hover logic
	h.PopupAdapter.Show()
}

// Hide implements Stateful interface.
func (h *HoverPopup) Hide() {
	h.model.ClearHover()
	h.PopupAdapter.Hide()
}

// AsHoverPopup returns the hover as a PopupComponent interface.
// This allows existing code to work with the new interface without changes.
func (m *Model) AsHoverPopup() components.PopupComponent {
	return NewHoverPopup(m)
}

// Ensure HoverPopup implements PopupComponent interface
var _ components.Component = (*HoverPopup)(nil)
var _ components.PopupComponent = (*HoverPopup)(nil)
var _ components.Stateful = (*HoverPopup)(nil)
var _ components.Sizeable = (*HoverPopup)(nil)
