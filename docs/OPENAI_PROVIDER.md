# OpenAI Provider Documentation

## Overview

The OpenAI provider implements the `Provider` interface to enable chat completions using OpenAI's GPT models. It is designed as ONE provider in a multi-provider architecture alongside Anthropic, Google, and other LLM providers.

## Features

- Full support for GPT-4 and GPT-3.5 Turbo models
- Streaming and non-streaming chat completions
- System message support
- Advanced configuration options (temperature, top_p, stop sequences)
- Comprehensive error handling with retries
- Organization support for OpenAI teams
- Context-aware request cancellation
- Proper resource management

## Installation

The OpenAI provider is part of the AINative-Code internal package structure:

```go
import (
    "github.com/AINative-studio/ainative-code/internal/provider"
    "github.com/AINative-studio/ainative-code/internal/provider/openai"
)
```

## Supported Models

The following OpenAI models are supported:

### GPT-4 Turbo
- `gpt-4-turbo-preview` - Latest GPT-4 Turbo preview
- `gpt-4-0125-preview` - GPT-4 Turbo (January 25, 2024)
- `gpt-4-1106-preview` - GPT-4 Turbo (November 6, 2023)

### GPT-4
- `gpt-4` - Standard GPT-4 (8K context)
- `gpt-4-0613` - GPT-4 snapshot from June 13, 2023
- `gpt-4-32k` - GPT-4 with 32K context window
- `gpt-4-32k-0613` - GPT-4 32K snapshot

### GPT-3.5 Turbo
- `gpt-3.5-turbo` - Latest GPT-3.5 Turbo
- `gpt-3.5-turbo-0125` - GPT-3.5 Turbo (January 25, 2024)
- `gpt-3.5-turbo-1106` - GPT-3.5 Turbo (November 6, 2023)
- `gpt-3.5-turbo-16k` - GPT-3.5 Turbo with 16K context
- `gpt-3.5-turbo-16k-0613` - GPT-3.5 Turbo 16K snapshot

## Quick Start

### Basic Chat Completion

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AINative-studio/ainative-code/internal/provider"
    "github.com/AINative-studio/ainative-code/internal/provider/openai"
)

