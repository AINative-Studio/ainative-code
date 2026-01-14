package layout

import (
	"testing"
)

// TestBoxLayoutVertical tests vertical box layout distribution
func TestBoxLayoutVertical(t *testing.T) {
	tests := []struct {
		name        string
		available   Rectangle
		components  []ComponentInfo
		want        map[string]Rectangle
		description string
	}{
		{
			name:      "equal distribution with two flex components",
			available: Rectangle{X: 0, Y: 0, Width: 100, Height: 100},
			components: []ComponentInfo{
				{ID: "top", Constraints: FlexConstraints(0, 1), Order: 0},
				{ID: "bottom", Constraints: FlexConstraints(0, 1), Order: 1},
			},
			want: map[string]Rectangle{
				"top":    {X: 0, Y: 0, Width: 100, Height: 50},
				"bottom": {X: 0, Y: 50, Width: 100, Height: 50},
			},
			description: "Two flex components with equal weight should split space 50/50",
		},
		{
			name:      "fixed and flex components",
			available: Rectangle{X: 0, Y: 0, Width: 80, Height: 50},
			components: []ComponentInfo{
				{ID: "header", Constraints: FixedConstraints(80, 5), Order: 0},
				{ID: "content", Constraints: FlexConstraints(10, 1), Order: 1},
				{ID: "footer", Constraints: FixedConstraints(80, 3), Order: 2},
			},
			want: map[string]Rectangle{
				"header":  {X: 0, Y: 0, Width: 80, Height: 5},
				"content": {X: 0, Y: 5, Width: 80, Height: 42},
				"footer":  {X: 0, Y: 47, Width: 80, Height: 3},
			},
			description: "Fixed components take their size, flex fills remaining",
		},
		{
			name:      "weighted flex components",
			available: Rectangle{X: 0, Y: 0, Width: 100, Height: 100},
			components: []ComponentInfo{
				{ID: "light", Constraints: Constraints{Grow: true, Weight: 1}, Order: 0},
				{ID: "heavy", Constraints: Constraints{Grow: true, Weight: 3}, Order: 1},
			},
			want: map[string]Rectangle{
				"light": {X: 0, Y: 0, Width: 100, Height: 25},
				"heavy": {X: 0, Y: 25, Width: 100, Height: 75},
			},
			description: "Components with weights 1:3 should get 25%:75% of space",
		},
		{
			name:      "minimum height constraints",
			available: Rectangle{X: 0, Y: 0, Width: 100, Height: 30},
			components: []ComponentInfo{
				{ID: "min1", Constraints: Constraints{MinHeight: 10, Grow: true, Weight: 1}, Order: 0},
				{ID: "min2", Constraints: Constraints{MinHeight: 20, Grow: true, Weight: 1}, Order: 1},
			},
			want: map[string]Rectangle{
				"min1": {X: 0, Y: 0, Width: 100, Height: 15},
				"min2": {X: 0, Y: 15, Width: 100, Height: 15},
			},
			description: "Minimum heights should be respected when distributing space",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := NewBoxLayout(Vertical)
			result := box.Calculate(tt.available, tt.components)

			// Check that all expected components are present
			if len(result) != len(tt.want) {
				t.Errorf("got %d components, want %d", len(result), len(tt.want))
			}

			// Check each component's bounds
			for id, wantBounds := range tt.want {
				gotBounds, exists := result[id]
				if !exists {
					t.Errorf("component %q not found in result", id)
					continue
				}

				if gotBounds != wantBounds {
					t.Errorf("component %q:\n  got  %+v\n  want %+v\n  (%s)",
						id, gotBounds, wantBounds, tt.description)
				}
			}
		})
	}
}

// TestBoxLayoutHorizontal tests horizontal box layout distribution
func TestBoxLayoutHorizontal(t *testing.T) {
	tests := []struct {
		name       string
		available  Rectangle
		components []ComponentInfo
		want       map[string]Rectangle
	}{
		{
			name:      "equal horizontal distribution",
			available: Rectangle{X: 0, Y: 0, Width: 100, Height: 50},
			components: []ComponentInfo{
				{ID: "left", Constraints: FlexConstraints(0, 1), Order: 0},
				{ID: "right", Constraints: FlexConstraints(0, 1), Order: 1},
			},
			want: map[string]Rectangle{
				"left":  {X: 0, Y: 0, Width: 50, Height: 50},
				"right": {X: 50, Y: 0, Width: 50, Height: 50},
			},
		},
		{
			name:      "fixed sidebar and flexible content",
			available: Rectangle{X: 0, Y: 0, Width: 100, Height: 50},
			components: []ComponentInfo{
				{ID: "sidebar", Constraints: FixedConstraints(20, 50), Order: 0},
				{ID: "content", Constraints: FlexConstraints(0, 1), Order: 1},
			},
			want: map[string]Rectangle{
				"sidebar": {X: 0, Y: 0, Width: 20, Height: 50},
				"content": {X: 20, Y: 0, Width: 80, Height: 50},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := NewBoxLayout(Horizontal)
			result := box.Calculate(tt.available, tt.components)

			for id, wantBounds := range tt.want {
				gotBounds, exists := result[id]
				if !exists {
					t.Errorf("component %q not found", id)
					continue
				}

				if gotBounds != wantBounds {
					t.Errorf("component %q: got %+v, want %+v", id, gotBounds, wantBounds)
				}
			}
		})
	}
}

