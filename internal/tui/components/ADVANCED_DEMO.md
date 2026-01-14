# Advanced Components Demo

This document provides comprehensive documentation for the advanced interactive components: Draggable, Resizable, SplitView, and MultiColumn layouts.

## Table of Contents

1. [DraggableComponent](#draggablecomponent)
2. [ResizableComponent](#resizablecomponent)
3. [SplitView](#splitview)
4. [MultiColumnLayout](#multicolumnlayout)
5. [Mouse Event Helpers](#mouse-event-helpers)
6. [Interactive Examples](#interactive-examples)

---

## DraggableComponent

Makes any component draggable with mouse and keyboard support.

### Features

- **Mouse Dragging**: Click and drag to move component
- **Keyboard Movement**: Use `Alt+Arrow` keys for precise positioning
- **Drag Handle**: Optional drag handle (e.g., title bar only)
- **Boundary Constraints**: Limit dragging within specific bounds
- **Snap to Grid**: Snap positions to grid points
- **Visual Feedback**: Border highlights when dragging

### Usage Example

```go
import (
    "github.com/aidevelopers/ainative/internal/tui/components"
)

// Create a draggable component
myComponent := components.NewTextComponent("Hello, World!")
draggable := components.NewDraggable(myComponent, 10, 5)

// Set drag handle (only title bar can be dragged)
draggable.SetDragHandle(0, 0, 40, 1)

// Set boundary constraints
draggable.SetBounds(0, 0, 80, 24)

// Enable snap to grid
draggable.EnableSnapToGrid(5)

// Set visual styles
borderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
draggable.SetBorderStyle(borderStyle)
```

### Mouse Interactions

| Action | Description |
|--------|-------------|
| **Click + Drag** | Move component (or drag handle if set) |
| **Release** | Stop dragging |

### Keyboard Shortcuts

| Key | Description |
|-----|-------------|
| `Alt+↑` | Move up by 1 unit (or grid size) |
| `Alt+↓` | Move down by 1 unit (or grid size) |
| `Alt+←` | Move left by 1 unit (or grid size) |
| `Alt+→` | Move right by 1 unit (or grid size) |

### API Reference

```go
// Create draggable wrapper
func NewDraggable(component Component, x, y int) *DraggableComponent

// Set drag handle area (relative to component)
func (d *DraggableComponent) SetDragHandle(x, y, width, height int)

// Set boundary constraints
func (d *DraggableComponent) SetBounds(x, y, width, height int)

// Enable grid snapping
func (d *DraggableComponent) EnableSnapToGrid(size int)

// Disable grid snapping
func (d *DraggableComponent) DisableSnapToGrid()

// Get current position
func (d *DraggableComponent) Position() (x, y int)

// Set position (with constraints applied)
func (d *DraggableComponent) SetPosition(x, y int)

// Check if currently dragging
func (d *DraggableComponent) IsDragging() bool

// Set component size (needed for boundary checking)
func (d *DraggableComponent) SetSize(width, height int)
```

---

## ResizableComponent

Makes any component resizable with 8 resize handles and keyboard support.

### Features

- **8 Resize Handles**: Corners (◢ ◣) and edges (│ ─)
- **Mouse Resizing**: Click and drag any handle to resize
- **Keyboard Resizing**: Use `Ctrl+Arrow` keys for precise resizing
- **Size Constraints**: Set minimum and maximum sizes
- **Aspect Ratio**: Optionally preserve aspect ratio
- **Visual Handles**: Shows resize cursors when focused

### Usage Example

```go
import (
    "github.com/aidevelopers/ainative/internal/tui/components"
)

// Create a resizable component
myComponent := components.NewTextComponent("Resizable Content")
resizable := components.NewResizable(myComponent, 40, 20)

// Set minimum size
resizable.SetMinSize(20, 10)

// Set maximum size
resizable.SetMaxSize(100, 50)

// Enable aspect ratio preservation
resizable.EnableAspectRatio(2.0) // width:height = 2:1

// Add all 8 resize handles
resizable.AddAllResizeHandles()
```

### Resize Handles

```
◢─────────◣
│         │
│ Content │
│         │
◣─────────◢
```

| Position | Symbol | Description |
|----------|--------|-------------|
| **Top-Left** | ◢ | Resize from top-left corner |
| **Top** | │ | Resize from top edge |
| **Top-Right** | ◣ | Resize from top-right corner |
| **Right** | ─ | Resize from right edge |
| **Bottom-Right** | ◢ | Resize from bottom-right corner |
| **Bottom** | │ | Resize from bottom edge |
| **Bottom-Left** | ◣ | Resize from bottom-left corner |
| **Left** | ─ | Resize from left edge |

### Mouse Interactions

| Action | Description |
|--------|-------------|
| **Click Handle + Drag** | Resize component |
| **Release** | Stop resizing |

### Keyboard Shortcuts

| Key | Description |
|-----|-------------|
| `Ctrl+↑` | Decrease height by 1 |
| `Ctrl+↓` | Increase height by 1 |
| `Ctrl+←` | Decrease width by 1 |
| `Ctrl+→` | Increase width by 1 |

### API Reference

```go
// Create resizable wrapper
func NewResizable(component Component, width, height int) *ResizableComponent

// Set minimum size constraints
func (r *ResizableComponent) SetMinSize(width, height int)

// Set maximum size constraints (0 = unlimited)
func (r *ResizableComponent) SetMaxSize(width, height int)

// Enable aspect ratio preservation
func (r *ResizableComponent) EnableAspectRatio(ratio float64)

// Disable aspect ratio preservation
func (r *ResizableComponent) DisableAspectRatio()

// Add resize handle at specific position
func (r *ResizableComponent) AddResizeHandle(pos ResizePosition)

// Add all 8 resize handles
func (r *ResizableComponent) AddAllResizeHandles()

// Get current size
func (r *ResizableComponent) Size() (width, height int)

// Set size (with constraints applied)
func (r *ResizableComponent) SetSize(width, height int)

// Check if currently resizing
func (r *ResizableComponent) IsResizing() bool
```

---

## SplitView

Creates resizable split pane layouts (horizontal or vertical).

### Features

- **Horizontal/Vertical Split**: Side-by-side or top-bottom layout
- **Draggable Divider**: Click and drag to adjust split ratio
- **Keyboard Adjustment**: Use `Ctrl+,` and `Ctrl+.` to adjust
- **Ratio Constraints**: Set minimum and maximum split ratios
- **Swap Panes**: Swap left/right or top/bottom panes
- **Customizable Divider**: Styled divider with visual feedback

### Usage Example

```go
import (
    "github.com/aidevelopers/ainative/internal/tui/components"
    "github.com/aidevelopers/ainative/internal/tui/layout"
)

// Create two panes
leftPane := components.NewTextComponent("Left Pane")
rightPane := components.NewTextComponent("Right Pane")

// Create horizontal split view (left/right)
split := components.NewSplitView(leftPane, rightPane, layout.Horizontal)
split.SetSize(100, 50)

// Set split ratio (0.0 - 1.0)
split.SetSplitRatio(0.6) // 60% left, 40% right

// Set ratio constraints
split.SetMinMaxRatio(0.2, 0.8) // Between 20% and 80%

// Customize divider
dividerStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("240")).
    Bold(true)
split.SetDividerStyle(dividerStyle)

// For vertical split (top/bottom)
verticalSplit := components.NewSplitView(topPane, bottomPane, layout.Vertical)
```

### Visual Layout

**Horizontal Split:**
```
┌──────────────┬───────────┐
│              │           │
│  Left Pane   │   Right   │
│              │   Pane    │
└──────────────┴───────────┘
       60%    │     40%
```

**Vertical Split:**
```
┌─────────────────────────┐
│      Top Pane           │
│         40%             │
├─────────────────────────┤
│      Bottom Pane        │
│         60%             │
└─────────────────────────┘
```

### Mouse Interactions

| Action | Description |
|--------|-------------|
| **Click Divider + Drag** | Adjust split ratio |
| **Release** | Stop dragging |

### Keyboard Shortcuts

| Key | Description |
|-----|-------------|
| `Ctrl+,` | Decrease split ratio (move divider left/up) |
| `Ctrl+.` | Increase split ratio (move divider right/down) |
| `Ctrl+S` | Swap left and right panes |
| `Ctrl+=` or `Ctrl+0` | Reset to 50/50 split |

### API Reference

```go
// Create split view
func NewSplitView(left, right Component, orientation layout.Orientation) *SplitView

// Set split ratio (0.0 - 1.0)
func (sv *SplitView) SetSplitRatio(ratio float64)

// Set minimum and maximum ratio constraints
func (sv *SplitView) SetMinMaxRatio(min, max float64)

// Set divider style
func (sv *SplitView) SetDividerStyle(style lipgloss.Style)

// Enable/disable divider
func (sv *SplitView) EnableDivider()
func (sv *SplitView) DisableDivider()

// Get panes
func (sv *SplitView) GetLeftPane() Component
func (sv *SplitView) GetRightPane() Component

// Swap panes
func (sv *SplitView) SwapPanes()

// Set total size
func (sv *SplitView) SetSize(width, height int)
```

---

## MultiColumnLayout

Displays content in multiple columns with flexible widths and alignment.

### Features

- **N Columns**: Support for any number of columns
- **Fixed/Flex Widths**: Mix fixed-width and flexible columns
- **Weight Distribution**: Control flex space distribution
- **Gaps**: Configurable spacing between columns
- **Vertical Alignment**: Top, Center, Bottom, Stretch
- **Dynamic Columns**: Add/remove columns at runtime

### Usage Example

```go
import (
    "github.com/aidevelopers/ainative/internal/tui/components"
)

// Create columns
col1 := components.NewTextComponent("Column 1")
col2 := components.NewTextComponent("Column 2")
col3 := components.NewTextComponent("Column 3")

// Create multi-column layout
mcl := components.NewMultiColumnLayout([]Component{col1, col2, col3})
mcl.SetSize(120, 40)

// Set column widths (0 = flex)
mcl.SetColumnWidths([]int{30, 0, 0}) // First fixed at 30, others flex

// Set flex weights for auto columns
mcl.SetColumnWeights([]int{0, 2, 1}) // Second gets 2x space of third

// Set gap between columns
mcl.SetGap(2)

// Set vertical alignment
mcl.SetAlignment(components.ColumnAlignTop)
```

### Layout Examples

**Fixed + Flex Widths:**
```
┌────────┬─────────────┬──────────┐
│        │             │          │
│ Fixed  │   Flex 2x   │  Flex 1x │
│  30    │             │          │
└────────┴─────────────┴──────────┘
```

**Equal Flex Widths:**
```
┌──────────┬──────────┬──────────┐
│          │          │          │
│   33%    │   33%    │   33%    │
│          │          │          │
└──────────┴──────────┴──────────┘
```

### Column Alignment

| Alignment | Description | Visual |
|-----------|-------------|--------|
| **ColumnAlignTop** | Align columns to top | Content starts at top |
| **ColumnAlignCenter** | Center columns vertically | Content centered |
| **ColumnAlignBottom** | Align columns to bottom | Content at bottom |
| **ColumnAlignStretch** | Stretch to fill height | Content fills height |

### Keyboard Shortcuts

| Key | Description |
|-----|-------------|
| `Tab` or `→` | Move to next column |
| `Shift+Tab` or `←` | Move to previous column |

### API Reference

```go
// Create multi-column layout
func NewMultiColumnLayout(columns []Component) *MultiColumnLayout

// Set column widths (0 = flex)
func (mcl *MultiColumnLayout) SetColumnWidths(widths []int)

// Set column weights for flex columns
func (mcl *MultiColumnLayout) SetColumnWeights(weights []int)

// Set gap between columns
func (mcl *MultiColumnLayout) SetGap(gap int)

// Set vertical alignment
func (mcl *MultiColumnLayout) SetAlignment(align ColumnAlignment)

// Add/remove columns
func (mcl *MultiColumnLayout) AddColumn(component Component)
func (mcl *MultiColumnLayout) RemoveColumn(index int)

// Set specific column width
func (mcl *MultiColumnLayout) SetColumnWidth(index, width int)

// Set specific column weight
func (mcl *MultiColumnLayout) SetColumnWeight(index, weight int)

// Get column
func (mcl *MultiColumnLayout) GetColumn(index int) Component

// Get column count
func (mcl *MultiColumnLayout) GetColumnCount() int

// Active column management
func (mcl *MultiColumnLayout) SetActiveColumn(index int)
func (mcl *MultiColumnLayout) GetActiveColumn() int
```

---

## Mouse Event Helpers

Utility functions for handling mouse events.

### MouseState

Tracks mouse state for drag and click operations.

```go
// Create mouse state tracker
ms := components.NewMouseState()

// Update with mouse message
ms.Update(mouseMsg)

// Check state
if ms.IsDragging {
    dx, dy := ms.DragDistance()
    // Handle drag
}
```

### Bounds Checking

```go
rect := layout.Rectangle{X: 10, Y: 10, Width: 40, Height: 20}

// Check if mouse is within bounds
if components.IsInBounds(mouseX, mouseY, rect) {
    // Handle click
}

// Alternative with separate coordinates
if components.IsInBoundsXY(mouseX, mouseY, x, y, width, height) {
    // Handle click
}
```

### Mouse Event Detection

```go
// Detect specific mouse events
if components.IsLeftClick(msg) {
    // Handle left click
}

if components.IsRightClick(msg) {
    // Handle right click
}

if components.IsMouseMotion(msg) {
    // Handle mouse motion
}

if components.IsScrollUp(msg) {
    // Handle scroll up
}
```

### Drag Handler

```go
// Create drag handler with callbacks
dh := components.NewDragHandler(
    func(msg tea.MouseMsg) tea.Cmd {
        // On drag start
        return nil
    },
    func(msg tea.MouseMsg) tea.Cmd {
        // On drag
        return nil
    },
    func(msg tea.MouseMsg) tea.Cmd {
        // On drag end
        return nil
    },
)

// Handle mouse message
cmd := dh.HandleMouseMsg(msg)

// Get drag delta
dx, dy := dh.DragDelta()
```

### Hover Detector

```go
// Create hover detector
hd := components.NewHoverDetector(
    bounds,
    onEnter,  // Called when mouse enters
    onLeave,  // Called when mouse leaves
    onHover,  // Called while hovering
)

// Update with mouse message
cmd := hd.Update(msg)

// Check hover state
if hd.IsHovering() {
    // Show tooltip, etc.
}
```

---

## Interactive Examples

### Example 1: Draggable Modal Window

```go
// Create modal content
modal := components.NewModalComponent("Settings")

// Make it draggable
draggable := components.NewDraggable(modal, 20, 10)
draggable.SetDragHandle(0, 0, 60, 1) // Only title bar
draggable.SetBounds(0, 0, 80, 24)    // Screen bounds
draggable.SetSize(60, 20)
```

### Example 2: Resizable Text Editor

```go
// Create text editor
editor := components.NewTextAreaComponent()

// Make it resizable
resizable := components.NewResizable(editor, 80, 30)
resizable.SetMinSize(40, 10)
resizable.SetMaxSize(120, 50)
```

### Example 3: Split View with File Tree and Editor

```go
// Create file tree and editor
fileTree := components.NewTreeComponent()
editor := components.NewTextAreaComponent()

// Create split view
split := components.NewSplitView(fileTree, editor, layout.Horizontal)
split.SetSplitRatio(0.3) // 30% file tree, 70% editor
split.SetMinMaxRatio(0.2, 0.5)
```

### Example 4: Three-Column Dashboard

```go
// Create dashboard columns
sidebar := components.NewMenuComponent()
content := components.NewContentComponent()
inspector := components.NewInspectorComponent()

// Create multi-column layout
dashboard := components.NewMultiColumnLayout([]Component{
    sidebar, content, inspector,
})
dashboard.SetColumnWidths([]int{20, 0, 30}) // Fixed, Flex, Fixed
dashboard.SetGap(1)
dashboard.SetAlignment(components.ColumnAlignStretch)
```

### Example 5: Nested Splits

```go
// Create complex layout with nested splits
topLeft := components.NewComponent1()
topRight := components.NewComponent2()
bottom := components.NewComponent3()

// Horizontal split for top
topSplit := components.NewSplitView(topLeft, topRight, layout.Horizontal)
topSplit.SetSplitRatio(0.5)

// Vertical split for entire view
mainSplit := components.NewSplitView(topSplit, bottom, layout.Vertical)
mainSplit.SetSplitRatio(0.7) // 70% top, 30% bottom
```

---

## Best Practices

### 1. Mouse Support

Enable mouse support in your Bubble Tea program:

```go
func main() {
    p := tea.NewProgram(
        model,
        tea.WithAltScreen(),
        tea.WithMouseAllMotion(), // Enable full mouse support
    )
    p.Run()
}
```

### 2. Coordinate Handling

When forwarding mouse events to child components, adjust coordinates:

```go
// In parent component
relativeMsg := msg
relativeMsg.X = msg.X - componentX
relativeMsg.Y = msg.Y - componentY
childComponent.Update(relativeMsg)
```

### 3. Focus Management

Only one component should handle keyboard input at a time:

```go
if component.Focused() {
    // Handle keyboard input
} else {
    // Ignore or forward to focused component
}
```

### 4. Performance

For complex layouts, cache rendered content when possible:

```go
if !needsUpdate {
    return cachedView
}
cachedView = renderView()
return cachedView
```

### 5. Accessibility

Always provide keyboard alternatives to mouse actions:

- Draggable: `Alt+Arrow` keys
- Resizable: `Ctrl+Arrow` keys
- Split View: `Ctrl+,` and `Ctrl+.`
- Multi-Column: `Tab` and `Shift+Tab`

---

## Troubleshooting

### Issue: Mouse events not working

**Solution:** Ensure mouse support is enabled:
```go
tea.WithMouseAllMotion()
```

### Issue: Components not sizing correctly

**Solution:** Implement the `Sizeable` interface on your components:
```go
func (c *MyComponent) SetSize(width, height int) {
    c.width = width
    c.height = height
}

func (c *MyComponent) GetSize() (int, int) {
    return c.width, c.height
}
```

### Issue: Drag/resize feels jerky

**Solution:** Process mouse motion events efficiently and avoid heavy operations in Update():
```go
case tea.MouseMsg:
    if msg.Action == tea.MouseActionMotion {
        // Quick position update only
        updatePosition(msg.X, msg.Y)
        return m, nil
    }
```

### Issue: Components overlap incorrectly

**Solution:** Use proper z-ordering and bounds checking:
```go
// Render back-to-front
renderComponent(background)
renderComponent(middleLayer)
renderComponent(foreground)
```

---

## Contributing

Found a bug or want to add a feature? Contributions are welcome!

1. Fork the repository
2. Create a feature branch
3. Add tests for your changes
4. Submit a pull request

---

## License

MIT License - See LICENSE file for details

---

Built by AINative Studio
Powered by AINative Cloud
