package components

import (
	"testing"

	"github.com/AINative-studio/ainative-code/internal/tui/layout"
)

// TestDraggableConstraints tests boundary constraint edge cases
func TestDraggableConstraints(t *testing.T) {
	t.Run("Drag beyond left boundary", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 10, 10)
		draggable.SetBounds(0, 0, 100, 100)
		draggable.SetSize(10, 10)

		draggable.SetPosition(-5, 10)

		x, y := draggable.Position()
		if x < 0 {
			t.Errorf("Expected x >= 0, got %d", x)
		}
		if y != 10 {
			t.Errorf("Expected y = 10, got %d", y)
		}
	})

	t.Run("Drag beyond right boundary", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 10, 10)
		draggable.SetBounds(0, 0, 100, 100)
		draggable.SetSize(10, 10)

		draggable.SetPosition(105, 10)

		x, _ := draggable.Position()
		// Max x should be bounds.Width - component.width = 100 - 10 = 90
		if x > 90 {
			t.Errorf("Expected x <= 90, got %d", x)
		}
	})

	t.Run("Drag beyond top boundary", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 10, 10)
		draggable.SetBounds(0, 0, 100, 100)
		draggable.SetSize(10, 10)

		draggable.SetPosition(10, -5)

		x, y := draggable.Position()
		if y < 0 {
			t.Errorf("Expected y >= 0, got %d", y)
		}
		if x != 10 {
			t.Errorf("Expected x = 10, got %d", x)
		}
	})

	t.Run("Drag beyond bottom boundary", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 10, 10)
		draggable.SetBounds(0, 0, 100, 100)
		draggable.SetSize(10, 10)

		draggable.SetPosition(10, 105)

		_, y := draggable.Position()
		// Max y should be bounds.Height - component.height = 100 - 10 = 90
		if y > 90 {
			t.Errorf("Expected y <= 90, got %d", y)
		}
	})

	t.Run("Drag with zero-size bounds", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 10, 10)
		draggable.SetBounds(0, 0, 0, 0)
		draggable.SetSize(10, 10)

		draggable.SetPosition(5, 5)

		x, y := draggable.Position()
		// With zero-size bounds, component should be constrained to origin
		if x > 0 || y > 0 {
			t.Errorf("Expected position (0, 0) or negative with zero bounds, got (%d, %d)", x, y)
		}
	})

	t.Run("Drag with component larger than bounds", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 0, 0)
		draggable.SetBounds(0, 0, 50, 50)
		draggable.SetSize(100, 100)

		draggable.SetPosition(10, 10)

		x, y := draggable.Position()
		// Component is larger than bounds, so max position is negative
		if x > 0 || y > 0 {
			t.Errorf("Expected non-positive position with oversized component, got (%d, %d)", x, y)
		}
	})

	t.Run("Drag without bounds set", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 10, 10)
		// Don't set bounds

		draggable.SetPosition(1000, 2000)

		x, y := draggable.Position()
		// Without bounds, position should be unconstrained
		if x != 1000 || y != 2000 {
			t.Errorf("Expected unconstrained position (1000, 2000), got (%d, %d)", x, y)
		}
	})
}

