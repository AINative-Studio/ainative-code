package bedrock

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBedrockProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config with access key and secret",
			config: Config{
				Region:    "us-east-1",
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
			},
			expectError: false,
		},
		{
			name: "valid config with session token",
			config: Config{
				Region:       "us-west-2",
				AccessKey:    "test-access-key",
				SecretKey:    "test-secret-key",
				SessionToken: "test-session-token",
			},
			expectError: false,
		},
		{
			name: "valid config with custom endpoint",
			config: Config{
				Region:    "us-east-1",
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
				Endpoint:  "https://custom.bedrock.endpoint.com",
			},
			expectError: false,
		},
		{
			name: "missing region defaults to us-east-1",
			config: Config{
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
			},
			expectError: false,
		},
		{
			name: "missing access key",
			config: Config{
				Region:    "us-east-1",
				SecretKey: "test-secret-key",
			},
			expectError: true,
			errorMsg:    "AccessKey is required",
		},
		{
			name: "missing secret key",
			config: Config{
				Region:    "us-east-1",
				AccessKey: "test-access-key",
			},
			expectError: true,
			errorMsg:    "SecretKey is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewBedrockProvider(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, p)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, p)
				assert.Equal(t, "bedrock", p.Name())

				// Check defaults
				if tt.config.Region == "" {
					assert.Equal(t, "us-east-1", p.region)
				}
			}
		})
	}
}

func TestBedrockProvider_Name(t *testing.T) {
	p, err := NewBedrockProvider(Config{
		Region:    "us-east-1",
		AccessKey: "test-key",
		SecretKey: "test-secret",
	})
	require.NoError(t, err)

	assert.Equal(t, "bedrock", p.Name())
}

func TestBedrockProvider_Models(t *testing.T) {
	p, err := NewBedrockProvider(Config{
		Region:    "us-east-1",
		AccessKey: "test-key",
		SecretKey: "test-secret",
	})
	require.NoError(t, err)

	models := p.Models()
	assert.NotNil(t, models)
	assert.Greater(t, len(models), 0)

	// Verify all expected Claude models are present
	expectedModels := []string{
		"anthropic.claude-3-5-sonnet-20241022-v2:0",
		"anthropic.claude-3-opus-20240229-v1:0",
		"anthropic.claude-3-sonnet-20240229-v1:0",
		"anthropic.claude-3-haiku-20240307-v1:0",
		"anthropic.claude-v2",
		"anthropic.claude-instant-v1",
	}

	for _, expected := range expectedModels {
		assert.Contains(t, models, expected)
	}
}

