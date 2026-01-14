package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/AINative-studio/ainative-code/internal/tui/layout"
)

// TestDraggableInit tests Init method
func TestDraggableInit(t *testing.T) {
	t.Run("Init passes through to wrapped component", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 10, 10)

		cmd := draggable.Init()
		// Mock returns nil
		if cmd != nil {
			t.Error("Expected nil cmd from Init()")
		}
	})
}

// TestDraggableView tests View rendering
func TestDraggableView(t *testing.T) {
	t.Run("View renders wrapped component", func(t *testing.T) {
		mock := newMockComponent("Draggable Content")
		draggable := NewDraggable(mock, 10, 10)

		view := draggable.View()
		// View should contain the wrapped component's content
		if view == "" {
			t.Error("Expected non-empty view")
		}
	})

	t.Run("View when dragging shows drag style", func(t *testing.T) {
		mock := newMockComponent("Dragging")
		draggable := NewDraggable(mock, 10, 10)

		draggable.StartDrag(15, 15)

		view := draggable.View()
		if view == "" {
			t.Error("Expected non-empty view while dragging")
		}
	})

	t.Run("View when focused", func(t *testing.T) {
		mock := newMockComponent("Focused")
		draggable := NewDraggable(mock, 10, 10)

		draggable.Focus()

		view := draggable.View()
		if view == "" {
			t.Error("Expected non-empty view when focused")
		}
	})
}

// TestDraggableComponentAccessors tests component getter/setter
func TestDraggableComponentAccessors(t *testing.T) {
	t.Run("GetComponent returns wrapped component", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 10, 10)

		component := draggable.GetComponent()
		if component != mock {
			t.Error("Expected GetComponent to return wrapped component")
		}
	})

	t.Run("SetComponent updates wrapped component", func(t *testing.T) {
		mock1 := newMockComponent("Test1")
		mock2 := newMockComponent("Test2")
		draggable := NewDraggable(mock1, 10, 10)

		draggable.SetComponent(mock2)

		component := draggable.GetComponent()
		if component != mock2 {
			t.Error("Expected component to be updated")
		}
	})
}

// TestDraggableGetSize tests GetSize method
func TestDraggableGetSize(t *testing.T) {
	t.Run("GetSize returns dimensions", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 10, 10)
		draggable.SetSize(50, 30)

		w, h := draggable.GetSize()
		if w != 50 || h != 30 {
			t.Errorf("Expected size (50, 30), got (%d, %d)", w, h)
		}
	})
}

// TestResizableInit tests Init method
func TestResizableInit(t *testing.T) {
	t.Run("Init passes through to wrapped component", func(t *testing.T) {
		mock := newMockComponent("Test")
		resizable := NewResizable(mock, 40, 20)

		cmd := resizable.Init()
		if cmd != nil {
			t.Error("Expected nil cmd from Init()")
		}
	})
}

// TestResizableView tests View rendering
func TestResizableView(t *testing.T) {
	t.Run("View renders wrapped component", func(t *testing.T) {
		mock := newMockComponent("Resizable Content")
		resizable := NewResizable(mock, 40, 20)

		view := resizable.View()
		if view == "" {
			t.Error("Expected non-empty view")
		}
	})

	t.Run("View when resizing", func(t *testing.T) {
		mock := newMockComponent("Resizing")
		resizable := NewResizable(mock, 40, 20)

		resizable.StartResize(45, 25)

		view := resizable.View()
		if view == "" {
			t.Error("Expected non-empty view while resizing")
		}
	})

	t.Run("View when focused", func(t *testing.T) {
		mock := newMockComponent("Focused")
		resizable := NewResizable(mock, 40, 20)

		resizable.Focus()

		view := resizable.View()
		if view == "" {
			t.Error("Expected non-empty view when focused")
		}
	})
}

