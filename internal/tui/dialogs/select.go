package dialogs

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SelectOption represents an option in the select dialog
type SelectOption struct {
	Label       string
	Value       string
	Description string
}

// SelectDialog represents a list selection dialog with search
type SelectDialog struct {
	id           string
	title        string
	description  string
	options      []SelectOption
	filteredOpts []SelectOption
	searchInput  textinput.Model
	selectedIdx  int
	searchMode   bool
	result       *string
	closing      bool
	width        int
	height       int
	maxVisible   int // Maximum visible items
}

// SelectDialogConfig contains configuration for a select dialog
type SelectDialogConfig struct {
	ID          string
	Title       string
	Description string
	Options     []SelectOption
	DefaultIdx  int  // Default selected index
	Searchable  bool // Enable search mode
}

// NewSelectDialog creates a new select dialog
func NewSelectDialog(config SelectDialogConfig) *SelectDialog {
	if config.ID == "" {
		config.ID = "select-dialog"
	}

	// Create search input
	ti := textinput.New()
	ti.Placeholder = "Type to search..."
	ti.CharLimit = 100
	ti.Width = 36

	// Start in search mode if searchable
	if config.Searchable {
		ti.Focus()
	}

	return &SelectDialog{
		id:           config.ID,
		title:        config.Title,
		description:  config.Description,
		options:      config.Options,
		filteredOpts: config.Options, // Initially show all
		searchInput:  ti,
		selectedIdx:  config.DefaultIdx,
		searchMode:   config.Searchable,
		result:       nil,
		closing:      false,
		width:        80,
		height:       24,
		maxVisible:   8, // Show max 8 items at a time
	}
}

// Init initializes the dialog
func (d *SelectDialog) Init() tea.Cmd {
	if d.searchMode {
		return textinput.Blink
	}
	return nil
}

// Update handles messages
func (d *SelectDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If in search mode, handle search-specific keys
		if d.searchMode {
			switch msg.String() {
			case "enter":
				// If search is empty or no results, exit search mode
				if strings.TrimSpace(d.searchInput.Value()) == "" || len(d.filteredOpts) == 0 {
					d.searchMode = false
					d.searchInput.Blur()
					return d, nil
				}
				// Otherwise, select the first filtered option
				if len(d.filteredOpts) > 0 {
					result := d.filteredOpts[d.selectedIdx].Value
					d.result = &result
					d.closing = true
					return d, nil
				}

			case "esc":
				// Exit search mode or cancel dialog
				if d.searchInput.Value() != "" {
					// Clear search
					d.searchInput.SetValue("")
					d.filteredOpts = d.options
					d.selectedIdx = 0
					return d, nil
				}
				// Cancel dialog
				d.result = nil
				d.closing = true
				return d, nil

			case "down", "ctrl+n":
				// Navigate down in filtered list
				if len(d.filteredOpts) > 0 {
					d.selectedIdx = (d.selectedIdx + 1) % len(d.filteredOpts)
				}
				return d, nil

			case "up", "ctrl+p":
				// Navigate up in filtered list
				if len(d.filteredOpts) > 0 {
					d.selectedIdx = (d.selectedIdx - 1 + len(d.filteredOpts)) % len(d.filteredOpts)
				}
				return d, nil

			case "tab":
				// Exit search mode
				d.searchMode = false
				d.searchInput.Blur()
				return d, nil

			default:
				// Update search input and filter options
				d.searchInput, cmd = d.searchInput.Update(msg)
				d.filterOptions()
				// Reset selection to first item
				d.selectedIdx = 0
				return d, cmd
			}
		} else {
			// Not in search mode - handle normal selection keys
			switch msg.String() {
			case "enter":
				// Select current option
				if len(d.filteredOpts) > 0 && d.selectedIdx < len(d.filteredOpts) {
					result := d.filteredOpts[d.selectedIdx].Value
					d.result = &result
					d.closing = true
				}
				return d, nil

			case "esc":
				// Cancel
				d.result = nil
				d.closing = true
				return d, nil

			case "down", "j":
				// Move down
				if len(d.filteredOpts) > 0 {
					d.selectedIdx = (d.selectedIdx + 1) % len(d.filteredOpts)
				}
				return d, nil

			case "up", "k":
				// Move up
				if len(d.filteredOpts) > 0 {
					d.selectedIdx = (d.selectedIdx - 1 + len(d.filteredOpts)) % len(d.filteredOpts)
				}
				return d, nil

			case "/":
				// Enter search mode
				d.searchMode = true
				d.searchInput.Focus()
				return d, textinput.Blink

			default:
				// Quick select by number
				if len(msg.String()) == 1 {
					num := int(msg.String()[0] - '0')
					if num > 0 && num <= len(d.filteredOpts) {
						result := d.filteredOpts[num-1].Value
						d.result = &result
						d.closing = true
						return d, nil
					}
				}
			}
		}
	}

	return d, nil
}

