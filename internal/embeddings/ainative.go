package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AINative-studio/ainative-code/internal/logger"
)

const (
	// DefaultEmbeddingsEndpoint is the default AINative platform embeddings API endpoint
	DefaultEmbeddingsEndpoint = "https://api.ainative.io/api/v1/embeddings"

	// DefaultTimeout for embeddings requests
	DefaultTimeout = 30 * time.Second

	// MaxBatchSize is the maximum number of texts to embed in a single request
	MaxBatchSize = 100
)

// AINativeEmbeddingsClient provides access to AINative platform embeddings API
// This is the ONLY way to get embeddings - we do NOT use OpenAI embeddings endpoints
type AINativeEmbeddingsClient struct {
	apiKey      string
	endpoint    string
	httpClient  *http.Client
	logger      logger.LoggerInterface
	maxRetries  int
	retryDelay  time.Duration
}

// Config contains configuration for the AINative embeddings client
type Config struct {
	APIKey     string                  // Required: AINative API key
	Endpoint   string                  // Optional: Custom endpoint (defaults to DefaultEmbeddingsEndpoint)
	HTTPClient *http.Client            // Optional: Custom HTTP client
	Logger     logger.LoggerInterface  // Optional: Logger for debugging
	MaxRetries int                     // Optional: Max retry attempts (default: 3)
	RetryDelay time.Duration           // Optional: Delay between retries (default: 1s)
}

// EmbeddingRequest represents a request to the embeddings API
type embeddingRequest struct {
	Texts     []string `json:"texts"`      // Texts to embed
	Model     string   `json:"model"`      // Embedding model to use
	Normalize bool     `json:"normalize"`  // Whether to normalize vectors
}

// EmbeddingResponse represents a response from the embeddings API
type embeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"` // Vector embeddings
	Model      string      `json:"model"`      // Model used
	Usage      struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// EmbeddingError represents an error from the embeddings API
type embeddingError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// EmbeddingResult contains the results of an embedding operation
type EmbeddingResult struct {
	Embeddings  [][]float32 // Vector embeddings (one per input text)
	Model       string      // Model used for embedding
	TotalTokens int         // Total tokens processed
}

// NewAINativeEmbeddingsClient creates a new AINative embeddings client
func NewAINativeEmbeddingsClient(config Config) (*AINativeEmbeddingsClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("AINative API key is required for embeddings")
	}

	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = DefaultEmbeddingsEndpoint
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: DefaultTimeout,
		}
	}

	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	retryDelay := config.RetryDelay
	if retryDelay == 0 {
		retryDelay = 1 * time.Second
	}

	return &AINativeEmbeddingsClient{
		apiKey:     config.APIKey,
		endpoint:   endpoint,
		httpClient: httpClient,
		logger:     config.Logger,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}, nil
}

// Embed generates embeddings for the given texts using AINative platform API
// This is the primary method for obtaining vector embeddings
func (c *AINativeEmbeddingsClient) Embed(ctx context.Context, texts []string, model string) (*EmbeddingResult, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("at least one text is required")
	}

	if len(texts) > MaxBatchSize {
		return nil, fmt.Errorf("batch size exceeds maximum of %d", MaxBatchSize)
	}

	if model == "" {
		model = "default" // AINative platform will use default embedding model
	}

	// Build request
	reqBody := embeddingRequest{
		Texts:     texts,
		Model:     model,
		Normalize: true, // Always normalize for cosine similarity
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Execute with retries
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay * time.Duration(attempt)):
			}

			if c.logger != nil {
				c.logger.Debug(fmt.Sprintf("Retrying embeddings request (attempt %d/%d)", attempt, c.maxRetries))
			}
		}

		result, err := c.doRequest(ctx, jsonBody)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry on client errors (400-499 except 429)
		if apiErr, ok := err.(*EmbeddingAPIError); ok {
			if apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != 429 {
				return nil, err
			}
		}
	}

	return nil, fmt.Errorf("embeddings request failed after %d retries: %w", c.maxRetries, lastErr)
}

// doRequest executes a single embeddings API request
func (c *AINativeEmbeddingsClient) doRequest(ctx context.Context, jsonBody []byte) (*EmbeddingResult, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("User-Agent", "AINative-Code/1.0")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp.StatusCode, body)
	}

	// Parse success response
	var apiResp embeddingResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &EmbeddingResult{
		Embeddings:  apiResp.Embeddings,
		Model:       apiResp.Model,
		TotalTokens: apiResp.Usage.TotalTokens,
	}, nil
}

// parseError parses an error response from the API
func (c *AINativeEmbeddingsClient) parseError(statusCode int, body []byte) error {
	var apiErr embeddingError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		// Can't parse error, return generic error
		return &EmbeddingAPIError{
			StatusCode: statusCode,
			Message:    string(body),
			Type:       "unknown",
		}
	}

	return &EmbeddingAPIError{
		StatusCode: statusCode,
		Message:    apiErr.Error.Message,
		Type:       apiErr.Error.Type,
		Code:       apiErr.Error.Code,
	}
}

// EmbeddingAPIError represents an error from the embeddings API
type EmbeddingAPIError struct {
	StatusCode int
	Message    string
	Type       string
	Code       string
}

func (e *EmbeddingAPIError) Error() string {
	return fmt.Sprintf("embeddings API error (status %d): %s", e.StatusCode, e.Message)
}

// IsAuthenticationError returns true if the error is due to authentication failure
func (e *EmbeddingAPIError) IsAuthenticationError() bool {
	return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
}

// IsRateLimitError returns true if the error is due to rate limiting
func (e *EmbeddingAPIError) IsRateLimitError() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsQuotaExceededError returns true if the error is due to quota exhaustion
func (e *EmbeddingAPIError) IsQuotaExceededError() bool {
	return e.Type == "quota_exceeded" || e.Code == "quota_exceeded"
}

// Close releases any resources held by the client
func (c *AINativeEmbeddingsClient) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
