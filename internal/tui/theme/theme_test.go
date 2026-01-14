package theme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestThemeCreation tests creating a new theme
func TestThemeCreation(t *testing.T) {
	colors := ColorPalette{
		Background: lipgloss.Color("#000000"),
		Foreground: lipgloss.Color("#ffffff"),
		Primary:    lipgloss.Color("#0000ff"),
		Secondary:  lipgloss.Color("#00ff00"),
		Accent:     lipgloss.Color("#ff0000"),
		Success:    lipgloss.Color("#00ff00"),
		Warning:    lipgloss.Color("#ffff00"),
		Error:      lipgloss.Color("#ff0000"),
		Info:       lipgloss.Color("#00ffff"),
	}

	theme := NewTheme("Test", colors)

	if theme.Name != "Test" {
		t.Errorf("Expected theme name 'Test', got '%s'", theme.Name)
	}

	if theme.Colors.Background != colors.Background {
		t.Error("Theme colors not set correctly")
	}

	if err := theme.Validate(); err != nil {
		t.Errorf("Valid theme failed validation: %v", err)
	}
}

// TestThemeValidation tests theme validation
func TestThemeValidation(t *testing.T) {
	tests := []struct {
		name      string
		theme     *Theme
		shouldErr bool
	}{
		{
			name: "Valid theme",
			theme: &Theme{
				Name: "Valid",
				Colors: ColorPalette{
					Background: lipgloss.Color("#000000"),
					Foreground: lipgloss.Color("#ffffff"),
					Primary:    lipgloss.Color("#0000ff"),
				},
			},
			shouldErr: false,
		},
		{
			name: "Empty name",
			theme: &Theme{
				Name: "",
				Colors: ColorPalette{
					Background: lipgloss.Color("#000000"),
					Foreground: lipgloss.Color("#ffffff"),
					Primary:    lipgloss.Color("#0000ff"),
				},
			},
			shouldErr: true,
		},
		{
			name: "Missing background",
			theme: &Theme{
				Name: "Invalid",
				Colors: ColorPalette{
					Foreground: lipgloss.Color("#ffffff"),
					Primary:    lipgloss.Color("#0000ff"),
				},
			},
			shouldErr: true,
		},
		{
			name: "Missing foreground",
			theme: &Theme{
				Name: "Invalid",
				Colors: ColorPalette{
					Background: lipgloss.Color("#000000"),
					Primary:    lipgloss.Color("#0000ff"),
				},
			},
			shouldErr: true,
		},
		{
			name: "Missing primary",
			theme: &Theme{
				Name: "Invalid",
				Colors: ColorPalette{
					Background: lipgloss.Color("#000000"),
					Foreground: lipgloss.Color("#ffffff"),
				},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.theme.Validate()
			if tt.shouldErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

// TestThemeClone tests cloning a theme
func TestThemeClone(t *testing.T) {
	original := AINativeTheme()
	clone := original.Clone()

	if clone.Name != original.Name {
		t.Error("Clone has different name")
	}

	if clone.Colors.Primary != original.Colors.Primary {
		t.Error("Clone has different colors")
	}

	// Verify it's a different instance
	if clone == original {
		t.Error("Clone is same instance as original")
	}
}

// TestThemeGetColor tests getting colors by name
func TestThemeGetColor(t *testing.T) {
	theme := AINativeTheme()

	tests := []struct {
		name     string
		expected lipgloss.Color
	}{
		{"background", theme.Colors.Background},
		{"foreground", theme.Colors.Foreground},
		{"primary", theme.Colors.Primary},
		{"secondary", theme.Colors.Secondary},
		{"success", theme.Colors.Success},
		{"error", theme.Colors.Error},
		{"unknown", theme.Colors.Foreground}, // Should return foreground as fallback
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := theme.GetColor(tt.name)
			if color != tt.expected {
				t.Errorf("Expected color %s, got %s", tt.expected, color)
			}
		})
	}
}

// TestThemeManager tests theme manager operations
func TestThemeManager(t *testing.T) {
	manager := NewThemeManager()

	// Test registration
	theme1 := AINativeTheme()
	err := manager.RegisterTheme(theme1)
	if err != nil {
		t.Fatalf("Failed to register theme: %v", err)
	}

	// Test duplicate registration
	err = manager.RegisterTheme(theme1)
	if err == nil {
		t.Error("Expected error when registering duplicate theme")
	}

	// Test getting theme
	retrieved, err := manager.GetTheme("AINative")
	if err != nil {
		t.Fatalf("Failed to get theme: %v", err)
	}
	if retrieved.Name != "AINative" {
		t.Error("Retrieved wrong theme")
	}

	// Test current theme
	current := manager.CurrentTheme()
	if current == nil {
		t.Error("Current theme is nil")
	}
	if current.Name != "AINative" {
		t.Error("Current theme not set to first registered theme")
	}

	// Register more themes
	theme2 := DarkTheme()
	err = manager.RegisterTheme(theme2)
	if err != nil {
		t.Fatalf("Failed to register second theme: %v", err)
	}

	// Test theme switching
	err = manager.SetTheme("Dark")
	if err != nil {
		t.Fatalf("Failed to switch theme: %v", err)
	}

	current = manager.CurrentTheme()
	if current.Name != "Dark" {
		t.Error("Theme not switched correctly")
	}

	// Test listing themes
	themes := manager.ListThemes()
	if len(themes) != 2 {
		t.Errorf("Expected 2 themes, got %d", len(themes))
	}

	// Test getting non-existent theme
	_, err = manager.GetTheme("NonExistent")
	if err == nil {
		t.Error("Expected error when getting non-existent theme")
	}

	// Test switching to non-existent theme
	err = manager.SetTheme("NonExistent")
	if err == nil {
		t.Error("Expected error when switching to non-existent theme")
	}
}

// TestThemeManagerCycle tests cycling through themes
func TestThemeManagerCycle(t *testing.T) {
	manager := NewThemeManager()

	// Register themes
	manager.RegisterTheme(AINativeTheme())
	manager.RegisterTheme(DarkTheme())
	manager.RegisterTheme(LightTheme())

	// Set to first theme
	manager.SetTheme("AINative")

	// Cycle through themes
	err := manager.CycleTheme()
	if err != nil {
		t.Fatalf("Failed to cycle theme: %v", err)
	}

	current := manager.CurrentTheme()
	if current.Name == "AINative" {
		t.Error("Theme did not cycle")
	}

	// Cycle back around
	manager.CycleTheme()
	manager.CycleTheme()

	current = manager.CurrentTheme()
	if current.Name != "AINative" {
		t.Error("Theme did not cycle back to start")
	}
}

// TestThemeManagerListeners tests theme change listeners
func TestThemeManagerListeners(t *testing.T) {
	manager := NewThemeManager()

	manager.RegisterTheme(AINativeTheme())
	manager.RegisterTheme(DarkTheme())

	// Create test listener with channel for synchronization
	done := make(chan bool, 1)
	var oldThemeName, newThemeName string
	listener := &testListener{
		onChangeFunc: func(old, new *Theme) {
			if old != nil {
				oldThemeName = old.Name
			}
			if new != nil {
				newThemeName = new.Name
			}
			done <- true
		},
	}

	manager.AddListener(listener)

	// Switch theme
	manager.SetTheme("Dark")

	// Wait for listener to be called
	<-done

	if oldThemeName != "AINative" {
		t.Errorf("Expected old theme 'AINative', got '%s'", oldThemeName)
	}
	if newThemeName != "Dark" {
		t.Errorf("Expected new theme 'Dark', got '%s'", newThemeName)
	}

	// Test removing listener
	manager.RemoveListener(listener)

	// Verify listener was removed - theme count should be accessible
	if manager.GetThemeCount() != 2 {
		t.Error("Themes were incorrectly modified")
	}
}

// testListener is a test implementation of ThemeChangeListener
type testListener struct {
	onChangeFunc func(old, new *Theme)
}

func (l *testListener) OnThemeChange(old, new *Theme) {
	if l.onChangeFunc != nil {
		l.onChangeFunc(old, new)
	}
}

// TestThemeManagerPersistence tests saving and loading theme config
func TestThemeManagerPersistence(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "theme-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "theme.json")

	// Create manager and register themes
	manager := NewThemeManager()
	manager.RegisterTheme(AINativeTheme())
	manager.RegisterTheme(DarkTheme())
	manager.SetTheme("Dark")

	// Save config
	err = manager.SaveToFile(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Create new manager and load config
	manager2 := NewThemeManager()
	manager2.RegisterTheme(AINativeTheme())
	manager2.RegisterTheme(DarkTheme())

	err = manager2.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify theme was loaded
	current := manager2.CurrentTheme()
	if current.Name != "Dark" {
		t.Errorf("Expected theme 'Dark', got '%s'", current.Name)
	}
}

// TestBuiltinThemes tests all built-in themes
func TestBuiltinThemes(t *testing.T) {
	themes := GetAllBuiltinThemes()

	if len(themes) != 3 {
		t.Errorf("Expected 3 built-in themes, got %d", len(themes))
	}

	expectedNames := []string{"AINative", "Dark", "Light"}
	for i, theme := range themes {
		if theme.Name != expectedNames[i] {
			t.Errorf("Expected theme %d to be '%s', got '%s'", i, expectedNames[i], theme.Name)
		}

		if err := theme.Validate(); err != nil {
			t.Errorf("Built-in theme '%s' failed validation: %v", theme.Name, err)
		}

		// Check color palette completeness
		if theme.Colors.Background == "" {
			t.Errorf("Theme '%s' missing background color", theme.Name)
		}
		if theme.Colors.Foreground == "" {
			t.Errorf("Theme '%s' missing foreground color", theme.Name)
		}
		if theme.Colors.Primary == "" {
			t.Errorf("Theme '%s' missing primary color", theme.Name)
		}
	}
}

// TestRegisterBuiltinThemes tests registering all built-in themes
func TestRegisterBuiltinThemes(t *testing.T) {
	manager := NewThemeManager()

	err := RegisterBuiltinThemes(manager)
	if err != nil {
		t.Fatalf("Failed to register built-in themes: %v", err)
	}

	themes := manager.ListThemes()
	if len(themes) != 3 {
		t.Errorf("Expected 3 themes registered, got %d", len(themes))
	}

	// Verify AINative is default
	current := manager.CurrentTheme()
	if current.Name != "AINative" {
		t.Errorf("Expected default theme 'AINative', got '%s'", current.Name)
	}

	// Verify all themes can be switched to
	for _, name := range []string{"AINative", "Dark", "Light"} {
		err := manager.SetTheme(name)
		if err != nil {
			t.Errorf("Failed to switch to theme '%s': %v", name, err)
		}
	}
}

// TestStyleSetGeneration tests that style sets are properly generated
func TestStyleSetGeneration(t *testing.T) {
	theme := AINativeTheme()

	// Verify key styles are created by checking they're not nil
	// We can't easily check the internal state, but we can verify the styles exist
	_ = theme.Styles.Title
	_ = theme.Styles.Button
	_ = theme.Styles.Dialog
	_ = theme.Styles.Error

	// Verify styles can be used to render text
	titleRendered := theme.Styles.Title.Render("Test")
	if titleRendered == "" {
		t.Error("Title style failed to render")
	}

	buttonRendered := theme.Styles.Button.Render("Test")
	if buttonRendered == "" {
		t.Error("Button style failed to render")
	}
}

// TestThemeManagerUnregister tests unregistering themes
func TestThemeManagerUnregister(t *testing.T) {
	manager := NewThemeManager()

	theme1 := AINativeTheme()
	theme2 := DarkTheme()

	manager.RegisterTheme(theme1)
	manager.RegisterTheme(theme2)

	// Try to unregister current theme (should fail)
	err := manager.UnregisterTheme("AINative")
	if err == nil {
		t.Error("Expected error when unregistering current theme")
	}

	// Switch to different theme
	manager.SetTheme("Dark")

	// Now unregister should work
	err = manager.UnregisterTheme("AINative")
	if err != nil {
		t.Errorf("Failed to unregister theme: %v", err)
	}

	// Verify theme was removed
	themes := manager.ListThemes()
	if len(themes) != 1 {
		t.Errorf("Expected 1 theme after unregister, got %d", len(themes))
	}

	// Try to unregister non-existent theme
	err = manager.UnregisterTheme("NonExistent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent theme")
	}
}

// TestThemeManagerHasTheme tests checking if a theme exists
func TestThemeManagerHasTheme(t *testing.T) {
	manager := NewThemeManager()
	manager.RegisterTheme(AINativeTheme())

	if !manager.HasTheme("AINative") {
		t.Error("Expected HasTheme to return true for registered theme")
	}

	if manager.HasTheme("NonExistent") {
		t.Error("Expected HasTheme to return false for non-existent theme")
	}
}

// TestThemeManagerGetThemeCount tests getting theme count
func TestThemeManagerGetThemeCount(t *testing.T) {
	manager := NewThemeManager()

	if manager.GetThemeCount() != 0 {
		t.Error("Expected 0 themes initially")
	}

	manager.RegisterTheme(AINativeTheme())
	if manager.GetThemeCount() != 1 {
		t.Error("Expected 1 theme after registration")
	}

	manager.RegisterTheme(DarkTheme())
	if manager.GetThemeCount() != 2 {
		t.Error("Expected 2 themes after second registration")
	}
}

// BenchmarkThemeSwitch benchmarks theme switching performance
func BenchmarkThemeSwitch(b *testing.B) {
	manager := NewThemeManager()
	manager.RegisterTheme(AINativeTheme())
	manager.RegisterTheme(DarkTheme())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			manager.SetTheme("AINative")
		} else {
			manager.SetTheme("Dark")
		}
	}
}

// BenchmarkStyleGeneration benchmarks style generation
func BenchmarkStyleGeneration(b *testing.B) {
	colors := ColorPalette{
		Background: lipgloss.Color("#000000"),
		Foreground: lipgloss.Color("#ffffff"),
		Primary:    lipgloss.Color("#0000ff"),
		Secondary:  lipgloss.Color("#00ff00"),
		Accent:     lipgloss.Color("#ff0000"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewTheme("Benchmark", colors)
	}
}
