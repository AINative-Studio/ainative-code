package ollama

import (
	"strings"

	"github.com/AINative-studio/ainative-code/internal/provider"
)

// convertMessages converts provider messages to Ollama format
// Returns the converted messages and any system prompt found in messages
func convertMessages(messages []provider.Message, systemPrompt string) ([]ollamaMessage, string) {
	var ollamaMessages []ollamaMessage
	var systemInMessages string

	for _, msg := range messages {
		// Collect system messages separately
		if msg.Role == "system" {
			if systemInMessages != "" {
				systemInMessages += "\n\n"
			}
			systemInMessages += msg.Content
			continue
		}

		ollamaMessages = append(ollamaMessages, ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Merge system prompts
	mergedSystem := mergeSystemPrompts(systemInMessages, systemPrompt)

	// If we have a merged system prompt, prepend it as a system message
	if mergedSystem != "" && len(ollamaMessages) > 0 {
		// Insert system message at the beginning
		ollamaMessages = append([]ollamaMessage{{
			Role:    "system",
			Content: mergedSystem,
		}}, ollamaMessages...)
		return ollamaMessages, ""
	}

	return ollamaMessages, mergedSystem
}

// mergeSystemPrompts combines system prompts from messages and options
func mergeSystemPrompts(systemInMsg, systemOption string) string {
	if systemInMsg != "" && systemOption != "" {
		return systemInMsg + "\n\n" + systemOption
	}
	if systemInMsg != "" {
		return systemInMsg
	}
	return systemOption
}

// buildOllamaRequest creates an Ollama API request from provider types
func buildOllamaRequest(config *Config, messages []provider.Message, options *provider.ChatOptions, stream bool) *ollamaRequest {
	// Convert messages
	ollamaMessages, systemPrompt := convertMessages(messages, options.SystemPrompt)

	// Build options
	opts := &ollamaOptions{
		NumCtx:      config.NumCtx,
		Temperature: config.Temperature,
		TopK:        config.TopK,
		TopP:        config.TopP,
	}

	// Override with options if provided
	if options.Temperature > 0 {
		opts.Temperature = options.Temperature
	}
	if options.TopP > 0 {
		opts.TopP = options.TopP
	}
	if len(options.StopSequences) > 0 {
		opts.Stop = options.StopSequences
	}

	// Create request
	req := &ollamaRequest{
		Model:    options.Model,
		Messages: ollamaMessages,
		Stream:   stream,
		Options:  opts,
	}

	// Add system prompt if not in messages
	if systemPrompt != "" {
		// Prepend system message
		req.Messages = append([]ollamaMessage{{
			Role:    "system",
			Content: systemPrompt,
		}}, req.Messages...)
	}

	return req
}

// formatPromptForModel applies model-specific formatting if needed
// Some Ollama models may have specific prompt templates
func formatPromptForModel(model string, content string) string {
	// Most Ollama models handle formatting internally
	// This function is here for future model-specific customizations

	modelLower := strings.ToLower(model)

	// CodeLlama might benefit from code-specific formatting
	if strings.Contains(modelLower, "codellama") || strings.Contains(modelLower, "code") {
		// CodeLlama uses standard format, no special handling needed
		return content
	}

	// Llama models use standard format
	if strings.Contains(modelLower, "llama") {
		return content
	}

	// Mistral models use standard format
	if strings.Contains(modelLower, "mistral") {
		return content
	}

	// Default: return as-is
	return content
}
