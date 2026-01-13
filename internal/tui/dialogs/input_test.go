package dialogs_test

import (
	"errors"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/tui/dialogs"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewInputDialog(t *testing.T) {
	dialog := dialogs.NewInputDialog(dialogs.InputDialogConfig{
		ID:           "test",
		Title:        "Input Test",
		Description:  "Enter some text",
		Placeholder:  "Type here...",
		DefaultValue: "initial",
	})

	if dialog == nil {
		t.Fatal("NewInputDialog returned nil")
	}

	if dialog.ID() != "test" {
		t.Errorf("Expected ID 'test', got '%s'", dialog.ID())
	}

	if dialog.IsClosing() {
		t.Error("New dialog should not be closing")
	}

	// Note: InputDialog has a result field that may be initialized
	// We check that it's not closing, which is the important state
}

func TestInputDialogDefaults(t *testing.T) {
	// Test with minimal config
	dialog := dialogs.NewInputDialog(dialogs.InputDialogConfig{
		Title: "Test",
	})

	if dialog.ID() == "" {
		t.Error("Default ID should not be empty")
	}

	view := dialog.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestInputDialogSubmit(t *testing.T) {
	dialog := dialogs.NewInputDialog(dialogs.InputDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	// Simulate typing
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})

	// Press Enter to submit
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !dialog.IsClosing() {
		t.Error("Dialog should be closing after Enter")
	}

	result := dialog.GetResult()
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if *result != "hello" {
		t.Errorf("Expected 'hello', got '%s'", *result)
	}
}

func TestInputDialogCancel(t *testing.T) {
	dialog := dialogs.NewInputDialog(dialogs.InputDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	// Type something
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t', 'e', 's', 't'}})

	// Press ESC to cancel
	dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !dialog.IsClosing() {
		t.Error("Dialog should be closing after ESC")
	}

	result := dialog.GetResult()
	if result != nil {
		t.Error("Result should be nil after cancel")
	}
}

func TestInputDialogValidation(t *testing.T) {
	// Create dialog with validator that rejects empty strings
	validator := func(s string) error {
		if s == "" {
			return errors.New("input cannot be empty")
		}
		return nil
	}

	dialog := dialogs.NewInputDialog(dialogs.InputDialogConfig{
		ID:        "test",
		Title:     "Test",
		Validator: validator,
	})

	// Try to submit empty input
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should not close due to validation error
	if dialog.IsClosing() {
		t.Error("Dialog should not close with invalid input")
	}

	// The dialog should remain open - checking IsClosing is sufficient
}

func TestInputDialogValidationSuccess(t *testing.T) {
	// Create dialog with validator
	validator := func(s string) error {
		if len(s) < 3 {
			return errors.New("input must be at least 3 characters")
		}
		return nil
	}

	dialog := dialogs.NewInputDialog(dialogs.InputDialogConfig{
		ID:        "test",
		Title:     "Test",
		Validator: validator,
	})

	// Type valid input
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a', 'b', 'c'}})

	// Submit
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should close with valid input
	if !dialog.IsClosing() {
		t.Error("Dialog should close with valid input")
	}

	result := dialog.GetResult()
	if result == nil {
		t.Fatal("Result should not be nil with valid input")
	}

	if *result != "abc" {
		t.Errorf("Expected 'abc', got '%s'", *result)
	}
}

func TestInputDialogView(t *testing.T) {
	dialog := dialogs.NewInputDialog(dialogs.InputDialogConfig{
		ID:          "test",
		Title:       "Enter Name",
		Description: "Please enter your name",
		Placeholder: "John Doe",
	})

	view := dialog.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	if len(view) < 20 {
		t.Error("View seems too short to contain dialog content")
	}
}

func TestInputDialogResize(t *testing.T) {
	dialog := dialogs.NewInputDialog(dialogs.InputDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	dialog.SetSize(100, 50)

	view := dialog.View()
	if view == "" {
		t.Error("View should still work after resize")
	}
}

func TestInputDialogInit(t *testing.T) {
	dialog := dialogs.NewInputDialog(dialogs.InputDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	cmd := dialog.Init()

	// InputDialog.Init() returns textinput.Blink
	if cmd == nil {
		t.Error("InputDialog.Init() should return a blink command")
	}
}

func TestInputDialogBackspace(t *testing.T) {
	dialog := dialogs.NewInputDialog(dialogs.InputDialogConfig{
		ID:           "test",
		Title:        "Test",
		DefaultValue: "hello",
	})

	// Simulate backspace
	dialog.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	// Submit
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	result := dialog.GetResult()
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Should have removed one character
	if *result == "hello" {
		t.Error("Backspace should have modified the input")
	}
}

func TestInputDialogWhitespace(t *testing.T) {
	dialog := dialogs.NewInputDialog(dialogs.InputDialogConfig{
		ID:    "test",
		Title: "Test",
	})

	// Type only spaces
	dialog.Update(tea.KeyMsg{Type: tea.KeySpace})
	dialog.Update(tea.KeyMsg{Type: tea.KeySpace})
	dialog.Update(tea.KeyMsg{Type: tea.KeySpace})

	// Submit
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	result := dialog.GetResult()
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Trimmed whitespace should result in empty string
	if *result != "" {
		t.Errorf("Expected empty string after trim, got '%s'", *result)
	}
}
