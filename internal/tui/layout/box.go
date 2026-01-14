package layout

// BoxLayout arranges components in a single direction (vertical or horizontal)
type BoxLayout struct {
	orientation Orientation
	spacing     int
}

// NewBoxLayout creates a new box layout with the specified orientation
func NewBoxLayout(orientation Orientation) *BoxLayout {
	return &BoxLayout{
		orientation: orientation,
		spacing:     0,
	}
}

// NewBoxLayoutWithSpacing creates a new box layout with custom spacing
func NewBoxLayoutWithSpacing(orientation Orientation, spacing int) *BoxLayout {
	return &BoxLayout{
		orientation: orientation,
		spacing:     spacing,
	}
}

// Calculate computes component bounds based on box layout algorithm
func (b *BoxLayout) Calculate(available Rectangle, components []ComponentInfo) map[string]Rectangle {
	if b.orientation == Vertical {
		return b.calculateVertical(available, components)
	}
	return b.calculateHorizontal(available, components)
}

// calculateVertical distributes components vertically (top to bottom)
func (b *BoxLayout) calculateVertical(available Rectangle, components []ComponentInfo) map[string]Rectangle {
	bounds := make(map[string]Rectangle)

	if len(components) == 0 {
		return bounds
	}

	// Calculate total spacing
	totalSpacing := b.spacing * (len(components) - 1)
	availableHeight := available.Height - totalSpacing
	if availableHeight < 0 {
		availableHeight = 0
	}

	// Phase 1: Calculate fixed and minimum sizes
	fixedHeight := 0
	flexComponents := []ComponentInfo{}
	totalWeight := 0

	for _, comp := range components {
		c := comp.Constraints

		// Fixed height components
		if c.MaxHeight > 0 && c.MinHeight == c.MaxHeight {
			fixedHeight += c.MaxHeight
		} else if c.Grow {
			// Flexible components
			flexComponents = append(flexComponents, comp)
			if c.Weight > 0 {
				totalWeight += c.Weight
			} else {
				totalWeight += 1 // Default weight
			}
		} else {
			// Components with minimum height but not flexible
			if c.MinHeight > 0 {
				fixedHeight += c.MinHeight
			}
		}
	}

	// Phase 2: Distribute remaining space among flexible components
	remainingHeight := availableHeight - fixedHeight
	if remainingHeight < 0 {
		remainingHeight = 0
	}

	// Calculate flex heights
	flexHeights := make(map[string]int)
	if totalWeight > 0 && remainingHeight > 0 {
		for _, comp := range flexComponents {
			weight := comp.Constraints.Weight
			if weight <= 0 {
				weight = 1
			}
			flexHeight := (remainingHeight * weight) / totalWeight

			// Respect min/max constraints
			if comp.Constraints.MinHeight > 0 && flexHeight < comp.Constraints.MinHeight {
				flexHeight = comp.Constraints.MinHeight
			}
			if comp.Constraints.MaxHeight > 0 && flexHeight > comp.Constraints.MaxHeight {
				flexHeight = comp.Constraints.MaxHeight
			}

			flexHeights[comp.ID] = flexHeight
		}
	}

	// Phase 3: Position components
	currentY := available.Y
	for _, comp := range components {
		c := comp.Constraints
		componentHeight := 0

		// Determine component height
		if flexHeight, isFlex := flexHeights[comp.ID]; isFlex {
			componentHeight = flexHeight
		} else if c.MaxHeight > 0 && c.MinHeight == c.MaxHeight {
			// Fixed height
			componentHeight = c.MaxHeight
		} else if c.MinHeight > 0 {
			// Minimum height
			componentHeight = c.MinHeight
		}

		// Ensure height doesn't exceed available space
		if currentY+componentHeight > available.Y+available.Height {
			componentHeight = available.Y + available.Height - currentY
			if componentHeight < 0 {
				componentHeight = 0
			}
		}

		// Calculate width (components fill container width by default)
		componentWidth := available.Width
		if c.MaxWidth > 0 && componentWidth > c.MaxWidth {
			componentWidth = c.MaxWidth
		}
		if c.MinWidth > 0 && componentWidth < c.MinWidth {
			componentWidth = c.MinWidth
		}

		bounds[comp.ID] = Rectangle{
			X:      available.X,
			Y:      currentY,
			Width:  componentWidth,
			Height: componentHeight,
		}

		currentY += componentHeight + b.spacing
	}

	return bounds
}

