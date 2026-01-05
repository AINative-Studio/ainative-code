package meta

import "time"

// Supported Meta LLAMA models
const (
	// Llama 4 Models - Mixture of Experts
	ModelLlama4Maverick = "Llama-4-Maverick-17B-128E-Instruct-FP8" // 400B total params, 17B active
	ModelLlama4Scout    = "Llama-4-Scout-17B-16E"                  // 109B total params, 17B active

	// Llama 3.3 Models - Dense
	ModelLlama33_70B = "Llama-3.3-70B-Instruct" // 70B params
	ModelLlama33_8B  = "Llama-3.3-8B-Instruct"  // 8B params (fast)

	// API endpoints
	DefaultBaseURL = "https://api.llama.com/compat/v1"
	DefaultTimeout = 60 * time.Second
)

// ChatRequest represents a Meta LLAMA API chat completion request
// Compatible with OpenAI Chat Completions API format
type ChatRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Temperature      float64   `json:"temperature,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	Stream           bool      `json:"stream,omitempty"`
	Stop             []string  `json:"stop,omitempty"`
	PresencePenalty  float64   `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64   `json:"frequency_penalty,omitempty"`
	N                int       `json:"n,omitempty"`
	User             string    `json:"user,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // "system", "user", or "assistant"
	Content string `json:"content"` // Message content
}

// ChatResponse represents a Meta LLAMA API chat completion response
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// StreamChoice represents a streaming completion choice
type StreamChoice struct {
	Index        int           `json:"index"`
	Delta        MessageDelta  `json:"delta"`
	FinishReason string        `json:"finish_reason,omitempty"`
}

// MessageDelta represents a partial message in streaming
type MessageDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// StreamResponse represents a streaming chat completion chunk
type StreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

// ModelInfo contains information about a Meta LLAMA model
type ModelInfo struct {
	ID             string
	Name           string
	Description    string
	ParameterCount string
	Architecture   string
	MaxTokens      int
	Recommended    bool
}

// GetSupportedModels returns information about all supported Meta LLAMA models
func GetSupportedModels() []ModelInfo {
	return []ModelInfo{
		{
			ID:             ModelLlama4Maverick,
			Name:           "Llama 4 Maverick",
			Description:    "Most capable model with 400B total parameters using Mixture of Experts",
			ParameterCount: "400B total, 17B active",
			Architecture:   "Mixture of Experts (128 experts)",
			MaxTokens:      8192,
			Recommended:    true,
		},
		{
			ID:             ModelLlama4Scout,
			Name:           "Llama 4 Scout",
			Description:    "Efficient model with 109B total parameters using Mixture of Experts",
			ParameterCount: "109B total, 17B active",
			Architecture:   "Mixture of Experts (16 experts)",
			MaxTokens:      8192,
			Recommended:    false,
		},
		{
			ID:             ModelLlama33_70B,
			Name:           "Llama 3.3 70B",
			Description:    "Large dense model with 70B parameters",
			ParameterCount: "70B",
			Architecture:   "Dense transformer",
			MaxTokens:      8192,
			Recommended:    false,
		},
		{
			ID:             ModelLlama33_8B,
			Name:           "Llama 3.3 8B",
			Description:    "Fast and efficient model with 8B parameters",
			ParameterCount: "8B",
			Architecture:   "Dense transformer",
			MaxTokens:      8192,
			Recommended:    false,
		},
	}
}

// IsValidModel checks if a model ID is valid
func IsValidModel(modelID string) bool {
	validModels := map[string]bool{
		ModelLlama4Maverick: true,
		ModelLlama4Scout:    true,
		ModelLlama33_70B:    true,
		ModelLlama33_8B:     true,
	}
	return validModels[modelID]
}