// TestResizableComponentAccessors tests component getter/setter
func TestResizableComponentAccessors(t *testing.T) {
	t.Run("GetComponent returns wrapped component", func(t *testing.T) {
		mock := newMockComponent("Test")
		resizable := NewResizable(mock, 40, 20)

		component := resizable.GetComponent()
		if component != mock {
			t.Error("Expected GetComponent to return wrapped component")
		}
	})

	t.Run("SetComponent updates wrapped component", func(t *testing.T) {
		mock1 := newMockComponent("Test1")
		mock2 := newMockComponent("Test2")
		resizable := NewResizable(mock1, 40, 20)

		resizable.SetComponent(mock2)

		component := resizable.GetComponent()
		if component != mock2 {
			t.Error("Expected component to be updated")
		}
	})
}

// TestSplitViewInit tests Init method
func TestSplitViewInit(t *testing.T) {
	t.Run("Init initializes both panes", func(t *testing.T) {
		left := newMockComponent("Left")
		right := newMockComponent("Right")
		split := NewSplitView(left, right, layout.Horizontal)

		cmd := split.Init()
		// Both mock components return nil
		if cmd != nil {
			t.Error("Expected combined cmd")
		}
	})
}

// TestSplitViewView tests View rendering
func TestSplitViewView(t *testing.T) {
	t.Run("View renders both panes", func(t *testing.T) {
		left := newMockComponent("Left")
		right := newMockComponent("Right")
		split := NewSplitView(left, right, layout.Horizontal)
		split.SetSize(100, 50)

		view := split.View()
		if view == "" {
			t.Error("Expected non-empty view")
		}
	})

	t.Run("View with divider", func(t *testing.T) {
		left := newMockComponent("Left")
		right := newMockComponent("Right")
		split := NewSplitView(left, right, layout.Horizontal)
		split.EnableDivider()
		split.SetSize(100, 50)

		view := split.View()
		if view == "" {
			t.Error("Expected non-empty view with divider")
		}
	})

	t.Run("View without divider", func(t *testing.T) {
		left := newMockComponent("Left")
		right := newMockComponent("Right")
		split := NewSplitView(left, right, layout.Horizontal)
		split.DisableDivider()
		split.SetSize(100, 50)

		view := split.View()
		if view == "" {
			t.Error("Expected non-empty view without divider")
		}
	})
}

// TestMultiColumnInit tests Init method
func TestMultiColumnInit(t *testing.T) {
	t.Run("Init initializes all columns", func(t *testing.T) {
		col1 := newMockComponent("Column 1")
		col2 := newMockComponent("Column 2")
		mcl := NewMultiColumnLayout([]Component{col1, col2})

		cmd := mcl.Init()
		// All mock components return nil
		if cmd != nil {
			t.Error("Expected combined cmd")
		}
	})
}

// TestMultiColumnView tests View rendering
func TestMultiColumnView(t *testing.T) {
	t.Run("View renders all columns", func(t *testing.T) {
		col1 := newMockComponent("Column 1")
		col2 := newMockComponent("Column 2")
		mcl := NewMultiColumnLayout([]Component{col1, col2})
		mcl.SetSize(100, 50)

		view := mcl.View()
		if view == "" {
			t.Error("Expected non-empty view")
		}
	})

	t.Run("View with gaps", func(t *testing.T) {
		col1 := newMockComponent("Column 1")
		col2 := newMockComponent("Column 2")
		mcl := NewMultiColumnLayout([]Component{col1, col2})
		mcl.SetGap(3)
		mcl.SetSize(100, 50)

		view := mcl.View()
		if view == "" {
			t.Error("Expected non-empty view with gaps")
		}
	})

	t.Run("View with different alignments", func(t *testing.T) {
		col1 := newMockComponent("Column 1")
		col2 := newMockComponent("Column 2")

		alignments := []ColumnAlignment{
			ColumnAlignTop,
			ColumnAlignCenter,
			ColumnAlignBottom,
			ColumnAlignStretch,
		}

		for _, align := range alignments {
			mcl := NewMultiColumnLayout([]Component{col1, col2})
			mcl.SetAlignment(align)
			mcl.SetSize(100, 50)

			view := mcl.View()
			if view == "" {
				t.Errorf("Expected non-empty view with alignment %s", align.String())
			}
		}
	})
}

