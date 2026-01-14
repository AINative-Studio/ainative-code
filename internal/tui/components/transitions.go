package components

import (
	"time"
)

// Pre-defined animation configs for common use cases
// These use different spring parameters to achieve various animation styles:
// - AngularFrequency: Controls animation speed (higher = faster response)
// - DampingRatio: Controls oscillation (< 1 = bouncy, 1 = smooth, > 1 = slow)

var (
	// FadeIn smoothly fades in from 0 to 1 with quick response
	FadeIn = AnimationConfig{
		Duration:         200 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 8.0,  // Fast response
		DampingRatio:     1.2,  // Slightly over-damped for smooth finish
		Loop:             false,
		Reverse:          false,
	}

	// FadeOut smoothly fades out from 1 to 0
	FadeOut = AnimationConfig{
		Duration:         150 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 10.0, // Very fast response
		DampingRatio:     1.0,  // Critical damping
		Loop:             false,
		Reverse:          false,
	}

	// SlideIn slides content in with smooth motion
	SlideIn = AnimationConfig{
		Duration:         250 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 7.0, // Medium response
		DampingRatio:     1.0, // Critical damping for smooth motion
		Loop:             false,
		Reverse:          false,
	}

	// SlideOut slides content out quickly
	SlideOut = AnimationConfig{
		Duration:         200 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 8.0, // Fast response
		DampingRatio:     1.0, // Critical damping
		Loop:             false,
		Reverse:          false,
	}

	// Spring creates a bouncy spring effect
	Spring = AnimationConfig{
		Duration:         400 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 5.0, // Slower response for visible spring
		DampingRatio:     0.5, // Under-damped for bounce
		Loop:             false,
		Reverse:          false,
	}

	// Bounce creates a bouncing effect with multiple oscillations
	Bounce = AnimationConfig{
		Duration:         500 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 6.0, // Medium response
		DampingRatio:     0.3, // Very under-damped for multiple bounces
		Loop:             false,
		Reverse:          false,
	}

	// Smooth provides a smooth, subtle animation with no overshoot
	Smooth = AnimationConfig{
		Duration:         300 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 6.0, // Standard response
		DampingRatio:     1.5, // Over-damped for very smooth motion
		Loop:             false,
		Reverse:          false,
	}

	// Spinner creates a continuous looping animation for spinners
	Spinner = AnimationConfig{
		Duration:         1000 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 6.0, // Standard response
		DampingRatio:     1.0, // Critical damping
		Loop:             true,
		Reverse:          false,
	}

	// Pulse creates a pulsing effect that reverses
	Pulse = AnimationConfig{
		Duration:         600 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 5.0, // Slower for visible pulse
		DampingRatio:     1.0, // Critical damping
		Loop:             true,
		Reverse:          true,
	}

	// Snap creates a quick, snappy animation
	Snap = AnimationConfig{
		Duration:         100 * time.Millisecond,
		FPS:              60,
		AngularFrequency: 12.0, // Very fast response
		DampingRatio:     1.0,  // Critical damping
		Loop:             false,
		Reverse:          false,
	}
)

// TransitionPresets provides easy access to all transition presets
var TransitionPresets = map[string]AnimationConfig{
	"fadeIn":   FadeIn,
	"fadeOut":  FadeOut,
	"slideIn":  SlideIn,
	"slideOut": SlideOut,
	"spring":   Spring,
	"bounce":   Bounce,
	"smooth":   Smooth,
	"spinner":  Spinner,
	"pulse":    Pulse,
	"snap":     Snap,
}

// Helper functions for common animation patterns

// FadeInComponent wraps a component with a fade-in animation
func FadeInComponent(c Component) *AnimatedComponent {
	ac := NewAnimatedComponent(c, FadeIn)
	return ac
}

// FadeOutComponent wraps a component with a fade-out animation
func FadeOutComponent(c Component) *AnimatedComponent {
	ac := NewAnimatedComponent(c, FadeOut)
	return ac
}

// SlideInComponent wraps a component with a slide-in animation
// from: starting position (e.g., -100 for off-screen left)
// to: ending position (e.g., 0 for on-screen)
func SlideInComponent(c Component, from, to float64) *AnimatedComponent {
	ac := NewAnimatedComponent(c, SlideIn)
	return ac
}

// SlideOutComponent wraps a component with a slide-out animation
func SlideOutComponent(c Component, from, to float64) *AnimatedComponent {
	ac := NewAnimatedComponent(c, SlideOut)
	return ac
}

// SpringComponent wraps a component with a spring animation
func SpringComponent(c Component) *AnimatedComponent {
	ac := NewAnimatedComponent(c, Spring)
	return ac
}

// BounceComponent wraps a component with a bounce animation
func BounceComponent(c Component) *AnimatedComponent {
	ac := NewAnimatedComponent(c, Bounce)
	return ac
}

// SmoothComponent wraps a component with a smooth animation
func SmoothComponent(c Component) *AnimatedComponent {
	ac := NewAnimatedComponent(c, Smooth)
	return ac
}

// SpinnerComponent wraps a component with a continuous spinner animation
// This is ideal for loading indicators and progress spinners
func SpinnerComponent(c Component) *AnimatedComponent {
	ac := NewAnimatedComponent(c, Spinner)
	return ac
}

// PulseComponent wraps a component with a pulsing animation
func PulseComponent(c Component) *AnimatedComponent {
	ac := NewAnimatedComponent(c, Pulse)
	return ac
}

// SnapComponent wraps a component with a quick snap animation
func SnapComponent(c Component) *AnimatedComponent {
	ac := NewAnimatedComponent(c, Snap)
	return ac
}

// GetTransitionPreset returns a preset animation configuration by name
func GetTransitionPreset(name string) (AnimationConfig, bool) {
	config, exists := TransitionPresets[name]
	return config, exists
}

// CustomTransition creates a custom animation configuration
func CustomTransition(duration time.Duration, fps int, angularFreq, dampingRatio float64, loop, reverse bool) AnimationConfig {
	return AnimationConfig{
		Duration:         duration,
		FPS:              fps,
		AngularFrequency: angularFreq,
		DampingRatio:     dampingRatio,
		Loop:             loop,
		Reverse:          reverse,
	}
}

// QuickTransition creates a fast animation (100ms)
func QuickTransition() AnimationConfig {
	return Snap
}

// MediumTransition creates a medium-speed animation (300ms)
func MediumTransition() AnimationConfig {
	return Smooth
}

// SlowTransition creates a slow animation (500ms)
func SlowTransition() AnimationConfig {
	return Bounce
}

// RotationAnimation creates a smooth rotation animation for spinners
func RotationAnimation(duration time.Duration, loop bool) AnimationConfig {
	return AnimationConfig{
		Duration:         duration,
		FPS:              60,
		AngularFrequency: 1.0, // Linear animation
		DampingRatio:     1.0, // No damping
		Loop:             loop,
		Reverse:          false,
	}
}
