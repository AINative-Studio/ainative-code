# LLM Provider Interface

This document describes the provider interface that all LLM provider implementations must follow.

## Interface Definition

```go
package provider

import (
    "context"
    "time"
)

// Provider defines the interface all LLM providers must implement
type Provider interface {
    // Name returns the provider identifier (e.g., "anthropic", "openai")
    Name() string

    // Chat sends a message and returns a complete response
    Chat(ctx context.Context, messages []Message, opts ...Option) (*Response, error)

    // Stream sends a message and returns a channel of streaming events
    Stream(ctx context.Context, messages []Message, opts ...Option) (<-chan StreamEvent, error)

    // ValidateConfig checks if the provider configuration is valid
    ValidateConfig(config *Config) error

    // Models returns the list of available models for this provider
    Models() []Model

    // MaxTokens returns the maximum token limit for a given model
    MaxTokens(model string) int
}

// Config holds provider-specific configuration
type Config struct {
    APIKey      string
    Model       string
    MaxTokens   int
    Temperature float64
    TopP        float64
    Endpoint    string // Optional custom endpoint
    Timeout     time.Duration
    RetryConfig *RetryConfig
}

// RetryConfig defines retry behavior
type RetryConfig struct {
    MaxRetries     int
    InitialBackoff time.Duration
    MaxBackoff     time.Duration
    Multiplier     float64
}
```

## Core Types

### Message

```go
// Message represents a chat message
type Message struct {
    Role    Role        // user, assistant, system
    Content interface{} // string or []ContentBlock
    Name    string      // Optional name for the message author
}

// Role defines message roles
type Role string

const (
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
    RoleSystem    Role = "system"
)

// ContentBlock represents structured content
type ContentBlock struct {
    Type string      // text, image, tool_use, tool_result
    Data interface{} // Type-specific data
}

// TextBlock represents text content
type TextBlock struct {
    Text string
}

// ImageBlock represents image content
type ImageBlock struct {
    Source ImageSource
}

// ImageSource defines image data source
type ImageSource struct {
    Type  string // base64, url
    Media string // MIME type (e.g., "image/png")
    Data  string // Base64 data or URL
}

// ToolUseBlock represents a tool call
type ToolUseBlock struct {
    ID    string
    Name  string
    Input map[string]interface{}
}

// ToolResultBlock represents tool execution result
type ToolResultBlock struct {
    ToolUseID string
    Content   string
    IsError   bool
}
```

### Response

```go
// Response represents a complete LLM response
type Response struct {
    ID         string
    Content    string
    ToolCalls  []ToolCall
    Usage      Usage
    StopReason StopReason
    Model      string
}

// ToolCall represents a request to execute a tool
type ToolCall struct {
    ID        string
    Name      string
    Arguments map[string]interface{}
}

// Usage represents token usage statistics
type Usage struct {
    InputTokens       int
    OutputTokens      int
    CacheReadTokens   int // For prompt caching
    CacheWriteTokens  int // For prompt caching
}

// StopReason indicates why the response ended
type StopReason string

const (
    StopReasonEndTurn      StopReason = "end_turn"
    StopReasonMaxTokens    StopReason = "max_tokens"
    StopReasonToolUse      StopReason = "tool_use"
    StopReasonStopSequence StopReason = "stop_sequence"
    StopReasonError        StopReason = "error"
)
```

### StreamEvent

```go
// StreamEvent represents a streaming chunk
type StreamEvent struct {
    Type       EventType
    Delta      string          // Text delta for streaming
    Content    string          // Complete content so far
    ToolCall   *ToolCall       // Tool call in progress
    Usage      *Usage          // Final usage stats
    StopReason *StopReason     // Reason for stopping
    Error      error           // Error if any
}

// EventType defines streaming event types
type EventType string

const (
    EventTypeStart       EventType = "start"
    EventTypeContentDelta EventType = "content_delta"
    EventTypeToolCall    EventType = "tool_call"
    EventTypeUsage       EventType = "usage"
    EventTypeDone        EventType = "done"
    EventTypeError       EventType = "error"
)
```

### Options

```go
// Option is a functional option for provider calls
type Option func(*CallOptions)

// CallOptions holds options for Chat/Stream calls
type CallOptions struct {
    Model          string
    MaxTokens      int
    Temperature    float64
    TopP           float64
    StopSequences  []string
    ToolChoice     string // auto, required, none, or specific tool name
    Stream         bool
    SystemPrompt   string
    Metadata       map[string]string
    CacheControl   bool // Enable prompt caching
}

// WithModel sets the model to use
func WithModel(model string) Option {
    return func(o *CallOptions) {
        o.Model = model
    }
}

// WithMaxTokens sets the maximum tokens
func WithMaxTokens(n int) Option {
    return func(o *CallOptions) {
        o.MaxTokens = n
    }
}

// WithTemperature sets the temperature
func WithTemperature(t float64) Option {
    return func(o *CallOptions) {
        o.Temperature = t
    }
}

// WithToolChoice sets tool calling behavior
func WithToolChoice(choice string) Option {
    return func(o *CallOptions) {
        o.ToolChoice = choice
    }
}

// WithCacheControl enables prompt caching
func WithCacheControl() Option {
    return func(o *CallOptions) {
        o.CacheControl = true
    }
}
```