// TestResizableConstraints tests resizing constraint edge cases
func TestResizableConstraints(t *testing.T) {
	t.Run("Resize below min width", func(t *testing.T) {
		mock := newMockComponent("Test")
		resizable := NewResizable(mock, 40, 20)
		resizable.SetMinSize(20, 10)

		resizable.SetSize(5, 20)

		w, _ := resizable.Size()
		if w < 20 {
			t.Errorf("Expected width >= 20, got %d", w)
		}
	})

	t.Run("Resize below min height", func(t *testing.T) {
		mock := newMockComponent("Test")
		resizable := NewResizable(mock, 40, 20)
		resizable.SetMinSize(20, 10)

		resizable.SetSize(40, 2)

		_, h := resizable.Size()
		if h < 10 {
			t.Errorf("Expected height >= 10, got %d", h)
		}
	})

	t.Run("Resize above max width", func(t *testing.T) {
		mock := newMockComponent("Test")
		resizable := NewResizable(mock, 40, 20)
		resizable.SetMaxSize(100, 50)

		resizable.SetSize(200, 20)

		w, _ := resizable.Size()
		if w > 100 {
			t.Errorf("Expected width <= 100, got %d", w)
		}
	})

	t.Run("Resize above max height", func(t *testing.T) {
		mock := newMockComponent("Test")
		resizable := NewResizable(mock, 40, 20)
		resizable.SetMaxSize(100, 50)

		resizable.SetSize(40, 200)

		_, h := resizable.Size()
		if h > 50 {
			t.Errorf("Expected height <= 50, got %d", h)
		}
	})

	t.Run("Resize with conflicting constraints (min > max)", func(t *testing.T) {
		mock := newMockComponent("Test")
		resizable := NewResizable(mock, 40, 20)
		// Set min larger than max (edge case)
		resizable.SetMinSize(100, 50)
		resizable.SetMaxSize(50, 25)

		resizable.SetSize(75, 30)

		w, h := resizable.Size()
		// Implementation may respect max over min or vice versa
		// Just verify constraints are applied (not the exact value)
		if w > 100 {
			t.Errorf("Expected width <= 100, got %d", w)
		}
		if h > 50 {
			t.Errorf("Expected height <= 50, got %d", h)
		}
		// Verify some constraint was applied
		if w == 75 && h == 30 {
			t.Error("Expected some constraint to be applied")
		}
	})

	t.Run("Resize to zero dimensions", func(t *testing.T) {
		mock := newMockComponent("Test")
		resizable := NewResizable(mock, 40, 20)

		resizable.SetSize(0, 0)

		w, h := resizable.Size()
		// Should be constrained by default min size
		if w < 10 {
			t.Errorf("Expected width >= default min (10), got %d", w)
		}
		if h < 3 {
			t.Errorf("Expected height >= default min (3), got %d", h)
		}
	})

	t.Run("Resize to negative dimensions", func(t *testing.T) {
		mock := newMockComponent("Test")
		resizable := NewResizable(mock, 40, 20)

		resizable.SetSize(-10, -5)

		w, h := resizable.Size()
		// Should be constrained to minimum positive values
		if w < 0 {
			t.Errorf("Expected non-negative width, got %d", w)
		}
		if h < 0 {
			t.Errorf("Expected non-negative height, got %d", h)
		}
	})

	t.Run("Resize with max = 0 (unlimited)", func(t *testing.T) {
		mock := newMockComponent("Test")
		resizable := NewResizable(mock, 40, 20)
		resizable.SetMaxSize(0, 0) // 0 means unlimited

		resizable.SetSize(1000, 500)

		w, h := resizable.Size()
		// Should not be constrained by max when max = 0
		if w != 1000 {
			t.Errorf("Expected unconstrained width 1000, got %d", w)
		}
		if h != 500 {
			t.Errorf("Expected unconstrained height 500, got %d", h)
		}
	})
}

// TestResizeHandleDetection tests resize handle position detection
func TestResizeHandleDetection(t *testing.T) {
	mock := newMockComponent("Test")
	resizable := NewResizable(mock, 40, 20)

	t.Run("All 8 handle positions present", func(t *testing.T) {
		if len(resizable.resizeHandles) != 8 {
			t.Errorf("Expected 8 resize handles, got %d", len(resizable.resizeHandles))
		}
	})

	t.Run("TopLeft handle exists", func(t *testing.T) {
		found := false
		for _, handle := range resizable.resizeHandles {
			if handle.Position == TopLeft {
				found = true
				break
			}
		}
		if !found {
			t.Error("TopLeft handle not found")
		}
	})

	t.Run("Top handle exists", func(t *testing.T) {
		found := false
		for _, handle := range resizable.resizeHandles {
			if handle.Position == Top {
				found = true
				break
			}
		}
		if !found {
			t.Error("Top handle not found")
		}
	})

	t.Run("TopRight handle exists", func(t *testing.T) {
		found := false
		for _, handle := range resizable.resizeHandles {
			if handle.Position == TopRight {
				found = true
				break
			}
		}
		if !found {
			t.Error("TopRight handle not found")
		}
	})

	t.Run("Right handle exists", func(t *testing.T) {
		found := false
		for _, handle := range resizable.resizeHandles {
			if handle.Position == Right {
				found = true
				break
			}
		}
		if !found {
			t.Error("Right handle not found")
		}
	})

	t.Run("BottomRight handle exists", func(t *testing.T) {
		found := false
		for _, handle := range resizable.resizeHandles {
			if handle.Position == BottomRight {
				found = true
				break
			}
		}
		if !found {
			t.Error("BottomRight handle not found")
		}
	})

	t.Run("Bottom handle exists", func(t *testing.T) {
		found := false
		for _, handle := range resizable.resizeHandles {
			if handle.Position == Bottom {
				found = true
				break
			}
		}
		if !found {
			t.Error("Bottom handle not found")
		}
	})

	t.Run("BottomLeft handle exists", func(t *testing.T) {
		found := false
		for _, handle := range resizable.resizeHandles {
			if handle.Position == BottomLeft {
				found = true
				break
			}
		}
		if !found {
			t.Error("BottomLeft handle not found")
		}
	})

	t.Run("Left handle exists", func(t *testing.T) {
		found := false
		for _, handle := range resizable.resizeHandles {
			if handle.Position == Left {
				found = true
				break
			}
		}
		if !found {
			t.Error("Left handle not found")
		}
	})

	t.Run("Handle position strings", func(t *testing.T) {
		positions := []ResizePosition{
			TopLeft, Top, TopRight, Right,
			BottomRight, Bottom, BottomLeft, Left,
		}
		expected := []string{
			"TopLeft", "Top", "TopRight", "Right",
			"BottomRight", "Bottom", "BottomLeft", "Left",
		}

		for i, pos := range positions {
			if pos.String() != expected[i] {
				t.Errorf("Expected position %s, got %s", expected[i], pos.String())
			}
		}
	})

	t.Run("Unknown resize position", func(t *testing.T) {
		pos := ResizePosition(999)
		if pos.String() != "Unknown" {
			t.Errorf("Expected 'Unknown' for invalid position, got '%s'", pos.String())
		}
	})
}