// TestMouseHelpers tests mouse utility functions
func TestMouseHelpers(t *testing.T) {
	t.Run("IsScrollUp", func(t *testing.T) {
		msg := tea.MouseMsg{
			Button: tea.MouseButtonWheelUp,
			Action: tea.MouseActionPress,
		}
		if !IsScrollUp(msg) {
			t.Error("Expected IsScrollUp to return true")
		}
	})

	t.Run("IsScrollDown", func(t *testing.T) {
		msg := tea.MouseMsg{
			Button: tea.MouseButtonWheelDown,
			Action: tea.MouseActionPress,
		}
		if !IsScrollDown(msg) {
			t.Error("Expected IsScrollDown to return true")
		}
	})
}

// TestTransitionPresetAccess tests transition preset access
func TestTransitionPresetAccess(t *testing.T) {
	t.Run("GetTransitionPreset for valid preset", func(t *testing.T) {
		config, exists := GetTransitionPreset("fadeIn")
		if !exists {
			t.Error("Expected fadeIn preset to exist")
		}
		if config.Duration != FadeIn.Duration {
			t.Error("Expected FadeIn config")
		}
	})

	t.Run("GetTransitionPreset for invalid preset", func(t *testing.T) {
		_, exists := GetTransitionPreset("nonexistent")
		if exists {
			t.Error("Expected preset to not exist")
		}
	})

	t.Run("All transition presets exist", func(t *testing.T) {
		presets := []string{
			"fadeIn", "fadeOut", "slideIn", "slideOut",
			"spring", "bounce", "smooth", "spinner", "pulse", "snap",
		}

		for _, name := range presets {
			_, exists := GetTransitionPreset(name)
			if !exists {
				t.Errorf("Expected preset '%s' to exist", name)
			}
		}
	})

	t.Run("QuickTransition returns Snap", func(t *testing.T) {
		config := QuickTransition()
		if config.Duration != Snap.Duration {
			t.Error("Expected QuickTransition to return Snap config")
		}
	})

	t.Run("MediumTransition returns Smooth", func(t *testing.T) {
		config := MediumTransition()
		if config.Duration != Smooth.Duration {
			t.Error("Expected MediumTransition to return Smooth config")
		}
	})

	t.Run("SlowTransition returns Bounce", func(t *testing.T) {
		config := SlowTransition()
		if config.Duration != Bounce.Duration {
			t.Error("Expected SlowTransition to return Bounce config")
		}
	})
}

// TestPopupAdapter tests PopupAdapter
func TestPopupAdapter(t *testing.T) {
	t.Run("NewPopupAdapter creates with defaults", func(t *testing.T) {
		popup := NewPopupAdapter()

		if popup == nil {
			t.Fatal("Expected non-nil popup adapter")
		}

		if popup.GetZIndex() != 100 {
			t.Errorf("Expected default z-index 100, got %d", popup.GetZIndex())
		}

		if popup.IsModal() {
			t.Error("Expected popup to not be modal by default")
		}
	})

	t.Run("SetPopupDimensions", func(t *testing.T) {
		popup := NewPopupAdapter()
		popup.SetPopupDimensions(200, 100)

		w, h := popup.GetPopupDimensions()
		if w != 200 || h != 100 {
			t.Errorf("Expected dimensions (200, 100), got (%d, %d)", w, h)
		}
	})

	t.Run("SetModal", func(t *testing.T) {
		popup := NewPopupAdapter()
		popup.SetModal(true)

		if !popup.IsModal() {
			t.Error("Expected popup to be modal")
		}
	})

	t.Run("SetZIndex", func(t *testing.T) {
		popup := NewPopupAdapter()
		popup.SetZIndex(200)

		if popup.GetZIndex() != 200 {
			t.Errorf("Expected z-index 200, got %d", popup.GetZIndex())
		}
	})

	t.Run("Close hides popup", func(t *testing.T) {
		popup := NewPopupAdapter()
		popup.Show()

		popup.Close()

		if popup.IsVisible() {
			t.Error("Expected popup to be hidden after Close()")
		}
	})

	t.Run("OnClose callback", func(t *testing.T) {
		popup := NewPopupAdapter()
		called := false

		popup.OnClose(func() {
			called = true
		})

		popup.Close()

		if !called {
			t.Error("Expected OnClose callback to be called")
		}
	})
}
