package bedrock

import (
	"testing"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/stretchr/testify/assert"
)

func TestConvertMessages(t *testing.T) {
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
		{
			name: "empty messages",
			messages: []provider.Message{},
			systemPrompt: "",
			expectedSystem: "",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiMessages, systemPrompt := convertMessages(tt.messages, tt.systemPrompt)

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

func TestBuildBedrockRequest(t *testing.T) {
	tests := []struct {
		name            string
		messages        []provider.Message
		options         *provider.ChatOptions
		expectedModel   string
		expectedMaxTokens int
		hasTemp         bool
		hasTopP         bool
		hasSystem       bool
	}{
		{
			name: "basic request",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			options: &provider.ChatOptions{
				Model:     "anthropic.claude-3-5-sonnet-20241022-v2:0",
				MaxTokens: 1024,
			},
			expectedModel:     "anthropic.claude-3-5-sonnet-20241022-v2:0",
			expectedMaxTokens: 1024,
			hasTemp:           false,
			hasTopP:           false,
			hasSystem:         false,
		},
		{
			name: "with temperature and top_p",
			messages: []provider.Message{
				{Role: "user", Content: "Test"},
			},
			options: &provider.ChatOptions{
				Model:       "anthropic.claude-3-haiku-20240307-v1:0",
				MaxTokens:   512,
				Temperature: 0.7,
				TopP:        0.9,
			},
			expectedModel:     "anthropic.claude-3-haiku-20240307-v1:0",
			expectedMaxTokens: 512,
			hasTemp:           true,
			hasTopP:           true,
			hasSystem:         false,
		},
		{
			name: "with system prompt",
			messages: []provider.Message{
				{Role: "system", Content: "Be helpful"},
				{Role: "user", Content: "Hello"},
			},
			options: &provider.ChatOptions{
				Model:     "anthropic.claude-v2",
				MaxTokens: 256,
			},
			expectedModel:     "anthropic.claude-v2",
			expectedMaxTokens: 256,
			hasTemp:           false,
			hasTopP:           false,
			hasSystem:         true,
		},
		{
			name: "with stop sequences",
			messages: []provider.Message{
				{Role: "user", Content: "Count to 10"},
			},
			options: &provider.ChatOptions{
				Model:         "anthropic.claude-instant-v1",
				MaxTokens:     100,
				StopSequences: []string{"STOP", "END"},
			},
			expectedModel:     "anthropic.claude-instant-v1",
			expectedMaxTokens: 100,
			hasTemp:           false,
			hasTopP:           false,
			hasSystem:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := buildBedrockRequest(tt.messages, tt.options)

			// Verify inference configuration
			assert.NotNil(t, req.InferenceConfig)
			assert.Equal(t, tt.expectedMaxTokens, req.InferenceConfig.MaxTokens)

			if tt.hasTemp {
				assert.NotNil(t, req.InferenceConfig.Temperature)
				assert.Equal(t, tt.options.Temperature, *req.InferenceConfig.Temperature)
			}

			if tt.hasTopP {
				assert.NotNil(t, req.InferenceConfig.TopP)
				assert.Equal(t, tt.options.TopP, *req.InferenceConfig.TopP)
			}

			// Verify messages
			assert.Greater(t, len(req.Messages), 0)

			// Verify system prompt
			if tt.hasSystem {
				assert.Len(t, req.System, 1)
				assert.NotEmpty(t, req.System[0].Text)
			}

			// Verify stop sequences
			if len(tt.options.StopSequences) > 0 {
				assert.Equal(t, tt.options.StopSequences, req.InferenceConfig.StopSequences)
			}
		})
	}
}

func TestParseBedrockResponse(t *testing.T) {
	tests := []struct {
		name            string
		response        *bedrockResponse
		model           string
		expectedContent string
		expectedInput   int
		expectedOutput  int
	}{
		{
			name: "single text content",
			response: &bedrockResponse{
				Output: bedrockOutput{
					Message: bedrockMessage{
						Role: "assistant",
						Content: []bedrockContent{
							{Type: "text", Text: "Hello!"},
						},
					},
				},
				Usage: bedrockUsage{
					InputTokens:  10,
					OutputTokens: 5,
				},
			},
			model:           "anthropic.claude-3-5-sonnet-20241022-v2:0",
			expectedContent: "Hello!",
			expectedInput:   10,
			expectedOutput:  5,
		},
		{
			name: "multiple text contents",
			response: &bedrockResponse{
				Output: bedrockOutput{
					Message: bedrockMessage{
						Role: "assistant",
						Content: []bedrockContent{
							{Type: "text", Text: "First part."},
							{Type: "text", Text: "Second part."},
						},
					},
				},
				Usage: bedrockUsage{
					InputTokens:  20,
					OutputTokens: 10,
				},
			},
			model:           "anthropic.claude-3-haiku-20240307-v1:0",
			expectedContent: "First part.\nSecond part.",
			expectedInput:   20,
			expectedOutput:  10,
		},
		{
			name: "empty content",
			response: &bedrockResponse{
				Output: bedrockOutput{
					Message: bedrockMessage{
						Role:    "assistant",
						Content: []bedrockContent{},
					},
				},
				Usage: bedrockUsage{
					InputTokens:  5,
					OutputTokens: 0,
				},
			},
			model:           "anthropic.claude-v2",
			expectedContent: "",
			expectedInput:   5,
			expectedOutput:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := parseBedrockResponse(tt.response, tt.model)

			assert.Equal(t, tt.expectedContent, resp.Content)
			assert.Equal(t, tt.model, resp.Model)
			assert.Equal(t, tt.expectedInput, resp.Usage.PromptTokens)
			assert.Equal(t, tt.expectedOutput, resp.Usage.CompletionTokens)
			assert.Equal(t, tt.expectedInput+tt.expectedOutput, resp.Usage.TotalTokens)
		})
	}
}
