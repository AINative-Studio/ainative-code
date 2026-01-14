package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestComponentAdapterLifecycle tests lifecycle hook execution
func TestComponentAdapterLifecycle(t *testing.T) {
	t.Run("OnInit() hook", func(t *testing.T) {
		adapter := NewComponentAdapter()

		if adapter.GetLifecycleState() != StateUninitialized {
			t.Errorf("Expected StateUninitialized, got %v", adapter.GetLifecycleState())
		}

		cmd := adapter.OnInit()
		if cmd != nil {
			t.Error("Expected nil cmd from OnInit()")
		}

		if adapter.GetLifecycleState() != StateInitializing {
			t.Errorf("Expected StateInitializing after OnInit(), got %v", adapter.GetLifecycleState())
		}
	})

	t.Run("OnMount() hook", func(t *testing.T) {
		adapter := NewComponentAdapter()

		if adapter.IsMounted() {
			t.Error("Expected component to not be mounted initially")
		}

		cmd := adapter.OnMount()
		if cmd != nil {
			t.Error("Expected nil cmd from OnMount()")
		}

		if !adapter.IsMounted() {
			t.Error("Expected component to be mounted after OnMount()")
		}

		if adapter.GetLifecycleState() != StateMounted {
			t.Errorf("Expected StateMounted after OnMount(), got %v", adapter.GetLifecycleState())
		}
	})

	t.Run("OnUnmount() hook", func(t *testing.T) {
		adapter := NewComponentAdapter()
		adapter.OnMount()

		if !adapter.IsMounted() {
			t.Error("Expected component to be mounted before unmount")
		}

		cmd := adapter.OnUnmount()
		if cmd != nil {
			t.Error("Expected nil cmd from OnUnmount()")
		}

		if adapter.IsMounted() {
			t.Error("Expected component to be unmounted after OnUnmount()")
		}

		if adapter.GetLifecycleState() != StateUnmounted {
			t.Errorf("Expected StateUnmounted after OnUnmount(), got %v", adapter.GetLifecycleState())
		}
	})

	t.Run("OnBeforeUpdate() hook", func(t *testing.T) {
		adapter := NewComponentAdapter()

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		shouldContinue := adapter.OnBeforeUpdate(msg)

		if !shouldContinue {
			t.Error("Expected OnBeforeUpdate to return true by default")
		}
	})

	t.Run("OnAfterUpdate() hook", func(t *testing.T) {
		adapter := NewComponentAdapter()

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		cmd := adapter.OnAfterUpdate(msg)

		if cmd != nil {
			t.Error("Expected nil cmd from OnAfterUpdate()")
		}
	})

	t.Run("OnShow() hook", func(t *testing.T) {
		adapter := NewComponentAdapter()

		cmd := adapter.OnShow()
		if cmd != nil {
			t.Error("Expected nil cmd from OnShow()")
		}
	})

	t.Run("OnHide() hook", func(t *testing.T) {
		adapter := NewComponentAdapter()

		cmd := adapter.OnHide()
		if cmd != nil {
			t.Error("Expected nil cmd from OnHide()")
		}
	})

	t.Run("OnResize() hook", func(t *testing.T) {
		adapter := NewComponentAdapter()

		initialW, initialH := adapter.GetSize()

		cmd := adapter.OnResize(100, 50)
		if cmd != nil {
			t.Error("Expected nil cmd from OnResize()")
		}

		w, h := adapter.GetSize()
		if w != 100 || h != 50 {
			t.Errorf("Expected size (100, 50), got (%d, %d)", w, h)
		}

		if w == initialW && h == initialH {
			t.Error("OnResize should have changed the size")
		}
	})

	t.Run("OnFocus() hook", func(t *testing.T) {
		adapter := NewComponentAdapter()

		cmd := adapter.OnFocus()
		if cmd != nil {
			t.Error("Expected nil cmd from OnFocus()")
		}
	})

	t.Run("OnBlur() hook", func(t *testing.T) {
		adapter := NewComponentAdapter()

		cmd := adapter.OnBlur()
		if cmd != nil {
			t.Error("Expected nil cmd from OnBlur()")
		}
	})
}

