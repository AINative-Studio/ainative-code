package meta

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetaError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *MetaError
		expected string
	}{
		{
			name: "with code",
			err: &MetaError{
				StatusCode: 401,
				Type:       ErrTypeAuthentication,
				Message:    "Invalid API key",
				Code:       "invalid_api_key",
			},
			expected: "meta api error (status 401, type: authentication_error, code: invalid_api_key): Invalid API key",
		},
		{
			name: "without code",
			err: &MetaError{
				StatusCode: 429,
				Type:       ErrTypeRateLimit,
				Message:    "Rate limit exceeded",
			},
			expected: "meta api error (status 429, type: rate_limit_error): Rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestMetaError_IsAuthenticationError(t *testing.T) {
	tests := []struct {
		name     string
		err      *MetaError
		expected bool
	}{
		{
			name:     "authentication error",
			err:      &MetaError{Type: ErrTypeAuthentication},
			expected: true,
		},
		{
			name:     "invalid API key",
			err:      &MetaError{Type: ErrTypeInvalidAPIKey},
			expected: true,
		},
		{
			name:     "other error",
			err:      &MetaError{Type: ErrTypeRateLimit},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.IsAuthenticationError())
		})
	}
}

func TestMetaError_IsRateLimitError(t *testing.T) {
	tests := []struct {
		name     string
		err      *MetaError
		expected bool
	}{
		{
			name:     "rate limit type",
			err:      &MetaError{Type: ErrTypeRateLimit},
			expected: true,
		},
		{
			name:     "429 status code",
			err:      &MetaError{StatusCode: 429},
			expected: true,
		},
		{
			name:     "other error",
			err:      &MetaError{StatusCode: 400, Type: ErrTypeInvalidRequest},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.IsRateLimitError())
		})
	}
}

func TestMetaError_IsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      *MetaError
		expected bool
	}{
		{
			name:     "rate limit",
			err:      &MetaError{Type: ErrTypeRateLimit},
			expected: true,
		},
		{
			name:     "timeout",
			err:      &MetaError{Type: ErrTypeTimeout},
			expected: true,
		},
		{
			name:     "500 error",
			err:      &MetaError{StatusCode: 500},
			expected: true,
		},
		{
			name:     "503 error",
			err:      &MetaError{StatusCode: 503},
			expected: true,
		},
		{
			name:     "authentication error",
			err:      &MetaError{StatusCode: 401, Type: ErrTypeAuthentication},
			expected: false,
		},
		{
			name:     "invalid request",
			err:      &MetaError{StatusCode: 400, Type: ErrTypeInvalidRequest},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.IsRetryable())
		})
	}
}

func TestParseErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		statusCode int
		wantErr    bool
		checkError func(*testing.T, error)
	}{
		{
			name: "valid error response",
			body: `{"error":{"message":"Invalid API key","type":"authentication_error","code":"invalid_api_key"}}`,
			statusCode: 401,
			wantErr:    true,
			checkError: func(t *testing.T, err error) {
				metaErr, ok := err.(*MetaError)
				assert.True(t, ok)
				assert.Equal(t, 401, metaErr.StatusCode)
				assert.Equal(t, "authentication_error", metaErr.Type)
				assert.Equal(t, "Invalid API key", metaErr.Message)
				assert.Equal(t, "invalid_api_key", metaErr.Code)
			},
		},
		{
			name:       "invalid JSON",
			body:       `{invalid json`,
			statusCode: 500,
			wantErr:    true,
			checkError: func(t *testing.T, err error) {
				metaErr, ok := err.(*MetaError)
				assert.True(t, ok)
				assert.Equal(t, 500, metaErr.StatusCode)
				assert.Contains(t, metaErr.Message, "failed to parse error response")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewReader([]byte(tt.body))),
			}

			err := parseErrorResponse(resp)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkError != nil {
					tt.checkError(t, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleHTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		contentType string
		wantErr    bool
		checkError func(*testing.T, error)
	}{
		{
			name:       "success 200",
			statusCode: 200,
			wantErr:    false,
		},
		{
			name:        "400 with JSON error",
			statusCode:  400,
			contentType: "application/json",
			body:        `{"error":{"message":"Bad request","type":"invalid_request_error"}}`,
			wantErr:     true,
			checkError: func(t *testing.T, err error) {
				metaErr, ok := err.(*MetaError)
				assert.True(t, ok)
				assert.Equal(t, 400, metaErr.StatusCode)
				assert.Equal(t, "invalid_request_error", metaErr.Type)
			},
		},
		{
			name:       "401 without JSON",
			statusCode: 401,
			wantErr:    true,
			checkError: func(t *testing.T, err error) {
				metaErr, ok := err.(*MetaError)
				assert.True(t, ok)
				assert.Equal(t, 401, metaErr.StatusCode)
				assert.Equal(t, ErrTypeAuthentication, metaErr.Type)
				assert.Contains(t, metaErr.Message, "Invalid or missing API key")
			},
		},
		{
			name:       "429 rate limit",
			statusCode: 429,
			wantErr:    true,
			checkError: func(t *testing.T, err error) {
				metaErr, ok := err.(*MetaError)
				assert.True(t, ok)
				assert.Equal(t, 429, metaErr.StatusCode)
				assert.Equal(t, ErrTypeRateLimit, metaErr.Type)
			},
		},
		{
			name:       "500 server error",
			statusCode: 500,
			wantErr:    true,
			checkError: func(t *testing.T, err error) {
				metaErr, ok := err.(*MetaError)
				assert.True(t, ok)
				assert.Equal(t, 500, metaErr.StatusCode)
				assert.Contains(t, metaErr.Message, "server error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := tt.body
			if body == "" {
				body = "{}"
			}

			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewReader([]byte(body))),
				Header:     http.Header{},
			}

			if tt.contentType != "" {
				resp.Header.Set("Content-Type", tt.contentType)
			}

			err := handleHTTPError(resp)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkError != nil {
					tt.checkError(t, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewErrors(t *testing.T) {
	t.Run("NewAuthenticationError", func(t *testing.T) {
		err := NewAuthenticationError("test message")
		assert.Equal(t, 401, err.StatusCode)
		assert.Equal(t, ErrTypeAuthentication, err.Type)
		assert.Equal(t, "test message", err.Message)
	})

	t.Run("NewInvalidRequestError", func(t *testing.T) {
		err := NewInvalidRequestError("test message")
		assert.Equal(t, 400, err.StatusCode)
		assert.Equal(t, ErrTypeInvalidRequest, err.Type)
		assert.Equal(t, "test message", err.Message)
	})

	t.Run("NewRateLimitError", func(t *testing.T) {
		err := NewRateLimitError("test message")
		assert.Equal(t, 429, err.StatusCode)
		assert.Equal(t, ErrTypeRateLimit, err.Type)
		assert.Equal(t, "test message", err.Message)
	})

	t.Run("NewTimeoutError", func(t *testing.T) {
		err := NewTimeoutError("test message")
		assert.Equal(t, 408, err.StatusCode)
		assert.Equal(t, ErrTypeTimeout, err.Type)
		assert.Equal(t, "test message", err.Message)
	})
}
