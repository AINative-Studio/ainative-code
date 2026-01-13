package dialogs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DialogManager manages a stack of dialogs with focus management
type DialogManager struct {
	stack  []Dialog
	width  int
	height int
}

// NewDialogManager creates a new dialog manager
func NewDialogManager() *DialogManager {
	return &DialogManager{
		stack:  make([]Dialog, 0),
		width:  80,
		height: 24,
	}
}

// Update handles dialog-related messages
func (dm *DialogManager) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case OpenDialogMsg:
		// Check if dialog with same ID already exists
		if !dm.hasDialog(msg.Dialog.ID()) {
			// Set size for new dialog
			msg.Dialog.SetSize(dm.width, dm.height)
			// Add to stack
			dm.stack = append(dm.stack, msg.Dialog)
		}
		return nil

	case CloseDialogMsg:
		if msg.DialogID == "" {
			// Close top dialog
			cmd := dm.CloseTop()
			return cmd
		} else {
			// Close specific dialog
			cmd := dm.CloseByID(msg.DialogID)
			return cmd
		}

	case CloseAllDialogsMsg:
		dm.stack = make([]Dialog, 0)
		return nil

	case tea.KeyMsg:
		// Handle ESC to close top dialog
		if msg.String() == "esc" && len(dm.stack) > 0 {
			return dm.CloseTop()
		}

		// Forward key events to top dialog
		if len(dm.stack) > 0 {
			topDialog := dm.stack[len(dm.stack)-1]
			updatedDialog, cmd := topDialog.Update(msg)
			dm.stack[len(dm.stack)-1] = updatedDialog.(Dialog)
			cmds = append(cmds, cmd)

			// Check if dialog is requesting to close
			if topDialog.IsClosing() {
				closeCmd := dm.CloseTop()
				cmds = append(cmds, closeCmd)
			}
		}

	case tea.WindowSizeMsg:
		dm.SetSize(msg.Width, msg.Height)
		// Propagate resize to all dialogs
		for i := range dm.stack {
			dm.stack[i].SetSize(dm.width, dm.height)
		}
	}

	// If there are dialogs on the stack, forward the message to the top dialog
	if len(dm.stack) > 0 {
		topDialog := dm.stack[len(dm.stack)-1]
		updatedDialog, cmd := topDialog.Update(msg)
		dm.stack[len(dm.stack)-1] = updatedDialog.(Dialog)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// View renders all dialogs as layers
func (dm *DialogManager) View() string {
	if len(dm.stack) == 0 {
		return ""
	}

	// Start with backdrop
	backdrop := RenderBackdrop(dm.width, dm.height)

	// Layer each dialog on top
	layers := []string{backdrop}
	for _, dialog := range dm.stack {
		dialogView := dialog.View()
		layers = append(layers, dialogView)
	}

	// Use lipgloss.Place to center dialogs
	result := layers[0] // Start with backdrop
	for i := 1; i < len(layers); i++ {
		// Center each dialog
		centered := lipgloss.Place(
			dm.width,
			dm.height,
			lipgloss.Center,
			lipgloss.Center,
			layers[i],
		)
		// Layer on top of previous
		result = dm.layerStrings(result, centered)
	}

	return result
}

// layerStrings layers two strings on top of each other
// The top string's non-space characters override the base string
func (dm *DialogManager) layerStrings(base, top string) string {
	baseLines := strings.Split(base, "\n")
	topLines := strings.Split(top, "\n")

	maxLines := len(baseLines)
	if len(topLines) > maxLines {
		maxLines = len(topLines)
	}

	result := make([]string, maxLines)
	for i := 0; i < maxLines; i++ {
		var baseLine, topLine string
		if i < len(baseLines) {
			baseLine = baseLines[i]
		}
		if i < len(topLines) {
			topLine = topLines[i]
		}

		// Merge lines - top non-space chars override base
		result[i] = dm.mergeLines(baseLine, topLine)
	}

	return strings.Join(result, "\n")
}

// mergeLines merges two lines, with top's non-space characters taking priority
func (dm *DialogManager) mergeLines(base, top string) string {
	if top == "" {
		return base
	}
	if base == "" {
		return top
	}

	baseRunes := []rune(base)
	topRunes := []rune(top)

	// Ensure base is at least as long as top
	if len(baseRunes) < len(topRunes) {
		padding := make([]rune, len(topRunes)-len(baseRunes))
		for i := range padding {
			padding[i] = ' '
		}
		baseRunes = append(baseRunes, padding...)
	}

	// Overlay top onto base
	for i, r := range topRunes {
		if r != ' ' && r != '\x00' {
			baseRunes[i] = r
		}
	}

	return string(baseRunes)
}

// SetSize updates the manager dimensions
func (dm *DialogManager) SetSize(width, height int) {
	dm.width = width
	dm.height = height

	// Update all dialogs
	for i := range dm.stack {
		dm.stack[i].SetSize(width, height)
	}
}

// HasDialogs returns true if there are any dialogs open
func (dm *DialogManager) HasDialogs() bool {
	return len(dm.stack) > 0
}

// GetTop returns the top dialog, or nil if stack is empty
func (dm *DialogManager) GetTop() Dialog {
	if len(dm.stack) == 0 {
		return nil
	}
	return dm.stack[len(dm.stack)-1]
}

// GetCount returns the number of dialogs in the stack
func (dm *DialogManager) GetCount() int {
	return len(dm.stack)
}

// CloseTop closes the top dialog and returns its result
func (dm *DialogManager) CloseTop() tea.Cmd {
	if len(dm.stack) == 0 {
		return nil
	}

	// Get the top dialog
	topDialog := dm.stack[len(dm.stack)-1]

	// Remove from stack
	dm.stack = dm.stack[:len(dm.stack)-1]

	// Send result message
	return SendDialogResult(topDialog.ID(), topDialog.Result(), nil)
}

// CloseByID closes a specific dialog by ID
func (dm *DialogManager) CloseByID(id string) tea.Cmd {
	for i, dialog := range dm.stack {
		if dialog.ID() == id {
			// Remove from stack
			dm.stack = append(dm.stack[:i], dm.stack[i+1:]...)
			// Send result message
			return SendDialogResult(dialog.ID(), dialog.Result(), nil)
		}
	}
	return nil
}

// hasDialog checks if a dialog with the given ID exists in the stack
func (dm *DialogManager) hasDialog(id string) bool {
	for _, dialog := range dm.stack {
		if dialog.ID() == id {
			return true
		}
	}
	return false
}

// Clear removes all dialogs from the stack
func (dm *DialogManager) Clear() {
	dm.stack = make([]Dialog, 0)
}
