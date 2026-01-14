package toast

import (
	"github.com/charmbracelet/lipgloss"
)

// Color constants for toast types
const (
	// Info colors (blue)
	InfoBorderColor      = "63"  // Blue
	InfoBackgroundColor  = "235" // Dark gray
	InfoTextColor        = "255" // White

	// Success colors (green)
	SuccessBorderColor      = "10"  // Green
	SuccessBackgroundColor  = "235" // Dark gray
	SuccessTextColor        = "255" // White

	// Warning colors (yellow/orange)
	WarningBorderColor      = "11"  // Yellow
	WarningBackgroundColor  = "235" // Dark gray
	WarningTextColor        = "255" // White

	// Error colors (red)
	ErrorBorderColor      = "9"   // Red
	ErrorBackgroundColor  = "235" // Dark gray
	ErrorTextColor        = "255" // White

	// Loading colors (purple/magenta)
	LoadingBorderColor      = "13"  // Magenta
	LoadingBackgroundColor  = "235" // Dark gray
	LoadingTextColor        = "255" // White
)

// GetToastStyle returns themed style for toast type
func GetToastStyle(toastType ToastType, width int) lipgloss.Style {
	base := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(width).
		Background(lipgloss.Color("235"))

	switch toastType {
	case ToastInfo:
		return base.
			BorderForeground(lipgloss.Color(InfoBorderColor)).
			Foreground(lipgloss.Color(InfoTextColor))

	case ToastSuccess:
		return base.
			BorderForeground(lipgloss.Color(SuccessBorderColor)).
			Foreground(lipgloss.Color(SuccessTextColor))

	case ToastWarning:
		return base.
			BorderForeground(lipgloss.Color(WarningBorderColor)).
			Foreground(lipgloss.Color(WarningTextColor))

	case ToastError:
		return base.
			BorderForeground(lipgloss.Color(ErrorBorderColor)).
			Foreground(lipgloss.Color(ErrorTextColor))

	case ToastLoading:
		return base.
			BorderForeground(lipgloss.Color(LoadingBorderColor)).
			Foreground(lipgloss.Color(LoadingTextColor))

	default:
		return base.
			BorderForeground(lipgloss.Color(InfoBorderColor)).
			Foreground(lipgloss.Color(InfoTextColor))
	}
}

// GetTitleStyle returns the style for toast titles
func GetTitleStyle(toastType ToastType) lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)

	switch toastType {
	case ToastInfo:
		return style.Foreground(lipgloss.Color(InfoBorderColor))
	case ToastSuccess:
		return style.Foreground(lipgloss.Color(SuccessBorderColor))
	case ToastWarning:
		return style.Foreground(lipgloss.Color(WarningBorderColor))
	case ToastError:
		return style.Foreground(lipgloss.Color(ErrorBorderColor))
	case ToastLoading:
		return style.Foreground(lipgloss.Color(LoadingBorderColor))
	default:
		return style.Foreground(lipgloss.Color(InfoBorderColor))
	}
}

// GetIconStyle returns the style for toast icons
func GetIconStyle(toastType ToastType) lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)

	switch toastType {
	case ToastInfo:
		return style.Foreground(lipgloss.Color(InfoBorderColor))
	case ToastSuccess:
		return style.Foreground(lipgloss.Color(SuccessBorderColor))
	case ToastWarning:
		return style.Foreground(lipgloss.Color(WarningBorderColor))
	case ToastError:
		return style.Foreground(lipgloss.Color(ErrorBorderColor))
	case ToastLoading:
		return style.Foreground(lipgloss.Color(LoadingBorderColor))
	default:
		return style.Foreground(lipgloss.Color(InfoBorderColor))
	}
}

// GetDismissStyle returns the style for the dismiss button
func GetDismissStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Bold(true)
}

// GetActionStyle returns the style for action buttons
func GetActionStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Underline(true)
}

