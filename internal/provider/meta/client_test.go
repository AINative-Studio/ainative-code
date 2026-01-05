package meta

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

func TestNewMetaProvider(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				APIKey:  "test-key",
				BaseURL: DefaultBaseURL,
				Model:   ModelLlama4Maverick,
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: &Config{
				BaseURL: DefaultBaseURL,
				Model:   ModelLlama4Maverick,
			},
			wantErr: true,
		},
		{
			name: "invalid model",
			config: &Config{
				APIKey:  "test-key",
				BaseURL: DefaultBaseURL,
				Model:   "invalid-model",
			},
			wantErr: true,
		},
		{
			name:    "nil config uses defaults",
			config:  nil,
			wantErr: true, // Will fail validation because no API key
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewMetaProvider(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestMetaProvider_Name(t *testing.T) {
	p := &MetaProvider{}
	assert.Equal(t, "meta", p.Name())
}

func TestMetaProvider_Models(t *testing.T) {
	p := &MetaProvider{}
	models := p.Models()

	assert.Len(t, models, 4)
	assert.Contains(t, models, ModelLlama4Maverick)
	assert.Contains(t, models, ModelLlama4Scout)
	assert.Contains(t, models, ModelLlama33_70B)
	assert.Contains(t, models, ModelLlama33_8B)
}

func TestMetaProvider_Chat(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer")

		// Send mock response
		resp := ChatResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   ModelLlama4Maverick,
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "Hello! How can I help you?",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create provider
	config := &Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   ModelLlama4Maverick,
	}
	p, err := NewMetaProvider(config)
	require.NoError(t, err)

	// Test chat
	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	resp, err := p.Chat(context.Background(), messages)
	require.NoError(t, err)

	assert.Equal(t, "Hello! How can I help you?", resp.Content)
	assert.Equal(t, 10, resp.Usage.PromptTokens)
	assert.Equal(t, 20, resp.Usage.CompletionTokens)
	assert.Equal(t, 30, resp.Usage.TotalTokens)
	assert.Equal(t, ModelLlama4Maverick, resp.Model)
}

func TestMetaProvider_ChatWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Verify options were applied
		assert.Equal(t, ModelLlama33_8B, req.Model)
		assert.Equal(t, 0.5, req.Temperature)
		assert.Equal(t, 100, req.MaxTokens)
		assert.Equal(t, 0.95, req.TopP)

		resp := ChatResponse{
			ID:     "test-id",
			Object: "chat.completion",
			Model:  req.Model,
			Choices: []Choice{
				{
					Message: Message{
						Role:    "assistant",
						Content: "Response",
					},
				},
			},
			Usage: Usage{TotalTokens: 50},
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := &Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   ModelLlama4Maverick,
	}
	p, err := NewMetaProvider(config)
	require.NoError(t, err)

	messages := []provider.Message{
		{Role: "user", Content: "Test"},
	}

	resp, err := p.Chat(context.Background(), messages,
		provider.WithModel(ModelLlama33_8B),
		provider.WithTemperature(0.5),
		provider.WithMaxTokens(100),
		provider.WithTopP(0.95),
	)

	require.NoError(t, err)
	assert.Equal(t, "Response", resp.Content)
	assert.Equal(t, ModelLlama33_8B, resp.Model)
}

func TestMetaProvider_ChatError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: ErrorDetail{
				Message: "Invalid API key",
				Type:    ErrTypeAuthentication,
			},
		})
	}))
	defer server.Close()

	config := &Config{
		APIKey:  "invalid-key",
		BaseURL: server.URL,
		Model:   ModelLlama4Maverick,
	}
	p, err := NewMetaProvider(config)
	require.NoError(t, err)

	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	_, err = p.Chat(context.Background(), messages)
	require.Error(t, err)

	metaErr, ok := err.(*MetaError)
	require.True(t, ok)
	assert.Equal(t, 401, metaErr.StatusCode)
	assert.True(t, metaErr.IsAuthenticationError())
}

func TestMetaProvider_Stream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.True(t, req.Stream)

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Send SSE chunks
		chunks := []string{
			`data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"llama-4","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}` + "\n\n",
			`data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"llama-4","choices":[{"index":0,"delta":{"content":" World"},"finish_reason":null}]}` + "\n\n",
			`data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"llama-4","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":"stop"}]}` + "\n\n",
			`data: [DONE]` + "\n\n",
		}

		for _, chunk := range chunks {
			w.Write([]byte(chunk))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	config := &Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   ModelLlama4Maverick,
	}
	p, err := NewMetaProvider(config)
	require.NoError(t, err)

	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	eventChan, err := p.Stream(context.Background(), messages)
	require.NoError(t, err)

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
		case provider.EventTypeError:
			t.Fatalf("Unexpected error: %v", event.Error)
		}
	}

	assert.True(t, gotStart, "Should receive start event")
	assert.True(t, gotEnd, "Should receive end event")
	assert.Contains(t, content, "Hello")
	assert.Contains(t, content, "World")
}

func TestMetaProvider_Close(t *testing.T) {
	p := &MetaProvider{}
	err := p.Close()
	assert.NoError(t, err)
}

func TestBuildRequest(t *testing.T) {
	config := &Config{
		APIKey:           "test-key",
		Model:            ModelLlama4Maverick,
		Temperature:      0.7,
		TopP:             0.9,
		MaxTokens:        2048,
		PresencePenalty:  0.1,
		FrequencyPenalty: 0.2,
		Stop:             []string{"STOP"},
	}

	p := &MetaProvider{config: config}

	messages := []provider.Message{
		{Role: "system", Content: "You are a helpful assistant"},
		{Role: "user", Content: "Hello"},
	}

	options := &provider.ChatOptions{
		Temperature: 0.5,
		MaxTokens:   100,
	}

	req := p.buildRequest(messages, options)

	assert.Equal(t, ModelLlama4Maverick, req.Model)
	assert.Equal(t, 0.5, req.Temperature) // Option overrides config
	assert.Equal(t, 100, req.MaxTokens)   // Option overrides config
	assert.Equal(t, 0.9, req.TopP)        // From config
	assert.Equal(t, 0.1, req.PresencePenalty)
	assert.Equal(t, 0.2, req.FrequencyPenalty)
	assert.Len(t, req.Messages, 2)
	assert.Equal(t, "system", req.Messages[0].Role)
	assert.Equal(t, "user", req.Messages[1].Role)
}
