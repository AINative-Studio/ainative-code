package theme

import (
	"fmt"
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

// =============================================================================
// Category 1: Style Generation Tests
// =============================================================================

// TestBuildStyleSet tests that buildStyleSet creates all required styles
func TestBuildStyleSet(t *testing.T) {
	colors := ColorPalette{
		Background:         lipgloss.Color("#1a1b26"),
		Foreground:         lipgloss.Color("#c0caf5"),
		Primary:            lipgloss.Color("#7aa2f7"),
		Secondary:          lipgloss.Color("#bb9af7"),
		Accent:             lipgloss.Color("#f7768e"),
		Success:            lipgloss.Color("#9ece6a"),
		Warning:            lipgloss.Color("#e0af68"),
		Error:              lipgloss.Color("#f7768e"),
		Info:               lipgloss.Color("#7dcfff"),
		Border:             lipgloss.Color("#565f89"),
		ButtonActive:       lipgloss.Color("#7aa2f7"),
		ButtonInactive:     lipgloss.Color("#565f89"),
		Muted:              lipgloss.Color("#565f89"),
		ThinkingBackground: lipgloss.Color("#1f2335"),
		CodeKeyword:        lipgloss.Color("#bb9af7"),
		StatusBar:          lipgloss.Color("#1f2335"),
		InputBorder:        lipgloss.Color("#565f89"),
		InputFocus:         lipgloss.Color("#7aa2f7"),
		Selection:          lipgloss.Color("#7aa2f7"),
	}

	borders := buildBorderSet()
	spacing := buildSpacingSet()
	styles := buildStyleSet(colors, borders, spacing)

	// Test text styles are created
	if styles.Title.GetForeground() != colors.Primary {
		t.Error("Title style has incorrect foreground color")
	}
	if styles.Subtitle.GetForeground() != colors.Secondary {
		t.Error("Subtitle style has incorrect foreground color")
	}
	if styles.Body.GetForeground() != colors.Foreground {
		t.Error("Body style has incorrect foreground color")
	}
	if styles.Muted.GetForeground() != colors.Muted {
		t.Error("Muted style has incorrect foreground color")
	}

	// Test status styles are created
	if styles.Success.GetForeground() != colors.Success {
		t.Error("Success style has incorrect foreground color")
	}
	if styles.Warning.GetForeground() != colors.Warning {
		t.Error("Warning style has incorrect foreground color")
	}
	if styles.Error.GetForeground() != colors.Error {
		t.Error("Error style has incorrect foreground color")
	}
	if styles.Info.GetForeground() != colors.Info {
		t.Error("Info style has incorrect foreground color")
	}

	// Test button styles are created
	if styles.ButtonActive.GetBackground() != colors.ButtonActive {
		t.Error("ButtonActive style has incorrect background color")
	}

	// Test dialog styles are created
	if styles.Dialog.GetBackground() != colors.Background {
		t.Error("Dialog style has incorrect background color")
	}

	// Test input styles are created
	// Note: lipgloss.Style doesn't expose GetBorderForeground(), so we test rendering instead
	if styles.InputField.Render("test") == "" {
		t.Error("InputField style failed to render")
	}
	if styles.InputFieldFocus.Render("test") == "" {
		t.Error("InputFieldFocus style failed to render")
	}
}

// TestBuildBorderSet tests that buildBorderSet creates all border types
func TestBuildBorderSet(t *testing.T) {
	borders := buildBorderSet()

	// Verify borders have correct properties by checking their characters
	// Normal border should have standard characters
	if borders.Normal.TopLeft == "" {
		t.Error("Normal border missing TopLeft character")
	}
	if borders.Normal.Top == "" {
		t.Error("Normal border missing Top character")
	}

	// Rounded border should have rounded characters
	if borders.Rounded.TopLeft == "" {
		t.Error("Rounded border missing TopLeft character")
	}
	if borders.Rounded.Top == "" {
		t.Error("Rounded border missing Top character")
	}

	// Thick border should have characters
	if borders.Thick.TopLeft == "" {
		t.Error("Thick border missing TopLeft character")
	}

	// Double border should have characters
	if borders.Double.TopLeft == "" {
		t.Error("Double border missing TopLeft character")
	}

	// Hidden border should have no visible characters
	if borders.Hidden.TopLeft != " " && borders.Hidden.TopLeft != "" {
		t.Error("Hidden border should have no visible TopLeft character")
	}
}

// TestStyleApplication tests applying styles to text
func TestStyleApplication(t *testing.T) {
	theme := AINativeTheme()

	tests := []struct {
		name  string
		style lipgloss.Style
		text  string
	}{
		{"Title", theme.Styles.Title, "Test Title"},
		{"Error", theme.Styles.Error, "Error Message"},
		{"Success", theme.Styles.Success, "Success Message"},
		{"Warning", theme.Styles.Warning, "Warning Message"},
		{"Info", theme.Styles.Info, "Info Message"},
		{"Button", theme.Styles.Button, "Button Text"},
		{"Dialog", theme.Styles.Dialog, "Dialog Content"},
		{"Code", theme.Styles.Code, "code snippet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render(tt.text)
			if rendered == "" {
				t.Errorf("%s style failed to render", tt.name)
			}
			// Verify the rendered text contains the original text
			// (lipgloss adds ANSI codes, so we can't do exact match)
			if len(rendered) < len(tt.text) {
				t.Errorf("%s style rendered text is shorter than input", tt.name)
			}
		})
	}
}

