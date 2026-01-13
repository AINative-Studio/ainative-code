package components_test

import (
	"testing"

	"github.com/AINative-studio/ainative-code/internal/tui/components"
	tea "github.com/charmbracelet/bubbletea"
)

// TestComponentAdapterImplementsInterfaces verifies ComponentAdapter implements required interfaces
func TestComponentAdapterImplementsInterfaces(t *testing.T) {
	adapter := components.NewComponentAdapter()

	// Test Component interface
	t.Run("Component interface", func(t *testing.T) {
		var _ components.Component = adapter

		cmd := adapter.Init()
		if cmd != nil {
			t.Error("ComponentAdapter.Init() should return nil")
		}

		_, cmd = adapter.Update(nil)
		if cmd != nil {
			t.Error("ComponentAdapter.Update() should return nil")
		}

		view := adapter.View()
		if view != "" {
			t.Error("ComponentAdapter.View() should return empty string")
		}
	})

	// Test Sizeable interface
	t.Run("Sizeable interface", func(t *testing.T) {
		var _ components.Sizeable = adapter

		adapter.SetSize(100, 50)
		width, height := adapter.GetSize()
		if width != 100 || height != 50 {
			t.Errorf("Expected size (100, 50), got (%d, %d)", width, height)
		}
	})

	// Test Focusable interface
	t.Run("Focusable interface", func(t *testing.T) {
		var _ components.Focusable = adapter

		if adapter.Focused() {
			t.Error("ComponentAdapter should not be focused initially")
		}

		adapter.Focus()
		if !adapter.Focused() {
			t.Error("ComponentAdapter should be focused after Focus()")
		}

		adapter.Blur()
		if adapter.Focused() {
			t.Error("ComponentAdapter should not be focused after Blur()")
		}
	})

	// Test Stateful interface
	t.Run("Stateful interface", func(t *testing.T) {
		var _ components.Stateful = adapter

		if !adapter.IsVisible() {
			t.Error("ComponentAdapter should be visible initially")
		}

		adapter.Hide()
		if adapter.IsVisible() {
			t.Error("ComponentAdapter should not be visible after Hide()")
		}

		adapter.Show()
		if !adapter.IsVisible() {
			t.Error("ComponentAdapter should be visible after Show()")
		}

		adapter.Toggle()
		if adapter.IsVisible() {
			t.Error("ComponentAdapter should not be visible after Toggle()")
		}

		adapter.Toggle()
		if !adapter.IsVisible() {
			t.Error("ComponentAdapter should be visible after second Toggle()")
		}
	})

	// Test Lifecycle interface
	t.Run("Lifecycle interface", func(t *testing.T) {
		// Create fresh adapter for lifecycle testing
		lifecycleAdapter := components.NewComponentAdapter()
		var _ components.Lifecycle = lifecycleAdapter

		// ComponentAdapter starts uninitialized
		if lifecycleAdapter.GetLifecycleState() != components.StateUninitialized {
			t.Errorf("Expected initial state Uninitialized, got %v", lifecycleAdapter.GetLifecycleState())
		}

		lifecycleAdapter.Init()
		if !lifecycleAdapter.IsInitialized() {
			t.Error("ComponentAdapter should be initialized after Init()")
		}

		if lifecycleAdapter.IsMounted() {
			t.Error("ComponentAdapter should not be mounted initially")
		}

		lifecycleAdapter.OnMount()
		if !lifecycleAdapter.IsMounted() {
			t.Error("ComponentAdapter should be mounted after OnMount()")
		}

		state := lifecycleAdapter.GetLifecycleState()
		if state != components.StateMounted {
			t.Errorf("Expected lifecycle state StateMounted, got %v", state)
		}

		lifecycleAdapter.OnUnmount()
		if lifecycleAdapter.IsMounted() {
			t.Error("ComponentAdapter should not be mounted after OnUnmount()")
		}

		state = lifecycleAdapter.GetLifecycleState()
		if state != components.StateUnmounted {
			t.Errorf("Expected lifecycle state StateUnmounted, got %v", state)
		}
	})
}

