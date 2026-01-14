package components

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// MockComponent is a simple mock component for testing
type MockComponent struct {
	initCalled   bool
	updateCalled int
	viewCalled   int
	lastMsg      tea.Msg
}

func (m *MockComponent) Init() tea.Cmd {
	m.initCalled = true
	return nil
}

func (m *MockComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	m.updateCalled++
	m.lastMsg = msg
	return m, nil
}

func (m *MockComponent) View() string {
	m.viewCalled++
	return "mock component"
}

func TestNewAnimatedComponent(t *testing.T) {
	mock := &MockComponent{}
	config := DefaultAnimationConfig()

	ac := NewAnimatedComponent(mock, config)

	if ac == nil {
		t.Fatal("NewAnimatedComponent returned nil")
	}

	if ac.component != mock {
		t.Error("Component not properly wrapped")
	}

	if ac.config.FPS != 60 {
		t.Errorf("Expected FPS 60, got %d", ac.config.FPS)
	}

	if ac.isAnimating {
		t.Error("Animation should not be running initially")
	}
}

func TestAnimationDefaults(t *testing.T) {
	mock := &MockComponent{}
	config := AnimationConfig{} // Empty config, should use defaults

	ac := NewAnimatedComponent(mock, config)

	if ac.config.FPS != 60 {
		t.Errorf("Expected default FPS 60, got %d", ac.config.FPS)
	}

	if ac.config.Duration != 300*time.Millisecond {
		t.Errorf("Expected default duration 300ms, got %v", ac.config.Duration)
	}

	if ac.config.AngularFrequency != 6.0 {
		t.Errorf("Expected default angular frequency 6.0, got %f", ac.config.AngularFrequency)
	}

	if ac.config.DampingRatio != 1.0 {
		t.Errorf("Expected default damping ratio 1.0, got %f", ac.config.DampingRatio)
	}
}

func TestStartAnimation(t *testing.T) {
	mock := &MockComponent{}
	config := AnimationConfig{
		Duration:         100 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 6.0,
		DampingRatio:     1.0,
	}

	ac := NewAnimatedComponent(mock, config)

	cmd := ac.StartAnimation(0.0, 100.0)

	if cmd == nil {
		t.Fatal("StartAnimation should return a command")
	}

	if !ac.IsAnimating() {
		t.Error("Animation should be running after StartAnimation")
	}

	if ac.from != 0.0 {
		t.Errorf("Expected from value 0.0, got %f", ac.from)
	}

	if ac.to != 100.0 {
		t.Errorf("Expected to value 100.0, got %f", ac.to)
	}

	if ac.current != 0.0 {
		t.Errorf("Expected current value 0.0, got %f", ac.current)
	}
}

func TestStopAnimation(t *testing.T) {
	mock := &MockComponent{}
	config := DefaultAnimationConfig()

	ac := NewAnimatedComponent(mock, config)
	ac.StartAnimation(0.0, 100.0)

	if !ac.IsAnimating() {
		t.Fatal("Animation should be running")
	}

	ac.StopAnimation()

	if ac.IsAnimating() {
		t.Error("Animation should be stopped")
	}
}

func TestAnimationTick(t *testing.T) {
	mock := &MockComponent{}
	config := AnimationConfig{
		Duration:         100 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 6.0,
		DampingRatio:     1.0,
	}

	ac := NewAnimatedComponent(mock, config)
	ac.StartAnimation(0.0, 100.0)

	// Simulate a tick message
	tickMsg := AnimationTickMsg{
		ID:    ac.GetID(),
		Value: 50.0,
		Time:  time.Now(),
	}

	newComponent, cmd := ac.Update(tickMsg)
	ac = newComponent.(*AnimatedComponent)

	if cmd == nil {
		t.Error("Update should return a command for next tick")
	}

	// Check that wrapped component received the message
	if mock.updateCalled == 0 {
		t.Error("Wrapped component should have received the message")
	}
}

func TestAnimationIgnoresWrongID(t *testing.T) {
	mock := &MockComponent{}
	config := DefaultAnimationConfig()

	ac := NewAnimatedComponent(mock, config)
	ac.StartAnimation(0.0, 100.0)

	initialValue := ac.GetValue()

	// Send tick with wrong ID
	tickMsg := AnimationTickMsg{
		ID:    "wrong-id",
		Value: 50.0,
		Time:  time.Now(),
	}

	newComponent, _ := ac.Update(tickMsg)
	ac = newComponent.(*AnimatedComponent)

	// Value should not change
	if ac.GetValue() != initialValue {
		t.Error("Animation should ignore ticks with wrong ID")
	}
}