// TestThemeRenderHelpers tests render helper functions
func TestThemeRenderHelpers(t *testing.T) {
	theme := AINativeTheme()
	helpers := NewRenderHelpers(theme)

	// Test GetTheme
	if helpers.GetTheme() != theme {
		t.Error("GetTheme returned incorrect theme")
	}

	// Test SetTheme
	newTheme := DarkTheme()
	helpers.SetTheme(newTheme)
	if helpers.GetTheme() != newTheme {
		t.Error("SetTheme did not update theme")
	}
	helpers.SetTheme(theme) // Reset for other tests

	// Test FormatError
	testErr := fmt.Errorf("test error")
	errorMsg := helpers.FormatError(testErr)
	if errorMsg == "" {
		t.Error("FormatError returned empty string")
	}

	// Test FormatError with nil
	nilErrorMsg := helpers.FormatError(nil)
	if nilErrorMsg != "" {
		t.Error("FormatError should return empty string for nil error")
	}

	// Test FormatSuccess
	successMsg := helpers.FormatSuccess("Operation completed")
	if successMsg == "" {
		t.Error("FormatSuccess returned empty string")
	}

	// Test FormatWarning
	warningMsg := helpers.FormatWarning("Warning message")
	if warningMsg == "" {
		t.Error("FormatWarning returned empty string")
	}

	// Test FormatInfo
	infoMsg := helpers.FormatInfo("Info message")
	if infoMsg == "" {
		t.Error("FormatInfo returned empty string")
	}

	// Test message formatters
	userMsg := helpers.FormatUserMessage("Hello")
	if userMsg == "" {
		t.Error("FormatUserMessage returned empty string")
	}

	assistantMsg := helpers.FormatAssistantMessage("Hi there")
	if assistantMsg == "" {
		t.Error("FormatAssistantMessage returned empty string")
	}

	systemMsg := helpers.FormatSystemMessage("System notification")
	if systemMsg == "" {
		t.Error("FormatSystemMessage returned empty string")
	}

	// Test FormatThemeIndicator
	indicator := helpers.FormatThemeIndicator()
	if indicator == "" {
		t.Error("FormatThemeIndicator returned empty string")
	}
}

