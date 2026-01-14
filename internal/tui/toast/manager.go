package toast

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToastManager manages multiple toasts with queue
type ToastManager struct {
	toasts      []*Toast      // Currently visible toasts
	queue       []*Toast      // Queued toasts waiting to be shown
	maxToasts   int           // Max visible toasts
	position    ToastPosition // Default position
	width       int           // Available width
	height      int           // Available height
	screenWidth int           // Total screen width
	screenHeight int          // Total screen height
	enabled     bool          // Whether toasts are enabled
}

// NewToastManager creates a new toast manager
func NewToastManager() *ToastManager {
	return &ToastManager{
		toasts:    make([]*Toast, 0),
		queue:     make([]*Toast, 0),
		maxToasts: 3, // Default max 3 visible toasts
		position:  TopRight,
		width:     40,
		height:    0,
		enabled:   true,
	}
}

// SetMaxToasts sets the maximum number of visible toasts
func (m *ToastManager) SetMaxToasts(max int) {
	if max < 1 {
		max = 1
	}
	m.maxToasts = max
	m.processQueue()
}

// GetMaxToasts returns the maximum number of visible toasts
func (m *ToastManager) GetMaxToasts() int {
	return m.maxToasts
}

// SetPosition sets the default toast position
func (m *ToastManager) SetPosition(pos ToastPosition) {
	m.position = pos
}

// GetPosition returns the default toast position
func (m *ToastManager) GetPosition() ToastPosition {
	return m.position
}

// SetSize sets the manager's dimensions
func (m *ToastManager) SetSize(width, height int) {
	m.screenWidth = width
	m.screenHeight = height

	// Update toast widths
	toastWidth := 40
	if width < 50 {
		toastWidth = width - 4
	}
	m.width = toastWidth

	for _, toast := range m.toasts {
		toast.SetWidth(toastWidth)
	}
}

// SetEnabled enables or disables toast notifications
func (m *ToastManager) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// IsEnabled returns whether toast notifications are enabled
func (m *ToastManager) IsEnabled() bool {
	return m.enabled
}

// ShowToast displays a new toast notification
func (m *ToastManager) ShowToast(config ToastConfig) tea.Cmd {
	if !m.enabled {
		return nil
	}

	// Use manager's default position if not specified
	if config.Position == 0 {
		config.Position = m.position
	}

	toast := NewToast(config)
	toast.SetWidth(m.width)

	// If we have room, show immediately
	if len(m.toasts) < m.maxToasts {
		m.toasts = append(m.toasts, toast)
		return toast.Init()
	}

	// Otherwise, add to queue
	m.queue = append(m.queue, toast)
	return nil
}

// ShowInfo displays an info toast
func (m *ToastManager) ShowInfo(message string) tea.Cmd {
	config := DefaultToastConfig(ToastInfo)
	config.Message = message
	return m.ShowToast(config)
}

// ShowSuccess displays a success toast
func (m *ToastManager) ShowSuccess(message string) tea.Cmd {
	config := DefaultToastConfig(ToastSuccess)
	config.Message = message
	return m.ShowToast(config)
}

// ShowWarning displays a warning toast
func (m *ToastManager) ShowWarning(message string) tea.Cmd {
	config := DefaultToastConfig(ToastWarning)
	config.Message = message
	return m.ShowToast(config)
}

// ShowError displays an error toast
func (m *ToastManager) ShowError(message string) tea.Cmd {
	config := DefaultToastConfig(ToastError)
	config.Message = message
	return m.ShowToast(config)
}

// ShowLoading displays a loading toast
func (m *ToastManager) ShowLoading(message string) tea.Cmd {
	config := DefaultToastConfig(ToastLoading)
	config.Message = message
	return m.ShowToast(config)
}

// DismissToast dismisses a specific toast by ID
func (m *ToastManager) DismissToast(id string) tea.Cmd {
	for _, toast := range m.toasts {
		if toast.ID() == id {
			return toast.Dismiss()
		}
	}
	return nil
}

