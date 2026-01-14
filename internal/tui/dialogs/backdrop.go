package dialogs

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BackdropStyle defines backdrop appearance and behavior
type BackdropStyle struct {
	Enabled   bool           // Whether backdrop is enabled
	Opacity   float64        // Opacity level (0.0 - 1.0)
	Color     lipgloss.Color // Backdrop color
	BlurChars string         // Characters to use for blur effect (e.g., "░▒▓")
}

// Pre-defined backdrop styles
var (
	// DarkBackdrop is a dark semi-transparent backdrop (60% opacity)
	DarkBackdrop = BackdropStyle{
		Enabled:   true,
		Opacity:   0.6,
		Color:     lipgloss.Color("#000000"),
		BlurChars: "",
	}

	// LightBackdrop is a light semi-transparent backdrop (40% opacity)
	LightBackdrop = BackdropStyle{
		Enabled:   true,
		Opacity:   0.4,
		Color:     lipgloss.Color("#FFFFFF"),
		BlurChars: "",
	}

	// BlurBackdrop is a backdrop with blur effect (70% opacity)
	BlurBackdrop = BackdropStyle{
		Enabled:   true,
		Opacity:   0.7,
		Color:     lipgloss.Color("#000000"),
		BlurChars: "░▒▓",
	}

	// NoBackdrop is an invisible backdrop (disabled)
	NoBackdrop = BackdropStyle{
		Enabled:   false,
		Opacity:   0.0,
		Color:     lipgloss.Color("#000000"),
		BlurChars: "",
	}

	// PurpleBackdrop is a purple-tinted backdrop matching AINative branding
	PurpleBackdrop = BackdropStyle{
		Enabled:   true,
		Opacity:   0.5,
		Color:     lipgloss.Color("#8B5CF6"),
		BlurChars: "",
	}

	// HeavyBlurBackdrop is a heavily blurred backdrop (80% opacity)
	HeavyBlurBackdrop = BackdropStyle{
		Enabled:   true,
		Opacity:   0.8,
		Color:     lipgloss.Color("#000000"),
		BlurChars: "░░▒▒▓▓",
	}
)

// BackdropRenderer creates semi-transparent overlays
type BackdropRenderer struct {
	width  int
	height int
	style  BackdropStyle
}

// NewBackdropRenderer creates a new backdrop renderer
func NewBackdropRenderer(width, height int, style BackdropStyle) *BackdropRenderer {
	return &BackdropRenderer{
		width:  width,
		height: height,
		style:  style,
	}
}

// Render creates the backdrop overlay
func (b *BackdropRenderer) Render() string {
	if !b.style.Enabled {
		return ""
	}

	// Calculate alpha channel based on opacity
	// Lipgloss doesn't support alpha transparency directly,
	// so we simulate it with character density or dimming
	var content string

	if b.style.BlurChars != "" {
		// Use blur characters to create a blur effect
		content = b.renderBlurred()
	} else {
		// Use solid color with dimming effect
		content = b.renderSolid()
	}

	return content
}

// renderSolid creates a solid backdrop with opacity simulation
func (b *BackdropRenderer) renderSolid() string {
	// Create a style with the backdrop color
	style := lipgloss.NewStyle().
		Width(b.width).
		Height(b.height).
		Background(b.style.Color)

	// For lower opacity, we use darker/lighter colors
	// This is a workaround since terminal colors don't support true alpha
	if b.style.Opacity < 1.0 {
		// Dim the color based on opacity
		dimmedColor := b.adjustColorOpacity(b.style.Color, b.style.Opacity)
		style = style.Background(dimmedColor)
	}

	return style.Render("")
}

// renderBlurred creates a backdrop with blur effect using characters
func (b *BackdropRenderer) renderBlurred() string {
	lines := make([]string, b.height)
	blurRunes := []rune(b.style.BlurChars)
	if len(blurRunes) == 0 {
		return b.renderSolid()
	}

	// Create pattern using blur characters
	for i := 0; i < b.height; i++ {
		var line strings.Builder
		for j := 0; j < b.width; j++ {
			// Use a pseudo-random pattern based on position
			// This creates a more natural blur effect
			charIndex := (i*b.width + j) % len(blurRunes)
			line.WriteRune(blurRunes[charIndex])
		}
		lines[i] = line.String()
	}

	content := strings.Join(lines, "\n")

	// Apply color styling
	style := lipgloss.NewStyle().
		Foreground(b.style.Color).
		Background(lipgloss.Color("#000000"))

	return style.Render(content)
}

// adjustColorOpacity adjusts color brightness based on opacity
// This is a helper to simulate opacity in terminals
func (b *BackdropRenderer) adjustColorOpacity(color lipgloss.Color, opacity float64) lipgloss.Color {
	// Map opacity to color brightness
	// For dark colors (black), lower opacity means lighter
	// For light colors (white), lower opacity means darker
	colorStr := string(color)

	if colorStr == "#000000" || colorStr == "#000" || colorStr == "0" {
		// Dark backdrop: lower opacity = lighter color
		if opacity >= 0.8 {
			return lipgloss.Color("#000000")
		} else if opacity >= 0.6 {
			return lipgloss.Color("#1A1A1A")
		} else if opacity >= 0.4 {
			return lipgloss.Color("#333333")
		} else if opacity >= 0.2 {
			return lipgloss.Color("#4D4D4D")
		}
		return lipgloss.Color("#666666")
	}

	if colorStr == "#FFFFFF" || colorStr == "#FFF" || colorStr == "15" {
		// Light backdrop: lower opacity = darker color
		if opacity >= 0.8 {
			return lipgloss.Color("#FFFFFF")
		} else if opacity >= 0.6 {
			return lipgloss.Color("#E5E5E5")
		} else if opacity >= 0.4 {
			return lipgloss.Color("#CCCCCC")
		} else if opacity >= 0.2 {
			return lipgloss.Color("#B3B3B3")
		}
		return lipgloss.Color("#999999")
	}

	// For other colors, adjust brightness proportionally
	// This is a simplified approach
	return color
}

// SetSize updates the backdrop dimensions
func (b *BackdropRenderer) SetSize(width, height int) {
	b.width = width
	b.height = height
}

// SetStyle updates the backdrop style
func (b *BackdropRenderer) SetStyle(style BackdropStyle) {
	b.style = style
}

// GetStyle returns the current backdrop style
func (b *BackdropRenderer) GetStyle() BackdropStyle {
	return b.style
}

// IsEnabled returns true if the backdrop is enabled
func (b *BackdropRenderer) IsEnabled() bool {
	return b.style.Enabled
}

// Enable enables the backdrop
func (b *BackdropRenderer) Enable() {
	b.style.Enabled = true
}

// Disable disables the backdrop
func (b *BackdropRenderer) Disable() {
	b.style.Enabled = false
}

// SetOpacity updates the backdrop opacity
func (b *BackdropRenderer) SetOpacity(opacity float64) {
	if opacity < 0.0 {
		opacity = 0.0
	}
	if opacity > 1.0 {
		opacity = 1.0
	}
	b.style.Opacity = opacity
}

// GetOpacity returns the current backdrop opacity
func (b *BackdropRenderer) GetOpacity() float64 {
	return b.style.Opacity
}