// ToastIcons maps toast types to their default icons
var ToastIcons = map[ToastType]string{
	ToastInfo:    "ℹ",
	ToastSuccess: "✓",
	ToastWarning: "⚠",
	ToastError:   "✗",
	ToastLoading: "⟳",
}

// GetDefaultIcon returns the default icon for a toast type
func GetDefaultIcon(toastType ToastType) string {
	if icon, ok := ToastIcons[toastType]; ok {
		return icon
	}
	return "ℹ"
}

// SpinnerFrames contains the animation frames for loading spinners
var SpinnerFrames = []string{
	"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
}

// AlternativeSpinnerFrames provides an alternative spinner style
var AlternativeSpinnerFrames = []string{
	"◐", "◓", "◑", "◒",
}

// DotsSpinnerFrames provides a dots-based spinner
var DotsSpinnerFrames = []string{
	"⠋", "⠙", "⠚", "⠞", "⠖", "⠦", "⠴", "⠲", "⠳", "⠓",
}

// GetSpinnerFrame returns the appropriate spinner frame for the given progress
func GetSpinnerFrame(progress float64, style string) string {
	var frames []string

	switch style {
	case "dots":
		frames = DotsSpinnerFrames
	case "alternative":
		frames = AlternativeSpinnerFrames
	default:
		frames = SpinnerFrames
	}

	index := int(progress*float64(len(frames))) % len(frames)
	return frames[index]
}

// ToastStylePresets provides pre-configured style combinations
type ToastStylePreset struct {
	Name            string
	BorderStyle     lipgloss.Border
	BorderColor     string
	BackgroundColor string
	TextColor       string
}

// StylePresets contains various toast style presets
var StylePresets = map[string]ToastStylePreset{
	"default": {
		Name:            "Default",
		BorderStyle:     lipgloss.RoundedBorder(),
		BorderColor:     "63",
		BackgroundColor: "235",
		TextColor:       "255",
	},
	"minimal": {
		Name:            "Minimal",
		BorderStyle:     lipgloss.NormalBorder(),
		BorderColor:     "241",
		BackgroundColor: "235",
		TextColor:       "255",
	},
	"bold": {
		Name:            "Bold",
		BorderStyle:     lipgloss.ThickBorder(),
		BorderColor:     "255",
		BackgroundColor: "235",
		TextColor:       "255",
	},
	"double": {
		Name:            "Double",
		BorderStyle:     lipgloss.DoubleBorder(),
		BorderColor:     "63",
		BackgroundColor: "235",
		TextColor:       "255",
	},
}

// ApplyPreset applies a style preset to a toast style
func ApplyPreset(style lipgloss.Style, presetName string) lipgloss.Style {
	preset, ok := StylePresets[presetName]
	if !ok {
		preset = StylePresets["default"]
	}

	return style.
		Border(preset.BorderStyle).
		BorderForeground(lipgloss.Color(preset.BorderColor)).
		Background(lipgloss.Color(preset.BackgroundColor)).
		Foreground(lipgloss.Color(preset.TextColor))
}

// CalculateToastWidth calculates the appropriate toast width based on screen size
func CalculateToastWidth(screenWidth int) int {
	if screenWidth < 50 {
		return screenWidth - 4
	}
	if screenWidth < 80 {
		return 40
	}
	if screenWidth < 120 {
		return 50
	}
	return 60
}

// CalculateMaxToasts calculates the maximum number of toasts based on screen height
func CalculateMaxToasts(screenHeight int) int {
	// Each toast is approximately 4-5 lines tall
	toastHeight := 5

	// Leave room for other UI elements (status bar, input, etc.)
	availableHeight := screenHeight - 10

	if availableHeight < toastHeight {
		return 1
	}

	maxToasts := availableHeight / toastHeight
	if maxToasts > 5 {
		maxToasts = 5 // Cap at 5 to avoid clutter
	}
	if maxToasts < 1 {
		maxToasts = 1
	}

	return maxToasts
}
