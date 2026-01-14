package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ThemeManager handles theme registration, switching, and persistence
type ThemeManager struct {
	current   *Theme
	themes    map[string]*Theme
	listeners []ThemeChangeListener
	mu        sync.RWMutex
	configDir string
}

// ThemeChangeListener is notified when the theme changes
type ThemeChangeListener interface {
	OnThemeChange(oldTheme, newTheme *Theme)
}

// ThemeConfig represents the persisted theme configuration
type ThemeConfig struct {
	CurrentTheme string `json:"current_theme"`
	Version      string `json:"version"`
}

// NewThemeManager creates a new theme manager
func NewThemeManager() *ThemeManager {
	configDir := getConfigDir()

	return &ThemeManager{
		current:   nil,
		themes:    make(map[string]*Theme),
		listeners: make([]ThemeChangeListener, 0),
		configDir: configDir,
	}
}

// RegisterTheme registers a new theme
func (tm *ThemeManager) RegisterTheme(theme *Theme) error {
	if theme == nil {
		return fmt.Errorf("cannot register nil theme")
	}

	if err := theme.Validate(); err != nil {
		return fmt.Errorf("invalid theme: %w", err)
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if theme with same name already exists
	if _, exists := tm.themes[theme.Name]; exists {
		return fmt.Errorf("theme '%s' already registered", theme.Name)
	}

	tm.themes[theme.Name] = theme

	// If this is the first theme, set it as current
	if tm.current == nil {
		tm.current = theme
	}

	return nil
}

// UnregisterTheme removes a theme from the manager
func (tm *ThemeManager) UnregisterTheme(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.themes[name]; !exists {
		return fmt.Errorf("theme '%s' not found", name)
	}

	// Don't allow unregistering the current theme
	if tm.current != nil && tm.current.Name == name {
		return fmt.Errorf("cannot unregister current theme '%s'", name)
	}

	delete(tm.themes, name)
	return nil
}

// SetTheme switches to a different theme by name
func (tm *ThemeManager) SetTheme(name string) error {
	tm.mu.Lock()

	newTheme, exists := tm.themes[name]
	if !exists {
		tm.mu.Unlock()
		return fmt.Errorf("theme '%s' not found", name)
	}

	oldTheme := tm.current
	tm.current = newTheme

	// Make a copy of listeners to avoid holding lock during notifications
	listeners := make([]ThemeChangeListener, len(tm.listeners))
	copy(listeners, tm.listeners)

	tm.mu.Unlock()

	// Notify listeners (outside of lock to prevent deadlocks)
	tm.notifyListeners(oldTheme, newTheme, listeners)

	return nil
}

// GetTheme returns a theme by name
func (tm *ThemeManager) GetTheme(name string) (*Theme, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	theme, exists := tm.themes[name]
	if !exists {
		return nil, fmt.Errorf("theme '%s' not found", name)
	}

	return theme, nil
}

// CurrentTheme returns the currently active theme
func (tm *ThemeManager) CurrentTheme() *Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.current
}

// ListThemes returns a list of all registered theme names
func (tm *ThemeManager) ListThemes() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	names := make([]string, 0, len(tm.themes))
	for name := range tm.themes {
		names = append(names, name)
	}

	return names
}

// CycleTheme switches to the next theme in the list
func (tm *ThemeManager) CycleTheme() error {
	tm.mu.RLock()

	if len(tm.themes) == 0 {
		tm.mu.RUnlock()
		return fmt.Errorf("no themes registered")
	}

	if len(tm.themes) == 1 {
		tm.mu.RUnlock()
		return nil // Only one theme, nothing to cycle
	}

	// Get ordered list of theme names
	names := make([]string, 0, len(tm.themes))
	for name := range tm.themes {
		names = append(names, name)
	}

	currentName := ""
	if tm.current != nil {
		currentName = tm.current.Name
	}

	tm.mu.RUnlock()

	// Find current theme index
	currentIdx := -1
	for i, name := range names {
		if name == currentName {
			currentIdx = i
			break
		}
	}

	// Calculate next index (wraps around)
	nextIdx := (currentIdx + 1) % len(names)
	nextTheme := names[nextIdx]

	return tm.SetTheme(nextTheme)
}

