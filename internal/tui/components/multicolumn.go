package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ColumnAlignment defines how columns are vertically aligned
type ColumnAlignment int

const (
	// ColumnAlignTop aligns columns to the top
	ColumnAlignTop ColumnAlignment = iota
	// ColumnAlignCenter centers columns vertically
	ColumnAlignCenter
	// ColumnAlignBottom aligns columns to the bottom
	ColumnAlignBottom
	// ColumnAlignStretch stretches columns to fill available height
	ColumnAlignStretch
)

// String returns the string representation of the alignment
func (ca ColumnAlignment) String() string {
	switch ca {
	case ColumnAlignTop:
		return "Top"
	case ColumnAlignCenter:
		return "Center"
	case ColumnAlignBottom:
		return "Bottom"
	case ColumnAlignStretch:
		return "Stretch"
	default:
		return "Unknown"
	}
}

// MultiColumnLayout displays content in multiple columns with flexible widths
type MultiColumnLayout struct {
	columns      []Component     // Column components
	columnWidths []int           // Fixed widths or 0 for flex
	weights      []int           // Flex weights for auto columns
	gaps         int             // Space between columns
	width        int             // Total width
	height       int             // Total height
	alignment    ColumnAlignment // Vertical alignment
	styles       []lipgloss.Style // Styles for each column
	focused      bool            // Whether component has keyboard focus
	activeColumn int             // Currently active/focused column
}

// NewMultiColumnLayout creates a new multi-column layout
func NewMultiColumnLayout(columns []Component) *MultiColumnLayout {
	columnCount := len(columns)
	return &MultiColumnLayout{
		columns:      columns,
		columnWidths: make([]int, columnCount),
		weights:      make([]int, columnCount),
		gaps:         1,
		width:        80,
		height:       24,
		alignment:    ColumnAlignTop,
		styles:       make([]lipgloss.Style, columnCount),
		focused:      false,
		activeColumn: 0,
	}
}

// SetColumnWidths sets fixed widths for columns (0 = flex)
func (mcl *MultiColumnLayout) SetColumnWidths(widths []int) {
	if len(widths) != len(mcl.columns) {
		return
	}
	mcl.columnWidths = widths
	mcl.updateColumnSizes()
}

// SetColumnWeights sets flex weights for auto-sized columns
func (mcl *MultiColumnLayout) SetColumnWeights(weights []int) {
	if len(weights) != len(mcl.columns) {
		return
	}
	mcl.weights = weights
	mcl.updateColumnSizes()
}

// SetGap sets the space between columns
func (mcl *MultiColumnLayout) SetGap(gap int) {
	mcl.gaps = gap
	mcl.updateColumnSizes()
}

// SetAlignment sets the vertical alignment of columns
func (mcl *MultiColumnLayout) SetAlignment(align ColumnAlignment) {
	mcl.alignment = align
}

// AddColumn adds a new column to the layout
func (mcl *MultiColumnLayout) AddColumn(component Component) {
	mcl.columns = append(mcl.columns, component)
	mcl.columnWidths = append(mcl.columnWidths, 0)
	mcl.weights = append(mcl.weights, 1)
	mcl.styles = append(mcl.styles, lipgloss.NewStyle())
	mcl.updateColumnSizes()
}

// RemoveColumn removes a column by index
func (mcl *MultiColumnLayout) RemoveColumn(index int) {
	if index < 0 || index >= len(mcl.columns) {
		return
	}

	mcl.columns = append(mcl.columns[:index], mcl.columns[index+1:]...)
	mcl.columnWidths = append(mcl.columnWidths[:index], mcl.columnWidths[index+1:]...)
	mcl.weights = append(mcl.weights[:index], mcl.weights[index+1:]...)
	mcl.styles = append(mcl.styles[:index], mcl.styles[index+1:]...)

	// Adjust active column if needed
	if mcl.activeColumn >= len(mcl.columns) {
		mcl.activeColumn = len(mcl.columns) - 1
	}
	if mcl.activeColumn < 0 {
		mcl.activeColumn = 0
	}

	mcl.updateColumnSizes()
}

// SetColumnWidth sets the width of a specific column (0 = flex)
func (mcl *MultiColumnLayout) SetColumnWidth(index, width int) {
	if index < 0 || index >= len(mcl.columnWidths) {
		return
	}
	mcl.columnWidths[index] = width
	mcl.updateColumnSizes()
}

