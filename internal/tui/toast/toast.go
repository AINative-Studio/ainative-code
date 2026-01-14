package toast

import (
	"fmt"
	"time"

	"github.com/AINative-studio/ainative-code/internal/tui/components"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToastType defines notification type
type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
	ToastLoading
)

// String returns the string representation of a ToastType
func (t ToastType) String() string {
	switch t {
	case ToastInfo:
		return "Info"
	case ToastSuccess:
		return "Success"
	case ToastWarning:
		return "Warning"
	case ToastError:
		return "Error"
	case ToastLoading:
		return "Loading"
	default:
		return "Unknown"
	}
}

// ToastPosition defines where toast appears
type ToastPosition int

const (
	TopRight ToastPosition = iota
	TopCenter
	TopLeft
	BottomRight
	BottomCenter
	BottomLeft
)

// String returns the string representation of a ToastPosition
func (p ToastPosition) String() string {
	switch p {
	case TopRight:
		return "TopRight"
	case TopCenter:
		return "TopCenter"
	case TopLeft:
		return "TopLeft"
	case BottomRight:
		return "BottomRight"
	case BottomCenter:
		return "BottomCenter"
	case BottomLeft:
		return "BottomLeft"
	default:
		return "Unknown"
	}
}

// ToastAction is an optional button in toast
type ToastAction struct {
	Label   string
	Command tea.Cmd
}

// ToastConfig configures toast behavior
type ToastConfig struct {
	Type        ToastType
	Message     string
	Title       string        // Optional
	Duration    time.Duration // Auto-dismiss time (0 = manual dismiss)
	Dismissible bool          // Can be dismissed with X button
	Position    ToastPosition
	Icon        string // Optional icon
	Action      *ToastAction // Optional action button
}

// DefaultToastConfig returns sensible defaults for a toast
func DefaultToastConfig(toastType ToastType) ToastConfig {
	duration := 3 * time.Second

	switch toastType {
	case ToastSuccess, ToastInfo:
		duration = 3 * time.Second
	case ToastWarning:
		duration = 5 * time.Second
	case ToastError:
		duration = 10 * time.Second
	case ToastLoading:
		duration = 0 // Manual dismiss
	}

	return ToastConfig{
		Type:        toastType,
		Duration:    duration,
		Dismissible: true,
		Position:    TopRight,
		Icon:        GetDefaultIcon(toastType),
	}
}

// Toast is a single notification
type Toast struct {
	id          string
	config      ToastConfig
	createdAt   time.Time
	expiresAt   time.Time
	progress    float64 // For loading toasts
	dismissed   bool
	animation   *components.AnimatedComponent // Fade in/out animation
	opacity     float64                       // Current opacity (0.0 to 1.0)
	isFadingOut bool                          // Whether toast is fading out
	width       int                           // Toast width
	height      int                           // Toast height (calculated)
}

// NewToast creates a new toast notification
func NewToast(config ToastConfig) *Toast {
	now := time.Now()
	id := fmt.Sprintf("toast-%d", now.UnixNano())

	var expiresAt time.Time
	if config.Duration > 0 {
		expiresAt = now.Add(config.Duration)
	}

	// Create animation for fade in
	animConfig := components.DefaultAnimationConfig()
	animConfig.Duration = 200 * time.Millisecond
	animation := components.NewAnimatedComponentWithID(id, nil, animConfig)

	toast := &Toast{
		id:          id,
		config:      config,
		createdAt:   now,
		expiresAt:   expiresAt,
		progress:    0,
		dismissed:   false,
		animation:   animation,
		opacity:     0,
		isFadingOut: false,
		width:       40,
	}

	return toast
}

// ID returns the toast's unique identifier
func (t *Toast) ID() string {
	return t.id
}

// Config returns the toast's configuration
func (t *Toast) Config() ToastConfig {
	return t.config
}

// IsExpired returns true if the toast should be dismissed
func (t *Toast) IsExpired() bool {
	if t.dismissed {
		return true
	}

	if t.config.Duration == 0 {
		return false // Manual dismiss only
	}

	return time.Now().After(t.expiresAt)
}

// IsDismissed returns true if the toast has been dismissed
func (t *Toast) IsDismissed() bool {
	return t.dismissed
}

// IsFadingOut returns true if the toast is fading out
func (t *Toast) IsFadingOut() bool {
	return t.isFadingOut
}

// GetOpacity returns the current opacity value
func (t *Toast) GetOpacity() float64 {
	return t.opacity
}