// calculateHorizontal distributes components horizontally (left to right)
func (b *BoxLayout) calculateHorizontal(available Rectangle, components []ComponentInfo) map[string]Rectangle {
	bounds := make(map[string]Rectangle)

	if len(components) == 0 {
		return bounds
	}

	// Calculate total spacing
	totalSpacing := b.spacing * (len(components) - 1)
	availableWidth := available.Width - totalSpacing
	if availableWidth < 0 {
		availableWidth = 0
	}

	// Phase 1: Calculate fixed and minimum sizes
	fixedWidth := 0
	flexComponents := []ComponentInfo{}
	totalWeight := 0

	for _, comp := range components {
		c := comp.Constraints

		// Fixed width components
		if c.MaxWidth > 0 && c.MinWidth == c.MaxWidth {
			fixedWidth += c.MaxWidth
		} else if c.Grow {
			// Flexible components
			flexComponents = append(flexComponents, comp)
			if c.Weight > 0 {
				totalWeight += c.Weight
			} else {
				totalWeight += 1
			}
		} else {
			// Components with minimum width but not flexible
			if c.MinWidth > 0 {
				fixedWidth += c.MinWidth
			}
		}
	}

	// Phase 2: Distribute remaining space among flexible components
	remainingWidth := availableWidth - fixedWidth
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	// Calculate flex widths
	flexWidths := make(map[string]int)
	if totalWeight > 0 && remainingWidth > 0 {
		for _, comp := range flexComponents {
			weight := comp.Constraints.Weight
			if weight <= 0 {
				weight = 1
			}
			flexWidth := (remainingWidth * weight) / totalWeight

			// Respect min/max constraints
			if comp.Constraints.MinWidth > 0 && flexWidth < comp.Constraints.MinWidth {
				flexWidth = comp.Constraints.MinWidth
			}
			if comp.Constraints.MaxWidth > 0 && flexWidth > comp.Constraints.MaxWidth {
				flexWidth = comp.Constraints.MaxWidth
			}

			flexWidths[comp.ID] = flexWidth
		}
	}

	// Phase 3: Position components
	currentX := available.X
	for _, comp := range components {
		c := comp.Constraints
		componentWidth := 0

		// Determine component width
		if flexWidth, isFlex := flexWidths[comp.ID]; isFlex {
			componentWidth = flexWidth
		} else if c.MaxWidth > 0 && c.MinWidth == c.MaxWidth {
			// Fixed width
			componentWidth = c.MaxWidth
		} else if c.MinWidth > 0 {
			// Minimum width
			componentWidth = c.MinWidth
		}

		// Ensure width doesn't exceed available space
		if currentX+componentWidth > available.X+available.Width {
			componentWidth = available.X + available.Width - currentX
			if componentWidth < 0 {
				componentWidth = 0
			}
		}

		// Calculate height (components fill container height by default)
		componentHeight := available.Height
		if c.MaxHeight > 0 && componentHeight > c.MaxHeight {
			componentHeight = c.MaxHeight
		}
		if c.MinHeight > 0 && componentHeight < c.MinHeight {
			componentHeight = c.MinHeight
		}

		bounds[comp.ID] = Rectangle{
			X:      currentX,
			Y:      available.Y,
			Width:  componentWidth,
			Height: componentHeight,
		}

		currentX += componentWidth + b.spacing
	}

	return bounds
}

// SetSpacing updates the spacing between components
func (b *BoxLayout) SetSpacing(spacing int) {
	b.spacing = spacing
}

// GetSpacing returns the current spacing between components
func (b *BoxLayout) GetSpacing() int {
	return b.spacing
}

// SetOrientation changes the layout orientation
func (b *BoxLayout) SetOrientation(orientation Orientation) {
	b.orientation = orientation
}

// GetOrientation returns the current layout orientation
func (b *BoxLayout) GetOrientation() Orientation {
	return b.orientation
}
