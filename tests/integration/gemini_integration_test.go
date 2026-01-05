package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/AINative-studio/ainative-code/internal/provider/gemini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGeminiProviderIntegration tests the complete workflow with Gemini provider
func TestGeminiProviderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create a comprehensive mock server that simulates Gemini API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify API key
		if !r.URL.Query().Has("key") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"code":    401,
					"message": "API key not valid",
					"status":  "UNAUTHENTICATED",
				},
			})
			return
		}

		// Handle different endpoints
		if r.URL.Path == "/models/gemini-pro:generateContent" {
			// Non-streaming response
			response := map[string]interface{}{
				"candidates": []map[string]interface{}{
					{
						"content": map[string]interface{}{
							"parts": []map[string]interface{}{
								{"text": "Hello! I'm Gemini, ready to help."},
							},
						},
						"finishReason": "STOP",
						"index":        0,
					},
				},
				"usageMetadata": map[string]interface{}{
					"promptTokenCount":     10,
					"candidatesTokenCount": 15,
					"totalTokenCount":      25,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		} else if r.URL.Path == "/models/gemini-pro:streamGenerateContent" {
			// Streaming response
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)

			chunks := []string{
				`data: {"candidates":[{"content":{"parts":[{"text":"Hello"}]}}]}`,
				`data: {"candidates":[{"content":{"parts":[{"text":" from"}]}}]}`,
				`data: {"candidates":[{"content":{"parts":[{"text":" Gemini!"}]},"finishReason":"STOP"}]}`,
			}

			for _, chunk := range chunks {
				w.Write([]byte(chunk + "\n\n"))
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				time.Sleep(10 * time.Millisecond)
			}

		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create provider
	p, err := gemini.NewGeminiProvider(gemini.Config{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)
	defer p.Close()

	t.Run("chat request", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		messages := []provider.Message{
			{Role: "user", Content: "Hello"},
		}

		resp, err := p.Chat(ctx, messages,
			provider.WithModel("gemini-pro"),
			provider.WithMaxTokens(100),
			provider.WithTemperature(0.7),
		)

		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
		assert.Contains(t, resp.Content, "Gemini")
		assert.Equal(t, "gemini-pro", resp.Model)
		assert.Equal(t, 10, resp.Usage.PromptTokens)
		assert.Equal(t, 15, resp.Usage.CompletionTokens)
		assert.Equal(t, 25, resp.Usage.TotalTokens)
	})

	t.Run("streaming request", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		messages := []provider.Message{
			{Role: "user", Content: "Hello"},
		}

		eventChan, err := p.Stream(ctx, messages, provider.StreamWithModel("gemini-pro"))
		require.NoError(t, err)
		require.NotNil(t, eventChan)

		var content string
		var gotStart, gotEnd bool

		for event := range eventChan {
			switch event.Type {
			case provider.EventTypeContentStart:
				gotStart = true
			case provider.EventTypeContentDelta:
				content += event.Content
			case provider.EventTypeContentEnd:
				gotEnd = true
				assert.True(t, event.Done)
			case provider.EventTypeError:
				t.Fatalf("unexpected error: %v", event.Error)
			}
		}

		assert.True(t, gotStart, "should receive content start event")
		assert.True(t, gotEnd, "should receive content end event")
		assert.Equal(t, "Hello from Gemini!", content)
	})

	t.Run("multi-turn conversation", func(t *testing.T) {
		ctx := context.Background()

		messages := []provider.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi!"},
			{Role: "user", Content: "How are you?"},
		}

		resp, err := p.Chat(ctx, messages, provider.WithModel("gemini-pro"))
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})

	t.Run("with system prompt", func(t *testing.T) {
		ctx := context.Background()

		messages := []provider.Message{
			{Role: "user", Content: "Hello"},
		}

		resp, err := p.Chat(ctx, messages,
			provider.WithModel("gemini-pro"),
			provider.WithSystemPrompt("You are a helpful assistant"),
		)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})
}

// TestGeminiErrorHandlingIntegration tests error scenarios
func TestGeminiErrorHandlingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    string
	}{
		{
			name: "invalid API key",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]interface{}{
						"code":    401,
						"message": "API key not valid",
						"status":  "UNAUTHENTICATED",
					},
				})
			},
			expectError: "authentication",
		},
		{
			name: "rate limit",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]interface{}{
						"code":    429,
						"message": "Quota exceeded",
						"status":  "RESOURCE_EXHAUSTED",
					},
				})
			},
			expectError: "rate limit",
		},
		{
			name: "content too long",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]interface{}{
						"code":    400,
						"message": "Request content exceeds token limit",
						"status":  "INVALID_ARGUMENT",
					},
				})
			},
			expectError: "context length",
		},
		{
			name: "safety block",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"promptFeedback": map[string]interface{}{
						"blockReason": "SAFETY",
					},
				})
			},
			expectError: "blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			p, err := gemini.NewGeminiProvider(gemini.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			require.NoError(t, err)

			ctx := context.Background()
			messages := []provider.Message{
				{Role: "user", Content: "Hello"},
			}

			_, err = p.Chat(ctx, messages, provider.WithModel("gemini-pro"))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

// TestGeminiMultiModalSupport tests multi-modal capabilities
func TestGeminiMultiModalSupport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{"text": "I can see an image in your request."},
						},
					},
					"finishReason": "STOP",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	p, err := gemini.NewGeminiProvider(gemini.Config{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "What do you see in this image?"},
	}

	// Note: Multi-modal content would be added via metadata in a real scenario
	resp, err := p.Chat(ctx, messages, provider.WithModel("gemini-pro-vision"))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
}

// TestGeminiContextCancellation tests context cancellation handling
func TestGeminiContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	p, err := gemini.NewGeminiProvider(gemini.Config{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	_, err = p.Chat(ctx, messages, provider.WithModel("gemini-pro"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

// TestGeminiProviderCleanup tests proper resource cleanup
func TestGeminiProviderCleanup(t *testing.T) {
	p, err := gemini.NewGeminiProvider(gemini.Config{
		APIKey: "test-api-key",
	})
	require.NoError(t, err)

	// Provider should clean up without errors
	err = p.Close()
	assert.NoError(t, err)
}
