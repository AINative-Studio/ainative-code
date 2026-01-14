package components

import (
	"testing"
)

// TestMultiColumnFocusManagement tests focus management in multi-column layouts
func TestMultiColumnFocusManagement(t *testing.T) {
	col1 := newMockComponent("Column 1")
	col2 := newMockComponent("Column 2")
	col3 := newMockComponent("Column 3")

	t.Run("SetActiveColumn()", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})

		mcl.SetActiveColumn(1)

		if mcl.GetActiveColumn() != 1 {
			t.Errorf("Expected active column 1, got %d", mcl.GetActiveColumn())
		}
	})

	t.Run("GetActiveColumn()", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})

		// Default should be 0
		if mcl.GetActiveColumn() != 0 {
			t.Errorf("Expected default active column 0, got %d", mcl.GetActiveColumn())
		}
	})

	t.Run("NextColumn()", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.SetActiveColumn(0)

		// Move to next
		mcl.SetActiveColumn((mcl.GetActiveColumn() + 1) % mcl.GetColumnCount())

		if mcl.GetActiveColumn() != 1 {
			t.Errorf("Expected active column 1, got %d", mcl.GetActiveColumn())
		}

		// Move to next again
		mcl.SetActiveColumn((mcl.GetActiveColumn() + 1) % mcl.GetColumnCount())

		if mcl.GetActiveColumn() != 2 {
			t.Errorf("Expected active column 2, got %d", mcl.GetActiveColumn())
		}
	})

	t.Run("PreviousColumn()", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.SetActiveColumn(2)

		// Move to previous
		newCol := mcl.GetActiveColumn() - 1
		if newCol < 0 {
			newCol = mcl.GetColumnCount() - 1
		}
		mcl.SetActiveColumn(newCol)

		if mcl.GetActiveColumn() != 1 {
			t.Errorf("Expected active column 1, got %d", mcl.GetActiveColumn())
		}
	})

	t.Run("Focus with invalid column index (negative)", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.SetActiveColumn(1)

		// Try to set negative index
		mcl.SetActiveColumn(-1)

		// Implementation should handle gracefully
		activeCol := mcl.GetActiveColumn()
		if activeCol < 0 || activeCol >= mcl.GetColumnCount() {
			t.Errorf("Expected valid column index, got %d", activeCol)
		}
	})

	t.Run("Focus with invalid column index (too large)", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.SetActiveColumn(1)

		// Try to set out of bounds index
		mcl.SetActiveColumn(10)

		// Implementation should handle gracefully
		activeCol := mcl.GetActiveColumn()
		if activeCol < 0 || activeCol >= mcl.GetColumnCount() {
			t.Errorf("Expected valid column index, got %d", activeCol)
		}
	})

	t.Run("Focus wrapping forward", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.SetActiveColumn(2) // Last column

		// Move to next (should wrap to 0)
		mcl.SetActiveColumn((mcl.GetActiveColumn() + 1) % mcl.GetColumnCount())

		if mcl.GetActiveColumn() != 0 {
			t.Errorf("Expected wrap to column 0, got %d", mcl.GetActiveColumn())
		}
	})

	t.Run("Focus wrapping backward", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.SetActiveColumn(0) // First column

		// Move to previous (should wrap to last)
		newCol := mcl.GetActiveColumn() - 1
		if newCol < 0 {
			newCol = mcl.GetColumnCount() - 1
		}
		mcl.SetActiveColumn(newCol)

		if mcl.GetActiveColumn() != 2 {
			t.Errorf("Expected wrap to column 2, got %d", mcl.GetActiveColumn())
		}
	})

	t.Run("Single column layout", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1})

		if mcl.GetColumnCount() != 1 {
			t.Errorf("Expected 1 column, got %d", mcl.GetColumnCount())
		}

		// Active column should stay at 0
		mcl.SetActiveColumn(0)
		if mcl.GetActiveColumn() != 0 {
			t.Errorf("Expected active column 0, got %d", mcl.GetActiveColumn())
		}

		// Next should wrap to 0
		mcl.SetActiveColumn((mcl.GetActiveColumn() + 1) % mcl.GetColumnCount())
		if mcl.GetActiveColumn() != 0 {
			t.Errorf("Expected active column to stay at 0, got %d", mcl.GetActiveColumn())
		}
	})

	t.Run("Empty column layout", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{})

		if mcl.GetColumnCount() != 0 {
			t.Errorf("Expected 0 columns, got %d", mcl.GetColumnCount())
		}
	})
}

