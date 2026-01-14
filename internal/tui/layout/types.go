package layout

// Rectangle represents a positioned area in the UI
type Rectangle struct {
	X      int // X position (left edge)
	Y      int // Y position (top edge)
	Width  int // Width in characters
	Height int // Height in lines
}

// Orientation defines the direction of layout flow
type Orientation int

const (
	// Vertical orientation stacks components top to bottom
	Vertical Orientation = iota
	// Horizontal orientation arranges components left to right
	Horizontal
)

// Constraints define size requirements and flexibility for a component
type Constraints struct {
	MinWidth  int  // Minimum width (0 = no minimum)
	MinHeight int  // Minimum height (0 = no minimum)
	MaxWidth  int  // Maximum width (0 = unlimited)
	MaxHeight int  // Maximum height (0 = unlimited)
	Grow      bool // Whether the component can grow to fill available space
	Shrink    bool // Whether the component can shrink below its preferred size
	Weight    int  // Flex weight for distributing extra space (higher = more space)
}

// DefaultConstraints returns constraints with sensible defaults
func DefaultConstraints() Constraints {
	return Constraints{
		MinWidth:  0,
		MinHeight: 0,
		MaxWidth:  0,
		MaxHeight: 0,
		Grow:      false,
		Shrink:    false,
		Weight:    1,
	}
}

// FixedConstraints creates constraints for a fixed-size component
func FixedConstraints(width, height int) Constraints {
	return Constraints{
		MinWidth:  width,
		MinHeight: height,
		MaxWidth:  width,
		MaxHeight: height,
		Grow:      false,
		Shrink:    false,
		Weight:    0,
	}
}

// FlexConstraints creates constraints for a flexible component
func FlexConstraints(minHeight, weight int) Constraints {
	return Constraints{
		MinWidth:  0,
		MinHeight: minHeight,
		MaxWidth:  0,
		MaxHeight: 0,
		Grow:      true,
		Shrink:    true,
		Weight:    weight,
	}
}
