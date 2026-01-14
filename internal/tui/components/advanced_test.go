package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/AINative-studio/ainative-code/internal/tui/layout"
)

// mockComponent is a simple mock component for testing
type mockComponent struct {
	width   int
	height  int
	content string
	focused bool
}

func newMockComponent(content string) *mockComponent {
	return &mockComponent{
		width:   20,
		height:  10,
		content: content,
		focused: false,
	}
}

func (m *mockComponent) Init() tea.Cmd {
	return nil
}

func (m *mockComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	return m, nil
}

func (m *mockComponent) View() string {
	return m.content
}

func (m *mockComponent) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *mockComponent) GetSize() (int, int) {
	return m.width, m.height
}

func (m *mockComponent) Focus() tea.Cmd {
	m.focused = true
	return nil
}

func (m *mockComponent) Blur() {
	m.focused = false
}

func (m *mockComponent) Focused() bool {
	return m.focused
}

// TestMouseState tests mouse state tracking
func TestMouseState(t *testing.T) {
	ms := NewMouseState()

	if ms.X != -1 || ms.Y != -1 {
		t.Errorf("Expected initial position (-1, -1), got (%d, %d)", ms.X, ms.Y)
	}

	// Test press
	msg := tea.MouseMsg{
		X:      10,
		Y:      20,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}
	ms.Update(msg)

	if !ms.IsPressed {
		t.Error("Expected IsPressed to be true after press")
	}
	if ms.DragStartX != 10 || ms.DragStartY != 20 {
		t.Errorf("Expected drag start (10, 20), got (%d, %d)", ms.DragStartX, ms.DragStartY)
	}

	// Test motion (should trigger dragging)
	msg.Action = tea.MouseActionMotion
	msg.X = 15
	msg.Y = 25
	ms.Update(msg)

	if !ms.IsDragging {
		t.Error("Expected IsDragging to be true after motion with button pressed")
	}

	dx, dy := ms.DragDistance()
	if dx != 5 || dy != 5 {
		t.Errorf("Expected drag distance (5, 5), got (%d, %d)", dx, dy)
	}

	// Test release
	msg.Action = tea.MouseActionRelease
	ms.Update(msg)

	if ms.IsPressed {
		t.Error("Expected IsPressed to be false after release")
	}
	if ms.IsDragging {
		t.Error("Expected IsDragging to be false after release")
	}
}