func main() {
    // Create provider
    p, err := openai.NewOpenAIProvider(openai.Config{
        APIKey: "your-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer p.Close()

    // Prepare messages
    messages := []provider.Message{
        {Role: "system", Content: "You are a helpful assistant."},
        {Role: "user", Content: "Hello!"},
    }

    // Send request
    ctx := context.Background()
    resp, err := p.Chat(ctx, messages,
        provider.WithModel("gpt-4"),
        provider.WithMaxTokens(100),
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Content)
}
```

### Streaming Chat

```go
// Stream responses for real-time output
eventChan, err := p.Stream(ctx, messages,
    provider.StreamWithModel("gpt-3.5-turbo"),
)
if err != nil {
    log.Fatal(err)
}

for event := range eventChan {
    switch event.Type {
    case provider.EventTypeContentDelta:
        fmt.Print(event.Content)
    case provider.EventTypeError:
        log.Printf("Error: %v", event.Error)
    }
}
```

## Configuration

### Basic Configuration

```go
config := openai.Config{
    APIKey: "sk-...",  // Required
}
```

### Advanced Configuration

```go
config := openai.Config{
    APIKey:       "sk-...",                          // Required
    BaseURL:      "https://api.openai.com/v1",      // Optional: Custom endpoint
    Organization: "org-...",                         // Optional: Organization ID
    HTTPClient:   customHTTPClient,                  // Optional: Custom HTTP client
    Logger:       customLogger,                      // Optional: Custom logger
}
```

## Chat Options

### Available Options

```go
// Model selection
provider.WithModel("gpt-4")

// Token limits
provider.WithMaxTokens(1000)

// Sampling parameters
provider.WithTemperature(0.7)   // 0.0 to 1.0
provider.WithTopP(0.9)           // 0.0 to 1.0

// Stop sequences
provider.WithStopSequences("\n\n", "END")

// System prompt
provider.WithSystemPrompt("You are a helpful assistant")
```

### Example with Multiple Options

```go
resp, err := p.Chat(ctx, messages,
    provider.WithModel("gpt-4-turbo-preview"),
    provider.WithMaxTokens(500),
    provider.WithTemperature(0.8),
    provider.WithTopP(0.95),
    provider.WithStopSequences("\n---\n"),
    provider.WithSystemPrompt("You are an expert programmer."),
)
```

## Multi-Provider Architecture

The OpenAI provider is designed to work alongside other providers:

```go
// Create registry
registry := provider.NewRegistry()

// Register multiple providers
openaiProvider, _ := openai.NewOpenAIProvider(openaiConfig)
registry.Register("openai", openaiProvider)

anthropicProvider, _ := anthropic.NewAnthropicProvider(anthropicConfig)
registry.Register("anthropic", anthropicProvider)

// Use providers by name
p, _ := registry.Get("openai")
resp, _ := p.Chat(ctx, messages, provider.WithModel("gpt-4"))
```

## Error Handling

### Error Types

The provider returns specific error types:

- **AuthenticationError**: Invalid API key or permissions
- **RateLimitError**: Rate limit exceeded
- **ContextLengthError**: Input exceeds model's context window
- **InvalidModelError**: Model not supported
- **ProviderError**: General provider errors

### Example Error Handling

```go
resp, err := p.Chat(ctx, messages, provider.WithModel("gpt-4"))
if err != nil {
    switch {
    case strings.Contains(err.Error(), "authentication"):
        log.Println("Check your API key")
    case strings.Contains(err.Error(), "rate limit"):
        log.Println("Rate limited, retry after delay")
    case strings.Contains(err.Error(), "context length"):
        log.Println("Reduce message length")
    default:
        log.Printf("Error: %v", err)
    }
    return
}
```

## Important Notes

### Embeddings

**CRITICAL**: For vector embeddings, use the AINative platform API, NOT OpenAI's embeddings endpoint:

```go
import "github.com/AINative-studio/ainative-code/internal/embeddings"

client, _ := embeddings.NewAINativeEmbeddingsClient(embeddings.Config{
    APIKey: "your-ainative-api-key",
})

result, _ := client.Embed(ctx, []string{"text to embed"}, "default")
```

See `EMBEDDINGS.md` for complete embeddings documentation.

### Best Practices

1. **Always close providers**: Use `defer p.Close()` to release resources
2. **Use context**: Pass context for cancellation and timeouts
3. **Handle errors**: Implement proper error handling with retries
4. **Choose appropriate models**: Balance cost, speed, and quality
5. **Set token limits**: Prevent unexpected costs with `WithMaxTokens()`
6. **Use streaming for long responses**: Better UX with `Stream()`

### Performance Considerations

- Use `gpt-3.5-turbo` for faster, cheaper responses
- Use `gpt-4` for complex reasoning tasks
- Enable streaming for long-running requests
- Implement rate limiting on your side to avoid API limits
- Reuse provider instances instead of creating new ones

## API Compatibility

The OpenAI provider implements the standard `Provider` interface:

```go
type Provider interface {
    Chat(ctx context.Context, messages []Message, opts ...ChatOption) (Response, error)
    Stream(ctx context.Context, messages []Message, opts ...StreamOption) (<-chan Event, error)
    Name() string
    Models() []string
    Close() error
}
```

This ensures compatibility with any code that uses the `Provider` interface, allowing seamless switching between OpenAI, Anthropic, Google, and other providers.

## Testing

The provider includes comprehensive tests:

- **Unit tests**: `internal/provider/openai/*_test.go`
- **Integration tests**: `tests/integration/openai_test.go`
- **Coverage**: 83.9%+ code coverage

Run tests:

```bash
# Unit tests
go test ./internal/provider/openai/...

# Integration tests
go test ./tests/integration/openai_test.go

# With coverage
go test -cover ./internal/provider/openai/...
```

## Examples

See `examples/openai_provider_example.go` for complete working examples including:

- Basic chat completions
- Streaming responses
- Multi-provider setup
- Advanced configuration
- Error handling

## Support

For issues or questions:

- Check the API documentation
- Review the examples
- Examine the unit tests for usage patterns
- Consult the Provider interface documentation
