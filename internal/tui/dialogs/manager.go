package dialogs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DialogManager manages a stack of modals with advanced features
type DialogManager struct {
	stack         []*Modal          // Stack of modals (top = last)
	idMap         map[string]int    // Map of dialog ID to stack index
	width         int               // Container width
	height        int               // Container height
	shortcuts     *ShortcutManager  // Global keyboard shortcuts
	focusTrap     *FocusTrap        // Focus trap for current modal
	nextZIndex    int               // Next auto-assigned z-index
	baseZIndex    int               // Base z-index (default: 100)
}

// NewDialogManager creates a new dialog manager
func NewDialogManager() *DialogManager {
	return &DialogManager{
		stack:      make([]*Modal, 0),
		idMap:      make(map[string]int),
		width:      80,
		height:     24,
		shortcuts:  NewShortcutManager(),
		focusTrap:  NewFocusTrap(),
		nextZIndex: 100,
		baseZIndex: 100,
	}
}

// Update handles dialog-related messages
func (dm *DialogManager) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Handle shortcut keys first (global)
	if handled, cmd := dm.shortcuts.HandleMessage(msg); handled {
		return cmd
	}

	switch msg := msg.(type) {
	case OpenDialogMsg:
		// Open dialog with default config
		config := DefaultModalConfig()
		return dm.OpenModal(msg.Dialog, config)

	case OpenModalMsg:
		// Open modal with custom config
		return dm.OpenModal(msg.Dialog, msg.Config)

	case CloseDialogMsg:
		if msg.DialogID == "" {
			// Close top dialog
			return dm.CloseTop()
		}
		// Close specific dialog
		return dm.CloseByID(msg.DialogID)

	case CloseAllDialogsMsg:
		dm.Clear()
		return nil

	case tea.KeyMsg:
		keyStr := msg.String()

		// Handle focus trap first
		if dm.focusTrap.IsActive() {
			if handled, nextFocus := dm.focusTrap.HandleKey(keyStr); handled {
				// Focus trap handled the key
				// In a real implementation, we would focus the next element
				_ = nextFocus
				return nil
			}
		}

		// Handle ESC to close top modal
		if keyStr == "esc" && len(dm.stack) > 0 {
			topModal := dm.stack[len(dm.stack)-1]
			if topModal.ShouldCloseOnEsc() {
				return dm.CloseTop()
			}
		}

		// Forward key events to top modal
		if len(dm.stack) > 0 {
			topModal := dm.stack[len(dm.stack)-1]
			updatedModal, cmd := topModal.Update(msg)
			dm.stack[len(dm.stack)-1] = updatedModal.(*Modal)
			cmds = append(cmds, cmd)

			// Check if dialog is requesting to close
			if topModal.IsClosing() {
				closeCmd := dm.CloseTop()
				cmds = append(cmds, closeCmd)
			}
		}

	case tea.WindowSizeMsg:
		dm.SetSize(msg.Width, msg.Height)
		// Propagate resize to all modals
		for _, modal := range dm.stack {
			modal.SetSize(dm.width, dm.height)
			modal.CalculatePosition(dm.width, dm.height)
		}
	}

	// Forward message to top modal if it exists
	if len(dm.stack) > 0 {
		topModal := dm.stack[len(dm.stack)-1]
		updatedModal, cmd := topModal.Update(msg)
		dm.stack[len(dm.stack)-1] = updatedModal.(*Modal)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// View renders all modals as layers
func (dm *DialogManager) View() string {
	if len(dm.stack) == 0 {
		return ""
	}

	// Sort modals by z-index (lowest to highest)
	sortedModals := dm.getSortedModals()

	// Start with empty base
	result := ""
	hasBackdrop := false

	for _, modal := range sortedModals {
		if !modal.IsVisible() {
			continue
		}

		// Render backdrop if this is the first modal with backdrop
		if !hasBackdrop && modal.GetConfig().Backdrop.Enabled {
			backdrop := NewBackdropRenderer(dm.width, dm.height, modal.GetConfig().Backdrop)
			backdropView := backdrop.Render()
			if result == "" {
				result = backdropView
			} else {
				result = dm.layerStrings(result, backdropView)
			}
			hasBackdrop = true
		}

		// Render modal
		modalView := modal.View()
		x, y := modal.GetPosition()

		// Position the modal
		positioned := dm.positionModal(modalView, x, y)

		// Layer on top
		if result == "" {
			result = positioned
		} else {
			result = dm.layerStrings(result, positioned)
		}
	}

	return result
}

// positionModal positions a modal at the given coordinates
func (dm *DialogManager) positionModal(modalView string, x, y int) string {
	// Use lipgloss.Place to position the modal
	return lipgloss.Place(
		dm.width,
		dm.height,
		lipgloss.Left,
		lipgloss.Top,
		modalView,
		lipgloss.WithWhitespaceChars(" "),
	)
}

// getSortedModals returns modals sorted by z-index
func (dm *DialogManager) getSortedModals() []*Modal {
	sorted := make([]*Modal, len(dm.stack))
	copy(sorted, dm.stack)

	// Simple bubble sort by z-index
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].GetZIndex() > sorted[j+1].GetZIndex() {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// layerStrings layers two strings on top of each other
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

// OpenModal opens a modal with the specified configuration
func (dm *DialogManager) OpenModal(dialog Dialog, config ModalConfig) tea.Cmd {
	// Check if dialog with same ID already exists
	if dm.hasDialog(dialog.ID()) {
		return nil
	}

	// Assign z-index if not set
	if config.ZIndex == 0 {
		config.ZIndex = dm.nextZIndex
		dm.nextZIndex += 100
	}

	// Create modal wrapper
	modal := NewModal(dialog, config)
	modal.SetZIndex(config.ZIndex)

	// Set size
	modal.SetSize(dm.width, dm.height)

	// Calculate position
	modal.CalculatePosition(dm.width, dm.height)

	// Add to stack
	dm.stack = append(dm.stack, modal)
	dm.idMap[dialog.ID()] = len(dm.stack) - 1

	// Activate focus trap if enabled
	if config.TrapFocus {
		dm.focusTrap.Activate()
	}

	return nil
}

// SetSize updates the manager dimensions
func (dm *DialogManager) SetSize(width, height int) {
	dm.width = width
	dm.height = height

	// Update all modals
	for _, modal := range dm.stack {
		modal.SetSize(width, height)
		modal.CalculatePosition(width, height)
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
	return dm.stack[len(dm.stack)-1].GetDialog()
}

// GetTopModal returns the top modal, or nil if stack is empty
func (dm *DialogManager) GetTopModal() *Modal {
	if len(dm.stack) == 0 {
		return nil
	}
	return dm.stack[len(dm.stack)-1]
}

// GetCount returns the number of dialogs in the stack
func (dm *DialogManager) GetCount() int {
	return len(dm.stack)
}

// CloseTop closes the top modal and returns its result
func (dm *DialogManager) CloseTop() tea.Cmd {
	if len(dm.stack) == 0 {
		return nil
	}

	// Get the top modal
	topModal := dm.stack[len(dm.stack)-1]

	// Remove from stack and map
	delete(dm.idMap, topModal.ID())
	dm.stack = dm.stack[:len(dm.stack)-1]

	// Update id map indices
	dm.rebuildIDMap()

	// Deactivate focus trap if no more modals
	if len(dm.stack) == 0 {
		dm.focusTrap.Deactivate()
	}

	// Send result message
	return SendDialogResult(topModal.ID(), topModal.Result(), nil)
}

// CloseByID closes a specific modal by ID
func (dm *DialogManager) CloseByID(id string) tea.Cmd {
	index, exists := dm.idMap[id]
	if !exists {
		return nil
	}

	modal := dm.stack[index]

	// Remove from stack
	dm.stack = append(dm.stack[:index], dm.stack[index+1:]...)
	delete(dm.idMap, id)

	// Update id map indices
	dm.rebuildIDMap()

	// Deactivate focus trap if no more modals
	if len(dm.stack) == 0 {
		dm.focusTrap.Deactivate()
	}

	// Send result message
	return SendDialogResult(modal.ID(), modal.Result(), nil)
}

// hasDialog checks if a dialog with the given ID exists in the stack
func (dm *DialogManager) hasDialog(id string) bool {
	_, exists := dm.idMap[id]
	return exists
}

// rebuildIDMap rebuilds the ID map after stack modifications
func (dm *DialogManager) rebuildIDMap() {
	dm.idMap = make(map[string]int)
	for i, modal := range dm.stack {
		dm.idMap[modal.ID()] = i
	}
}

// Clear removes all modals from the stack
func (dm *DialogManager) Clear() {
	dm.stack = make([]*Modal, 0)
	dm.idMap = make(map[string]int)
	dm.focusTrap.Deactivate()
	dm.nextZIndex = dm.baseZIndex
}

// SetZIndex sets the z-index for a specific modal
func (dm *DialogManager) SetZIndex(id string, zIndex int) tea.Cmd {
	index, exists := dm.idMap[id]
	if !exists {
		return nil
	}

	dm.stack[index].SetZIndex(zIndex)
	return nil
}

// EnableFocusTrap enables focus trapping
func (dm *DialogManager) EnableFocusTrap() tea.Cmd {
	dm.focusTrap.Enable()
	if len(dm.stack) > 0 {
		dm.focusTrap.Activate()
	}
	return nil
}

// DisableFocusTrap disables focus trapping
func (dm *DialogManager) DisableFocusTrap() tea.Cmd {
	dm.focusTrap.Disable()
	dm.focusTrap.Deactivate()
	return nil
}

// RegisterShortcut registers a global keyboard shortcut
func (dm *DialogManager) RegisterShortcut(key string, handler func() tea.Msg) {
	dm.shortcuts.RegisterShortcut(key, handler)
}

// GetShortcutManager returns the shortcut manager
func (dm *DialogManager) GetShortcutManager() *ShortcutManager {
	return dm.shortcuts
}

// GetFocusTrap returns the focus trap
func (dm *DialogManager) GetFocusTrap() *FocusTrap {
	return dm.focusTrap
}

// OpenModalMsg signals to open a modal with custom configuration
type OpenModalMsg struct {
	Dialog Dialog
	Config ModalConfig
}

// OpenModalWithConfig creates a command to open a modal with configuration
func OpenModalWithConfig(dialog Dialog, config ModalConfig) tea.Cmd {
	return func() tea.Msg {
		return OpenModalMsg{
			Dialog: dialog,
			Config: config,
		}
	}
}
