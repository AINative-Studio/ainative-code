package embeddings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAINativeEmbeddingsClient(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			config: Config{
				APIKey: "test-api-key",
			},
			expectError: false,
		},
		{
			name: "valid with custom endpoint",
			config: Config{
				APIKey:   "test-api-key",
				Endpoint: "https://custom.api.com/embeddings",
			},
			expectError: false,
		},
		{
			name:        "missing API key",
			config:      Config{},
			expectError: true,
			errorMsg:    "API key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewAINativeEmbeddingsClient(tt.config)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
				assert.NotEmpty(t, client.endpoint)
			}
		})
	}
}

func TestAINativeEmbeddingsClient_Embed(t *testing.T) {
	tests := []struct {
		name           string
		texts          []string
		model          string
		mockResponse   string
		mockStatusCode int
		expectError    bool
		errorContains  string
		validateResult func(t *testing.T, result *EmbeddingResult)
	}{
		{
			name:  "successful embedding",
			texts: []string{"Hello world", "Test text"},
			model: "text-embedding-ada-002",
			mockResponse: `{
				"embeddings": [
					[0.1, 0.2, 0.3],
					[0.4, 0.5, 0.6]
				],
				"model": "text-embedding-ada-002",
				"usage": {
					"total_tokens": 10
				}
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			validateResult: func(t *testing.T, result *EmbeddingResult) {
				assert.Len(t, result.Embeddings, 2)
				assert.Len(t, result.Embeddings[0], 3)
				assert.Equal(t, float32(0.1), result.Embeddings[0][0])
				assert.Equal(t, float32(0.2), result.Embeddings[0][1])
				assert.Equal(t, float32(0.3), result.Embeddings[0][2])
				assert.Equal(t, "text-embedding-ada-002", result.Model)
				assert.Equal(t, 10, result.TotalTokens)
			},
		},
		{
			name:  "single text embedding",
			texts: []string{"Single text"},
			model: "default",
			mockResponse: `{
				"embeddings": [[0.1, 0.2, 0.3, 0.4, 0.5]],
				"model": "default",
				"usage": {"total_tokens": 5}
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			validateResult: func(t *testing.T, result *EmbeddingResult) {
				assert.Len(t, result.Embeddings, 1)
				assert.Len(t, result.Embeddings[0], 5)
			},
		},
		{
			name:          "empty texts",
			texts:         []string{},
			model:         "default",
			expectError:   true,
			errorContains: "at least one text is required",
		},
		{
			name:          "batch size too large",
			texts:         make([]string, MaxBatchSize+1),
			model:         "default",
			expectError:   true,
			errorContains: "batch size exceeds maximum",
		},
		{
			name:  "authentication error",
			texts: []string{"test"},
			model: "default",
			mockResponse: `{
				"error": {
					"message": "Invalid API key",
					"type": "authentication_error",
					"code": "invalid_api_key"
				}
			}`,
			mockStatusCode: http.StatusUnauthorized,
			expectError:    true,
			errorContains:  "status 401",
		},
		{
			name:  "rate limit error",
			texts: []string{"test"},
			model: "default",
			mockResponse: `{
				"error": {
					"message": "Rate limit exceeded",
					"type": "rate_limit_error",
					"code": "rate_limit"
				}
			}`,
			mockStatusCode: http.StatusTooManyRequests,
			expectError:    true,
			errorContains:  "status 429",
		},
		{
			name:  "quota exceeded error",
			texts: []string{"test"},
			model: "default",
			mockResponse: `{
				"error": {
					"message": "Quota exceeded",
					"type": "quota_exceeded",
					"code": "quota_exceeded"
				}
			}`,
			mockStatusCode: http.StatusForbidden,
			expectError:    true,
			errorContains:  "status 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify headers
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.True(t, len(r.Header.Get("Authorization")) > 0)
				assert.Contains(t, r.Header.Get("Authorization"), "Bearer")

				// Verify request body
				var reqBody embeddingRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				if err == nil {
					assert.True(t, reqBody.Normalize) // Should always normalize
				}

				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Create client with mock server
			client, err := NewAINativeEmbeddingsClient(Config{
				APIKey:   "test-key",
				Endpoint: server.URL,
			})
			require.NoError(t, err)

			// Execute embedding
			ctx := context.Background()
			result, err := client.Embed(ctx, tt.texts, tt.model)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestAINativeEmbeddingsClient_Retry(t *testing.T) {
	attempts := 0

	// Create server that fails first 2 times, then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": {"message": "Server error", "type": "server_error"}}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"embeddings": [[0.1, 0.2, 0.3]],
			"model": "default",
			"usage": {"total_tokens": 5}
		}`))
	}))
	defer server.Close()

	client, err := NewAINativeEmbeddingsClient(Config{
		APIKey:     "test-key",
		Endpoint:   server.URL,
		MaxRetries: 3,
		RetryDelay: 10 * time.Millisecond,
	})
	require.NoError(t, err)

	ctx := context.Background()
	result, err := client.Embed(ctx, []string{"test"}, "default")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 3, attempts) // Should have retried twice
}

func TestAINativeEmbeddingsClient_NoRetryOnClientError(t *testing.T) {
	attempts := 0

	// Create server that always returns 400
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"message": "Bad request", "type": "invalid_request"}}`))
	}))
	defer server.Close()

	client, err := NewAINativeEmbeddingsClient(Config{
		APIKey:     "test-key",
		Endpoint:   server.URL,
		MaxRetries: 3,
		RetryDelay: 10 * time.Millisecond,
	})
	require.NoError(t, err)

	ctx := context.Background()
	_, err = client.Embed(ctx, []string{"test"}, "default")

	require.Error(t, err)
	assert.Equal(t, 1, attempts) // Should NOT have retried
}

