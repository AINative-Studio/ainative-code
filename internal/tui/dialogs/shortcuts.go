package dialogs

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ShortcutHandler is a function that handles a keyboard shortcut
type ShortcutHandler func() tea.Msg

// ShortcutManager handles global modal keyboard shortcuts
type ShortcutManager struct {
	shortcuts map[string]ShortcutHandler // Map of key -> handler
	enabled   bool                        // Whether shortcuts are enabled
}

// NewShortcutManager creates a new shortcut manager
func NewShortcutManager() *ShortcutManager {
	return &ShortcutManager{
		shortcuts: make(map[string]ShortcutHandler),
		enabled:   true,
	}
}

// RegisterShortcut adds a global shortcut
func (s *ShortcutManager) RegisterShortcut(key string, handler ShortcutHandler) {
	s.shortcuts[key] = handler
}

// UnregisterShortcut removes a global shortcut
func (s *ShortcutManager) UnregisterShortcut(key string) {
	delete(s.shortcuts, key)
}

// ClearShortcuts removes all shortcuts
func (s *ShortcutManager) ClearShortcuts() {
	s.shortcuts = make(map[string]ShortcutHandler)
}

// HasShortcut checks if a shortcut is registered
func (s *ShortcutManager) HasShortcut(key string) bool {
	_, exists := s.shortcuts[key]
	return exists
}

// GetShortcut returns the handler for a shortcut, or nil if not found
func (s *ShortcutManager) GetShortcut(key string) ShortcutHandler {
	return s.shortcuts[key]
}

// GetAllShortcuts returns all registered shortcuts
func (s *ShortcutManager) GetAllShortcuts() map[string]ShortcutHandler {
	// Return a copy to prevent external modification
	shortcuts := make(map[string]ShortcutHandler, len(s.shortcuts))
	for k, v := range s.shortcuts {
		shortcuts[k] = v
	}
	return shortcuts
}

// GetShortcutKeys returns all registered shortcut keys
func (s *ShortcutManager) GetShortcutKeys() []string {
	keys := make([]string, 0, len(s.shortcuts))
	for k := range s.shortcuts {
		keys = append(keys, k)
	}
	return keys
}

// HandleKey processes a key event and triggers the appropriate shortcut
// Returns (handled, command)
func (s *ShortcutManager) HandleKey(key string) (bool, tea.Cmd) {
	if !s.enabled {
		return false, nil
	}

	handler, exists := s.shortcuts[key]
	if !exists {
		return false, nil
	}

	// Execute the handler and return the message as a command
	msg := handler()
	if msg == nil {
		return true, nil
	}

	return true, func() tea.Msg {
		return msg
	}
}

// HandleMessage processes a Bubble Tea message for shortcuts
func (s *ShortcutManager) HandleMessage(msg tea.Msg) (bool, tea.Cmd) {
	if !s.enabled {
		return false, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return s.HandleKey(msg.String())
	}

	return false, nil
}

// Enable enables shortcut handling
func (s *ShortcutManager) Enable() {
	s.enabled = true
}

// Disable disables shortcut handling
func (s *ShortcutManager) Disable() {
	s.enabled = false
}

// IsEnabled returns true if shortcut handling is enabled
func (s *ShortcutManager) IsEnabled() bool {
	return s.enabled
}

// Toggle toggles shortcut handling on/off
func (s *ShortcutManager) Toggle() {
	s.enabled = !s.enabled
}

// GetShortcutCount returns the number of registered shortcuts
func (s *ShortcutManager) GetShortcutCount() int {
	return len(s.shortcuts)
}

// Common shortcut message types

// CommandPaletteMsg signals to open the command palette
type CommandPaletteMsg struct{}

// FilePickerMsg signals to open the file picker
type FilePickerMsg struct{}

// SearchMsg signals to open the search dialog
type SearchMsg struct{}

// SettingsMsg signals to open the settings dialog
type SettingsMsg struct{}

// HelpMsg signals to open the help dialog
type HelpMsg struct{}

// Helper functions to create common shortcuts

// RegisterCommandPalette registers Ctrl+K for command palette
func (s *ShortcutManager) RegisterCommandPalette(handler ShortcutHandler) {
	if handler == nil {
		handler = func() tea.Msg {
			return CommandPaletteMsg{}
		}
	}
	s.RegisterShortcut("ctrl+k", handler)
}

// RegisterFilePicker registers Ctrl+P for file picker
func (s *ShortcutManager) RegisterFilePicker(handler ShortcutHandler) {
	if handler == nil {
		handler = func() tea.Msg {
			return FilePickerMsg{}
		}
	}
	s.RegisterShortcut("ctrl+p", handler)
}

// RegisterSearch registers Ctrl+F for search
func (s *ShortcutManager) RegisterSearch(handler ShortcutHandler) {
	if handler == nil {
		handler = func() tea.Msg {
			return SearchMsg{}
		}
	}
	s.RegisterShortcut("ctrl+f", handler)
}

// RegisterSettings registers Ctrl+, for settings
func (s *ShortcutManager) RegisterSettings(handler ShortcutHandler) {
	if handler == nil {
		handler = func() tea.Msg {
			return SettingsMsg{}
		}
	}
	s.RegisterShortcut("ctrl+,", handler)
}

// RegisterHelp registers F1 or Ctrl+? for help
func (s *ShortcutManager) RegisterHelp(handler ShortcutHandler) {
	if handler == nil {
		handler = func() tea.Msg {
			return HelpMsg{}
		}
	}
	s.RegisterShortcut("f1", handler)
	s.RegisterShortcut("ctrl+?", handler)
}

// RegisterCommonShortcuts registers all common shortcuts with default handlers
func (s *ShortcutManager) RegisterCommonShortcuts() {
	s.RegisterCommandPalette(nil)
	s.RegisterFilePicker(nil)
	s.RegisterSearch(nil)
	s.RegisterSettings(nil)
	s.RegisterHelp(nil)
}

// ShortcutInfo represents information about a shortcut
type ShortcutInfo struct {
	Key         string // The key combination
	Description string // What the shortcut does
}

// GetShortcutHelp returns help information for all registered shortcuts
// This is useful for displaying help dialogs
func GetCommonShortcutHelp() []ShortcutInfo {
	return []ShortcutInfo{
		{Key: "Ctrl+K", Description: "Open command palette"},
		{Key: "Ctrl+P", Description: "Open file picker"},
		{Key: "Ctrl+F", Description: "Search"},
		{Key: "Ctrl+,", Description: "Open settings"},
		{Key: "F1 / Ctrl+?", Description: "Show help"},
		{Key: "ESC", Description: "Close modal"},
		{Key: "Tab", Description: "Next element"},
		{Key: "Shift+Tab", Description: "Previous element"},
	}
}