// TestLifecycleStateTransitions tests state transitions
func TestLifecycleStateTransitions(t *testing.T) {
	t.Run("Uninitialized → Initializing", func(t *testing.T) {
		adapter := NewComponentAdapter()

		if adapter.GetLifecycleState() != StateUninitialized {
			t.Fatal("Expected StateUninitialized initially")
		}

		adapter.OnInit()

		if adapter.GetLifecycleState() != StateInitializing {
			t.Errorf("Expected StateInitializing, got %v", adapter.GetLifecycleState())
		}
	})

	t.Run("Initializing → Initialized", func(t *testing.T) {
		adapter := NewComponentAdapter()
		adapter.OnInit()

		adapter.Init()

		if !adapter.IsInitialized() {
			t.Error("Expected component to be initialized")
		}

		if adapter.GetLifecycleState() != StateInitialized {
			t.Errorf("Expected StateInitialized, got %v", adapter.GetLifecycleState())
		}
	})

	t.Run("Initialized → Mounted", func(t *testing.T) {
		adapter := NewComponentAdapter()
		adapter.OnInit()
		adapter.Init()

		adapter.OnMount()

		if adapter.GetLifecycleState() != StateMounted {
			t.Errorf("Expected StateMounted, got %v", adapter.GetLifecycleState())
		}
	})

	t.Run("Mounted → Unmounted", func(t *testing.T) {
		adapter := NewComponentAdapter()
		adapter.OnInit()
		adapter.Init()
		adapter.OnMount()

		if adapter.GetLifecycleState() != StateMounted {
			t.Fatal("Expected StateMounted before unmount")
		}

		adapter.OnUnmount()

		if adapter.GetLifecycleState() != StateUnmounted {
			t.Errorf("Expected StateUnmounted, got %v", adapter.GetLifecycleState())
		}
	})

	t.Run("Full lifecycle sequence", func(t *testing.T) {
		adapter := NewComponentAdapter()

		states := []LifecycleState{
			StateUninitialized,
			StateInitializing,
			StateInitialized,
			StateMounted,
			StateUnmounted,
		}

		stateNames := []string{
			"uninitialized",
			"initializing",
			"initialized",
			"mounted",
			"unmounted",
		}

		// Check initial state
		if adapter.GetLifecycleState() != states[0] {
			t.Errorf("Expected %v, got %v", states[0], adapter.GetLifecycleState())
		}

		// OnInit
		adapter.OnInit()
		if adapter.GetLifecycleState() != states[1] {
			t.Errorf("Expected %v, got %v", states[1], adapter.GetLifecycleState())
		}

		// Init
		adapter.Init()
		if adapter.GetLifecycleState() != states[2] {
			t.Errorf("Expected %v, got %v", states[2], adapter.GetLifecycleState())
		}

		// OnMount
		adapter.OnMount()
		if adapter.GetLifecycleState() != states[3] {
			t.Errorf("Expected %v, got %v", states[3], adapter.GetLifecycleState())
		}

		// OnUnmount
		adapter.OnUnmount()
		if adapter.GetLifecycleState() != states[4] {
			t.Errorf("Expected %v, got %v", states[4], adapter.GetLifecycleState())
		}

		// Verify state names
		for i, expected := range stateNames {
			if states[i].String() != expected {
				t.Errorf("Expected state name '%s', got '%s'", expected, states[i].String())
			}
		}
	})
}

