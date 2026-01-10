// Package branding contains AINative Code branding constants and utilities.
// © 2024 AINative Studio. All rights reserved.
package branding

const (
	// Product Information
	ProductName        = "AINative Code"
	ProductTagline     = "AI-Native Development, Natively"
	ProductDescription = "A next-generation terminal-based AI coding assistant"
	CompanyName        = "AINative Studio"
	Copyright          = "© 2024 AINative Studio. All rights reserved."

	// URLs
	WebsiteURL       = "https://code.ainative.studio"
	DocsURL          = "https://docs.ainative.studio/code"
	SupportEmail     = "support@ainative.studio"
	RepositoryURL    = "https://github.com/AINative-studio/ainative-code"
	IssuesURL        = "https://github.com/AINative-studio/ainative-code/issues"
	DiscussionsURL   = "https://github.com/AINative-studio/ainative-code/discussions"

	// Binary and Configuration
	BinaryName       = "ainative-code"
	ConfigFileName   = ".ainative-code.yaml"
	ConfigDirName    = "ainative-code"
	DataDirName      = "ainative-code"

	// Environment Variable Prefix
	EnvPrefix = "AINATIVE_CODE"

	// Brand Colors (Hex)
	ColorPrimary   = "#6366F1" // Indigo
	ColorSecondary = "#8B5CF6" // Purple
	ColorSuccess   = "#10B981" // Green
	ColorError     = "#EF4444" // Red
	ColorAccent    = "#EC4899" // Pink
	ColorWarning   = "#F59E0B" // Amber
	ColorInfo      = "#3B82F6" // Blue

	// Service Endpoints (AINative Platform)
	AuthServiceURL   = "https://auth.ainative.studio"
	ZeroDBServiceURL = "https://api.zerodb.ainative.studio"
	DesignServiceURL = "https://design.ainative.studio"
	StrapiServiceURL = "https://strapi.ainative.studio"
	RLHFServiceURL   = "https://rlhf.ainative.studio"

	// Version (should be updated with each release)
	Version = "0.1.8"
)

// GetConfigPath returns the full path to the config directory.
// On Unix systems: ~/.config/ainative-code/
// On Windows: %APPDATA%\ainative-code\
func GetConfigPath() string {
	// This is a placeholder - actual implementation would use os.UserConfigDir()
	return "$HOME/.config/" + ConfigDirName
}

// GetDataPath returns the full path to the data directory.
// On Unix systems: ~/.local/share/ainative-code/
// On Windows: %LOCALAPPDATA%\ainative-code\
func GetDataPath() string {
	// This is a placeholder - actual implementation would use os.UserCacheDir()
	return "$HOME/.local/share/" + DataDirName
}

// GetFullProductName returns the complete product name with tagline.
func GetFullProductName() string {
	return ProductName + " - " + ProductTagline
}

// GetVersionString returns a formatted version string.
func GetVersionString() string {
	return ProductName + " v" + Version
}

// GetCopyrightNotice returns the complete copyright notice.
func GetCopyrightNotice() string {
	return Copyright
}

// GetWelcomeMessage returns a formatted welcome message for the CLI.
func GetWelcomeMessage() string {
	return `
╔════════════════════════════════════════════════════════════╗
║                     AINative Code                          ║
║            AI-Native Development, Natively                 ║
╚════════════════════════════════════════════════════════════╝

© 2024 AINative Studio. All rights reserved.
Version: ` + Version + `

Type 'ainative-code --help' for usage information.
Visit ` + DocsURL + ` for documentation.
`
}
