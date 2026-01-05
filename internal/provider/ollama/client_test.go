package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOllamaProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: Config{
				BaseURL: "http://localhost:11434",
				Model:   "llama2",
			},
			expectError: false,
		},
		{
			name: "missing model",
			config: Config{
				BaseURL: "http://localhost:11434",
			},
			expectError: true,
			errorMsg:    "model name is required",
		},
		{
			name: "config with defaults",
			config: Config{
				Model: "llama3",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOllamaProvider(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, "ollama", provider.Name())
			}
		})
	}
}

func TestOllamaProvider_Name(t *testing.T) {
	config := Config{
		Model: "llama2",
	}
	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	assert.Equal(t, "ollama", provider.Name())
}

func TestOllamaProvider_Models(t *testing.T) {
	config := Config{
		Model: "llama2",
	}
	prov, err := NewOllamaProvider(config)
	require.NoError(t, err)

	models := prov.Models()
	assert.NotEmpty(t, models)
	assert.Contains(t, models, "llama2")
	assert.Contains(t, models, "llama3")
	assert.Contains(t, models, "codellama")
}

func TestOllamaProvider_Chat(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			// Parse request
			var req ollamaRequest
			json.NewDecoder(r.Body).Decode(&req)

			// Return mock response
			resp := ollamaResponse{
				Model:     req.Model,
				CreatedAt: time.Now(),
				Message: ollamaMessage{
					Role:    "assistant",
					Content: "This is a test response from " + req.Model,
				},
				Done:            true,
				PromptEvalCount: 10,
				EvalCount:       20,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	prov, err := NewOllamaProvider(config)
	require.NoError(t, err)

	t.Run("successful chat request", func(t *testing.T) {
		messages := []provider.Message{
			{Role: "user", Content: "Hello, how are you?"},
		}

		ctx := context.Background()
		resp, err := prov.Chat(ctx, messages, provider.WithModel("llama2"))

		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
		assert.Contains(t, resp.Content, "test response")
		assert.Equal(t, "llama2", resp.Model)
		assert.Equal(t, 10, resp.Usage.PromptTokens)
		assert.Equal(t, 20, resp.Usage.CompletionTokens)
		assert.Equal(t, 30, resp.Usage.TotalTokens)
	})

	t.Run("chat with system prompt", func(t *testing.T) {
		messages := []provider.Message{
			{Role: "user", Content: "Hello"},
		}

		ctx := context.Background()
		resp, err := prov.Chat(ctx, messages,
			provider.WithModel("llama2"),
			provider.WithSystemPrompt("You are a helpful assistant"),
		)

		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})

	t.Run("chat with options", func(t *testing.T) {
		messages := []provider.Message{
			{Role: "user", Content: "Write code"},
		}

		ctx := context.Background()
		resp, err := prov.Chat(ctx, messages,
			provider.WithModel("llama2"),
			provider.WithTemperature(0.7),
			provider.WithMaxTokens(2048),
		)

		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})

	t.Run("chat with cancelled context", func(t *testing.T) {
		messages := []provider.Message{
			{Role: "user", Content: "Hello"},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := prov.Chat(ctx, messages, provider.WithModel("llama2"))
		assert.Error(t, err)
	})
}

func TestOllamaProvider_Stream(t *testing.T) {
	// Create mock streaming server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.Header().Set("Content-Type", "application/x-ndjson")

			// Simulate streaming chunks
			chunks := []ollamaStreamResponse{
				{
					Model:     "llama2",
					CreatedAt: time.Now(),
					Message:   ollamaMessage{Role: "assistant", Content: "Hello"},
					Done:      false,
				},
				{
					Model:     "llama2",
					CreatedAt: time.Now(),
					Message:   ollamaMessage{Role: "assistant", Content: " there"},
					Done:      false,
				},
				{
					Model:           "llama2",
					CreatedAt:       time.Now(),
					Message:         ollamaMessage{Role: "assistant", Content: "!"},
					Done:            true,
					PromptEvalCount: 5,
					EvalCount:       3,
				},
			}

			encoder := json.NewEncoder(w)
			for _, chunk := range chunks {
				encoder.Encode(chunk)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		}
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	prov, err := NewOllamaProvider(config)
	require.NoError(t, err)

	t.Run("successful streaming", func(t *testing.T) {
		messages := []provider.Message{
			{Role: "user", Content: "Hello"},
		}

		ctx := context.Background()
		eventChan, err := prov.Stream(ctx, messages, provider.StreamWithModel("llama2"))

		require.NoError(t, err)
		require.NotNil(t, eventChan)

		var events []provider.Event
		for event := range eventChan {
			events = append(events, event)
		}

		assert.NotEmpty(t, events)

		// Check for start event
		assert.Equal(t, provider.EventTypeContentStart, events[0].Type)

		// Check for content deltas
		var content string
		for _, event := range events {
			if event.Type == provider.EventTypeContentDelta {
				content += event.Content
			}
		}
		assert.Equal(t, "Hello there!", content)

		// Check for end event
		lastEvent := events[len(events)-1]
		assert.Equal(t, provider.EventTypeContentEnd, lastEvent.Type)
		assert.True(t, lastEvent.Done)
	})

	t.Run("stream with cancelled context", func(t *testing.T) {
		messages := []provider.Message{
			{Role: "user", Content: "Hello"},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		eventChan, err := prov.Stream(ctx, messages, provider.StreamWithModel("llama2"))
		require.NoError(t, err)

		// Wait for cancellation
		time.Sleep(20 * time.Millisecond)

		// Drain channel
		for range eventChan {
		}
	})
}

func TestOllamaProvider_ChatError(t *testing.T) {
	// Create error server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "model 'unknown' not found",
		})
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Model:   "unknown",
	}

	prov, err := NewOllamaProvider(config)
	require.NoError(t, err)

	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	ctx := context.Background()
	_, err = prov.Chat(ctx, messages, provider.WithModel("unknown"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestOllamaProvider_Close(t *testing.T) {
	config := Config{
		Model: "llama2",
	}

	prov, err := NewOllamaProvider(config)
	require.NoError(t, err)

	err = prov.Close()
	assert.NoError(t, err)
}

func TestOllamaProvider_ConversationContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			var req ollamaRequest
			json.NewDecoder(r.Body).Decode(&req)

			// Verify multiple messages for conversation
			assert.GreaterOrEqual(t, len(req.Messages), 2)

			resp := ollamaResponse{
				Model:   req.Model,
				Message: ollamaMessage{Role: "assistant", Content: "Response to conversation"},
				Done:    true,
			}

			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	prov, err := NewOllamaProvider(config)
	require.NoError(t, err)

	messages := []provider.Message{
		{Role: "user", Content: "First message"},
		{Role: "assistant", Content: "First response"},
		{Role: "user", Content: "Second message"},
	}

	ctx := context.Background()
	resp, err := prov.Chat(ctx, messages, provider.WithModel("llama2"))

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
}

func TestOllamaProvider_HelperMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			modelsResp := ollamaModelsResponse{
				Models: []ollamaModelInfo{
					{Name: "llama2", Size: 7365960704},
					{Name: "mistral", Size: 7000000000},
				},
			}
			json.NewEncoder(w).Encode(modelsResp)
		} else if r.URL.Path == "/api/chat" {
			resp := ollamaResponse{
				Model:   "llama2",
				Message: ollamaMessage{Role: "assistant", Content: "test"},
				Done:    true,
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	prov, err := NewOllamaProvider(config)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("health check", func(t *testing.T) {
		err := prov.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("list available models", func(t *testing.T) {
		models, err := prov.ListAvailableModels(ctx)
		require.NoError(t, err)
		assert.Len(t, models, 2)
	})

	t.Run("get model info", func(t *testing.T) {
		info, err := prov.GetModelInfo(ctx, "llama2")
		require.NoError(t, err)
		assert.Equal(t, "llama2", info.Name)
	})

	t.Run("is model available", func(t *testing.T) {
		available, err := prov.IsModelAvailable(ctx, "llama2")
		require.NoError(t, err)
		assert.True(t, available)

		available, err = prov.IsModelAvailable(ctx, "unknown")
		require.NoError(t, err)
		assert.False(t, available)
	})
}

func TestNewOllamaProviderWithLogger(t *testing.T) {
	config := Config{
		Model: "llama2",
	}

	// Mock logger can be nil for this test
	prov, err := NewOllamaProviderWithLogger(config, nil)
	require.NoError(t, err)
	assert.NotNil(t, prov)
}

func TestNewOllamaProviderForModel(t *testing.T) {
	prov, err := NewOllamaProviderForModel("llama3")
	require.NoError(t, err)
	assert.NotNil(t, prov)
	assert.Equal(t, "ollama", prov.Name())
}

func TestOllamaProvider_DifferentModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaRequest
		json.NewDecoder(r.Body).Decode(&req)

		resp := ollamaResponse{
			Model:   req.Model,
			Message: ollamaMessage{Role: "assistant", Content: "Response from " + req.Model},
			Done:    true,
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	models := []string{"llama2", "llama3", "codellama", "mistral"}

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			config := Config{
				BaseURL: server.URL,
				Model:   model,
			}

			prov, err := NewOllamaProvider(config)
			require.NoError(t, err)

			messages := []provider.Message{
				{Role: "user", Content: "Test message"},
			}

			ctx := context.Background()
			resp, err := prov.Chat(ctx, messages, provider.WithModel(model))

			require.NoError(t, err)
			assert.Contains(t, resp.Content, model)
			assert.Equal(t, model, resp.Model)
		})
	}
}
