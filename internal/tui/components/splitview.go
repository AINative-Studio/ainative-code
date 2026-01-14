package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/AINative-studio/ainative-code/internal/tui/layout"
)

// SplitView creates a resizable split pane layout
type SplitView struct {
	left          Component        // Left or top pane
	right         Component        // Right or bottom pane
	orientation   layout.Orientation // Horizontal or Vertical
	splitRatio    float64          // Split ratio (0.0 - 1.0)
	minRatio      float64          // Minimum split ratio
	maxRatio      float64          // Maximum split ratio
	dividerSize   int              // Divider width/height (in characters)
	isDragging    bool             // Whether divider is being dragged
	showDivider   bool             // Whether to show divider
	width         int              // Total width
	height        int              // Total height
	dividerStyle  lipgloss.Style   // Style for divider
	dragStyle     lipgloss.Style   // Style when dragging
	dragStartPos  int              // Mouse position when drag started
	dragStartRatio float64         // Split ratio when drag started
	focused       bool             // Whether component has keyboard focus
}

// NewSplitView creates a new split view with two panes
func NewSplitView(left, right Component, orientation layout.Orientation) *SplitView {
	return &SplitView{
		left:          left,
		right:         right,
		orientation:   orientation,
		splitRatio:    0.5,
		minRatio:      0.2,
		maxRatio:      0.8,
		dividerSize:   1,
		isDragging:    false,
		showDivider:   true,
		width:         80,
		height:        24,
		dividerStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		dragStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true),
		dragStartPos:  0,
		dragStartRatio: 0.5,
		focused:       false,
	}
}

// SetSplitRatio sets the split ratio (constrained by min/max)
func (sv *SplitView) SetSplitRatio(ratio float64) {
	if ratio < sv.minRatio {
		ratio = sv.minRatio
	}
	if ratio > sv.maxRatio {
		ratio = sv.maxRatio
	}
	sv.splitRatio = ratio
	sv.updatePaneSizes()
}

// SetMinMaxRatio sets the minimum and maximum split ratios
func (sv *SplitView) SetMinMaxRatio(min, max float64) {
	sv.minRatio = min
	sv.maxRatio = max
	// Re-apply current ratio to ensure it's within bounds
	sv.SetSplitRatio(sv.splitRatio)
}

// SetDividerStyle sets the divider style
func (sv *SplitView) SetDividerStyle(style lipgloss.Style) {
	sv.dividerStyle = style
}

// SetDragStyle sets the style used when dragging
func (sv *SplitView) SetDragStyle(style lipgloss.Style) {
	sv.dragStyle = style
}

// EnableDivider shows the divider
func (sv *SplitView) EnableDivider() {
	sv.showDivider = true
}

// DisableDivider hides the divider
func (sv *SplitView) DisableDivider() {
	sv.showDivider = false
}

// SetDividerSize sets the divider size (width for vertical, height for horizontal)
func (sv *SplitView) SetDividerSize(size int) {
	sv.dividerSize = size
	sv.updatePaneSizes()
}

// StartDividerDrag initiates a divider drag operation
func (sv *SplitView) StartDividerDrag(mouseX, mouseY int) {
	// Check if click is on the divider
	if !sv.isOnDivider(mouseX, mouseY) {
		return
	}

	sv.isDragging = true
	sv.dragStartRatio = sv.splitRatio

	if sv.orientation == layout.Horizontal {
		sv.dragStartPos = mouseX
	} else {
		sv.dragStartPos = mouseY
	}
}

// UpdateDividerDrag updates the split ratio during a drag operation
func (sv *SplitView) UpdateDividerDrag(mouseX, mouseY int) {
	if !sv.isDragging {
		return
	}

	var currentPos int
	var totalSize int

	if sv.orientation == layout.Horizontal {
		currentPos = mouseX
		totalSize = sv.width
	} else {
		currentPos = mouseY
		totalSize = sv.height
	}

	// Calculate new ratio based on mouse position
	newRatio := float64(currentPos) / float64(totalSize)
	sv.SetSplitRatio(newRatio)
}

// EndDividerDrag ends the drag operation
func (sv *SplitView) EndDividerDrag() {
	sv.isDragging = false
}

// GetLeftPane returns the left/top pane component
func (sv *SplitView) GetLeftPane() Component {
	return sv.left
}

// GetRightPane returns the right/bottom pane component
func (sv *SplitView) GetRightPane() Component {
	return sv.right
}

// SetLeftPane sets the left/top pane component
func (sv *SplitView) SetLeftPane(component Component) {
	sv.left = component
	sv.updatePaneSizes()
}

// SetRightPane sets the right/bottom pane component
func (sv *SplitView) SetRightPane(component Component) {
	sv.right = component
	sv.updatePaneSizes()
}

// SwapPanes swaps the left and right panes
func (sv *SplitView) SwapPanes() {
	sv.left, sv.right = sv.right, sv.left
}

// SetSize sets the total size of the split view
func (sv *SplitView) SetSize(width, height int) {
	sv.width = width
	sv.height = height
	sv.updatePaneSizes()
}