// DismissAll dismisses all visible toasts
func (m *ToastManager) DismissAll() tea.Cmd {
	var cmds []tea.Cmd
	for _, toast := range m.toasts {
		cmd := toast.Dismiss()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

// RemoveToast removes a toast from the visible list
func (m *ToastManager) RemoveToast(id string) {
	for i, toast := range m.toasts {
		if toast.ID() == id {
			m.toasts = append(m.toasts[:i], m.toasts[i+1:]...)
			break
		}
	}
}

// processQueue moves toasts from queue to visible list
func (m *ToastManager) processQueue() tea.Cmd {
	if len(m.queue) == 0 {
		return nil
	}

	var cmds []tea.Cmd

	// Move queued toasts to visible list if there's room
	for len(m.toasts) < m.maxToasts && len(m.queue) > 0 {
		toast := m.queue[0]
		m.queue = m.queue[1:]
		m.toasts = append(m.toasts, toast)
		cmds = append(cmds, toast.Init())
	}

	return tea.Batch(cmds...)
}

// HasToasts returns true if there are visible toasts
func (m *ToastManager) HasToasts() bool {
	return len(m.toasts) > 0
}

// GetToasts returns the list of visible toasts
func (m *ToastManager) GetToasts() []*Toast {
	return m.toasts
}

// GetQueueLength returns the number of queued toasts
func (m *ToastManager) GetQueueLength() int {
	return len(m.queue)
}

// Update updates all toasts and handles toast-related messages
func (m *ToastManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ShowToastMsg:
		cmd := m.ShowToast(msg.Config)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case DismissToastMsg:
		cmd := m.DismissToast(msg.ID)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case DismissAllToastsMsg:
		cmd := m.DismissAll()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case ToastExpiredMsg:
		// Remove expired toast and process queue
		m.RemoveToast(msg.ID)
		cmd := m.processQueue()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case ToastActionMsg:
		// Execute action and dismiss toast
		if msg.Action != nil && msg.Action.Command != nil {
			cmds = append(cmds, msg.Action.Command)
		}
		cmd := m.DismissToast(msg.ToastID)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Update all visible toasts
	var toRemove []string
	for _, toast := range m.toasts {
		toastModel, cmd := toast.Update(msg)
		toast = toastModel.(*Toast)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Remove toasts that have completed fade out
		if toast.IsFadingOut() && toast.GetOpacity() <= 0 {
			toRemove = append(toRemove, toast.ID())
		}
	}

	// Remove fully faded out toasts
	for _, id := range toRemove {
		m.RemoveToast(id)
		// Process queue after removing
		cmd := m.processQueue()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// Init initializes the toast manager
func (m *ToastManager) Init() tea.Cmd {
	return nil
}

// View renders all visible toasts
func (m *ToastManager) View() string {
	if len(m.toasts) == 0 {
		return ""
	}

	// Render each toast
	var toastViews []string
	for _, toast := range m.toasts {
		view := toast.View()
		if view != "" {
			toastViews = append(toastViews, view)
		}
	}

	if len(toastViews) == 0 {
		return ""
	}

	// Stack toasts vertically with spacing
	content := lipgloss.JoinVertical(lipgloss.Left, toastViews...)

	// Position the toast stack based on configuration
	return m.positionToasts(content)
}

// positionToasts positions the toast stack on the screen
func (m *ToastManager) positionToasts(content string) string {
	if m.screenWidth == 0 || m.screenHeight == 0 {
		return content
	}

	// Determine horizontal and vertical alignment
	var hPos lipgloss.Position
	var vPos lipgloss.Position

	switch m.position {
	case TopLeft:
		hPos = lipgloss.Left
		vPos = lipgloss.Top
	case TopCenter:
		hPos = lipgloss.Center
		vPos = lipgloss.Top
	case TopRight:
		hPos = lipgloss.Right
		vPos = lipgloss.Top
	case BottomLeft:
		hPos = lipgloss.Left
		vPos = lipgloss.Bottom
	case BottomCenter:
		hPos = lipgloss.Center
		vPos = lipgloss.Bottom
	case BottomRight:
		hPos = lipgloss.Right
		vPos = lipgloss.Bottom
	default:
		hPos = lipgloss.Right
		vPos = lipgloss.Top
	}

	// Use Place to position the content
	return lipgloss.Place(
		m.screenWidth,
		m.screenHeight,
		hPos,
		vPos,
		content,
		lipgloss.WithWhitespaceChars(""),
	)
}

// ClearQueue clears all queued toasts
func (m *ToastManager) ClearQueue() {
	m.queue = make([]*Toast, 0)
}

// Stats returns statistics about the toast manager
func (m *ToastManager) Stats() string {
	return fmt.Sprintf("Visible: %d/%d, Queued: %d",
		len(m.toasts), m.maxToasts, len(m.queue))
}
