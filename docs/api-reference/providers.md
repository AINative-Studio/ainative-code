# Providers API Reference

**Import Path**: `github.com/AINative-studio/ainative-code/internal/provider`

The provider package defines the interface for LLM providers and includes implementations for multiple AI services.

## Table of Contents

- [Provider Interface](#provider-interface)
- [Base Provider](#base-provider)
- [Message Types](#message-types)
- [Chat Options](#chat-options)
- [Streaming](#streaming)
- [Error Handling](#error-handling)
- [Provider Registry](#provider-registry)
- [Usage Examples](#usage-examples)

## Provider Interface

### Provider

```go
type Provider interface {
    Chat(ctx context.Context, messages []Message, opts ...ChatOption) (Response, error)
    Stream(ctx context.Context, messages []Message, opts ...StreamOption) (<-chan Event, error)
    Name() string
    Models() []string
    Close() error
}
```

The Provider interface defines the contract that all LLM providers must implement.

**Methods**:

- `Chat` - Sends a complete chat request and waits for the full response
- `Stream` - Sends a streaming chat request and returns a channel for events
- `Name` - Returns the provider's name (e.g., "anthropic", "openai")
- `Models` - Returns the list of supported model identifiers
- `Close` - Releases any resources held by the provider

## Message Types

### Message

```go
type Message struct {
    Role    string // "user", "assistant", "system"
    Content string
}
```

Represents a chat message in a conversation.

**Example**:

```go
messages := []provider.Message{
    {Role: "system", Content: "You are a helpful coding assistant."},
    {Role: "user", Content: "How do I implement OAuth in Go?"},
}
```

### Response

```go
type Response struct {
    Content string
    Usage   Usage
    Model   string
}
```

Represents a complete chat response from a provider.

### Usage

```go
type Usage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}
```

Token usage statistics for a request.

**Example**:

```go
fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n",
    response.Usage.TotalTokens,
    response.Usage.PromptTokens,
    response.Usage.CompletionTokens)
```

### Event

```go
type Event struct {
    Type    EventType
    Content string
    Error   error
    Done    bool
}
```

Represents a streaming event.

### EventType

```go
type EventType int

const (
    EventTypeContentDelta EventType = iota // Incremental content
    EventTypeContentStart                   // Stream started
    EventTypeContentEnd                     // Stream completed
    EventTypeError                          // Error occurred
)
```

## Chat Options

### ChatOption Functions

```go
func WithModel(model string) ChatOption
func WithMaxTokens(maxTokens int) ChatOption
func WithTemperature(temperature float64) ChatOption
func WithTopP(topP float64) ChatOption
func WithStopSequences(sequences ...string) ChatOption
func WithSystemPrompt(prompt string) ChatOption
func WithMetadata(key, value string) ChatOption
```

**Example**:

```go
response, err := provider.Chat(ctx, messages,
    provider.WithModel("claude-3-sonnet-20240229"),
    provider.WithTemperature(0.7),
    provider.WithMaxTokens(4096),
    provider.WithSystemPrompt("You are a helpful assistant."),
)
```

### StreamOption Functions

```go
func StreamWithModel(model string) StreamOption
func StreamWithMaxTokens(maxTokens int) StreamOption
func StreamWithTemperature(temperature float64) StreamOption
func StreamWithTopP(topP float64) StreamOption
func StreamWithStopSequences(sequences ...string) StreamOption
func StreamWithSystemPrompt(prompt string) StreamOption
func StreamWithMetadata(key, value string) StreamOption
```

### DefaultChatOptions

```go
func DefaultChatOptions() *ChatOptions
```

Returns ChatOptions with sensible defaults:
- MaxTokens: 1024
- Temperature: 0.7
- TopP: 1.0
- Stream: false

## Base Provider

The BaseProvider provides common functionality for all provider implementations.

### Types

#### BaseProvider

```go
type BaseProvider struct {
    // Contains filtered or unexported fields
}
```

#### RetryConfig

```go
type RetryConfig struct {
    MaxRetries           int
    InitialBackoff       time.Duration
    MaxBackoff           time.Duration
    Multiplier           float64
    RetryableStatusCodes []int
}
```

Configuration for retry behavior.

### Functions

#### NewBaseProvider

```go
func NewBaseProvider(config BaseProviderConfig) *BaseProvider
```

Creates a new BaseProvider with the given configuration.

**Example**:

```go
baseProvider := provider.NewBaseProvider(provider.BaseProviderConfig{
    Name:        "my-provider",
    HTTPClient:  httpClient,
    Logger:      logger,
    RetryConfig: provider.DefaultRetryConfig(),
})
```

#### DefaultRetryConfig

```go
func DefaultRetryConfig() RetryConfig
```

Returns sensible default retry configuration:
- MaxRetries: 3
- InitialBackoff: 1 second
- MaxBackoff: 30 seconds
- Multiplier: 2.0
- RetryableStatusCodes: [429, 500, 502, 503, 504]

### Methods

#### DoRequest

```go
func (b *BaseProvider) DoRequest(ctx context.Context, req *http.Request) (*http.Response, error)
```

Executes an HTTP request with retry logic and error handling.

Features:
- Exponential backoff with jitter
- Context-aware cancellation
- Automatic retry on transient errors
- Rate limit header parsing

**Example**:

```go
req, _ := http.NewRequest("POST", endpoint, body)
req.Header.Set("Content-Type", "application/json")

resp, err := baseProvider.DoRequest(ctx, req)
if err != nil {
    log.Fatalf("Request failed: %v", err)
}
defer resp.Body.Close()
```

## Streaming

### Using Stream Method

```go
func (p Provider) Stream(ctx context.Context, messages []Message, opts ...StreamOption) (<-chan Event, error)
```

**Example**:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AINative-studio/ainative-code/internal/provider"
)

func main() {
    ctx := context.Background()

    // Get provider
    p, _ := provider.GetRegistry().Get("anthropic")

    // Stream chat
    messages := []provider.Message{
        {Role: "user", Content: "Explain quantum computing"},
    }

    eventChan, err := p.Stream(ctx, messages,
        provider.StreamWithModel("claude-3-sonnet-20240229"),
        provider.StreamWithMaxTokens(2048),
    )
    if err != nil {
        log.Fatalf("Stream failed: %v", err)
    }

    // Process events
    var fullContent string
    for event := range eventChan {
        switch event.Type {
        case provider.EventTypeContentStart:
            fmt.Println("Stream started...")

        case provider.EventTypeContentDelta:
            fullContent += event.Content
            fmt.Print(event.Content) // Print incrementally

        case provider.EventTypeContentEnd:
            fmt.Println("\nStream completed")

        case provider.EventTypeError:
            log.Printf("Error: %v", event.Error)
        }

        if event.Done {
            break
        }
    }

    fmt.Printf("\nFull response: %s\n", fullContent)
}
```

## Error Handling

### Provider-Specific Errors

```go
// From internal/provider/errors.go
func NewAuthenticationError(providerName, message string) error
func NewRateLimitError(providerName string, retryAfter int) error
func NewProviderError(providerName, model string, err error) error
func NewInvalidModelError(providerName, model string, supportedModels []string) error
```

**Example**:

```go
import "github.com/AINative-studio/ainative-code/internal/errors"

response, err := provider.Chat(ctx, messages)
if err != nil {
    // Check error type
    if errors.GetCode(err) == errors.ErrCodeProviderRateLimit {
        log.Println("Rate limited, backing off...")
        time.Sleep(30 * time.Second)
        // Retry
    }

    // Check if retryable
    if errors.IsRetryable(err) {
        log.Println("Retryable error, attempting retry...")
    }

    return err
}
```

## Provider Registry

### Registry

The provider registry manages all registered providers and provides lookup functionality.

#### GetRegistry

```go
func GetRegistry() *Registry
```

Returns the global provider registry.

#### Register

```go
func (r *Registry) Register(provider Provider) error
```

Registers a new provider.

**Example**:

```go
// Register a provider
provider := &MyCustomProvider{}
err := provider.GetRegistry().Register(provider)
if err != nil {
    log.Fatalf("Failed to register provider: %v", err)
}
```

#### Get

```go
func (r *Registry) Get(name string) (Provider, error)
```

Retrieves a provider by name.

**Example**:

```go
// Get Anthropic provider
anthropic, err := provider.GetRegistry().Get("anthropic")
if err != nil {
    log.Fatalf("Provider not found: %v", err)
}

// Use the provider
response, err := anthropic.Chat(ctx, messages)
```

#### List

```go
func (r *Registry) List() []string
```

Returns a list of all registered provider names.

**Example**:

```go
providers := provider.GetRegistry().List()
fmt.Printf("Available providers: %v\n", providers)
```

## Usage Examples

### Basic Chat Request

```go
package main

import (
    "context"
    "log"

    "github.com/AINative-studio/ainative-code/internal/provider"
)

func main() {
    ctx := context.Background()

    // Get provider
    p, err := provider.GetRegistry().Get("anthropic")
    if err != nil {
        log.Fatalf("Failed to get provider: %v", err)
    }

    // Prepare messages
    messages := []provider.Message{
        {
            Role:    "system",
            Content: "You are a helpful coding assistant.",
        },
        {
            Role:    "user",
            Content: "Write a function to calculate fibonacci numbers in Go.",
        },
    }

    // Send chat request
    response, err := p.Chat(ctx, messages,
        provider.WithModel("claude-3-sonnet-20240229"),
        provider.WithMaxTokens(2048),
        provider.WithTemperature(0.7),
    )
    if err != nil {
        log.Fatalf("Chat failed: %v", err)
    }

    // Print response
    log.Printf("Response: %s", response.Content)
    log.Printf("Tokens used: %d", response.Usage.TotalTokens)
}
```

### Multi-Turn Conversation

```go
func multiTurnConversation(ctx context.Context) {
    p, _ := provider.GetRegistry().Get("anthropic")

    messages := []provider.Message{
        {Role: "system", Content: "You are a helpful assistant."},
    }

    // Turn 1
    messages = append(messages, provider.Message{
        Role:    "user",
        Content: "What is the capital of France?",
    })

    resp1, _ := p.Chat(ctx, messages, provider.WithModel("claude-3-sonnet-20240229"))
    messages = append(messages, provider.Message{
        Role:    "assistant",
        Content: resp1.Content,
    })

    // Turn 2
    messages = append(messages, provider.Message{
        Role:    "user",
        Content: "What is its population?",
    })

    resp2, _ := p.Chat(ctx, messages, provider.WithModel("claude-3-sonnet-20240229"))
    fmt.Printf("Response: %s\n", resp2.Content)
}
```

### Error Recovery

```go
func chatWithRetry(ctx context.Context, p provider.Provider, messages []provider.Message) (*provider.Response, error) {
    maxRetries := 3
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        response, err := p.Chat(ctx, messages,
            provider.WithModel("claude-3-sonnet-20240229"),
        )

        if err == nil {
            return &response, nil
        }

        lastErr = err

        // Check if retryable
        if !errors.IsRetryable(err) {
            return nil, err
        }

        // Exponential backoff
        backoff := time.Duration(1<<uint(attempt)) * time.Second
        log.Printf("Attempt %d failed, retrying in %v: %v", attempt+1, backoff, err)
        time.Sleep(backoff)
    }

    return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}
```

### Custom Provider Implementation

```go
package main

import (
    "context"
    "fmt"

    "github.com/AINative-studio/ainative-code/internal/provider"
)

type MyProvider struct {
    *provider.BaseProvider
    apiKey string
}

func NewMyProvider(apiKey string) *MyProvider {
    return &MyProvider{
        BaseProvider: provider.NewBaseProvider(provider.BaseProviderConfig{
            Name:        "my-provider",
            RetryConfig: provider.DefaultRetryConfig(),
        }),
        apiKey: apiKey,
    }
}

func (p *MyProvider) Chat(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error) {
    // Apply options
    options := provider.DefaultChatOptions()
    provider.ApplyChatOptions(options, opts...)

    // Implement chat logic...
    return provider.Response{
        Content: "Response from custom provider",
        Usage: provider.Usage{
            TotalTokens: 100,
        },
        Model: options.Model,
    }, nil
}

func (p *MyProvider) Stream(ctx context.Context, messages []provider.Message, opts ...provider.StreamOption) (<-chan provider.Event, error) {
    // Implement streaming logic...
    eventChan := make(chan provider.Event)
    go func() {
        defer close(eventChan)
        // Send events...
    }()
    return eventChan, nil
}

func (p *MyProvider) Models() []string {
    return []string{"my-model-1", "my-model-2"}
}
```

## Best Practices

### 1. Always Use Context

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := provider.Chat(ctx, messages)
```

### 2. Handle Streaming Properly

```go
eventChan, err := provider.Stream(ctx, messages)
if err != nil {
    return err
}

for event := range eventChan {
    if event.Type == provider.EventTypeError {
        log.Printf("Stream error: %v", event.Error)
    }
    if event.Done {
        break
    }
}
```

### 3. Implement Retry Logic

```go
for attempt := 0; attempt < maxRetries; attempt++ {
    response, err := provider.Chat(ctx, messages)
    if err == nil {
        return response, nil
    }
    if !errors.IsRetryable(err) {
        return nil, err
    }
    time.Sleep(calculateBackoff(attempt))
}
```

### 4. Monitor Token Usage

```go
response, err := provider.Chat(ctx, messages)
if err != nil {
    return err
}

log.Printf("Tokens - Prompt: %d, Completion: %d, Total: %d",
    response.Usage.PromptTokens,
    response.Usage.CompletionTokens,
    response.Usage.TotalTokens)

// Check against budget
if response.Usage.TotalTokens > tokenBudget {
    log.Warn("Token budget exceeded")
}
```

### 5. Use Provider Registry

```go
// Don't hardcode provider selection
// Bad:
// provider := anthropic.New(...)

// Good:
provider, err := provider.GetRegistry().Get(config.ProviderName)
if err != nil {
    return err
}
```

## Related Documentation

- [Configuration](configuration.md) - Provider configuration
- [Errors](errors.md) - Error handling
- [Core Packages](core-packages.md) - Client and session management
