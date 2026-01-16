package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

// Client represents an HTTP client for AINative platform API interactions.
type Client struct {
	httpClient *http.Client
	authClient auth.Client
	baseURL    string
	timeout    time.Duration
	maxRetries int
}

// New creates a new API client with the specified options.
func New(opts ...Option) *Client {
	client := &Client{
		timeout:    30 * time.Second,
		maxRetries: 3,
	}

	for _, opt := range opts {
		opt(client)
	}

	// Only create default HTTP client if one wasn't provided
	if client.httpClient == nil {
		client.httpClient = &http.Client{
			Timeout: client.timeout,
		}
	}

	return client
}

// Get performs a GET request to the specified path.
func (c *Client) Get(ctx context.Context, path string, opts ...RequestOption) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil, opts...)
}

// Post performs a POST request to the specified path with the given body.
func (c *Client) Post(ctx context.Context, path string, body interface{}, opts ...RequestOption) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, path, body, opts...)
}

// Put performs a PUT request to the specified path with the given body.
func (c *Client) Put(ctx context.Context, path string, body interface{}, opts ...RequestOption) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPut, path, body, opts...)
}

// Patch performs a PATCH request to the specified path with the given body.
func (c *Client) Patch(ctx context.Context, path string, body interface{}, opts ...RequestOption) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPatch, path, body, opts...)
}

// Delete performs a DELETE request to the specified path.
func (c *Client) Delete(ctx context.Context, path string, opts ...RequestOption) ([]byte, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil, opts...)
}

// doRequest performs an HTTP request with automatic token injection and retry logic.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, opts ...RequestOption) ([]byte, error) {
	// Build request options
	reqOpts := &requestOptions{}
	for _, opt := range opts {
		opt(reqOpts)
	}

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	url := c.buildURL(path, reqOpts.queryParams)

	// Retry loop with exponential backoff
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s...
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			logger.DebugEvent().
				Int("attempt", attempt).
				Dur("backoff", backoff).
				Msg("Retrying request after backoff")
			time.Sleep(backoff)

			// Reset body reader for retry
			if body != nil {
				jsonData, _ := json.Marshal(body)
				bodyReader = bytes.NewReader(jsonData)
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set content type for POST/PUT/PATCH requests
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		// Add custom headers
		for key, value := range reqOpts.headers {
			req.Header.Set(key, value)
		}

		// Inject JWT bearer token if auth client is configured and not skipped
		if c.authClient != nil && !reqOpts.skipAuth {
			if err := c.injectAuthToken(ctx, req); err != nil {
				logger.WarnEvent().Err(err).Msg("Failed to inject auth token")
				// Continue without token - API might be public
			}
		}

		// Log request
		logger.DebugEvent().
			Str("method", method).
			Str("url", url).
			Int("attempt", attempt+1).
			Msg("Sending HTTP request")

		// Execute request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			logger.WarnEvent().Err(lastErr).Msg("Request failed, will retry")
			continue
		}

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			logger.WarnEvent().Err(lastErr).Msg("Failed to read response")
			continue
		}

		// Log response
		logger.DebugEvent().
			Int("status", resp.StatusCode).
			Int("body_size", len(respBody)).
			Msg("Received HTTP response")

		// Handle 401 Unauthorized - token might be expired
		if resp.StatusCode == http.StatusUnauthorized && c.authClient != nil {
			logger.InfoEvent().Msg("Received 401, attempting token refresh")

			// Try to refresh token
			tokens, err := c.authClient.GetStoredTokens(ctx)
			if err == nil && tokens.RefreshToken != nil {
				_, err := c.authClient.RefreshToken(ctx, tokens.RefreshToken)
				if err == nil {
					// Token refreshed successfully, retry the request
					logger.InfoEvent().Msg("Token refreshed successfully, retrying request")
					continue
				}
			}

			// Token refresh failed or no refresh token available
			return nil, fmt.Errorf("authentication failed: %s", string(respBody))
		}

		// Handle other error status codes
		if resp.StatusCode >= 400 {
			// Check if we should retry
			if c.shouldRetry(resp.StatusCode) && attempt < c.maxRetries {
				lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
				logger.WarnEvent().
					Int("status", resp.StatusCode).
					Msg("Request failed with retryable error")
				continue
			}

			// Non-retryable error or max retries exceeded
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}

		// Success
		return respBody, nil
	}

	// All retries exhausted
	return nil, fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// injectAuthToken retrieves the access token and adds it to the request.
func (c *Client) injectAuthToken(ctx context.Context, req *http.Request) error {
	tokens, err := c.authClient.GetStoredTokens(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stored tokens: %w", err)
	}

	if tokens == nil || tokens.AccessToken == nil {
		return fmt.Errorf("no access token available")
	}

	// Check if token needs refresh
	if tokens.NeedsRefresh() && tokens.RefreshToken != nil {
		logger.DebugEvent().Msg("Access token needs refresh")
		newTokens, err := c.authClient.RefreshToken(ctx, tokens.RefreshToken)
		if err != nil {
			logger.WarnEvent().Err(err).Msg("Failed to refresh token, using existing")
		} else {
			tokens = newTokens
		}
	}

	// Add bearer token to request
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken.Raw)
	return nil
}

// buildURL constructs the full URL from base URL, path, and query parameters.
func (c *Client) buildURL(path string, queryParams map[string]string) string {
	url := c.baseURL + path

	if len(queryParams) > 0 {
		query := "?"
		first := true
		for key, value := range queryParams {
			if !first {
				query += "&"
			}
			query += key + "=" + value
			first = false
		}
		url += query
	}

	return url
}

// shouldRetry determines if a request should be retried based on status code.
func (c *Client) shouldRetry(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, // 429 - Rate limited
		http.StatusInternalServerError,     // 500 - Server error
		http.StatusBadGateway,               // 502 - Bad gateway
		http.StatusServiceUnavailable,       // 503 - Service unavailable
		http.StatusGatewayTimeout:           // 504 - Gateway timeout
		return true
	default:
		return false
	}
}