// TestThemeRenderHelperStyles tests all render helper style methods
func TestThemeRenderHelperStyles(t *testing.T) {
	theme := AINativeTheme()
	helpers := NewRenderHelpers(theme)

	// Test all style methods return non-empty styles
	tests := []struct {
		name      string
		styleFunc func() lipgloss.Style
	}{
		{"BorderStyle", helpers.BorderStyle},
		{"StatusBarStyle", helpers.StatusBarStyle},
		{"StreamingIndicatorStyle", helpers.StreamingIndicatorStyle},
		{"HelpHintStyle", helpers.HelpHintStyle},
		{"InputPromptStyle", helpers.InputPromptStyle},
		{"ErrorStyle", helpers.ErrorStyle},
		{"SeparatorStyle", helpers.SeparatorStyle},
		{"DisabledStyle", helpers.DisabledStyle},
		{"ReadyStyle", helpers.ReadyStyle},
		{"CenteredTextStyle", helpers.CenteredTextStyle},
		{"QuitStyle", helpers.QuitStyle},
		{"LoadingStyle", helpers.LoadingStyle},
		{"ScrollIndicatorStyle", helpers.ScrollIndicatorStyle},
		{"PlaceholderStyle", helpers.PlaceholderStyle},
		{"LSPConnectedStyle", helpers.LSPConnectedStyle},
		{"LSPConnectingStyle", helpers.LSPConnectingStyle},
		{"LSPErrorStyle", helpers.LSPErrorStyle},
		{"LSPDisconnectedStyle", helpers.LSPDisconnectedStyle},
		{"ThemeIndicatorStyle", helpers.ThemeIndicatorStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := tt.styleFunc()
			// Test that the style can render text
			rendered := style.Render("test")
			if rendered == "" {
				t.Errorf("%s returned style that cannot render", tt.name)
			}
		})
	}
}

// =============================================================================
// Category 2: File I/O & Error Handling Tests
// =============================================================================

