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
