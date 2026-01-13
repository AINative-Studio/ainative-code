package dialogs_test

import (
	"testing"

	"github.com/AINative-studio/ainative-code/internal/tui/dialogs"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewConfirmDialog(t *testing.T) {
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:          "test",
		Title:       "Confirm Test",
		Description: "This is a test",
		YesLabel:    "Accept",
		NoLabel:     "Decline",
		DefaultYes:  true,
	})

	if dialog == nil {
		t.Fatal("NewConfirmDialog returned nil")
	}

	if dialog.ID() != "test" {
		t.Errorf("Expected ID 'test', got '%s'", dialog.ID())
	}

	if dialog.IsClosing() {
		t.Error("New dialog should not be closing")
	}

	// Note: ConfirmDialog has a result field that may have a default value
	// We check that it's not closing, which is the important state
}

func TestConfirmDialogDefaults(t *testing.T) {
	// Test with minimal config
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		Title: "Test",
	})

	if dialog.ID() == "" {
		t.Error("Default ID should not be empty")
	}

	// View should work with defaults
	view := dialog.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestConfirmDialogNavigation(t *testing.T) {
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:         "test",
		Title:      "Test",
		DefaultYes: false, // Start on No
	})

	// Test right arrow (move to No)
	dialog.Update(tea.KeyMsg{Type: tea.KeyRight})

	// Test left arrow (move to Yes)
	dialog.Update(tea.KeyMsg{Type: tea.KeyLeft})

	// Test tab key
	dialog.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Should not be closing yet
	if dialog.IsClosing() {
		t.Error("Dialog should not be closing after navigation")
	}
}

func TestConfirmDialogYesSelection(t *testing.T) {
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:         "test",
		Title:      "Test",
		DefaultYes: true,
	})

	// Press Enter to confirm Yes
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !dialog.IsClosing() {
		t.Error("Dialog should be closing after Enter")
	}

	result := dialog.GetResult()
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if !*result {
		t.Error("Result should be true (Yes)")
	}
}

func TestConfirmDialogNoSelection(t *testing.T) {
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:         "test",
		Title:      "Test",
		DefaultYes: false, // Start on No
	})

	// Press Enter to confirm No
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !dialog.IsClosing() {
		t.Error("Dialog should be closing after Enter")
	}

	result := dialog.GetResult()
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if *result {
		t.Error("Result should be false (No)")
	}
}

func TestConfirmDialogQuickYes(t *testing.T) {
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	// Press 'y' for quick Yes
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	if !dialog.IsClosing() {
		t.Error("Dialog should be closing after 'y'")
	}

	result := dialog.GetResult()
	if result == nil || !*result {
		t.Error("Result should be true after 'y'")
	}
}

func TestConfirmDialogQuickNo(t *testing.T) {
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	// Press 'n' for quick No
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if !dialog.IsClosing() {
		t.Error("Dialog should be closing after 'n'")
	}

	result := dialog.GetResult()
	if result == nil || *result {
		t.Error("Result should be false after 'n'")
	}
}

func TestConfirmDialogEscape(t *testing.T) {
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	// Press ESC to cancel
	dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !dialog.IsClosing() {
		t.Error("Dialog should be closing after ESC")
	}

	result := dialog.GetResult()
	if result == nil {
		t.Fatal("Result should not be nil after ESC")
	}

	if *result {
		t.Error("ESC should result in false (No)")
	}
}

func TestConfirmDialogView(t *testing.T) {
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:          "test",
		Title:       "Confirm Action",
		Description: "Are you sure you want to proceed?",
		YesLabel:    "Proceed",
		NoLabel:     "Cancel",
	})

	view := dialog.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// View should contain title
	// (We can't easily test exact content due to styling, but length check is reasonable)
	if len(view) < 20 {
		t.Error("View seems too short to contain all dialog content")
	}
}

func TestConfirmDialogResize(t *testing.T) {
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	// Test resize
	dialog.SetSize(100, 50)

	// Should not crash or cause issues
	view := dialog.View()
	if view == "" {
		t.Error("View should still work after resize")
	}
}

func TestConfirmDialogInit(t *testing.T) {
	dialog := dialogs.NewConfirmDialog(dialogs.ConfirmDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	cmd := dialog.Init()

	// ConfirmDialog.Init() returns nil
	if cmd != nil {
		t.Error("ConfirmDialog.Init() should return nil")
	}
}
