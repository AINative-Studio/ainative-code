package ollama

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/provider"
)

// OllamaConnectionError represents a connection failure to Ollama server
type OllamaConnectionError struct {
	BaseURL string
	Err     error
}

// Error implements the error interface
func (e *OllamaConnectionError) Error() string {
	return fmt.Sprintf("ollama server not running at %s: %v (ensure Ollama is installed and running)", e.BaseURL, e.Err)
}

// Unwrap returns the underlying error
func (e *OllamaConnectionError) Unwrap() error {
	return e.Err
}

// NewOllamaConnectionError creates a new connection error
func NewOllamaConnectionError(baseURL string, err error) *OllamaConnectionError {
	return &OllamaConnectionError{
		BaseURL: baseURL,
		Err:     err,
	}
}

// OllamaOutOfMemoryError represents an out of memory error
type OllamaOutOfMemoryError struct {
	Model string
}

// Error implements the error interface
func (e *OllamaOutOfMemoryError) Error() string {
	return fmt.Sprintf("out of memory loading model %q: try using a smaller model or increase available RAM", e.Model)
}

// NewOllamaOutOfMemoryError creates a new out of memory error
func NewOllamaOutOfMemoryError(model string) *OllamaOutOfMemoryError {
	return &OllamaOutOfMemoryError{
		Model: model,
	}
}

// ollamaErrorResponse represents the error response from Ollama API
type ollamaErrorResponse struct {
	Error string `json:"error"`
}

// parseOllamaError converts Ollama API errors to provider errors
func parseOllamaError(statusCode int, body []byte, model string) error {
	// Handle connection errors (no status code)
	if statusCode == 0 {
		return NewOllamaConnectionError(DefaultOllamaURL, errors.New("connection failed"))
	}

	// Try to parse error response
	var errResp ollamaErrorResponse
	if len(body) > 0 {
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return mapOllamaError(statusCode, errResp.Error, model)
		}
	}

	// Fallback to status code based errors
	switch statusCode {
	case http.StatusNotFound:
		return NewOllamaModelNotFoundError(model)
	case http.StatusBadRequest:
		return provider.NewProviderError("ollama", model, fmt.Errorf("bad request: %s", string(body)))
	case http.StatusInternalServerError:
		return provider.NewProviderError("ollama", model, fmt.Errorf("internal server error: %s", string(body)))
	default:
		return provider.NewProviderError("ollama", model, fmt.Errorf("HTTP %d: %s", statusCode, string(body)))
	}
}

// mapOllamaError maps Ollama error messages to appropriate error types
func mapOllamaError(statusCode int, errMsg string, model string) error {
	errLower := strings.ToLower(errMsg)

	// Model not found
	if strings.Contains(errLower, "not found") || strings.Contains(errLower, "model") && statusCode == http.StatusNotFound {
		return NewOllamaModelNotFoundError(model)
	}

	// Out of memory
	if strings.Contains(errLower, "out of memory") || strings.Contains(errLower, "oom") {
		return NewOllamaOutOfMemoryError(model)
	}

	// Context length exceeded
	if strings.Contains(errLower, "context length") || strings.Contains(errLower, "exceeds maximum") {
		// Try to extract token counts if available
		return provider.NewContextLengthError("ollama", model, 0, 0)
	}

	// Connection refused / server not running
	if strings.Contains(errLower, "connection refused") || strings.Contains(errLower, "connect: connection refused") {
		return NewOllamaConnectionError(DefaultOllamaURL, fmt.Errorf("%s", errMsg))
	}

	// Generic error
	return provider.NewProviderError("ollama", model, fmt.Errorf("%s", errMsg))
}

// NewOllamaModelNotFoundError creates an error for when a model is not found
func NewOllamaModelNotFoundError(model string) error {
	// Return an InvalidModelError that provides helpful suggestions
	err := provider.NewInvalidModelError("ollama", model, []string{})

	// Wrap with a more helpful message
	return fmt.Errorf("model '%s' not found: use 'ollama pull %s' to download it, or use 'ollama list' to see available models: %w", model, model, err)
}

// IsConnectionError checks if an error is an Ollama connection error
func IsConnectionError(err error) bool {
	var connErr *OllamaConnectionError
	return errors.As(err, &connErr)
}

// IsModelNotFoundError checks if an error is a model not found error
func IsModelNotFoundError(err error) bool {
	var invalidModelErr *provider.InvalidModelError
	return errors.As(err, &invalidModelErr)
}

// IsOutOfMemoryError checks if an error is an out of memory error
func IsOutOfMemoryError(err error) bool {
	var oomErr *OllamaOutOfMemoryError
	return errors.As(err, &oomErr)
}