// TestPopupAdapterImplementsInterfaces verifies PopupAdapter implements required interfaces
func TestPopupAdapterImplementsInterfaces(t *testing.T) {
	popup := components.NewPopupAdapter()

	// Test PopupComponent interface
	t.Run("PopupComponent interface", func(t *testing.T) {
		var _ components.PopupComponent = popup

		// Test position
		popup.SetPopupPosition(10, 20)
		x, y, _, _ := popup.GetPopupPosition()
		if x != 10 || y != 20 {
			t.Errorf("Expected position (10, 20), got (%d, %d)", x, y)
		}

		// Test dimensions
		popup.SetPopupDimensions(80, 40)
		width, height := popup.GetPopupDimensions()
		if width != 80 || height != 40 {
			t.Errorf("Expected dimensions (80, 40), got (%d, %d)", width, height)
		}

		// Test modal
		if popup.IsModal() {
			t.Error("Popup should not be modal by default")
		}
		popup.SetModal(true)
		if !popup.IsModal() {
			t.Error("Popup should be modal after SetModal(true)")
		}

		// Test z-index
		defaultZIndex := popup.GetZIndex()
		if defaultZIndex != 100 {
			t.Errorf("Expected default z-index 100, got %d", defaultZIndex)
		}
		popup.SetZIndex(200)
		if popup.GetZIndex() != 200 {
			t.Errorf("Expected z-index 200, got %d", popup.GetZIndex())
		}
	})

	// Test config
	t.Run("PopupConfig", func(t *testing.T) {
		config := components.DefaultPopupConfig()

		if config.Width != 60 {
			t.Errorf("Expected default width 60, got %d", config.Width)
		}
		if config.Height != 20 {
			t.Errorf("Expected default height 20, got %d", config.Height)
		}
		if config.ZIndex != 100 {
			t.Errorf("Expected default z-index 100, got %d", config.ZIndex)
		}
		if !config.CloseOnEscape {
			t.Error("Expected default CloseOnEscape to be true")
		}
		if config.Modal {
			t.Error("Expected default Modal to be false")
		}

		popup.SetConfig(config)
		retrievedConfig := popup.GetConfig()
		if retrievedConfig.Width != config.Width {
			t.Errorf("Expected config width %d, got %d", config.Width, retrievedConfig.Width)
		}
	})

	// Test close callback
	t.Run("Close callback", func(t *testing.T) {
		callbackCalled := false
		popup.OnClose(func() {
			callbackCalled = true
		})

		popup.Close()
		if !callbackCalled {
			t.Error("OnClose callback should be called when popup closes")
		}
		if popup.IsVisible() {
			t.Error("Popup should be hidden after Close()")
		}
	})
}

