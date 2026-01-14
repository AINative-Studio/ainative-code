package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/AINative-studio/ainative-code/internal/tui/layout"
)

// MouseState tracks the current mouse state for drag and click operations
type MouseState struct {
	X          int  // Current X position
	Y          int  // Current Y position
	Button     int  // Currently pressed button
	IsPressed  bool // Whether any button is pressed
	IsDragging bool // Whether a drag operation is in progress
	DragStartX int  // X position where drag started
	DragStartY int  // Y position where drag started
}

// NewMouseState creates a new mouse state tracker
func NewMouseState() *MouseState {
	return &MouseState{
		X:          -1,
		Y:          -1,
		Button:     -1,
		IsPressed:  false,
		IsDragging: false,
		DragStartX: 0,
		DragStartY: 0,
	}
}

// Update updates the mouse state based on a tea.MouseMsg
func (ms *MouseState) Update(msg tea.MouseMsg) {
	ms.X = msg.X
	ms.Y = msg.Y

	switch msg.Action {
	case tea.MouseActionPress:
		ms.IsPressed = true
		ms.Button = int(msg.Button)
		ms.DragStartX = msg.X
		ms.DragStartY = msg.Y
		ms.IsDragging = false

	case tea.MouseActionRelease:
		ms.IsPressed = false
		ms.IsDragging = false
		ms.Button = -1

	case tea.MouseActionMotion:
		// If button is pressed and mouse moved, it's a drag
		if ms.IsPressed {
			ms.IsDragging = true
		}
	}
}

// DragDistance returns the distance dragged from the start position
func (ms *MouseState) DragDistance() (dx, dy int) {
	if !ms.IsDragging {
		return 0, 0
	}
	return ms.X - ms.DragStartX, ms.Y - ms.DragStartY
}

// IsInBounds checks if the mouse position is within the specified rectangle
func IsInBounds(mouseX, mouseY int, rect layout.Rectangle) bool {
	return mouseX >= rect.X && mouseX < rect.X+rect.Width &&
		mouseY >= rect.Y && mouseY < rect.Y+rect.Height
}

// IsInBoundsXY is a convenience function for checking bounds with separate coordinates
func IsInBoundsXY(mouseX, mouseY, x, y, width, height int) bool {
	return mouseX >= x && mouseX < x+width &&
		mouseY >= y && mouseY < y+height
}

// GetMousePosition extracts the X and Y coordinates from a tea.MouseMsg
func GetMousePosition(msg tea.MouseMsg) (x, y int) {
	return msg.X, msg.Y
}

// IsLeftClick checks if the message is a left mouse button press
func IsLeftClick(msg tea.MouseMsg) bool {
	return msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress
}

// IsLeftRelease checks if the message is a left mouse button release
func IsLeftRelease(msg tea.MouseMsg) bool {
	return msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionRelease
}

// IsRightClick checks if the message is a right mouse button press
func IsRightClick(msg tea.MouseMsg) bool {
	return msg.Button == tea.MouseButtonRight && msg.Action == tea.MouseActionPress
}

// IsMiddleClick checks if the message is a middle mouse button press
func IsMiddleClick(msg tea.MouseMsg) bool {
	return msg.Button == tea.MouseButtonMiddle && msg.Action == tea.MouseActionPress
}

// IsMouseMotion checks if the message is a mouse motion event
func IsMouseMotion(msg tea.MouseMsg) bool {
	return msg.Action == tea.MouseActionMotion
}

// IsScrollUp checks if the message is a scroll up event
func IsScrollUp(msg tea.MouseMsg) bool {
	return msg.Button == tea.MouseButtonWheelUp && msg.Action == tea.MouseActionPress
}

// IsScrollDown checks if the message is a scroll down event
func IsScrollDown(msg tea.MouseMsg) bool {
	return msg.Button == tea.MouseButtonWheelDown && msg.Action == tea.MouseActionPress
}

// MouseEventHandler is a function type for handling mouse events
type MouseEventHandler func(msg tea.MouseMsg) tea.Cmd

// DragHandler tracks drag operations
type DragHandler struct {
	IsActive   bool
	StartX     int
	StartY     int
	CurrentX   int
	CurrentY   int
	OnDragStart MouseEventHandler
	OnDrag     MouseEventHandler
	OnDragEnd  MouseEventHandler
}