func TestGetValue(t *testing.T) {
	mock := &MockComponent{}
	config := DefaultAnimationConfig()

	ac := NewAnimatedComponent(mock, config)
	ac.SetValue(42.5)

	value := ac.GetValue()
	if value != 42.5 {
		t.Errorf("Expected value 42.5, got %f", value)
	}
}

func TestSetValue(t *testing.T) {
	mock := &MockComponent{}
	config := DefaultAnimationConfig()

	ac := NewAnimatedComponent(mock, config)
	ac.SetValue(123.456)

	if ac.current != 123.456 {
		t.Errorf("Expected current value 123.456, got %f", ac.current)
	}
}

func TestProgress(t *testing.T) {
	mock := &MockComponent{}
	config := DefaultAnimationConfig()

	ac := NewAnimatedComponent(mock, config)
	ac.from = 0.0
	ac.to = 100.0
	ac.current = 50.0

	progress := ac.Progress()
	if progress != 0.5 {
		t.Errorf("Expected progress 0.5, got %f", progress)
	}

	ac.current = 0.0
	progress = ac.Progress()
	if progress != 0.0 {
		t.Errorf("Expected progress 0.0 at start, got %f", progress)
	}

	ac.current = 100.0
	progress = ac.Progress()
	if progress != 1.0 {
		t.Errorf("Expected progress 1.0 at end, got %f", progress)
	}
}

func TestProgressSameFromTo(t *testing.T) {
	mock := &MockComponent{}
	config := DefaultAnimationConfig()

	ac := NewAnimatedComponent(mock, config)
	ac.from = 50.0
	ac.to = 50.0
	ac.current = 50.0

	progress := ac.Progress()
	if progress != 1.0 {
		t.Errorf("Expected progress 1.0 when from equals to, got %f", progress)
	}
}

func TestReset(t *testing.T) {
	mock := &MockComponent{}
	config := DefaultAnimationConfig()

	ac := NewAnimatedComponent(mock, config)
	ac.StartAnimation(0.0, 100.0)
	ac.current = 50.0
	ac.reversed = true

	ac.Reset()

	if ac.IsAnimating() {
		t.Error("Animation should be stopped after reset")
	}

	if ac.current != 0.0 {
		t.Errorf("Expected current to be reset to from value, got %f", ac.current)
	}

	if ac.reversed {
		t.Error("Reversed flag should be reset")
	}
}

func TestComponentPassthrough(t *testing.T) {
	mock := &MockComponent{}
	config := DefaultAnimationConfig()

	ac := NewAnimatedComponent(mock, config)

	// Test Init
	ac.Init()
	if !mock.initCalled {
		t.Error("Init should be passed to wrapped component")
	}

	// Test View
	view := ac.View()
	if view != "mock component" {
		t.Errorf("Expected 'mock component', got '%s'", view)
	}
	if mock.viewCalled == 0 {
		t.Error("View should be passed to wrapped component")
	}

	// Test Update with non-animation message
	customMsg := tea.KeyMsg{}
	ac.Update(customMsg)
	if mock.updateCalled == 0 {
		t.Error("Update should be passed to wrapped component")
	}
}

func TestAnimationWithLoop(t *testing.T) {
	mock := &MockComponent{}
	config := AnimationConfig{
		Duration:         50 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 6.0,
		DampingRatio:     1.0,
		Loop:             true,
		Reverse:          false,
	}

	ac := NewAnimatedComponent(mock, config)
	ac.StartAnimation(0.0, 100.0)

	if !ac.config.Loop {
		t.Error("Loop should be enabled")
	}
}

func TestAnimationWithReverse(t *testing.T) {
	mock := &MockComponent{}
	config := AnimationConfig{
		Duration:         50 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 6.0,
		DampingRatio:     1.0,
		Loop:             true,
		Reverse:          true,
	}

	ac := NewAnimatedComponent(mock, config)
	ac.StartAnimation(0.0, 100.0)

	if !ac.config.Reverse {
		t.Error("Reverse should be enabled")
	}
}

func TestGetID(t *testing.T) {
	mock := &MockComponent{}
	config := DefaultAnimationConfig()

	ac := NewAnimatedComponent(mock, config)
	id1 := ac.GetID()

	if id1 == "" {
		t.Error("GetID should return non-empty ID")
	}

	// Sleep briefly to ensure different nanosecond timestamp
	time.Sleep(1 * time.Microsecond)

	// Create another component, should have different ID
	ac2 := NewAnimatedComponent(mock, config)
	id2 := ac2.GetID()

	if id1 == id2 {
		t.Error("Different components should have different IDs")
	}
}

