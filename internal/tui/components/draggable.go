package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/AINative-studio/ainative-code/internal/tui/layout"
)

// DraggableComponent wraps any component and makes it draggable with mouse and keyboard
type DraggableComponent struct {
	component    Component           // The wrapped component
	x            int                 // Current X position
	y            int                 // Current Y position
	startX       int                 // Position when drag started
	startY       int                 // Position when drag started
	offsetX      int                 // Mouse offset from component origin
	offsetY      int                 // Mouse offset from component origin
	isDragging   bool                // Whether currently dragging
	dragHandle   layout.Rectangle    // Area that can be dragged (e.g., title bar)
	bounds       layout.Rectangle    // Constraint boundaries
	snapToGrid   bool                // Whether to snap to grid
	gridSize     int                 // Grid size for snapping
	width        int                 // Component width
	height       int                 // Component height
	hasBounds    bool                // Whether bounds are set
	hasDragHandle bool               // Whether drag handle is set
	borderStyle  lipgloss.Style      // Border style for visual feedback
	dragStyle    lipgloss.Style      // Style when dragging
	focused      bool                // Whether component has keyboard focus
}

// NewDraggable creates a new draggable wrapper around a component
func NewDraggable(component Component, x, y int) *DraggableComponent {
	return &DraggableComponent{
		component:     component,
		x:             x,
		y:             y,
		startX:        x,
		startY:        y,
		offsetX:       0,
		offsetY:       0,
		isDragging:    false,
		dragHandle:    layout.Rectangle{X: 0, Y: 0, Width: 0, Height: 0},
		bounds:        layout.Rectangle{X: 0, Y: 0, Width: 0, Height: 0},
		snapToGrid:    false,
		gridSize:      1,
		width:         40,
		height:        10,
		hasBounds:     false,
		hasDragHandle: false,
		borderStyle:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
		dragStyle:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("86")),
		focused:       false,
	}
}

// SetDragHandle sets the area that can be used to drag (relative to component)
func (d *DraggableComponent) SetDragHandle(x, y, width, height int) {
	d.dragHandle = layout.Rectangle{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
	d.hasDragHandle = true
}

// SetBounds sets the constraint boundaries for dragging
func (d *DraggableComponent) SetBounds(x, y, width, height int) {
	d.bounds = layout.Rectangle{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
	d.hasBounds = true
}

// EnableSnapToGrid enables snapping to a grid with the specified size
func (d *DraggableComponent) EnableSnapToGrid(size int) {
	d.snapToGrid = true
	d.gridSize = size
}

// DisableSnapToGrid disables grid snapping
func (d *DraggableComponent) DisableSnapToGrid() {
	d.snapToGrid = false
}

// StartDrag initiates a drag operation
func (d *DraggableComponent) StartDrag(mouseX, mouseY int) {
	// Check if click is on drag handle (if set)
	if d.hasDragHandle {
		handleX := d.x + d.dragHandle.X
		handleY := d.y + d.dragHandle.Y
		if !IsInBoundsXY(mouseX, mouseY, handleX, handleY, d.dragHandle.Width, d.dragHandle.Height) {
			return
		}
	}

	d.isDragging = true
	d.startX = d.x
	d.startY = d.y
	d.offsetX = mouseX - d.x
	d.offsetY = mouseY - d.y
}

// UpdateDrag updates the position during a drag operation
func (d *DraggableComponent) UpdateDrag(mouseX, mouseY int) {
	if !d.isDragging {
		return
	}

	// Calculate new position
	newX := mouseX - d.offsetX
	newY := mouseY - d.offsetY

	// Apply grid snapping
	if d.snapToGrid && d.gridSize > 0 {
		newX = (newX / d.gridSize) * d.gridSize
		newY = (newY / d.gridSize) * d.gridSize
	}

	// Apply boundary constraints
	if d.hasBounds {
		if newX < d.bounds.X {
			newX = d.bounds.X
		}
		if newY < d.bounds.Y {
			newY = d.bounds.Y
		}
		if newX+d.width > d.bounds.X+d.bounds.Width {
			newX = d.bounds.X + d.bounds.Width - d.width
		}
		if newY+d.height > d.bounds.Y+d.bounds.Height {
			newY = d.bounds.Y + d.bounds.Height - d.height
		}
	}

	d.x = newX
	d.y = newY
}

// EndDrag ends the drag operation
func (d *DraggableComponent) EndDrag() {
	d.isDragging = false
}

// Position returns the current position
func (d *DraggableComponent) Position() (x, y int) {
	return d.x, d.y
}

// SetPosition sets the position (with constraints)
func (d *DraggableComponent) SetPosition(x, y int) {
	// Apply grid snapping
	if d.snapToGrid && d.gridSize > 0 {
		x = (x / d.gridSize) * d.gridSize
		y = (y / d.gridSize) * d.gridSize
	}

	// Apply boundary constraints
	if d.hasBounds {
		if x < d.bounds.X {
			x = d.bounds.X
		}
		if y < d.bounds.Y {
			y = d.bounds.Y
		}
		if x+d.width > d.bounds.X+d.bounds.Width {
			x = d.bounds.X + d.bounds.Width - d.width
		}
		if y+d.height > d.bounds.Y+d.bounds.Height {
			y = d.bounds.Y + d.bounds.Height - d.height
		}
	}

	d.x = x
	d.y = y
}

// IsDragging returns whether currently dragging
func (d *DraggableComponent) IsDragging() bool {
	return d.isDragging
}

// SetSize sets the component size (needed for boundary checking)
func (d *DraggableComponent) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// GetSize returns the component size
func (d *DraggableComponent) GetSize() (width, height int) {
	return d.width, d.height
}

// Focus gives keyboard focus to the component
func (d *DraggableComponent) Focus() tea.Cmd {
	d.focused = true
	if focusable, ok := d.component.(Focusable); ok {
		return focusable.Focus()
	}
	return nil
}

// Blur removes keyboard focus
func (d *DraggableComponent) Blur() {
	d.focused = false
	if focusable, ok := d.component.(Focusable); ok {
		focusable.Blur()
	}
}

// Focused returns whether component has focus
func (d *DraggableComponent) Focused() bool {
	return d.focused
}

// SetBorderStyle sets the border style
func (d *DraggableComponent) SetBorderStyle(style lipgloss.Style) {
	d.borderStyle = style
}

// SetDragStyle sets the style used when dragging
func (d *DraggableComponent) SetDragStyle(style lipgloss.Style) {
	d.dragStyle = style
}

// Init initializes the component
func (d *DraggableComponent) Init() tea.Cmd {
	return d.component.Init()
}

// Update handles messages and updates state
func (d *DraggableComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		return d.handleMouseMsg(msg)

	case tea.KeyMsg:
		return d.handleKeyMsg(msg)
	}

	// Forward to wrapped component
	updatedComp, cmd := d.component.Update(msg)
	d.component = updatedComp
	return d, cmd
}

// handleMouseMsg processes mouse events
func (d *DraggableComponent) handleMouseMsg(msg tea.MouseMsg) (Component, tea.Cmd) {
	mouseX, mouseY := msg.X, msg.Y

	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonLeft {
			// Check if click is on the component
			if IsInBoundsXY(mouseX, mouseY, d.x, d.y, d.width, d.height) {
				d.StartDrag(mouseX, mouseY)
				d.Focus()
			}
		}

	case tea.MouseActionMotion:
		if d.isDragging {
			d.UpdateDrag(mouseX, mouseY)
		}

	case tea.MouseActionRelease:
		if msg.Button == tea.MouseButtonLeft {
			d.EndDrag()
		}
	}

	// Forward to wrapped component (adjust coordinates relative to component)
	relativeMsg := msg
	relativeMsg.X = mouseX - d.x
	relativeMsg.Y = mouseY - d.y
	updatedComp, cmd := d.component.Update(relativeMsg)
	d.component = updatedComp

	return d, cmd
}

