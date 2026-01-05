package ollama

import (
	"errors"
	"net/http"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/stretchr/testify/assert"
)

func TestParseOllamaError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		body           []byte
		model          string
		expectedType   error
		expectedMsg    string
	}{
		{
			name:       "model not found",
			statusCode: http.StatusNotFound,
			body:       []byte(`{"error":"model 'unknown' not found"}`),
			model:      "unknown",
			expectedType: &provider.InvalidModelError{},
			expectedMsg:  "unknown",
		},
		{
			name:       "connection refused",
			statusCode: 0,
			body:       nil,
			model:      "llama2",
			expectedType: &OllamaConnectionError{},
			expectedMsg:  "ollama server not running",
		},
		{
			name:       "out of memory",
			statusCode: http.StatusInternalServerError,
			body:       []byte(`{"error":"out of memory"}`),
			model:      "llama2",
			expectedType: &OllamaOutOfMemoryError{},
			expectedMsg:  "out of memory",
		},
		{
			name:       "invalid request",
			statusCode: http.StatusBadRequest,
			body:       []byte(`{"error":"invalid prompt format"}`),
			model:      "llama2",
			expectedType: &provider.ProviderError{},
			expectedMsg:  "invalid prompt format",
		},
		{
			name:       "context length exceeded",
			statusCode: http.StatusBadRequest,
			body:       []byte(`{"error":"context length exceeded: 4096 > 2048"}`),
			model:      "llama2",
			expectedType: &provider.ContextLengthError{},
			expectedMsg:  "context length exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseOllamaError(tt.statusCode, tt.body, tt.model)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedMsg)

			// Check error type using errors.As for wrapped errors
			switch tt.expectedType.(type) {
			case *provider.InvalidModelError:
				var invalidErr *provider.InvalidModelError
				assert.True(t, errors.As(err, &invalidErr), "expected InvalidModelError")
			case *OllamaConnectionError:
				var connErr *OllamaConnectionError
				assert.True(t, errors.As(err, &connErr), "expected OllamaConnectionError")
			case *OllamaOutOfMemoryError:
				var oomErr *OllamaOutOfMemoryError
				assert.True(t, errors.As(err, &oomErr), "expected OllamaOutOfMemoryError")
			case *provider.ContextLengthError:
				var ctxErr *provider.ContextLengthError
				assert.True(t, errors.As(err, &ctxErr), "expected ContextLengthError")
			case *provider.ProviderError:
				// Should be a provider error
				var provErr *provider.ProviderError
				assert.True(t, errors.As(err, &provErr), "expected ProviderError")
			}
		})
	}
}

func TestOllamaConnectionError(t *testing.T) {
	err := NewOllamaConnectionError("http://localhost:11434", errors.New("connection refused"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ollama server not running")
	assert.Contains(t, err.Error(), "http://localhost:11434")
	assert.Contains(t, err.Error(), "connection refused")
}

func TestOllamaModelNotFoundError(t *testing.T) {
	err := NewOllamaModelNotFoundError("unknown-model")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model 'unknown-model' not found")
	assert.Contains(t, err.Error(), "ollama pull")

	// Check it's an InvalidModelError
	var invalidModelErr *provider.InvalidModelError
	assert.True(t, errors.As(err, &invalidModelErr))
}

func TestOllamaOutOfMemoryError(t *testing.T) {
	err := NewOllamaOutOfMemoryError("llama2:70b")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of memory")
	assert.Contains(t, err.Error(), "llama2:70b")
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "connection error",
			err:      NewOllamaConnectionError("http://localhost:11434", errors.New("refused")),
			expected: true,
		},
		{
			name:     "model not found error",
			err:      NewOllamaModelNotFoundError("unknown"),
			expected: false,
		},
		{
			name:     "generic provider error",
			err:      provider.NewProviderError("ollama", "llama2", errors.New("test")),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConnectionError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsModelNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "model not found error",
			err:      NewOllamaModelNotFoundError("unknown"),
			expected: true,
		},
		{
			name:     "connection error",
			err:      NewOllamaConnectionError("http://localhost:11434", errors.New("refused")),
			expected: false,
		},
		{
			name:     "invalid model error",
			err:      provider.NewInvalidModelError("ollama", "test", []string{"llama2"}),
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsModelNotFoundError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsOutOfMemoryError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "out of memory error",
			err:      NewOllamaOutOfMemoryError("llama2:70b"),
			expected: true,
		},
		{
			name:     "connection error",
			err:      NewOllamaConnectionError("http://localhost:11434", errors.New("refused")),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsOutOfMemoryError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
