package components

import (
	"fmt"
	"math"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AnimatedSpinner demonstrates using AnimatedComponent for a smooth rotating spinner
type AnimatedSpinner struct {
	animation *AnimatedComponent
	frames    []string
	message   string
	style     lipgloss.Style
}

// NewAnimatedSpinner creates a new animated spinner component
func NewAnimatedSpinner(message string) *AnimatedSpinner {
	// Create spinner frames
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	// Create a dummy component (the animation wrapper will handle the logic)
	baseComponent := &dummyComponent{}

	// Create animation with spinner preset (continuous loop)
	animation := NewAnimatedComponent(baseComponent, Spinner)

	return &AnimatedSpinner{
		animation: animation,
		frames:    frames,
		message:   message,
		style:     lipgloss.NewStyle().Foreground(lipgloss.Color("141")),
	}
}

// Start begins the spinner animation
func (s *AnimatedSpinner) Start() tea.Cmd {
	return s.animation.StartAnimation(0, float64(len(s.frames)))
}

// Stop stops the spinner animation
func (s *AnimatedSpinner) Stop() tea.Cmd {
	return s.animation.StopAnimation()
}

// Init initializes the animated spinner
func (s *AnimatedSpinner) Init() tea.Cmd {
	return s.animation.Init()
}

// Update handles animation updates
func (s *AnimatedSpinner) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
	updatedComponent, cmd := s.animation.Update(msg)
	if animComp, ok := updatedComponent.(*AnimatedComponent); ok {
		s.animation = animComp
	}
	return s, cmd
}

// View renders the spinner with smooth animation
func (s *AnimatedSpinner) View() string {
	if !s.animation.IsAnimating() {
		return ""
	}

	// Get current animation value (0 to len(frames))
	value := s.animation.GetValue()

	// Use modulo to cycle through frames
	frameIndex := int(value) % len(s.frames)
	if frameIndex < 0 {
		frameIndex = 0
	}
	if frameIndex >= len(s.frames) {
		frameIndex = len(s.frames) - 1
	}

	frame := s.frames[frameIndex]

	if s.message != "" {
		return s.style.Render(fmt.Sprintf("%s %s", frame, s.message))
	}

	return s.style.Render(frame)
}

// AnimatedProgress demonstrates a smooth progress bar animation
type AnimatedProgress struct {
	animation *AnimatedComponent
	width     int
	style     lipgloss.Style
	label     string
}

// NewAnimatedProgress creates a new animated progress bar
func NewAnimatedProgress(width int, label string) *AnimatedProgress {
	baseComponent := &dummyComponent{}
	animation := NewAnimatedComponent(baseComponent, Smooth)

	return &AnimatedProgress{
		animation: animation,
		width:     width,
		label:     label,
		style:     lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
	}
}

// SetProgress animates the progress bar to the target percentage (0.0 to 1.0)
func (p *AnimatedProgress) SetProgress(target float64) tea.Cmd {
	current := p.animation.GetValue()
	return p.animation.StartAnimation(current, target)
}

// Init initializes the progress bar
func (p *AnimatedProgress) Init() tea.Cmd {
	return p.animation.Init()
}

// Update handles animation updates
func (p *AnimatedProgress) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
	updatedComponent, cmd := p.animation.Update(msg)
	if animComp, ok := updatedComponent.(*AnimatedComponent); ok {
		p.animation = animComp
	}
	return p, cmd
}

// View renders the animated progress bar
func (p *AnimatedProgress) View() string {
	progress := p.animation.GetValue()

	filled := int(progress * float64(p.width))
	if filled > p.width {
		filled = p.width
	}
	empty := p.width - filled

	bar := "["
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	bar += "]"

	percentage := int(progress * 100)
	result := fmt.Sprintf("%s %d%%", bar, percentage)

	if p.label != "" {
		result = fmt.Sprintf("%s: %s", p.label, result)
	}

	return p.style.Render(result)
}

// AnimatedPulse demonstrates a pulsing animation effect
type AnimatedPulse struct {
	animation *AnimatedComponent
	text      string
	baseColor lipgloss.Color
}

// NewAnimatedPulse creates a new pulsing text animation
func NewAnimatedPulse(text string, color lipgloss.Color) *AnimatedPulse {
	baseComponent := &dummyComponent{}
	animation := NewAnimatedComponent(baseComponent, Pulse)

	return &AnimatedPulse{
		animation: animation,
		text:      text,
		baseColor: color,
	}
}

// Start begins the pulse animation
func (p *AnimatedPulse) Start() tea.Cmd {
	return p.animation.StartAnimation(0.3, 1.0) // Pulse between 30% and 100% opacity
}

