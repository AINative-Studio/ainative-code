package components

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
)

// AnimationConfig configures animation behavior
type AnimationConfig struct {
	Duration         time.Duration // Animation duration (used to calculate spring params)
	FPS              int           // Frames per second (default: 60)
	AngularFrequency float64       // Spring angular frequency (default: 6.0)
	DampingRatio     float64       // Spring damping ratio (default: 1.0 for critical damping)
	Loop             bool          // Repeat animation when complete
	Reverse          bool          // Reverse direction on loop
}

// DefaultAnimationConfig returns sensible defaults for animations
func DefaultAnimationConfig() AnimationConfig {
	return AnimationConfig{
		Duration:         300 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 6.0,
		DampingRatio:     1.0, // Critical damping for smooth motion
		Loop:             false,
		Reverse:          false,
	}
}

// AnimatedComponent wraps any component with animation support
type AnimatedComponent struct {
	id          string            // Unique identifier for this animation
	component   Component         // The wrapped component
	spring      harmonica.Spring  // Physics-based animation spring
	config      AnimationConfig   // Animation configuration
	from        float64           // Starting value
	to          float64           // Target value
	current     float64           // Current animated value
	velocity    float64           // Current velocity
	isAnimating bool              // Whether animation is active
	reversed    bool              // Current direction (for reverse mode)
	frameTime   time.Duration     // Time per frame (calculated from FPS)
	startTime   time.Time         // When the animation started
}

// AnimationTickMsg is sent on each animation frame update
type AnimationTickMsg struct {
	ID    string  // Animation ID
	Value float64 // Current interpolated value
	Time  time.Time
}

// AnimationCompleteMsg is sent when an animation completes
type AnimationCompleteMsg struct {
	ID    string  // Animation ID
	Value float64 // Final value
}

// NewAnimatedComponent creates a new animated component wrapper
func NewAnimatedComponent(component Component, config AnimationConfig) *AnimatedComponent {
	// Set defaults
	if config.FPS <= 0 {
		config.FPS = 60
	}
	if config.AngularFrequency <= 0 {
		config.AngularFrequency = 6.0
	}
	if config.DampingRatio <= 0 {
		config.DampingRatio = 1.0
	}
	if config.Duration <= 0 {
		config.Duration = 300 * time.Millisecond
	}

	deltaTime := harmonica.FPS(config.FPS)

	ac := &AnimatedComponent{
		id:        fmt.Sprintf("anim-%d", time.Now().UnixNano()),
		component: component,
		config:    config,
		current:   0,
		velocity:  0,
		frameTime: time.Second / time.Duration(config.FPS),
		spring:    harmonica.NewSpring(deltaTime, config.AngularFrequency, config.DampingRatio),
	}

	return ac
}

// NewAnimatedComponentWithID creates an animated component with a custom ID
func NewAnimatedComponentWithID(id string, component Component, config AnimationConfig) *AnimatedComponent {
	ac := NewAnimatedComponent(component, config)
	ac.id = id
	return ac
}

// StartAnimation begins animating from the current value to the target
func (a *AnimatedComponent) StartAnimation(from, to float64) tea.Cmd {
	a.from = from
	a.to = to
	a.current = from
	a.velocity = 0
	a.isAnimating = true
	a.reversed = false
	a.startTime = time.Now()

	// Start ticking
	return a.tick()
}

// StopAnimation stops the animation
func (a *AnimatedComponent) StopAnimation() tea.Cmd {
	a.isAnimating = false
	return nil
}

// IsAnimating returns whether the animation is currently running
func (a *AnimatedComponent) IsAnimating() bool {
	return a.isAnimating
}

// GetValue returns the current animated value
func (a *AnimatedComponent) GetValue() float64 {
	return a.current
}

// GetID returns the animation's unique identifier
func (a *AnimatedComponent) GetID() string {
	return a.id
}

// SetValue directly sets the current value (bypasses animation)
func (a *AnimatedComponent) SetValue(value float64) {
	a.current = value
}

// tick generates the next animation frame
func (a *AnimatedComponent) tick() tea.Cmd {
	return tea.Tick(a.frameTime, func(t time.Time) tea.Msg {
		return AnimationTickMsg{
			ID:    a.id,
			Value: a.current,
			Time:  t,
		}
	})
}

// Update handles animation ticks and updates the wrapped component
func (a *AnimatedComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case AnimationTickMsg:
		// Only process ticks for this animation
		if msg.ID != a.id {
			break
		}

		if !a.isAnimating {
			break
		}

		// Update spring animation using harmonica's physics
		a.current, a.velocity = a.spring.Update(a.current, a.velocity, a.to)

		// Check if animation is complete (position near target and velocity near zero)
		elapsed := time.Since(a.startTime)
		threshold := 0.01
		isComplete := elapsed >= a.config.Duration ||
			(abs(a.current-a.to) < threshold && abs(a.velocity) < threshold)

		if isComplete {
			a.current = a.to // Snap to final value
			a.velocity = 0

			if a.config.Loop {
				// Handle looping
				if a.config.Reverse {
					// Swap from/to for reverse
					a.reversed = !a.reversed
					a.from, a.to = a.to, a.from
				} else {
					// Reset to start
					a.current = a.from
				}

				a.velocity = 0
				a.startTime = time.Now()
				cmds = append(cmds, a.tick())
			} else {
				// Animation complete
				a.isAnimating = false
				cmds = append(cmds, func() tea.Msg {
					return AnimationCompleteMsg{
						ID:    a.id,
						Value: a.current,
					}
				})
			}
		} else {
			// Continue animation
			cmds = append(cmds, a.tick())
		}
	}

	// Pass message to wrapped component
	if a.component != nil {
		var cmd tea.Cmd
		a.component, cmd = a.component.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return a, tea.Batch(cmds...)
}

// abs returns the absolute value of x
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Init initializes both the animation and wrapped component
func (a *AnimatedComponent) Init() tea.Cmd {
	if a.component != nil {
		return a.component.Init()
	}
	return nil
}

// View renders the wrapped component
func (a *AnimatedComponent) View() string {
	if a.component != nil {
		return a.component.View()
	}
	return ""
}

// Component returns the wrapped component (useful for accessing specific interfaces)
func (a *AnimatedComponent) Component() Component {
	return a.component
}

// Progress returns the animation progress as a percentage (0.0 to 1.0)
func (a *AnimatedComponent) Progress() float64 {
	if a.to == a.from {
		return 1.0
	}
	return (a.current - a.from) / (a.to - a.from)
}

// Reset resets the animation to its initial state
func (a *AnimatedComponent) Reset() {
	a.isAnimating = false
	a.current = a.from
	a.velocity = 0
	a.reversed = false
	a.startTime = time.Time{}
}
