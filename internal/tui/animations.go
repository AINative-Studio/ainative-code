package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AnimationType represents the type of animation to display
type AnimationType int

const (
	AnimationNone AnimationType = iota
	AnimationLoading
	AnimationThinking
	AnimationProcessing
	AnimationSuccess
	AnimationError
)

// AnimationState manages all animation state
type AnimationState struct {
	Spinner       spinner.Model
	AnimationType AnimationType
	Message       string
	Progress      float64 // 0.0 to 1.0 for progress bars
	StartTime     time.Time
	Visible       bool
}

// NewAnimationState creates a new animation state with default settings
func NewAnimationState() *AnimationState {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &AnimationState{
		Spinner:       s,
		AnimationType: AnimationNone,
		Visible:       false,
		Progress:      0.0,
	}
}

// StartAnimation starts an animation with the given type and message
func (a *AnimationState) StartAnimation(animType AnimationType, message string) {
	a.AnimationType = animType
	a.Message = message
	a.Visible = true
	a.StartTime = time.Now()
	a.Progress = 0.0

	// Set appropriate spinner style based on animation type
	switch animType {
	case AnimationLoading:
		a.Spinner.Spinner = spinner.Dot
		a.Spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Blue
	case AnimationThinking:
		a.Spinner.Spinner = spinner.MiniDot
		a.Spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("141")) // Purple
	case AnimationProcessing:
		a.Spinner.Spinner = spinner.Globe
		a.Spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
	case AnimationSuccess:
		a.Spinner.Spinner = spinner.Points
		a.Spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
	case AnimationError:
		a.Spinner.Spinner = spinner.Meter
		a.Spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red
	}
}

// StopAnimation stops the current animation
func (a *AnimationState) StopAnimation() {
	a.Visible = false
	a.AnimationType = AnimationNone
	a.Message = ""
	a.Progress = 0.0
}

// SetProgress updates the progress value (0.0 to 1.0)
func (a *AnimationState) SetProgress(progress float64) {
	if progress < 0.0 {
		progress = 0.0
	}
	if progress > 1.0 {
		progress = 1.0
	}
	a.Progress = progress
}

// UpdateMessage updates the animation message
func (a *AnimationState) UpdateMessage(message string) {
	a.Message = message
}

// GetElapsedTime returns the time elapsed since animation started
func (a *AnimationState) GetElapsedTime() time.Duration {
	if !a.Visible {
		return 0
	}
	return time.Since(a.StartTime)
}

// Render renders the animation based on current state
func (a *AnimationState) Render() string {
	if !a.Visible {
		return ""
	}

	var style lipgloss.Style
	var icon string

	switch a.AnimationType {
	case AnimationLoading:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
		icon = a.Spinner.View()
	case AnimationThinking:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
		icon = a.Spinner.View()
	case AnimationProcessing:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		icon = a.Spinner.View()
	case AnimationSuccess:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		icon = "✓"
	case AnimationError:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		icon = "✗"
	default:
		return ""
	}

	// Build the animation display
	result := fmt.Sprintf("%s %s", icon, a.Message)

	// Add elapsed time for long operations
	if a.AnimationType == AnimationLoading || a.AnimationType == AnimationProcessing {
		elapsed := a.GetElapsedTime()
		if elapsed > 3*time.Second {
			result += fmt.Sprintf(" (%ds)", int(elapsed.Seconds()))
		}
	}

	// Add progress bar if progress is set
	if a.Progress > 0 {
		result += " " + RenderProgressBar(a.Progress, 20)
	}

	return style.Render(result)
}

// RenderProgressBar renders a progress bar with the given completion percentage
func RenderProgressBar(progress float64, width int) string {
	if width < 2 {
		width = 2
	}

	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}

	empty := width - filled

	bar := "["
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	bar += "]"

	percentage := int(progress * 100)
	return fmt.Sprintf("%s %d%%", bar, percentage)
}

// LoadingIndicator renders a simple loading indicator
func LoadingIndicator(message string) string {
	dots := spinner.Dot
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	return style.Render(fmt.Sprintf("%s %s", dots.Frames[0], message))
}

// ThinkingIndicator renders a thinking indicator with animated dots
func ThinkingIndicator(tick int) string {
	dots := []string{"   ", ".  ", ".. ", "..."}
	dotIndex := tick % len(dots)

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
	return style.Render(fmt.Sprintf("Thinking%s", dots[dotIndex]))
}

// StreamingIndicator renders a streaming indicator
func StreamingIndicator(tick int) string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frameIndex := tick % len(frames)

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	return style.Render(fmt.Sprintf("%s Streaming...", frames[frameIndex]))
}

// SuccessIndicator renders a success indicator
func SuccessIndicator(message string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	return style.Render(fmt.Sprintf("✓ %s", message))
}

// ErrorIndicator renders an error indicator
func ErrorIndicator(message string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	return style.Render(fmt.Sprintf("✗ %s", message))
}

// PulseAnimation creates a pulsing effect for text
func PulseAnimation(text string, tick int) string {
	colors := []lipgloss.Color{
		lipgloss.Color("12"),
		lipgloss.Color("14"),
		lipgloss.Color("10"),
		lipgloss.Color("11"),
	}

	colorIndex := (tick / 2) % len(colors)
	style := lipgloss.NewStyle().Foreground(colors[colorIndex])
	return style.Render(text)
}

// TypewriterEffect simulates a typewriter effect for text
func TypewriterEffect(text string, position int) string {
	if position >= len(text) {
		return text
	}
	return text[:position]
}

// FadeInEffect creates a fade-in effect using different brightness levels
func FadeInEffect(text string, step int, maxSteps int) string {
	if step >= maxSteps {
		return text
	}

	// Map step to brightness (darker to lighter)
	brightness := []lipgloss.Color{
		lipgloss.Color("240"),
		lipgloss.Color("242"),
		lipgloss.Color("244"),
		lipgloss.Color("246"),
		lipgloss.Color("248"),
		lipgloss.Color("250"),
		lipgloss.Color("252"),
		lipgloss.Color("15"),
	}

	index := (step * len(brightness)) / maxSteps
	if index >= len(brightness) {
		index = len(brightness) - 1
	}

	style := lipgloss.NewStyle().Foreground(brightness[index])
	return style.Render(text)
}

// AnimationTickMsg is sent to update animation state
type AnimationTickMsg struct {
	Time time.Time
}

// AnimationCmd returns a command that sends animation ticks
func AnimationCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return AnimationTickMsg{Time: t}
	})
}