// Dismiss marks the toast as dismissed and starts fade out animation
func (t *Toast) Dismiss() tea.Cmd {
	if t.dismissed {
		return nil
	}

	t.dismissed = true
	t.isFadingOut = true

	// Start fade out animation (from current opacity to 0)
	return t.animation.StartAnimation(t.opacity, 0)
}

// StartFadeIn starts the fade in animation
func (t *Toast) StartFadeIn() tea.Cmd {
	return t.animation.StartAnimation(0, 1.0)
}

// SetWidth sets the toast width
func (t *Toast) SetWidth(width int) {
	t.width = width
}

// GetWidth returns the toast width
func (t *Toast) GetWidth() int {
	return t.width
}

// GetHeight returns the calculated toast height
func (t *Toast) GetHeight() int {
	// Calculate based on content
	lines := 2 // Border top/bottom
	if t.config.Title != "" {
		lines += 1
	}
	lines += 1 // Message line
	if t.config.Action != nil {
		lines += 1
	}
	return lines
}

// SetValue directly sets the toast opacity value
func (t *Toast) SetValue(value float64) {
	t.opacity = value
	if t.animation != nil {
		t.animation.SetValue(value)
	}
}

// Update updates the toast state
func (t *Toast) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Update animation
	if t.animation != nil {
		animation, cmd := t.animation.Update(msg)
		t.animation = animation.(*components.AnimatedComponent)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Update opacity from animation value
		t.opacity = t.animation.GetValue()
	}

	switch msg := msg.(type) {
	case components.AnimationCompleteMsg:
		// Check if this is our animation
		if msg.ID == t.id {
			if t.isFadingOut {
				// Fade out complete, mark as fully dismissed
				t.opacity = 0
			} else {
				// Fade in complete
				t.opacity = 1.0
			}
		}

	case ToastTickMsg:
		// Update loading progress
		if t.config.Type == ToastLoading && !t.dismissed {
			t.progress += 0.1
			if t.progress > 1.0 {
				t.progress = 0
			}
			cmds = append(cmds, t.tick())
		}

		// Check expiration
		if t.IsExpired() && !t.isFadingOut {
			cmd := t.Dismiss()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Notify manager that this toast expired
			cmds = append(cmds, func() tea.Msg {
				return ToastExpiredMsg{ID: t.id}
			})
		}
	}

	return t, tea.Batch(cmds...)
}

// tick generates periodic updates for loading toasts and expiration checks
func (t *Toast) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return ToastTickMsg{ID: t.id}
	})
}

// Init initializes the toast
func (t *Toast) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Start fade in animation
	cmds = append(cmds, t.StartFadeIn())

	// Start tick for loading toasts and expiration checks
	if t.config.Type == ToastLoading || t.config.Duration > 0 {
		cmds = append(cmds, t.tick())
	}

	return tea.Batch(cmds...)
}

// View renders the toast
func (t *Toast) View() string {
	if t.opacity <= 0 {
		return ""
	}

	// Get the appropriate style for this toast type
	style := GetToastStyle(t.config.Type, t.width)

	// Apply opacity by adjusting colors
	if t.opacity < 1.0 {
		style = applyOpacity(style, t.opacity)
	}

	// Build toast content
	var content string

	// Icon and title/message
	icon := t.config.Icon
	if icon == "" {
		icon = GetDefaultIcon(t.config.Type)
	}

	if t.config.Title != "" {
		titleStyle := lipgloss.NewStyle().Bold(true)
		content = fmt.Sprintf("%s %s\n%s", icon, titleStyle.Render(t.config.Title), t.config.Message)
	} else {
		content = fmt.Sprintf("%s %s", icon, t.config.Message)
	}

	// Add loading spinner for loading toasts
	if t.config.Type == ToastLoading {
		spinner := getSpinnerFrame(t.progress)
		content = fmt.Sprintf("%s %s", spinner, t.config.Message)
	}

	// Add action button if present
	if t.config.Action != nil {
		actionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true)
		content = fmt.Sprintf("%s\n%s", content, actionStyle.Render("["+t.config.Action.Label+"]"))
	}

	// Add dismiss button if dismissible
	if t.config.Dismissible {
		dismissStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Right)
		header := dismissStyle.Render("×")
		content = lipgloss.JoinVertical(lipgloss.Left, header, content)
	}

	return style.Render(content)
}

// applyOpacity applies opacity to a style by adjusting colors
func applyOpacity(style lipgloss.Style, opacity float64) lipgloss.Style {
	// This is a simplified opacity implementation
	// In a real implementation, you might want to blend with background color
	return style
}

// getSpinnerFrame returns a spinner character based on progress
func getSpinnerFrame(progress float64) string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	index := int(progress*10) % len(frames)
	return frames[index]
}