func TestNewAnimatedComponentWithID(t *testing.T) {
	mock := &MockComponent{}
	config := DefaultAnimationConfig()
	customID := "custom-animation-id"

	ac := NewAnimatedComponentWithID(customID, mock, config)

	if ac.GetID() != customID {
		t.Errorf("Expected ID '%s', got '%s'", customID, ac.GetID())
	}
}

func TestComponentMethod(t *testing.T) {
	mock := &MockComponent{}
	config := DefaultAnimationConfig()

	ac := NewAnimatedComponent(mock, config)
	wrapped := ac.Component()

	if wrapped != mock {
		t.Error("Component() should return the wrapped component")
	}
}

// Test transition presets
func TestFadeInComponent(t *testing.T) {
	mock := &MockComponent{}
	ac := FadeInComponent(mock)

	if ac == nil {
		t.Fatal("FadeInComponent returned nil")
	}

	if ac.config.Duration != 200*time.Millisecond {
		t.Errorf("Expected FadeIn duration 200ms, got %v", ac.config.Duration)
	}
}

func TestFadeOutComponent(t *testing.T) {
	mock := &MockComponent{}
	ac := FadeOutComponent(mock)

	if ac == nil {
		t.Fatal("FadeOutComponent returned nil")
	}

	if ac.config.Duration != 150*time.Millisecond {
		t.Errorf("Expected FadeOut duration 150ms, got %v", ac.config.Duration)
	}
}

func TestSlideInComponent(t *testing.T) {
	mock := &MockComponent{}
	ac := SlideInComponent(mock, -100, 0)

	if ac == nil {
		t.Fatal("SlideInComponent returned nil")
	}

	if ac.config.Duration != 250*time.Millisecond {
		t.Errorf("Expected SlideIn duration 250ms, got %v", ac.config.Duration)
	}
}

func TestSpringComponent(t *testing.T) {
	mock := &MockComponent{}
	ac := SpringComponent(mock)

	if ac == nil {
		t.Fatal("SpringComponent returned nil")
	}

	if ac.config.Duration != 400*time.Millisecond {
		t.Errorf("Expected Spring duration 400ms, got %v", ac.config.Duration)
	}
}

func TestSpinnerComponent(t *testing.T) {
	mock := &MockComponent{}
	ac := SpinnerComponent(mock)

	if ac == nil {
		t.Fatal("SpinnerComponent returned nil")
	}

	if !ac.config.Loop {
		t.Error("Spinner should have Loop enabled")
	}

	if ac.config.Duration != 1000*time.Millisecond {
		t.Errorf("Expected Spinner duration 1000ms, got %v", ac.config.Duration)
	}
}

func TestGetTransitionPreset(t *testing.T) {
	config, exists := GetTransitionPreset("fadeIn")
	if !exists {
		t.Error("fadeIn preset should exist")
	}
	if config.Duration != 200*time.Millisecond {
		t.Errorf("Expected fadeIn duration 200ms, got %v", config.Duration)
	}

	_, exists = GetTransitionPreset("nonexistent")
	if exists {
		t.Error("nonexistent preset should not exist")
	}
}

func TestCustomTransition(t *testing.T) {
	config := CustomTransition(
		500*time.Millisecond,
		30,
		5.0,  // angular frequency
		0.8,  // damping ratio
		true,
		true,
	)

	if config.Duration != 500*time.Millisecond {
		t.Errorf("Expected duration 500ms, got %v", config.Duration)
	}

	if config.FPS != 30 {
		t.Errorf("Expected FPS 30, got %d", config.FPS)
	}

	if config.AngularFrequency != 5.0 {
		t.Errorf("Expected AngularFrequency 5.0, got %f", config.AngularFrequency)
	}

	if config.DampingRatio != 0.8 {
		t.Errorf("Expected DampingRatio 0.8, got %f", config.DampingRatio)
	}

	if !config.Loop {
		t.Error("Expected Loop to be true")
	}

	if !config.Reverse {
		t.Error("Expected Reverse to be true")
	}
}

func TestTransitionPresets(t *testing.T) {
	expectedPresets := []string{
		"fadeIn", "fadeOut", "slideIn", "slideOut",
		"spring", "bounce", "smooth", "spinner", "pulse", "snap",
	}

	for _, name := range expectedPresets {
		_, exists := TransitionPresets[name]
		if !exists {
			t.Errorf("Expected preset '%s' to exist in TransitionPresets", name)
		}
	}
}

func TestRotationAnimation(t *testing.T) {
	config := RotationAnimation(2*time.Second, true)

	if config.Duration != 2*time.Second {
		t.Errorf("Expected duration 2s, got %v", config.Duration)
	}

	if !config.Loop {
		t.Error("Expected Loop to be true for rotation")
	}

	if config.FPS != 60 {
		t.Errorf("Expected FPS 60, got %d", config.FPS)
	}
}
