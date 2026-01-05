package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/AINative-studio/ainative-code/internal/provider/bedrock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBedrockIntegration_ChatWorkflow(t *testing.T) {
	// Create mock Bedrock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/model/")
		assert.Contains(t, r.URL.Path, "/invoke")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify AWS signature headers
		assert.NotEmpty(t, r.Header.Get("Authorization"))
		assert.NotEmpty(t, r.Header.Get("X-Amz-Date"))
		assert.NotEmpty(t, r.Header.Get("X-Amz-Content-Sha256"))

		// Return successful response
		response := map[string]interface{}{
			"output": map[string]interface{}{
				"message": map[string]interface{}{
					"role": "assistant",
					"content": []map[string]interface{}{
						{"text": "Hello! I'm Claude, an AI assistant. How can I help you today?"},
					},
				},
			},
			"usage": map[string]interface{}{
				"inputTokens":  25,
				"outputTokens": 18,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create provider
	p, err := bedrock.NewBedrockProvider(bedrock.Config{
		Region:    "us-east-1",
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Endpoint:  server.URL,
	})
	require.NoError(t, err)
	defer p.Close()

	// Test chat
	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "Hello, Claude! Introduce yourself."},
	}

	resp, err := p.Chat(ctx, messages,
		provider.WithModel("anthropic.claude-3-5-sonnet-20241022-v2:0"),
		provider.WithMaxTokens(100),
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
	assert.Contains(t, resp.Content, "Claude")
	assert.Equal(t, "anthropic.claude-3-5-sonnet-20241022-v2:0", resp.Model)
	assert.Greater(t, resp.Usage.PromptTokens, 0)
	assert.Greater(t, resp.Usage.CompletionTokens, 0)
	assert.Equal(t, resp.Usage.PromptTokens+resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
}

func TestBedrockIntegration_StreamingWorkflow(t *testing.T) {
	// Create mock Bedrock streaming server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/invoke-with-response-stream")
		assert.Equal(t, "application/vnd.amazon.eventstream", r.Header.Get("Accept"))

		// Verify AWS signature
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		// Return streaming response
		w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		// Send streaming events
		events := []string{
			`{"messageStart":{"role":"assistant"}}`,
			`{"contentBlockDelta":{"delta":{"text":"Hello"},"contentBlockIndex":0}}`,
			`{"contentBlockDelta":{"delta":{"text":" from"},"contentBlockIndex":0}}`,
			`{"contentBlockDelta":{"delta":{"text":" Bedrock"},"contentBlockIndex":0}}`,
			`{"contentBlockDelta":{"delta":{"text":"!"},"contentBlockIndex":0}}`,
			`{"messageStop":{}}`,
		}

		for _, event := range events {
			w.Write([]byte(event + "\n"))
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	// Create provider
	p, err := bedrock.NewBedrockProvider(bedrock.Config{
		Region:    "us-east-1",
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Endpoint:  server.URL,
	})
	require.NoError(t, err)
	defer p.Close()

	// Test streaming
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messages := []provider.Message{
		{Role: "user", Content: "Say hello"},
	}

	eventChan, err := p.Stream(ctx, messages,
		provider.StreamWithModel("anthropic.claude-3-haiku-20240307-v1:0"),
		provider.StreamWithMaxTokens(50),
	)
	require.NoError(t, err)

	// Collect events
	var events []provider.Event
	var fullText string

	for event := range eventChan {
		events = append(events, event)

		if event.Type == provider.EventTypeContentDelta {
			fullText += event.Content
		}

		if event.Type == provider.EventTypeError {
			t.Fatalf("Unexpected error: %v", event.Error)
		}
	}

	// Verify events
	assert.Greater(t, len(events), 0)
	assert.Equal(t, provider.EventTypeContentStart, events[0].Type)
	assert.Equal(t, provider.EventTypeContentEnd, events[len(events)-1].Type)
	assert.True(t, events[len(events)-1].Done)
	assert.Equal(t, "Hello from Bedrock!", fullText)
}

func TestBedrockIntegration_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedError  string
	}{
		{
			name:       "authentication error",
			statusCode: http.StatusForbidden,
			responseBody: `{"message":"The security token included in the request is invalid."}`,
			expectedError: "authentication",
		},
		{
			name:       "throttling error",
			statusCode: http.StatusTooManyRequests,
			responseBody: `{"message":"Rate exceeded"}`,
			expectedError: "rate limit", // The BaseProvider retry logic wraps this
		},
		{
			name:       "validation error",
			statusCode: http.StatusBadRequest,
			responseBody: `{"message":"Validation error: messages is required"}`,
			expectedError: "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create provider
			p, err := bedrock.NewBedrockProvider(bedrock.Config{
				Region:    "us-east-1",
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
				Endpoint:  server.URL,
			})
			require.NoError(t, err)
			defer p.Close()

			// Execute request
			ctx := context.Background()
			messages := []provider.Message{
				{Role: "user", Content: "Test"},
			}

			_, err = p.Chat(ctx, messages,
				provider.WithModel("anthropic.claude-3-5-sonnet-20241022-v2:0"),
			)

			// Verify error
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestBedrockIntegration_MultipleModels(t *testing.T) {
	models := []string{
		"anthropic.claude-3-5-sonnet-20241022-v2:0",
		"anthropic.claude-3-opus-20240229-v1:0",
		"anthropic.claude-3-haiku-20240307-v1:0",
		"anthropic.claude-v2",
		"anthropic.claude-instant-v1",
	}

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify model in URL
				assert.Contains(t, r.URL.Path, model)

				response := map[string]interface{}{
					"output": map[string]interface{}{
						"message": map[string]interface{}{
							"role": "assistant",
							"content": []map[string]interface{}{
								{"text": "Response from " + model},
							},
						},
					},
					"usage": map[string]interface{}{
						"inputTokens":  10,
						"outputTokens": 5,
					},
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			// Create provider
			p, err := bedrock.NewBedrockProvider(bedrock.Config{
				Region:    "us-east-1",
				AccessKey: "test-key",
				SecretKey: "test-secret",
				Endpoint:  server.URL,
			})
			require.NoError(t, err)
			defer p.Close()

			// Test
			ctx := context.Background()
			messages := []provider.Message{
				{Role: "user", Content: "Test"},
			}

			resp, err := p.Chat(ctx, messages, provider.WithModel(model))
			assert.NoError(t, err)
			assert.NotEmpty(t, resp.Content)
		})
	}
}

func TestBedrockIntegration_SystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and verify request body
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)

		// Verify system prompt is present
		assert.Contains(t, reqBody, "system")
		system := reqBody["system"].([]interface{})
		assert.Len(t, system, 1)
		systemText := system[0].(map[string]interface{})["text"].(string)
		assert.Equal(t, "You are a helpful assistant.", systemText)

		response := map[string]interface{}{
			"output": map[string]interface{}{
				"message": map[string]interface{}{
					"role": "assistant",
					"content": []map[string]interface{}{
						{"text": "I'll be helpful!"},
					},
				},
			},
			"usage": map[string]interface{}{
				"inputTokens":  15,
				"outputTokens": 5,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	p, err := bedrock.NewBedrockProvider(bedrock.Config{
		Region:    "us-east-1",
		AccessKey: "test-key",
		SecretKey: "test-secret",
		Endpoint:  server.URL,
	})
	require.NoError(t, err)
	defer p.Close()

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello"},
	}

	resp, err := p.Chat(ctx, messages,
		provider.WithModel("anthropic.claude-3-5-sonnet-20241022-v2:0"),
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
}
