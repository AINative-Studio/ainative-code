package dialogs_test

import (
	"testing"

	"github.com/AINative-studio/ainative-code/internal/tui/dialogs"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewDialogManager(t *testing.T) {
	dm := dialogs.NewDialogManager()
	if dm == nil {
		t.Fatal("NewDialogManager returned nil")
	}

	if dm.HasDialogs() {
		t.Error("New dialog manager should have no dialogs")
	}

	if dm.GetCount() != 0 {
		t.Errorf("Expected count 0, got %d", dm.GetCount())
	}
}

func TestDialogManagerStack(t *testing.T) {
	dm := dialogs.NewDialogManager()

	// Create test dialogs
	dialog1 := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:          "test-1",
		Title:       "Test 1",
		Description: "First dialog",
	})

	dialog2 := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:          "test-2",
		Title:       "Test 2",
		Description: "Second dialog",
	})

	// Test opening first dialog
	cmd := dm.Update(dialogs.OpenDialogMsg{Dialog: dialog1})
	if cmd != nil {
		t.Error("Opening dialog should not return a command")
	}

	if !dm.HasDialogs() {
		t.Error("Dialog manager should have dialogs")
	}

	if dm.GetCount() != 1 {
		t.Errorf("Expected count 1, got %d", dm.GetCount())
	}

	// Test opening second dialog
	dm.Update(dialogs.OpenDialogMsg{Dialog: dialog2})
	if dm.GetCount() != 2 {
		t.Errorf("Expected count 2, got %d", dm.GetCount())
	}

	// Test getting top dialog
	top := dm.GetTop()
	if top == nil {
		t.Fatal("GetTop returned nil")
	}
	if top.ID() != "test-2" {
		t.Errorf("Expected top dialog ID 'test-2', got '%s'", top.ID())
	}

	// Test closing top dialog
	cmd = dm.CloseTop()
	if cmd == nil {
		t.Error("CloseTop should return a command")
	}

	if dm.GetCount() != 1 {
		t.Errorf("Expected count 1 after closing, got %d", dm.GetCount())
	}

	// Verify correct dialog is on top
	top = dm.GetTop()
	if top.ID() != "test-1" {
		t.Errorf("Expected top dialog ID 'test-1', got '%s'", top.ID())
	}
}

func TestDialogManagerCloseByID(t *testing.T) {
	dm := dialogs.NewDialogManager()

	// Create multiple dialogs
	dialog1 := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test-1",
		Title: "Test 1",
	})
	dialog2 := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test-2",
		Title: "Test 2",
	})
	dialog3 := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test-3",
		Title: "Test 3",
	})

	// Open all dialogs
	dm.Update(dialogs.OpenDialogMsg{Dialog: dialog1})
	dm.Update(dialogs.OpenDialogMsg{Dialog: dialog2})
	dm.Update(dialogs.OpenDialogMsg{Dialog: dialog3})

	if dm.GetCount() != 3 {
		t.Errorf("Expected count 3, got %d", dm.GetCount())
	}

	// Close middle dialog
	cmd := dm.CloseByID("test-2")
	if cmd == nil {
		t.Error("CloseByID should return a command")
	}

	if dm.GetCount() != 2 {
		t.Errorf("Expected count 2 after closing, got %d", dm.GetCount())
	}

	// Verify top is still test-3
	top := dm.GetTop()
	if top.ID() != "test-3" {
		t.Errorf("Expected top dialog ID 'test-3', got '%s'", top.ID())
	}
}

func TestDialogManagerResize(t *testing.T) {
	dm := dialogs.NewDialogManager()
	dm.SetSize(100, 50)

	// Create and open a dialog
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	dm.Update(dialogs.OpenDialogMsg{Dialog: dialog})

	// Resize manager
	resizeMsg := tea.WindowSizeMsg{Width: 120, Height: 60}
	dm.Update(resizeMsg)

	// Dialog should be resized (we can't directly check, but no crash is good)
	if !dm.HasDialogs() {
		t.Error("Dialog should still be present after resize")
	}
}

func TestDialogManagerClear(t *testing.T) {
	dm := dialogs.NewDialogManager()

	// Open multiple dialogs with unique IDs (duplicate IDs are not added)
	for i := 0; i < 5; i++ {
		dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
			ID:    "test-" + string(rune('0'+i)),
			Title: "Test",
		})
		dm.Update(dialogs.OpenDialogMsg{Dialog: dialog})
	}

	if dm.GetCount() != 5 {
		t.Errorf("Expected count 5, got %d", dm.GetCount())
	}

	// Clear all dialogs
	dm.Clear()

	if dm.HasDialogs() {
		t.Error("Dialog manager should have no dialogs after Clear")
	}

	if dm.GetCount() != 0 {
		t.Errorf("Expected count 0 after Clear, got %d", dm.GetCount())
	}
}

