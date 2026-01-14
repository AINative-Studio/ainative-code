# Animation Component Wrapper

A reusable animation component wrapper that makes it easy to add smooth 60 FPS animations to any component using physics-based spring animations powered by charmbracelet/harmonica.

## Features

- **Smooth 60 FPS animations** - Frame timing at 16.67ms
- **Physics-based motion** - Uses harmonica.Spring for natural movement
- **Component integration** - Wraps any existing component
- **Pre-defined transitions** - FadeIn, FadeOut, SlideIn, Spring, Bounce, Pulse, Spinner, and more
- **Loop and reverse support** - Create continuous or ping-pong animations
- **Custom configurations** - Fine-tune angular frequency and damping ratio

## Quick Start

### Basic Usage

```go
// Wrap any component with animation
baseComponent := &MyComponent{}
config := components.FadeIn
animatedComponent := components.NewAnimatedComponent(baseComponent, config)

// Start the animation
cmd := animatedComponent.StartAnimation(0.0, 1.0) // From 0 to 1
```

### Pre-defined Transitions

```go
// Fade in animation
fadeIn := components.FadeInComponent(myComponent)
cmd := fadeIn.StartAnimation(0.0, 1.0)

// Slide in from left
slideIn := components.SlideInComponent(myComponent, -100, 0)
cmd := slideIn.StartAnimation(-100, 0)

// Spring bounce effect
spring := components.SpringComponent(myComponent)
cmd := spring.StartAnimation(0.0, 100.0)

// Continuous spinner
spinner := components.SpinnerComponent(myComponent)
cmd := spinner.StartAnimation(0.0, 360.0)
```

## Available Transitions

### FadeIn
- Duration: 200ms
- Angular Frequency: 8.0 (fast response)
- Damping Ratio: 1.2 (slightly over-damped for smooth finish)
- Use for: Smooth fade-in effects

### FadeOut
- Duration: 150ms
- Angular Frequency: 10.0 (very fast response)
- Damping Ratio: 1.0 (critical damping)
- Use for: Quick fade-out effects

### SlideIn
- Duration: 250ms
- Angular Frequency: 7.0 (medium response)
- Damping Ratio: 1.0 (critical damping)
- Use for: Sliding content into view

### Spring
- Duration: 400ms
- Angular Frequency: 5.0 (slower for visible spring)
- Damping Ratio: 0.5 (under-damped for bounce)
- Use for: Bouncy spring effects

### Bounce
- Duration: 500ms
- Angular Frequency: 6.0 (medium response)
- Damping Ratio: 0.3 (very under-damped for multiple bounces)
- Use for: Multiple bounce effects

### Smooth
- Duration: 300ms
- Angular Frequency: 6.0 (standard response)
- Damping Ratio: 1.5 (over-damped for very smooth motion)
- Use for: Subtle, smooth animations with no overshoot

### Spinner
- Duration: 1000ms
- Loop: true
- Use for: Loading indicators and continuous rotation

### Pulse
- Duration: 600ms
- Loop: true
- Reverse: true
- Use for: Attention-grabbing pulsing effects

### Snap
- Duration: 100ms
- Angular Frequency: 12.0 (very fast)
- Use for: Quick, snappy animations

## Custom Configurations

```go
// Create a custom animation
config := components.CustomTransition(
    500*time.Millisecond,  // Duration
    60,                     // FPS
    5.0,                    // Angular frequency (controls speed)
    0.8,                    // Damping ratio (controls oscillation)
    false,                  // Loop
    false,                  // Reverse
)

animatedComponent := components.NewAnimatedComponent(myComponent, config)
```

## Understanding Spring Parameters

### Angular Frequency
Controls animation speed:
- **1.0-3.0**: Slow, gentle motion
- **4.0-8.0**: Standard responsive motion
- **9.0-15.0**: Fast, snappy motion

### Damping Ratio
Controls oscillation:
- **< 1.0**: Under-damped (bouncy, with overshoot)
- **= 1.0**: Critically damped (smooth, no overshoot)
- **> 1.0**: Over-damped (slow approach, no overshoot)

## Integration with Component Interface

AnimatedComponent implements the Component interface:

```go
type Component interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (Component, tea.Cmd)
    View() string
}
```

This means you can use it anywhere a Component is expected.

## Animation Messages

