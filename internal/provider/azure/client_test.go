package azure

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAzureProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid config",
			config: Config{
				Endpoint:   "https://test.openai.azure.com",
				APIKey:     "test-key",
				Deployment: "gpt-4",
			},
			expectError: false,
		},
		{
			name: "missing endpoint",
			config: Config{
				APIKey:     "test-key",
				Deployment: "gpt-4",
			},
			expectError: true,
		},
		{
			name: "missing API key",
			config: Config{
				Endpoint:   "https://test.openai.azure.com",
				Deployment: "gpt-4",
			},
			expectError: true,
		},
		{
			name: "missing deployment",
			config: Config{
				Endpoint: "https://test.openai.azure.com",
				APIKey:   "test-key",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prov, err := NewAzureProvider(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, prov)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, prov)
				assert.Equal(t, "azure", prov.Name())
			}
		})
	}
}

func TestAzureProvider_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/openai/deployments/gpt-4/chat/completions", r.URL.Path)
		assert.Contains(t, r.URL.Query().Get("api-version"), "2024")
		assert.Equal(t, "test-key", r.Header.Get("api-key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
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
				"completion_tokens": 7,
				"total_tokens": 17
			}
		}`))
	}))
	defer server.Close()

	config := Config{
		Endpoint:   server.URL,
		APIKey:     "test-key",
		Deployment: "gpt-4",
	}

	prov, err := NewAzureProvider(config)
	require.NoError(t, err)

	messages := []provider.Message{
		{
			Role:    "user",
			Content: "Hello",
		},
	}

	resp, err := prov.Chat(context.Background(), messages)
	require.NoError(t, err)

	assert.Equal(t, "Hello! How can I help you?", resp.Content)
	assert.Equal(t, 10, resp.Usage.PromptTokens)
	assert.Equal(t, 7, resp.Usage.CompletionTokens)
	assert.Equal(t, 17, resp.Usage.TotalTokens)
	assert.Equal(t, "gpt-4", resp.Model)
}

func TestAzureProvider_ChatError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{
			"error": {
				"message": "Invalid API key",
				"type": "invalid_request_error"
			}
		}`))
	}))
	defer server.Close()

	config := Config{
		Endpoint:   server.URL,
		APIKey:     "invalid-key",
		Deployment: "gpt-4",
	}

	prov, err := NewAzureProvider(config)
	require.NoError(t, err)

	messages := []provider.Message{
		{
			Role:    "user",
			Content: "Hello",
		},
	}

	_, err = prov.Chat(context.Background(), messages)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid API key")
}

func TestAzureProvider_Stream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/openai/deployments/gpt-4/chat/completions", r.URL.Path)
		assert.Contains(t, r.URL.Query().Get("api-version"), "2024")
		assert.Equal(t, "test-key", r.Header.Get("api-key"))

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Send streaming chunks
		chunks := []string{
			`data: {"id":"1","object":"chat.completion.chunk","created":1234,"model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}`,
			`data: {"id":"1","object":"chat.completion.chunk","created":1234,"model":"gpt-4","choices":[{"index":0,"delta":{"content":" there"},"finish_reason":null}]}`,
			`data: {"id":"1","object":"chat.completion.chunk","created":1234,"model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			w.Write([]byte(chunk + "\n\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	config := Config{
		Endpoint:   server.URL,
		APIKey:     "test-key",
		Deployment: "gpt-4",
	}

	prov, err := NewAzureProvider(config)
	require.NoError(t, err)

	messages := []provider.Message{
		{
			Role:    "user",
			Content: "Hello",
		},
	}

	eventChan, err := prov.Stream(context.Background(), messages)
	require.NoError(t, err)

	var content string
	for event := range eventChan {
		require.NoError(t, event.Error)
		content += event.Content
	}

	assert.Equal(t, "Hello there", content)
}

func TestAzureProvider_StreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{
			"error": {
				"message": "Rate limit exceeded",
				"type": "rate_limit_error"
			}
		}`))
	}))
	defer server.Close()

	config := Config{
		Endpoint:   server.URL,
		APIKey:     "test-key",
		Deployment: "gpt-4",
	}

	prov, err := NewAzureProvider(config)
	require.NoError(t, err)

	messages := []provider.Message{
		{
			Role:    "user",
			Content: "Hello",
		},
	}

	_, err = prov.Stream(context.Background(), messages)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit")
}

func TestAzureProvider_Models(t *testing.T) {
	config := Config{
		Endpoint:   "https://test.openai.azure.com",
		APIKey:     "test-key",
		Deployment: "my-gpt4-deployment",
	}

	prov, err := NewAzureProvider(config)
	require.NoError(t, err)

	models := prov.Models()
	assert.Len(t, models, 1)
	assert.Equal(t, "my-gpt4-deployment", models[0])
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Endpoint:   "https://test.openai.azure.com",
				APIKey:     "test-key",
				Deployment: "gpt-4",
			},
			wantErr: false,
		},
		{
			name: "missing endpoint",
			config: Config{
				APIKey:     "test-key",
				Deployment: "gpt-4",
			},
			wantErr: true,
		},
		{
			name: "missing API key",
			config: Config{
				Endpoint:   "https://test.openai.azure.com",
				Deployment: "gpt-4",
			},
			wantErr: true,
		},
		{
			name: "missing deployment",
			config: Config{
				Endpoint: "https://test.openai.azure.com",
				APIKey:   "test-key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_SetDefaults(t *testing.T) {
	config := Config{
		Endpoint:   "https://test.openai.azure.com",
		APIKey:     "test-key",
		Deployment: "gpt-4",
	}

	config.SetDefaults()

	assert.Equal(t, DefaultAPIVersion, config.APIVersion)
	assert.Equal(t, DefaultTimeout, config.Timeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.NotNil(t, config.HTTPClient)
}
