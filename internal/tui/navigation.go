package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/tui/components"
	"github.com/AINative-studio/ainative-code/pkg/lsp"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	maxNavigationResults = 20
	navigationWidth      = 70
)

var (
	navigationBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("141")).
				Padding(1, 2).
				Width(navigationWidth)

	navigationHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("141")).
				Bold(true)

	navigationItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	navigationFileStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("141")).
				Bold(true)

	navigationLineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("105"))

	navigationHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Italic(true)
)

// RequestDefinition requests definition locations from LSP
func RequestDefinition(ctx context.Context, m *Model, params lsp.DefinitionParams) ([]lsp.Location, error) {
	if !m.IsLSPEnabled() {
		return nil, fmt.Errorf("LSP not enabled")
	}

	locations, err := m.lspClient.GetDefinition(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get definition: %w", err)
	}

	return locations, nil
}

// RequestReferences requests reference locations from LSP
func RequestReferences(ctx context.Context, m *Model, params lsp.ReferencesParams) ([]lsp.Location, error) {
	if !m.IsLSPEnabled() {
		return nil, fmt.Errorf("LSP not enabled")
	}

	locations, err := m.lspClient.GetReferences(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
	}

	return locations, nil
}

// RenderNavigation renders the navigation results popup
func RenderNavigation(m *Model) string {
	if !m.showNavigation || len(m.navigationResult) == 0 {
		return ""
	}

	var sb strings.Builder

	// Header
	header := "Navigation Results"
	if len(m.navigationResult) > maxNavigationResults {
		header = fmt.Sprintf("Navigation Results (showing %d of %d)", maxNavigationResults, len(m.navigationResult))
	}
	sb.WriteString(navigationHeaderStyle.Render(header))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", navigationWidth-4))
	sb.WriteString("\n\n")

	// Group results by file
	fileGroups := make(map[string][]lsp.Location)
	for _, location := range m.navigationResult {
		fileName := ExtractFileName(location.URI)
		fileGroups[fileName] = append(fileGroups[fileName], location)
	}

	// Render results
	count := 0
	for fileName, locations := range fileGroups {
		if count >= maxNavigationResults {
			break
		}

		// File header
		sb.WriteString(navigationFileStyle.Render(fileName))
		sb.WriteString("\n")

		// Locations in this file
		for _, location := range locations {
			if count >= maxNavigationResults {
				break
			}

			line := location.Range.Start.Line + 1 // Convert to 1-indexed
			char := location.Range.Start.Character

			// Format: "  Line 42:10"
			sb.WriteString("  ")
			sb.WriteString(navigationLineStyle.Render(fmt.Sprintf("Line %d:%d", line, char)))
			sb.WriteString("\n")

			count++
		}

		sb.WriteString("\n")
	}

	// Show hint
	sb.WriteString(strings.Repeat("─", navigationWidth-4))
	sb.WriteString("\n")
	sb.WriteString(navigationHintStyle.Render("Press Esc to close"))

	return navigationBoxStyle.Render(sb.String())
}

// FormatLocation formats a location for display
func FormatLocation(location lsp.Location) string {
	fileName := ExtractFileName(location.URI)
	line := location.Range.Start.Line + 1 // Convert to 1-indexed
	char := location.Range.Start.Character

	return fmt.Sprintf("%s:%d:%d", fileName, line, char)
}

// ExtractFileName extracts the file name from a URI
func ExtractFileName(uri string) string {
	// Remove file:// scheme if present
	path := strings.TrimPrefix(uri, "file://")

	// Get base name
	return filepath.Base(path)
}

// ExtractFilePath extracts the full file path from a URI
func ExtractFilePath(uri string) string {
	// Remove file:// scheme if present
	return strings.TrimPrefix(uri, "file://")
}

// GotoDefinition navigates to the definition of a symbol
func GotoDefinition(ctx context.Context, m *Model, documentURI string, line, char int) error {
	if !m.IsLSPEnabled() {
		return fmt.Errorf("LSP not enabled")
	}

	params := lsp.DefinitionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: documentURI,
		},
		Position: lsp.Position{
			Line:      line,
			Character: char,
		},
	}

	locations, err := RequestDefinition(ctx, m, params)
	if err != nil {
		return err
	}

	if len(locations) == 0 {
		// No definition found
		m.ClearNavigation()
		return nil
	}

	m.SetNavigationResult(locations)
	return nil
}

// FindReferences finds all references to a symbol
func FindReferences(ctx context.Context, m *Model, documentURI string, line, char int, includeDeclaration bool) error {
	if !m.IsLSPEnabled() {
		return fmt.Errorf("LSP not enabled")
	}

	params := lsp.ReferencesParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: documentURI,
		},
		Position: lsp.Position{
			Line:      line,
			Character: char,
		},
		Context: lsp.ReferencesContext{
			IncludeDeclaration: includeDeclaration,
		},
	}

	locations, err := RequestReferences(ctx, m, params)
	if err != nil {
		return err
	}

	if len(locations) == 0 {
		// No references found
		m.ClearNavigation()
		return nil
	}

	m.SetNavigationResult(locations)
	return nil
}

// NavigationPopup is a wrapper that makes Model's navigation functionality
// implement the PopupComponent interface. This is NON-BREAKING - it only
// adds new methods and doesn't modify existing behavior.
type NavigationPopup struct {
	model *Model
	*components.PopupAdapter
}

// NewNavigationPopup creates a new navigation popup wrapper around a Model.
func NewNavigationPopup(m *Model) *NavigationPopup {
	adapter := components.NewPopupAdapter()
	adapter.SetSize(navigationWidth, maxNavigationResults+5)
	return &NavigationPopup{
		model:        m,
		PopupAdapter: adapter,
	}
}

// Init implements Component interface.
func (n *NavigationPopup) Init() tea.Cmd {
	return n.PopupAdapter.Init()
}

// Update implements Component interface.
func (n *NavigationPopup) Update(msg tea.Msg) (components.Component, tea.Cmd) {
	return n, nil
}

// View implements Component interface.
func (n *NavigationPopup) View() string {
	return n.RenderPopup()
}

// RenderPopup implements PopupComponent interface.
func (n *NavigationPopup) RenderPopup() string {
	return RenderNavigation(n.model)
}

// IsVisible implements Stateful interface.
func (n *NavigationPopup) IsVisible() bool {
	return n.model.GetShowNavigation()
}

// Show implements Stateful interface.
func (n *NavigationPopup) Show() {
	// This would be controlled by the model's navigation logic
	n.PopupAdapter.Show()
}

// Hide implements Stateful interface.
func (n *NavigationPopup) Hide() {
	n.model.ClearNavigation()
	n.PopupAdapter.Hide()
}

// AsNavigationPopup returns the navigation as a PopupComponent interface.
// This allows existing code to work with the new interface without changes.
func (m *Model) AsNavigationPopup() components.PopupComponent {
	return NewNavigationPopup(m)
}

// Ensure NavigationPopup implements PopupComponent interface
var _ components.Component = (*NavigationPopup)(nil)
var _ components.PopupComponent = (*NavigationPopup)(nil)
var _ components.Stateful = (*NavigationPopup)(nil)
var _ components.Sizeable = (*NavigationPopup)(nil)
