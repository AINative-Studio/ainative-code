package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLogger implements a simple test logger
type mockLogger struct{}

func (m *mockLogger) Debug(msg string)              {}
func (m *mockLogger) Info(msg string)               {}
func (m *mockLogger) Warn(msg string)               {}
func (m *mockLogger) Error(msg string)              {}
func (m *mockLogger) Fatal(msg string)              {}
func (m *mockLogger) WithField(key string, value interface{}) logger.LoggerInterface { return m }
func (m *mockLogger) WithFields(fields map[string]interface{}) logger.LoggerInterface { return m }
func (m *mockLogger) WithError(err error) logger.LoggerInterface                      { return m }

func TestNewOpenAIProvider(t *testing.T) {
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
			name: "valid configuration with custom base URL",
			config: Config{
				APIKey:  "test-api-key",
				BaseURL: "https://custom.openai.com/v1",
			},
			expectError: false,
		},
		{
			name: "valid configuration with organization",
			config: Config{
				APIKey:       "test-api-key",
				Organization: "org-123",
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
			p, err := NewOpenAIProvider(tt.config)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, p)
			} else {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.Equal(t, "openai", p.Name())
			}
		})
	}
}

func TestOpenAIProvider_Name(t *testing.T) {
	p, err := NewOpenAIProvider(Config{
		APIKey: "test-key",
	})
	require.NoError(t, err)

	assert.Equal(t, "openai", p.Name())
}

func TestOpenAIProvider_Models(t *testing.T) {
	p, err := NewOpenAIProvider(Config{
		APIKey: "test-key",
	})
	require.NoError(t, err)

	models := p.Models()
	assert.NotEmpty(t, models)
	assert.Contains(t, models, "gpt-4-turbo-preview")
	assert.Contains(t, models, "gpt-4")
	assert.Contains(t, models, "gpt-3.5-turbo")
	assert.Contains(t, models, "gpt-3.5-turbo-16k")
}

func TestOpenAIProvider_Chat(t *testing.T) {
	tests := []struct {
		name           string
		messages       []provider.Message
		options        []provider.ChatOption
		mockResponse   string
		mockStatusCode int
		expectError    bool
		errorContains  string
		validateResp   func(t *testing.T, resp provider.Response)
	}{
		{
			name: "successful chat request",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			options: []provider.ChatOption{
				provider.WithModel("gpt-4"),
			},
			mockResponse: `{
				"id": "chatcmpl-123",
				"object": "chat.completion",
				"created": 1677652288,
				"model": "gpt-4",
				"choices": [{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "Hello! How can I help you?"
					},
					"finish_reason": "stop"
				}],
				"usage": {
					"prompt_tokens": 10,
					"completion_tokens": 20,
					"total_tokens": 30
				}
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			validateResp: func(t *testing.T, resp provider.Response) {
				assert.Equal(t, "Hello! How can I help you?", resp.Content)
				assert.Equal(t, "gpt-4", resp.Model)
				assert.Equal(t, 10, resp.Usage.PromptTokens)
				assert.Equal(t, 20, resp.Usage.CompletionTokens)
				assert.Equal(t, 30, resp.Usage.TotalTokens)
			},
		},
		{
			name: "chat with system prompt",
			messages: []provider.Message{
				{Role: "user", Content: "What's 2+2?"},
			},
			options: []provider.ChatOption{
				provider.WithModel("gpt-3.5-turbo"),
				provider.WithSystemPrompt("You are a helpful math assistant."),
			},
			mockResponse: `{
				"id": "chatcmpl-456",
				"object": "chat.completion",
				"created": 1677652288,
				"model": "gpt-3.5-turbo",
				"choices": [{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "2+2 equals 4"
					},
					"finish_reason": "stop"
				}],
				"usage": {
					"prompt_tokens": 15,
					"completion_tokens": 8,
					"total_tokens": 23
				}
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			validateResp: func(t *testing.T, resp provider.Response) {
				assert.Equal(t, "2+2 equals 4", resp.Content)
			},
		},
		{
			name: "invalid model error",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			options: []provider.ChatOption{
				provider.WithModel("invalid-model"),
			},
			mockStatusCode: http.StatusOK,
			expectError:    true,
			errorContains:  "invalid model",
		},
		{
			name: "authentication error",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			options: []provider.ChatOption{
				provider.WithModel("gpt-4"),
			},
			mockResponse: `{
				"error": {
					"message": "Invalid API key",
					"type": "invalid_request_error",
					"code": "invalid_api_key"
				}
			}`,
			mockStatusCode: http.StatusUnauthorized,
			expectError:    true,
			errorContains:  "authentication",
		},
		{
			name: "rate limit error",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			options: []provider.ChatOption{
				provider.WithModel("gpt-4"),
			},
			mockResponse: `{
				"error": {
					"message": "Rate limit exceeded",
					"type": "rate_limit_error"
				}
			}`,
			mockStatusCode: http.StatusTooManyRequests,
			expectError:    true,
			errorContains:  "rate limit",
		},
		{
			name: "context length error",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			options: []provider.ChatOption{
				provider.WithModel("gpt-4"),
			},
			mockResponse: `{
				"error": {
					"message": "This model's maximum context length is 8192 tokens",
					"type": "context_length_exceeded"
				}
			}`,
			mockStatusCode: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "context length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify headers
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "Bearer "))

				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Create provider with mock server
			p, err := NewOpenAIProvider(Config{
				APIKey:  "test-key",
				BaseURL: server.URL,
				Logger:  &mockLogger{},
			})
			require.NoError(t, err)

			// Execute chat
			ctx := context.Background()
			resp, err := p.Chat(ctx, tt.messages, tt.options...)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errorContains))
				}
			} else {
				require.NoError(t, err)
				if tt.validateResp != nil {
					tt.validateResp(t, resp)
				}
			}
		})
	}
}

