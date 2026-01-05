package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// HelpState manages the help overlay state
type HelpState struct {
	Visible         bool
	FullHelp        bool // false = compact, true = full
	CurrentCategory string
	HelpModel       help.Model
}

// KeyBinding represents a keyboard shortcut
type KeyBinding struct {
	Key         string
	Description string
	Category    string // "navigation", "editing", "view", "system"
}

// NewHelpState creates a new help state
func NewHelpState() *HelpState {
	h := help.New()
	h.ShowAll = false

	return &HelpState{
		Visible:         false,
		FullHelp:        false,
		CurrentCategory: "all",
		HelpModel:       h,
	}
}

// Show displays the help overlay
func (h *HelpState) Show() {
	h.Visible = true
}

// Hide hides the help overlay
func (h *HelpState) Hide() {
	h.Visible = false
}

// Toggle toggles the help overlay
func (h *HelpState) Toggle() {
	h.Visible = !h.Visible
}

// ToggleFullHelp toggles between compact and full help
func (h *HelpState) ToggleFullHelp() {
	h.FullHelp = !h.FullHelp
	h.HelpModel.ShowAll = h.FullHelp
}

// SetCategory sets the current help category
func (h *HelpState) SetCategory(category string) {
	h.CurrentCategory = category
}

// GetAllKeyBindings returns all available key bindings
func GetAllKeyBindings() []KeyBinding {
	return []KeyBinding{
		// Navigation
		{Key: "↑/↓", Description: "Scroll up/down", Category: "navigation"},
		{Key: "PgUp/PgDn", Description: "Page up/down", Category: "navigation"},
		{Key: "Home", Description: "Jump to top", Category: "navigation"},
		{Key: "End", Description: "Jump to bottom", Category: "navigation"},
		{Key: "Ctrl+U", Description: "Scroll half page up", Category: "navigation"},
		{Key: "Ctrl+D", Description: "Scroll half page down", Category: "navigation"},

		// Editing
		{Key: "Enter", Description: "Send message", Category: "editing"},
		{Key: "Ctrl+L", Description: "Clear conversation", Category: "editing"},
		{Key: "Esc", Description: "Cancel input/close overlay", Category: "editing"},

		// View
		{Key: "t", Description: "Toggle thinking display", Category: "view"},
		{Key: "e", Description: "Expand all thinking blocks", Category: "view"},
		{Key: "c", Description: "Collapse all thinking blocks", Category: "view"},
		{Key: "Ctrl+R", Description: "Refresh display", Category: "view"},
		{Key: "Ctrl+T", Description: "Toggle theme", Category: "view"},

		// Help & System
		{Key: "?", Description: "Show/hide this help", Category: "help"},
		{Key: "Ctrl+H", Description: "Toggle compact/full help", Category: "help"},
		{Key: "Ctrl+C", Description: "Quit application", Category: "system"},
		{Key: "Ctrl+Z", Description: "Suspend (background)", Category: "system"},
	}
}

// GetKeyBindingsByCategory returns key bindings filtered by category
func GetKeyBindingsByCategory(category string) []KeyBinding {
	allBindings := GetAllKeyBindings()

	if category == "all" {
		return allBindings
	}

	var filtered []KeyBinding
	for _, binding := range allBindings {
		if binding.Category == category {
			filtered = append(filtered, binding)
		}
	}

	return filtered
}

// RenderHelp renders the help overlay
func (h *HelpState) RenderHelp(width, height int) string {
	if !h.Visible {
		return ""
	}

	if width < 40 || height < 10 {
		return h.renderCompactHelp(width, height)
	}

	if h.FullHelp {
		return h.renderFullHelp(width, height)
	}

	return h.renderCategorizedHelp(width, height)
}

// renderCompactHelp renders a minimal help view for small terminals
func (h *HelpState) renderCompactHelp(width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Align(lipgloss.Center)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	var lines []string
	lines = append(lines, titleStyle.Render("Keyboard Shortcuts"))
	lines = append(lines, "")

	// Show only essential shortcuts in compact mode
	essentialBindings := []KeyBinding{
		{Key: "Enter", Description: "Send message"},
		{Key: "↑/↓", Description: "Scroll"},
		{Key: "?", Description: "Toggle help"},
		{Key: "Ctrl+C", Description: "Quit"},
	}

	for _, binding := range essentialBindings {
		line := fmt.Sprintf("%s  %s",
			keyStyle.Render(fmt.Sprintf("%-10s", binding.Key)),
			descStyle.Render(binding.Description))
		lines = append(lines, line)
	}

	lines = append(lines, "")
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	lines = append(lines, hintStyle.Render("Press Ctrl+H for full help"))

	content := strings.Join(lines, "\n")

	// Center the help in the available space
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Align(lipgloss.Center)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, boxStyle.Render(content))
}

