package dialogs

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// InputDialog represents a text input dialog with validation
type InputDialog struct {
	id          string
	title       string
	description string
	textInput   textinput.Model
	validator   func(string) error
	result      *string
	errorMsg    string
	closing     bool
	width       int
	height      int
}

// InputDialogConfig contains configuration for an input dialog
type InputDialogConfig struct {
	ID           string
	Title        string
	Description  string
	Placeholder  string
	DefaultValue string
	Validator    func(string) error // Optional validation function
}

// NewInputDialog creates a new input dialog
func NewInputDialog(config InputDialogConfig) *InputDialog {
	if config.ID == "" {
		config.ID = "input-dialog"
	}
	if config.Placeholder == "" {
		config.Placeholder = "Enter text..."
	}

	// Create text input
	ti := textinput.New()
	ti.Placeholder = config.Placeholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 36
	ti.SetValue(config.DefaultValue)

	return &InputDialog{
		id:          config.ID,
		title:       config.Title,
		description: config.Description,
		textInput:   ti,
		validator:   config.Validator,
		result:      nil,
		errorMsg:    "",
		closing:     false,
		width:       80,
		height:      24,
	}
}

// Init initializes the dialog
func (d *InputDialog) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (d *InputDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Validate input
			value := strings.TrimSpace(d.textInput.Value())
			if d.validator != nil {
				if err := d.validator(value); err != nil {
					d.errorMsg = err.Error()
					return d, nil
				}
			}

			// Accept input
			d.result = &value
			d.closing = true
			return d, nil

		case "esc":
			// Cancel
			d.result = nil
			d.closing = true
			return d, nil

		default:
			// Clear error on any other key
			d.errorMsg = ""
		}
	}

	// Update text input
	d.textInput, cmd = d.textInput.Update(msg)
	return d, cmd
}

// View renders the dialog
func (d *InputDialog) View() string {
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

	// Input field
	inputView := d.textInput.View()
	inputStyled := InputFieldFocusedStyle.Render(inputView)
	content.WriteString(inputStyled)
	content.WriteString("\n")

	// Error message (if any)
	if d.errorMsg != "" {
		errorText := ErrorTextStyle.Width(36).Render("⚠ " + d.errorMsg)
		content.WriteString("\n")
		content.WriteString(errorText)
		content.WriteString("\n")
	}

	// Help text
	helpText := HelpTextStyle.Width(40).Render("Enter to submit • ESC to cancel")
	content.WriteString("\n")
	content.WriteString(helpText)

	// Wrap in dialog box
	return RenderDialogBox(content.String(), 44)
}

// ID returns the dialog ID
func (d *InputDialog) ID() string {
	return d.id
}

// SetSize updates the dialog dimensions
func (d *InputDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// IsClosing returns true if the dialog is requesting to be closed
func (d *InputDialog) IsClosing() bool {
	return d.closing
}

// Result returns the dialog result
func (d *InputDialog) Result() interface{} {
	return d.result
}

// GetResult returns the string result (convenience method)
func (d *InputDialog) GetResult() *string {
	return d.result
}