// TestLayoutManager tests the layout manager integration
func TestLayoutManager(t *testing.T) {
	t.Run("register and unregister components", func(t *testing.T) {
		mgr := NewManager(NewBoxLayout(Vertical))

		// Register components
		err := mgr.RegisterComponent("comp1", DefaultConstraints())
		if err != nil {
			t.Fatalf("failed to register comp1: %v", err)
		}

		err = mgr.RegisterComponent("comp2", DefaultConstraints())
		if err != nil {
			t.Fatalf("failed to register comp2: %v", err)
		}

		if mgr.GetComponentCount() != 2 {
			t.Errorf("got %d components, want 2", mgr.GetComponentCount())
		}

		// Unregister component
		err = mgr.UnregisterComponent("comp1")
		if err != nil {
			t.Fatalf("failed to unregister comp1: %v", err)
		}

		if mgr.GetComponentCount() != 1 {
			t.Errorf("got %d components, want 1", mgr.GetComponentCount())
		}
	})

	t.Run("dirty flag management", func(t *testing.T) {
		mgr := NewManager(NewBoxLayout(Vertical))
		mgr.SetAvailableSize(100, 100)

		if !mgr.IsDirty() {
			t.Error("manager should be dirty after size change")
		}

		_ = mgr.RecalculateLayout()

		if mgr.IsDirty() {
			t.Error("manager should not be dirty after recalculation")
		}

		_ = mgr.RegisterComponent("test", DefaultConstraints())

		if !mgr.IsDirty() {
			t.Error("manager should be dirty after component registration")
		}
	})

	t.Run("bounds calculation", func(t *testing.T) {
		mgr := NewManager(NewBoxLayout(Vertical))
		mgr.SetAvailableSize(100, 100)

		_ = mgr.RegisterComponent("viewport", FlexConstraints(10, 1))
		_ = mgr.RegisterComponent("input", FixedConstraints(100, 3))
		_ = mgr.RegisterComponent("status", FixedConstraints(100, 1))

		// Force recalculation
		_ = mgr.RecalculateLayout()

		// Check viewport bounds (should take remaining space)
		viewportBounds := mgr.GetComponentBounds("viewport")
		if viewportBounds.Height != 96 { // 100 - 3 - 1 = 96
			t.Errorf("viewport height: got %d, want 96", viewportBounds.Height)
		}

		// Check input bounds
		inputBounds := mgr.GetComponentBounds("input")
		if inputBounds.Height != 3 {
			t.Errorf("input height: got %d, want 3", inputBounds.Height)
		}
		if inputBounds.Y != 96 {
			t.Errorf("input Y: got %d, want 96", inputBounds.Y)
		}

		// Check status bounds
		statusBounds := mgr.GetComponentBounds("status")
		if statusBounds.Height != 1 {
			t.Errorf("status height: got %d, want 1", statusBounds.Height)
		}
		if statusBounds.Y != 99 {
			t.Errorf("status Y: got %d, want 99", statusBounds.Y)
		}
	})
}

// TestResponsiveLayout tests responsive layout breakpoints
func TestResponsiveLayout(t *testing.T) {
	t.Run("breakpoint selection", func(t *testing.T) {
		compactLayout := NewBoxLayout(Vertical)
		fullLayout := NewBoxLayout(Horizontal)

		responsive := NewResponsiveLayout(compactLayout)
		responsive.AddBreakpoint(BreakpointBasic, fullLayout)

		// Test compact width (< 80)
		components := []ComponentInfo{
			{ID: "test", Constraints: FlexConstraints(0, 1), Order: 0},
		}

		// Width < 80 should use compact layout (vertical)
		result1 := responsive.Calculate(Rectangle{Width: 40, Height: 100}, components)
		if _, exists := result1["test"]; !exists {
			t.Error("component not found in compact layout")
		}

		// Width >= 80 should use full layout (horizontal)
		result2 := responsive.Calculate(Rectangle{Width: 100, Height: 100}, components)
		if _, exists := result2["test"]; !exists {
			t.Error("component not found in full layout")
		}
	})

	t.Run("breakpoint helper functions", func(t *testing.T) {
		if !IsCompactMode(30) {
			t.Error("30 should be compact mode")
		}
		if !IsBasicMode(50) {
			t.Error("50 should be basic mode")
		}
		if !IsFullMode(90) {
			t.Error("90 should be full mode")
		}
		if !IsExtendedMode(120) {
			t.Error("120 should be extended mode")
		}
	})

	t.Run("breakpoint names", func(t *testing.T) {
		responsive := NewResponsiveLayout(NewBoxLayout(Vertical))

		tests := []struct {
			width int
			want  string
		}{
			{30, "compact"},
			{50, "basic"},
			{90, "full"},
			{120, "extended"},
		}

		for _, tt := range tests {
			got := responsive.GetCurrentBreakpoint(tt.width)
			if got != tt.want {
				t.Errorf("width %d: got %q, want %q", tt.width, got, tt.want)
			}
		}
	})
}