// AddListener registers a theme change listener
func (tm *ThemeManager) AddListener(listener ThemeChangeListener) {
	if listener == nil {
		return
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.listeners = append(tm.listeners, listener)
}

// RemoveListener unregisters a theme change listener
func (tm *ThemeManager) RemoveListener(listener ThemeChangeListener) {
	if listener == nil {
		return
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i, l := range tm.listeners {
		if l == listener {
			// Remove listener by replacing with last element and truncating
			tm.listeners[i] = tm.listeners[len(tm.listeners)-1]
			tm.listeners = tm.listeners[:len(tm.listeners)-1]
			break
		}
	}
}

// ClearListeners removes all theme change listeners
func (tm *ThemeManager) ClearListeners() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.listeners = make([]ThemeChangeListener, 0)
}

// notifyListeners notifies all listeners of a theme change
func (tm *ThemeManager) notifyListeners(oldTheme, newTheme *Theme, listeners []ThemeChangeListener) {
	for _, listener := range listeners {
		// Call listener in goroutine to prevent blocking
		go listener.OnThemeChange(oldTheme, newTheme)
	}
}

// LoadFromFile loads theme configuration from a file
func (tm *ThemeManager) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, not an error - use defaults
			return nil
		}
		return fmt.Errorf("failed to read theme config: %w", err)
	}

	var config ThemeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse theme config: %w", err)
	}

	// Try to set the theme from config
	if config.CurrentTheme != "" {
		if err := tm.SetTheme(config.CurrentTheme); err != nil {
			// Theme not found, not a fatal error - will use default
			return nil
		}
	}

	return nil
}

// SaveToFile saves the current theme configuration to a file
func (tm *ThemeManager) SaveToFile(path string) error {
	tm.mu.RLock()
	currentName := ""
	if tm.current != nil {
		currentName = tm.current.Name
	}
	tm.mu.RUnlock()

	config := ThemeConfig{
		CurrentTheme: currentName,
		Version:      "1.0",
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme config: %w", err)
	}

	return nil
}

// LoadConfig loads theme configuration from default location
func (tm *ThemeManager) LoadConfig() error {
	configPath := tm.getConfigPath()
	return tm.LoadFromFile(configPath)
}

// SaveConfig saves theme configuration to default location
func (tm *ThemeManager) SaveConfig() error {
	configPath := tm.getConfigPath()
	return tm.SaveToFile(configPath)
}

// getConfigPath returns the default theme configuration file path
func (tm *ThemeManager) getConfigPath() string {
	return filepath.Join(tm.configDir, "theme.json")
}

// getConfigDir returns the application config directory
func getConfigDir() string {
	// Try XDG_CONFIG_HOME first
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome != "" {
		return filepath.Join(configHome, "ainative-code")
	}

	// Fall back to ~/.config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Last resort: use current directory
		return ".ainative-code"
	}

	return filepath.Join(homeDir, ".config", "ainative-code")
}

// GetThemeCount returns the number of registered themes
func (tm *ThemeManager) GetThemeCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return len(tm.themes)
}

// HasTheme checks if a theme is registered
func (tm *ThemeManager) HasTheme(name string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	_, exists := tm.themes[name]
	return exists
}

// ResetToDefault resets to the first registered theme (usually the default)
func (tm *ThemeManager) ResetToDefault() error {
	tm.mu.RLock()
	names := tm.ListThemes()
	tm.mu.RUnlock()

	if len(names) == 0 {
		return fmt.Errorf("no themes registered")
	}

	// Use first theme as default
	return tm.SetTheme(names[0])
}