## Model Information

```go
// Model represents information about an available model
type Model struct {
    ID              string
    Name            string
    Description     string
    MaxTokens       int
    InputPricing    float64 // Per million tokens
    OutputPricing   float64 // Per million tokens
    Features        []Feature
    ContextWindow   int
    TrainingCutoff  string
}

// Feature represents a model capability
type Feature string

const (
    FeatureVision          Feature = "vision"
    FeatureToolUse         Feature = "tool_use"
    FeatureStreaming       Feature = "streaming"
    FeaturePromptCaching   Feature = "prompt_caching"
    FeatureFunctionCalling Feature = "function_calling"
    FeatureExtendedThinking Feature = "extended_thinking"
)
```

## Implementation Example

Here's an example implementation for a hypothetical provider:

```go
package example

import (
    "context"
    "fmt"

    "github.com/AINative-studio/ainative-code/internal/provider"
)

type ExampleProvider struct {
    config *provider.Config
    client *HTTPClient
}

// NewExampleProvider creates a new provider instance
func NewExampleProvider(config *provider.Config) (*ExampleProvider, error) {
    if err := validateConfig(config); err != nil {
        return nil, err
    }

    return &ExampleProvider{
        config: config,
        client: NewHTTPClient(config.APIKey, config.Endpoint),
    }, nil
}

// Name returns the provider identifier
func (p *ExampleProvider) Name() string {
    return "example"
}

// Chat sends a message and returns a complete response
func (p *ExampleProvider) Chat(ctx context.Context, messages []provider.Message, opts ...provider.Option) (*provider.Response, error) {
    // Apply options
    options := p.applyOptions(opts)

    // Format messages for this provider's API
    apiMessages := p.formatMessages(messages)

    // Create request
    req := &APIRequest{
        Model:       options.Model,
        Messages:    apiMessages,
        MaxTokens:   options.MaxTokens,
        Temperature: options.Temperature,
    }

    // Make API call
    resp, err := p.client.Chat(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("chat request failed: %w", err)
    }

    // Convert to common response format
    return p.convertResponse(resp), nil
}

// Stream sends a message and returns a channel of streaming events
func (p *ExampleProvider) Stream(ctx context.Context, messages []provider.Message, opts ...provider.Option) (<-chan provider.StreamEvent, error) {
    events := make(chan provider.StreamEvent)

    go func() {
        defer close(events)

        // Apply options
        options := p.applyOptions(opts)
        options.Stream = true

        // Format messages
        apiMessages := p.formatMessages(messages)

        // Create streaming request
        req := &APIRequest{
            Model:       options.Model,
            Messages:    apiMessages,
            MaxTokens:   options.MaxTokens,
            Temperature: options.Temperature,
            Stream:      true,
        }

        // Stream response
        eventStream, err := p.client.StreamChat(ctx, req)
        if err != nil {
            events <- provider.StreamEvent{
                Type:  provider.EventTypeError,
                Error: err,
            }
            return
        }

        // Process stream
        for apiEvent := range eventStream {
            event := p.convertStreamEvent(apiEvent)

            select {
            case events <- event:
            case <-ctx.Done():
                return
            }
        }
    }()

    return events, nil
}

// ValidateConfig checks if the provider configuration is valid
func (p *ExampleProvider) ValidateConfig(config *provider.Config) error {
    if config.APIKey == "" {
        return fmt.Errorf("API key is required")
    }

    if config.Model == "" {
        return fmt.Errorf("model is required")
    }

    if config.MaxTokens <= 0 || config.MaxTokens > 100000 {
        return fmt.Errorf("max_tokens must be between 1 and 100000")
    }

    if config.Temperature < 0 || config.Temperature > 2 {
        return fmt.Errorf("temperature must be between 0 and 2")
    }

    return nil
}

// Models returns the list of available models
func (p *ExampleProvider) Models() []provider.Model {
    return []provider.Model{
        {
            ID:            "example-pro",
            Name:          "Example Pro",
            Description:   "Most capable model",
            MaxTokens:     100000,
            InputPricing:  3.0,
            OutputPricing: 15.0,
            Features: []provider.Feature{
                provider.FeatureVision,
                provider.FeatureToolUse,
                provider.FeatureStreaming,
            },
            ContextWindow:  200000,
            TrainingCutoff: "2024-01",
        },
    }
}

// MaxTokens returns the maximum token limit for a model
func (p *ExampleProvider) MaxTokens(model string) int {
    for _, m := range p.Models() {
        if m.ID == model {
            return m.MaxTokens
        }
    }
    return 4096 // Default
}

// Helper methods

func (p *ExampleProvider) applyOptions(opts []provider.Option) *provider.CallOptions {
    options := &provider.CallOptions{
        Model:       p.config.Model,
        MaxTokens:   p.config.MaxTokens,
        Temperature: p.config.Temperature,
    }

    for _, opt := range opts {
        opt(options)
    }

    return options
}

func (p *ExampleProvider) formatMessages(messages []provider.Message) []APIMessage {
    // Convert common message format to provider-specific format
    var apiMessages []APIMessage

    for _, msg := range messages {
        apiMsg := APIMessage{
            Role: string(msg.Role),
        }

        // Handle different content types
        switch content := msg.Content.(type) {
        case string:
            apiMsg.Content = content
        case []provider.ContentBlock:
            apiMsg.Content = p.formatContentBlocks(content)
        }

        apiMessages = append(apiMessages, apiMsg)
    }

    return apiMessages
}

func (p *ExampleProvider) convertResponse(apiResp *APIResponse) *provider.Response {
    return &provider.Response{
        ID:      apiResp.ID,
        Content: apiResp.Content,
        Usage: provider.Usage{
            InputTokens:  apiResp.Usage.InputTokens,
            OutputTokens: apiResp.Usage.OutputTokens,
        },
        StopReason: provider.StopReason(apiResp.StopReason),
        Model:      apiResp.Model,
    }
}

func (p *ExampleProvider) convertStreamEvent(apiEvent *APIStreamEvent) provider.StreamEvent {
    switch apiEvent.Type {
    case "content_delta":
        return provider.StreamEvent{
            Type:  provider.EventTypeContentDelta,
            Delta: apiEvent.Delta,
        }
    case "done":
        return provider.StreamEvent{
            Type: provider.EventTypeDone,
            Usage: &provider.Usage{
                InputTokens:  apiEvent.Usage.InputTokens,
                OutputTokens: apiEvent.Usage.OutputTokens,
            },
        }
    default:
        return provider.StreamEvent{
            Type: provider.EventTypeError,
            Error: fmt.Errorf("unknown event type: %s", apiEvent.Type),
        }
    }
}
```

