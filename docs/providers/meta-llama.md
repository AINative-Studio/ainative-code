# Meta LLAMA Provider

Official provider for Meta's hosted LLAMA API, supporting the latest Llama 4 and Llama 3.3 models.

## Overview

The Meta LLAMA provider enables integration with Meta's official LLAMA API, providing access to state-of-the-art large language models with Mixture of Experts (MoE) architecture. The API is OpenAI-compatible, making it easy to use with existing code.

## Features

- ✅ OpenAI-compatible API format
- ✅ Chat completions (both sync and streaming)
- ✅ 4 LLAMA models supported (Llama 4 Maverick, Llama 4 Scout, Llama 3.3 70B/8B)
- ✅ Comprehensive error handling with retry support
- ✅ Token usage tracking
- ✅ Configurable parameters (temperature, top_p, max_tokens, etc.)
- ✅ 80%+ test coverage

## Supported Models

### Llama 4 Models (Mixture of Experts)

| Model ID | Parameters | Active Params | Architecture | Max Tokens | Recommended |
|----------|-----------|---------------|--------------|------------|-------------|
| `Llama-4-Maverick-17B-128E-Instruct-FP8` | 400B total | 17B active | 128 experts | 8192 | ✅ |
| `Llama-4-Scout-17B-16E` | 109B total | 17B active | 16 experts | 8192 | - |

### Llama 3.3 Models (Dense)

| Model ID | Parameters | Architecture | Max Tokens | Use Case |
|----------|-----------|--------------|------------|----------|
| `Llama-3.3-70B-Instruct` | 70B | Dense | 8192 | Large tasks |
| `Llama-3.3-8B-Instruct` | 8B | Dense | 8192 | Fast inference |

## Installation

The Meta provider is included in the AINative-Code project. No additional installation required.

## Configuration

### Environment Variables

```bash
# Required
META_API_KEY="LLM|<app_id>|<token>"

# Optional
META_BASE_URL="https://api.llama.com/compat/v1"  # Default
META_MODEL="Llama-4-Maverick-17B-128E-Instruct-FP8"  # Default
META_API_TIMEOUT="60"  # Seconds
```

### Programmatic Configuration

```go
import (
    "github.com/AINative-studio/ainative-code/internal/provider/meta"
    "time"
)

config := &meta.Config{
    APIKey:      os.Getenv("META_API_KEY"),
    BaseURL:     meta.DefaultBaseURL,
    Model:       meta.ModelLlama4Maverick,
    Temperature: 0.7,
    TopP:        0.9,
    MaxTokens:   2048,
    Timeout:     60 * time.Second,
}

provider, err := meta.NewMetaProvider(config)
if err != nil {
    log.Fatal(err)
}
defer provider.Close()
```

## Usage Examples

### Basic Chat Completion

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/AINative-studio/ainative-code/internal/provider"
    "github.com/AINative-studio/ainative-code/internal/provider/meta"
)

