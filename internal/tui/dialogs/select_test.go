package dialogs_test

import (
	"testing"

	"github.com/AINative-studio/ainative-code/internal/tui/dialogs"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSelectDialog(t *testing.T) {
	options := []dialogs.SelectOption{
		{Label: "Option 1", Value: "opt1", Description: "First option"},
		{Label: "Option 2", Value: "opt2", Description: "Second option"},
		{Label: "Option 3", Value: "opt3", Description: "Third option"},
	}

	dialog := dialogs.NewSelectDialog(dialogs.SelectDialogConfig{
		ID:          "test",
		Title:       "Select Test",
		Description: "Choose an option",
		Options:     options,
		DefaultIdx:  0,
		Searchable:  true,
	})

	if dialog == nil {
		t.Fatal("NewSelectDialog returned nil")
	}

	if dialog.ID() != "test" {
		t.Errorf("Expected ID 'test', got '%s'", dialog.ID())
	}

	if dialog.IsClosing() {
		t.Error("New dialog should not be closing")
	}

	// Note: SelectDialog has a result field that may be initialized
	// We check that it's not closing, which is the important state
}

func TestSelectDialogDefaults(t *testing.T) {
	options := []dialogs.SelectOption{
		{Label: "Option 1", Value: "opt1"},
	}

	dialog := dialogs.NewSelectDialog(dialogs.SelectDialogConfig{
		Title:   "Test",
		Options: options,
	})

	if dialog.ID() == "" {
		t.Error("Default ID should not be empty")
	}

	view := dialog.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestSelectDialogNavigation(t *testing.T) {
	options := []dialogs.SelectOption{
		{Label: "Option 1", Value: "opt1"},
		{Label: "Option 2", Value: "opt2"},
		{Label: "Option 3", Value: "opt3"},
	}

	dialog := dialogs.NewSelectDialog(dialogs.SelectDialogConfig{
		ID:         "test",
		Title:      "Test",
		Options:    options,
		DefaultIdx: 0,
	})

	// Test down navigation
	dialog.Update(tea.KeyMsg{Type: tea.KeyDown})
	dialog.Update(tea.KeyMsg{Type: tea.KeyDown})

	// Should not be closing
	if dialog.IsClosing() {
		t.Error("Dialog should not be closing after navigation")
	}

	// Test up navigation
	dialog.Update(tea.KeyMsg{Type: tea.KeyUp})

	if dialog.IsClosing() {
		t.Error("Dialog should not be closing after navigation")
	}
}

func TestSelectDialogSelection(t *testing.T) {
	options := []dialogs.SelectOption{
		{Label: "Option 1", Value: "opt1"},
		{Label: "Option 2", Value: "opt2"},
		{Label: "Option 3", Value: "opt3"},
	}

	dialog := dialogs.NewSelectDialog(dialogs.SelectDialogConfig{
		ID:         "test",
		Title:      "Test",
		Options:    options,
		DefaultIdx: 1, // Start on second option
	})

	// Press Enter to select
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !dialog.IsClosing() {
		t.Error("Dialog should be closing after Enter")
	}

	result := dialog.GetResult()
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if *result != "opt2" {
		t.Errorf("Expected 'opt2', got '%s'", *result)
	}
}

func TestSelectDialogCancel(t *testing.T) {
	options := []dialogs.SelectOption{
		{Label: "Option 1", Value: "opt1"},
	}

	dialog := dialogs.NewSelectDialog(dialogs.SelectDialogConfig{
		ID:      "test",
		Title:   "Test",
		Options: options,
	})

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

func TestSelectDialogVimKeys(t *testing.T) {
	options := []dialogs.SelectOption{
		{Label: "Option 1", Value: "opt1"},
		{Label: "Option 2", Value: "opt2"},
	}

	dialog := dialogs.NewSelectDialog(dialogs.SelectDialogConfig{
		ID:      "test",
		Title:   "Test",
		Options: options,
	})

	// Test 'j' for down
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Test 'k' for up
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

	if dialog.IsClosing() {
		t.Error("Dialog should not be closing after vim navigation")
	}
}

func TestSelectDialogWrapAround(t *testing.T) {
	options := []dialogs.SelectOption{
		{Label: "Option 1", Value: "opt1"},
		{Label: "Option 2", Value: "opt2"},
		{Label: "Option 3", Value: "opt3"},
	}

	dialog := dialogs.NewSelectDialog(dialogs.SelectDialogConfig{
		ID:         "test",
		Title:      "Test",
		Options:    options,
		DefaultIdx: 0, // Start at first
	})

	// Go up from first (should wrap to last)
	dialog.Update(tea.KeyMsg{Type: tea.KeyUp})

	// Submit
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	result := dialog.GetResult()
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Should have wrapped to last option
	if *result != "opt3" {
		t.Errorf("Expected 'opt3' (wrap around), got '%s'", *result)
	}
}

func TestSelectDialogSearch(t *testing.T) {
	options := []dialogs.SelectOption{
		{Label: "Apple", Value: "apple"},
		{Label: "Banana", Value: "banana"},
		{Label: "Cherry", Value: "cherry"},
	}

	dialog := dialogs.NewSelectDialog(dialogs.SelectDialogConfig{
		ID:         "test",
		Title:      "Test",
		Options:    options,
		Searchable: true,
	})

	// Type to search (dialog starts in search mode if searchable)
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	// Submit (should select filtered option)
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !dialog.IsClosing() {
		t.Error("Dialog should be closing after Enter")
	}

	result := dialog.GetResult()
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if *result != "banana" {
		t.Errorf("Expected 'banana', got '%s'", *result)
	}
}

func TestSelectDialogSearchClear(t *testing.T) {
	options := []dialogs.SelectOption{
		{Label: "Option 1", Value: "opt1"},
		{Label: "Option 2", Value: "opt2"},
	}

	dialog := dialogs.NewSelectDialog(dialogs.SelectDialogConfig{
		ID:         "test",
		Title:      "Test",
		Options:    options,
		Searchable: true,
	})

	// Type search query
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x', 'y', 'z'}})

	// Press ESC to clear search (first ESC clears, second closes)
	dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Should not be closing yet (just cleared search)
	if dialog.IsClosing() {
		t.Error("Dialog should not close on first ESC (should clear search)")
	}
}

func TestSelectDialogEmptyOptions(t *testing.T) {
	dialog := dialogs.NewSelectDialog(dialogs.SelectDialogConfig{
		ID:      "test",
		Title:   "Test",
		Options: []dialogs.SelectOption{},
	})

	view := dialog.View()
	if view == "" {
		t.Error("View should not be empty even with no options")
	}

	// Should be able to cancel with ESC
	dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !dialog.IsClosing() {
		t.Error("Dialog should close on ESC even with no options")
	}
}

func TestSelectDialogView(t *testing.T) {
	options := []dialogs.SelectOption{
		{Label: "Option 1", Value: "opt1", Description: "First"},
		{Label: "Option 2", Value: "opt2", Description: "Second"},
	}

	dialog := dialogs.NewSelectDialog(dialogs.SelectDialogConfig{
		ID:          "test",
		Title:       "Select Item",
		Description: "Choose one",
		Options:     options,
	})

	view := dialog.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	if len(view) < 30 {
		t.Error("View seems too short to contain dialog content")
	}
}

func TestSelectDialogResize(t *testing.T) {
	options := []dialogs.SelectOption{
		{Label: "Option 1", Value: "opt1"},
	}

	dialog := dialogs.NewSelectDialog(dialogs.SelectDialogConfig{
		ID:      "test",
		Title:   "Test",
		Options: options,
	})

	dialog.SetSize(100, 50)

	view := dialog.View()
	if view == "" {
		t.Error("View should still work after resize")
	}
}

func TestSelectDialogGetSelectedOption(t *testing.T) {
	options := []dialogs.SelectOption{
		{Label: "Option 1", Value: "opt1", Description: "First"},
		{Label: "Option 2", Value: "opt2", Description: "Second"},
	}

	dialog := dialogs.NewSelectDialog(dialogs.SelectDialogConfig{
		ID:         "test",
		Title:      "Test",
		Options:    options,
		DefaultIdx: 1,
	})

	// Select option
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	selectedOpt := dialog.GetSelectedOption()
	if selectedOpt == nil {
		t.Fatal("GetSelectedOption should not return nil")
	}

	if selectedOpt.Value != "opt2" {
		t.Errorf("Expected value 'opt2', got '%s'", selectedOpt.Value)
	}

	if selectedOpt.Label != "Option 2" {
		t.Errorf("Expected label 'Option 2', got '%s'", selectedOpt.Label)
	}
}