// TestLifecycleStateString tests the String() method of LifecycleState
func TestLifecycleStateString(t *testing.T) {
	tests := []struct {
		state    components.LifecycleState
		expected string
	}{
		{components.StateUninitialized, "uninitialized"},
		{components.StateInitializing, "initializing"},
		{components.StateInitialized, "initialized"},
		{components.StateMounting, "mounting"},
		{components.StateMounted, "mounted"},
		{components.StateUpdating, "updating"},
		{components.StateUnmounting, "unmounting"},
		{components.StateUnmounted, "unmounted"},
		{components.StateError, "error"},
		{components.StateDisposed, "disposed"},
		{components.LifecycleState(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestLifecycleEventTypeString tests the String() method of LifecycleEventType
func TestLifecycleEventTypeString(t *testing.T) {
	tests := []struct {
		eventType components.LifecycleEventType
		expected  string
	}{
		{components.EventInit, "init"},
		{components.EventMount, "mount"},
		{components.EventUnmount, "unmount"},
		{components.EventUpdate, "update"},
		{components.EventShow, "show"},
		{components.EventHide, "hide"},
		{components.EventResize, "resize"},
		{components.EventFocus, "focus"},
		{components.EventBlur, "blur"},
		{components.EventError, "error"},
		{components.EventDispose, "dispose"},
		{components.LifecycleEventType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.eventType.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestLifecycleHooksExecution tests LifecycleHooks execution
func TestLifecycleHooksExecution(t *testing.T) {
	hooks := &components.LifecycleHooks{}

	// Test nil function calls
	t.Run("Nil functions", func(t *testing.T) {
		if cmd := hooks.ExecuteInit(); cmd != nil {
			t.Error("ExecuteInit should return nil when OnInitFunc is nil")
		}
		if cmd := hooks.ExecuteMount(); cmd != nil {
			t.Error("ExecuteMount should return nil when OnMountFunc is nil")
		}
		if cmd := hooks.ExecuteUnmount(); cmd != nil {
			t.Error("ExecuteUnmount should return nil when OnUnmountFunc is nil")
		}
		if !hooks.ExecuteBeforeUpdate(nil) {
			t.Error("ExecuteBeforeUpdate should return true when OnBeforeUpdateFunc is nil")
		}
		if cmd := hooks.ExecuteAfterUpdate(nil); cmd != nil {
			t.Error("ExecuteAfterUpdate should return nil when OnAfterUpdateFunc is nil")
		}
		if cmd := hooks.ExecuteShow(); cmd != nil {
			t.Error("ExecuteShow should return nil when OnShowFunc is nil")
		}
		if cmd := hooks.ExecuteHide(); cmd != nil {
			t.Error("ExecuteHide should return nil when OnHideFunc is nil")
		}
		if cmd := hooks.ExecuteResize(100, 50); cmd != nil {
			t.Error("ExecuteResize should return nil when OnResizeFunc is nil")
		}
		if cmd := hooks.ExecuteFocus(); cmd != nil {
			t.Error("ExecuteFocus should return nil when OnFocusFunc is nil")
		}
		if cmd := hooks.ExecuteBlur(); cmd != nil {
			t.Error("ExecuteBlur should return nil when OnBlurFunc is nil")
		}
	})

	// Test function calls
	t.Run("Function execution", func(t *testing.T) {
		initCalled := false
		hooks.OnInitFunc = func() tea.Cmd {
			initCalled = true
			return nil
		}
		hooks.ExecuteInit()
		if !initCalled {
			t.Error("OnInitFunc should be called")
		}

		mountCalled := false
		hooks.OnMountFunc = func() tea.Cmd {
			mountCalled = true
			return nil
		}
		hooks.ExecuteMount()
		if !mountCalled {
			t.Error("OnMountFunc should be called")
		}

		unmountCalled := false
		hooks.OnUnmountFunc = func() tea.Cmd {
			unmountCalled = true
			return nil
		}
		hooks.ExecuteUnmount()
		if !unmountCalled {
			t.Error("OnUnmountFunc should be called")
		}

		beforeUpdateCalled := false
		hooks.OnBeforeUpdateFunc = func(msg tea.Msg) bool {
			beforeUpdateCalled = true
			return false
		}
		result := hooks.ExecuteBeforeUpdate(nil)
		if !beforeUpdateCalled || result {
			t.Error("OnBeforeUpdateFunc should be called and return false")
		}

		afterUpdateCalled := false
		hooks.OnAfterUpdateFunc = func(msg tea.Msg) tea.Cmd {
			afterUpdateCalled = true
			return nil
		}
		hooks.ExecuteAfterUpdate(nil)
		if !afterUpdateCalled {
			t.Error("OnAfterUpdateFunc should be called")
		}

		showCalled := false
		hooks.OnShowFunc = func() tea.Cmd {
			showCalled = true
			return nil
		}
		hooks.ExecuteShow()
		if !showCalled {
			t.Error("OnShowFunc should be called")
		}

		hideCalled := false
		hooks.OnHideFunc = func() tea.Cmd {
			hideCalled = true
			return nil
		}
		hooks.ExecuteHide()
		if !hideCalled {
			t.Error("OnHideFunc should be called")
		}

		resizeCalled := false
		var resizeWidth, resizeHeight int
		hooks.OnResizeFunc = func(width, height int) tea.Cmd {
			resizeCalled = true
			resizeWidth = width
			resizeHeight = height
			return nil
		}
		hooks.ExecuteResize(100, 50)
		if !resizeCalled || resizeWidth != 100 || resizeHeight != 50 {
			t.Error("OnResizeFunc should be called with correct parameters")
		}

		focusCalled := false
		hooks.OnFocusFunc = func() tea.Cmd {
			focusCalled = true
			return nil
		}
		hooks.ExecuteFocus()
		if !focusCalled {
			t.Error("OnFocusFunc should be called")
		}

		blurCalled := false
		hooks.OnBlurFunc = func() tea.Cmd {
			blurCalled = true
			return nil
		}
		hooks.ExecuteBlur()
		if !blurCalled {
			t.Error("OnBlurFunc should be called")
		}
	})
}

// TestPopupAlignmentConstants tests that popup alignment constants are defined correctly
func TestPopupAlignmentConstants(t *testing.T) {
	alignments := []components.PopupAlignment{
		components.AlignTopLeft,
		components.AlignTopCenter,
		components.AlignTopRight,
		components.AlignCenterLeft,
		components.AlignCenter,
		components.AlignCenterRight,
		components.AlignBottomLeft,
		components.AlignBottomCenter,
		components.AlignBottomRight,
	}

	// Ensure all alignments are unique
	seen := make(map[components.PopupAlignment]bool)
	for _, alignment := range alignments {
		if seen[alignment] {
			t.Errorf("Duplicate alignment value: %v", alignment)
		}
		seen[alignment] = true
	}

	// Ensure we have 9 distinct alignments
	if len(seen) != 9 {
		t.Errorf("Expected 9 unique alignments, got %d", len(seen))
	}
}

// TestComponentAdapterResize tests the OnResize functionality
func TestComponentAdapterResize(t *testing.T) {
	adapter := components.NewComponentAdapter()

	// Initial size should be 0, 0
	width, height := adapter.GetSize()
	if width != 0 || height != 0 {
		t.Errorf("Expected initial size (0, 0), got (%d, %d)", width, height)
	}

	// Test OnResize
	adapter.OnResize(200, 100)
	width, height = adapter.GetSize()
	if width != 200 || height != 100 {
		t.Errorf("Expected size after resize (200, 100), got (%d, %d)", width, height)
	}
}

// TestPopupComponentStateTransitions tests popup visibility state transitions
func TestPopupComponentStateTransitions(t *testing.T) {
	popup := components.NewPopupAdapter()

	// Initial state: visible
	if !popup.IsVisible() {
		t.Error("Popup should be visible initially")
	}

	// Hide the popup
	popup.Hide()
	if popup.IsVisible() {
		t.Error("Popup should be hidden after Hide()")
	}

	// Show the popup
	popup.Show()
	if !popup.IsVisible() {
		t.Error("Popup should be visible after Show()")
	}

	// Toggle the popup
	popup.Toggle()
	if popup.IsVisible() {
		t.Error("Popup should be hidden after first Toggle()")
	}

	popup.Toggle()
	if !popup.IsVisible() {
		t.Error("Popup should be visible after second Toggle()")
	}
}
