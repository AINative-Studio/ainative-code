package anthropic

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnthropicProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config with API key",
			config: Config{
				APIKey: "test-api-key",
			},
			expectError: false,
		},
		{
			name: "valid config with custom base URL",
			config: Config{
				APIKey:  "test-api-key",
				BaseURL: "https://custom.api.com/v1",
			},
			expectError: false,
		},
		{
			name: "missing API key",
			config: Config{
				BaseURL: "https://api.anthropic.com/v1",
			},
			expectError: true,
			errorMsg:    "API key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewAnthropicProvider(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, "anthropic", provider.Name())

				// Check base URL
				expectedURL := tt.config.BaseURL
				if expectedURL == "" {
					expectedURL = AnthropicAPIURL
				}
				assert.Equal(t, expectedURL, provider.baseURL)
				assert.Equal(t, tt.config.APIKey, provider.apiKey)
			}
		})
	}
}

func TestAnthropicProvider_Name(t *testing.T) {
	provider, err := NewAnthropicProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	assert.Equal(t, "anthropic", provider.Name())
}

func TestAnthropicProvider_Models(t *testing.T) {
	provider, err := NewAnthropicProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	models := provider.Models()
	assert.NotNil(t, models)
	assert.Len(t, models, len(supportedModels))

	// Verify all expected models are present
	expectedModels := []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-opus-20240229",
		"claude-3-haiku-20240307",
		"claude-3-5-haiku-20241022",
		"claude-3-sonnet-20240229",
	}

	for _, expected := range expectedModels {
		assert.Contains(t, models, expected)
	}
}

