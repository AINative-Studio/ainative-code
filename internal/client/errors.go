package client

import (
	"errors"
	"fmt"
)

var (
	// ErrHTTPRequest indicates an HTTP request failed.
	ErrHTTPRequest = errors.New("HTTP request failed")

	// ErrHTTPResponse indicates an HTTP response error.
	ErrHTTPResponse = errors.New("HTTP response error")

	// ErrUnauthorized indicates a 401 Unauthorized response.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates a 403 Forbidden response.
	ErrForbidden = errors.New("forbidden")

	// ErrNotFound indicates a 404 Not Found response.
	ErrNotFound = errors.New("not found")

	// ErrRateLimited indicates a 429 Too Many Requests response.
	ErrRateLimited = errors.New("rate limited")

	// ErrServerError indicates a 5xx server error response.
	ErrServerError = errors.New("server error")

	// ErrNoAuthClient indicates no auth client is configured.
	ErrNoAuthClient = errors.New("no auth client configured")

	// ErrTokenRefreshFailed indicates token refresh failed.
	ErrTokenRefreshFailed = errors.New("token refresh failed")

	// ErrMaxRetriesExceeded indicates maximum retry attempts exceeded.
	ErrMaxRetriesExceeded = errors.New("maximum retry attempts exceeded")

	// ErrInvalidURL indicates an invalid URL.
	ErrInvalidURL = errors.New("invalid URL")

	// ErrInvalidRequest indicates an invalid request.
	ErrInvalidRequest = errors.New("invalid request")
)

// HTTPError represents an HTTP error with status code and message.
type HTTPError struct {
	StatusCode int
	Message    string
	Body       []byte
	Err        error
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("HTTP %d: %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// Unwrap returns the underlying error.
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates a new HTTP error.
func NewHTTPError(statusCode int, message string, body []byte, err error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
		Body:       body,
		Err:        err,
	}
}

// IsRetryable returns true if the error is retryable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types that are retryable
	if errors.Is(err, ErrRateLimited) ||
		errors.Is(err, ErrServerError) {
		return true
	}

	// Check for HTTP errors
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		// Retry on 429, 500, 502, 503, 504
		return httpErr.StatusCode == 429 ||
			httpErr.StatusCode == 500 ||
			httpErr.StatusCode == 502 ||
			httpErr.StatusCode == 503 ||
			httpErr.StatusCode == 504
	}

	return false
}

// IsAuthError returns true if the error is an authentication error.
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific auth error types
	if errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrForbidden) {
		return true
	}

	// Check for HTTP errors
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == 401 || httpErr.StatusCode == 403
	}

	return false
}