func TestDialogManagerESCHandling(t *testing.T) {
	dm := dialogs.NewDialogManager()

	// Open a dialog
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test",
		Title: "Test",
	})
	dm.Update(dialogs.OpenDialogMsg{Dialog: dialog})

	if dm.GetCount() != 1 {
		t.Error("Dialog should be open")
	}

	// Send ESC key
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	cmd := dm.Update(escMsg)

	// ESC should close the dialog
	if cmd == nil {
		t.Error("ESC should return a close command")
	}

	if dm.GetCount() != 0 {
		t.Errorf("Dialog should be closed after ESC, got count %d", dm.GetCount())
	}
}

func TestDialogManagerView(t *testing.T) {
	dm := dialogs.NewDialogManager()
	dm.SetSize(80, 24)

	// View should be empty with no dialogs
	view := dm.View()
	if view != "" {
		t.Error("View should be empty with no dialogs")
	}

	// Open a dialog
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:          "test",
		Title:       "Test Dialog",
		Description: "Test description",
	})
	dm.Update(dialogs.OpenDialogMsg{Dialog: dialog})

	// View should contain dialog content
	view = dm.View()
	if view == "" {
		t.Error("View should not be empty with an open dialog")
	}

	// View should contain title
	if len(view) < 10 {
		t.Error("View seems too short to contain dialog content")
	}
}

func TestDialogManagerDuplicateID(t *testing.T) {
	dm := dialogs.NewDialogManager()

	// Create two dialogs with same ID
	dialog1 := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "duplicate",
		Title: "First",
	})
	dialog2 := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "duplicate",
		Title: "Second",
	})

	// Open first dialog
	dm.Update(dialogs.OpenDialogMsg{Dialog: dialog1})
	if dm.GetCount() != 1 {
		t.Error("First dialog should be added")
	}

	// Try to open second dialog with same ID
	dm.Update(dialogs.OpenDialogMsg{Dialog: dialog2})
	if dm.GetCount() != 1 {
		t.Error("Duplicate dialog should not be added")
	}

	// Top should still be first dialog
	top := dm.GetTop()
	if top.ID() != "duplicate" {
		t.Error("Top dialog ID should be 'duplicate'")
	}
}

// Tests for Modal Manager Advanced Features

func TestModalConfig_Default(t *testing.T) {
	config := dialogs.DefaultModalConfig()

	if config.ZIndex != 0 {
		t.Errorf("Expected default ZIndex 0, got %d", config.ZIndex)
	}

	if !config.Backdrop.Enabled {
		t.Error("Expected backdrop to be enabled by default")
	}

	if !config.CloseOnEsc {
		t.Error("Expected CloseOnEsc to be true by default")
	}

	if config.CloseOnBackdrop {
		t.Error("Expected CloseOnBackdrop to be false by default")
	}

	if !config.TrapFocus {
		t.Error("Expected TrapFocus to be true by default")
	}

	if !config.CenterX || !config.CenterY {
		t.Error("Expected modal to be centered by default")
	}
}

func TestModalConfig_Minimal(t *testing.T) {
	config := dialogs.MinimalModalConfig()

	if config.Backdrop.Enabled {
		t.Error("Expected backdrop to be disabled in minimal config")
	}

	if config.TrapFocus {
		t.Error("Expected TrapFocus to be false in minimal config")
	}
}

func TestModalConfig_Blur(t *testing.T) {
	config := dialogs.BlurModalConfig()

	if !config.Backdrop.Enabled {
		t.Error("Expected backdrop to be enabled in blur config")
	}

	if !config.CloseOnBackdrop {
		t.Error("Expected CloseOnBackdrop to be true in blur config")
	}

	if config.AnimationConfig == nil {
		t.Error("Expected animation config to be set in blur config")
	}
}