func TestAINativeEmbeddingsClient_ContextCancellation(t *testing.T) {
	// Create server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"embeddings": [[0.1]], "model": "default", "usage": {"total_tokens": 1}}`))
	}))
	defer server.Close()

	client, err := NewAINativeEmbeddingsClient(Config{
		APIKey:   "test-key",
		Endpoint: server.URL,
	})
	require.NoError(t, err)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = client.Embed(ctx, []string{"test"}, "default")
	assert.Error(t, err) // Should timeout
}

func TestEmbeddingAPIError_Methods(t *testing.T) {
	tests := []struct {
		name                   string
		err                    *EmbeddingAPIError
		isAuth                 bool
		isRateLimit            bool
		isQuotaExceeded        bool
	}{
		{
			name: "authentication error",
			err: &EmbeddingAPIError{
				StatusCode: http.StatusUnauthorized,
				Message:    "Invalid API key",
				Type:       "authentication_error",
			},
			isAuth:          true,
			isRateLimit:     false,
			isQuotaExceeded: false,
		},
		{
			name: "rate limit error",
			err: &EmbeddingAPIError{
				StatusCode: http.StatusTooManyRequests,
				Message:    "Rate limit exceeded",
				Type:       "rate_limit_error",
			},
			isAuth:          false,
			isRateLimit:     true,
			isQuotaExceeded: false,
		},
		{
			name: "quota exceeded error",
			err: &EmbeddingAPIError{
				StatusCode: http.StatusForbidden,
				Message:    "Quota exceeded",
				Type:       "quota_exceeded",
			},
			isAuth:          true, // 403 is considered auth error
			isRateLimit:     false,
			isQuotaExceeded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isAuth, tt.err.IsAuthenticationError())
			assert.Equal(t, tt.isRateLimit, tt.err.IsRateLimitError())
			assert.Equal(t, tt.isQuotaExceeded, tt.err.IsQuotaExceededError())
			assert.NotEmpty(t, tt.err.Error())
		})
	}
}

func TestAINativeEmbeddingsClient_Close(t *testing.T) {
	client, err := NewAINativeEmbeddingsClient(Config{
		APIKey: "test-key",
	})
	require.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
}

func TestAINativeEmbeddingsClient_DefaultModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody embeddingRequest
		json.NewDecoder(r.Body).Decode(&reqBody)

		// When empty model is passed, it should default to "default"
		assert.Equal(t, "default", reqBody.Model)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"embeddings": [[0.1, 0.2]],
			"model": "default",
			"usage": {"total_tokens": 3}
		}`))
	}))
	defer server.Close()

	client, err := NewAINativeEmbeddingsClient(Config{
		APIKey:   "test-key",
		Endpoint: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	result, err := client.Embed(ctx, []string{"test"}, "") // Empty model

	require.NoError(t, err)
	assert.Equal(t, "default", result.Model)
}