// TestConstraintHelpers tests constraint helper functions
func TestConstraintHelpers(t *testing.T) {
	t.Run("default constraints", func(t *testing.T) {
		c := DefaultConstraints()
		if c.Grow || c.Shrink {
			t.Error("default constraints should not grow or shrink")
		}
		if c.Weight != 1 {
			t.Errorf("default weight: got %d, want 1", c.Weight)
		}
	})

	t.Run("fixed constraints", func(t *testing.T) {
		c := FixedConstraints(100, 50)
		if c.MinWidth != 100 || c.MaxWidth != 100 {
			t.Error("fixed constraints should have equal min/max width")
		}
		if c.MinHeight != 50 || c.MaxHeight != 50 {
			t.Error("fixed constraints should have equal min/max height")
		}
		if c.Grow || c.Shrink {
			t.Error("fixed constraints should not grow or shrink")
		}
	})

	t.Run("flex constraints", func(t *testing.T) {
		c := FlexConstraints(10, 2)
		if c.MinHeight != 10 {
			t.Errorf("flex min height: got %d, want 10", c.MinHeight)
		}
		if c.Weight != 2 {
			t.Errorf("flex weight: got %d, want 2", c.Weight)
		}
		if !c.Grow || !c.Shrink {
			t.Error("flex constraints should grow and shrink")
		}
	})
}

// TestBoxLayoutSpacing tests layout with spacing between components
func TestBoxLayoutSpacing(t *testing.T) {
	t.Run("vertical spacing", func(t *testing.T) {
		box := NewBoxLayoutWithSpacing(Vertical, 2)

		components := []ComponentInfo{
			{ID: "a", Constraints: FixedConstraints(100, 10), Order: 0},
			{ID: "b", Constraints: FixedConstraints(100, 10), Order: 1},
			{ID: "c", Constraints: FixedConstraints(100, 10), Order: 2},
		}

		result := box.Calculate(Rectangle{Width: 100, Height: 100}, components)

		// Check Y positions account for spacing
		if result["a"].Y != 0 {
			t.Errorf("component a Y: got %d, want 0", result["a"].Y)
		}
		if result["b"].Y != 12 { // 10 + 2 spacing
			t.Errorf("component b Y: got %d, want 12", result["b"].Y)
		}
		if result["c"].Y != 24 { // 10 + 2 + 10 + 2
			t.Errorf("component c Y: got %d, want 24", result["c"].Y)
		}
	})
}

// TestEmptyComponents tests layouts with no components
func TestEmptyComponents(t *testing.T) {
	box := NewBoxLayout(Vertical)
	result := box.Calculate(Rectangle{Width: 100, Height: 100}, []ComponentInfo{})

	if len(result) != 0 {
		t.Errorf("empty components should return empty bounds map, got %d items", len(result))
	}
}

// TestInvalidRegistration tests error handling in manager
func TestInvalidRegistration(t *testing.T) {
	mgr := NewManager(NewBoxLayout(Vertical))

	// Test empty ID
	err := mgr.RegisterComponent("", DefaultConstraints())
	if err == nil {
		t.Error("should return error for empty component ID")
	}

	// Test unregister non-existent
	err = mgr.UnregisterComponent("nonexistent")
	if err == nil {
		t.Error("should return error when unregistering non-existent component")
	}
}

// TestMinimalTerminalSize tests layout with very small terminal sizes
func TestMinimalTerminalSize(t *testing.T) {
	mgr := NewManager(NewBoxLayout(Vertical))
	mgr.SetAvailableSize(10, 5) // Very small terminal

	_ = mgr.RegisterComponent("viewport", FlexConstraints(1, 1))
	_ = mgr.RegisterComponent("input", FixedConstraints(10, 1))

	_ = mgr.RecalculateLayout()

	viewportBounds := mgr.GetComponentBounds("viewport")
	inputBounds := mgr.GetComponentBounds("input")

	// Viewport should get most space
	if viewportBounds.Height < 1 {
		t.Error("viewport should have at least 1 line in minimal terminal")
	}

	// Input should maintain its fixed height
	if inputBounds.Height != 1 {
		t.Errorf("input height: got %d, want 1", inputBounds.Height)
	}

	// Total height should not exceed available
	totalHeight := viewportBounds.Height + inputBounds.Height
	if totalHeight > 5 {
		t.Errorf("total height %d exceeds available height 5", totalHeight)
	}
}
