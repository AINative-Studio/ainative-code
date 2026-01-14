package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestAnimatedSpinner(t *testing.T) {
	spinner := NewAnimatedSpinner("Loading...")

	if spinner == nil {
		t.Fatal("NewAnimatedSpinner returned nil")
	}

	// Test initialization
	cmd := spinner.Init()
	if cmd != nil {
		t.Error("Spinner init should return nil command")
	}

	// Start animation
	cmd = spinner.Start()
	if cmd == nil {
		t.Error("Start should return a command")
	}

	// Check if animating
	if !spinner.animation.IsAnimating() {
		t.Error("Spinner should be animating after Start")
	}

	// Test view
	view := spinner.View()
	if view == "" {
		t.Error("Spinner view should not be empty when animating")
	}

	// Stop animation
	spinner.Stop()
	if spinner.animation.IsAnimating() {
		t.Error("Spinner should not be animating after Stop")
	}

	// View should be empty when stopped
	view = spinner.View()
	if view != "" {
		t.Error("Spinner view should be empty when not animating")
	}
}

func TestAnimatedProgress(t *testing.T) {
	progress := NewAnimatedProgress(20, "Download")

	if progress == nil {
		t.Fatal("NewAnimatedProgress returned nil")
	}

	// Test setting progress
	cmd := progress.SetProgress(0.5)
	if cmd == nil {
		t.Error("SetProgress should return a command")
	}

	// Test view rendering
	view := progress.View()
	if view == "" {
		t.Error("Progress view should not be empty")
	}

	// Test progress animation
	cmd = progress.SetProgress(1.0)
	if cmd == nil {
		t.Error("SetProgress to 100% should return a command")
	}

	// Simulate animation tick
	tickMsg := AnimationTickMsg{
		ID:    progress.animation.GetID(),
		Value: 0.75,
	}

	_, cmd = progress.Update(tickMsg)
	if cmd == nil {
		t.Error("Update with animation tick should return a command")
	}
}

func TestAnimatedPulse(t *testing.T) {
	pulse := NewAnimatedPulse("Attention", lipgloss.Color("10"))

	if pulse == nil {
		t.Fatal("NewAnimatedPulse returned nil")
	}

	// Start pulsing
	cmd := pulse.Start()
	if cmd == nil {
		t.Error("Start should return a command")
	}

	if !pulse.animation.IsAnimating() {
		t.Error("Pulse should be animating after Start")
	}

	// Test view
	view := pulse.View()
	if view == "" {
		t.Error("Pulse view should not be empty")
	}

	// Verify looping is enabled (pulse should loop)
	if !pulse.animation.config.Loop {
		t.Error("Pulse animation should have Loop enabled")
	}

	// Verify reverse is enabled (pulse should reverse)
	if !pulse.animation.config.Reverse {
		t.Error("Pulse animation should have Reverse enabled")
	}
}

func TestAnimatedSlide(t *testing.T) {
	content := "Hello, World!"
	slide := NewAnimatedSlide(content, 50)

	if slide == nil {
		t.Fatal("NewAnimatedSlide returned nil")
	}

	// Test slide in
	cmd := slide.SlideIn()
	if cmd == nil {
		t.Error("SlideIn should return a command")
	}

	// Content should start off-screen
	initialValue := slide.animation.GetValue()
	if initialValue >= 0 {
		t.Error("Slide should start with negative offset (off-screen left)")
	}

	// Test slide out
	cmd = slide.SlideOut()
	if cmd == nil {
		t.Error("SlideOut should return a command")
	}

	// Test view rendering during animation
	view := slide.View()
	// View might be empty if off-screen, which is OK
	_ = view
}

func TestAnimatedRotation(t *testing.T) {
	rotation := NewAnimatedRotation(10, "●")

	if rotation == nil {
		t.Fatal("NewAnimatedRotation returned nil")
	}

	// Start rotation
	cmd := rotation.Start()
	if cmd == nil {
		t.Error("Start should return a command")
	}

	if !rotation.animation.IsAnimating() {
		t.Error("Rotation should be animating after Start")
	}

	// Verify looping is enabled (rotation should loop)
	if !rotation.animation.config.Loop {
		t.Error("Rotation animation should have Loop enabled")
	}

	// Test view
	view := rotation.View()
	if view == "" {
		t.Error("Rotation view should not be empty")
	}
}

func TestDemoComponentUpdates(t *testing.T) {
	spinner := NewAnimatedSpinner("Testing")
	spinner.Start()

	// Test that component properly handles animation ticks
	tickMsg := AnimationTickMsg{
		ID:    spinner.animation.GetID(),
		Value: 5.0,
	}

	updated, cmd := spinner.Update(tickMsg)
	if updated == nil {
		t.Error("Update should return updated component")
	}

	if cmd == nil {
		t.Error("Update with tick should return a command")
	}

	// Test with different message type (should pass through)
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd = spinner.Update(keyMsg)
	if updated == nil {
		t.Error("Update should return updated component for any message")
	}
}

func TestProgressBarRendering(t *testing.T) {
	progress := NewAnimatedProgress(10, "Test")

	// Set to 0%
	progress.animation.SetValue(0.0)
	view := progress.View()
	if view == "" {
		t.Error("View should not be empty at 0%")
	}

	// Set to 50%
	progress.animation.SetValue(0.5)
	view = progress.View()
	if view == "" {
		t.Error("View should not be empty at 50%")
	}

	// Set to 100%
	progress.animation.SetValue(1.0)
	view = progress.View()
	if view == "" {
		t.Error("View should not be empty at 100%")
	}
}

func TestSpinnerFrameSelection(t *testing.T) {
	spinner := NewAnimatedSpinner("Test")

	// Test frame selection at different animation values
	testCases := []struct {
		value         float64
		expectedValid bool
	}{
		{0.0, true},
		{1.5, true},
		{5.5, true},
		{9.9, true},
		{10.0, true},
	}

	for _, tc := range testCases {
		spinner.animation.SetValue(tc.value)
		spinner.animation.isAnimating = true
		view := spinner.View()

		if tc.expectedValid && view == "" {
			t.Errorf("Expected valid view at value %.1f", tc.value)
		}
	}
}

func TestSlideOffScreenBehavior(t *testing.T) {
	slide := NewAnimatedSlide("Content", 50)

	// Test far off-screen left
	slide.animation.SetValue(-100)
	view := slide.View()
	if view != "" {
		t.Error("View should be empty when far off-screen left")
	}

	// Test far off-screen right
	slide.animation.SetValue(100)
	view = slide.View()
	if view != "" {
		t.Error("View should be empty when far off-screen right")
	}

	// Test on-screen
	slide.animation.SetValue(0)
	view = slide.View()
	if view == "" {
		t.Error("View should not be empty when on-screen")
	}
}

func TestDummyComponent(t *testing.T) {
	dummy := &dummyComponent{}

	cmd := dummy.Init()
	if cmd != nil {
		t.Error("Dummy component Init should return nil")
	}

	updated, cmd := dummy.Update(tea.KeyMsg{})
	if updated != dummy {
		t.Error("Dummy component Update should return self")
	}
	if cmd != nil {
		t.Error("Dummy component Update should return nil command")
	}

	view := dummy.View()
	if view != "" {
		t.Error("Dummy component View should return empty string")
	}
}