func TestBedrockProvider_Chat(t *testing.T) {
	tests := []struct {
		name           string
		messages       []provider.Message
		options        []provider.ChatOption
		mockResponse   string
		mockStatusCode int
		expectError    bool
		errorType      string
		validateReq    func(t *testing.T, req *http.Request)
	}{
		{
			name: "successful chat completion",
			messages: []provider.Message{
				{Role: "user", Content: "Hello, Claude!"},
			},
			options: []provider.ChatOption{
				provider.WithModel("anthropic.claude-3-5-sonnet-20241022-v2:0"),
				provider.WithMaxTokens(1024),
			},
			mockResponse: `{
				"output": {
					"message": {
						"role": "assistant",
						"content": [{"text": "Hello! How can I help you?"}]
					}
				},
				"usage": {
					"inputTokens": 10,
					"outputTokens": 20
				}
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			validateReq: func(t *testing.T, req *http.Request) {
				assert.Equal(t, "POST", req.Method)
				assert.Contains(t, req.URL.Path, "/model/")
				assert.Contains(t, req.URL.Path, "/invoke")
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

				// Check for AWS signature headers
				assert.NotEmpty(t, req.Header.Get("Authorization"))
				assert.NotEmpty(t, req.Header.Get("X-Amz-Date"))
			},
		},
		{
			name: "with system prompt",
			messages: []provider.Message{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "Hello!"},
			},
			options: []provider.ChatOption{
				provider.WithModel("anthropic.claude-3-haiku-20240307-v1:0"),
				provider.WithMaxTokens(512),
			},
			mockResponse: `{
				"output": {
					"message": {
						"role": "assistant",
						"content": [{"text": "Hi there!"}]
					}
				},
				"usage": {
					"inputTokens": 15,
					"outputTokens": 5
				}
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name: "with temperature and top_p",
			messages: []provider.Message{
				{Role: "user", Content: "Test"},
			},
			options: []provider.ChatOption{
				provider.WithModel("anthropic.claude-3-opus-20240229-v1:0"),
				provider.WithMaxTokens(100),
				provider.WithTemperature(0.7),
				provider.WithTopP(0.9),
			},
			mockResponse: `{
				"output": {
					"message": {
						"role": "assistant",
						"content": [{"text": "Response"}]
					}
				},
				"usage": {
					"inputTokens": 5,
					"outputTokens": 3
				}
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name: "invalid model",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			options: []provider.ChatOption{
				provider.WithModel("invalid-model"),
			},
			expectError: true,
			errorType:   "invalid_model",
		},
		{
			name: "authentication error",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			options: []provider.ChatOption{
				provider.WithModel("anthropic.claude-3-5-sonnet-20241022-v2:0"),
			},
			mockResponse: `{
				"message": "The security token included in the request is invalid."
			}`,
			mockStatusCode: http.StatusForbidden,
			expectError:    true,
			errorType:      "authentication",
		},
		{
			name: "throttling error",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			options: []provider.ChatOption{
				provider.WithModel("anthropic.claude-3-5-sonnet-20241022-v2:0"),
			},
			mockResponse: `{
				"message": "Rate exceeded"
			}`,
			mockStatusCode: http.StatusTooManyRequests,
			expectError:    true,
			errorType:      "rate_limit",
		},
		{
			name: "validation error",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			options: []provider.ChatOption{
				provider.WithModel("anthropic.claude-3-5-sonnet-20241022-v2:0"),
			},
			mockResponse: `{
				"message": "Validation error: max_tokens is required"
			}`,
			mockStatusCode: http.StatusBadRequest,
			expectError:    true,
			errorType:      "provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.validateReq != nil {
					tt.validateReq(t, r)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Create provider with custom endpoint
			p, err := NewBedrockProvider(Config{
				Region:    "us-east-1",
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
				Endpoint:  server.URL,
			})
			require.NoError(t, err)

			// Execute chat
			ctx := context.Background()
			resp, err := p.Chat(ctx, tt.messages, tt.options...)

			if tt.expectError {
				assert.Error(t, err)

				// Check error type
				switch tt.errorType {
				case "invalid_model":
					var invalidModelErr *provider.InvalidModelError
					assert.ErrorAs(t, err, &invalidModelErr)
				case "authentication":
					var authErr *provider.AuthenticationError
					assert.ErrorAs(t, err, &authErr)
				case "rate_limit":
					var rateLimitErr *provider.RateLimitError
					assert.ErrorAs(t, err, &rateLimitErr)
				case "provider":
					var providerErr *provider.ProviderError
					assert.ErrorAs(t, err, &providerErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Content)
				assert.Greater(t, resp.Usage.PromptTokens, 0)
				assert.Greater(t, resp.Usage.CompletionTokens, 0)
				assert.Equal(t, resp.Usage.PromptTokens+resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
			}
		})
	}
}

func TestBedrockProvider_Stream(t *testing.T) {
	tests := []struct {
		name         string
		messages     []provider.Message
		options      []provider.StreamOption
		mockEvents   []string
		expectError  bool
		expectedText string
	}{
		{
			name: "successful streaming",
			messages: []provider.Message{
				{Role: "user", Content: "Hello!"},
			},
			options: []provider.StreamOption{
				provider.StreamWithModel("anthropic.claude-3-5-sonnet-20241022-v2:0"),
				provider.StreamWithMaxTokens(100),
			},
			mockEvents: []string{
				`{"messageStart":{"role":"assistant"}}` + "\n",
				`{"contentBlockDelta":{"delta":{"text":"Hello"},"contentBlockIndex":0}}` + "\n",
				`{"contentBlockDelta":{"delta":{"text":" there"},"contentBlockIndex":0}}` + "\n",
				`{"contentBlockDelta":{"delta":{"text":"!"},"contentBlockIndex":0}}` + "\n",
				`{"messageStop":{}}` + "\n",
			},
			expectError:  false,
			expectedText: "Hello there!",
		},
		{
			name: "streaming with error",
			messages: []provider.Message{
				{Role: "user", Content: "Test"},
			},
			options: []provider.StreamOption{
				provider.StreamWithModel("anthropic.claude-3-haiku-20240307-v1:0"),
			},
			mockEvents: []string{
				`{"messageStart":{"role":"assistant"}}` + "\n",
				`{"error":{"message":"Internal error"}}` + "\n",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
				w.WriteHeader(http.StatusOK)

				flusher, ok := w.(http.Flusher)
				require.True(t, ok)

				for _, event := range tt.mockEvents {
					w.Write([]byte(event))
					flusher.Flush()
					time.Sleep(10 * time.Millisecond)
				}
			}))
			defer server.Close()

			// Create provider
			p, err := NewBedrockProvider(Config{
				Region:    "us-east-1",
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
				Endpoint:  server.URL,
			})
			require.NoError(t, err)

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Execute stream
			eventChan, err := p.Stream(ctx, tt.messages, tt.options...)
			require.NoError(t, err)

			// Collect events
			var events []provider.Event
			var fullText string
			var gotError error

			for event := range eventChan {
				events = append(events, event)

				if event.Type == provider.EventTypeContentDelta {
					fullText += event.Content
				}

				if event.Type == provider.EventTypeError {
					gotError = event.Error
				}
			}

			if tt.expectError {
				assert.Error(t, gotError)
			} else {
				assert.NoError(t, gotError)
				assert.Equal(t, tt.expectedText, fullText)

				// Verify event sequence
				assert.Greater(t, len(events), 0)
				assert.Equal(t, provider.EventTypeContentStart, events[0].Type)
				assert.Equal(t, provider.EventTypeContentEnd, events[len(events)-1].Type)
				assert.True(t, events[len(events)-1].Done)
			}
		})
	}
}

func TestBedrockProvider_Close(t *testing.T) {
	p, err := NewBedrockProvider(Config{
		Region:    "us-east-1",
		AccessKey: "test-key",
		SecretKey: "test-secret",
	})
	require.NoError(t, err)

	err = p.Close()
	assert.NoError(t, err)
}

func TestBedrockProvider_BuildInvokeURL(t *testing.T) {
	p, err := NewBedrockProvider(Config{
		Region:    "us-east-1",
		AccessKey: "test-key",
		SecretKey: "test-secret",
	})
	require.NoError(t, err)

	tests := []struct {
		name     string
		model    string
		stream   bool
		expected string
	}{
		{
			name:     "non-streaming invoke",
			model:    "anthropic.claude-3-5-sonnet-20241022-v2:0",
			stream:   false,
			expected: "/model/anthropic.claude-3-5-sonnet-20241022-v2:0/invoke",
		},
		{
			name:     "streaming invoke",
			model:    "anthropic.claude-3-haiku-20240307-v1:0",
			stream:   true,
			expected: "/model/anthropic.claude-3-haiku-20240307-v1:0/invoke-with-response-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := p.buildInvokeURL(tt.model, tt.stream)
			assert.Contains(t, url, tt.expected)
		})
	}
}

func TestBedrockProvider_ParseResponse(t *testing.T) {
	p, err := NewBedrockProvider(Config{
		Region:    "us-east-1",
		AccessKey: "test-key",
		SecretKey: "test-secret",
	})
	require.NoError(t, err)

	tests := []struct {
		name            string
		responseJSON    string
		model           string
		expectedContent string
		expectError     bool
	}{
		{
			name: "single text content",
			responseJSON: `{
				"output": {
					"message": {
						"role": "assistant",
						"content": [{"text": "Hello!"}]
					}
				},
				"usage": {
					"inputTokens": 10,
					"outputTokens": 5
				}
			}`,
			model:           "anthropic.claude-3-5-sonnet-20241022-v2:0",
			expectedContent: "Hello!",
			expectError:     false,
		},
		{
			name: "multiple text contents",
			responseJSON: `{
				"output": {
					"message": {
						"role": "assistant",
						"content": [
							{"text": "First part."},
							{"text": "Second part."}
						]
					}
				},
				"usage": {
					"inputTokens": 20,
					"outputTokens": 10
				}
			}`,
			model:           "anthropic.claude-3-haiku-20240307-v1:0",
			expectedContent: "First part.\nSecond part.",
			expectError:     false,
		},
		{
			name:         "invalid JSON",
			responseJSON: `{invalid json`,
			model:        "test-model",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := p.parseResponse([]byte(tt.responseJSON), tt.model)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedContent, resp.Content)
				assert.Equal(t, tt.model, resp.Model)
				assert.Greater(t, resp.Usage.TotalTokens, 0)
			}
		})
	}
}