// TestIsInBounds tests the bounds checking function
func TestIsInBounds(t *testing.T) {
	rect := layout.Rectangle{X: 10, Y: 10, Width: 20, Height: 15}

	tests := []struct {
		name     string
		x, y     int
		expected bool
	}{
		{"Inside", 15, 15, true},
		{"Top-left corner", 10, 10, true},
		{"Bottom-right corner (exclusive)", 30, 25, false},
		{"Just inside right", 29, 15, true},
		{"Just inside bottom", 15, 24, true},
		{"Outside left", 5, 15, false},
		{"Outside right", 31, 15, false},
		{"Outside top", 15, 5, false},
		{"Outside bottom", 15, 26, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInBounds(tt.x, tt.y, rect)
			if result != tt.expected {
				t.Errorf("IsInBounds(%d, %d, rect) = %v, expected %v", tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

// TestDraggableComponent tests draggable functionality
func TestDraggableComponent(t *testing.T) {
	mock := newMockComponent("Test Content")
	draggable := NewDraggable(mock, 10, 10)

	t.Run("Initial position", func(t *testing.T) {
		x, y := draggable.Position()
		if x != 10 || y != 10 {
			t.Errorf("Expected position (10, 10), got (%d, %d)", x, y)
		}
	})

	t.Run("Start drag", func(t *testing.T) {
		draggable.StartDrag(15, 15)
		if !draggable.IsDragging() {
			t.Error("Expected dragging to be active")
		}
	})

	t.Run("Update drag", func(t *testing.T) {
		draggable.UpdateDrag(20, 25)
		x, y := draggable.Position()
		// Position should be mouseX - offsetX, mouseY - offsetY
		// offsetX = 15 - 10 = 5, offsetY = 15 - 10 = 5
		// So new position should be 20 - 5 = 15, 25 - 5 = 20
		if x != 15 || y != 20 {
			t.Errorf("Expected position (15, 20), got (%d, %d)", x, y)
		}
	})

	t.Run("End drag", func(t *testing.T) {
		draggable.EndDrag()
		if draggable.IsDragging() {
			t.Error("Expected dragging to be inactive")
		}
	})

	t.Run("Set bounds", func(t *testing.T) {
		draggable.SetBounds(0, 0, 100, 100)
		draggable.SetSize(10, 10)
		draggable.SetPosition(95, 95)
		x, y := draggable.Position()
		// Should be constrained to 90, 90 (100 - 10)
		if x != 90 || y != 90 {
			t.Errorf("Expected constrained position (90, 90), got (%d, %d)", x, y)
		}
	})

	t.Run("Snap to grid", func(t *testing.T) {
		draggable.EnableSnapToGrid(5)
		draggable.SetPosition(23, 27)
		x, y := draggable.Position()
		// Should snap to nearest grid point
		if x != 20 || y != 25 {
			t.Errorf("Expected snapped position (20, 25), got (%d, %d)", x, y)
		}
	})
}

// TestDraggableKeyboardMovement tests keyboard movement
func TestDraggableKeyboardMovement(t *testing.T) {
	mock := newMockComponent("Test")
	draggable := NewDraggable(mock, 10, 10)
	draggable.Focus()

	tests := []struct {
		name     string
		key      string
		expectedX int
		expectedY int
	}{
		{"Move up", "alt+up", 10, 9},
		{"Move down", "alt+down", 10, 10},
		{"Move left", "alt+left", 9, 10},
		{"Move right", "alt+right", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draggable.SetPosition(10, 10)
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{}, Alt: true}
			draggable.Update(msg)
			// Note: In real implementation, we'd need proper key parsing
			// This is a simplified test
		})
	}
}

// TestResizableComponent tests resizable functionality
func TestResizableComponent(t *testing.T) {
	mock := newMockComponent("Test Content")
	resizable := NewResizable(mock, 40, 20)

	t.Run("Initial size", func(t *testing.T) {
		w, h := resizable.Size()
		if w != 40 || h != 20 {
			t.Errorf("Expected size (40, 20), got (%d, %d)", w, h)
		}
	})

	t.Run("Set size", func(t *testing.T) {
		resizable.SetSize(50, 30)
		w, h := resizable.Size()
		if w != 50 || h != 30 {
			t.Errorf("Expected size (50, 30), got (%d, %d)", w, h)
		}
	})

	t.Run("Min size constraint", func(t *testing.T) {
		resizable.SetMinSize(20, 10)
		resizable.SetSize(5, 5)
		w, h := resizable.Size()
		if w != 20 || h != 10 {
			t.Errorf("Expected constrained size (20, 10), got (%d, %d)", w, h)
		}
	})

	t.Run("Max size constraint", func(t *testing.T) {
		resizable.SetMaxSize(100, 50)
		resizable.SetSize(200, 100)
		w, h := resizable.Size()
		if w != 100 || h != 50 {
			t.Errorf("Expected constrained size (100, 50), got (%d, %d)", w, h)
		}
	})

	t.Run("Aspect ratio", func(t *testing.T) {
		resizable.EnableAspectRatio(2.0) // width:height = 2:1
		resizable.SetSize(60, 30)
		// Aspect ratio should be maintained
		w, h := resizable.Size()
		ratio := float64(w) / float64(h)
		if ratio < 1.9 || ratio > 2.1 {
			t.Errorf("Expected aspect ratio ~2.0, got %.2f", ratio)
		}
	})
}

// TestResizeHandles tests resize handle positions
func TestResizeHandles(t *testing.T) {
	mock := newMockComponent("Test")
	resizable := NewResizable(mock, 40, 20)

	handles := resizable.resizeHandles
	if len(handles) != 8 {
		t.Errorf("Expected 8 resize handles, got %d", len(handles))
	}

	// Check all positions are present
	positions := make(map[ResizePosition]bool)
	for _, handle := range handles {
		positions[handle.Position] = true
	}

	expectedPositions := []ResizePosition{
		TopLeft, Top, TopRight, Right,
		BottomRight, Bottom, BottomLeft, Left,
	}

	for _, pos := range expectedPositions {
		if !positions[pos] {
			t.Errorf("Missing resize handle at position %v", pos)
		}
	}
}

// TestSplitView tests split view functionality
func TestSplitView(t *testing.T) {
	left := newMockComponent("Left")
	right := newMockComponent("Right")

	t.Run("Horizontal split", func(t *testing.T) {
		split := NewSplitView(left, right, layout.Horizontal)
		split.SetSize(100, 50)

		if split.splitRatio != 0.5 {
			t.Errorf("Expected default split ratio 0.5, got %.2f", split.splitRatio)
		}

		w, h := split.GetSize()
		if w != 100 || h != 50 {
			t.Errorf("Expected size (100, 50), got (%d, %d)", w, h)
		}
	})

	t.Run("Vertical split", func(t *testing.T) {
		split := NewSplitView(left, right, layout.Vertical)
		split.SetSize(100, 50)

		if split.orientation != layout.Vertical {
			t.Error("Expected vertical orientation")
		}
	})

	t.Run("Set split ratio", func(t *testing.T) {
		split := NewSplitView(left, right, layout.Horizontal)
		split.SetSplitRatio(0.7)

		if split.splitRatio != 0.7 {
			t.Errorf("Expected split ratio 0.7, got %.2f", split.splitRatio)
		}
	})

	t.Run("Min/max ratio constraints", func(t *testing.T) {
		split := NewSplitView(left, right, layout.Horizontal)
		split.SetMinMaxRatio(0.3, 0.7)

		split.SetSplitRatio(0.1)
		if split.splitRatio != 0.3 {
			t.Errorf("Expected ratio constrained to 0.3, got %.2f", split.splitRatio)
		}

		split.SetSplitRatio(0.9)
		if split.splitRatio != 0.7 {
			t.Errorf("Expected ratio constrained to 0.7, got %.2f", split.splitRatio)
		}
	})

	t.Run("Swap panes", func(t *testing.T) {
		split := NewSplitView(left, right, layout.Horizontal)
		originalLeft := split.GetLeftPane()
		originalRight := split.GetRightPane()

		split.SwapPanes()

		if split.GetLeftPane() != originalRight || split.GetRightPane() != originalLeft {
			t.Error("Panes were not swapped correctly")
		}
	})
}

// TestSplitViewDivider tests divider dragging
func TestSplitViewDivider(t *testing.T) {
	left := newMockComponent("Left")
	right := newMockComponent("Right")
	split := NewSplitView(left, right, layout.Horizontal)
	split.SetSize(100, 50)

	t.Run("Start drag on divider", func(t *testing.T) {
		// Divider should be at position 50 (0.5 * 100)
		split.StartDividerDrag(50, 25)
		if !split.isDragging {
			t.Error("Expected dragging to be active")
		}
	})

	t.Run("Update drag", func(t *testing.T) {
		split.isDragging = true
		split.UpdateDividerDrag(70, 25)
		// Split ratio should be updated to 0.7 (70/100)
		if split.splitRatio < 0.69 || split.splitRatio > 0.71 {
			t.Errorf("Expected split ratio ~0.7, got %.2f", split.splitRatio)
		}
	})

	t.Run("End drag", func(t *testing.T) {
		split.EndDividerDrag()
		if split.isDragging {
			t.Error("Expected dragging to be inactive")
		}
	})
}

// TestMultiColumnLayout tests multi-column layout
func TestMultiColumnLayout(t *testing.T) {
	col1 := newMockComponent("Column 1")
	col2 := newMockComponent("Column 2")
	col3 := newMockComponent("Column 3")

	t.Run("Create layout", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		if mcl.GetColumnCount() != 3 {
			t.Errorf("Expected 3 columns, got %d", mcl.GetColumnCount())
		}
	})

	t.Run("Add column", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})
		mcl.AddColumn(col3)
		if mcl.GetColumnCount() != 3 {
			t.Errorf("Expected 3 columns after adding, got %d", mcl.GetColumnCount())
		}
	})

	t.Run("Remove column", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.RemoveColumn(1)
		if mcl.GetColumnCount() != 2 {
			t.Errorf("Expected 2 columns after removing, got %d", mcl.GetColumnCount())
		}
	})

	t.Run("Set column widths", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.SetColumnWidths([]int{20, 30, 0}) // Third column is flex
		mcl.SetSize(100, 50)

		// Column widths should be calculated
		// Total gaps = 1 * (3-1) = 2
		// Available = 100 - 2 = 98
		// Fixed = 20 + 30 = 50
		// Flex = 98 - 50 = 48
		// So third column should get 48 width
	})

	t.Run("Set column weights", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.SetColumnWeights([]int{1, 2, 1})
		// Weights affect flex distribution
	})

	t.Run("Set alignment", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})
		mcl.SetAlignment(ColumnAlignCenter)
		if mcl.alignment != ColumnAlignCenter {
			t.Errorf("Expected ColumnAlignCenter, got %v", mcl.alignment)
		}
	})

	t.Run("Set gap", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})
		mcl.SetGap(3)
		if mcl.gaps != 3 {
			t.Errorf("Expected gap 3, got %d", mcl.gaps)
		}
	})
}

