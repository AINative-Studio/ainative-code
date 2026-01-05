package ollama

import (
	"testing"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/stretchr/testify/assert"
)

func TestConvertMessages(t *testing.T) {
	tests := []struct {
		name             string
		messages         []provider.Message
		systemPrompt     string
		expectedMessages []ollamaMessage
		expectedSystem   string
	}{
		{
			name: "simple user and assistant messages",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
				{Role: "user", Content: "How are you?"},
			},
			systemPrompt: "",
			expectedMessages: []ollamaMessage{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
				{Role: "user", Content: "How are you?"},
			},
			expectedSystem: "",
		},
		{
			name: "messages with system prompt in options",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			systemPrompt: "You are a helpful assistant",
			expectedMessages: []ollamaMessage{
				{Role: "system", Content: "You are a helpful assistant"},
				{Role: "user", Content: "Hello"},
			},
			expectedSystem: "",
		},
		{
			name: "messages with system message",
			messages: []provider.Message{
				{Role: "system", Content: "You are helpful"},
				{Role: "user", Content: "Hello"},
			},
			systemPrompt: "",
			expectedMessages: []ollamaMessage{
				{Role: "system", Content: "You are helpful"},
				{Role: "user", Content: "Hello"},
			},
			expectedSystem: "",
		},
		{
			name: "messages with both system message and prompt",
			messages: []provider.Message{
				{Role: "system", Content: "Context: you are an AI"},
				{Role: "user", Content: "Hello"},
			},
			systemPrompt: "Additional instruction",
			expectedMessages: []ollamaMessage{
				{Role: "system", Content: "Context: you are an AI\n\nAdditional instruction"},
				{Role: "user", Content: "Hello"},
			},
			expectedSystem: "",
		},
		{
			name: "empty messages",
			messages: []provider.Message{},
			systemPrompt: "System",
			expectedMessages: []ollamaMessage{},
			expectedSystem: "System", // When no messages, system stays as return value
		},
		{
			name: "multiple system messages merged",
			messages: []provider.Message{
				{Role: "system", Content: "First system"},
				{Role: "system", Content: "Second system"},
				{Role: "user", Content: "Hello"},
			},
			systemPrompt: "",
			expectedMessages: []ollamaMessage{
				{Role: "system", Content: "First system\n\nSecond system"},
				{Role: "user", Content: "Hello"},
			},
			expectedSystem: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, systemInMessages := convertMessages(tt.messages, tt.systemPrompt)

			assert.Equal(t, len(tt.expectedMessages), len(messages))
			for i, expected := range tt.expectedMessages {
				if i < len(messages) {
					assert.Equal(t, expected.Role, messages[i].Role)
					assert.Equal(t, expected.Content, messages[i].Content)
				}
			}

			assert.Equal(t, tt.expectedSystem, systemInMessages)
		})
	}
}

func TestBuildOllamaRequest(t *testing.T) {
	config := &Config{
		Model:       "llama2",
		NumCtx:      2048,
		Temperature: 0.7,
		TopK:        40,
		TopP:        0.9,
	}

	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	options := &provider.ChatOptions{
		Model:        "llama2",
		MaxTokens:    1024,
		Temperature:  0.8,
		TopP:         0.95,
		StopSequences: []string{"END"},
		SystemPrompt: "You are helpful",
	}

	t.Run("build request for chat", func(t *testing.T) {
		req := buildOllamaRequest(config, messages, options, false)

		assert.Equal(t, "llama2", req.Model)
		assert.False(t, req.Stream)
		assert.NotNil(t, req.Options)
		assert.Equal(t, 2048, req.Options.NumCtx)
		assert.Equal(t, 0.8, req.Options.Temperature)
		assert.Equal(t, 0.95, req.Options.TopP)
		assert.Equal(t, 40, req.Options.TopK)
		assert.Equal(t, []string{"END"}, req.Options.Stop)
		// System prompt is prepended, so we expect 2 messages (system + user)
		assert.GreaterOrEqual(t, len(req.Messages), 1)
	})

	t.Run("build request for streaming", func(t *testing.T) {
		req := buildOllamaRequest(config, messages, options, true)

		assert.Equal(t, "llama2", req.Model)
		assert.True(t, req.Stream)
	})

	t.Run("use config defaults when options not set", func(t *testing.T) {
		minimalOptions := &provider.ChatOptions{
			Model: "llama2",
		}

		req := buildOllamaRequest(config, messages, minimalOptions, false)

		assert.Equal(t, config.Temperature, req.Options.Temperature)
		assert.Equal(t, config.TopK, req.Options.TopK)
		assert.Equal(t, config.TopP, req.Options.TopP)
	})
}

func TestMergeSystemPrompts(t *testing.T) {
	tests := []struct {
		name          string
		systemInMsg   string
		systemOption  string
		expected      string
	}{
		{
			name:         "both present",
			systemInMsg:  "System in message",
			systemOption: "System in option",
			expected:     "System in message\n\nSystem in option",
		},
		{
			name:         "only message",
			systemInMsg:  "System in message",
			systemOption: "",
			expected:     "System in message",
		},
		{
			name:         "only option",
			systemInMsg:  "",
			systemOption: "System in option",
			expected:     "System in option",
		},
		{
			name:         "both empty",
			systemInMsg:  "",
			systemOption: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeSystemPrompts(tt.systemInMsg, tt.systemOption)
			assert.Equal(t, tt.expected, result)
		})
	}
}
