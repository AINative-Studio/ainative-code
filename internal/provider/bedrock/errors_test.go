package bedrock

import (
	"net/http"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/stretchr/testify/assert"
)

func TestParseBedrockError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		errorType  string
	}{
		{
			name:       "authentication error - invalid credentials",
			statusCode: http.StatusForbidden,
			body:       `{"message":"The security token included in the request is invalid."}`,
			errorType:  "authentication",
		},
		{
			name:       "authentication error - expired token",
			statusCode: http.StatusForbidden,
			body:       `{"message":"The security token included in the request is expired"}`,
			errorType:  "authentication",
		},
		{
			name:       "authentication error - unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{"message":"Access denied"}`,
			errorType:  "authentication",
		},
		{
			name:       "throttling error",
			statusCode: http.StatusTooManyRequests,
			body:       `{"message":"Rate exceeded"}`,
			errorType:  "rate_limit",
		},
		{
			name:       "validation error - missing field",
			statusCode: http.StatusBadRequest,
			body:       `{"message":"Validation error: messages is required"}`,
			errorType:  "validation",
		},
		{
			name:       "validation error - invalid parameter",
			statusCode: http.StatusBadRequest,
			body:       `{"message":"Invalid request: max_tokens must be greater than 0"}`,
			errorType:  "validation",
		},
		{
			name:       "model not found",
			statusCode: http.StatusNotFound,
			body:       `{"message":"Could not resolve the foundation model"}`,
			errorType:  "model_not_found",
		},
		{
			name:       "service unavailable",
			statusCode: http.StatusServiceUnavailable,
			body:       `{"message":"Service temporarily unavailable"}`,
			errorType:  "service_unavailable",
		},
		{
			name:       "internal error",
			statusCode: http.StatusInternalServerError,
			body:       `{"message":"Internal server error"}`,
			errorType:  "internal",
		},
		{
			name:       "context length exceeded",
			statusCode: http.StatusBadRequest,
			body:       `{"message":"Prompt is too long. The prompt has 150000 tokens but the model only supports 100000"}`,
			errorType:  "context_length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseBedrockError(tt.statusCode, []byte(tt.body), "test-model")
			assert.Error(t, err)

			switch tt.errorType {
			case "authentication":
				var authErr *provider.AuthenticationError
				assert.ErrorAs(t, err, &authErr)
			case "rate_limit":
				var rateLimitErr *provider.RateLimitError
				assert.ErrorAs(t, err, &rateLimitErr)
			case "context_length":
				var contextErr *provider.ContextLengthError
				assert.ErrorAs(t, err, &contextErr)
			case "validation", "model_not_found", "service_unavailable", "internal":
				var providerErr *provider.ProviderError
				assert.ErrorAs(t, err, &providerErr)
			}
		})
	}
}

func TestParseErrorResponse(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		expectedMessage string
		expectError     bool
	}{
		{
			name:            "valid error response",
			body:            `{"message":"Error occurred"}`,
			expectedMessage: "Error occurred",
			expectError:     false,
		},
		{
			name:            "error with __type field",
			body:            `{"__type":"ValidationException","message":"Invalid input"}`,
			expectedMessage: "Invalid input",
			expectError:     false,
		},
		{
			name:            "empty message",
			body:            `{"message":""}`,
			expectedMessage: "",
			expectError:     false,
		},
		{
			name:        "invalid JSON",
			body:        `{invalid json}`,
			expectError: true,
		},
		{
			name:            "missing message field",
			body:            `{"error":"something"}`,
			expectedMessage: "",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errResp, err := parseErrorResponse([]byte(tt.body))

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMessage, errResp.Message)
			}
		})
	}
}

func TestIsAuthenticationError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		expected   bool
	}{
		{
			name:       "403 with invalid token",
			statusCode: http.StatusForbidden,
			message:    "The security token included in the request is invalid",
			expected:   true,
		},
		{
			name:       "403 with expired token",
			statusCode: http.StatusForbidden,
			message:    "Token has expired",
			expected:   true,
		},
		{
			name:       "401 unauthorized",
			statusCode: http.StatusUnauthorized,
			message:    "Access denied",
			expected:   true,
		},
		{
			name:       "403 without auth message",
			statusCode: http.StatusForbidden,
			message:    "Resource not accessible",
			expected:   true, // Still auth error because of status code
		},
		{
			name:       "400 bad request",
			statusCode: http.StatusBadRequest,
			message:    "Invalid input",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAuthenticationError(tt.statusCode, tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsThrottlingError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		expected   bool
	}{
		{
			name:       "429 rate limit",
			statusCode: http.StatusTooManyRequests,
			message:    "Rate exceeded",
			expected:   true,
		},
		{
			name:       "400 with throttling message",
			statusCode: http.StatusBadRequest,
			message:    "ThrottlingException: Rate exceeded",
			expected:   true,
		},
		{
			name:       "400 without throttling",
			statusCode: http.StatusBadRequest,
			message:    "Validation error",
			expected:   false,
		},
		{
			name:       "500 internal error",
			statusCode: http.StatusInternalServerError,
			message:    "Internal error",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isThrottlingError(tt.statusCode, tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsContextLengthError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "prompt too long",
			message:  "Prompt is too long. The prompt has 150000 tokens",
			expected: true,
		},
		{
			name:     "exceeds maximum",
			message:  "Input exceeds maximum context length",
			expected: true,
		},
		{
			name:     "too many tokens",
			message:  "Request has too many tokens: 200000",
			expected: true,
		},
		{
			name:     "context window exceeded",
			message:  "The context window has been exceeded",
			expected: true,
		},
		{
			name:     "regular validation error",
			message:  "messages field is required",
			expected: false,
		},
		{
			name:     "different error",
			message:  "Model not found",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isContextLengthError(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		expected   bool
	}{
		{
			name:       "validation exception",
			statusCode: http.StatusBadRequest,
			message:    "ValidationException: messages is required",
			expected:   true,
		},
		{
			name:       "validation error message",
			statusCode: http.StatusBadRequest,
			message:    "Validation error: invalid parameter",
			expected:   true,
		},
		{
			name:       "invalid request",
			statusCode: http.StatusBadRequest,
			message:    "Invalid request body",
			expected:   true,
		},
		{
			name:       "400 without validation",
			statusCode: http.StatusBadRequest,
			message:    "Prompt is too long",
			expected:   false, // This is context length error
		},
		{
			name:       "500 error",
			statusCode: http.StatusInternalServerError,
			message:    "Validation failed",
			expected:   false, // Not a 400 error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidationError(tt.statusCode, tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractContextLengthInfo(t *testing.T) {
	tests := []struct {
		name               string
		message            string
		expectedRequested  int
		expectedMax        int
	}{
		{
			name:              "with token counts",
			message:           "Prompt is too long. The prompt has 150000 tokens but the model only supports 100000",
			expectedRequested: 150000,
			expectedMax:       100000,
		},
		{
			name:              "with different format",
			message:           "Input exceeds maximum: requested 200000, maximum 128000 tokens",
			expectedRequested: 200000,
			expectedMax:       128000,
		},
		{
			name:              "no specific numbers",
			message:           "Context length exceeded",
			expectedRequested: 0,
			expectedMax:       0,
		},
		{
			name:              "partial information",
			message:           "Prompt has 150000 tokens",
			expectedRequested: 150000,
			expectedMax:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requested, max := extractContextLengthInfo(tt.message)
			assert.Equal(t, tt.expectedRequested, requested)
			assert.Equal(t, tt.expectedMax, max)
		})
	}
}
