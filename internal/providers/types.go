package providers

import "time"

// Role represents the role of a message sender
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

// Message represents a chat message
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Response represents a complete LLM response
type Response struct {
	Content      string        `json:"content"`
	Model        string        `json:"model"`
	Provider     string        `json:"provider"`
	FinishReason string        `json:"finish_reason,omitempty"`
	Usage        *UsageInfo    `json:"usage,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

// EventType represents the type of streaming event
type EventType string

const (
	EventTextDelta    EventType = "text_delta"
	EventContentStart EventType = "content_start"
	EventContentEnd   EventType = "content_end"
	EventMessageStart EventType = "message_start"
	EventMessageStop  EventType = "message_stop"
	EventError        EventType = "error"
	EventUsage        EventType = "usage"
	EventThinking     EventType = "thinking"
)

// Event represents a streaming event from the LLM
type Event struct {
	Type      EventType              `json:"type"`
	Data      string                 `json:"data,omitempty"`
	Error     error                  `json:"error,omitempty"`
	Usage     *UsageInfo             `json:"usage,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// UsageInfo contains token usage information
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Model represents an LLM model
type Model struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Provider    string   `json:"provider"`
	MaxTokens   int      `json:"max_tokens"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// ChatRequest represents a request to the Chat method
type ChatRequest struct {
	Messages      []Message              `json:"messages"`
	Model         string                 `json:"model"`
	MaxTokens     int                    `json:"max_tokens,omitempty"`
	Temperature   float64                `json:"temperature,omitempty"`
	TopP          float64                `json:"top_p,omitempty"`
	StopSequences []string               `json:"stop_sequences,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// StreamRequest represents a request to the Stream method
type StreamRequest struct {
	Messages      []Message              `json:"messages"`
	Model         string                 `json:"model"`
	MaxTokens     int                    `json:"max_tokens,omitempty"`
	Temperature   float64                `json:"temperature,omitempty"`
	TopP          float64                `json:"top_p,omitempty"`
	StopSequences []string               `json:"stop_sequences,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}
