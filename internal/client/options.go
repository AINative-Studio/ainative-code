package client

import (
	"net/http"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth"
)

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithAuthClient sets the authentication client for JWT token management.
func WithAuthClient(authClient auth.Client) Option {
	return func(c *Client) {
		c.authClient = authClient
	}
}

// WithBaseURL sets the base URL for API requests.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retry attempts for failed requests.
func WithMaxRetries(maxRetries int) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// RequestOption is a functional option for per-request configuration.
type RequestOption func(*requestOptions)

// requestOptions contains per-request options.
type requestOptions struct {
	headers      map[string]string
	queryParams  map[string]string
	skipAuth     bool
	disableRetry bool
}

// WithHeader adds a custom header to the request.
func WithHeader(key, value string) RequestOption {
	return func(opts *requestOptions) {
		if opts.headers == nil {
			opts.headers = make(map[string]string)
		}
		opts.headers[key] = value
	}
}

// WithHeaders adds multiple custom headers to the request.
func WithHeaders(headers map[string]string) RequestOption {
	return func(opts *requestOptions) {
		if opts.headers == nil {
			opts.headers = make(map[string]string)
		}
		for k, v := range headers {
			opts.headers[k] = v
		}
	}
}

// WithQueryParam adds a query parameter to the request.
func WithQueryParam(key, value string) RequestOption {
	return func(opts *requestOptions) {
		if opts.queryParams == nil {
			opts.queryParams = make(map[string]string)
		}
		opts.queryParams[key] = value
	}
}

// WithQueryParams adds multiple query parameters to the request.
func WithQueryParams(params map[string]string) RequestOption {
	return func(opts *requestOptions) {
		if opts.queryParams == nil {
			opts.queryParams = make(map[string]string)
		}
		for k, v := range params {
			opts.queryParams[k] = v
		}
	}
}

// WithSkipAuth skips JWT token injection for this request.
func WithSkipAuth() RequestOption {
	return func(opts *requestOptions) {
		opts.skipAuth = true
	}
}

// WithDisableRetry disables retry logic for this request.
func WithDisableRetry() RequestOption {
	return func(opts *requestOptions) {
		opts.disableRetry = true
	}
}