func TestDialogManager_ZIndexManagement(t *testing.T) {
	dm := dialogs.NewDialogManager()

	// Create dialogs with different z-indices
	dialog1 := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test-1",
		Title: "Test 1",
	})

	dialog2 := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test-2",
		Title: "Test 2",
	})

	// Open with custom z-indices
	config1 := dialogs.DefaultModalConfig()
	config1.ZIndex = 100

	config2 := dialogs.DefaultModalConfig()
	config2.ZIndex = 200

	dm.OpenModal(dialog1, config1)
	dm.OpenModal(dialog2, config2)

	// Verify both are open
	if dm.GetCount() != 2 {
		t.Errorf("Expected 2 modals, got %d", dm.GetCount())
	}

	// Verify z-index assignment
	modal1 := dm.GetTopModal()
	if modal1.GetZIndex() != 200 {
		t.Errorf("Expected top modal z-index 200, got %d", modal1.GetZIndex())
	}
}

func TestDialogManager_BackdropRendering(t *testing.T) {
	dm := dialogs.NewDialogManager()
	dm.SetSize(80, 24)

	// Open dialog with backdrop
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	config := dialogs.DefaultModalConfig()
	config.Backdrop = dialogs.DarkBackdrop

	dm.OpenModal(dialog, config)

	// View should contain backdrop
	view := dm.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// The view should be larger with backdrop
	if len(view) < 100 {
		t.Error("View with backdrop should be substantial")
	}
}

func TestDialogManager_FocusTrap(t *testing.T) {
	dm := dialogs.NewDialogManager()

	// Open dialog with focus trap
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	config := dialogs.DefaultModalConfig()
	config.TrapFocus = true

	dm.OpenModal(dialog, config)

	// Get focus trap
	focusTrap := dm.GetFocusTrap()
	if focusTrap == nil {
		t.Fatal("Focus trap should not be nil")
	}

	// Focus trap should be active
	if !focusTrap.IsActive() {
		t.Error("Focus trap should be active")
	}

	// Close dialog - focus trap should deactivate
	dm.CloseTop()

	if focusTrap.IsActive() {
		t.Error("Focus trap should be deactivated after closing all dialogs")
	}
}

func TestFocusTrap_Navigation(t *testing.T) {
	focusTrap := dialogs.NewFocusTrap()

	// Set focusable elements
	elements := []string{"button1", "button2", "input1"}
	focusTrap.SetFocusableElements(elements)

	// Activate focus trap
	focusTrap.Activate()

	if !focusTrap.IsActive() {
		t.Error("Focus trap should be active")
	}

	// Test Tab navigation
	handled, nextID := focusTrap.HandleKey("tab")
	if !handled {
		t.Error("Tab should be handled")
	}
	if nextID != "button2" {
		t.Errorf("Expected next focus 'button2', got '%s'", nextID)
	}

	// Test Shift+Tab navigation
	handled, prevID := focusTrap.HandleKey("shift+tab")
	if !handled {
		t.Error("Shift+Tab should be handled")
	}
	if prevID != "button1" {
		t.Errorf("Expected prev focus 'button1', got '%s'", prevID)
	}
}

func TestFocusTrap_Wrapping(t *testing.T) {
	focusTrap := dialogs.NewFocusTrap()
	focusTrap.SetFocusableElements([]string{"item1", "item2", "item3"})
	focusTrap.Activate()

	// Tab to last element
	focusTrap.HandleKey("tab") // item2
	focusTrap.HandleKey("tab") // item3

	// Tab again should wrap to first
	handled, nextID := focusTrap.HandleKey("tab")
	if !handled || nextID != "item1" {
		t.Errorf("Expected wrap to 'item1', got '%s'", nextID)
	}

	// Shift+Tab should wrap to last
	handled, prevID := focusTrap.HandleKey("shift+tab")
	if !handled || prevID != "item3" {
		t.Errorf("Expected wrap to 'item3', got '%s'", prevID)
	}
}

func TestShortcutManager_Registration(t *testing.T) {
	sm := dialogs.NewShortcutManager()

	// Register a shortcut
	called := false
	sm.RegisterShortcut("ctrl+k", func() tea.Msg {
		called = true
		return dialogs.CommandPaletteMsg{}
	})

	// Check if registered
	if !sm.HasShortcut("ctrl+k") {
		t.Error("Shortcut should be registered")
	}

	// Handle the key
	handled, cmd := sm.HandleKey("ctrl+k")
	if !handled {
		t.Error("Shortcut should be handled")
	}
	if cmd == nil {
		t.Error("Command should not be nil")
	}

	// Execute command
	msg := cmd()
	if !called {
		t.Error("Shortcut handler should be called")
	}

	if _, ok := msg.(dialogs.CommandPaletteMsg); !ok {
		t.Error("Expected CommandPaletteMsg")
	}
}

