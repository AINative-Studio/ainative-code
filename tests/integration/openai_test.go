package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/AINative-studio/ainative-code/internal/provider/openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAIProvider_Integration tests the complete workflow
func TestOpenAIProvider_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create a mock server that simulates OpenAI API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request format
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer")

		// Return mock response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "chatcmpl-test",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-4",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "This is a test response from the mock OpenAI API."
				},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 15,
				"completion_tokens": 25,
				"total_tokens": 40
			}
		}`))
	}))
	defer server.Close()

	// Create provider
	p, err := openai.NewOpenAIProvider(openai.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)
	defer p.Close()

	// Test chat
	ctx := context.Background()
	messages := []provider.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello, how are you?"},
	}

	resp, err := p.Chat(ctx, messages,
		provider.WithModel("gpt-4"),
		provider.WithMaxTokens(100),
		provider.WithTemperature(0.7),
	)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
	assert.Equal(t, "gpt-4", resp.Model)
	assert.Greater(t, resp.Usage.TotalTokens, 0)
}

// TestOpenAIProvider_StreamingIntegration tests streaming workflow
func TestOpenAIProvider_StreamingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create mock streaming server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		// Send SSE events
		streamData := []string{
			`data: {"id":"1","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}`,
			``,
			`data: {"id":"1","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}`,
			``,
			`data: {"id":"1","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":" there"},"finish_reason":null}]}`,
			``,
			`data: {"id":"1","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":null}]}`,
			``,
			`data: {"id":"1","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
			``,
			`data: [DONE]`,
			``,
		}

		for _, data := range streamData {
			w.Write([]byte(data + "\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	// Create provider
	p, err := openai.NewOpenAIProvider(openai.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)
	defer p.Close()

	// Test streaming
	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "Say hello"},
	}

	eventChan, err := p.Stream(ctx, messages, provider.StreamWithModel("gpt-3.5-turbo"))
	require.NoError(t, err)

	// Collect events
	var events []provider.Event
	var fullContent string

	for event := range eventChan {
		events = append(events, event)

		if event.Type == provider.EventTypeContentDelta {
			fullContent += event.Content
		}

		if event.Error != nil {
			t.Logf("Event error: %v", event.Error)
		}
	}

	// Verify results
	assert.NotEmpty(t, events)
	assert.Equal(t, "Hello there!", fullContent)

	// Check for start and end events
	assert.Equal(t, provider.EventTypeContentStart, events[0].Type)
	lastEvent := events[len(events)-1]
	assert.Equal(t, provider.EventTypeContentEnd, lastEvent.Type)
	assert.True(t, lastEvent.Done)
}

// TestOpenAIProvider_ErrorHandling tests error scenarios
func TestOpenAIProvider_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	tests := []struct {
		name           string
		statusCode     int
		response       string
		expectedErrMsg string
	}{
		{
			name:       "authentication error",
			statusCode: http.StatusUnauthorized,
			response: `{
				"error": {
					"message": "Invalid API key provided",
					"type": "invalid_request_error"
				}
			}`,
			expectedErrMsg: "authentication",
		},
		{
			name:       "rate limit error",
			statusCode: http.StatusTooManyRequests,
			response: `{
				"error": {
					"message": "Rate limit exceeded",
					"type": "rate_limit_error"
				}
			}`,
			expectedErrMsg: "rate limit",
		},
		{
			name:       "context length error",
			statusCode: http.StatusBadRequest,
			response: `{
				"error": {
					"message": "This model's maximum context length is 8192 tokens",
					"type": "invalid_request_error"
				}
			}`,
			expectedErrMsg: "context length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			p, err := openai.NewOpenAIProvider(openai.Config{
				APIKey:  "test-key",
				BaseURL: server.URL,
			})
			require.NoError(t, err)
			defer p.Close()

			ctx := context.Background()
			messages := []provider.Message{{Role: "user", Content: "test"}}

			_, err = p.Chat(ctx, messages, provider.WithModel("gpt-4"))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErrMsg)
		})
	}
}

// TestOpenAIProvider_MultiProvider tests using multiple providers
func TestOpenAIProvider_MultiProvider(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create registry
	registry := provider.NewRegistry()

	// Create OpenAI mock server
	openaiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "test",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-4",
			"choices": [{
				"index": 0,
				"message": {"role": "assistant", "content": "OpenAI response"},
				"finish_reason": "stop"
			}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
		}`))
	}))
	defer openaiServer.Close()

	// Create and register OpenAI provider
	openaiProvider, err := openai.NewOpenAIProvider(openai.Config{
		APIKey:  "test-key",
		BaseURL: openaiServer.URL,
	})
	require.NoError(t, err)

	err = registry.Register("openai", openaiProvider)
	require.NoError(t, err)

	// Verify provider is registered
	assert.True(t, registry.Has("openai"))
	assert.Equal(t, 1, registry.Count())

	// Get and use provider
	p, err := registry.Get("openai")
	require.NoError(t, err)
	assert.Equal(t, "openai", p.Name())

	ctx := context.Background()
	messages := []provider.Message{{Role: "user", Content: "test"}}

	resp, err := p.Chat(ctx, messages, provider.WithModel("gpt-4"))
	require.NoError(t, err)
	assert.Contains(t, resp.Content, "OpenAI response")

	// Cleanup
	err = registry.Close()
	assert.NoError(t, err)
}

// TestOpenAIProvider_Timeout tests request timeout handling
func TestOpenAIProvider_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices":[{"message":{"content":"test"}}],"usage":{}}`))
	}))
	defer server.Close()

	// Create provider with custom HTTP client (short timeout)
	httpClient := &http.Client{
		Timeout: 50 * time.Millisecond,
	}

	p, err := openai.NewOpenAIProvider(openai.Config{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		HTTPClient: httpClient,
	})
	require.NoError(t, err)
	defer p.Close()

	ctx := context.Background()
	messages := []provider.Message{{Role: "user", Content: "test"}}

	_, err = p.Chat(ctx, messages, provider.WithModel("gpt-4"))
	assert.Error(t, err) // Should timeout
}
