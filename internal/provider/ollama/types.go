package ollama

import "time"

// ollamaRequest represents a request to the Ollama chat API
type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages,omitempty"`
	Prompt   string          `json:"prompt,omitempty"` // For legacy API
	Stream   bool            `json:"stream"`
	Options  *ollamaOptions  `json:"options,omitempty"`
	Context  []int           `json:"context,omitempty"` // Conversation context
}

// ollamaMessage represents a message in Ollama format
type ollamaMessage struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// ollamaOptions represents generation options for Ollama
type ollamaOptions struct {
	NumCtx      int     `json:"num_ctx,omitempty"`       // Context window size
	Temperature float64 `json:"temperature,omitempty"`   // Sampling temperature
	TopK        int     `json:"top_k,omitempty"`         // Top-k sampling
	TopP        float64 `json:"top_p,omitempty"`         // Nucleus sampling
	RepeatPenalty float64 `json:"repeat_penalty,omitempty"` // Repeat penalty
	Stop        []string `json:"stop,omitempty"`          // Stop sequences
}

// ollamaResponse represents a complete response from Ollama
type ollamaResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Message   ollamaMessage `json:"message"`
	Done      bool      `json:"done"`
	Context   []int     `json:"context,omitempty"`

	// Usage statistics
	TotalDuration      int64 `json:"total_duration,omitempty"`
	LoadDuration       int64 `json:"load_duration,omitempty"`
	PromptEvalCount    int   `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64 `json:"prompt_eval_duration,omitempty"`
	EvalCount          int   `json:"eval_count,omitempty"`
	EvalDuration       int64 `json:"eval_duration,omitempty"`
}

// ollamaStreamResponse represents a streaming response chunk from Ollama
type ollamaStreamResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Message   ollamaMessage `json:"message,omitempty"`
	Done      bool      `json:"done"`

	// Present in final chunk
	Context            []int `json:"context,omitempty"`
	TotalDuration      int64 `json:"total_duration,omitempty"`
	LoadDuration       int64 `json:"load_duration,omitempty"`
	PromptEvalCount    int   `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64 `json:"prompt_eval_duration,omitempty"`
	EvalCount          int   `json:"eval_count,omitempty"`
	EvalDuration       int64 `json:"eval_duration,omitempty"`

	// Error information
	Error string `json:"error,omitempty"`
}

// ollamaModelInfo represents information about a model
type ollamaModelInfo struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Details    ollamaModelDetails `json:"details"`
}

// ollamaModelDetails contains detailed model information
type ollamaModelDetails struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families,omitempty"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// ollamaModelsResponse represents the response from /api/tags
type ollamaModelsResponse struct {
	Models []ollamaModelInfo `json:"models"`
}

// ollamaPullRequest represents a request to pull a model
type ollamaPullRequest struct {
	Name   string `json:"name"`
	Stream bool   `json:"stream"`
}

// ollamaPullResponse represents a response from pulling a model
type ollamaPullResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// ModelInfo represents public model information
type ModelInfo struct {
	Name          string
	Size          int64
	ParameterSize string
	Family        string
	ModifiedAt    time.Time
	Digest        string
}

// convertToModelInfo converts internal model info to public format
func convertToModelInfo(info ollamaModelInfo) ModelInfo {
	return ModelInfo{
		Name:          info.Name,
		Size:          info.Size,
		ParameterSize: info.Details.ParameterSize,
		Family:        info.Details.Family,
		ModifiedAt:    info.ModifiedAt,
		Digest:        info.Digest,
	}
}
