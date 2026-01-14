package components

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestAnimatedComponentView tests the View() method during various animation states
func TestAnimatedComponentView(t *testing.T) {
	t.Run("View() during animation", func(t *testing.T) {
		mock := newMockComponent("Animated Content")
		config := DefaultAnimationConfig()
		config.Duration = 100 * time.Millisecond

		ac := NewAnimatedComponent(mock, config)
		ac.StartAnimation(0.0, 1.0)

		// View should return wrapped component's view
		view := ac.View()
		if view != "Animated Content" {
			t.Errorf("Expected 'Animated Content', got '%s'", view)
		}
	})

	t.Run("View() at start position", func(t *testing.T) {
		mock := newMockComponent("Start Position")
		config := DefaultAnimationConfig()

		ac := NewAnimatedComponent(mock, config)
		ac.SetValue(0.0)

		view := ac.View()
		if view != "Start Position" {
			t.Errorf("Expected 'Start Position', got '%s'", view)
		}

		if ac.GetValue() != 0.0 {
			t.Errorf("Expected value 0.0, got %.2f", ac.GetValue())
		}
	})

	t.Run("View() at end position", func(t *testing.T) {
		mock := newMockComponent("End Position")
		config := DefaultAnimationConfig()

		ac := NewAnimatedComponent(mock, config)
		ac.SetValue(1.0)

		view := ac.View()
		if view != "End Position" {
			t.Errorf("Expected 'End Position', got '%s'", view)
		}

		if ac.GetValue() != 1.0 {
			t.Errorf("Expected value 1.0, got %.2f", ac.GetValue())
		}
	})

	t.Run("View() while not animating", func(t *testing.T) {
		mock := newMockComponent("Static Content")
		config := DefaultAnimationConfig()

		ac := NewAnimatedComponent(mock, config)
		ac.SetValue(0.5)

		if ac.IsAnimating() {
			t.Error("Expected animation to be inactive")
		}

		view := ac.View()
		if view != "Static Content" {
			t.Errorf("Expected 'Static Content', got '%s'", view)
		}
	})

	t.Run("View() with nil component", func(t *testing.T) {
		config := DefaultAnimationConfig()
		ac := NewAnimatedComponent(nil, config)

		view := ac.View()
		if view != "" {
			t.Errorf("Expected empty string for nil component, got '%s'", view)
		}
	})
}

// TestAnimatedComponentWithTransitions tests View() with different transition presets
func TestAnimatedComponentWithTransitions(t *testing.T) {
	t.Run("FadeInComponent View()", func(t *testing.T) {
		mock := newMockComponent("Fade In")
		ac := FadeInComponent(mock)

		view := ac.View()
		if view != "Fade In" {
			t.Errorf("Expected 'Fade In', got '%s'", view)
		}

		// Check config
		if ac.config.Duration != FadeIn.Duration {
			t.Errorf("Expected duration %v, got %v", FadeIn.Duration, ac.config.Duration)
		}
	})

	t.Run("FadeOutComponent View()", func(t *testing.T) {
		mock := newMockComponent("Fade Out")
		ac := FadeOutComponent(mock)

		view := ac.View()
		if view != "Fade Out" {
			t.Errorf("Expected 'Fade Out', got '%s'", view)
		}

		if ac.config.Duration != FadeOut.Duration {
			t.Errorf("Expected duration %v, got %v", FadeOut.Duration, ac.config.Duration)
		}
	})

	t.Run("SlideInComponent View()", func(t *testing.T) {
		mock := newMockComponent("Slide In")
		ac := SlideInComponent(mock, -100, 0)

		view := ac.View()
		if view != "Slide In" {
			t.Errorf("Expected 'Slide In', got '%s'", view)
		}

		if ac.config.Duration != SlideIn.Duration {
			t.Errorf("Expected duration %v, got %v", SlideIn.Duration, ac.config.Duration)
		}
	})

	t.Run("SpringComponent View()", func(t *testing.T) {
		mock := newMockComponent("Spring")
		ac := SpringComponent(mock)

		view := ac.View()
		if view != "Spring" {
			t.Errorf("Expected 'Spring', got '%s'", view)
		}

		if ac.config.DampingRatio != Spring.DampingRatio {
			t.Errorf("Expected damping ratio %.2f, got %.2f", Spring.DampingRatio, ac.config.DampingRatio)
		}
	})

	t.Run("PulseComponent View()", func(t *testing.T) {
		mock := newMockComponent("Pulse")
		ac := PulseComponent(mock)

		view := ac.View()
		if view != "Pulse" {
			t.Errorf("Expected 'Pulse', got '%s'", view)
		}

		if !ac.config.Loop || !ac.config.Reverse {
			t.Error("Expected pulse animation to loop and reverse")
		}
	})
}

