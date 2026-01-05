package client

import (
	"net/http"
	"time"
)

// Options contains configuration options for the HTTP client.
type Options struct {
	// BaseURL is the base URL for all API requests
	BaseURL string

	// Timeout is the default timeout for HTTP requests
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// RetryBackoff is the initial backoff duration for retries
	RetryBackoff time.Duration

	// EnableLogging enables request/response logging
	EnableLogging bool

	// HTTPClient is the underlying HTTP client (optional)
	HTTPClient *http.Client

	// UserAgent is the User-Agent header value
	UserAgent string
}

// DefaultOptions returns the default client options.
func DefaultOptions() *Options {
	return &Options{
		BaseURL:       "https://api.ainative.studio",
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryBackoff:  1 * time.Second,
		EnableLogging: true,
		HTTPClient:    nil, // Will create default client
		UserAgent:     "ainative-code/1.0.0",
	}
}

// Response represents an HTTP response with parsed data.
type Response struct {
	// StatusCode is the HTTP status code
	StatusCode int

	// Headers are the response headers
	Headers http.Header

	// Body is the raw response body
	Body []byte

	// Duration is the time taken for the request
	Duration time.Duration
}

// RequestOptions contains per-request options.
type RequestOptions struct {
	// Headers are additional headers to include in the request
	Headers map[string]string

	// QueryParams are query parameters to include in the request
	QueryParams map[string]string

	// Timeout overrides the default client timeout for this request
	Timeout time.Duration

	// SkipAuth skips JWT token injection for this request
	SkipAuth bool

	// DisableRetry disables retry logic for this request
	DisableRetry bool
}