func TestAnthropicProvider_Chat(t *testing.T) {
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
				provider.WithModel("claude-3-5-sonnet-20241022"),
				provider.WithMaxTokens(1024),
			},
			mockResponse: `{
				"id": "msg_123",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "Hello! How can I help you?"}],
				"model": "claude-3-5-sonnet-20241022",
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 10, "output_tokens": 20}
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			validateReq: func(t *testing.T, req *http.Request) {
				assert.Equal(t, "POST", req.Method)
				assert.Equal(t, "/messages", req.URL.Path)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				assert.Equal(t, "test-api-key", req.Header.Get("x-api-key"))
				assert.Equal(t, AnthropicAPIVersion, req.Header.Get("anthropic-version"))

				// Parse request body
				body, err := io.ReadAll(req.Body)
				require.NoError(t, err)

				var reqBody anthropicRequest
				err = json.Unmarshal(body, &reqBody)
				require.NoError(t, err)

				assert.Equal(t, "claude-3-5-sonnet-20241022", reqBody.Model)
				assert.Equal(t, 1024, reqBody.MaxTokens)
				assert.False(t, reqBody.Stream)
				assert.Len(t, reqBody.Messages, 1)
				assert.Equal(t, "user", reqBody.Messages[0].Role)
			},
		},
		{
			name: "with system prompt",
			messages: []provider.Message{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "Hello!"},
			},
			options: []provider.ChatOption{
				provider.WithModel("claude-3-haiku-20240307"),
				provider.WithMaxTokens(512),
			},
			mockResponse: `{
				"id": "msg_456",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "Hi there!"}],
				"model": "claude-3-haiku-20240307",
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 15, "output_tokens": 5}
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			validateReq: func(t *testing.T, req *http.Request) {
				body, err := io.ReadAll(req.Body)
				require.NoError(t, err)

				var reqBody anthropicRequest
				err = json.Unmarshal(body, &reqBody)
				require.NoError(t, err)

				// System message should be in System field, not Messages
				assert.Equal(t, "You are a helpful assistant.", reqBody.System)
				assert.Len(t, reqBody.Messages, 1)
				assert.Equal(t, "user", reqBody.Messages[0].Role)
			},
		},
		{
			name: "with temperature and top_p",
			messages: []provider.Message{
				{Role: "user", Content: "Test"},
			},
			options: []provider.ChatOption{
				provider.WithModel("claude-3-opus-20240229"),
				provider.WithMaxTokens(100),
				provider.WithTemperature(0.7),
				provider.WithTopP(0.9),
			},
			mockResponse: `{
				"id": "msg_789",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "Response"}],
				"model": "claude-3-opus-20240229",
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 5, "output_tokens": 3}
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			validateReq: func(t *testing.T, req *http.Request) {
				body, err := io.ReadAll(req.Body)
				require.NoError(t, err)

				var reqBody anthropicRequest
				err = json.Unmarshal(body, &reqBody)
				require.NoError(t, err)

				assert.NotNil(t, reqBody.Temperature)
				assert.Equal(t, 0.7, *reqBody.Temperature)
				assert.NotNil(t, reqBody.TopP)
				assert.Equal(t, 0.9, *reqBody.TopP)
			},
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
				provider.WithModel("claude-3-5-sonnet-20241022"),
			},
			mockResponse: `{
				"type": "error",
				"error": {
					"type": "authentication_error",
					"message": "Invalid API key"
				}
			}`,
			mockStatusCode: http.StatusUnauthorized,
			expectError:    true,
			errorType:      "authentication",
		},
		{
			name: "rate limit error",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			options: []provider.ChatOption{
				provider.WithModel("claude-3-5-sonnet-20241022"),
			},
			mockResponse: `{
				"type": "error",
				"error": {
					"type": "rate_limit_error",
					"message": "Rate limit exceeded"
				}
			}`,
			mockStatusCode: http.StatusTooManyRequests,
			expectError:    true,
			errorType:      "rate_limit",
		},
		{
			name: "context length error",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			options: []provider.ChatOption{
				provider.WithModel("claude-3-5-sonnet-20241022"),
			},
			mockResponse: `{
				"type": "error",
				"error": {
					"type": "invalid_request_error",
					"message": "Your prompt is too long"
				}
			}`,
			mockStatusCode: http.StatusBadRequest,
			expectError:    true,
			errorType:      "context_length",
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

			// Create provider with custom base URL
			p, err := NewAnthropicProvider(Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
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
				case "context_length":
					var contextErr *provider.ContextLengthError
					assert.ErrorAs(t, err, &contextErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Content)
				assert.NotEmpty(t, resp.Model)
				assert.Greater(t, resp.Usage.PromptTokens, 0)
				assert.Greater(t, resp.Usage.CompletionTokens, 0)
				assert.Equal(t, resp.Usage.PromptTokens+resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
			}
		})
	}
}

func TestAnthropicProvider_Stream(t *testing.T) {
	tests := []struct {
		name         string
		messages     []provider.Message
		options      []provider.StreamOption
		mockEvents   []string
		expectError  bool
		expectedText string
		validateReq  func(t *testing.T, req *http.Request)
	}{
		{
			name: "successful streaming",
			messages: []provider.Message{
				{Role: "user", Content: "Hello!"},
			},
			options: []provider.StreamOption{
				provider.StreamWithModel("claude-3-5-sonnet-20241022"),
				provider.StreamWithMaxTokens(100),
			},
			mockEvents: []string{
				"event: message_start\ndata: {\"type\":\"message_start\"}\n\n",
				"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Hello\"}}\n\n",
				"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\" there\"}}\n\n",
				"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"!\"}}\n\n",
				"event: message_stop\ndata: {}\n\n",
			},
			expectError:  false,
			expectedText: "Hello there!",
			validateReq: func(t *testing.T, req *http.Request) {
				body, err := io.ReadAll(req.Body)
				require.NoError(t, err)

				var reqBody anthropicRequest
				err = json.Unmarshal(body, &reqBody)
				require.NoError(t, err)

				assert.True(t, reqBody.Stream)
			},
		},
		{
			name: "streaming with error event",
			messages: []provider.Message{
				{Role: "user", Content: "Test"},
			},
			options: []provider.StreamOption{
				provider.StreamWithModel("claude-3-haiku-20240307"),
			},
			mockEvents: []string{
				"event: message_start\ndata: {\"type\":\"message_start\"}\n\n",
				"event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"rate_limit_error\",\"message\":\"Rate limit\"}}\n\n",
			},
			expectError: true,
		},
		{
			name: "context cancellation",
			messages: []provider.Message{
				{Role: "user", Content: "Test"},
			},
			options: []provider.StreamOption{
				provider.StreamWithModel("claude-3-5-sonnet-20241022"),
			},
			mockEvents: []string{
				"event: message_start\ndata: {\"type\":\"message_start\"}\n\n",
				// Delay to allow cancellation
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.validateReq != nil {
					tt.validateReq(t, r)
				}

				w.Header().Set("Content-Type", "text/event-stream")
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
			p, err := NewAnthropicProvider(Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			require.NoError(t, err)

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Special handling for cancellation test
			if tt.name == "context cancellation" {
				ctx, cancel = context.WithCancel(context.Background())
				// Cancel after a short delay
				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()
			}

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

func TestAnthropicProvider_ConvertMessages(t *testing.T) {
	p, err := NewAnthropicProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	tests := []struct {
		name           string
		messages       []provider.Message
		systemPrompt   string
		expectedSystem string
		expectedCount  int
	}{
		{
			name: "user and assistant messages",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there"},
				{Role: "user", Content: "How are you?"},
			},
			systemPrompt:   "",
			expectedSystem: "",
			expectedCount:  3,
		},
		{
			name: "extract system message",
			messages: []provider.Message{
				{Role: "system", Content: "You are helpful."},
				{Role: "user", Content: "Hello"},
			},
			systemPrompt:   "",
			expectedSystem: "You are helpful.",
			expectedCount:  1,
		},
		{
			name: "multiple system messages",
			messages: []provider.Message{
				{Role: "system", Content: "First instruction."},
				{Role: "system", Content: "Second instruction."},
				{Role: "user", Content: "Hello"},
			},
			systemPrompt:   "",
			expectedSystem: "First instruction.\n\nSecond instruction.",
			expectedCount:  1,
		},
		{
			name: "combine extracted and provided system prompts",
			messages: []provider.Message{
				{Role: "system", Content: "From messages."},
				{Role: "user", Content: "Hello"},
			},
			systemPrompt:   "From options.",
			expectedSystem: "From messages.\n\nFrom options.",
			expectedCount:  1,
		},
		{
			name: "only provided system prompt",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			systemPrompt:   "Custom system prompt.",
			expectedSystem: "Custom system prompt.",
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiMessages, systemPrompt := p.convertMessages(tt.messages, tt.systemPrompt)

			assert.Equal(t, tt.expectedSystem, systemPrompt)
			assert.Len(t, apiMessages, tt.expectedCount)

			// Verify no system messages in apiMessages
			for _, msg := range apiMessages {
				assert.NotEqual(t, "system", msg.Role)
				assert.Len(t, msg.Content, 1)
				assert.Equal(t, "text", msg.Content[0].Type)
			}
		})
	}
}

func TestAnthropicProvider_ConvertAPIError(t *testing.T) {
	p, err := NewAnthropicProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	tests := []struct {
		name      string
		apiErr    *anthropicError
		errorType string
	}{
		{
			name: "authentication error",
			apiErr: &anthropicError{
				Type: "error",
				Error: struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				}{
					Type:    "authentication_error",
					Message: "Invalid API key",
				},
			},
			errorType: "authentication",
		},
		{
			name: "permission error",
			apiErr: &anthropicError{
				Type: "error",
				Error: struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				}{
					Type:    "permission_error",
					Message: "No access",
				},
			},
			errorType: "authentication",
		},
		{
			name: "rate limit error",
			apiErr: &anthropicError{
				Type: "error",
				Error: struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				}{
					Type:    "rate_limit_error",
					Message: "Too many requests",
				},
			},
			errorType: "rate_limit",
		},
		{
			name: "not found error - model deprecated/retired",
			apiErr: &anthropicError{
				Type: "error",
				Error: struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				}{
					Type:    "not_found_error",
					Message: "model: claude-3-5-sonnet-20241022",
				},
			},
			errorType: "invalid_model",
		},
		{
			name: "context length error - prompt too long",
			apiErr: &anthropicError{
				Type: "error",
				Error: struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				}{
					Type:    "invalid_request_error",
					Message: "Your prompt is too long",
				},
			},
			errorType: "context_length",
		},
		{
			name: "context length error - max tokens",
			apiErr: &anthropicError{
				Type: "error",
				Error: struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				}{
					Type:    "invalid_request_error",
					Message: "max_tokens exceeds limit",
				},
			},
			errorType: "context_length",
		},
		{
			name: "generic invalid request",
			apiErr: &anthropicError{
				Type: "error",
				Error: struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				}{
					Type:    "invalid_request_error",
					Message: "Missing required field",
				},
			},
			errorType: "provider",
		},
		{
			name: "unknown error type",
			apiErr: &anthropicError{
				Type: "error",
				Error: struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				}{
					Type:    "unknown_error",
					Message: "Something went wrong",
				},
			},
			errorType: "provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.convertAPIError(tt.apiErr, "claude-3-5-sonnet-20241022")
			assert.Error(t, err)

			switch tt.errorType {
			case "authentication":
				var authErr *provider.AuthenticationError
				assert.ErrorAs(t, err, &authErr)
			case "rate_limit":
				var rateLimitErr *provider.RateLimitError
				assert.ErrorAs(t, err, &rateLimitErr)
			case "invalid_model":
				var invalidModelErr *provider.InvalidModelError
				assert.ErrorAs(t, err, &invalidModelErr)
			case "context_length":
				var contextErr *provider.ContextLengthError
				assert.ErrorAs(t, err, &contextErr)
			case "provider":
				var providerErr *provider.ProviderError
				assert.ErrorAs(t, err, &providerErr)
			}
		})
	}
}

func TestAnthropicProvider_Close(t *testing.T) {
	p, err := NewAnthropicProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	err = p.Close()
	assert.NoError(t, err)
}

func TestAnthropicProvider_ParseResponse(t *testing.T) {
	p, err := NewAnthropicProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	tests := []struct {
		name            string
		responseJSON    string
		expectedContent string
		expectedModel   string
		expectError     bool
	}{
		{
			name: "single text block",
			responseJSON: `{
				"id": "msg_1",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "Hello!"}],
				"model": "claude-3-5-sonnet-20241022",
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 10, "output_tokens": 5}
			}`,
			expectedContent: "Hello!",
			expectedModel:   "claude-3-5-sonnet-20241022",
			expectError:     false,
		},
		{
			name: "multiple text blocks",
			responseJSON: `{
				"id": "msg_2",
				"type": "message",
				"role": "assistant",
				"content": [
					{"type": "text", "text": "First part."},
					{"type": "text", "text": "Second part."}
				],
				"model": "claude-3-haiku-20240307",
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 20, "output_tokens": 10}
			}`,
			expectedContent: "First part.\nSecond part.",
			expectedModel:   "claude-3-haiku-20240307",
			expectError:     false,
		},
		{
			name:         "invalid JSON",
			responseJSON: `{invalid json`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := p.parseResponse([]byte(tt.responseJSON), "test-model")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedContent, resp.Content)
				assert.Equal(t, tt.expectedModel, resp.Model)
				assert.Greater(t, resp.Usage.TotalTokens, 0)
			}
		})
	}
}

func TestAnthropicProvider_BuildRequest(t *testing.T) {
	p, err := NewAnthropicProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	options := &provider.ChatOptions{
		Model:         "claude-3-5-sonnet-20241022",
		MaxTokens:     1024,
		Temperature:   0.7,
		TopP:          0.9,
		StopSequences: []string{"STOP"},
		SystemPrompt:  "Be helpful",
	}

	req, err := p.buildRequest(context.Background(), messages, options, false)
	require.NoError(t, err)

	assert.Equal(t, "POST", req.Method)
	assert.True(t, strings.HasSuffix(req.URL.String(), "/messages"))
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "test-key", req.Header.Get("x-api-key"))
	assert.Equal(t, AnthropicAPIVersion, req.Header.Get("anthropic-version"))

	// Parse body
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)

	var reqBody anthropicRequest
	err = json.Unmarshal(body, &reqBody)
	require.NoError(t, err)

	assert.Equal(t, "claude-3-5-sonnet-20241022", reqBody.Model)
	assert.Equal(t, 1024, reqBody.MaxTokens)
	assert.False(t, reqBody.Stream)
	assert.Equal(t, "Be helpful", reqBody.System)
	assert.NotNil(t, reqBody.Temperature)
	assert.Equal(t, 0.7, *reqBody.Temperature)
	assert.NotNil(t, reqBody.TopP)
	assert.Equal(t, 0.9, *reqBody.TopP)
	assert.Equal(t, []string{"STOP"}, reqBody.StopSequences)
}
