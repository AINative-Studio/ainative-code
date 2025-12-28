package provider

import "context"

// Provider defines the interface for LLM providers
type Provider interface {
	// Chat sends a complete chat request and waits for the full response
	Chat(ctx context.Context, messages []Message, opts ...ChatOption) (Response, error)

	// Stream sends a streaming chat request and returns a channel for events
	Stream(ctx context.Context, messages []Message, opts ...StreamOption) (<-chan Event, error)

	// Name returns the provider's name
	Name() string

	// Models returns the list of supported model identifiers
	Models() []string

	// Close releases any resources held by the provider
	Close() error
}

// Message represents a chat message
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

// Response represents a complete chat response
type Response struct {
	Content string
	Usage   Usage
	Model   string
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Event represents a streaming event
type Event struct {
	Type    EventType
	Content string
	Error   error
	Done    bool
}

// EventType represents the type of streaming event
type EventType int

const (
	EventTypeContentDelta EventType = iota // Incremental content
	EventTypeContentStart                   // Stream started
	EventTypeContentEnd                     // Stream completed
	EventTypeError                          // Error occurred
)

// String returns the string representation of EventType
func (e EventType) String() string {
	switch e {
	case EventTypeContentDelta:
		return "ContentDelta"
	case EventTypeContentStart:
		return "ContentStart"
	case EventTypeContentEnd:
		return "ContentEnd"
	case EventTypeError:
		return "Error"
	default:
		return "Unknown"
	}
}