func main() {
    // Create provider
    config := &meta.Config{
        APIKey: os.Getenv("META_API_KEY"),
        Model:  meta.ModelLlama4Maverick,
    }

    provider, err := meta.NewMetaProvider(config)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Close()

    // Send chat request
    messages := []provider.Message{
        {
            Role:    "user",
            Content: "Explain what Llama 4 is in one sentence.",
        },
    }

    resp, err := provider.Chat(context.Background(), messages)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\n", resp.Content)
    fmt.Printf("Tokens: %d total (%d prompt + %d completion)\n",
        resp.Usage.TotalTokens,
        resp.Usage.PromptTokens,
        resp.Usage.CompletionTokens)
}
```

### Streaming Chat

```go
func streamExample() {
    provider, err := meta.NewMetaProvider(&meta.Config{
        APIKey: os.Getenv("META_API_KEY"),
        Model:  meta.ModelLlama4Maverick,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Close()

    messages := []provider.Message{
        {
            Role:    "user",
            Content: "Write a haiku about AI.",
        },
    }

    eventChan, err := provider.Stream(context.Background(), messages)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Print("Response: ")
    for event := range eventChan {
        switch event.Type {
        case provider.EventTypeContentDelta:
            fmt.Print(event.Content)
        case provider.EventTypeError:
            log.Fatal(event.Error)
        case provider.EventTypeContentEnd:
            fmt.Println() // New line at end
        }
    }
}
```

### System Prompts

```go
messages := []provider.Message{
    {
        Role:    "system",
        Content: "You are a helpful AI assistant that speaks like a pirate.",
    },
    {
        Role:    "user",
        Content: "Tell me about the ocean.",
    },
}

resp, err := provider.Chat(context.Background(), messages)
```

### Custom Parameters

```go
resp, err := provider.Chat(
    context.Background(),
    messages,
    provider.WithTemperature(0.5),      // More focused responses
    provider.WithMaxTokens(100),        // Limit response length
    provider.WithTopP(0.95),            // Nucleus sampling
    provider.WithStopSequences("END"),  // Stop at specific sequence
)
```

### Using Different Models

```go
// Fast model for quick responses
resp, err := provider.Chat(
    ctx,
    messages,
    provider.WithModel(meta.ModelLlama33_8B),
)

// Large model for complex tasks
resp, err := provider.Chat(
    ctx,
    messages,
    provider.WithModel(meta.ModelLlama33_70B),
)
```

## API Reference

### Configuration Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `APIKey` | string | **required** | Meta LLAMA API key |
| `BaseURL` | string | `https://api.llama.com/compat/v1` | API endpoint |
| `Model` | string | `Llama-4-Maverick-17B-128E-Instruct-FP8` | Model to use |
| `Temperature` | float64 | `0.7` | Sampling temperature (0-2) |
| `TopP` | float64 | `0.9` | Nucleus sampling (0-1) |
| `MaxTokens` | int | `2048` | Maximum tokens to generate |
| `Timeout` | duration | `60s` | Request timeout |
| `PresencePenalty` | float64 | `0.0` | Presence penalty (-2 to 2) |
| `FrequencyPenalty` | float64 | `0.0` | Frequency penalty (-2 to 2) |

### Error Handling

The provider returns `*MetaError` for API errors:

```go
resp, err := provider.Chat(ctx, messages)
if err != nil {
    if metaErr, ok := err.(*meta.MetaError); ok {
        if metaErr.IsAuthenticationError() {
            log.Printf("Invalid API key: %s", metaErr.Message)
        } else if metaErr.IsRateLimitError() {
            log.Printf("Rate limited, retry after delay")
        } else if metaErr.IsRetryable() {
            log.Printf("Temporary error, can retry: %s", metaErr.Message)
        }
    }
}
```

### Error Types

- `ErrTypeAuthentication` - Invalid or missing API key
- `ErrTypeRateLimit` - Rate limit exceeded (429)
- `ErrTypeInvalidRequest` - Bad request (400)
- `ErrTypeTimeout` - Request timeout
- `ErrTypeAPI` - Server error (5xx)

## Getting an API Key

1. Visit [Meta LLAMA Developer Portal](https://llama.developer.meta.com/)
2. Create an account or sign in
3. Navigate to API Keys section
4. Generate a new API key
5. Format: `LLM|<app_id>|<token>`

## Best Practices

### 1. Use Appropriate Models

- **Llama-4-Maverick**: Best overall performance, good for complex tasks
- **Llama-4-Scout**: Faster than Maverick, still very capable
- **Llama-3.3-70B**: Large dense model for specific use cases
- **Llama-3.3-8B**: Fast responses, lower cost

### 2. Implement Retry Logic

```go
func chatWithRetry(provider *meta.MetaProvider, messages []provider.Message) (provider.Response, error) {
    maxRetries := 3
    var resp provider.Response
    var err error

    for i := 0; i < maxRetries; i++ {
        resp, err = provider.Chat(context.Background(), messages)
        if err == nil {
            return resp, nil
        }

        if metaErr, ok := err.(*meta.MetaError); ok {
            if !metaErr.IsRetryable() {
                return resp, err // Don't retry auth errors, etc.
            }
        }

        time.Sleep(time.Second * time.Duration(i+1))
    }

    return resp, err
}
```

### 3. Handle Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := provider.Chat(ctx, messages)
```

### 4. Monitor Token Usage

```go
resp, err := provider.Chat(ctx, messages)
if err == nil {
    log.Printf("Tokens used: %d (prompt: %d, completion: %d)",
        resp.Usage.TotalTokens,
        resp.Usage.PromptTokens,
        resp.Usage.CompletionTokens)
}
```

## Performance

- **Chat Completion**: ~1-3 seconds for short responses
- **Streaming**: ~100-200ms for first token
- **Max Context**: 8192 tokens across all models
- **Rate Limits**: Varies by plan (check Meta's documentation)

## Testing

Run unit tests:
```bash
go test -v ./internal/provider/meta/...
```

Run integration tests (requires API key):
```bash
META_API_KEY="your-key" go test -v ./tests/integration/meta_test.go
```

## Troubleshooting

### "Invalid API key" Error

- Verify your API key is correct
- Ensure it follows the format: `LLM|<app_id>|<token>`
- Check that the API key hasn't expired

### "Rate limit exceeded" Error

- Implement exponential backoff retry logic
- Consider upgrading your Meta API plan
- Reduce request frequency

### Timeout Errors

- Increase the `Timeout` configuration
- Use shorter `MaxTokens` for faster responses
- Check network connectivity

### Model Not Found

- Verify you're using a supported model ID
- Check [Meta's API documentation](https://api.llama.com/compat/v1) for latest models
- Ensure model ID is spelled correctly

## Additional Resources

- [Meta LLAMA Documentation](https://llama.developer.meta.com/docs/overview/)
- [API Reference](https://api.llama.com/compat/v1/)
- [Model Card - Llama 4](https://ai.meta.com/llama/)
- [OpenAI Compatibility Guide](https://platform.openai.com/docs/api-reference/chat)

## Support

For issues specific to this provider:
- Check [GitHub Issues](https://github.com/AINative-Studio/ainative-code/issues)
- Review test examples in `internal/provider/meta/*_test.go`

For Meta API issues:
- Visit [Meta Developer Portal](https://llama.developer.meta.com/)
- Contact Meta Support

## License

This provider implementation is part of the AINative-Code project and follows the same license.

---

**Note**: Meta LLAMA is a product of Meta Platforms, Inc. This provider is an independent integration and is not officially endorsed by Meta.