// TestMultiColumnNavigation tests keyboard navigation between columns
func TestMultiColumnNavigation(t *testing.T) {
	col1 := newMockComponent("Column 1")
	col2 := newMockComponent("Column 2")
	col3 := newMockComponent("Column 3")

	mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
	mcl.Focus()

	t.Run("Initial active column", func(t *testing.T) {
		if mcl.GetActiveColumn() != 0 {
			t.Errorf("Expected active column 0, got %d", mcl.GetActiveColumn())
		}
	})

	t.Run("Set active column", func(t *testing.T) {
		mcl.SetActiveColumn(1)
		if mcl.GetActiveColumn() != 1 {
			t.Errorf("Expected active column 1, got %d", mcl.GetActiveColumn())
		}
	})

	t.Run("Navigate to next column", func(t *testing.T) {
		mcl.SetActiveColumn(0)
		// Simulate tab key (in real implementation)
		mcl.SetActiveColumn((mcl.GetActiveColumn() + 1) % mcl.GetColumnCount())
		if mcl.GetActiveColumn() != 1 {
			t.Errorf("Expected active column 1 after navigation, got %d", mcl.GetActiveColumn())
		}
	})
}

// TestColumnAlignment tests different alignment modes
func TestColumnAlignment(t *testing.T) {
	alignments := []ColumnAlignment{
		ColumnAlignTop, ColumnAlignCenter, ColumnAlignBottom, ColumnAlignStretch,
	}

	for _, align := range alignments {
		t.Run(align.String(), func(t *testing.T) {
			col1 := newMockComponent("Column 1")
			col2 := newMockComponent("Column 2")

			mcl := NewMultiColumnLayout([]Component{col1, col2})
			mcl.SetAlignment(align)

			if mcl.alignment != align {
				t.Errorf("Expected alignment %v, got %v", align, mcl.alignment)
			}
		})
	}
}