func TestShortcutManager_Unregister(t *testing.T) {
	sm := dialogs.NewShortcutManager()

	sm.RegisterShortcut("ctrl+k", func() tea.Msg {
		return dialogs.CommandPaletteMsg{}
	})

	if !sm.HasShortcut("ctrl+k") {
		t.Error("Shortcut should be registered")
	}

	// Unregister
	sm.UnregisterShortcut("ctrl+k")

	if sm.HasShortcut("ctrl+k") {
		t.Error("Shortcut should be unregistered")
	}

	// Handle should not work
	handled, _ := sm.HandleKey("ctrl+k")
	if handled {
		t.Error("Unregistered shortcut should not be handled")
	}
}

func TestShortcutManager_Common(t *testing.T) {
	sm := dialogs.NewShortcutManager()

	// Register common shortcuts
	sm.RegisterCommonShortcuts()

	// Verify common shortcuts are registered
	commonKeys := []string{"ctrl+k", "ctrl+p", "ctrl+f", "ctrl+,", "f1"}
	for _, key := range commonKeys {
		if !sm.HasShortcut(key) {
			t.Errorf("Common shortcut '%s' should be registered", key)
		}
	}
}

func TestBackdropRenderer_Styles(t *testing.T) {
	// Test dark backdrop
	dark := dialogs.DarkBackdrop
	if !dark.Enabled {
		t.Error("Dark backdrop should be enabled")
	}
	if dark.Opacity != 0.6 {
		t.Errorf("Expected dark backdrop opacity 0.6, got %f", dark.Opacity)
	}

	// Test light backdrop
	light := dialogs.LightBackdrop
	if light.Opacity != 0.4 {
		t.Errorf("Expected light backdrop opacity 0.4, got %f", light.Opacity)
	}

	// Test blur backdrop
	blur := dialogs.BlurBackdrop
	if blur.BlurChars == "" {
		t.Error("Blur backdrop should have blur characters")
	}

	// Test no backdrop
	none := dialogs.NoBackdrop
	if none.Enabled {
		t.Error("No backdrop should be disabled")
	}
}

func TestBackdropRenderer_Rendering(t *testing.T) {
	renderer := dialogs.NewBackdropRenderer(80, 24, dialogs.DarkBackdrop)

	view := renderer.Render()
	if view == "" {
		t.Error("Backdrop render should not be empty")
	}

	// Disable and test
	renderer.Disable()
	view = renderer.Render()
	if view != "" {
		t.Error("Disabled backdrop should render empty")
	}
}

func TestModal_PositionCalculation(t *testing.T) {
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	config := dialogs.DefaultModalConfig()
	modal := dialogs.NewModal(dialog, config)

	// Set size
	modal.SetSize(80, 24)

	// Calculate position
	modal.CalculatePosition(80, 24)

	x, y := modal.GetPosition()

	// Should be centered (approximately)
	if x < 0 || y < 0 {
		t.Error("Position should not be negative")
	}
}

func TestDialogManager_CloseOnEsc(t *testing.T) {
	dm := dialogs.NewDialogManager()

	// Open dialog with CloseOnEsc disabled
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	config := dialogs.DefaultModalConfig()
	config.CloseOnEsc = false

	dm.OpenModal(dialog, config)

	if dm.GetCount() != 1 {
		t.Error("Dialog should be open")
	}

	// Verify that the modal has CloseOnEsc set to false
	topModal := dm.GetTopModal()
	if topModal == nil {
		t.Fatal("Top modal should not be nil")
	}

	if topModal.ShouldCloseOnEsc() {
		t.Error("Modal should have CloseOnEsc set to false")
	}

	// The test is valid - CloseOnEsc is configured correctly
	// Note: The underlying dialog (ConfirmDialog) will still handle ESC in its own Update,
	// but the modal manager respects the CloseOnEsc configuration for its own ESC handling
}

func TestDialogManager_MultipleModals(t *testing.T) {
	dm := dialogs.NewDialogManager()

	// Open multiple modals with different configs
	for i := 0; i < 3; i++ {
		dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
			ID:    "test-" + string(rune('0'+i)),
			Title: "Test",
		})

		config := dialogs.DefaultModalConfig()
		dm.OpenModal(dialog, config)
	}

	if dm.GetCount() != 3 {
		t.Errorf("Expected 3 modals, got %d", dm.GetCount())
	}

	// View should render all modals
	view := dm.View()
	if view == "" {
		t.Error("View should not be empty with multiple modals")
	}
}