// GetSize returns the total size
func (sv *SplitView) GetSize() (width, height int) {
	return sv.width, sv.height
}

// Focus gives keyboard focus to the component
func (sv *SplitView) Focus() tea.Cmd {
	sv.focused = true
	return nil
}

// Blur removes keyboard focus
func (sv *SplitView) Blur() {
	sv.focused = false
}

// Focused returns whether component has focus
func (sv *SplitView) Focused() bool {
	return sv.focused
}

// isOnDivider checks if the mouse is on the divider
func (sv *SplitView) isOnDivider(mouseX, mouseY int) bool {
	if !sv.showDivider {
		return false
	}

	if sv.orientation == layout.Horizontal {
		dividerX := int(float64(sv.width) * sv.splitRatio)
		return mouseX >= dividerX && mouseX < dividerX+sv.dividerSize
	} else {
		dividerY := int(float64(sv.height) * sv.splitRatio)
		return mouseY >= dividerY && mouseY < dividerY+sv.dividerSize
	}
}

// updatePaneSizes updates the sizes of both panes
func (sv *SplitView) updatePaneSizes() {
	if sv.orientation == layout.Horizontal {
		// Horizontal split (left/right)
		leftWidth := int(float64(sv.width) * sv.splitRatio)
		rightWidth := sv.width - leftWidth - sv.dividerSize

		if leftWidth < 0 {
			leftWidth = 0
		}
		if rightWidth < 0 {
			rightWidth = 0
		}

		if leftSizeable, ok := sv.left.(Sizeable); ok {
			leftSizeable.SetSize(leftWidth, sv.height)
		}
		if rightSizeable, ok := sv.right.(Sizeable); ok {
			rightSizeable.SetSize(rightWidth, sv.height)
		}
	} else {
		// Vertical split (top/bottom)
		topHeight := int(float64(sv.height) * sv.splitRatio)
		bottomHeight := sv.height - topHeight - sv.dividerSize

		if topHeight < 0 {
			topHeight = 0
		}
		if bottomHeight < 0 {
			bottomHeight = 0
		}

		if leftSizeable, ok := sv.left.(Sizeable); ok {
			leftSizeable.SetSize(sv.width, topHeight)
		}
		if rightSizeable, ok := sv.right.(Sizeable); ok {
			rightSizeable.SetSize(sv.width, bottomHeight)
		}
	}
}

// Init initializes the component
func (sv *SplitView) Init() tea.Cmd {
	leftCmd := sv.left.Init()
	rightCmd := sv.right.Init()
	return tea.Batch(leftCmd, rightCmd)
}

// Update handles messages and updates state
func (sv *SplitView) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		return sv.handleMouseMsg(msg)

	case tea.KeyMsg:
		return sv.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		sv.SetSize(msg.Width, msg.Height)
	}

	// Forward to both panes
	leftComp, leftCmd := sv.left.Update(msg)
	sv.left = leftComp

	rightComp, rightCmd := sv.right.Update(msg)
	sv.right = rightComp

	return sv, tea.Batch(leftCmd, rightCmd)
}

// handleMouseMsg processes mouse events
func (sv *SplitView) handleMouseMsg(msg tea.MouseMsg) (Component, tea.Cmd) {
	mouseX, mouseY := msg.X, msg.Y

	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonLeft {
			sv.StartDividerDrag(mouseX, mouseY)
			if sv.isDragging {
				sv.Focus()
				return sv, nil
			}
		}

	case tea.MouseActionMotion:
		if sv.isDragging {
			sv.UpdateDividerDrag(mouseX, mouseY)
			return sv, nil
		}

	case tea.MouseActionRelease:
		if msg.Button == tea.MouseButtonLeft {
			sv.EndDividerDrag()
			return sv, nil
		}
	}

	// Forward to panes if not dragging
	if !sv.isDragging {
		var leftCmd, rightCmd tea.Cmd

		if sv.orientation == layout.Horizontal {
			dividerX := int(float64(sv.width) * sv.splitRatio)
			if mouseX < dividerX {
				// Click is in left pane
				leftComp, cmd := sv.left.Update(msg)
				sv.left = leftComp
				leftCmd = cmd
			} else if mouseX >= dividerX+sv.dividerSize {
				// Click is in right pane
				// Adjust coordinates relative to right pane
				relativeMsg := msg
				relativeMsg.X = mouseX - dividerX - sv.dividerSize
				rightComp, cmd := sv.right.Update(relativeMsg)
				sv.right = rightComp
				rightCmd = cmd
			}
		} else {
			dividerY := int(float64(sv.height) * sv.splitRatio)
			if mouseY < dividerY {
				// Click is in top pane
				topComp, cmd := sv.left.Update(msg)
				sv.left = topComp
				leftCmd = cmd
			} else if mouseY >= dividerY+sv.dividerSize {
				// Click is in bottom pane
				// Adjust coordinates relative to bottom pane
				relativeMsg := msg
				relativeMsg.Y = mouseY - dividerY - sv.dividerSize
				bottomComp, cmd := sv.right.Update(relativeMsg)
				sv.right = bottomComp
				rightCmd = cmd
			}
		}

		return sv, tea.Batch(leftCmd, rightCmd)
	}

	return sv, nil
}

