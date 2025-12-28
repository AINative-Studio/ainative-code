# LLM Provider Interface

A unified, extensible Go interface for integrating multiple Large Language Model (LLM) providers including Anthropic, OpenAI, Google Gemini, AWS Bedrock, Azure OpenAI, and Ollama.

## Architecture Overview

The provider interface implements several Go design patterns to create a clean, maintainable, and type-safe abstraction:

- **Provider Interface Pattern**: Unified contract for all LLM providers
- **Functional Options Pattern**: Flexible request configuration
- **Factory Pattern**: Dynamic provider instantiation
- **Registry Pattern**: Centralized provider management
- **Event Streaming**: Real-time response handling via Go channels

## Core Components

### 1. Provider Interface

The `Provider` interface defines the contract that all LLM provider implementations must satisfy:

```go
type Provider interface {
    // Chat sends a single request and returns the complete response
    Chat(ctx context.Context, req *ChatRequest, opts ...Option) (*Response, error)

    // Stream sends a request and returns a channel for streaming events
    Stream(ctx context.Context, req *StreamRequest, opts ...Option) (<-chan Event, error)

    // Name returns the provider's identifier
    Name() string

    // Models returns available models for this provider
    Models(ctx context.Context) ([]Model, error)

    // Close releases provider resources
    Close() error
}
```

### 2. Type System

#### Messages and Roles

```go
const (
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
    RoleSystem    Role = "system"
)

type Message struct {
    Role    Role
    Content string
}
```

#### Request Types

```go
type ChatRequest struct {
    Messages      []Message
    Model         string
    MaxTokens     int
    Temperature   float64
    TopP          float64
    StopSequences []string
    Metadata      map[string]interface{}
}

type StreamRequest struct {
    Messages      []Message
    Model         string
    MaxTokens     int
    Temperature   float64
    TopP          float64
    StopSequences []string
    Metadata      map[string]interface{}
}
```

#### Response Types

```go
type Response struct {
    Content      string
    Model        string
    Provider     string
    FinishReason string
    Usage        *UsageInfo
    Metadata     map[string]interface{}
    CreatedAt    time.Time
}

type UsageInfo struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}
```

#### Event System

```go
const (
    EventTextDelta     EventType = "text_delta"
    EventContentStart  EventType = "content_start"
    EventContentEnd    EventType = "content_end"
    EventMessageStart  EventType = "message_start"
    EventMessageStop   EventType = "message_stop"
    EventError         EventType = "error"
    EventUsage         EventType = "usage"
    EventThinking      EventType = "thinking"
)

type Event struct {
    Type      EventType
    Data      interface{}
    Usage     *UsageInfo
    Timestamp time.Time
}
```

## Usage Examples

### Example 1: Basic Chat Request

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AINative-studio/ainative-code/internal/providers"
)

func main() {
    // Get the global registry
    registry := providers.GetGlobalRegistry()

    // Create a provider instance (assumes factory is registered)
    config := providers.Config{
        APIKey:  "your-api-key",
        BaseURL: "https://api.anthropic.com",
    }

    provider, err := registry.Create("anthropic", config)
    if err != nil {
        log.Fatalf("Failed to create provider: %v", err)
    }
    defer provider.Close()

    // Create a chat request
    req := &providers.ChatRequest{
        Messages: []providers.Message{
            {Role: providers.RoleUser, Content: "What is the capital of France?"},
        },
        Model: "claude-3-sonnet-20240229",
    }

    // Send the request
    ctx := context.Background()
    resp, err := provider.Chat(ctx, req)
    if err != nil {
        log.Fatalf("Chat request failed: %v", err)
    }

    // Display the response
    fmt.Printf("Response: %s\n", resp.Content)
    fmt.Printf("Model: %s\n", resp.Model)
    fmt.Printf("Tokens used: %d\n", resp.Usage.TotalTokens)
}
```

### Example 2: Using Functional Options

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AINative-studio/ainative-code/internal/providers"
)

func main() {
    registry := providers.GetGlobalRegistry()
    provider, _ := registry.Get("anthropic")

    req := &providers.ChatRequest{
        Messages: []providers.Message{
            {Role: providers.RoleSystem, Content: "You are a helpful assistant."},
            {Role: providers.RoleUser, Content: "Write a haiku about coding."},
        },
        Model: "claude-3-sonnet-20240229",
    }

    // Apply functional options for fine-grained control
    providers.ApplyOptions(req,
        providers.WithMaxTokens(1024),
        providers.WithTemperature(0.7),
        providers.WithTopP(0.9),
        providers.WithStopSequences("\n\n"),
        providers.WithMetadata("session_id", "abc-123"),
        providers.WithMetadata("user_id", 42),
    )

    ctx := context.Background()
    resp, err := provider.Chat(ctx, req)
    if err != nil {
        log.Fatalf("Chat failed: %v", err)
    }

    fmt.Printf("Haiku:\n%s\n", resp.Content)
}
```