// TestMultiColumnWidthCalculation tests column width calculations
func TestMultiColumnWidthCalculation(t *testing.T) {
	col1 := newMockComponent("Column 1")
	col2 := newMockComponent("Column 2")
	col3 := newMockComponent("Column 3")

	t.Run("Fixed width columns", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.SetColumnWidths([]int{20, 30, 25})
		mcl.SetSize(100, 50)

		// Verify widths are set
		if len(mcl.columnWidths) != 3 {
			t.Errorf("Expected 3 column widths, got %d", len(mcl.columnWidths))
		}

		if mcl.columnWidths[0] != 20 || mcl.columnWidths[1] != 30 || mcl.columnWidths[2] != 25 {
			t.Errorf("Expected widths [20, 30, 25], got %v", mcl.columnWidths)
		}
	})

	t.Run("Flex width columns (0 = flex)", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.SetColumnWidths([]int{20, 0, 0}) // First fixed, others flex
		mcl.SetSize(100, 50)

		// Widths should be stored
		if mcl.columnWidths[0] != 20 {
			t.Errorf("Expected first column width 20, got %d", mcl.columnWidths[0])
		}

		if mcl.columnWidths[1] != 0 || mcl.columnWidths[2] != 0 {
			t.Error("Expected flex columns to have width 0")
		}
	})

	t.Run("Mixed fixed and flex columns", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.SetColumnWidths([]int{20, 0, 30}) // Fixed, flex, fixed
		mcl.SetSize(100, 50)

		if mcl.columnWidths[0] != 20 || mcl.columnWidths[1] != 0 || mcl.columnWidths[2] != 30 {
			t.Errorf("Expected widths [20, 0, 30], got %v", mcl.columnWidths)
		}
	})

	t.Run("Weight distribution", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.SetColumnWeights([]int{1, 2, 1}) // Second column gets 2x weight
		mcl.SetSize(100, 50)

		// Weights should be stored
		if len(mcl.weights) != 3 {
			t.Errorf("Expected 3 weights, got %d", len(mcl.weights))
		}

		if mcl.weights[0] != 1 || mcl.weights[1] != 2 || mcl.weights[2] != 1 {
			t.Errorf("Expected weights [1, 2, 1], got %v", mcl.weights)
		}
	})

	t.Run("Invalid width array (wrong length)", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})

		// Try to set wrong number of widths
		mcl.SetColumnWidths([]int{20, 30}) // Only 2, but need 3

		// Widths should not be updated
		if len(mcl.columnWidths) != 3 {
			t.Error("Column widths should not be updated with wrong length array")
		}
	})

	t.Run("Invalid weight array (wrong length)", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})

		// Try to set wrong number of weights
		mcl.SetColumnWeights([]int{1, 2}) // Only 2, but need 3

		// Weights should not be updated
		if len(mcl.weights) != 3 {
			t.Error("Column weights should not be updated with wrong length array")
		}
	})

	t.Run("With insufficient total width", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})
		mcl.SetColumnWidths([]int{40, 40, 40}) // Total 120 but width is 100
		mcl.SetSize(100, 50)

		// Should still store the widths
		if mcl.columnWidths[0] != 40 {
			t.Errorf("Expected width 40, got %d", mcl.columnWidths[0])
		}
	})
}

// TestMultiColumnAlignment tests vertical alignment modes
func TestMultiColumnAlignment(t *testing.T) {
	col1 := newMockComponent("Column 1")
	col2 := newMockComponent("Column 2")

	t.Run("AlignTop", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})
		mcl.SetAlignment(ColumnAlignTop)

		if mcl.alignment != ColumnAlignTop {
			t.Errorf("Expected ColumnAlignTop, got %v", mcl.alignment)
		}

		if mcl.alignment.String() != "Top" {
			t.Errorf("Expected 'Top', got '%s'", mcl.alignment.String())
		}
	})

	t.Run("AlignCenter", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})
		mcl.SetAlignment(ColumnAlignCenter)

		if mcl.alignment != ColumnAlignCenter {
			t.Errorf("Expected ColumnAlignCenter, got %v", mcl.alignment)
		}

		if mcl.alignment.String() != "Center" {
			t.Errorf("Expected 'Center', got '%s'", mcl.alignment.String())
		}
	})

	t.Run("AlignBottom", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})
		mcl.SetAlignment(ColumnAlignBottom)

		if mcl.alignment != ColumnAlignBottom {
			t.Errorf("Expected ColumnAlignBottom, got %v", mcl.alignment)
		}

		if mcl.alignment.String() != "Bottom" {
			t.Errorf("Expected 'Bottom', got '%s'", mcl.alignment.String())
		}
	})

	t.Run("AlignStretch", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})
		mcl.SetAlignment(ColumnAlignStretch)

		if mcl.alignment != ColumnAlignStretch {
			t.Errorf("Expected ColumnAlignStretch, got %v", mcl.alignment)
		}

		if mcl.alignment.String() != "Stretch" {
			t.Errorf("Expected 'Stretch', got '%s'", mcl.alignment.String())
		}
	})

	t.Run("Default alignment", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})

		// Default should be Top
		if mcl.alignment != ColumnAlignTop {
			t.Errorf("Expected default ColumnAlignTop, got %v", mcl.alignment)
		}
	})

	t.Run("Unknown alignment string", func(t *testing.T) {
		align := ColumnAlignment(999)

		if align.String() != "Unknown" {
			t.Errorf("Expected 'Unknown', got '%s'", align.String())
		}
	})
}