func TestOpenAIProvider_Stream(t *testing.T) {
	tests := []struct {
		name           string
		messages       []provider.Message
		options        []provider.StreamOption
		mockResponse   string
		mockStatusCode int
		expectError    bool
		validateEvents func(t *testing.T, events []provider.Event)
	}{
		{
			name: "successful streaming",
			messages: []provider.Message{
				{Role: "user", Content: "Count to 3"},
			},
			options: []provider.StreamOption{
				provider.StreamWithModel("gpt-3.5-turbo"),
			},
			mockResponse: `data: {"id":"1","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"1","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":"1"},"finish_reason":null}]}

data: {"id":"1","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":", "},"finish_reason":null}]}

data: {"id":"1","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":"2"},"finish_reason":null}]}

data: {"id":"1","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":", "},"finish_reason":null}]}

data: {"id":"1","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":"3"},"finish_reason":null}]}

data: {"id":"1","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]

`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			validateEvents: func(t *testing.T, events []provider.Event) {
				assert.NotEmpty(t, events)

				// First event should be start
				assert.Equal(t, provider.EventTypeContentStart, events[0].Type)

				// Last event should be end with done=true
				lastEvent := events[len(events)-1]
				assert.Equal(t, provider.EventTypeContentEnd, lastEvent.Type)
				assert.True(t, lastEvent.Done)

				// Collect all content deltas
				var content strings.Builder
				for _, event := range events {
					if event.Type == provider.EventTypeContentDelta {
						content.WriteString(event.Content)
					}
				}
				assert.Equal(t, "1, 2, 3", content.String())
			},
		},
		{
			name: "streaming with invalid model",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			options: []provider.StreamOption{
				provider.StreamWithModel("invalid-model"),
			},
			mockStatusCode: http.StatusOK,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request body contains stream=true
				var reqBody openAIRequest
				json.NewDecoder(r.Body).Decode(&reqBody)
				assert.True(t, reqBody.Stream)

				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Create provider with mock server
			p, err := NewOpenAIProvider(Config{
				APIKey:  "test-key",
				BaseURL: server.URL,
				Logger:  &mockLogger{},
			})
			require.NoError(t, err)

			// Execute stream
			ctx := context.Background()
			eventChan, err := p.Stream(ctx, tt.messages, tt.options...)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, eventChan)

				// Collect all events
				var events []provider.Event
				for event := range eventChan {
					events = append(events, event)
					if event.Error != nil {
						t.Logf("Stream error: %v", event.Error)
					}
				}

				if tt.validateEvents != nil {
					tt.validateEvents(t, events)
				}
			}
		})
	}
}