### Example 3: Streaming Responses

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AINative-studio/ainative-code/internal/providers"
)

func main() {
    registry := providers.GetGlobalRegistry()
    provider, _ := registry.Get("anthropic")

    req := &providers.StreamRequest{
        Messages: []providers.Message{
            {Role: providers.RoleUser, Content: "Tell me a story about a brave knight."},
        },
        Model: "claude-3-sonnet-20240229",
    }

    providers.ApplyStreamOptions(req,
        providers.WithMaxTokens(2048),
        providers.WithTemperature(0.8),
    )

    ctx := context.Background()
    eventChan, err := provider.Stream(ctx, req)
    if err != nil {
        log.Fatalf("Stream failed: %v", err)
    }

    // Process streaming events
    var fullContent string
    for event := range eventChan {
        switch event.Type {
        case providers.EventMessageStart:
            fmt.Println("Stream started...")

        case providers.EventTextDelta:
            delta := event.Data.(string)
            fmt.Print(delta)
            fullContent += delta

        case providers.EventUsage:
            usage := event.Usage
            fmt.Printf("\nTokens used: %d\n", usage.TotalTokens)

        case providers.EventError:
            fmt.Printf("Error: %v\n", event.Data)

        case providers.EventMessageStop:
            fmt.Println("\nStream complete")
        }
    }
}
```

### Example 4: Provider Registry Management

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AINative-studio/ainative-code/internal/providers"
)

// Custom provider factory function
func createOpenAIProvider(config providers.Config) (providers.Provider, error) {
    // Implementation would create OpenAI client
    // This is a placeholder example
    return &CustomOpenAIProvider{
        apiKey: config.APIKey,
        baseURL: config.BaseURL,
    }, nil
}

func main() {
    registry := providers.NewRegistry()

    // Register a provider factory
    err := registry.RegisterFactory("openai", createOpenAIProvider)
    if err != nil {
        log.Fatalf("Factory registration failed: %v", err)
    }

    // Create provider instances with different configurations
    config1 := providers.Config{
        APIKey:  "key-for-service-a",
        BaseURL: "https://api.openai.com/v1",
    }
    provider1, err := registry.Create("openai", config1)
    if err != nil {
        log.Fatalf("Provider creation failed: %v", err)
    }

    // List all registered providers
    providers := registry.List()
    fmt.Printf("Registered providers: %v\n", providers)

    // Retrieve a specific provider
    provider, err := registry.Get("openai")
    if err != nil {
        log.Fatalf("Provider not found: %v", err)
    }

    // Use the provider
    req := &providers.ChatRequest{
        Messages: []providers.Message{
            {Role: providers.RoleUser, Content: "Hello!"},
        },
        Model: "gpt-4",
    }

    resp, _ := provider.Chat(context.Background(), req)
    fmt.Printf("Response: %s\n", resp.Content)

    // Clean up - close all providers
    defer registry.Close()
}
```

### Example 5: Context-Based Cancellation and Timeout

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/AINative-studio/ainative-code/internal/providers"
)

