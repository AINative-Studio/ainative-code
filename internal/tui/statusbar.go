package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/AINative-studio/ainative-code/internal/tui/components"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusBarState represents the state of the status bar
type StatusBarState struct {
	Provider       string
	Model          string
	TokensUsed     int
	TokensTotal    int
	SessionStart   time.Time
	ConnectionOK   bool
	CurrentMode    string // "chat", "help", "settings", etc.
	CustomMessage  string
	ShowKeyHints   bool
	AnimationTick  int // For animated elements
}

// NewStatusBarState creates a new status bar state with defaults
func NewStatusBarState() *StatusBarState {
	return &StatusBarState{
		Provider:      "Unknown",
		Model:         "Unknown",
		TokensUsed:    0,
		TokensTotal:   0,
		SessionStart:  time.Now(),
		ConnectionOK:  true,
		CurrentMode:   "chat",
		ShowKeyHints:  true,
		AnimationTick: 0,
	}
}

// SetProvider sets the AI provider name
func (s *StatusBarState) SetProvider(provider string) {
	s.Provider = provider
}

// SetModel sets the AI model name
func (s *StatusBarState) SetModel(model string) {
	s.Model = model
}

// SetTokens sets the token usage
func (s *StatusBarState) SetTokens(used, total int) {
	s.TokensUsed = used
	s.TokensTotal = total
}

// AddTokens adds to the token count
func (s *StatusBarState) AddTokens(tokens int) {
	s.TokensUsed += tokens
}

// SetConnectionStatus sets the connection status
func (s *StatusBarState) SetConnectionStatus(ok bool) {
	s.ConnectionOK = ok
}

// SetMode sets the current mode
func (s *StatusBarState) SetMode(mode string) {
	s.CurrentMode = mode
}

// SetCustomMessage sets a custom message to display
func (s *StatusBarState) SetCustomMessage(message string) {
	s.CustomMessage = message
}

// ClearCustomMessage clears the custom message
func (s *StatusBarState) ClearCustomMessage() {
	s.CustomMessage = ""
}

// ToggleKeyHints toggles the display of keyboard hints
func (s *StatusBarState) ToggleKeyHints() {
	s.ShowKeyHints = !s.ShowKeyHints
}

// IncrementAnimationTick increments the animation tick counter
func (s *StatusBarState) IncrementAnimationTick() {
	s.AnimationTick++
}

// GetSessionDuration returns the duration of the current session
func (s *StatusBarState) GetSessionDuration() time.Duration {
	return time.Since(s.SessionStart)
}

// RenderStatusBar renders the complete status bar
func (s *StatusBarState) RenderStatusBar(width int, isStreaming bool, hasError bool) string {
	if width < 40 {
		return s.renderCompactStatusBar(width, isStreaming, hasError)
	}

	var sections []string

	// Left section: Status indicator and mode
	leftSection := s.renderLeftSection(isStreaming, hasError)
	sections = append(sections, leftSection)

	// Middle section: Provider, model, and tokens
	if width >= 80 {
		middleSection := s.renderMiddleSection()
		if middleSection != "" {
			sections = append(sections, middleSection)
		}
	}

	// Right section: Session info and key hints
	rightSection := s.renderRightSection(width >= 100)
	sections = append(sections, rightSection)

	return s.layoutSections(sections, width)
}

// renderLeftSection renders the left section of the status bar
func (s *StatusBarState) renderLeftSection(isStreaming bool, hasError bool) string {
	var parts []string

	// Connection status indicator
	if !s.ConnectionOK {
		disconnectedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)
		parts = append(parts, disconnectedStyle.Render("⚠ Disconnected"))
	} else if isStreaming {
		// Animated streaming indicator
		streamingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)
		parts = append(parts, streamingStyle.Render(StreamingIndicator(s.AnimationTick)))
	} else if hasError {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)
		parts = append(parts, errorStyle.Render("✗ Error"))
	} else {
		readyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))
		parts = append(parts, readyStyle.Render("● Ready"))
	}

	// Current mode
	if s.CurrentMode != "chat" {
		modeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true)
		parts = append(parts, modeStyle.Render(fmt.Sprintf("[%s]", s.CurrentMode)))
	}

	// Custom message if set
	if s.CustomMessage != "" {
		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))
		parts = append(parts, messageStyle.Render(s.CustomMessage))
	}

	return strings.Join(parts, " ")
}