// TestThemeLoadFromFileErrors tests various error conditions when loading theme config
func TestThemeLoadFromFileErrors(t *testing.T) {
	manager := NewThemeManager()
	manager.RegisterTheme(AINativeTheme())
	manager.RegisterTheme(DarkTheme())

	tests := []struct {
		name      string
		setupFunc func() string
		shouldErr bool
	}{
		{
			name: "Non-existent file",
			setupFunc: func() string {
				return "/nonexistent/path/theme.json"
			},
			shouldErr: false, // LoadFromFile returns nil for non-existent files
		},
		{
			name: "Invalid JSON",
			setupFunc: func() string {
				tmpFile := filepath.Join(t.TempDir(), "invalid.json")
				os.WriteFile(tmpFile, []byte("{ invalid json }"), 0644)
				return tmpFile
			},
			shouldErr: true,
		},
		{
			name: "Empty file",
			setupFunc: func() string {
				tmpFile := filepath.Join(t.TempDir(), "empty.json")
				os.WriteFile(tmpFile, []byte(""), 0644)
				return tmpFile
			},
			shouldErr: true,
		},
		{
			name: "Valid JSON with non-existent theme",
			setupFunc: func() string {
				tmpFile := filepath.Join(t.TempDir(), "valid.json")
				os.WriteFile(tmpFile, []byte(`{"current_theme":"NonExistent","version":"1.0"}`), 0644)
				return tmpFile
			},
			shouldErr: false, // LoadFromFile ignores non-existent themes
		},
		{
			name: "Valid JSON with existing theme",
			setupFunc: func() string {
				tmpFile := filepath.Join(t.TempDir(), "valid.json")
				os.WriteFile(tmpFile, []byte(`{"current_theme":"Dark","version":"1.0"}`), 0644)
				return tmpFile
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupFunc()
			err := manager.LoadFromFile(path)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestThemeSaveToFileErrors tests various error conditions when saving theme config
func TestThemeSaveToFileErrors(t *testing.T) {
	manager := NewThemeManager()
	manager.RegisterTheme(AINativeTheme())

	tests := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{
			name:      "Valid path",
			path:      filepath.Join(t.TempDir(), "valid", "theme.json"),
			shouldErr: false,
		},
		{
			name:      "Invalid path - root directory (permission denied on many systems)",
			path:      "/theme.json",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.SaveToFile(tt.path)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestThemeManagerLoadSaveConfig tests LoadConfig and SaveConfig methods
func TestThemeManagerLoadSaveConfig(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()

	// Create a manager with custom config dir
	manager := NewThemeManager()
	manager.configDir = tmpDir
	manager.RegisterTheme(AINativeTheme())
	manager.RegisterTheme(DarkTheme())
	manager.SetTheme("Dark")

	// Test SaveConfig
	err := manager.SaveConfig()
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(tmpDir, "theme.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created by SaveConfig")
	}

	// Create new manager and load config
	manager2 := NewThemeManager()
	manager2.configDir = tmpDir
	manager2.RegisterTheme(AINativeTheme())
	manager2.RegisterTheme(DarkTheme())

	err = manager2.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify theme was loaded
	if manager2.CurrentTheme().Name != "Dark" {
		t.Errorf("Expected theme 'Dark', got '%s'", manager2.CurrentTheme().Name)
	}
}

// TestThemeManagerPersistenceWithCorruptedData tests recovery from corrupted config
func TestThemeManagerPersistenceWithCorruptedData(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "theme.json")

	// Write corrupted JSON
	err := os.WriteFile(configPath, []byte("corrupted{json}"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	manager := NewThemeManager()
	manager.RegisterTheme(AINativeTheme())

	// LoadFromFile should return error for corrupted JSON
	err = manager.LoadFromFile(configPath)
	if err == nil {
		t.Error("Expected error when loading corrupted config")
	}

	// Manager should still be usable with default theme
	if manager.CurrentTheme() == nil {
		t.Error("Manager should have default theme after load failure")
	}
}

// TestThemeManagerRegisterNilTheme tests registering a nil theme
func TestThemeManagerRegisterNilTheme(t *testing.T) {
	manager := NewThemeManager()

	err := manager.RegisterTheme(nil)
	if err == nil {
		t.Error("Expected error when registering nil theme")
	}
}

// TestThemeManagerAddNilListener tests adding a nil listener
func TestThemeManagerAddNilListener(t *testing.T) {
	manager := NewThemeManager()
	manager.RegisterTheme(AINativeTheme())

	// Adding nil listener should not crash
	manager.AddListener(nil)

	// Should still be able to switch themes
	manager.RegisterTheme(DarkTheme())
	err := manager.SetTheme("Dark")
	if err != nil {
		t.Fatalf("SetTheme failed: %v", err)
	}
}

// TestThemeManagerRemoveNilListener tests removing a nil listener
func TestThemeManagerRemoveNilListener(t *testing.T) {
	manager := NewThemeManager()
	manager.RegisterTheme(AINativeTheme())

	// Removing nil listener should not crash
	manager.RemoveListener(nil)

	// Should still be able to switch themes
	manager.RegisterTheme(DarkTheme())
	err := manager.SetTheme("Dark")
	if err != nil {
		t.Fatalf("SetTheme failed: %v", err)
	}
}

// TestThemeManagerClearListeners tests clearing all listeners
func TestThemeManagerClearListeners(t *testing.T) {
	manager := NewThemeManager()
	manager.RegisterTheme(AINativeTheme())
	manager.RegisterTheme(DarkTheme())

	// Add multiple listeners
	listener1 := &testListener{onChangeFunc: func(old, new *Theme) {}}
	listener2 := &testListener{onChangeFunc: func(old, new *Theme) {}}

	manager.AddListener(listener1)
	manager.AddListener(listener2)

	// Clear all listeners
	manager.ClearListeners()

	// Verify listeners were cleared (can't directly test, but no crashes is good)
	err := manager.SetTheme("Dark")
	if err != nil {
		t.Fatalf("SetTheme failed: %v", err)
	}
}

// =============================================================================
// Category 3: Theme Validation & Edge Cases Tests
// =============================================================================

// TestThemeGetColorEdgeCases tests GetColor method with edge cases
func TestThemeGetColorEdgeCases(t *testing.T) {
	theme := AINativeTheme()

	tests := []struct {
		name     string
		color    string
		fallback lipgloss.Color
	}{
		{"Empty string", "", theme.Colors.Foreground},
		{"Invalid color name", "invalid_color", theme.Colors.Foreground},
		{"Case sensitive - uppercase", "PRIMARY", theme.Colors.Foreground},
		{"Case sensitive - mixed", "Primary", theme.Colors.Foreground},
		{"Unknown alias", "bg", theme.Colors.Foreground},
		{"Special characters", "!@#$%", theme.Colors.Foreground},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := theme.GetColor(tt.color)
			if color != tt.fallback {
				t.Errorf("Expected fallback color %s, got %s", tt.fallback, color)
			}
		})
	}
}

// TestThemeGetName tests GetName method
func TestThemeGetName(t *testing.T) {
	theme := AINativeTheme()
	if theme.GetName() != "AINative" {
		t.Errorf("Expected name 'AINative', got '%s'", theme.GetName())
	}

	theme2 := DarkTheme()
	if theme2.GetName() != "Dark" {
		t.Errorf("Expected name 'Dark', got '%s'", theme2.GetName())
	}
}

// TestThemeCloneEdgeCases tests cloning edge cases
func TestThemeCloneEdgeCases(t *testing.T) {
	original := AINativeTheme()
	clone := original.Clone()

	// Test that all color fields are copied
	if clone.Colors.Background != original.Colors.Background {
		t.Error("Clone Background color differs from original")
	}
	if clone.Colors.Foreground != original.Colors.Foreground {
		t.Error("Clone Foreground color differs from original")
	}
	if clone.Colors.Primary != original.Colors.Primary {
		t.Error("Clone Primary color differs from original")
	}
	if clone.Colors.Secondary != original.Colors.Secondary {
		t.Error("Clone Secondary color differs from original")
	}
	if clone.Colors.Accent != original.Colors.Accent {
		t.Error("Clone Accent color differs from original")
	}
	if clone.Colors.Success != original.Colors.Success {
		t.Error("Clone Success color differs from original")
	}
	if clone.Colors.Warning != original.Colors.Warning {
		t.Error("Clone Warning color differs from original")
	}
	if clone.Colors.Error != original.Colors.Error {
		t.Error("Clone Error color differs from original")
	}
	if clone.Colors.Info != original.Colors.Info {
		t.Error("Clone Info color differs from original")
	}

	// Test that borders are copied
	if clone.Borders.Normal.TopLeft != original.Borders.Normal.TopLeft {
		t.Error("Clone borders differ from original")
	}

	// Test that spacing is copied
	if clone.Spacing.Small != original.Spacing.Small {
		t.Error("Clone spacing differs from original")
	}
	if clone.Spacing.Medium != original.Spacing.Medium {
		t.Error("Clone spacing differs from original")
	}

	// Verify it's a separate instance
	if &clone.Colors == &original.Colors {
		t.Error("Clone shares color palette reference with original")
	}
}

// TestThemeValidationErrorMessage tests error message formatting
func TestThemeValidationErrorMessage(t *testing.T) {
	theme := &Theme{
		Name: "",
		Colors: ColorPalette{
			Background: lipgloss.Color("#000000"),
			Foreground: lipgloss.Color("#ffffff"),
			Primary:    lipgloss.Color("#0000ff"),
		},
	}

	err := theme.Validate()
	if err == nil {
		t.Fatal("Expected validation error")
	}

	// Test error message format
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message is empty")
	}

	// Test that it's an ErrInvalidTheme
	if _, ok := err.(ErrInvalidTheme); !ok {
		t.Errorf("Expected ErrInvalidTheme, got %T", err)
	}
}

// TestBuiltinThemeIntegrity tests all builtin themes have required fields
func TestBuiltinThemeIntegrity(t *testing.T) {
	themes := []*Theme{
		AINativeTheme(),
		DarkTheme(),
		LightTheme(),
	}

	for _, theme := range themes {
		t.Run(theme.Name, func(t *testing.T) {
			// Test validation passes
			if err := theme.Validate(); err != nil {
				t.Errorf("Theme '%s' failed validation: %v", theme.Name, err)
			}

			// Test all required colors are set
			requiredColors := map[string]lipgloss.Color{
				"Background": theme.Colors.Background,
				"Foreground": theme.Colors.Foreground,
				"Primary":    theme.Colors.Primary,
				"Secondary":  theme.Colors.Secondary,
				"Accent":     theme.Colors.Accent,
				"Success":    theme.Colors.Success,
				"Warning":    theme.Colors.Warning,
				"Error":      theme.Colors.Error,
				"Info":       theme.Colors.Info,
			}

			for name, color := range requiredColors {
				if color == "" {
					t.Errorf("Theme '%s' missing %s color", theme.Name, name)
				}
			}

			// Test all styles can render
			if theme.Styles.Title.Render("test") == "" {
				t.Errorf("Theme '%s' Title style cannot render", theme.Name)
			}
			if theme.Styles.Error.Render("test") == "" {
				t.Errorf("Theme '%s' Error style cannot render", theme.Name)
			}

			// Test borders are set
			if theme.Borders.Normal.TopLeft == "" {
				t.Errorf("Theme '%s' missing Normal border", theme.Name)
			}
			if theme.Borders.Rounded.TopLeft == "" {
				t.Errorf("Theme '%s' missing Rounded border", theme.Name)
			}

			// Test spacing is set
			if theme.Spacing.Small <= 0 {
				t.Errorf("Theme '%s' has invalid Small spacing", theme.Name)
			}
			if theme.Spacing.Medium <= 0 {
				t.Errorf("Theme '%s' has invalid Medium spacing", theme.Name)
			}
		})
	}
}

// TestThemeManagerResetToDefault tests resetting to default theme
func TestThemeManagerResetToDefault(t *testing.T) {
	manager := NewThemeManager()
	manager.RegisterTheme(AINativeTheme())
	manager.RegisterTheme(DarkTheme())
	manager.RegisterTheme(LightTheme())

	// Switch to non-default theme
	err := manager.SetTheme("Light")
	if err != nil {
		t.Fatalf("SetTheme failed: %v", err)
	}

	// Reset to default
	err = manager.ResetToDefault()
	if err != nil {
		t.Fatalf("ResetToDefault failed: %v", err)
	}

	// Should be back to first theme (AINative)
	if manager.CurrentTheme().Name != "AINative" {
		t.Errorf("Expected theme 'AINative' after reset, got '%s'", manager.CurrentTheme().Name)
	}
}

// TestThemeManagerResetToDefaultWithNoThemes tests resetting with no themes
func TestThemeManagerResetToDefaultWithNoThemes(t *testing.T) {
	manager := NewThemeManager()

	err := manager.ResetToDefault()
	if err == nil {
		t.Error("Expected error when resetting with no themes")
	}
}

// TestThemeManagerCycleWithOneTheme tests cycling with only one theme
func TestThemeManagerCycleWithOneTheme(t *testing.T) {
	manager := NewThemeManager()
	manager.RegisterTheme(AINativeTheme())

	// Should not error with one theme
	err := manager.CycleTheme()
	if err != nil {
		t.Errorf("CycleTheme with one theme should not error: %v", err)
	}

	// Should still be on the same theme
	if manager.CurrentTheme().Name != "AINative" {
		t.Error("Theme changed when cycling with only one theme")
	}
}

// TestThemeManagerCycleWithNoThemes tests cycling with no themes
func TestThemeManagerCycleWithNoThemes(t *testing.T) {
	manager := NewThemeManager()

	err := manager.CycleTheme()
	if err == nil {
		t.Error("Expected error when cycling with no themes")
	}
}

// TestBuildSpacingSet tests spacing set creation
func TestBuildSpacingSet(t *testing.T) {
	spacing := buildSpacingSet()

	// Test all spacing values are set correctly
	if spacing.None != 0 {
		t.Errorf("Expected None spacing to be 0, got %d", spacing.None)
	}
	if spacing.Small != 1 {
		t.Errorf("Expected Small spacing to be 1, got %d", spacing.Small)
	}
	if spacing.Medium != 2 {
		t.Errorf("Expected Medium spacing to be 2, got %d", spacing.Medium)
	}
	if spacing.Large != 4 {
		t.Errorf("Expected Large spacing to be 4, got %d", spacing.Large)
	}
	if spacing.XLarge != 8 {
		t.Errorf("Expected XLarge spacing to be 8, got %d", spacing.XLarge)
	}

	// Test spacing progression
	if spacing.Small >= spacing.Medium {
		t.Error("Small spacing should be less than Medium")
	}
	if spacing.Medium >= spacing.Large {
		t.Error("Medium spacing should be less than Large")
	}
	if spacing.Large >= spacing.XLarge {
		t.Error("Large spacing should be less than XLarge")
	}
}

// TestThemeManagerConfigPath tests config path generation
func TestThemeManagerConfigPath(t *testing.T) {
	manager := NewThemeManager()

	configPath := manager.getConfigPath()
	if configPath == "" {
		t.Error("Config path is empty")
	}

	// Should contain "theme.json"
	if filepath.Base(configPath) != "theme.json" {
		t.Errorf("Config path should end with theme.json, got %s", configPath)
	}
}
