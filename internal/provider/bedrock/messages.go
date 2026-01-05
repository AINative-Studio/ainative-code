package bedrock

import (
	"github.com/AINative-studio/ainative-code/internal/provider"
)

// bedrockRequest represents a request to the Bedrock API
type bedrockRequest struct {
	Messages        []bedrockMessage       `json:"messages"`
	System          []bedrockSystemMessage `json:"system,omitempty"`
	InferenceConfig bedrockInferenceConfig `json:"inferenceConfig"`
}

// bedrockMessage represents a message in Bedrock format
type bedrockMessage struct {
	Role    string           `json:"role"`
	Content []bedrockContent `json:"content"`
}

// bedrockContent represents content within a message
type bedrockContent struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

// bedrockSystemMessage represents a system message
type bedrockSystemMessage struct {
	Text string `json:"text"`
}

// bedrockInferenceConfig contains inference parameters
type bedrockInferenceConfig struct {
	MaxTokens     int       `json:"maxTokens"`
	Temperature   *float64  `json:"temperature,omitempty"`
	TopP          *float64  `json:"topP,omitempty"`
	StopSequences []string  `json:"stopSequences,omitempty"`
}

// bedrockResponse represents a response from the Bedrock API
type bedrockResponse struct {
	Output bedrockOutput `json:"output"`
	Usage  bedrockUsage  `json:"usage"`
}

// bedrockOutput contains the response message
type bedrockOutput struct {
	Message bedrockMessage `json:"message"`
}

// bedrockUsage represents token usage statistics
type bedrockUsage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
}

// convertMessages converts provider messages to Bedrock format
// Returns the messages and extracted system prompt
func convertMessages(messages []provider.Message, systemPrompt string) ([]bedrockMessage, string) {
	var bedrockMessages []bedrockMessage
	var extractedSystem string

	for _, msg := range messages {
		// Extract system messages separately
		if msg.Role == "system" {
			if extractedSystem != "" {
				extractedSystem += "\n\n"
			}
			extractedSystem += msg.Content
			continue
		}

		bedrockMessages = append(bedrockMessages, bedrockMessage{
			Role: msg.Role,
			Content: []bedrockContent{
				{
					Type: "text",
					Text: msg.Content,
				},
			},
		})
	}

	// Combine extracted system messages with provided system prompt
	finalSystem := extractedSystem
	if systemPrompt != "" {
		if finalSystem != "" {
			finalSystem += "\n\n"
		}
		finalSystem += systemPrompt
	}

	return bedrockMessages, finalSystem
}

// buildBedrockRequest builds a Bedrock API request
func buildBedrockRequest(messages []provider.Message, options *provider.ChatOptions) *bedrockRequest {
	// Convert messages
	bedrockMessages, systemPrompt := convertMessages(messages, options.SystemPrompt)

	// Build inference config
	inferenceConfig := bedrockInferenceConfig{
		MaxTokens: options.MaxTokens,
	}

	// Add optional parameters
	if options.Temperature > 0 {
		inferenceConfig.Temperature = &options.Temperature
	}
	if options.TopP > 0 && options.TopP < 1.0 {
		inferenceConfig.TopP = &options.TopP
	}
	if len(options.StopSequences) > 0 {
		inferenceConfig.StopSequences = options.StopSequences
	}

	// Build request
	req := &bedrockRequest{
		Messages:        bedrockMessages,
		InferenceConfig: inferenceConfig,
	}

	// Add system prompt if present
	if systemPrompt != "" {
		req.System = []bedrockSystemMessage{
			{Text: systemPrompt},
		}
	}

	return req
}

// parseBedrockResponse parses a Bedrock API response
func parseBedrockResponse(resp *bedrockResponse, model string) provider.Response {
	// Extract text content
	var content string
	for i, block := range resp.Output.Message.Content {
		if block.Type == "text" || block.Type == "" {
			if i > 0 && content != "" {
				content += "\n"
			}
			content += block.Text
		}
	}

	return provider.Response{
		Content: content,
		Model:   model,
		Usage: provider.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}
}
