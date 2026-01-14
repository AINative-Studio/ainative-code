package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ResizePosition defines the position of a resize handle
type ResizePosition int

const (
	TopLeft ResizePosition = iota
	Top
	TopRight
	Right
	BottomRight
	Bottom
	BottomLeft
	Left
)

// String returns the string representation of the resize position
func (rp ResizePosition) String() string {
	switch rp {
	case TopLeft:
		return "TopLeft"
	case Top:
		return "Top"
	case TopRight:
		return "TopRight"
	case Right:
		return "Right"
	case BottomRight:
		return "BottomRight"
	case Bottom:
		return "Bottom"
	case BottomLeft:
		return "BottomLeft"
	case Left:
		return "Left"
	default:
		return "Unknown"
	}
}

// ResizeHandle defines a resize grab area
type ResizeHandle struct {
	Position ResizePosition
	X        int    // X position relative to component
	Y        int    // Y position relative to component
	Width    int    // Handle width
	Height   int    // Handle height
	Cursor   string // Character to display (e.g., "◢")
}

// ResizableComponent wraps a component and makes it resizable
type ResizableComponent struct {
	component       Component        // The wrapped component
	x               int              // X position
	y               int              // Y position
	width           int              // Current width
	height          int              // Current height
	minWidth        int              // Minimum width
	minHeight       int              // Minimum height
	maxWidth        int              // Maximum width (0 = unlimited)
	maxHeight       int              // Maximum height (0 = unlimited)
	resizeHandles   []ResizeHandle   // Resize handles
	activeHandle    *ResizeHandle    // Currently active handle
	isResizing      bool             // Whether currently resizing
	preserveAspect  bool             // Whether to preserve aspect ratio
	aspectRatio     float64          // Aspect ratio (width/height)
	startWidth      int              // Width when resize started
	startHeight     int              // Height when resize started
	startMouseX     int              // Mouse X when resize started
	startMouseY     int              // Mouse Y when resize started
	borderStyle     lipgloss.Style   // Border style
	resizeStyle     lipgloss.Style   // Style when resizing
	handleStyle     lipgloss.Style   // Style for resize handles
	focused         bool             // Whether component has keyboard focus
}

// NewResizable creates a new resizable wrapper around a component
func NewResizable(component Component, width, height int) *ResizableComponent {
	r := &ResizableComponent{
		component:       component,
		x:               0,
		y:               0,
		width:           width,
		height:          height,
		minWidth:        10,
		minHeight:       3,
		maxWidth:        0,
		maxHeight:       0,
		resizeHandles:   []ResizeHandle{},
		activeHandle:    nil,
		isResizing:      false,
		preserveAspect:  false,
		aspectRatio:     float64(width) / float64(height),
		startWidth:      width,
		startHeight:     height,
		startMouseX:     0,
		startMouseY:     0,
		borderStyle:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
		resizeStyle:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("86")),
		handleStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("86")),
		focused:         false,
	}

	// Add default resize handles at all positions
	r.AddAllResizeHandles()

	return r
}

// SetMinSize sets the minimum size constraints
func (r *ResizableComponent) SetMinSize(width, height int) {
	r.minWidth = width
	r.minHeight = height
}

// SetMaxSize sets the maximum size constraints (0 = unlimited)
func (r *ResizableComponent) SetMaxSize(width, height int) {
	r.maxWidth = width
	r.maxHeight = height
}

// EnableAspectRatio enables aspect ratio preservation
func (r *ResizableComponent) EnableAspectRatio(ratio float64) {
	r.preserveAspect = true
	r.aspectRatio = ratio
}

// DisableAspectRatio disables aspect ratio preservation
func (r *ResizableComponent) DisableAspectRatio() {
	r.preserveAspect = false
}

// AddResizeHandle adds a resize handle at the specified position
func (r *ResizableComponent) AddResizeHandle(pos ResizePosition) {
	handle := ResizeHandle{
		Position: pos,
		Width:    1,
		Height:   1,
		Cursor:   r.getCursorForPosition(pos),
	}
	r.resizeHandles = append(r.resizeHandles, handle)
	r.updateHandlePositions()
}

// AddAllResizeHandles adds resize handles at all 8 positions
func (r *ResizableComponent) AddAllResizeHandles() {
	r.resizeHandles = []ResizeHandle{}
	for pos := TopLeft; pos <= Left; pos++ {
		r.AddResizeHandle(pos)
	}
}

// getCursorForPosition returns the appropriate cursor character for a position
func (r *ResizableComponent) getCursorForPosition(pos ResizePosition) string {
	switch pos {
	case TopLeft, BottomRight:
		return "◢"
	case TopRight, BottomLeft:
		return "◣"
	case Top, Bottom:
		return "│"
	case Left, Right:
		return "─"
	default:
		return "+"
	}
}