func main() {
    registry := providers.GetGlobalRegistry()
    provider, _ := registry.Get("anthropic")

    req := &providers.ChatRequest{
        Messages: []providers.Message{
            {Role: providers.RoleUser, Content: "Explain quantum computing in detail."},
        },
        Model: "claude-3-sonnet-20240229",
    }

    // Set a 10-second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    resp, err := provider.Chat(ctx, req)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            log.Fatal("Request timed out after 10 seconds")
        }
        log.Fatalf("Request failed: %v", err)
    }

    fmt.Printf("Response: %s\n", resp.Content)
}

func streamWithCancellation() {
    registry := providers.GetGlobalRegistry()
    provider, _ := registry.Get("anthropic")

    req := &providers.StreamRequest{
        Messages: []providers.Message{
            {Role: providers.RoleUser, Content: "Write a long essay."},
        },
        Model: "claude-3-sonnet-20240229",
    }

    // Create cancellable context
    ctx, cancel := context.WithCancel(context.Background())

    // Cancel after 5 seconds
    go func() {
        time.Sleep(5 * time.Second)
        cancel()
    }()

    eventChan, err := provider.Stream(ctx, req)
    if err != nil {
        log.Fatalf("Stream failed: %v", err)
    }

    // Process events until cancelled
    for event := range eventChan {
        if event.Type == providers.EventTextDelta {
            fmt.Print(event.Data.(string))
        }
    }

    if ctx.Err() == context.Canceled {
        fmt.Println("\nStream cancelled by user")
    }
}
```

### Example 6: Implementing a Custom Provider

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/AINative-studio/ainative-code/internal/providers"
)

// CustomAnthropicProvider implements the Provider interface
type CustomAnthropicProvider struct {
    name    string
    apiKey  string
    baseURL string
    client  *http.Client
}

func (p *CustomAnthropicProvider) Chat(ctx context.Context, req *providers.ChatRequest, opts ...providers.Option) (*providers.Response, error) {
    // Apply any functional options
    providers.ApplyOptions(req, opts...)

    // Build API request (implementation specific)
    apiReq := buildAnthropicRequest(req)

    // Make HTTP call to Anthropic API
    apiResp, err := p.client.Do(apiReq.WithContext(ctx))
    if err != nil {
        return nil, fmt.Errorf("API request failed: %w", err)
    }
    defer apiResp.Body.Close()

    // Parse response and convert to unified format
    response := &providers.Response{
        Content:      parseContent(apiResp),
        Model:        req.Model,
        Provider:     p.name,
        FinishReason: parseFinishReason(apiResp),
        Usage: &providers.UsageInfo{
            PromptTokens:     parsePromptTokens(apiResp),
            CompletionTokens: parseCompletionTokens(apiResp),
            TotalTokens:      parseTotalTokens(apiResp),
        },
        CreatedAt: time.Now(),
    }

    return response, nil
}

func (p *CustomAnthropicProvider) Stream(ctx context.Context, req *providers.StreamRequest, opts ...providers.Option) (<-chan providers.Event, error) {
    providers.ApplyStreamOptions(req, opts...)

    // Create event channel
    eventChan := make(chan providers.Event, 100)

    // Start streaming in goroutine
    go func() {
        defer close(eventChan)

        // Send message start event
        eventChan <- providers.Event{
            Type:      providers.EventMessageStart,
            Timestamp: time.Now(),
        }

        // Build and execute streaming request
        apiReq := buildAnthropicStreamRequest(req)
        resp, err := p.client.Do(apiReq.WithContext(ctx))
        if err != nil {
            eventChan <- providers.Event{
                Type:      providers.EventError,
                Data:      err.Error(),
                Timestamp: time.Now(),
            }
            return
        }
        defer resp.Body.Close()

        // Parse SSE stream
        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            select {
            case <-ctx.Done():
                return
            default:
                line := scanner.Text()
                event := parseSSEEvent(line)

                // Convert to unified event format
                switch event.Type {
                case "content_block_delta":
                    eventChan <- providers.Event{
                        Type:      providers.EventTextDelta,
                        Data:      event.Delta.Text,
                        Timestamp: time.Now(),
                    }
                case "message_stop":
                    eventChan <- providers.Event{
                        Type:      providers.EventMessageStop,
                        Usage:     parseUsage(event),
                        Timestamp: time.Now(),
                    }
                }
            }
        }
    }()

    return eventChan, nil
}

func (p *CustomAnthropicProvider) Name() string {
    return p.name
}

func (p *CustomAnthropicProvider) Models(ctx context.Context) ([]providers.Model, error) {
    return []providers.Model{
        {
            ID:           "claude-3-opus-20240229",
            Name:         "Claude 3 Opus",
            Provider:     p.name,
            MaxTokens:    200000,
            Capabilities: []string{"chat", "streaming", "vision"},
        },
        {
            ID:           "claude-3-sonnet-20240229",
            Name:         "Claude 3 Sonnet",
            Provider:     p.name,
            MaxTokens:    200000,
            Capabilities: []string{"chat", "streaming", "vision"},
        },
    }, nil
}

func (p *CustomAnthropicProvider) Close() error {
    // Clean up resources (close connections, etc.)
    p.client.CloseIdleConnections()
    return nil
}

// Factory function for creating Anthropic provider instances
func NewAnthropicProvider(config providers.Config) (providers.Provider, error) {
    if config.APIKey == "" {
        return nil, fmt.Errorf("API key is required")
    }

    return &CustomAnthropicProvider{
        name:    "anthropic",
        apiKey:  config.APIKey,
        baseURL: config.BaseURL,
        client:  &http.Client{Timeout: 60 * time.Second},
    }, nil
}
```