// NewDragHandler creates a new drag handler with the specified callbacks
func NewDragHandler(onStart, onDrag, onEnd MouseEventHandler) *DragHandler {
	return &DragHandler{
		IsActive:    false,
		OnDragStart: onStart,
		OnDrag:      onDrag,
		OnDragEnd:   onEnd,
	}
}

// HandleMouseMsg processes mouse messages for drag operations
func (dh *DragHandler) HandleMouseMsg(msg tea.MouseMsg) tea.Cmd {
	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonLeft {
			dh.IsActive = true
			dh.StartX = msg.X
			dh.StartY = msg.Y
			dh.CurrentX = msg.X
			dh.CurrentY = msg.Y
			if dh.OnDragStart != nil {
				return dh.OnDragStart(msg)
			}
		}

	case tea.MouseActionMotion:
		if dh.IsActive {
			dh.CurrentX = msg.X
			dh.CurrentY = msg.Y
			if dh.OnDrag != nil {
				return dh.OnDrag(msg)
			}
		}

	case tea.MouseActionRelease:
		if dh.IsActive && msg.Button == tea.MouseButtonLeft {
			dh.IsActive = false
			if dh.OnDragEnd != nil {
				return dh.OnDragEnd(msg)
			}
		}
	}

	return nil
}

// DragDelta returns the current drag delta from the start position
func (dh *DragHandler) DragDelta() (dx, dy int) {
	if !dh.IsActive {
		return 0, 0
	}
	return dh.CurrentX - dh.StartX, dh.CurrentY - dh.StartY
}

// Reset resets the drag handler state
func (dh *DragHandler) Reset() {
	dh.IsActive = false
	dh.StartX = 0
	dh.StartY = 0
	dh.CurrentX = 0
	dh.CurrentY = 0
}

// HoverDetector detects mouse hover over areas
type HoverDetector struct {
	bounds   layout.Rectangle
	isHover  bool
	onEnter  MouseEventHandler
	onLeave  MouseEventHandler
	onHover  MouseEventHandler
}

// NewHoverDetector creates a new hover detector
func NewHoverDetector(bounds layout.Rectangle, onEnter, onLeave, onHover MouseEventHandler) *HoverDetector {
	return &HoverDetector{
		bounds:  bounds,
		isHover: false,
		onEnter: onEnter,
		onLeave: onLeave,
		onHover: onHover,
	}
}

// Update updates the hover state
func (hd *HoverDetector) Update(msg tea.MouseMsg) tea.Cmd {
	wasHover := hd.isHover
	hd.isHover = IsInBounds(msg.X, msg.Y, hd.bounds)

	// Detect enter/leave transitions
	if hd.isHover && !wasHover {
		// Mouse entered
		if hd.onEnter != nil {
			return hd.onEnter(msg)
		}
	} else if !hd.isHover && wasHover {
		// Mouse left
		if hd.onLeave != nil {
			return hd.onLeave(msg)
		}
	} else if hd.isHover {
		// Mouse is hovering
		if hd.onHover != nil {
			return hd.onHover(msg)
		}
	}

	return nil
}

// IsHovering returns whether the mouse is currently hovering
func (hd *HoverDetector) IsHovering() bool {
	return hd.isHover
}

// SetBounds updates the hover detection bounds
func (hd *HoverDetector) SetBounds(bounds layout.Rectangle) {
	hd.bounds = bounds
}

// ClickDetector detects click events within a bounds
type ClickDetector struct {
	bounds        layout.Rectangle
	onClick       MouseEventHandler
	onDoubleClick MouseEventHandler
	lastClickTime int64
}

// NewClickDetector creates a new click detector
func NewClickDetector(bounds layout.Rectangle, onClick, onDoubleClick MouseEventHandler) *ClickDetector {
	return &ClickDetector{
		bounds:        bounds,
		onClick:       onClick,
		onDoubleClick: onDoubleClick,
		lastClickTime: 0,
	}
}

// HandleClick processes click events
func (cd *ClickDetector) HandleClick(msg tea.MouseMsg) tea.Cmd {
	if !IsLeftClick(msg) {
		return nil
	}

	if !IsInBounds(msg.X, msg.Y, cd.bounds) {
		return nil
	}

	// Simple click detection (double-click detection would need timestamp comparison)
	if cd.onClick != nil {
		return cd.onClick(msg)
	}

	return nil
}

// SetBounds updates the click detection bounds
func (cd *ClickDetector) SetBounds(bounds layout.Rectangle) {
	cd.bounds = bounds
}
