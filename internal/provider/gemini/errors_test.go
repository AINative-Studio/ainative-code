package gemini

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleAPIError tests various error scenarios
func TestHandleAPIError(t *testing.T) {
	p, err := NewGeminiProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	tests := []struct {
		name          string
		statusCode    int
		errorResponse geminiError
		expectedType  string // "authentication", "rate_limit", "context_length", "invalid_model", "generic"
	}{
		{
			name:       "authentication error - 401",
			statusCode: http.StatusUnauthorized,
			errorResponse: geminiError{
				Error: errorDetails{
					Code:    401,
					Message: "Invalid API key",
					Status:  "UNAUTHENTICATED",
				},
			},
			expectedType: "authentication",
		},
		{
			name:       "authentication error - 403",
			statusCode: http.StatusForbidden,
			errorResponse: geminiError{
				Error: errorDetails{
					Code:    403,
					Message: "Permission denied",
					Status:  "PERMISSION_DENIED",
				},
			},
			expectedType: "authentication",
		},
		{
			name:       "rate limit error - 429",
			statusCode: http.StatusTooManyRequests,
			errorResponse: geminiError{
				Error: errorDetails{
					Code:    429,
					Message: "Too many requests",
					Status:  "RESOURCE_EXHAUSTED",
				},
			},
			expectedType: "rate_limit",
		},
		{
			name:       "invalid model error - 404",
			statusCode: http.StatusNotFound,
			errorResponse: geminiError{
				Error: errorDetails{
					Code:    404,
					Message: "model not found",
					Status:  "NOT_FOUND",
				},
			},
			expectedType: "invalid_model",
		},
		{
			name:       "context length error",
			statusCode: http.StatusBadRequest,
			errorResponse: geminiError{
				Error: errorDetails{
					Code:    400,
					Message: "Request content exceeds token limit",
					Status:  "INVALID_ARGUMENT",
				},
			},
			expectedType: "context_length",
		},
		{
			name:       "content too long error",
			statusCode: http.StatusBadRequest,
			errorResponse: geminiError{
				Error: errorDetails{
					Code:    400,
					Message: "content is too long",
					Status:  "INVALID_ARGUMENT",
				},
			},
			expectedType: "context_length",
		},
		{
			name:       "invalid API key in message",
			statusCode: http.StatusBadRequest,
			errorResponse: geminiError{
				Error: errorDetails{
					Code:    400,
					Message: "API key is invalid",
					Status:  "INVALID_ARGUMENT",
				},
			},
			expectedType: "authentication",
		},
		{
			name:       "server error - 500",
			statusCode: http.StatusInternalServerError,
			errorResponse: geminiError{
				Error: errorDetails{
					Code:    500,
					Message: "Internal server error",
					Status:  "INTERNAL",
				},
			},
			expectedType: "generic",
		},
		{
			name:       "service unavailable - 503",
			statusCode: http.StatusServiceUnavailable,
			errorResponse: geminiError{
				Error: errorDetails{
					Code:    503,
					Message: "Service temporarily unavailable",
					Status:  "UNAVAILABLE",
				},
			},
			expectedType: "generic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.errorResponse)

			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       http.NoBody,
			}

			err := p.handleAPIError(resp, body, "gemini-pro")
			require.Error(t, err)

			// Check error type
			switch tt.expectedType {
			case "authentication":
				_, ok := err.(*provider.AuthenticationError)
				assert.True(t, ok, "expected authentication error")
			case "rate_limit":
				_, ok := err.(*provider.RateLimitError)
				assert.True(t, ok, "expected rate limit error")
			case "context_length":
				_, ok := err.(*provider.ContextLengthError)
				assert.True(t, ok, "expected context length error")
			case "invalid_model":
				_, ok := err.(*provider.InvalidModelError)
				assert.True(t, ok, "expected invalid model error")
			case "generic":
				_, ok := err.(*provider.ProviderError)
				assert.True(t, ok, "expected generic provider error")
			}
		})
	}
}

// TestConvertAPIError tests the error conversion logic
func TestConvertAPIError(t *testing.T) {
	p, err := NewGeminiProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	tests := []struct {
		name         string
		apiError     *geminiError
		statusCode   int
		model        string
		expectedType string
		expectedMsg  string
	}{
		{
			name: "invalid argument",
			apiError: &geminiError{
				Error: errorDetails{
					Code:    400,
					Message: "Invalid parameter",
					Status:  "INVALID_ARGUMENT",
				},
			},
			statusCode:   400,
			model:        "gemini-pro",
			expectedType: "generic",
			expectedMsg:  "invalid request",
		},
		{
			name: "model not found in message",
			apiError: &geminiError{
				Error: errorDetails{
					Code:    400,
					Message: "The requested model was not found",
					Status:  "INVALID_ARGUMENT",
				},
			},
			statusCode:   400,
			model:        "invalid-model",
			expectedType: "invalid_model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.convertAPIError(tt.apiError, tt.statusCode, tt.model)
			require.Error(t, err)

			if tt.expectedMsg != "" {
				assert.Contains(t, err.Error(), tt.expectedMsg)
			}
		})
	}
}

// TestErrorHandlingIntegration tests error handling with mock server
func TestErrorHandlingIntegration(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		errorBody    geminiError
		expectErrMsg string
	}{
		{
			name:       "authentication failure",
			statusCode: http.StatusUnauthorized,
			errorBody: geminiError{
				Error: errorDetails{
					Code:    401,
					Message: "API key not valid",
					Status:  "UNAUTHENTICATED",
				},
			},
			expectErrMsg: "authentication",
		},
		{
			name:       "rate limited",
			statusCode: http.StatusTooManyRequests,
			errorBody: geminiError{
				Error: errorDetails{
					Code:    429,
					Message: "Quota exceeded",
					Status:  "RESOURCE_EXHAUSTED",
				},
			},
			expectErrMsg: "rate limit",
		},
		{
			name:       "model not found",
			statusCode: http.StatusNotFound,
			errorBody: geminiError{
				Error: errorDetails{
					Code:    404,
					Message: "Model not found",
					Status:  "NOT_FOUND",
				},
			},
			expectErrMsg: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.errorBody)
			}))
			defer server.Close()

			p, err := NewGeminiProvider(Config{
				APIKey:  "test-key",
				BaseURL: server.URL,
			})
			require.NoError(t, err)

			ctx := httptest.NewRequest("GET", "/", nil).Context()
			messages := []provider.Message{
				{Role: "user", Content: "Hello"},
			}

			_, err = p.Chat(ctx, messages, provider.WithModel("gemini-pro"))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectErrMsg)
		})
	}
}

// TestMalformedErrorResponse tests handling of malformed error responses
func TestMalformedErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	p, err := NewGeminiProvider(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := httptest.NewRequest("GET", "/", nil).Context()
	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	_, err = p.Chat(ctx, messages, provider.WithModel("gemini-pro"))
	require.Error(t, err)
	// Should still handle error gracefully even if JSON parsing fails
}

// TestErrorDetails tests that error details are preserved
func TestErrorDetails(t *testing.T) {
	p, err := NewGeminiProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	apiErr := &geminiError{
		Error: errorDetails{
			Code:    400,
			Message: "Specific error message",
			Status:  "INVALID_ARGUMENT",
			Details: []errorDetail{
				{
					Type:   "type.googleapis.com/google.rpc.BadRequest",
					Reason: "FIELD_VIOLATION",
					Domain: "googleapis.com",
				},
			},
		},
	}

	err = p.convertAPIError(apiErr, 400, "gemini-pro")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Specific error message")
}