// handleKeyMsg processes keyboard events
func (sv *SplitView) handleKeyMsg(msg tea.KeyMsg) (Component, tea.Cmd) {
	// Handle split adjustment when focused
	if sv.focused {
		adjustSpeed := 0.05

		handled := false
		switch msg.String() {
		case "ctrl+,":
			// Decrease split ratio (move divider left/up)
			sv.SetSplitRatio(sv.splitRatio - adjustSpeed)
			handled = true
		case "ctrl+.":
			// Increase split ratio (move divider right/down)
			sv.SetSplitRatio(sv.splitRatio + adjustSpeed)
			handled = true
		case "ctrl+s":
			// Swap panes
			sv.SwapPanes()
			handled = true
		case "ctrl+=", "ctrl+0":
			// Reset to 50/50
			sv.SetSplitRatio(0.5)
			handled = true
		}

		if handled {
			return sv, nil
		}
	}

	// Forward to both panes
	leftComp, leftCmd := sv.left.Update(msg)
	sv.left = leftComp

	rightComp, rightCmd := sv.right.Update(msg)
	sv.right = rightComp

	return sv, tea.Batch(leftCmd, rightCmd)
}

// View renders the split view
func (sv *SplitView) View() string {
	leftView := sv.left.View()
	rightView := sv.right.View()

	if sv.orientation == layout.Horizontal {
		return sv.renderHorizontal(leftView, rightView)
	} else {
		return sv.renderVertical(leftView, rightView)
	}
}

// renderHorizontal renders horizontal split (left/right)
func (sv *SplitView) renderHorizontal(leftView, rightView string) string {
	leftWidth := int(float64(sv.width) * sv.splitRatio)
	rightWidth := sv.width - leftWidth - sv.dividerSize

	leftLines := strings.Split(leftView, "\n")
	rightLines := strings.Split(rightView, "\n")

	// Ensure we have enough lines
	maxLines := sv.height
	if len(leftLines) > maxLines {
		leftLines = leftLines[:maxLines]
	}
	if len(rightLines) > maxLines {
		rightLines = rightLines[:maxLines]
	}

	// Build the output
	result := make([]string, 0, maxLines)

	for i := 0; i < maxLines; i++ {
		var leftLine, rightLine string

		if i < len(leftLines) {
			leftLine = leftLines[i]
		}
		if i < len(rightLines) {
			rightLine = rightLines[i]
		}

		// Truncate or pad left line
		if lipgloss.Width(leftLine) > leftWidth {
			leftLine = leftLine[:leftWidth]
		} else {
			leftLine = leftLine + strings.Repeat(" ", leftWidth-lipgloss.Width(leftLine))
		}

		// Truncate or pad right line
		if lipgloss.Width(rightLine) > rightWidth {
			rightLine = rightLine[:rightWidth]
		} else {
			rightLine = rightLine + strings.Repeat(" ", rightWidth-lipgloss.Width(rightLine))
		}

		// Create divider
		var divider string
		if sv.showDivider {
			style := sv.dividerStyle
			if sv.isDragging {
				style = sv.dragStyle
			}
			divider = style.Render("│")
			for j := 1; j < sv.dividerSize; j++ {
				divider += style.Render("│")
			}
		}

		// Combine left, divider, and right
		line := leftLine + divider + rightLine
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// renderVertical renders vertical split (top/bottom)
func (sv *SplitView) renderVertical(topView, bottomView string) string {
	topHeight := int(float64(sv.height) * sv.splitRatio)
	bottomHeight := sv.height - topHeight - sv.dividerSize

	topLines := strings.Split(topView, "\n")
	bottomLines := strings.Split(bottomView, "\n")

	result := make([]string, 0, sv.height)

	// Add top pane lines
	for i := 0; i < topHeight; i++ {
		var line string
		if i < len(topLines) {
			line = topLines[i]
		}
		// Pad to width
		if lipgloss.Width(line) < sv.width {
			line = line + strings.Repeat(" ", sv.width-lipgloss.Width(line))
		} else if lipgloss.Width(line) > sv.width {
			line = line[:sv.width]
		}
		result = append(result, line)
	}

	// Add divider
	if sv.showDivider {
		for i := 0; i < sv.dividerSize; i++ {
			style := sv.dividerStyle
			if sv.isDragging {
				style = sv.dragStyle
			}
			divider := style.Render(strings.Repeat("─", sv.width))
			result = append(result, divider)
		}
	}

	// Add bottom pane lines
	for i := 0; i < bottomHeight; i++ {
		var line string
		if i < len(bottomLines) {
			line = bottomLines[i]
		}
		// Pad to width
		if lipgloss.Width(line) < sv.width {
			line = line + strings.Repeat(" ", sv.width-lipgloss.Width(line))
		} else if lipgloss.Width(line) > sv.width {
			line = line[:sv.width]
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
