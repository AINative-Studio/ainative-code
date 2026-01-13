package dialogs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmDialog represents a Yes/No confirmation dialog
type ConfirmDialog struct {
	id          string
	title       string
	description string
	yesLabel    string
	noLabel     string
	defaultYes  bool
	selectedYes bool
	result      *bool
	closing     bool
	width       int
	height      int
}

// ConfirmDialogConfig contains configuration for a confirm dialog
type ConfirmDialogConfig struct {
	ID          string
	Title       string
	Description string
	YesLabel    string // Default: "Yes"
	NoLabel     string // Default: "No"
	DefaultYes  bool   // Default: false (No is selected)
}

// NewConfirmDialog creates a new confirm dialog
func NewConfirmDialog(config ConfirmDialogConfig) *ConfirmDialog {
	// Set defaults
	if config.YesLabel == "" {
		config.YesLabel = "Yes"
	}
	if config.NoLabel == "" {
		config.NoLabel = "No"
	}
	if config.ID == "" {
		config.ID = "confirm-dialog"
	}

	return &ConfirmDialog{
		id:          config.ID,
		title:       config.Title,
		description: config.Description,
		yesLabel:    config.YesLabel,
		noLabel:     config.NoLabel,
		defaultYes:  config.DefaultYes,
		selectedYes: config.DefaultYes,
		result:      nil,
		closing:     false,
		width:       80,
		height:      24,
	}
}

// Init initializes the dialog
func (d *ConfirmDialog) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (d *ConfirmDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "shift+tab":
			// Move to Yes
			d.selectedYes = true
			return d, nil

		case "right", "l", "tab":
			// Move to No
			d.selectedYes = false
			return d, nil

		case "y", "Y":
			// Quick Yes
			result := true
			d.result = &result
			d.closing = true
			return d, nil

		case "n", "N":
			// Quick No
			result := false
			d.result = &result
			d.closing = true
			return d, nil

		case "enter":
			// Confirm selection
			d.result = &d.selectedYes
			d.closing = true
			return d, nil

		case "esc":
			// Cancel - treat as No
			result := false
			d.result = &result
			d.closing = true
			return d, nil
		}
	}

	return d, nil
}

// View renders the dialog
func (d *ConfirmDialog) View() string {
	var content strings.Builder

	// Title
	content.WriteString(DialogTitleStyle.Render(d.title))
	content.WriteString("\n\n")

	// Description
	if d.description != "" {
		desc := DialogDescriptionStyle.Width(36).Render(d.description)
		content.WriteString(desc)
		content.WriteString("\n\n")
	}

	// Buttons
	var yesButton, noButton string
	if d.selectedYes {
		yesButton = ButtonActiveStyle.Render(d.yesLabel)
		noButton = ButtonInactiveStyle.Render(d.noLabel)
	} else {
		yesButton = ButtonInactiveStyle.Render(d.yesLabel)
		noButton = ButtonActiveStyle.Render(d.noLabel)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Left, yesButton, noButton)
	centeredButtons := lipgloss.NewStyle().
		Width(40).
		Align(lipgloss.Center).
		Render(buttons)
	content.WriteString(centeredButtons)
	content.WriteString("\n\n")

	// Help text
	helpText := HelpTextStyle.Width(40).Render("← → / Tab to switch • Enter to confirm • ESC to cancel")
	content.WriteString(helpText)

	// Wrap in dialog box
	return RenderDialogBox(content.String(), 44)
}

// ID returns the dialog ID
func (d *ConfirmDialog) ID() string {
	return d.id
}

// SetSize updates the dialog dimensions
func (d *ConfirmDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// IsClosing returns true if the dialog is requesting to be closed
func (d *ConfirmDialog) IsClosing() bool {
	return d.closing
}

// Result returns the dialog result
func (d *ConfirmDialog) Result() interface{} {
	return d.result
}

// GetResult returns the boolean result (convenience method)
func (d *ConfirmDialog) GetResult() *bool {
	return d.result
}
