package provider

import "errors"

// ProviderInfo represents an LLM provider with its capabilities
type ProviderInfo struct {
	Name                    string
	DisplayName             string
	SupportsVision          bool
	SupportsFunctionCalling bool
	SupportsStreaming       bool
	MaxTokens               int
	LowCreditWarning        bool
}

// User represents a user with their credit balance and tier
type User struct {
	Email   string
	Credits int
	Tier    string
}

// SelectionRequest represents requirements for provider selection
type SelectionRequest struct {
	Model                   string
	RequiresVision          bool
	RequiresFunctionCalling bool
	RequiresStreaming       bool
}

// Error definitions
var (
	ErrInsufficientCredits = errors.New("insufficient credits")
	ErrNoProviderAvailable = errors.New("no provider available")
)
