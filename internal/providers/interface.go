package providers

import (
	"context"
	"io"
)

// Provider defines the unified interface for all LLM providers
type Provider interface {
	// Chat sends a chat request and returns a complete response
	Chat(ctx context.Context, req *ChatRequest, opts ...Option) (*Response, error)

	// Stream sends a streaming chat request and returns a channel of events
	Stream(ctx context.Context, req *StreamRequest, opts ...Option) (<-chan Event, error)

	// Name returns the provider's name (e.g., "anthropic", "openai")
	Name() string

	// Models returns the list of available models for this provider
	Models(ctx context.Context) ([]Model, error)

	// Close closes the provider and releases any resources
	io.Closer
}

// Config holds common provider configuration
type Config struct {
	APIKey      string
	BaseURL     string
	MaxRetries  int
	Timeout     int // in seconds
	DefaultModel string
	Metadata    map[string]interface{}
}

// ProviderFactory is a function that creates a new provider instance
type ProviderFactory func(config Config) (Provider, error)
