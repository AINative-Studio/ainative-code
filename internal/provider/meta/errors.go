package meta

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Error types
const (
	ErrTypeInvalidRequest     = "invalid_request_error"
	ErrTypeAuthentication     = "authentication_error"
	ErrTypePermission         = "permission_error"
	ErrTypeNotFound           = "not_found_error"
	ErrTypeRateLimit          = "rate_limit_error"
	ErrTypeAPI                = "api_error"
	ErrTypeTimeout            = "timeout_error"
	ErrTypeInvalidAPIKey      = "invalid_api_key"
	ErrTypeInsufficientQuota  = "insufficient_quota"
)

// MetaError represents an error from the Meta LLAMA API
type MetaError struct {
	StatusCode int
	Type       string
	Message    string
	Param      string
	Code       string
}

// Error implements the error interface
func (e *MetaError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("meta api error (status %d, type: %s, code: %s): %s",
			e.StatusCode, e.Type, e.Code, e.Message)
	}
	return fmt.Sprintf("meta api error (status %d, type: %s): %s",
		e.StatusCode, e.Type, e.Message)
}

// IsAuthenticationError checks if the error is an authentication error
func (e *MetaError) IsAuthenticationError() bool {
	return e.Type == ErrTypeAuthentication || e.Type == ErrTypeInvalidAPIKey
}

// IsRateLimitError checks if the error is a rate limit error
func (e *MetaError) IsRateLimitError() bool {
	return e.Type == ErrTypeRateLimit || e.StatusCode == 429
}

// IsQuotaError checks if the error is an insufficient quota error
func (e *MetaError) IsQuotaError() bool {
	return e.Type == ErrTypeInsufficientQuota
}

// IsRetryable checks if the error is retryable
func (e *MetaError) IsRetryable() bool {
	// Retry on rate limits, timeouts, and 5xx errors
	return e.IsRateLimitError() ||
		e.Type == ErrTypeTimeout ||
		e.StatusCode >= 500
}

// parseErrorResponse parses an error response from the Meta API
func parseErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &MetaError{
			StatusCode: resp.StatusCode,
			Type:       ErrTypeAPI,
			Message:    fmt.Sprintf("failed to read error response: %v", err),
		}
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return &MetaError{
			StatusCode: resp.StatusCode,
			Type:       ErrTypeAPI,
			Message:    fmt.Sprintf("failed to parse error response: %s", string(body)),
		}
	}

	return &MetaError{
		StatusCode: resp.StatusCode,
		Type:       errResp.Error.Type,
		Message:    errResp.Error.Message,
		Param:      errResp.Error.Param,
		Code:       errResp.Error.Code,
	}
}

// handleHTTPError handles HTTP errors and returns appropriate MetaError
func handleHTTPError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Try to parse structured error response
	if resp.Header.Get("Content-Type") == "application/json" {
		return parseErrorResponse(resp)
	}

	// Fallback to generic error based on status code
	var errType string
	var message string

	switch resp.StatusCode {
	case 400:
		errType = ErrTypeInvalidRequest
		message = "Invalid request parameters"
	case 401:
		errType = ErrTypeAuthentication
		message = "Invalid or missing API key"
	case 403:
		errType = ErrTypePermission
		message = "Permission denied"
	case 404:
		errType = ErrTypeNotFound
		message = "Resource not found"
	case 429:
		errType = ErrTypeRateLimit
		message = "Rate limit exceeded"
	case 500, 502, 503, 504:
		errType = ErrTypeAPI
		message = "Meta API server error"
	default:
		errType = ErrTypeAPI
		message = fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
	}

	return &MetaError{
		StatusCode: resp.StatusCode,
		Type:       errType,
		Message:    message,
	}
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(message string) *MetaError {
	return &MetaError{
		StatusCode: 401,
		Type:       ErrTypeAuthentication,
		Message:    message,
	}
}

// NewInvalidRequestError creates a new invalid request error
func NewInvalidRequestError(message string) *MetaError {
	return &MetaError{
		StatusCode: 400,
		Type:       ErrTypeInvalidRequest,
		Message:    message,
	}
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(message string) *MetaError {
	return &MetaError{
		StatusCode: 429,
		Type:       ErrTypeRateLimit,
		Message:    message,
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(message string) *MetaError {
	return &MetaError{
		StatusCode: 408,
		Type:       ErrTypeTimeout,
		Message:    message,
	}
}