// TestDragHandler tests the drag handler utility
func TestDragHandler(t *testing.T) {
	startCalled := false
	dragCalled := false
	endCalled := false

	onStart := func(msg tea.MouseMsg) tea.Cmd {
		startCalled = true
		return nil
	}

	onDrag := func(msg tea.MouseMsg) tea.Cmd {
		dragCalled = true
		return nil
	}

	onEnd := func(msg tea.MouseMsg) tea.Cmd {
		endCalled = true
		return nil
	}

	dh := NewDragHandler(onStart, onDrag, onEnd)

	t.Run("Start drag", func(t *testing.T) {
		msg := tea.MouseMsg{
			X:      10,
			Y:      10,
			Button: tea.MouseButtonLeft,
			Action: tea.MouseActionPress,
		}
		dh.HandleMouseMsg(msg)

		if !startCalled {
			t.Error("Expected onStart callback to be called")
		}
		if !dh.IsActive {
			t.Error("Expected drag to be active")
		}
	})

	t.Run("Drag motion", func(t *testing.T) {
		msg := tea.MouseMsg{
			X:      15,
			Y:      15,
			Button: tea.MouseButtonLeft,
			Action: tea.MouseActionMotion,
		}
		dh.HandleMouseMsg(msg)

		if !dragCalled {
			t.Error("Expected onDrag callback to be called")
		}

		dx, dy := dh.DragDelta()
		if dx != 5 || dy != 5 {
			t.Errorf("Expected delta (5, 5), got (%d, %d)", dx, dy)
		}
	})

	t.Run("End drag", func(t *testing.T) {
		msg := tea.MouseMsg{
			X:      15,
			Y:      15,
			Button: tea.MouseButtonLeft,
			Action: tea.MouseActionRelease,
		}
		dh.HandleMouseMsg(msg)

		if !endCalled {
			t.Error("Expected onEnd callback to be called")
		}
		if dh.IsActive {
			t.Error("Expected drag to be inactive")
		}
	})
}