// handleKeyMsg processes keyboard events for movement
func (d *DraggableComponent) handleKeyMsg(msg tea.KeyMsg) (Component, tea.Cmd) {
	// Only handle keyboard movement when focused and Alt is pressed
	if !d.focused {
		updatedComp, cmd := d.component.Update(msg)
		d.component = updatedComp
		return d, cmd
	}

	// Alt+Arrow keys for movement
	moveSpeed := 1
	if d.snapToGrid {
		moveSpeed = d.gridSize
	}

	handled := false
	switch msg.String() {
	case "alt+up":
		d.SetPosition(d.x, d.y-moveSpeed)
		handled = true
	case "alt+down":
		d.SetPosition(d.x, d.y+moveSpeed)
		handled = true
	case "alt+left":
		d.SetPosition(d.x-moveSpeed, d.y)
		handled = true
	case "alt+right":
		d.SetPosition(d.x+moveSpeed, d.y)
		handled = true
	}

	// If not handled, forward to wrapped component
	if !handled {
		updatedComp, cmd := d.component.Update(msg)
		d.component = updatedComp
		return d, cmd
	}

	return d, nil
}

// View renders the component with positioning
func (d *DraggableComponent) View() string {
	// Get the wrapped component's view
	content := d.component.View()

	// Apply style based on state
	style := d.borderStyle
	if d.isDragging || d.focused {
		style = d.dragStyle
	}

	// Apply border if dragging or focused
	if d.isDragging || d.focused {
		content = style.Render(content)
	}

	// Get content dimensions
	lines := strings.Split(content, "\n")
	if len(lines) > 0 {
		d.height = len(lines)
		d.width = 0
		for _, line := range lines {
			if len(line) > d.width {
				d.width = lipgloss.Width(line)
			}
		}
	}

	// Add padding for positioning (ANSI escape sequences for positioning)
	// For simplicity, we'll use spaces for now
	// In a real implementation, you'd use ANSI cursor positioning
	positioned := make([]string, d.y)
	for i := 0; i < d.y; i++ {
		positioned = append(positioned, "")
	}

	// Add the content with left padding
	for _, line := range lines {
		paddedLine := strings.Repeat(" ", d.x) + line
		positioned = append(positioned, paddedLine)
	}

	return strings.Join(positioned, "\n")
}

// GetComponent returns the wrapped component
func (d *DraggableComponent) GetComponent() Component {
	return d.component
}

// SetComponent sets the wrapped component
func (d *DraggableComponent) SetComponent(component Component) {
	d.component = component
}
