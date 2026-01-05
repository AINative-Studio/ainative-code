package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/AINative-studio/ainative-code/internal/provider/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetaProvider_Integration tests the Meta LLAMA provider with the real API
// Run with: go test -v -tags=integration ./tests/integration -run TestMetaProvider
func TestMetaProvider_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiKey := os.Getenv("META_API_KEY")
	if apiKey == "" {
		t.Skip("META_API_KEY environment variable not set")
	}

	config := &meta.Config{
		APIKey:      apiKey,
		BaseURL:     os.Getenv("META_BASE_URL"),
		Model:       os.Getenv("META_MODEL"),
		Temperature: 0.7,
		MaxTokens:   100,
		Timeout:     60 * time.Second,
	}

	// Use defaults if env vars not set
	if config.BaseURL == "" {
		config.BaseURL = meta.DefaultBaseURL
	}
	if config.Model == "" {
		config.Model = meta.ModelLlama4Maverick
	}

	provider, err := meta.NewMetaProvider(config)
	require.NoError(t, err)
	defer provider.Close()

	t.Run("Chat", func(t *testing.T) {
		messages := []provider.Message{
			{
				Role:    "user",
				Content: "Say 'Hello, World!' and nothing else.",
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := provider.Chat(ctx, messages)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.Content)
		assert.Contains(t, resp.Content, "Hello")
		assert.Greater(t, resp.Usage.TotalTokens, 0)
		assert.Greater(t, resp.Usage.PromptTokens, 0)
		assert.Greater(t, resp.Usage.CompletionTokens, 0)
		assert.NotEmpty(t, resp.Model)

		t.Logf("Response: %s", resp.Content)
		t.Logf("Usage: %d prompt + %d completion = %d total tokens",
			resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
		t.Logf("Model: %s", resp.Model)
	})

	t.Run("Chat with Options", func(t *testing.T) {
		messages := []provider.Message{
			{
				Role:    "user",
				Content: "Count from 1 to 5.",
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := provider.Chat(ctx, messages,
			provider.WithTemperature(0.5),
			provider.WithMaxTokens(50),
		)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.Content)
		assert.Greater(t, resp.Usage.TotalTokens, 0)

		t.Logf("Response: %s", resp.Content)
	})

	t.Run("Stream", func(t *testing.T) {
		messages := []provider.Message{
			{
				Role:    "user",
				Content: "Write a haiku about coding.",
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		eventChan, err := provider.Stream(ctx, messages)
		require.NoError(t, err)

		var fullContent string
		var gotStart, gotEnd bool
		var deltaCount int

		for event := range eventChan {
			switch event.Type {
			case provider.EventTypeContentStart:
				gotStart = true
				t.Log("Stream started")

			case provider.EventTypeContentDelta:
				fullContent += event.Content
				deltaCount++
				t.Logf("Delta %d: %q", deltaCount, event.Content)

			case provider.EventTypeContentEnd:
				gotEnd = true
				t.Log("Stream ended")

			case provider.EventTypeError:
				t.Fatalf("Stream error: %v", event.Error)
			}
		}

		assert.True(t, gotStart, "Should receive start event")
		assert.True(t, gotEnd, "Should receive end event")
		assert.Greater(t, deltaCount, 0, "Should receive at least one delta")
		assert.NotEmpty(t, fullContent, "Should receive content")

		t.Logf("\nFull response (%d deltas):\n%s", deltaCount, fullContent)
	})

	t.Run("Multiple Models", func(t *testing.T) {
		models := []string{
			meta.ModelLlama4Maverick,
			meta.ModelLlama33_8B,
		}

		for _, model := range models {
			t.Run(model, func(t *testing.T) {
				messages := []provider.Message{
					{
						Role:    "user",
						Content: "Say hi in one word.",
					},
				}

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				resp, err := provider.Chat(ctx, messages, provider.WithModel(model))
				require.NoError(t, err)

				assert.NotEmpty(t, resp.Content)
				t.Logf("%s response: %s", model, resp.Content)
			})
		}
	})

	t.Run("System Message", func(t *testing.T) {
		messages := []provider.Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant that responds in pirate speak.",
			},
			{
				Role:    "user",
				Content: "Tell me about the weather.",
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := provider.Chat(ctx, messages)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.Content)
		t.Logf("Pirate response: %s", resp.Content)
	})
}

// TestMetaProvider_ModelValidation tests model validation
func TestMetaProvider_ModelValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiKey := os.Getenv("META_API_KEY")
	if apiKey == "" {
		t.Skip("META_API_KEY environment variable not set")
	}

	t.Run("invalid model", func(t *testing.T) {
		config := &meta.Config{
			APIKey:  apiKey,
			BaseURL: meta.DefaultBaseURL,
			Model:   "gpt-4", // Invalid Meta model
		}

		_, err := meta.NewMetaProvider(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid model")
	})

	t.Run("all valid models", func(t *testing.T) {
		models := []string{
			meta.ModelLlama4Maverick,
			meta.ModelLlama4Scout,
			meta.ModelLlama33_70B,
			meta.ModelLlama33_8B,
		}

		for _, model := range models {
			config := &meta.Config{
				APIKey:  apiKey,
				BaseURL: meta.DefaultBaseURL,
				Model:   model,
			}

			provider, err := meta.NewMetaProvider(config)
			assert.NoError(t, err, "Model %s should be valid", model)
			if provider != nil {
				provider.Close()
			}
		}
	})
}

// TestMetaProvider_ErrorHandling tests error scenarios
func TestMetaProvider_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("invalid API key", func(t *testing.T) {
		config := &meta.Config{
			APIKey:  "invalid-key",
			BaseURL: meta.DefaultBaseURL,
			Model:   meta.ModelLlama4Maverick,
		}

		provider, err := meta.NewMetaProvider(config)
		require.NoError(t, err)

		messages := []provider.Message{
			{Role: "user", Content: "Hello"},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = provider.Chat(ctx, messages)
		assert.Error(t, err)

		// Check if it's a Meta error
		if metaErr, ok := err.(*meta.MetaError); ok {
			assert.True(t, metaErr.IsAuthenticationError() || metaErr.StatusCode == 401)
			t.Logf("Expected error: %v", metaErr)
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		apiKey := os.Getenv("META_API_KEY")
		if apiKey == "" {
			t.Skip("META_API_KEY environment variable not set")
		}

		config := &meta.Config{
			APIKey:  apiKey,
			BaseURL: meta.DefaultBaseURL,
			Model:   meta.ModelLlama4Maverick,
		}

		provider, err := meta.NewMetaProvider(config)
		require.NoError(t, err)

		messages := []provider.Message{
			{Role: "user", Content: "Write a very long essay."},
		}

		// Set very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		_, err = provider.Chat(ctx, messages)
		assert.Error(t, err)
		t.Logf("Expected timeout error: %v", err)
	})
}