// TestLifecycleHooks tests LifecycleHooks utility
func TestLifecycleHooks(t *testing.T) {
	t.Run("ExecuteInit with defined hook", func(t *testing.T) {
		called := false
		hooks := &LifecycleHooks{
			OnInitFunc: func() tea.Cmd {
				called = true
				return nil
			},
		}

		cmd := hooks.ExecuteInit()
		if !called {
			t.Error("Expected OnInitFunc to be called")
		}
		if cmd != nil {
			t.Error("Expected nil cmd")
		}
	})

	t.Run("ExecuteInit with nil hook", func(t *testing.T) {
		hooks := &LifecycleHooks{}

		cmd := hooks.ExecuteInit()
		if cmd != nil {
			t.Error("Expected nil cmd when hook is not defined")
		}
	})

	t.Run("ExecuteMount with defined hook", func(t *testing.T) {
		called := false
		hooks := &LifecycleHooks{
			OnMountFunc: func() tea.Cmd {
				called = true
				return nil
			},
		}

		cmd := hooks.ExecuteMount()
		if !called {
			t.Error("Expected OnMountFunc to be called")
		}
		if cmd != nil {
			t.Error("Expected nil cmd")
		}
	})

	t.Run("ExecuteUnmount with defined hook", func(t *testing.T) {
		called := false
		hooks := &LifecycleHooks{
			OnUnmountFunc: func() tea.Cmd {
				called = true
				return nil
			},
		}

		cmd := hooks.ExecuteUnmount()
		if !called {
			t.Error("Expected OnUnmountFunc to be called")
		}
		if cmd != nil {
			t.Error("Expected nil cmd")
		}
	})

	t.Run("ExecuteBeforeUpdate with defined hook", func(t *testing.T) {
		called := false
		hooks := &LifecycleHooks{
			OnBeforeUpdateFunc: func(msg tea.Msg) bool {
				called = true
				return false // Cancel update
			},
		}

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		shouldContinue := hooks.ExecuteBeforeUpdate(msg)

		if !called {
			t.Error("Expected OnBeforeUpdateFunc to be called")
		}
		if shouldContinue {
			t.Error("Expected update to be cancelled")
		}
	})

	t.Run("ExecuteBeforeUpdate with nil hook returns true", func(t *testing.T) {
		hooks := &LifecycleHooks{}

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		shouldContinue := hooks.ExecuteBeforeUpdate(msg)

		if !shouldContinue {
			t.Error("Expected update to continue when hook is not defined")
		}
	})

	t.Run("ExecuteAfterUpdate with defined hook", func(t *testing.T) {
		called := false
		hooks := &LifecycleHooks{
			OnAfterUpdateFunc: func(msg tea.Msg) tea.Cmd {
				called = true
				return nil
			},
		}

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		cmd := hooks.ExecuteAfterUpdate(msg)

		if !called {
			t.Error("Expected OnAfterUpdateFunc to be called")
		}
		if cmd != nil {
			t.Error("Expected nil cmd")
		}
	})

	t.Run("ExecuteShow with defined hook", func(t *testing.T) {
		called := false
		hooks := &LifecycleHooks{
			OnShowFunc: func() tea.Cmd {
				called = true
				return nil
			},
		}

		cmd := hooks.ExecuteShow()
		if !called {
			t.Error("Expected OnShowFunc to be called")
		}
		if cmd != nil {
			t.Error("Expected nil cmd")
		}
	})

	t.Run("ExecuteHide with defined hook", func(t *testing.T) {
		called := false
		hooks := &LifecycleHooks{
			OnHideFunc: func() tea.Cmd {
				called = true
				return nil
			},
		}

		cmd := hooks.ExecuteHide()
		if !called {
			t.Error("Expected OnHideFunc to be called")
		}
		if cmd != nil {
			t.Error("Expected nil cmd")
		}
	})

	t.Run("ExecuteResize with defined hook", func(t *testing.T) {
		called := false
		var capturedW, capturedH int

		hooks := &LifecycleHooks{
			OnResizeFunc: func(width, height int) tea.Cmd {
				called = true
				capturedW = width
				capturedH = height
				return nil
			},
		}

		cmd := hooks.ExecuteResize(100, 50)
		if !called {
			t.Error("Expected OnResizeFunc to be called")
		}
		if capturedW != 100 || capturedH != 50 {
			t.Errorf("Expected resize (100, 50), got (%d, %d)", capturedW, capturedH)
		}
		if cmd != nil {
			t.Error("Expected nil cmd")
		}
	})

	t.Run("ExecuteFocus with defined hook", func(t *testing.T) {
		called := false
		hooks := &LifecycleHooks{
			OnFocusFunc: func() tea.Cmd {
				called = true
				return nil
			},
		}

		cmd := hooks.ExecuteFocus()
		if !called {
			t.Error("Expected OnFocusFunc to be called")
		}
		if cmd != nil {
			t.Error("Expected nil cmd")
		}
	})

	t.Run("ExecuteBlur with defined hook", func(t *testing.T) {
		called := false
		hooks := &LifecycleHooks{
			OnBlurFunc: func() tea.Cmd {
				called = true
				return nil
			},
		}

		cmd := hooks.ExecuteBlur()
		if !called {
			t.Error("Expected OnBlurFunc to be called")
		}
		if cmd != nil {
			t.Error("Expected nil cmd")
		}
	})
}