// TestMultiColumnLayoutOperations tests add/remove operations
func TestMultiColumnLayoutOperations(t *testing.T) {
	col1 := newMockComponent("Column 1")
	col2 := newMockComponent("Column 2")
	col3 := newMockComponent("Column 3")

	t.Run("AddColumn increases count", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})

		initialCount := mcl.GetColumnCount()
		mcl.AddColumn(col3)

		if mcl.GetColumnCount() != initialCount+1 {
			t.Errorf("Expected column count %d, got %d", initialCount+1, mcl.GetColumnCount())
		}
	})

	t.Run("RemoveColumn decreases count", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})

		initialCount := mcl.GetColumnCount()
		mcl.RemoveColumn(1)

		if mcl.GetColumnCount() != initialCount-1 {
			t.Errorf("Expected column count %d, got %d", initialCount-1, mcl.GetColumnCount())
		}
	})

	t.Run("RemoveColumn at start", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})

		mcl.RemoveColumn(0)

		if mcl.GetColumnCount() != 2 {
			t.Errorf("Expected 2 columns, got %d", mcl.GetColumnCount())
		}
	})

	t.Run("RemoveColumn at end", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})

		mcl.RemoveColumn(2)

		if mcl.GetColumnCount() != 2 {
			t.Errorf("Expected 2 columns, got %d", mcl.GetColumnCount())
		}
	})

	t.Run("RemoveColumn with invalid index", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})

		initialCount := mcl.GetColumnCount()

		// Try to remove invalid index
		mcl.RemoveColumn(10)

		// Count should not change
		if mcl.GetColumnCount() != initialCount {
			t.Error("Column count should not change with invalid index")
		}
	})

	t.Run("RemoveColumn with negative index", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})

		initialCount := mcl.GetColumnCount()

		// Try to remove negative index
		mcl.RemoveColumn(-1)

		// Count should not change
		if mcl.GetColumnCount() != initialCount {
			t.Error("Column count should not change with negative index")
		}
	})

	t.Run("GetColumn returns correct component", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})

		column := mcl.GetColumn(1)
		if column != col2 {
			t.Error("Expected GetColumn(1) to return col2")
		}
	})

	t.Run("GetColumn with invalid index", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2, col3})

		column := mcl.GetColumn(10)
		if column != nil {
			t.Error("Expected GetColumn with invalid index to return nil")
		}
	})
}

// TestMultiColumnGapSetting tests gap configuration
func TestMultiColumnGapSetting(t *testing.T) {
	col1 := newMockComponent("Column 1")
	col2 := newMockComponent("Column 2")

	t.Run("SetGap updates gap size", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})

		mcl.SetGap(3)

		if mcl.gaps != 3 {
			t.Errorf("Expected gap 3, got %d", mcl.gaps)
		}
	})

	t.Run("Default gap size", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})

		if mcl.gaps != 1 {
			t.Errorf("Expected default gap 1, got %d", mcl.gaps)
		}
	})

	t.Run("Zero gap size", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})

		mcl.SetGap(0)

		if mcl.gaps != 0 {
			t.Errorf("Expected gap 0, got %d", mcl.gaps)
		}
	})

	t.Run("Large gap size", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})

		mcl.SetGap(10)

		if mcl.gaps != 10 {
			t.Errorf("Expected gap 10, got %d", mcl.gaps)
		}
	})
}

// TestMultiColumnSizeManagement tests size setting and retrieval
func TestMultiColumnSizeManagement(t *testing.T) {
	col1 := newMockComponent("Column 1")
	col2 := newMockComponent("Column 2")

	t.Run("SetSize updates dimensions", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})

		mcl.SetSize(120, 60)

		w, h := mcl.GetSize()
		if w != 120 || h != 60 {
			t.Errorf("Expected size (120, 60), got (%d, %d)", w, h)
		}
	})

	t.Run("GetSize returns current dimensions", func(t *testing.T) {
		mcl := NewMultiColumnLayout([]Component{col1, col2})

		w, h := mcl.GetSize()
		// Default size
		if w != 80 || h != 24 {
			t.Errorf("Expected default size (80, 24), got (%d, %d)", w, h)
		}
	})
}
