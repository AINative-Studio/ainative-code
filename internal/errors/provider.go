package errors

import (
	"fmt"
	"time"
)

// ProviderError represents errors related to external AI provider interactions
type ProviderError struct {
	*BaseError
	ProviderName string
	Model        string
	RequestID    string
	StatusCode   int
	RetryAfter   *time.Duration
}

// NewProviderError creates a new provider error
func NewProviderError(code ErrorCode, message string, providerName string) *ProviderError {
	baseErr := newError(code, message, SeverityHigh, false)
	return &ProviderError{
		BaseError:    baseErr,
		ProviderName: providerName,
	}
}

// NewProviderUnavailableError creates an error for unavailable providers
func NewProviderUnavailableError(providerName string, cause error) *ProviderError {
	msg := fmt.Sprintf("Provider '%s' is currently unavailable", providerName)
	userMsg := fmt.Sprintf("Service temporarily unavailable: The AI provider '%s' is not responding. Please try again in a few moments.", providerName)

	baseErr := newError(ErrCodeProviderUnavailable, msg, SeverityHigh, true)
	baseErr.cause = cause
	baseErr.userMsg = userMsg

	return &ProviderError{
		BaseError:    baseErr,
		ProviderName: providerName,
	}
}

// NewProviderTimeoutError creates an error for provider request timeouts
func NewProviderTimeoutError(providerName, model string, timeout time.Duration) *ProviderError {
	msg := fmt.Sprintf("Request to provider '%s' (model: %s) timed out after %v", providerName, model, timeout)
	userMsg := fmt.Sprintf("Request timeout: The AI provider took too long to respond. Please try again.")

	err := NewProviderError(ErrCodeProviderTimeout, msg, providerName)
	err.userMsg = userMsg
	err.Model = model
	err.retryable = true
	return err
}

// NewProviderRateLimitError creates an error for rate limit exceeded
func NewProviderRateLimitError(providerName string, retryAfter time.Duration) *ProviderError {
	msg := fmt.Sprintf("Rate limit exceeded for provider '%s'", providerName)
	userMsg := fmt.Sprintf("Rate limit reached: Too many requests to the AI provider. Please wait %v before trying again.", retryAfter.Round(time.Second))

	err := NewProviderError(ErrCodeProviderRateLimit, msg, providerName)
	err.userMsg = userMsg
	err.RetryAfter = &retryAfter
	err.retryable = true
	return err
}

// NewProviderInvalidResponseError creates an error for invalid provider responses
func NewProviderInvalidResponseError(providerName string, reason string, cause error) *ProviderError {
	msg := fmt.Sprintf("Invalid response from provider '%s': %s", providerName, reason)
	userMsg := "The AI provider returned an unexpected response. Please try again or contact support if the issue persists."

	baseErr := newError(ErrCodeProviderInvalidResponse, msg, SeverityMedium, true)
	baseErr.cause = cause
	baseErr.userMsg = userMsg

	return &ProviderError{
		BaseError:    baseErr,
		ProviderName: providerName,
	}
}

// NewProviderNotFoundError creates an error for provider not found
func NewProviderNotFoundError(providerName string) *ProviderError {
	msg := fmt.Sprintf("Provider '%s' not found or not configured", providerName)
	userMsg := fmt.Sprintf("Provider error: '%s' is not available. Please check your configuration.", providerName)

	err := NewProviderError(ErrCodeProviderNotFound, msg, providerName)
	err.userMsg = userMsg
	err.severity = SeverityCritical
	return err
}

// WithModel sets the model name
func (e *ProviderError) WithModel(model string) *ProviderError {
	e.Model = model
	return e
}

// WithRequestID sets the request ID for tracking
func (e *ProviderError) WithRequestID(requestID string) *ProviderError {
	e.RequestID = requestID
	return e
}

// WithStatusCode sets the HTTP status code
func (e *ProviderError) WithStatusCode(statusCode int) *ProviderError {
	e.StatusCode = statusCode
	return e
}

// ShouldRetry determines if the provider error should be retried
func (e *ProviderError) ShouldRetry() bool {
	return e.retryable
}

// GetRetryDelay returns the suggested retry delay
func (e *ProviderError) GetRetryDelay() time.Duration {
	if e.RetryAfter != nil {
		return *e.RetryAfter
	}
	// Default exponential backoff can be implemented in the retry strategy
	return 0
}