// Init initializes the pulse animation
func (p *AnimatedPulse) Init() tea.Cmd {
	return p.animation.Init()
}

// Update handles animation updates
func (p *AnimatedPulse) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
	updatedComponent, cmd := p.animation.Update(msg)
	if animComp, ok := updatedComponent.(*AnimatedComponent); ok {
		p.animation = animComp
	}
	return p, cmd
}

// View renders the pulsing text
func (p *AnimatedPulse) View() string {
	opacity := p.animation.GetValue()

	// Interpolate color brightness based on opacity
	// Convert opacity to a brightness level (240-255 range)
	brightness := int(240 + (15 * opacity))
	if brightness > 255 {
		brightness = 255
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", brightness)))
	return style.Render(p.text)
}

// AnimatedSlide demonstrates a sliding animation
type AnimatedSlide struct {
	animation *AnimatedComponent
	content   string
	width     int
}

// NewAnimatedSlide creates a new sliding content animation
func NewAnimatedSlide(content string, width int) *AnimatedSlide {
	baseComponent := &dummyComponent{}
	animation := NewAnimatedComponent(baseComponent, SlideIn)

	return &AnimatedSlide{
		animation: animation,
		content:   content,
		width:     width,
	}
}

// SlideIn animates the content sliding in from the left
func (s *AnimatedSlide) SlideIn() tea.Cmd {
	return s.animation.StartAnimation(float64(-s.width), 0)
}

// SlideOut animates the content sliding out to the right
func (s *AnimatedSlide) SlideOut() tea.Cmd {
	s.animation.config = SlideOut
	return s.animation.StartAnimation(0, float64(s.width))
}

// Init initializes the slide animation
func (s *AnimatedSlide) Init() tea.Cmd {
	return s.animation.Init()
}

// Update handles animation updates
func (s *AnimatedSlide) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
	updatedComponent, cmd := s.animation.Update(msg)
	if animComp, ok := updatedComponent.(*AnimatedComponent); ok {
		s.animation = animComp
	}
	return s, cmd
}

// View renders the sliding content
func (s *AnimatedSlide) View() string {
	offset := int(s.animation.GetValue())

	if offset < -s.width || offset > s.width {
		return "" // Content is off-screen
	}

	// Create padding based on offset
	if offset > 0 {
		// Sliding right (out)
		visible := s.width - offset
		if visible <= 0 {
			return ""
		}
		if visible > len(s.content) {
			visible = len(s.content)
		}
		return s.content[:visible]
	} else if offset < 0 {
		// Sliding left (in)
		padding := -offset
		if padding >= s.width {
			return ""
		}
		spaces := ""
		for i := 0; i < padding; i++ {
			spaces += " "
		}
		return spaces + s.content
	}

	return s.content
}

// AnimatedRotation demonstrates a rotating animation (for spinners)
type AnimatedRotation struct {
	animation *AnimatedComponent
	radius    float64
	char      string
}

// NewAnimatedRotation creates a new rotating animation
func NewAnimatedRotation(radius float64, char string) *AnimatedRotation {
	baseComponent := &dummyComponent{}
	config := RotationAnimation(1000, true) // 1 second per rotation, loop
	animation := NewAnimatedComponent(baseComponent, config)

	return &AnimatedRotation{
		animation: animation,
		radius:    radius,
		char:      char,
	}
}

// Start begins the rotation animation
func (r *AnimatedRotation) Start() tea.Cmd {
	return r.animation.StartAnimation(0, 2*math.Pi) // Full circle in radians
}

// Init initializes the rotation animation
func (r *AnimatedRotation) Init() tea.Cmd {
	return r.animation.Init()
}

// Update handles animation updates
func (r *AnimatedRotation) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
	updatedComponent, cmd := r.animation.Update(msg)
	if animComp, ok := updatedComponent.(*AnimatedComponent); ok {
		r.animation = animComp
	}
	return r, cmd
}

// View renders the rotating element
func (r *AnimatedRotation) View() string {
	angle := r.animation.GetValue()

	// Calculate position on circle
	x := int(r.radius * math.Cos(angle))
	y := int(r.radius * math.Sin(angle))

	// Simple representation - in real usage, you'd position this in 2D space
	return fmt.Sprintf("Rotation: %.1f° (x:%d, y:%d)", angle*180/math.Pi, x, y)
}

// dummyComponent is a minimal component implementation for animation-only components
type dummyComponent struct{}

func (d *dummyComponent) Init() tea.Cmd                           { return nil }
func (d *dummyComponent) Update(msg tea.Msg) (Component, tea.Cmd) { return d, nil }
func (d *dummyComponent) View() string                            { return "" }
