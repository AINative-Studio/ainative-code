package dialogs

import tea "github.com/charmbracelet/bubbletea"

// Dialog message types for Bubble Tea event handling

// Dialog represents the interface that all dialogs must implement
type Dialog interface {
	tea.Model
	// ID returns a unique identifier for this dialog
	ID() string
	// SetSize updates the dialog dimensions
	SetSize(width, height int)
	// IsClosing returns true if the dialog is requesting to be closed
	IsClosing() bool
	// Result returns the dialog result (if any)
	Result() interface{}
}

// OpenDialogMsg signals to open a new dialog
type OpenDialogMsg struct {
	Dialog Dialog
}

// CloseDialogMsg signals to close the top dialog
type CloseDialogMsg struct {
	// DialogID is optional - if empty, closes the top dialog
	DialogID string
}

// CloseAllDialogsMsg signals to close all dialogs
type CloseAllDialogsMsg struct{}

// DialogResultMsg is sent when a dialog completes with a result
type DialogResultMsg struct {
	DialogID string
	Result   interface{}
	Error    error
}

// Helper functions to create commands

// OpenDialog creates a command to open a dialog
func OpenDialog(dialog Dialog) tea.Cmd {
	return func() tea.Msg {
		return OpenDialogMsg{Dialog: dialog}
	}
}

// CloseDialog creates a command to close a dialog
func CloseDialog(dialogID string) tea.Cmd {
	return func() tea.Msg {
		return CloseDialogMsg{DialogID: dialogID}
	}
}

// CloseTopDialog creates a command to close the top dialog
func CloseTopDialog() tea.Cmd {
	return func() tea.Msg {
		return CloseDialogMsg{DialogID: ""}
	}
}

// CloseAllDialogs creates a command to close all dialogs
func CloseAllDialogs() tea.Cmd {
	return func() tea.Msg {
		return CloseAllDialogsMsg{}
	}
}

// SendDialogResult creates a command to send a dialog result
func SendDialogResult(dialogID string, result interface{}, err error) tea.Cmd {
	return func() tea.Msg {
		return DialogResultMsg{
			DialogID: dialogID,
			Result:   result,
			Error:    err,
		}
	}
}
