package dialogs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ModalConfig configures modal behavior and appearance
type ModalConfig struct {
	ZIndex          int               // Z-index for layering (default: auto-increment)
	Backdrop        BackdropStyle     // Backdrop appearance
	CloseOnEsc      bool              // Allow ESC key to close (default: true)
	CloseOnBackdrop bool              // Allow clicking backdrop to close (default: false)
	TrapFocus       bool              // Trap focus within modal (default: true)
	MaxWidth        int               // Maximum modal width (0 = no limit)
	MaxHeight       int               // Maximum modal height (0 = no limit)
	CenterX         bool              // Center horizontally (default: true)
	CenterY         bool              // Center vertically (default: true)
	AnimationConfig *AnimationConfig  // Optional animation configuration
}

// AnimationConfig defines animation behavior for modals
type AnimationConfig struct {
	FadeIn       bool    // Enable fade-in effect
	FadeOut      bool    // Enable fade-out effect
	Duration     int     // Animation duration in milliseconds
	InitialAlpha float64 // Initial opacity (0.0 - 1.0)
	FinalAlpha   float64 // Final opacity (0.0 - 1.0)
}

// DefaultModalConfig returns a sensible default modal configuration
func DefaultModalConfig() ModalConfig {
	return ModalConfig{
		ZIndex:          0, // Will be auto-assigned
		Backdrop:        DarkBackdrop,
		CloseOnEsc:      true,
		CloseOnBackdrop: false,
		TrapFocus:       true,
		MaxWidth:        0,
		MaxHeight:       0,
		CenterX:         true,
		CenterY:         true,
		AnimationConfig: nil, // No animation by default
	}
}

// MinimalModalConfig returns a minimal modal configuration with no backdrop
func MinimalModalConfig() ModalConfig {
	return ModalConfig{
		ZIndex:          0,
		Backdrop:        NoBackdrop,
		CloseOnEsc:      true,
		CloseOnBackdrop: false,
		TrapFocus:       false,
		MaxWidth:        0,
		MaxHeight:       0,
		CenterX:         true,
		CenterY:         true,
		AnimationConfig: nil,
	}
}

// BlurModalConfig returns a modal configuration with blur backdrop effect
func BlurModalConfig() ModalConfig {
	return ModalConfig{
		ZIndex:          0,
		Backdrop:        BlurBackdrop,
		CloseOnEsc:      true,
		CloseOnBackdrop: true, // Allow closing by clicking backdrop
		TrapFocus:       true,
		MaxWidth:        0,
		MaxHeight:       0,
		CenterX:         true,
		CenterY:         true,
		AnimationConfig: &AnimationConfig{
			FadeIn:       true,
			FadeOut:      true,
			Duration:     200,
			InitialAlpha: 0.0,
			FinalAlpha:   1.0,
		},
	}
}

// Modal wraps a Dialog with advanced modal features
type Modal struct {
	dialog  Dialog      // The underlying dialog
	config  ModalConfig // Modal configuration
	zIndex  int         // Current z-index
	visible bool        // Visibility state
	focused bool        // Focus state
	x, y    int         // Position coordinates
}

// NewModal creates a new modal wrapper around a dialog
func NewModal(dialog Dialog, config ModalConfig) *Modal {
	return &Modal{
		dialog:  dialog,
		config:  config,
		zIndex:  config.ZIndex,
		visible: true,
		focused: true,
		x:       0,
		y:       0,
	}
}

// Init initializes the modal
func (m *Modal) Init() tea.Cmd {
	return m.dialog.Init()
}

// Update handles messages for the modal
func (m *Modal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Forward message to underlying dialog
	updatedDialog, cmd := m.dialog.Update(msg)
	m.dialog = updatedDialog.(Dialog)
	return m, cmd
}

// View renders the modal
func (m *Modal) View() string {
	if !m.visible {
		return ""
	}
	return m.dialog.View()
}

// ID returns the modal's dialog ID
func (m *Modal) ID() string {
	return m.dialog.ID()
}

// SetSize updates the modal dimensions
func (m *Modal) SetSize(width, height int) {
	m.dialog.SetSize(width, height)
}

// IsClosing returns true if the dialog is requesting to be closed
func (m *Modal) IsClosing() bool {
	return m.dialog.IsClosing()
}

// Result returns the dialog result
func (m *Modal) Result() interface{} {
	return m.dialog.Result()
}

// GetDialog returns the underlying dialog
func (m *Modal) GetDialog() Dialog {
	return m.dialog
}

// GetConfig returns the modal configuration
func (m *Modal) GetConfig() ModalConfig {
	return m.config
}

// SetZIndex updates the modal's z-index
func (m *Modal) SetZIndex(zIndex int) {
	m.zIndex = zIndex
	m.config.ZIndex = zIndex
}

// GetZIndex returns the modal's current z-index
func (m *Modal) GetZIndex() int {
	return m.zIndex
}

// SetVisible controls modal visibility
func (m *Modal) SetVisible(visible bool) {
	m.visible = visible
}

// IsVisible returns true if the modal is visible
func (m *Modal) IsVisible() bool {
	return m.visible
}

// SetFocused controls modal focus state
func (m *Modal) SetFocused(focused bool) {
	m.focused = focused
}

// IsFocused returns true if the modal is focused
func (m *Modal) IsFocused() bool {
	return m.focused
}

// SetPosition sets the modal's position
func (m *Modal) SetPosition(x, y int) {
	m.x = x
	m.y = y
}

// GetPosition returns the modal's position
func (m *Modal) GetPosition() (int, int) {
	return m.x, m.y
}

// CalculatePosition calculates the centered position for the modal
func (m *Modal) CalculatePosition(containerWidth, containerHeight int) {
	if !m.config.CenterX && !m.config.CenterY {
		return
	}

	// Get the rendered view to calculate dimensions
	view := m.View()
	lines := lipgloss.Height(view)
	width := lipgloss.Width(view)

	// Apply max width/height if configured
	if m.config.MaxWidth > 0 && width > m.config.MaxWidth {
		width = m.config.MaxWidth
	}
	if m.config.MaxHeight > 0 && lines > m.config.MaxHeight {
		lines = m.config.MaxHeight
	}

	// Calculate centered position
	if m.config.CenterX {
		m.x = (containerWidth - width) / 2
		if m.x < 0 {
			m.x = 0
		}
	}

	if m.config.CenterY {
		m.y = (containerHeight - lines) / 2
		if m.y < 0 {
			m.y = 0
		}
	}
}

// ShouldCloseOnEsc returns true if ESC should close this modal
func (m *Modal) ShouldCloseOnEsc() bool {
	return m.config.CloseOnEsc
}

// ShouldCloseOnBackdrop returns true if clicking backdrop should close this modal
func (m *Modal) ShouldCloseOnBackdrop() bool {
	return m.config.CloseOnBackdrop
}

// HasFocusTrap returns true if focus should be trapped in this modal
func (m *Modal) HasFocusTrap() bool {
	return m.config.TrapFocus
}