// TestAnimatedComponentOpacity tests opacity calculations during fade animations
func TestAnimatedComponentOpacity(t *testing.T) {
	t.Run("Opacity at 0.0 (invisible)", func(t *testing.T) {
		mock := newMockComponent("Invisible")
		ac := FadeInComponent(mock)
		ac.SetValue(0.0)

		if ac.GetValue() != 0.0 {
			t.Errorf("Expected opacity 0.0, got %.2f", ac.GetValue())
		}

		// Component should still render
		view := ac.View()
		if view != "Invisible" {
			t.Errorf("Expected 'Invisible', got '%s'", view)
		}
	})

	t.Run("Opacity at 1.0 (fully visible)", func(t *testing.T) {
		mock := newMockComponent("Visible")
		ac := FadeInComponent(mock)
		ac.SetValue(1.0)

		if ac.GetValue() != 1.0 {
			t.Errorf("Expected opacity 1.0, got %.2f", ac.GetValue())
		}

		view := ac.View()
		if view != "Visible" {
			t.Errorf("Expected 'Visible', got '%s'", view)
		}
	})

	t.Run("Opacity during fade (0.5)", func(t *testing.T) {
		mock := newMockComponent("Half Visible")
		ac := FadeInComponent(mock)
		ac.from = 0.0
		ac.to = 1.0
		ac.SetValue(0.5)

		if ac.GetValue() != 0.5 {
			t.Errorf("Expected opacity 0.5, got %.2f", ac.GetValue())
		}

		progress := ac.Progress()
		if progress != 0.5 {
			t.Errorf("Expected progress 0.5, got %.2f", progress)
		}
	})
}

// TestAnimatedComponentProgress tests progress calculation
func TestAnimatedComponentProgress(t *testing.T) {
	t.Run("Progress at start (0%)", func(t *testing.T) {
		mock := newMockComponent("Test")
		ac := NewAnimatedComponent(mock, DefaultAnimationConfig())

		ac.StartAnimation(0.0, 1.0)
		progress := ac.Progress()

		if progress != 0.0 {
			t.Errorf("Expected progress 0.0, got %.2f", progress)
		}
	})

	t.Run("Progress at midpoint (50%)", func(t *testing.T) {
		mock := newMockComponent("Test")
		ac := NewAnimatedComponent(mock, DefaultAnimationConfig())

		ac.from = 0.0
		ac.to = 1.0
		ac.SetValue(0.5)

		progress := ac.Progress()
		if progress < 0.49 || progress > 0.51 {
			t.Errorf("Expected progress ~0.5, got %.2f", progress)
		}
	})

	t.Run("Progress at end (100%)", func(t *testing.T) {
		mock := newMockComponent("Test")
		ac := NewAnimatedComponent(mock, DefaultAnimationConfig())

		ac.from = 0.0
		ac.to = 1.0
		ac.SetValue(1.0)

		progress := ac.Progress()
		if progress != 1.0 {
			t.Errorf("Expected progress 1.0, got %.2f", progress)
		}
	})

	t.Run("Progress with same from/to", func(t *testing.T) {
		mock := newMockComponent("Test")
		ac := NewAnimatedComponent(mock, DefaultAnimationConfig())

		ac.from = 0.5
		ac.to = 0.5
		ac.SetValue(0.5)

		progress := ac.Progress()
		if progress != 1.0 {
			t.Errorf("Expected progress 1.0 when from==to, got %.2f", progress)
		}
	})

	t.Run("Progress with negative range", func(t *testing.T) {
		mock := newMockComponent("Test")
		ac := NewAnimatedComponent(mock, DefaultAnimationConfig())

		ac.from = 1.0
		ac.to = 0.0
		ac.SetValue(0.5)

		progress := ac.Progress()
		if progress < 0.49 || progress > 0.51 {
			t.Errorf("Expected progress ~0.5 for reverse animation, got %.2f", progress)
		}
	})
}

// TestAnimatedComponentInit tests Init() method
func TestAnimatedComponentInit(t *testing.T) {
	t.Run("Init() with wrapped component", func(t *testing.T) {
		mock := newMockComponent("Test")
		ac := NewAnimatedComponent(mock, DefaultAnimationConfig())

		cmd := ac.Init()
		// Mock component returns nil from Init()
		if cmd != nil {
			t.Error("Expected nil cmd from Init()")
		}
	})

	t.Run("Init() with nil component", func(t *testing.T) {
		ac := NewAnimatedComponent(nil, DefaultAnimationConfig())

		cmd := ac.Init()
		if cmd != nil {
			t.Error("Expected nil cmd from Init() with nil component")
		}
	})
}

