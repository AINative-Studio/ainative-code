package provider

import "fmt"

// ProviderError represents a provider-specific error with additional context
type ProviderError struct {
	Provider string
	Model    string
	Err      error
}

// Error implements the error interface
func (e *ProviderError) Error() string {
	if e.Model != "" {
		return fmt.Sprintf("provider %q (model %q): %v", e.Provider, e.Model, e.Err)
	}
	return fmt.Sprintf("provider %q: %v", e.Provider, e.Err)
}

// Unwrap returns the underlying error
func (e *ProviderError) Unwrap() error {
	return e.Err
}

// RateLimitError represents a rate limiting error
type RateLimitError struct {
	Provider     string
	RetryAfter   int // seconds until retry is allowed
	LimitType    string
	CurrentUsage int
	Limit        int
}

// Error implements the error interface
func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limit exceeded for provider %q: retry after %d seconds", e.Provider, e.RetryAfter)
	}
	if e.LimitType != "" {
		return fmt.Sprintf("rate limit exceeded for provider %q: %s limit reached (%d/%d)", e.Provider, e.LimitType, e.CurrentUsage, e.Limit)
	}
	return fmt.Sprintf("rate limit exceeded for provider %q", e.Provider)
}

// AuthenticationError represents an authentication failure
type AuthenticationError struct {
	Provider string
	Reason   string
}

// Error implements the error interface
func (e *AuthenticationError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("authentication failed for provider %q: %s", e.Provider, e.Reason)
	}
	return fmt.Sprintf("authentication failed for provider %q", e.Provider)
}

// ContextLengthError represents a context length exceeded error
type ContextLengthError struct {
	Provider      string
	Model         string
	RequestTokens int
	MaxTokens     int
}

// Error implements the error interface
func (e *ContextLengthError) Error() string {
	if e.RequestTokens > 0 && e.MaxTokens > 0 {
		return fmt.Sprintf("context length exceeded for provider %q model %q: %d tokens exceeds maximum of %d", e.Provider, e.Model, e.RequestTokens, e.MaxTokens)
	}
	return fmt.Sprintf("context length exceeded for provider %q model %q", e.Provider, e.Model)
}

// InvalidModelError represents an invalid model identifier error
type InvalidModelError struct {
	Provider       string
	Model          string
	SupportedModels []string
}

// Error implements the error interface
func (e *InvalidModelError) Error() string {
	if len(e.SupportedModels) > 0 {
		return fmt.Sprintf("invalid model %q for provider %q: supported models are %v", e.Model, e.Provider, e.SupportedModels)
	}
	return fmt.Sprintf("invalid model %q for provider %q", e.Model, e.Provider)
}

// StreamingNotSupportedError represents an error when streaming is not supported
type StreamingNotSupportedError struct {
	Provider string
	Model    string
	Reason   string
}

// Error implements the error interface
func (e *StreamingNotSupportedError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("streaming not supported for provider %q model %q: %s", e.Provider, e.Model, e.Reason)
	}
	if e.Model != "" {
		return fmt.Sprintf("streaming not supported for provider %q model %q", e.Provider, e.Model)
	}
	return fmt.Sprintf("streaming not supported for provider %q", e.Provider)
}

// Helper functions for creating common errors

// NewProviderError creates a new ProviderError
func NewProviderError(provider, model string, err error) *ProviderError {
	return &ProviderError{
		Provider: provider,
		Model:    model,
		Err:      err,
	}
}

// NewRateLimitError creates a new RateLimitError with retry information
func NewRateLimitError(provider string, retryAfter int) *RateLimitError {
	return &RateLimitError{
		Provider:   provider,
		RetryAfter: retryAfter,
	}
}

// NewRateLimitErrorWithDetails creates a detailed RateLimitError
func NewRateLimitErrorWithDetails(provider, limitType string, currentUsage, limit int) *RateLimitError {
	return &RateLimitError{
		Provider:     provider,
		LimitType:    limitType,
		CurrentUsage: currentUsage,
		Limit:        limit,
	}
}

// NewAuthenticationError creates a new AuthenticationError
func NewAuthenticationError(provider, reason string) *AuthenticationError {
	return &AuthenticationError{
		Provider: provider,
		Reason:   reason,
	}
}

// NewContextLengthError creates a new ContextLengthError
func NewContextLengthError(provider, model string, requestTokens, maxTokens int) *ContextLengthError {
	return &ContextLengthError{
		Provider:      provider,
		Model:         model,
		RequestTokens: requestTokens,
		MaxTokens:     maxTokens,
	}
}

// NewInvalidModelError creates a new InvalidModelError
func NewInvalidModelError(provider, model string, supportedModels []string) *InvalidModelError {
	return &InvalidModelError{
		Provider:       provider,
		Model:          model,
		SupportedModels: supportedModels,
	}
}

// NewStreamingNotSupportedError creates a new StreamingNotSupportedError
func NewStreamingNotSupportedError(provider, model, reason string) *StreamingNotSupportedError {
	return &StreamingNotSupportedError{
		Provider: provider,
		Model:    model,
		Reason:   reason,
	}
}