// renderCategorizedHelp renders help organized by categories
func (h *HelpState) renderCategorizedHelp(width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Align(lipgloss.Center)

	categoryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Bold(true).
		Underline(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	var lines []string
	lines = append(lines, titleStyle.Render("Keyboard Shortcuts"))
	lines = append(lines, "")

	categories := []string{"navigation", "editing", "view", "help", "system"}
	categoryNames := map[string]string{
		"navigation": "Navigation",
		"editing":    "Editing",
		"view":       "View Options",
		"help":       "Help & Info",
		"system":     "System",
	}

	for _, cat := range categories {
		bindings := GetKeyBindingsByCategory(cat)
		if len(bindings) == 0 {
			continue
		}

		lines = append(lines, categoryStyle.Render(categoryNames[cat]))

		for _, binding := range bindings {
			line := fmt.Sprintf("  %s  %s",
				keyStyle.Render(fmt.Sprintf("%-12s", binding.Key)),
				descStyle.Render(binding.Description))
			lines = append(lines, line)
		}

		lines = append(lines, "")
	}

	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	lines = append(lines, hintStyle.Render("Press ? or ESC to close | Ctrl+H for compact view"))

	content := strings.Join(lines, "\n")

	// Create scrollable content if needed
	if len(lines) > height-6 {
		content = h.truncateContent(lines, height-6)
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(width - 4).
		Height(height - 2)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, boxStyle.Render(content))
}

// renderFullHelp renders complete help with all details
func (h *HelpState) renderFullHelp(width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Align(lipgloss.Center)

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("13")).
		Bold(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	var lines []string

	// Header
	lines = append(lines, titleStyle.Render("AINative-Code - Complete Keyboard Reference"))
	lines = append(lines, "")

	// About section
	lines = append(lines, sectionStyle.Render("About"))
	lines = append(lines, descStyle.Render("AINative-Code is a terminal-based AI coding assistant."))
	lines = append(lines, descStyle.Render("Use keyboard shortcuts for efficient navigation and control."))
	lines = append(lines, "")

	// Shortcuts by category
	categories := []string{"navigation", "editing", "view", "help", "system"}
	categoryNames := map[string]string{
		"navigation": "Navigation Shortcuts",
		"editing":    "Editing & Input",
		"view":       "View & Display Options",
		"help":       "Help & Information",
		"system":     "System Controls",
	}

	for _, cat := range categories {
		bindings := GetKeyBindingsByCategory(cat)
		if len(bindings) == 0 {
			continue
		}

		lines = append(lines, sectionStyle.Render(categoryNames[cat]))

		for _, binding := range bindings {
			line := fmt.Sprintf("  %s  %s",
				keyStyle.Render(fmt.Sprintf("%-15s", binding.Key)),
				descStyle.Render(binding.Description))
			lines = append(lines, line)
		}

		lines = append(lines, "")
	}

	// Tips section
	lines = append(lines, sectionStyle.Render("Tips"))
	tipStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	lines = append(lines, tipStyle.Render("• Use arrow keys or PgUp/PgDn to scroll through long conversations"))
	lines = append(lines, tipStyle.Render("• Press 't' to toggle thinking blocks for Claude's extended thinking"))
	lines = append(lines, tipStyle.Render("• Use Ctrl+L to clear conversation and start fresh"))
	lines = append(lines, tipStyle.Render("• Mouse wheel scrolling is supported in compatible terminals"))
	lines = append(lines, "")

	// Footer
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	lines = append(lines, hintStyle.Render("Press ? or ESC to close help | Ctrl+H for compact view"))

	content := strings.Join(lines, "\n")

	// Create scrollable content if needed
	if len(lines) > height-6 {
		content = h.truncateContent(lines, height-6)
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 3).
		Width(width - 6).
		Height(height - 2)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, boxStyle.Render(content))
}

// truncateContent truncates content to fit within maxLines
func (h *HelpState) truncateContent(lines []string, maxLines int) string {
	if len(lines) <= maxLines {
		return strings.Join(lines, "\n")
	}

	truncated := lines[:maxLines-1]
	moreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	truncated = append(truncated, moreStyle.Render(fmt.Sprintf("... (%d more lines)", len(lines)-maxLines+1)))

	return strings.Join(truncated, "\n")
}

// RenderInlineHelp renders a single-line help hint
func RenderInlineHelp(hints []string) string {
	if len(hints) == 0 {
		return ""
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	return style.Render(strings.Join(hints, " • "))
}

// GetContextualHelp returns context-sensitive help text
func GetContextualHelp(mode string, streaming bool) []string {
	var hints []string

	if streaming {
		hints = append(hints, "Streaming in progress...")
		return hints
	}

	switch mode {
	case "chat":
		hints = append(hints, "Enter: send", "↑/↓: scroll", "?: help")
	case "help":
		hints = append(hints, "ESC: close", "Ctrl+H: toggle view")
	default:
		hints = append(hints, "?: help", "Ctrl+C: quit")
	}

	return hints
}

// Key bindings for bubble tea help model
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding
	Enter    key.Binding
	Help     key.Binding
	Quit     key.Binding
}

// ShortHelp returns a short help text
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Help, k.Quit}
}

// FullHelp returns full help text
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Home, k.End, k.Enter},
		{k.Help, k.Quit},
	}
}

// DefaultKeyMap returns the default key map
func DefaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("pgup/b", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "f"),
			key.WithHelp("pgdn/f", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "bottom"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "send"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}