## Error Handling

Providers should return standardized errors:

```go
package provider

import "errors"

var (
    // ErrInvalidAPIKey indicates the API key is invalid
    ErrInvalidAPIKey = errors.New("invalid API key")

    // ErrRateLimited indicates rate limit exceeded
    ErrRateLimited = errors.New("rate limit exceeded")

    // ErrModelNotFound indicates the model doesn't exist
    ErrModelNotFound = errors.New("model not found")

    // ErrInvalidRequest indicates malformed request
    ErrInvalidRequest = errors.New("invalid request")

    // ErrServerError indicates server-side error
    ErrServerError = errors.New("server error")

    // ErrTimeout indicates request timeout
    ErrTimeout = errors.New("request timeout")

    // ErrContentFiltered indicates content was filtered
    ErrContentFiltered = errors.New("content filtered")
)

// ProviderError wraps provider-specific errors
type ProviderError struct {
    Provider string
    Type     string
    Message  string
    Code     int
    Retryable bool
    Err      error
}

func (e *ProviderError) Error() string {
    return fmt.Sprintf("%s provider error: %s (code: %d)", e.Provider, e.Message, e.Code)
}

func (e *ProviderError) Unwrap() error {
    return e.Err
}
```

## Testing

All provider implementations should include comprehensive tests:

```go
func TestProviderChat(t *testing.T) {
    provider := setupTestProvider(t)

    messages := []provider.Message{
        {
            Role:    provider.RoleUser,
            Content: "Hello, world!",
        },
    }

    resp, err := provider.Chat(context.Background(), messages)
    assert.NoError(t, err)
    assert.NotEmpty(t, resp.Content)
    assert.Greater(t, resp.Usage.OutputTokens, 0)
}

func TestProviderStream(t *testing.T) {
    provider := setupTestProvider(t)

    messages := []provider.Message{
        {
            Role:    provider.RoleUser,
            Content: "Count to 5",
        },
    }

    events, err := provider.Stream(context.Background(), messages)
    assert.NoError(t, err)

    var content strings.Builder
    for event := range events {
        if event.Type == provider.EventTypeContentDelta {
            content.WriteString(event.Delta)
        }
    }

    assert.NotEmpty(t, content.String())
}
```

## References

- [System Overview](../architecture/system-overview.md)
- [Component Design](../architecture/component-design.md)
- [Provider Implementations](../providers/README.md)

---

**Document Version**: 1.0
**Last Updated**: January 2025
