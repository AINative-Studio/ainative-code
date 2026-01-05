# OpenAI Provider Quick Start Guide

Get started with the OpenAI provider in 5 minutes.

## Installation

```bash
go get github.com/AINative-studio/ainative-code
```

## Basic Usage

### 1. Import Packages

```go
import (
    "context"
    "github.com/AINative-studio/ainative-code/internal/provider"
    "github.com/AINative-studio/ainative-code/internal/provider/openai"
)
```

### 2. Create Provider

```go
p, err := openai.NewOpenAIProvider(openai.Config{
    APIKey: "sk-...", // Your OpenAI API key
})
if err != nil {
    panic(err)
}
defer p.Close()
```

### 3. Send Chat Request

```go
ctx := context.Background()
messages := []provider.Message{
    {Role: "user", Content: "Hello, how are you?"},
}

resp, err := p.Chat(ctx, messages,
    provider.WithModel("gpt-4"),
    provider.WithMaxTokens(100),
)
if err != nil {
    panic(err)
}

fmt.Println(resp.Content)
```

## Streaming Example

```go
eventChan, err := p.Stream(ctx, messages,
    provider.StreamWithModel("gpt-3.5-turbo"),
)

for event := range eventChan {
    if event.Type == provider.EventTypeContentDelta {
        fmt.Print(event.Content)
    }
}
```

## Environment Variables

```bash
export OPENAI_API_KEY="sk-..."
export OPENAI_ORG_ID="org-..."  # Optional
```

Then in code:

```go
import "os"

p, _ := openai.NewOpenAIProvider(openai.Config{
    APIKey:       os.Getenv("OPENAI_API_KEY"),
    Organization: os.Getenv("OPENAI_ORG_ID"),
})
```

## Popular Models

| Model | Use Case | Context |
|-------|----------|---------|
| `gpt-4-turbo-preview` | Latest GPT-4 | 128K |
| `gpt-4` | Complex reasoning | 8K |
| `gpt-3.5-turbo` | Fast, cheap | 4K |
| `gpt-3.5-turbo-16k` | Longer context | 16K |

## Common Options

```go
resp, _ := p.Chat(ctx, messages,
    provider.WithModel("gpt-4"),              // Choose model
    provider.WithMaxTokens(500),              // Limit response length
    provider.WithTemperature(0.7),            // Creativity (0-1)
    provider.WithSystemPrompt("Be concise"),  // System instruction
)
```

## Error Handling

```go
resp, err := p.Chat(ctx, messages, provider.WithModel("gpt-4"))
if err != nil {
    switch {
    case strings.Contains(err.Error(), "authentication"):
        fmt.Println("Invalid API key")
    case strings.Contains(err.Error(), "rate limit"):
        fmt.Println("Too many requests, wait and retry")
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Embeddings

**IMPORTANT**: Use AINative platform for embeddings, not OpenAI:

```go
import "github.com/AINative-studio/ainative-code/internal/embeddings"

client, _ := embeddings.NewAINativeEmbeddingsClient(embeddings.Config{
    APIKey: os.Getenv("AINATIVE_API_KEY"),
})
defer client.Close()

result, _ := client.Embed(ctx, []string{"Hello world"}, "default")
fmt.Printf("Embedding dimension: %d\n", len(result.Embeddings[0]))
```

## Multi-Provider Setup

```go
// Create registry
registry := provider.NewRegistry()

// Add providers
openaiProvider, _ := openai.NewOpenAIProvider(openaiConfig)
registry.Register("openai", openaiProvider)

// Use by name
p, _ := registry.Get("openai")
resp, _ := p.Chat(ctx, messages, provider.WithModel("gpt-4"))
```

## Configuration File

Create `config.yaml`:

```yaml
llm:
  default_provider: "openai"
  openai:
    api_key: "${OPENAI_API_KEY}"
    model: "gpt-4"
    max_tokens: 1000
    temperature: 0.7
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/AINative-studio/ainative-code/internal/provider"
    "github.com/AINative-studio/ainative-code/internal/provider/openai"
)

func main() {
    // Create provider
    p, err := openai.NewOpenAIProvider(openai.Config{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
    if err != nil {
        log.Fatal(err)
    }
    defer p.Close()

    // Prepare messages
    messages := []provider.Message{
        {Role: "system", Content: "You are a helpful assistant."},
        {Role: "user", Content: "What is the capital of France?"},
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

    fmt.Printf("Response: %s\n", resp.Content)
    fmt.Printf("Tokens: %d\n", resp.Usage.TotalTokens)
}
```

## Next Steps

- Read full documentation: `OPENAI_PROVIDER.md`
- Check examples: `examples/openai_provider_example.go`
- Learn about embeddings: `EMBEDDINGS.md`
- Explore multi-provider: See registry examples

## Tips

1. **Use context**: Always pass context for timeout control
2. **Set max tokens**: Prevent runaway costs
3. **Choose right model**: gpt-3.5-turbo for speed, gpt-4 for quality
4. **Stream long responses**: Better UX with streaming
5. **Handle errors**: Implement retry logic for rate limits

## Support

- Documentation: `/docs/OPENAI_PROVIDER.md`
- Examples: `/examples/openai_provider_example.go`
- Tests: `/internal/provider/openai/*_test.go`
- Issues: GitHub issue tracker