// updateHandlePositions updates the positions of resize handles
func (r *ResizableComponent) updateHandlePositions() {
	for i := range r.resizeHandles {
		handle := &r.resizeHandles[i]
		switch handle.Position {
		case TopLeft:
			handle.X = 0
			handle.Y = 0
		case Top:
			handle.X = r.width / 2
			handle.Y = 0
		case TopRight:
			handle.X = r.width - 1
			handle.Y = 0
		case Right:
			handle.X = r.width - 1
			handle.Y = r.height / 2
		case BottomRight:
			handle.X = r.width - 1
			handle.Y = r.height - 1
		case Bottom:
			handle.X = r.width / 2
			handle.Y = r.height - 1
		case BottomLeft:
			handle.X = 0
			handle.Y = r.height - 1
		case Left:
			handle.X = 0
			handle.Y = r.height / 2
		}
	}
}

// StartResize initiates a resize operation
func (r *ResizableComponent) StartResize(mouseX, mouseY int) {
	// Find which handle was clicked
	for i := range r.resizeHandles {
		handle := &r.resizeHandles[i]
		handleX := r.x + handle.X
		handleY := r.y + handle.Y

		// Check if mouse is on this handle (with some tolerance)
		if mouseX >= handleX-1 && mouseX <= handleX+1 &&
			mouseY >= handleY-1 && mouseY <= handleY+1 {
			r.isResizing = true
			r.activeHandle = handle
			r.startWidth = r.width
			r.startHeight = r.height
			r.startMouseX = mouseX
			r.startMouseY = mouseY
			return
		}
	}
}

// UpdateResize updates the size during a resize operation
func (r *ResizableComponent) UpdateResize(mouseX, mouseY int) {
	if !r.isResizing || r.activeHandle == nil {
		return
	}

	deltaX := mouseX - r.startMouseX
	deltaY := mouseY - r.startMouseY

	newWidth := r.startWidth
	newHeight := r.startHeight

	// Calculate new size based on handle position
	switch r.activeHandle.Position {
	case TopLeft:
		newWidth = r.startWidth - deltaX
		newHeight = r.startHeight - deltaY
	case Top:
		newHeight = r.startHeight - deltaY
	case TopRight:
		newWidth = r.startWidth + deltaX
		newHeight = r.startHeight - deltaY
	case Right:
		newWidth = r.startWidth + deltaX
	case BottomRight:
		newWidth = r.startWidth + deltaX
		newHeight = r.startHeight + deltaY
	case Bottom:
		newHeight = r.startHeight + deltaY
	case BottomLeft:
		newWidth = r.startWidth - deltaX
		newHeight = r.startHeight + deltaY
	case Left:
		newWidth = r.startWidth - deltaX
	}

	// Apply aspect ratio if enabled
	if r.preserveAspect && r.aspectRatio > 0 {
		// Preserve aspect ratio based on which dimension changed more
		if absInt(deltaX) > absInt(deltaY) {
			newHeight = int(float64(newWidth) / r.aspectRatio)
		} else {
			newWidth = int(float64(newHeight) * r.aspectRatio)
		}
	}

	// Apply constraints
	if newWidth < r.minWidth {
		newWidth = r.minWidth
	}
	if newHeight < r.minHeight {
		newHeight = r.minHeight
	}
	if r.maxWidth > 0 && newWidth > r.maxWidth {
		newWidth = r.maxWidth
	}
	if r.maxHeight > 0 && newHeight > r.maxHeight {
		newHeight = r.maxHeight
	}

	r.width = newWidth
	r.height = newHeight

	// Update component size if it implements Sizeable
	if sizeable, ok := r.component.(Sizeable); ok {
		sizeable.SetSize(r.width, r.height)
	}

	// Update handle positions
	r.updateHandlePositions()
}

// EndResize ends the resize operation
func (r *ResizableComponent) EndResize() {
	r.isResizing = false
	r.activeHandle = nil
}

// Size returns the current size
func (r *ResizableComponent) Size() (width, height int) {
	return r.width, r.height
}

// SetSize sets the size (with constraints)
func (r *ResizableComponent) SetSize(width, height int) {
	// Apply constraints
	if width < r.minWidth {
		width = r.minWidth
	}
	if height < r.minHeight {
		height = r.minHeight
	}
	if r.maxWidth > 0 && width > r.maxWidth {
		width = r.maxWidth
	}
	if r.maxHeight > 0 && height > r.maxHeight {
		height = r.maxHeight
	}

	r.width = width
	r.height = height

	// Update component size if it implements Sizeable
	if sizeable, ok := r.component.(Sizeable); ok {
		sizeable.SetSize(r.width, r.height)
	}

	r.updateHandlePositions()
}

// GetSize returns the current size
func (r *ResizableComponent) GetSize() (width, height int) {
	return r.width, r.height
}

// IsResizing returns whether currently resizing
func (r *ResizableComponent) IsResizing() bool {
	return r.isResizing
}

// SetPosition sets the position
func (r *ResizableComponent) SetPosition(x, y int) {
	r.x = x
	r.y = y
}

// Position returns the position
func (r *ResizableComponent) Position() (x, y int) {
	return r.x, r.y
}

