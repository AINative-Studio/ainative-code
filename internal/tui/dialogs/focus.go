package dialogs

import (
	tea "github.com/charmbracelet/bubbletea"
)

// FocusTrap manages focus within modal boundaries
// It prevents Tab/Shift+Tab from leaving the modal
type FocusTrap struct {
	focusableIDs []string // List of focusable element IDs
	currentIndex int      // Current focus index
	trapped      bool     // Whether focus is currently trapped
	enabled      bool     // Whether trap is enabled
}

// NewFocusTrap creates a new focus trap
func NewFocusTrap() *FocusTrap {
	return &FocusTrap{
		focusableIDs: make([]string, 0),
		currentIndex: 0,
		trapped:      false,
		enabled:      true,
	}
}

// Enable enables focus trapping
func (f *FocusTrap) Enable() {
	f.enabled = true
}

// Disable disables focus trapping
func (f *FocusTrap) Disable() {
	f.enabled = false
}

// IsEnabled returns true if focus trapping is enabled
func (f *FocusTrap) IsEnabled() bool {
	return f.enabled
}

// Activate activates the focus trap
func (f *FocusTrap) Activate() {
	if f.enabled {
		f.trapped = true
	}
}

// Deactivate deactivates the focus trap
func (f *FocusTrap) Deactivate() {
	f.trapped = false
}

// IsActive returns true if focus is currently trapped
func (f *FocusTrap) IsActive() bool {
	return f.trapped && f.enabled
}

// SetFocusableElements sets the list of focusable element IDs
func (f *FocusTrap) SetFocusableElements(ids []string) {
	f.focusableIDs = ids
	if len(ids) > 0 && f.currentIndex >= len(ids) {
		f.currentIndex = 0
	}
}

// AddFocusableElement adds a focusable element ID
func (f *FocusTrap) AddFocusableElement(id string) {
	f.focusableIDs = append(f.focusableIDs, id)
}

// RemoveFocusableElement removes a focusable element ID
func (f *FocusTrap) RemoveFocusableElement(id string) {
	for i, fid := range f.focusableIDs {
		if fid == id {
			f.focusableIDs = append(f.focusableIDs[:i], f.focusableIDs[i+1:]...)
			if f.currentIndex >= len(f.focusableIDs) && len(f.focusableIDs) > 0 {
				f.currentIndex = len(f.focusableIDs) - 1
			}
			break
		}
	}
}

// ClearFocusableElements clears all focusable elements
func (f *FocusTrap) ClearFocusableElements() {
	f.focusableIDs = make([]string, 0)
	f.currentIndex = 0
}

// GetFocusableElements returns the list of focusable element IDs
func (f *FocusTrap) GetFocusableElements() []string {
	return f.focusableIDs
}

// HandleKey handles keyboard navigation within the focus trap
// Returns (handled, nextFocusID)
func (f *FocusTrap) HandleKey(key string) (handled bool, nextFocus string) {
	if !f.IsActive() {
		return false, ""
	}

	if len(f.focusableIDs) == 0 {
		return false, ""
	}

	switch key {
	case "tab":
		// Move to next focusable element
		f.currentIndex = (f.currentIndex + 1) % len(f.focusableIDs)
		return true, f.focusableIDs[f.currentIndex]

	case "shift+tab":
		// Move to previous focusable element
		f.currentIndex--
		if f.currentIndex < 0 {
			f.currentIndex = len(f.focusableIDs) - 1
		}
		return true, f.focusableIDs[f.currentIndex]
	}

	return false, ""
}

// HandleMessage handles Bubble Tea messages for focus navigation
func (f *FocusTrap) HandleMessage(msg tea.Msg) (handled bool, nextFocus string, cmd tea.Cmd) {
	if !f.IsActive() {
		return false, "", nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		handled, nextFocus := f.HandleKey(msg.String())
		return handled, nextFocus, nil
	}

	return false, "", nil
}

// GetCurrentFocusID returns the ID of the currently focused element
func (f *FocusTrap) GetCurrentFocusID() string {
	if len(f.focusableIDs) == 0 {
		return ""
	}
	if f.currentIndex >= 0 && f.currentIndex < len(f.focusableIDs) {
		return f.focusableIDs[f.currentIndex]
	}
	return ""
}

// SetCurrentFocusIndex sets the current focus index
func (f *FocusTrap) SetCurrentFocusIndex(index int) {
	if index >= 0 && index < len(f.focusableIDs) {
		f.currentIndex = index
	}
}

// GetCurrentFocusIndex returns the current focus index
func (f *FocusTrap) GetCurrentFocusIndex() int {
	return f.currentIndex
}

// FocusFirst sets focus to the first focusable element
func (f *FocusTrap) FocusFirst() string {
	if len(f.focusableIDs) > 0 {
		f.currentIndex = 0
		return f.focusableIDs[0]
	}
	return ""
}

// FocusLast sets focus to the last focusable element
func (f *FocusTrap) FocusLast() string {
	if len(f.focusableIDs) > 0 {
		f.currentIndex = len(f.focusableIDs) - 1
		return f.focusableIDs[f.currentIndex]
	}
	return ""
}

// NextFocusable returns the ID of the next focusable element
func (f *FocusTrap) NextFocusable() string {
	if len(f.focusableIDs) == 0 {
		return ""
	}
	nextIndex := (f.currentIndex + 1) % len(f.focusableIDs)
	return f.focusableIDs[nextIndex]
}

// PrevFocusable returns the ID of the previous focusable element
func (f *FocusTrap) PrevFocusable() string {
	if len(f.focusableIDs) == 0 {
		return ""
	}
	prevIndex := f.currentIndex - 1
	if prevIndex < 0 {
		prevIndex = len(f.focusableIDs) - 1
	}
	return f.focusableIDs[prevIndex]
}

// HasFocusableElements returns true if there are any focusable elements
func (f *FocusTrap) HasFocusableElements() bool {
	return len(f.focusableIDs) > 0
}

// GetFocusableCount returns the number of focusable elements
func (f *FocusTrap) GetFocusableCount() int {
	return len(f.focusableIDs)
}

// Reset resets the focus trap to its initial state
func (f *FocusTrap) Reset() {
	f.currentIndex = 0
	f.trapped = false
}