// filterOptions filters options based on search query
func (d *SelectDialog) filterOptions() {
	query := strings.ToLower(strings.TrimSpace(d.searchInput.Value()))
	if query == "" {
		d.filteredOpts = d.options
		return
	}

	filtered := make([]SelectOption, 0)
	for _, opt := range d.options {
		// Search in both label and description
		if strings.Contains(strings.ToLower(opt.Label), query) ||
			strings.Contains(strings.ToLower(opt.Description), query) {
			filtered = append(filtered, opt)
		}
	}
	d.filteredOpts = filtered
}

// View renders the dialog
func (d *SelectDialog) View() string {
	var content strings.Builder

	// Title
	content.WriteString(DialogTitleStyle.Render(d.title))
	content.WriteString("\n\n")

	// Description
	if d.description != "" {
		desc := DialogDescriptionStyle.Width(40).Render(d.description)
		content.WriteString(desc)
		content.WriteString("\n\n")
	}

	// Search box (if in search mode)
	if d.searchMode {
		searchBox := d.searchInput.View()
		searchStyled := InputFieldFocusedStyle.Render(searchBox)
		content.WriteString(searchStyled)
		content.WriteString("\n\n")
	}

	// Options list
	if len(d.filteredOpts) == 0 {
		noResults := ErrorTextStyle.Width(40).Render("No matching options")
		content.WriteString(noResults)
		content.WriteString("\n")
	} else {
		// Calculate visible range
		startIdx := 0
		endIdx := len(d.filteredOpts)
		if len(d.filteredOpts) > d.maxVisible {
			// Center the selected item
			halfVisible := d.maxVisible / 2
			startIdx = d.selectedIdx - halfVisible
			if startIdx < 0 {
				startIdx = 0
			}
			endIdx = startIdx + d.maxVisible
			if endIdx > len(d.filteredOpts) {
				endIdx = len(d.filteredOpts)
				startIdx = endIdx - d.maxVisible
				if startIdx < 0 {
					startIdx = 0
				}
			}
		}

		// Show scroll indicator if needed
		if startIdx > 0 {
			scrollUp := HelpTextStyle.Render("  ▲ More above")
			content.WriteString(scrollUp)
			content.WriteString("\n")
		}

		// Render visible options
		for i := startIdx; i < endIdx; i++ {
			opt := d.filteredOpts[i]
			prefix := "  "
			if i == d.selectedIdx {
				prefix = "▶ "
			}

			var optText string
			if opt.Description != "" {
				optText = opt.Label + " - " + opt.Description
			} else {
				optText = opt.Label
			}

			// Truncate if too long
			maxLen := 38
			if len(optText) > maxLen {
				optText = optText[:maxLen-3] + "..."
			}

			if i == d.selectedIdx {
				line := ListItemSelectedStyle.Width(40).Render(prefix + optText)
				content.WriteString(line)
			} else {
				line := ListItemStyle.Width(40).Render(prefix + optText)
				content.WriteString(line)
			}
			content.WriteString("\n")
		}

		// Show scroll indicator if needed
		if endIdx < len(d.filteredOpts) {
			scrollDown := HelpTextStyle.Render("  ▼ More below")
			content.WriteString(scrollDown)
			content.WriteString("\n")
		}
	}

	// Help text
	var helpText string
	if d.searchMode {
		helpText = "↑↓ navigate • Enter select • Tab exit search • ESC cancel"
	} else {
		helpText = "↑↓ / jk navigate • Enter select • / search • ESC cancel"
	}
	help := HelpTextStyle.Width(44).Render(helpText)
	content.WriteString("\n")
	content.WriteString(help)

	// Wrap in dialog box
	return RenderDialogBox(content.String(), 48)
}

// ID returns the dialog ID
func (d *SelectDialog) ID() string {
	return d.id
}

// SetSize updates the dialog dimensions
func (d *SelectDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
	// Adjust max visible based on height
	d.maxVisible = (height / 3) - 4
	if d.maxVisible < 3 {
		d.maxVisible = 3
	}
}

// IsClosing returns true if the dialog is requesting to be closed
func (d *SelectDialog) IsClosing() bool {
	return d.closing
}

// Result returns the dialog result
func (d *SelectDialog) Result() interface{} {
	return d.result
}

// GetResult returns the string result (convenience method)
func (d *SelectDialog) GetResult() *string {
	return d.result
}

// GetSelectedOption returns the selected option (convenience method)
func (d *SelectDialog) GetSelectedOption() *SelectOption {
	if d.result != nil {
		for _, opt := range d.options {
			if opt.Value == *d.result {
				return &opt
			}
		}
	}
	return nil
}