// TestLifecycleStateStrings tests state string representations
func TestLifecycleStateStrings(t *testing.T) {
	tests := []struct {
		state    LifecycleState
		expected string
	}{
		{StateUninitialized, "uninitialized"},
		{StateInitializing, "initializing"},
		{StateInitialized, "initialized"},
		{StateMounting, "mounting"},
		{StateMounted, "mounted"},
		{StateUpdating, "updating"},
		{StateUnmounting, "unmounting"},
		{StateUnmounted, "unmounted"},
		{StateError, "error"},
		{StateDisposed, "disposed"},
		{LifecycleState(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestLifecycleEventTypeStrings tests event type string representations
func TestLifecycleEventTypeStrings(t *testing.T) {
	tests := []struct {
		eventType LifecycleEventType
		expected  string
	}{
		{EventInit, "init"},
		{EventMount, "mount"},
		{EventUnmount, "unmount"},
		{EventUpdate, "update"},
		{EventShow, "show"},
		{EventHide, "hide"},
		{EventResize, "resize"},
		{EventFocus, "focus"},
		{EventBlur, "blur"},
		{EventError, "error"},
		{EventDispose, "dispose"},
		{LifecycleEventType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.eventType.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestComponentAdapterStateful tests Stateful interface implementation
func TestComponentAdapterStateful(t *testing.T) {
	t.Run("Show() makes component visible", func(t *testing.T) {
		adapter := NewComponentAdapter()
		adapter.Hide()

		if adapter.IsVisible() {
			t.Error("Expected component to be hidden")
		}

		adapter.Show()

		if !adapter.IsVisible() {
			t.Error("Expected component to be visible after Show()")
		}
	})

	t.Run("Hide() makes component invisible", func(t *testing.T) {
		adapter := NewComponentAdapter()

		if !adapter.IsVisible() {
			t.Error("Expected component to be visible initially")
		}

		adapter.Hide()

		if adapter.IsVisible() {
			t.Error("Expected component to be hidden after Hide()")
		}
	})

	t.Run("Toggle() switches visibility", func(t *testing.T) {
		adapter := NewComponentAdapter()

		initialState := adapter.IsVisible()
		adapter.Toggle()

		if adapter.IsVisible() == initialState {
			t.Error("Expected Toggle() to change visibility")
		}

		adapter.Toggle()

		if adapter.IsVisible() != initialState {
			t.Error("Expected Toggle() to restore original visibility")
		}
	})
}

// TestComponentAdapterFocusable tests Focusable interface implementation
func TestComponentAdapterFocusable(t *testing.T) {
	t.Run("Focus() sets focused state", func(t *testing.T) {
		adapter := NewComponentAdapter()

		if adapter.Focused() {
			t.Error("Expected component to not be focused initially")
		}

		cmd := adapter.Focus()
		if cmd != nil {
			t.Error("Expected nil cmd from Focus()")
		}

		if !adapter.Focused() {
			t.Error("Expected component to be focused after Focus()")
		}
	})

	t.Run("Blur() clears focused state", func(t *testing.T) {
		adapter := NewComponentAdapter()
		adapter.Focus()

		if !adapter.Focused() {
			t.Error("Expected component to be focused before blur")
		}

		adapter.Blur()

		if adapter.Focused() {
			t.Error("Expected component to not be focused after Blur()")
		}
	})
}

// TestComponentAdapterSizeable tests Sizeable interface implementation
func TestComponentAdapterSizeable(t *testing.T) {
	t.Run("SetSize() updates dimensions", func(t *testing.T) {
		adapter := NewComponentAdapter()

		w, h := adapter.GetSize()
		if w != 0 || h != 0 {
			t.Errorf("Expected initial size (0, 0), got (%d, %d)", w, h)
		}

		adapter.SetSize(100, 50)

		w, h = adapter.GetSize()
		if w != 100 || h != 50 {
			t.Errorf("Expected size (100, 50), got (%d, %d)", w, h)
		}
	})

	t.Run("GetSize() returns current dimensions", func(t *testing.T) {
		adapter := NewComponentAdapter()
		adapter.SetSize(75, 25)

		w, h := adapter.GetSize()
		if w != 75 || h != 25 {
			t.Errorf("Expected size (75, 25), got (%d, %d)", w, h)
		}
	})
}