## Best Practices

### 1. Always Use Context

Pass context to all Chat() and Stream() calls to enable cancellation and timeout control:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := provider.Chat(ctx, req)
```

### 2. Handle Streaming Events Properly

Always consume the entire event channel to prevent goroutine leaks:

```go
eventChan, err := provider.Stream(ctx, req)
if err != nil {
    return err
}

for event := range eventChan {
    // Process event
}
```

### 3. Close Providers When Done

Use defer to ensure providers are properly closed:

```go
provider, err := registry.Create("anthropic", config)
if err != nil {
    return err
}
defer provider.Close()
```

### 4. Use Functional Options for Flexibility

Leverage functional options to keep the API clean while supporting advanced configuration:

```go
// Reusable options
productionOpts := []providers.Option{
    providers.WithMaxTokens(4096),
    providers.WithTemperature(0.7),
    providers.WithMetadata("env", "production"),
}

providers.ApplyOptions(req, productionOpts...)
```

### 5. Implement Proper Error Handling

Check for context errors separately from provider errors:

```go
resp, err := provider.Chat(ctx, req)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        // Handle timeout
    } else if ctx.Err() == context.Canceled {
        // Handle cancellation
    } else {
        // Handle provider error
    }
}
```

### 6. Use Registry for Multi-Provider Applications

When supporting multiple providers, use the registry pattern:

```go
registry := providers.GetGlobalRegistry()

// Register multiple providers
registry.RegisterFactory("anthropic", NewAnthropicProvider)
registry.RegisterFactory("openai", NewOpenAIProvider)
registry.RegisterFactory("gemini", NewGeminiProvider)

// Use provider by name
provider, _ := registry.Get(userSelectedProvider)
resp, _ := provider.Chat(ctx, req)
```

## Thread Safety

The Registry implementation is thread-safe and can be accessed concurrently:

- `Get()`, `List()`, and `Create()` use read locks
- `Register()`, `RegisterFactory()`, `Unregister()`, and `Close()` use write locks
- Provider implementations should ensure their own thread-safety

## Testing

The package includes comprehensive test coverage (100%) with examples of:

- Unit tests for all types and constants
- Functional options pattern testing
- Registry operations and thread safety
- Mock provider implementations
- Error handling and edge cases

See `*_test.go` files for testing patterns and examples.

## Contributing

When implementing a new provider:

1. Implement all methods of the `Provider` interface
2. Handle context cancellation properly in both Chat() and Stream()
3. Convert provider-specific types to unified types
4. Map provider events to unified event types
5. Implement proper resource cleanup in Close()
6. Create a factory function that accepts `Config`
7. Write comprehensive unit tests
8. Document provider-specific configuration requirements

## License

See LICENSE file for details.
