package provider

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/AINative-studio/ainative-code/internal/logger"
)

// BaseProvider provides common functionality for all provider implementations
type BaseProvider struct {
	name       string
	httpClient *http.Client
	logger     logger.LoggerInterface
	retryConfig RetryConfig
}

// RetryConfig configures retry behavior for failed requests
type RetryConfig struct {
	MaxRetries     int           // Maximum number of retry attempts
	InitialBackoff time.Duration // Initial backoff duration
	MaxBackoff     time.Duration // Maximum backoff duration
	Multiplier     float64       // Backoff multiplier for exponential backoff
	RetryableStatusCodes []int    // HTTP status codes that should trigger retries
}

// DefaultRetryConfig returns sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
		RetryableStatusCodes: []int{
			http.StatusTooManyRequests,      // 429
			http.StatusInternalServerError,  // 500
			http.StatusBadGateway,           // 502
			http.StatusServiceUnavailable,   // 503
			http.StatusGatewayTimeout,       // 504
		},
	}
}

// BaseProviderConfig contains configuration for creating a BaseProvider
type BaseProviderConfig struct {
	Name        string
	HTTPClient  *http.Client
	Logger      logger.LoggerInterface
	RetryConfig RetryConfig
}

// NewBaseProvider creates a new BaseProvider with the given configuration
func NewBaseProvider(config BaseProviderConfig) *BaseProvider {
	// Use default HTTP client if none provided
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	}

	// Use default retry config if not provided
	retryConfig := config.RetryConfig
	if retryConfig.MaxRetries == 0 {
		retryConfig = DefaultRetryConfig()
	}

	return &BaseProvider{
		name:        config.Name,
		httpClient:  httpClient,
		logger:      config.Logger,
		retryConfig: retryConfig,
	}
}

// Name returns the provider's name
func (b *BaseProvider) Name() string {
	return b.name
}

// DoRequest executes an HTTP request with retry logic and error handling
func (b *BaseProvider) DoRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= b.retryConfig.MaxRetries; attempt++ {
		// Check if context is cancelled before attempting request
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("request cancelled: %w", err)
		}

		// Log the attempt
		if b.logger != nil && attempt > 0 {
			b.logger.Debug(fmt.Sprintf("Retry attempt %d/%d for %s %s",
				attempt, b.retryConfig.MaxRetries, req.Method, req.URL.String()))
		}

		// Execute the request with context
		resp, err := b.httpClient.Do(req.WithContext(ctx))

		// Request succeeded
		if err == nil {
			// Check if status code indicates success or non-retryable error
			if !b.shouldRetry(resp.StatusCode) {
				return resp, nil
			}

			// Status code indicates we should retry
			lastErr = fmt.Errorf("request failed with status %d", resp.StatusCode)

			// Close the response body before retrying
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			// Check for rate limit headers
			if resp.StatusCode == http.StatusTooManyRequests {
				retryAfter := b.parseRetryAfter(resp)
				if b.logger != nil {
					b.logger.Warn(fmt.Sprintf("Rate limited by provider, retry after %d seconds", retryAfter))
				}
				lastErr = NewRateLimitError(b.name, retryAfter)
			}
		} else {
			// Network error occurred
			lastErr = err
			if b.logger != nil {
				b.logger.Error(fmt.Sprintf("Request failed: %v", err))
			}
		}

		// Don't sleep after the last attempt
		if attempt < b.retryConfig.MaxRetries {
			backoff := b.calculateBackoff(attempt)

			if b.logger != nil {
				b.logger.Debug(fmt.Sprintf("Backing off for %v before retry", backoff))
			}

			// Sleep with context awareness
			select {
			case <-time.After(backoff):
				// Continue to next retry
			case <-ctx.Done():
				return nil, fmt.Errorf("request cancelled during backoff: %w", ctx.Err())
			}
		}
	}

	// All retries exhausted
	if b.logger != nil {
		b.logger.Error(fmt.Sprintf("Request failed after %d retries", b.retryConfig.MaxRetries))
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", b.retryConfig.MaxRetries, lastErr)
}

// shouldRetry determines if a request should be retried based on status code
func (b *BaseProvider) shouldRetry(statusCode int) bool {
	for _, retryableCode := range b.retryConfig.RetryableStatusCodes {
		if statusCode == retryableCode {
			return true
		}
	}
	return false
}

// calculateBackoff calculates the backoff duration for a given retry attempt
// Uses exponential backoff with jitter
func (b *BaseProvider) calculateBackoff(attempt int) time.Duration {
	// Calculate exponential backoff
	backoff := float64(b.retryConfig.InitialBackoff) * math.Pow(b.retryConfig.Multiplier, float64(attempt))

	// Cap at max backoff
	if backoff > float64(b.retryConfig.MaxBackoff) {
		backoff = float64(b.retryConfig.MaxBackoff)
	}

	// Add jitter (±10% randomization to prevent thundering herd)
	jitter := 0.9 + (0.2 * (float64(time.Now().UnixNano()%100) / 100.0))
	backoff = backoff * jitter

	return time.Duration(backoff)
}

// parseRetryAfter attempts to parse Retry-After header from response
// Returns the number of seconds to wait, or 0 if header not present
func (b *BaseProvider) parseRetryAfter(resp *http.Response) int {
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return 0
	}

	// Try to parse as integer (seconds)
	var seconds int
	if _, err := fmt.Sscanf(retryAfter, "%d", &seconds); err == nil {
		return seconds
	}

	// Try to parse as HTTP date
	if t, err := http.ParseTime(retryAfter); err == nil {
		duration := time.Until(t)
		if duration > 0 {
			return int(duration.Seconds())
		}
	}

	return 0
}

// Close releases resources held by the base provider
func (b *BaseProvider) Close() error {
	// Close idle connections
	b.httpClient.CloseIdleConnections()

	if b.logger != nil {
		b.logger.Debug(fmt.Sprintf("Closed provider: %s", b.name))
	}

	return nil
}

// LogRequest logs details about an HTTP request (for debugging)
func (b *BaseProvider) LogRequest(req *http.Request) {
	if b.logger == nil {
		return
	}

	b.logger.Debug(fmt.Sprintf("Request: %s %s", req.Method, req.URL.String()))
}

// LogResponse logs details about an HTTP response (for debugging)
func (b *BaseProvider) LogResponse(resp *http.Response) {
	if b.logger == nil {
		return
	}

	b.logger.Debug(fmt.Sprintf("Response: %d from %s", resp.StatusCode, resp.Request.URL.String()))
}

// HandleHTTPError converts HTTP errors to provider-specific errors
func (b *BaseProvider) HandleHTTPError(resp *http.Response, body []byte) error {
	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return NewAuthenticationError(b.name, string(body))

	case http.StatusTooManyRequests:
		retryAfter := b.parseRetryAfter(resp)
		return NewRateLimitError(b.name, retryAfter)

	case http.StatusBadRequest:
		// Check if it's a context length error (common across providers)
		// Actual parsing would be provider-specific
		return NewProviderError(b.name, "", fmt.Errorf("bad request: %s", string(body)))

	default:
		return NewProviderError(b.name, "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)))
	}
}

// ValidateModel checks if a model is supported and returns an error if not
func (b *BaseProvider) ValidateModel(model string, supportedModels []string) error {
	if model == "" {
		return NewInvalidModelError(b.name, model, supportedModels)
	}

	for _, supported := range supportedModels {
		if model == supported {
			return nil
		}
	}

	return NewInvalidModelError(b.name, model, supportedModels)
}