// SetColumnWeight sets the flex weight of a specific column
func (mcl *MultiColumnLayout) SetColumnWeight(index, weight int) {
	if index < 0 || index >= len(mcl.weights) {
		return
	}
	mcl.weights[index] = weight
	mcl.updateColumnSizes()
}

// SetColumnStyle sets the style for a specific column
func (mcl *MultiColumnLayout) SetColumnStyle(index int, style lipgloss.Style) {
	if index < 0 || index >= len(mcl.styles) {
		return
	}
	mcl.styles[index] = style
}

// GetColumn returns the component at the specified index
func (mcl *MultiColumnLayout) GetColumn(index int) Component {
	if index < 0 || index >= len(mcl.columns) {
		return nil
	}
	return mcl.columns[index]
}

// GetColumnCount returns the number of columns
func (mcl *MultiColumnLayout) GetColumnCount() int {
	return len(mcl.columns)
}

// SetActiveColumn sets the active column for focus
func (mcl *MultiColumnLayout) SetActiveColumn(index int) {
	if index < 0 || index >= len(mcl.columns) {
		return
	}
	mcl.activeColumn = index
}

// GetActiveColumn returns the index of the active column
func (mcl *MultiColumnLayout) GetActiveColumn() int {
	return mcl.activeColumn
}

// SetSize sets the total size of the layout
func (mcl *MultiColumnLayout) SetSize(width, height int) {
	mcl.width = width
	mcl.height = height
	mcl.updateColumnSizes()
}

// GetSize returns the total size
func (mcl *MultiColumnLayout) GetSize() (width, height int) {
	return mcl.width, mcl.height
}

// Focus gives keyboard focus to the component
func (mcl *MultiColumnLayout) Focus() tea.Cmd {
	mcl.focused = true
	if mcl.activeColumn >= 0 && mcl.activeColumn < len(mcl.columns) {
		if focusable, ok := mcl.columns[mcl.activeColumn].(Focusable); ok {
			return focusable.Focus()
		}
	}
	return nil
}

// Blur removes keyboard focus
func (mcl *MultiColumnLayout) Blur() {
	mcl.focused = false
	for _, col := range mcl.columns {
		if focusable, ok := col.(Focusable); ok {
			focusable.Blur()
		}
	}
}

// Focused returns whether component has focus
func (mcl *MultiColumnLayout) Focused() bool {
	return mcl.focused
}

// updateColumnSizes calculates and updates column sizes
func (mcl *MultiColumnLayout) updateColumnSizes() {
	if len(mcl.columns) == 0 {
		return
	}

	// Calculate total gap space
	totalGaps := mcl.gaps * (len(mcl.columns) - 1)
	availableWidth := mcl.width - totalGaps

	// Calculate fixed width columns total
	fixedTotal := 0
	flexCount := 0
	totalWeight := 0

	for i, width := range mcl.columnWidths {
		if width > 0 {
			fixedTotal += width
		} else {
			flexCount++
			totalWeight += mcl.weights[i]
		}
	}

	// Calculate flex width
	flexWidth := availableWidth - fixedTotal
	if flexWidth < 0 {
		flexWidth = 0
	}

	// Calculate actual widths and update components
	calculatedWidths := make([]int, len(mcl.columns))
	for i := range mcl.columns {
		if mcl.columnWidths[i] > 0 {
			calculatedWidths[i] = mcl.columnWidths[i]
		} else {
			if totalWeight > 0 {
				calculatedWidths[i] = (flexWidth * mcl.weights[i]) / totalWeight
			} else {
				calculatedWidths[i] = flexWidth / flexCount
			}
		}

		// Update component size if it implements Sizeable
		if sizeable, ok := mcl.columns[i].(Sizeable); ok {
			height := mcl.height
			if mcl.alignment != ColumnAlignStretch {
				// For non-stretch alignment, let component use its natural height
				_, naturalHeight := sizeable.GetSize()
				if naturalHeight > 0 && naturalHeight < height {
					height = naturalHeight
				}
			}
			sizeable.SetSize(calculatedWidths[i], height)
		}
	}
}

// Init initializes the component
func (mcl *MultiColumnLayout) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(mcl.columns))
	for _, col := range mcl.columns {
		if col != nil {
			cmds = append(cmds, col.Init())
		}
	}
	return tea.Batch(cmds...)
}

