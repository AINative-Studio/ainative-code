package provider

import (
	"context"
	"fmt"
)

// Selector handles intelligent provider selection based on user preferences,
// credit balance, and capability requirements
type Selector struct {
	providers       []string
	userPreference  string
	creditThreshold int
	fallbackEnabled bool
	capabilities    map[string]ProviderInfo
}

// SelectorOption is a functional option for configuring Selector
type SelectorOption func(*Selector)

// WithProviders sets the available providers for selection
func WithProviders(providers ...string) SelectorOption {
	return func(s *Selector) {
		s.providers = providers
	}
}

// WithUserPreference sets the user's preferred provider
func WithUserPreference(pref string) SelectorOption {
	return func(s *Selector) {
		s.userPreference = pref
	}
}

// WithCreditThreshold sets the credit threshold for low credit warnings
func WithCreditThreshold(threshold int) SelectorOption {
	return func(s *Selector) {
		s.creditThreshold = threshold
	}
}

// WithFallback enables or disables fallback to alternative providers
func WithFallback(enabled bool) SelectorOption {
	return func(s *Selector) {
		s.fallbackEnabled = enabled
	}
}

// NewSelector creates a new provider selector with the given options
func NewSelector(opts ...SelectorOption) *Selector {
	s := &Selector{
		capabilities:    ProviderCapabilities,
		creditThreshold: 50,
		fallbackEnabled: true,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Select intelligently selects a provider based on user preferences,
// credit balance, and capability requirements
func (s *Selector) Select(ctx context.Context, user *User, req ...*SelectionRequest) (*ProviderInfo, error) {
	// Check credits first
	if user != nil && user.Credits == 0 {
		return nil, ErrInsufficientCredits
	}

	// If no providers configured, return error
	if len(s.providers) == 0 {
		return nil, ErrNoProviderAvailable
	}

	// Extract request if provided
	var selectionReq *SelectionRequest
	if len(req) > 0 {
		selectionReq = req[0]
	}

	// Try user preference first if set
	if s.userPreference != "" && s.IsAvailable(s.userPreference) {
		provider := s.capabilities[s.userPreference]

		// Check if preferred provider meets capability requirements
		if selectionReq != nil && !s.meetsRequirements(&provider, selectionReq) {
			// Preferred provider doesn't meet requirements, fallback to capability-based selection
			return s.selectByCapabilities(user, selectionReq)
		}

		// Apply credit warning if needed
		if user != nil && user.Credits < s.creditThreshold {
			provider.LowCreditWarning = true
		}

		return &provider, nil
	}

	// If request has capability requirements, select by capabilities
	if selectionReq != nil {
		return s.selectByCapabilities(user, selectionReq)
	}

	// Default to first available provider
	provider := s.capabilities[s.providers[0]]
	if user != nil && user.Credits < s.creditThreshold {
		provider.LowCreditWarning = true
	}

	return &provider, nil
}

// IsAvailable checks if a provider is available in the configured providers list
func (s *Selector) IsAvailable(name string) bool {
	for _, p := range s.providers {
		if p == name {
			return true
		}
	}
	return false
}

// meetsRequirements checks if a provider meets the given capability requirements
func (s *Selector) meetsRequirements(p *ProviderInfo, req *SelectionRequest) bool {
	if req.RequiresVision && !p.SupportsVision {
		return false
	}
	if req.RequiresFunctionCalling && !p.SupportsFunctionCalling {
		return false
	}
	if req.RequiresStreaming && !p.SupportsStreaming {
		return false
	}
	return true
}

// selectByCapabilities selects a provider that meets the capability requirements
func (s *Selector) selectByCapabilities(user *User, req *SelectionRequest) (*ProviderInfo, error) {
	for _, name := range s.providers {
		provider := s.capabilities[name]
		if s.meetsRequirements(&provider, req) {
			// Apply credit warning if needed
			if user != nil && user.Credits < s.creditThreshold {
				provider.LowCreditWarning = true
			}
			return &provider, nil
		}
	}

	return nil, fmt.Errorf("no provider meets requirements: %w", ErrNoProviderAvailable)
}