// TestDragOperationEdgeCases tests edge cases in drag operations
func TestDragOperationEdgeCases(t *testing.T) {
	t.Run("UpdateDrag without StartDrag", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 10, 10)

		initialX, initialY := draggable.Position()

		// Try to update drag without starting
		draggable.UpdateDrag(20, 20)

		x, y := draggable.Position()
		// Position should not change without active drag
		if x != initialX || y != initialY {
			t.Errorf("Expected position unchanged (%d, %d), got (%d, %d)", initialX, initialY, x, y)
		}
	})

	t.Run("EndDrag without StartDrag", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 10, 10)

		// Should not panic
		draggable.EndDrag()

		if draggable.IsDragging() {
			t.Error("Expected dragging to be false")
		}
	})

	t.Run("Multiple StartDrag calls", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 10, 10)

		draggable.StartDrag(15, 15)
		draggable.StartDrag(20, 20)

		if !draggable.IsDragging() {
			t.Error("Expected dragging to be active")
		}
	})

	t.Run("Drag with drag handle set", func(t *testing.T) {
		mock := newMockComponent("Test")
		draggable := NewDraggable(mock, 10, 10)
		draggable.SetDragHandle(0, 0, 10, 2)

		// Dragging should only work when mouse is in drag handle area
		if !draggable.hasDragHandle {
			t.Error("Expected drag handle to be set")
		}
	})
}

// TestResizeOperationEdgeCases tests edge cases in resize operations
func TestResizeOperationEdgeCases(t *testing.T) {
	t.Run("StartResize without handle", func(t *testing.T) {
		mock := newMockComponent("Test")
		resizable := NewResizable(mock, 40, 20)

		initialW, initialH := resizable.Size()

		// Direct resize should still work
		resizable.SetSize(50, 25)

		w, h := resizable.Size()
		if w == initialW || h == initialH {
			t.Error("Expected size to change")
		}
	})

	t.Run("IsResizing state management", func(t *testing.T) {
		mock := newMockComponent("Test")
		resizable := NewResizable(mock, 40, 20)

		if resizable.IsResizing() {
			t.Error("Expected not resizing initially")
		}

		resizable.StartResize(0, 0)

		if !resizable.IsResizing() {
			t.Error("Expected resizing after StartResize")
		}

		resizable.EndResize()

		if resizable.IsResizing() {
			t.Error("Expected not resizing after EndResize")
		}
	})

	t.Run("UpdateResize without StartResize", func(t *testing.T) {
		mock := newMockComponent("Test")
		resizable := NewResizable(mock, 40, 20)

		initialW, initialH := resizable.Size()

		// Try to update without starting
		resizable.UpdateResize(100, 100)

		w, h := resizable.Size()
		// Size should not change without active resize
		if w != initialW || h != initialH {
			t.Errorf("Expected size unchanged (%d, %d), got (%d, %d)", initialW, initialH, w, h)
		}
	})
}

// TestBoundaryConditions tests additional boundary conditions
func TestBoundaryConditions(t *testing.T) {
	t.Run("IsInBounds with negative coordinates", func(t *testing.T) {
		rect := layout.Rectangle{X: 0, Y: 0, Width: 100, Height: 100}

		if IsInBounds(-5, -5, rect) {
			t.Error("Expected negative coordinates to be out of bounds")
		}
	})

	t.Run("IsInBounds with exact boundaries", func(t *testing.T) {
		rect := layout.Rectangle{X: 10, Y: 10, Width: 20, Height: 20}

		// Top-left corner (inclusive)
		if !IsInBounds(10, 10, rect) {
			t.Error("Expected top-left corner to be in bounds")
		}

		// Bottom-right corner (exclusive)
		if IsInBounds(30, 30, rect) {
			t.Error("Expected bottom-right corner to be out of bounds (exclusive)")
		}

		// Just inside bottom-right
		if !IsInBounds(29, 29, rect) {
			t.Error("Expected just inside bottom-right to be in bounds")
		}
	})

	t.Run("IsInBounds with zero-size rectangle", func(t *testing.T) {
		rect := layout.Rectangle{X: 10, Y: 10, Width: 0, Height: 0}

		if IsInBounds(10, 10, rect) {
			t.Error("Expected zero-size rectangle to contain no points")
		}
	})

	t.Run("IsInBounds with negative-size rectangle", func(t *testing.T) {
		rect := layout.Rectangle{X: 10, Y: 10, Width: -10, Height: -10}

		if IsInBounds(5, 5, rect) {
			t.Error("Expected negative-size rectangle to be invalid")
		}
	})
}