// Update handles messages and updates state
func (mcl *MultiColumnLayout) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return mcl.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		mcl.SetSize(msg.Width, msg.Height)
	}

	// Forward to all columns
	cmds := make([]tea.Cmd, 0, len(mcl.columns))
	for i, col := range mcl.columns {
		if col != nil {
			updatedCol, cmd := col.Update(msg)
			mcl.columns[i] = updatedCol
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return mcl, tea.Batch(cmds...)
}

// handleKeyMsg processes keyboard events
func (mcl *MultiColumnLayout) handleKeyMsg(msg tea.KeyMsg) (Component, tea.Cmd) {
	// Handle column navigation when focused
	if mcl.focused {
		handled := false
		switch msg.String() {
		case "tab", "right":
			// Move to next column
			mcl.activeColumn = (mcl.activeColumn + 1) % len(mcl.columns)
			handled = true
		case "shift+tab", "left":
			// Move to previous column
			mcl.activeColumn--
			if mcl.activeColumn < 0 {
				mcl.activeColumn = len(mcl.columns) - 1
			}
			handled = true
		}

		if handled {
			// Focus the new active column
			for i, col := range mcl.columns {
				if focusable, ok := col.(Focusable); ok {
					if i == mcl.activeColumn {
						focusable.Focus()
					} else {
						focusable.Blur()
					}
				}
			}
			return mcl, nil
		}
	}

	// Forward to all columns
	cmds := make([]tea.Cmd, 0, len(mcl.columns))
	for i, col := range mcl.columns {
		if col != nil {
			updatedCol, cmd := col.Update(msg)
			mcl.columns[i] = updatedCol
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return mcl, tea.Batch(cmds...)
}

// View renders the multi-column layout
func (mcl *MultiColumnLayout) View() string {
	if len(mcl.columns) == 0 {
		return ""
	}

	// Render each column
	columnViews := make([]string, len(mcl.columns))
	columnHeights := make([]int, len(mcl.columns))

	for i, col := range mcl.columns {
		if col != nil {
			view := col.View()
			// Apply style if set (check using String() method)
			if mcl.styles[i].String() != "" {
				view = mcl.styles[i].Render(view)
			}
			columnViews[i] = view
			lines := strings.Split(view, "\n")
			columnHeights[i] = len(lines)
		}
	}

	// Find max height for alignment
	maxHeight := 0
	for _, h := range columnHeights {
		if h > maxHeight {
			maxHeight = h
		}
	}

	// Apply vertical alignment
	alignedViews := make([][]string, len(mcl.columns))
	for i, view := range columnViews {
		lines := strings.Split(view, "\n")
		height := len(lines)

		switch mcl.alignment {
		case ColumnAlignTop:
			alignedViews[i] = lines
			// Pad bottom
			for j := height; j < maxHeight; j++ {
				alignedViews[i] = append(alignedViews[i], "")
			}

		case ColumnAlignCenter:
			topPad := (maxHeight - height) / 2
			bottomPad := maxHeight - height - topPad
			alignedViews[i] = make([]string, 0, maxHeight)
			for j := 0; j < topPad; j++ {
				alignedViews[i] = append(alignedViews[i], "")
			}
			alignedViews[i] = append(alignedViews[i], lines...)
			for j := 0; j < bottomPad; j++ {
				alignedViews[i] = append(alignedViews[i], "")
			}

		case ColumnAlignBottom:
			topPad := maxHeight - height
			alignedViews[i] = make([]string, 0, maxHeight)
			for j := 0; j < topPad; j++ {
				alignedViews[i] = append(alignedViews[i], "")
			}
			alignedViews[i] = append(alignedViews[i], lines...)

		case ColumnAlignStretch:
			alignedViews[i] = lines
			// Pad to max height
			for j := height; j < maxHeight; j++ {
				alignedViews[i] = append(alignedViews[i], "")
			}
		}
	}

	// Combine columns horizontally
	result := make([]string, 0, maxHeight)
	gap := strings.Repeat(" ", mcl.gaps)

	for row := 0; row < maxHeight; row++ {
		line := ""
		for col := 0; col < len(mcl.columns); col++ {
			if col > 0 {
				line += gap
			}

			if row < len(alignedViews[col]) {
				line += alignedViews[col][row]
			}
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