func TestOpenAIProvider_ConvertMessages(t *testing.T) {
	p, err := NewOpenAIProvider(Config{
		APIKey: "test-key",
	})
	require.NoError(t, err)

	tests := []struct {
		name         string
		messages     []provider.Message
		systemPrompt string
		expected     int
		validateMsg  func(t *testing.T, msgs []openAIMessage)
	}{
		{
			name: "convert user messages",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
				{Role: "user", Content: "How are you?"},
			},
			systemPrompt: "",
			expected:     3,
			validateMsg: func(t *testing.T, msgs []openAIMessage) {
				assert.Equal(t, "user", msgs[0].Role)
				assert.Equal(t, "Hello", msgs[0].Content)
				assert.Equal(t, "assistant", msgs[1].Role)
				assert.Equal(t, "Hi there!", msgs[1].Content)
			},
		},
		{
			name: "with system prompt",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			systemPrompt: "You are helpful",
			expected:     2,
			validateMsg: func(t *testing.T, msgs []openAIMessage) {
				assert.Equal(t, "system", msgs[0].Role)
				assert.Equal(t, "You are helpful", msgs[0].Content)
				assert.Equal(t, "user", msgs[1].Role)
			},
		},
		{
			name: "system message in messages array",
			messages: []provider.Message{
				{Role: "system", Content: "Be concise"},
				{Role: "user", Content: "Hello"},
			},
			systemPrompt: "",
			expected:     2,
			validateMsg: func(t *testing.T, msgs []openAIMessage) {
				assert.Equal(t, "system", msgs[0].Role)
				assert.Equal(t, "Be concise", msgs[0].Content)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.convertMessages(tt.messages, tt.systemPrompt)
			assert.Len(t, result, tt.expected)
			if tt.validateMsg != nil {
				tt.validateMsg(t, result)
			}
		})
	}
}

func TestOpenAIProvider_Close(t *testing.T) {
	p, err := NewOpenAIProvider(Config{
		APIKey: "test-key",
	})
	require.NoError(t, err)

	err = p.Close()
	assert.NoError(t, err)
}

func TestOpenAIProvider_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices":[{"message":{"content":"test"}}],"usage":{}}`))
	}))
	defer server.Close()

	p, err := NewOpenAIProvider(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	messages := []provider.Message{{Role: "user", Content: "test"}}
	_, err = p.Chat(ctx, messages, provider.WithModel("gpt-4"))

	// Should get context deadline exceeded error
	assert.Error(t, err)
}

func TestOpenAIProvider_WithOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify organization header
		assert.Equal(t, "org-test-123", r.Header.Get("OpenAI-Organization"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "test",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-4",
			"choices": [{"index": 0, "message": {"role": "assistant", "content": "test"}, "finish_reason": "stop"}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
		}`))
	}))
	defer server.Close()

	p, err := NewOpenAIProvider(Config{
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Organization: "org-test-123",
	})
	require.NoError(t, err)

	ctx := context.Background()
	messages := []provider.Message{{Role: "user", Content: "test"}}
	_, err = p.Chat(ctx, messages, provider.WithModel("gpt-4"))
	require.NoError(t, err)
}

func TestOpenAIProvider_ParseResponseEdgeCases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "test",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-4",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": [
						{"type": "text", "text": "Part 1"},
						{"type": "text", "text": "Part 2"}
					]
				},
				"finish_reason": "stop"
			}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
		}`))
	}))
	defer server.Close()

	p, err := NewOpenAIProvider(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	messages := []provider.Message{{Role: "user", Content: "test"}}
	resp, err := p.Chat(ctx, messages, provider.WithModel("gpt-4"))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
}

func TestOpenAIProvider_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "test",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-4",
			"choices": [],
			"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
		}`))
	}))
	defer server.Close()

	p, err := NewOpenAIProvider(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	messages := []provider.Message{{Role: "user", Content: "test"}}
	_, err = p.Chat(ctx, messages, provider.WithModel("gpt-4"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no choices")
}

func TestOpenAIProvider_ModelNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{
			"error": {
				"message": "The model 'invalid-model' does not exist",
				"type": "invalid_request_error"
			}
		}`))
	}))
	defer server.Close()

	p, err := NewOpenAIProvider(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	messages := []provider.Message{{Role: "user", Content: "test"}}
	_, err = p.Chat(ctx, messages, provider.WithModel("gpt-4"))
	require.Error(t, err)
}

func TestOpenAIProvider_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody openAIRequest
		json.NewDecoder(r.Body).Decode(&reqBody)

		// Verify options were applied
		assert.Equal(t, 100, reqBody.MaxTokens)
		assert.NotNil(t, reqBody.Temperature)
		assert.Equal(t, 0.8, *reqBody.Temperature)
		assert.NotNil(t, reqBody.TopP)
		assert.Equal(t, 0.9, *reqBody.TopP)
		assert.Len(t, reqBody.Stop, 1)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "test",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-4",
			"choices": [{"index": 0, "message": {"role": "assistant", "content": "test"}, "finish_reason": "stop"}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
		}`))
	}))
	defer server.Close()

	p, err := NewOpenAIProvider(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	messages := []provider.Message{{Role: "user", Content: "test"}}
	_, err = p.Chat(ctx, messages,
		provider.WithModel("gpt-4"),
		provider.WithMaxTokens(100),
		provider.WithTemperature(0.8),
		provider.WithTopP(0.9),
		provider.WithStopSequences("STOP"),
	)
	require.NoError(t, err)
}