// TestAnimatedComponentWrapper tests Component() accessor
func TestAnimatedComponentWrapper(t *testing.T) {
	t.Run("Component() returns wrapped component", func(t *testing.T) {
		mock := newMockComponent("Test")
		ac := NewAnimatedComponent(mock, DefaultAnimationConfig())

		wrapped := ac.Component()
		if wrapped != mock {
			t.Error("Expected Component() to return wrapped component")
		}
	})

	t.Run("Component() with nil", func(t *testing.T) {
		ac := NewAnimatedComponent(nil, DefaultAnimationConfig())

		wrapped := ac.Component()
		if wrapped != nil {
			t.Error("Expected Component() to return nil")
		}
	})
}

// TestAnimatedComponentID tests ID generation and management
func TestAnimatedComponentID(t *testing.T) {
	t.Run("GetID() returns non-empty ID", func(t *testing.T) {
		mock := newMockComponent("Test")
		ac := NewAnimatedComponent(mock, DefaultAnimationConfig())

		id := ac.GetID()
		if id == "" {
			t.Error("Expected non-empty ID")
		}
	})

	t.Run("NewAnimatedComponentWithID() sets custom ID", func(t *testing.T) {
		mock := newMockComponent("Test")
		customID := "my-custom-animation"
		ac := NewAnimatedComponentWithID(customID, mock, DefaultAnimationConfig())

		id := ac.GetID()
		if id != customID {
			t.Errorf("Expected ID '%s', got '%s'", customID, id)
		}
	})

	t.Run("Each animation has unique ID", func(t *testing.T) {
		mock1 := newMockComponent("Test1")
		mock2 := newMockComponent("Test2")

		ac1 := NewAnimatedComponent(mock1, DefaultAnimationConfig())
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
		ac2 := NewAnimatedComponent(mock2, DefaultAnimationConfig())

		id1 := ac1.GetID()
		id2 := ac2.GetID()

		if id1 == id2 {
			t.Error("Expected unique IDs for different animations")
		}
	})
}

// TestAnimatedComponentReset tests Reset() method
func TestAnimatedComponentReset(t *testing.T) {
	t.Run("Reset() clears animation state", func(t *testing.T) {
		mock := newMockComponent("Test")
		ac := NewAnimatedComponent(mock, DefaultAnimationConfig())

		ac.StartAnimation(0.0, 1.0)
		ac.current = 0.5
		ac.velocity = 10.0

		ac.Reset()

		if ac.IsAnimating() {
			t.Error("Expected animation to be stopped after reset")
		}

		if ac.current != ac.from {
			t.Errorf("Expected current to be reset to from value, got %.2f", ac.current)
		}

		if ac.velocity != 0 {
			t.Errorf("Expected velocity to be 0, got %.2f", ac.velocity)
		}
	})

	t.Run("Reset() clears reversed flag", func(t *testing.T) {
		mock := newMockComponent("Test")
		config := DefaultAnimationConfig()
		config.Loop = true
		config.Reverse = true

		ac := NewAnimatedComponent(mock, config)
		ac.reversed = true

		ac.Reset()

		if ac.reversed {
			t.Error("Expected reversed flag to be cleared")
		}
	})
}

// TestAnimatedComponentUpdate tests Update() message handling
func TestAnimatedComponentUpdate(t *testing.T) {
	t.Run("Update() passes message to wrapped component", func(t *testing.T) {
		mock := &mockComponentWithUpdate{
			mockComponent: newMockComponent("Test"),
			updateCount:   0,
		}
		ac := NewAnimatedComponent(mock, DefaultAnimationConfig())

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		ac.Update(msg)

		if mock.updateCount != 1 {
			t.Errorf("Expected update to be called once, got %d", mock.updateCount)
		}
	})

	t.Run("Update() ignores ticks for different animation IDs", func(t *testing.T) {
		mock := newMockComponent("Test")
		ac := NewAnimatedComponent(mock, DefaultAnimationConfig())
		ac.StartAnimation(0.0, 1.0)

		initialValue := ac.GetValue()

		// Send tick with different ID
		msg := AnimationTickMsg{
			ID:    "different-id",
			Value: 0.5,
			Time:  time.Now(),
		}
		ac.Update(msg)

		// Value should not change
		if ac.GetValue() != initialValue {
			t.Error("Animation should ignore ticks with different IDs")
		}
	})
}

// mockComponentWithUpdate extends mockComponent to track Update calls
type mockComponentWithUpdate struct {
	*mockComponent
	updateCount int
}

func (m *mockComponentWithUpdate) Update(msg tea.Msg) (Component, tea.Cmd) {
	m.updateCount++
	return m, nil
}