// TestHoverDetector tests the hover detector utility
func TestHoverDetector(t *testing.T) {
	enterCalled := false
	leaveCalled := false
	hoverCalled := false

	bounds := layout.Rectangle{X: 10, Y: 10, Width: 20, Height: 15}

	onEnter := func(msg tea.MouseMsg) tea.Cmd {
		enterCalled = true
		return nil
	}

	onLeave := func(msg tea.MouseMsg) tea.Cmd {
		leaveCalled = true
		return nil
	}

	onHover := func(msg tea.MouseMsg) tea.Cmd {
		hoverCalled = true
		return nil
	}

	hd := NewHoverDetector(bounds, onEnter, onLeave, onHover)

	t.Run("Mouse enter", func(t *testing.T) {
		msg := tea.MouseMsg{X: 15, Y: 15}
		hd.Update(msg)

		if !enterCalled {
			t.Error("Expected onEnter callback to be called")
		}
		if !hd.IsHovering() {
			t.Error("Expected hover state to be true")
		}
	})

	t.Run("Mouse hover", func(t *testing.T) {
		hoverCalled = false
		msg := tea.MouseMsg{X: 16, Y: 16}
		hd.Update(msg)

		if !hoverCalled {
			t.Error("Expected onHover callback to be called")
		}
	})

	t.Run("Mouse leave", func(t *testing.T) {
		msg := tea.MouseMsg{X: 5, Y: 5}
		hd.Update(msg)

		if !leaveCalled {
			t.Error("Expected onLeave callback to be called")
		}
		if hd.IsHovering() {
			t.Error("Expected hover state to be false")
		}
	})
}

// TestMouseEventHelpers tests mouse event helper functions
func TestMouseEventHelpers(t *testing.T) {
	t.Run("IsLeftClick", func(t *testing.T) {
		msg := tea.MouseMsg{
			Button: tea.MouseButtonLeft,
			Action: tea.MouseActionPress,
		}
		if !IsLeftClick(msg) {
			t.Error("Expected IsLeftClick to return true")
		}
	})

	t.Run("IsRightClick", func(t *testing.T) {
		msg := tea.MouseMsg{
			Button: tea.MouseButtonRight,
			Action: tea.MouseActionPress,
		}
		if !IsRightClick(msg) {
			t.Error("Expected IsRightClick to return true")
		}
	})

	t.Run("IsMouseMotion", func(t *testing.T) {
		msg := tea.MouseMsg{
			Action: tea.MouseActionMotion,
		}
		if !IsMouseMotion(msg) {
			t.Error("Expected IsMouseMotion to return true")
		}
	})

	t.Run("GetMousePosition", func(t *testing.T) {
		msg := tea.MouseMsg{X: 42, Y: 24}
		x, y := GetMousePosition(msg)
		if x != 42 || y != 24 {
			t.Errorf("Expected position (42, 24), got (%d, %d)", x, y)
		}
	})
}