// renderMiddleSection renders the middle section of the status bar
func (s *StatusBarState) renderMiddleSection() string {
	var parts []string

	// Provider and model
	if s.Provider != "Unknown" || s.Model != "Unknown" {
		providerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("14"))
		modelInfo := fmt.Sprintf("%s:%s", s.Provider, s.Model)
		parts = append(parts, providerStyle.Render(modelInfo))
	}

	// Token usage
	if s.TokensTotal > 0 {
		tokenStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("13"))

		percentage := float64(s.TokensUsed) / float64(s.TokensTotal) * 100
		tokenInfo := fmt.Sprintf("Tokens: %d/%d (%.0f%%)", s.TokensUsed, s.TokensTotal, percentage)

		// Warn if token usage is high
		if percentage > 80 {
			tokenStyle = tokenStyle.Foreground(lipgloss.Color("11")) // Yellow
		}
		if percentage > 95 {
			tokenStyle = tokenStyle.Foreground(lipgloss.Color("9")) // Red
		}

		parts = append(parts, tokenStyle.Render(tokenInfo))
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " | ")
}

// renderRightSection renders the right section of the status bar
func (s *StatusBarState) renderRightSection(showExtended bool) string {
	var parts []string

	// Session duration
	if showExtended {
		duration := s.GetSessionDuration()
		durationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

		var durationStr string
		if duration < time.Minute {
			durationStr = fmt.Sprintf("Session: %ds", int(duration.Seconds()))
		} else if duration < time.Hour {
			durationStr = fmt.Sprintf("Session: %dm", int(duration.Minutes()))
		} else {
			durationStr = fmt.Sprintf("Session: %dh %dm",
				int(duration.Hours()),
				int(duration.Minutes())%60)
		}

		parts = append(parts, durationStyle.Render(durationStr))
	}

	// Keyboard hints
	if s.ShowKeyHints {
		hintStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

		if s.CurrentMode == "help" {
			parts = append(parts, hintStyle.Render("Press ESC to close"))
		} else {
			parts = append(parts, hintStyle.Render("? for help"))
		}
	}

	return strings.Join(parts, " | ")
}

// renderCompactStatusBar renders a compact version for narrow terminals
func (s *StatusBarState) renderCompactStatusBar(width int, isStreaming bool, hasError bool) string {
	var icon string
	var color lipgloss.Color

	if !s.ConnectionOK {
		icon = "⚠"
		color = lipgloss.Color("9")
	} else if isStreaming {
		icon = "●"
		color = lipgloss.Color("10")
	} else if hasError {
		icon = "✗"
		color = lipgloss.Color("9")
	} else {
		icon = "●"
		color = lipgloss.Color("10")
	}

	style := lipgloss.NewStyle().Foreground(color)
	statusText := fmt.Sprintf("%s %s", icon, s.CurrentMode)

	if s.ShowKeyHints {
		statusText += " | ? help"
	}

	return style.Render(statusText)
}

// layoutSections distributes sections across the status bar width
func (s *StatusBarState) layoutSections(sections []string, width int) string {
	if len(sections) == 0 {
		return ""
	}

	// Calculate total content width
	totalWidth := 0
	for _, section := range sections {
		totalWidth += lipgloss.Width(section)
	}

	// If content is wider than available space, truncate
	if totalWidth >= width {
		return s.truncateSections(sections, width)
	}

	// Distribute sections with spacing
	if len(sections) == 1 {
		padding := strings.Repeat(" ", width-lipgloss.Width(sections[0]))
		return sections[0] + padding
	}

	if len(sections) == 2 {
		// Left and right alignment
		spacingWidth := width - totalWidth
		if spacingWidth < 0 {
			spacingWidth = 0
		}
		spacing := strings.Repeat(" ", spacingWidth)
		return sections[0] + spacing + sections[1]
	}

	// Three sections: left, center, right
	leftWidth := lipgloss.Width(sections[0])
	centerWidth := lipgloss.Width(sections[1])
	rightWidth := lipgloss.Width(sections[2])

	// Calculate spacing
	remainingSpace := width - leftWidth - centerWidth - rightWidth
	if remainingSpace < 2 {
		return s.truncateSections(sections, width)
	}

	leftSpacing := remainingSpace / 2
	rightSpacing := remainingSpace - leftSpacing

	return sections[0] + strings.Repeat(" ", leftSpacing) +
		sections[1] + strings.Repeat(" ", rightSpacing) + sections[2]
}