// Focus gives keyboard focus to the component
func (r *ResizableComponent) Focus() tea.Cmd {
	r.focused = true
	if focusable, ok := r.component.(Focusable); ok {
		return focusable.Focus()
	}
	return nil
}

// Blur removes keyboard focus
func (r *ResizableComponent) Blur() {
	r.focused = false
	if focusable, ok := r.component.(Focusable); ok {
		focusable.Blur()
	}
}

// Focused returns whether component has focus
func (r *ResizableComponent) Focused() bool {
	return r.focused
}

// SetBorderStyle sets the border style
func (r *ResizableComponent) SetBorderStyle(style lipgloss.Style) {
	r.borderStyle = style
}

// SetResizeStyle sets the style used when resizing
func (r *ResizableComponent) SetResizeStyle(style lipgloss.Style) {
	r.resizeStyle = style
}

// Init initializes the component
func (r *ResizableComponent) Init() tea.Cmd {
	return r.component.Init()
}

// Update handles messages and updates state
func (r *ResizableComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		return r.handleMouseMsg(msg)

	case tea.KeyMsg:
		return r.handleKeyMsg(msg)
	}

	// Forward to wrapped component
	updatedComp, cmd := r.component.Update(msg)
	r.component = updatedComp
	return r, cmd
}

// handleMouseMsg processes mouse events
func (r *ResizableComponent) handleMouseMsg(msg tea.MouseMsg) (Component, tea.Cmd) {
	mouseX, mouseY := msg.X, msg.Y

	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonLeft {
			r.StartResize(mouseX, mouseY)
			if r.isResizing {
				r.Focus()
			}
		}

	case tea.MouseActionMotion:
		if r.isResizing {
			r.UpdateResize(mouseX, mouseY)
		}

	case tea.MouseActionRelease:
		if msg.Button == tea.MouseButtonLeft {
			r.EndResize()
		}
	}

	// Forward to wrapped component if not resizing
	if !r.isResizing {
		updatedComp, cmd := r.component.Update(msg)
		r.component = updatedComp
		return r, cmd
	}

	return r, nil
}

// handleKeyMsg processes keyboard events for resizing
func (r *ResizableComponent) handleKeyMsg(msg tea.KeyMsg) (Component, tea.Cmd) {
	// Only handle keyboard resize when focused and Ctrl is pressed
	if !r.focused {
		updatedComp, cmd := r.component.Update(msg)
		r.component = updatedComp
		return r, cmd
	}

	// Ctrl+Arrow keys for resizing
	resizeSpeed := 1

	handled := false
	switch msg.String() {
	case "ctrl+up":
		r.SetSize(r.width, r.height-resizeSpeed)
		handled = true
	case "ctrl+down":
		r.SetSize(r.width, r.height+resizeSpeed)
		handled = true
	case "ctrl+left":
		r.SetSize(r.width-resizeSpeed, r.height)
		handled = true
	case "ctrl+right":
		r.SetSize(r.width+resizeSpeed, r.height)
		handled = true
	}

	// If not handled, forward to wrapped component
	if !handled {
		updatedComp, cmd := r.component.Update(msg)
		r.component = updatedComp
		return r, cmd
	}

	return r, nil
}

// View renders the component with resize handles
func (r *ResizableComponent) View() string {
	// Get the wrapped component's view
	content := r.component.View()

	// Apply style based on state
	style := r.borderStyle
	if r.isResizing || r.focused {
		style = r.resizeStyle
	}

	// Resize content to fit dimensions
	lines := strings.Split(content, "\n")
	resizedLines := make([]string, 0, r.height)

	for i := 0; i < r.height; i++ {
		var line string
		if i < len(lines) {
			line = lines[i]
		} else {
			line = ""
		}

		// Truncate or pad to width
		if lipgloss.Width(line) > r.width {
			line = line[:r.width]
		} else if lipgloss.Width(line) < r.width {
			line = line + strings.Repeat(" ", r.width-lipgloss.Width(line))
		}

		resizedLines = append(resizedLines, line)
	}

	content = strings.Join(resizedLines, "\n")

	// Apply border
	content = style.Width(r.width).Height(r.height).Render(content)

	// Add resize handles if focused or resizing
	if r.focused || r.isResizing {
		content = r.addResizeHandles(content)
	}

	return content
}

// addResizeHandles adds visual resize handles to the content
func (r *ResizableComponent) addResizeHandles(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return content
	}

	// Add handles at corner and edge positions
	for _, handle := range r.resizeHandles {
		if handle.Y >= 0 && handle.Y < len(lines) {
			line := lines[handle.Y]
			if handle.X >= 0 && handle.X < len(line) {
				// Replace character at handle position with cursor
				runes := []rune(line)
				if handle.X < len(runes) {
					runes[handle.X] = []rune(r.handleStyle.Render(handle.Cursor))[0]
					lines[handle.Y] = string(runes)
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// GetComponent returns the wrapped component
func (r *ResizableComponent) GetComponent() Component {
	return r.component
}

// SetComponent sets the wrapped component
func (r *ResizableComponent) SetComponent(component Component) {
	r.component = component
}

// absInt returns the absolute value of an integer
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
