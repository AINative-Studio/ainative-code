package layout

import "sort"

// Breakpoint constants from statusbar.go (preserved from existing implementation)
const (
	// BreakpointCompact is the threshold for compact mode (< 40 cols)
	BreakpointCompact = 40
	// BreakpointBasic is the threshold for basic mode (40-80 cols)
	BreakpointBasic = 80
	// BreakpointFull is the threshold for full mode (80-100 cols)
	BreakpointFull = 100
	// BreakpointExtended is the threshold for extended mode (100+ cols)
	BreakpointExtended = 100
)

// BreakpointConfig defines a layout configuration for a specific breakpoint
type BreakpointConfig struct {
	MinWidth int    // Minimum width for this breakpoint
	Layout   Layout // Layout to use at this breakpoint
}

// ResponsiveLayout switches between different layouts based on container width
type ResponsiveLayout struct {
	breakpoints []BreakpointConfig
	defaultLayout Layout
	lastWidth   int
	cachedLayout Layout
}

// NewResponsiveLayout creates a new responsive layout with a default layout
func NewResponsiveLayout(defaultLayout Layout) *ResponsiveLayout {
	return &ResponsiveLayout{
		breakpoints:   []BreakpointConfig{},
		defaultLayout: defaultLayout,
		lastWidth:     -1,
		cachedLayout:  defaultLayout,
	}
}

// AddBreakpoint adds a layout configuration for a specific width breakpoint
func (r *ResponsiveLayout) AddBreakpoint(minWidth int, layout Layout) {
	r.breakpoints = append(r.breakpoints, BreakpointConfig{
		MinWidth: minWidth,
		Layout:   layout,
	})

	// Sort breakpoints by width (descending) for correct matching
	sort.Slice(r.breakpoints, func(i, j int) bool {
		return r.breakpoints[i].MinWidth > r.breakpoints[j].MinWidth
	})

	// Clear cached layout since breakpoints changed
	r.lastWidth = -1
}

// Calculate selects the appropriate layout based on width and calculates component bounds
func (r *ResponsiveLayout) Calculate(available Rectangle, components []ComponentInfo) map[string]Rectangle {
	// Select layout based on width
	selectedLayout := r.selectLayout(available.Width)

	// Calculate using selected layout
	return selectedLayout.Calculate(available, components)
}

// selectLayout chooses the appropriate layout for the given width
func (r *ResponsiveLayout) selectLayout(width int) Layout {
	// Return cached layout if width hasn't changed
	if width == r.lastWidth && r.cachedLayout != nil {
		return r.cachedLayout
	}

	// Find the first matching breakpoint (sorted descending)
	for _, bp := range r.breakpoints {
		if width >= bp.MinWidth {
			r.lastWidth = width
			r.cachedLayout = bp.Layout
			return bp.Layout
		}
	}

	// No breakpoint matched, use default
	r.lastWidth = width
	r.cachedLayout = r.defaultLayout
	return r.defaultLayout
}

// GetCurrentBreakpoint returns the name of the current breakpoint for the given width
func (r *ResponsiveLayout) GetCurrentBreakpoint(width int) string {
	if width < BreakpointCompact {
		return "compact"
	} else if width < BreakpointBasic {
		return "basic"
	} else if width < BreakpointFull {
		return "full"
	}
	return "extended"
}

// IsCompactMode returns true if width is in compact mode
func IsCompactMode(width int) bool {
	return width < BreakpointCompact
}

// IsBasicMode returns true if width is in basic mode
func IsBasicMode(width int) bool {
	return width >= BreakpointCompact && width < BreakpointBasic
}

// IsFullMode returns true if width is in full mode
func IsFullMode(width int) bool {
	return width >= BreakpointBasic && width < BreakpointFull
}

// IsExtendedMode returns true if width is in extended mode
func IsExtendedMode(width int) bool {
	return width >= BreakpointExtended
}

// ClearBreakpoints removes all breakpoint configurations
func (r *ResponsiveLayout) ClearBreakpoints() {
	r.breakpoints = []BreakpointConfig{}
	r.lastWidth = -1
	r.cachedLayout = r.defaultLayout
}

// GetBreakpointCount returns the number of configured breakpoints
func (r *ResponsiveLayout) GetBreakpointCount() int {
	return len(r.breakpoints)
}

// SetDefaultLayout changes the default layout
func (r *ResponsiveLayout) SetDefaultLayout(layout Layout) {
	r.defaultLayout = layout
	r.lastWidth = -1
	r.cachedLayout = nil
}