// truncateSections truncates sections to fit within width
func (s *StatusBarState) truncateSections(sections []string, width int) string {
	if len(sections) == 0 {
		return ""
	}

	// Always show at least the first section (status indicator)
	result := sections[0]
	currentWidth := lipgloss.Width(result)

	for i := 1; i < len(sections) && currentWidth < width; i++ {
		sectionWidth := lipgloss.Width(sections[i])
		if currentWidth+sectionWidth+3 <= width { // +3 for " | "
			result += " | " + sections[i]
			currentWidth += sectionWidth + 3
		} else {
			break
		}
	}

	// Pad to full width
	if currentWidth < width {
		result += strings.Repeat(" ", width-currentWidth)
	}

	return result
}

// RenderMinimalStatusBar renders a minimal status bar (single line indicator)
func RenderMinimalStatusBar(isStreaming bool, hasError bool) string {
	var icon string
	var color lipgloss.Color
	var text string

	if hasError {
		icon = "✗"
		color = lipgloss.Color("9")
		text = "Error"
	} else if isStreaming {
		icon = "●"
		color = lipgloss.Color("10")
		text = "Streaming"
	} else {
		icon = "●"
		color = lipgloss.Color("10")
		text = "Ready"
	}

	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	return style.Render(fmt.Sprintf("%s %s", icon, text))
}

// StatusBarComponent is a wrapper that makes StatusBarState implement the Component interface.
// This is NON-BREAKING - it only adds new methods and doesn't modify existing behavior.
type StatusBarComponent struct {
	state *StatusBarState
	*components.ComponentAdapter
	isStreaming bool
	hasError    bool
}

// NewStatusBarComponent creates a new status bar component wrapper around StatusBarState.
func NewStatusBarComponent(state *StatusBarState) *StatusBarComponent {
	adapter := components.NewComponentAdapter()
	return &StatusBarComponent{
		state:            state,
		ComponentAdapter: adapter,
		isStreaming:      false,
		hasError:         false,
	}
}

// Init implements Component interface.
func (s *StatusBarComponent) Init() tea.Cmd {
	return s.ComponentAdapter.Init()
}

// Update implements Component interface.
func (s *StatusBarComponent) Update(msg tea.Msg) (components.Component, tea.Cmd) {
	// Handle window size messages
	if wsMsg, ok := msg.(tea.WindowSizeMsg); ok {
		s.SetSize(wsMsg.Width, 1) // Status bar is always 1 line tall
	}
	return s, nil
}

// View implements Component interface.
func (s *StatusBarComponent) View() string {
	if !s.IsVisible() {
		return ""
	}
	width, _ := s.GetSize()
	if width == 0 {
		width = 80 // Default width
	}
	return s.state.RenderStatusBar(width, s.isStreaming, s.hasError)
}

// SetStreaming sets the streaming state for rendering.
func (s *StatusBarComponent) SetStreaming(streaming bool) {
	s.isStreaming = streaming
}

// SetError sets the error state for rendering.
func (s *StatusBarComponent) SetError(hasError bool) {
	s.hasError = hasError
}

// GetState returns the underlying StatusBarState.
func (s *StatusBarComponent) GetState() *StatusBarState {
	return s.state
}

// AsComponent returns the status bar as a Component interface.
// This allows existing code to work with the new interface without changes.
func (s *StatusBarState) AsComponent() components.Component {
	return NewStatusBarComponent(s)
}

// Ensure StatusBarComponent implements Component interface
var _ components.Component = (*StatusBarComponent)(nil)
var _ components.Sizeable = (*StatusBarComponent)(nil)
var _ components.Stateful = (*StatusBarComponent)(nil)
var _ components.Lifecycle = (*StatusBarComponent)(nil)