### AnimationTickMsg
Sent on each animation frame:
```go
type AnimationTickMsg struct {
    ID    string  // Animation ID
    Value float64 // Current interpolated value
    Time  time.Time
}
```

### AnimationCompleteMsg
Sent when animation completes:
```go
type AnimationCompleteMsg struct {
    ID    string  // Animation ID
    Value float64 // Final value
}
```

## Demo Components

See `animation_demo.go` for complete examples:

### AnimatedSpinner
Smooth rotating spinner with customizable frames:
```go
spinner := components.NewAnimatedSpinner("Loading...")
cmd := spinner.Start()
```

### AnimatedProgress
Animated progress bar that smoothly transitions:
```go
progress := components.NewAnimatedProgress(20, "Download")
cmd := progress.SetProgress(0.75) // Animate to 75%
```

### AnimatedPulse
Pulsing text effect:
```go
pulse := components.NewAnimatedPulse("Attention", lipgloss.Color("10"))
cmd := pulse.Start()
```

### AnimatedSlide
Sliding content in and out:
```go
slide := components.NewAnimatedSlide("Hello", 50)
cmd := slide.SlideIn()  // Slide in from left
cmd = slide.SlideOut()  // Slide out to right
```

### AnimatedRotation
Rotating element around a circle:
```go
rotation := components.NewAnimatedRotation(10.0, "●")
cmd := rotation.Start()
```

## Methods

### StartAnimation(from, to float64) tea.Cmd
Starts animating from `from` value to `to` value.

### StopAnimation() tea.Cmd
Stops the animation.

### IsAnimating() bool
Returns true if animation is currently running.

### GetValue() float64
Returns the current animated value.

### SetValue(value float64)
Directly sets the current value (bypasses animation).

### Progress() float64
Returns animation progress as a percentage (0.0 to 1.0).

### Reset()
Resets the animation to its initial state.

## Performance

- **60 FPS**: Frame timing at 16.67ms
- **Non-blocking**: Uses tea.Tick for frame updates
- **Efficient**: Physics-based spring calculations are optimized
- **Smooth**: Harmonica ensures consistent, natural motion

## Testing

Run the test suite:
```bash
go test ./internal/tui/components/... -v
```

All 63 tests passing:
- 26 animation tests
- 10 demo component tests
- 27 other component tests

## Best Practices

1. **Choose appropriate transitions**: Use FadeIn/FadeOut for opacity, SlideIn for position, Spring for bouncy effects
2. **Set reasonable durations**: 100-500ms for most UI animations
3. **Use critical damping (1.0) for smooth motion**: Avoids overshoot
4. **Use under-damping (< 1.0) for playful effects**: Creates bounce
5. **Enable loop for continuous animations**: Like spinners and progress indicators
6. **Use unique IDs**: When managing multiple animations simultaneously

## Examples

### Smooth Loading Indicator
```go
spinner := components.NewAnimatedSpinner("Thinking...")
spinner.Start()
// Will continuously loop through frames at 60 FPS
```

### Progress Bar with Smooth Updates
```go
progress := components.NewAnimatedProgress(30, "Upload")
progress.SetProgress(0.0)   // Start at 0%
progress.SetProgress(0.5)   // Smoothly animate to 50%
progress.SetProgress(1.0)   // Smoothly animate to 100%
```

### Attention-Grabbing Pulse
```go
pulse := components.NewAnimatedPulse("New Message!", lipgloss.Color("10"))
pulse.Start()
// Will continuously pulse between 30% and 100% opacity
```

### Slide-In Panel
```go
panel := components.NewAnimatedSlide(panelContent, terminalWidth)
panel.SlideIn()  // Animate from off-screen left to on-screen
```

## Integration with Existing Code

The animation wrapper is designed to work seamlessly with existing components:

```go
// Wrap existing component
existingComponent := &MyExistingComponent{}
animated := components.FadeInComponent(existingComponent)

// Use in your Update loop
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.animatedComponent, cmd = m.animatedComponent.Update(msg)
    return m, cmd
}

// Render in View
func (m Model) View() string {
    return m.animatedComponent.View()
}
```

## Future Enhancements

- Additional easing functions
- Chained animations
- Animation groups (multiple animations in sequence)
- Bezier curve support
- Custom interpolation functions

---

Built with ❤️ using [charmbracelet/harmonica](https://github.com/charmbracelet/harmonica)
