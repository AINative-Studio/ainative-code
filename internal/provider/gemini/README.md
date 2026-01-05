# Google Gemini Provider

The Gemini provider implements the Provider interface for Google's Gemini API, supporting both standard and multi-modal interactions.

## Features

- **Multiple Models**: Support for gemini-pro, gemini-pro-vision, gemini-ultra, and gemini-1.5-pro
- **Chat Completions**: Non-streaming chat requests
- **Streaming Support**: Server-Sent Events (SSE) for real-time responses
- **Multi-Modal**: Support for text and image inputs (via gemini-pro-vision)
- **Safety Settings**: Built-in content safety filtering
- **Function Calling**: Support for tool/function calling
- **Error Handling**: Comprehensive error mapping and recovery

## Supported Models

| Model | Description | Context Length | Capabilities |
|-------|-------------|----------------|--------------|
| gemini-pro | General purpose model | 32,768 tokens | Text generation, chat |
| gemini-pro-vision | Multi-modal model | 16,384 tokens | Text + image understanding |
| gemini-ultra | Most capable model | 32,768 tokens | Complex reasoning, long context |
| gemini-1.5-pro | Latest production model | 1,048,576 tokens | Extended context, improved performance |
| gemini-1.5-flash | Fast, efficient model | 1,048,576 tokens | Low latency, cost-effective |

## Installation

```go
import (
    "github.com/AINative-studio/ainative-code/internal/provider/gemini"
    "github.com/AINative-studio/ainative-code/internal/provider"
)
```

## Basic Usage

### Creating a Provider

```go
// Create Gemini provider with API key
config := gemini.Config{
    APIKey: "your-google-api-key",
}

provider, err := gemini.NewGeminiProvider(config)
if err != nil {
    log.Fatal(err)
}
defer provider.Close()
```

### Simple Chat Request

```go
ctx := context.Background()

messages := []provider.Message{
    {Role: "user", Content: "What is the capital of France?"},
}

response, err := provider.Chat(ctx, messages,
    provider.WithModel("gemini-pro"),
    provider.WithMaxTokens(100),
    provider.WithTemperature(0.7),
)

if err != nil {
    log.Fatal(err)
}

fmt.Println("Response:", response.Content)
fmt.Printf("Tokens: %d prompt + %d completion = %d total\n",
    response.Usage.PromptTokens,
    response.Usage.CompletionTokens,
    response.Usage.TotalTokens)
```

### Streaming Chat

```go
ctx := context.Background()

messages := []provider.Message{
    {Role: "user", Content: "Write a haiku about Go programming"},
}

eventChan, err := provider.Stream(ctx, messages,
    provider.StreamWithModel("gemini-pro"),
    provider.StreamWithTemperature(0.8),
)

if err != nil {
    log.Fatal(err)
}

for event := range eventChan {
    switch event.Type {
    case provider.EventTypeContentStart:
        fmt.Print("Starting... ")
    case provider.EventTypeContentDelta:
        fmt.Print(event.Content)
    case provider.EventTypeContentEnd:
        fmt.Println("\nDone!")
    case provider.EventTypeError:
        log.Printf("Error: %v", event.Error)
    }
}
```

### Multi-Turn Conversation

```go
ctx := context.Background()

messages := []provider.Message{
    {Role: "user", Content: "Hello! What's your name?"},
    {Role: "assistant", Content: "I'm Gemini, an AI assistant created by Google."},
    {Role: "user", Content: "What can you help me with?"},
}

response, err := provider.Chat(ctx, messages,
    provider.WithModel("gemini-pro"),
    provider.WithMaxTokens(200),
)

if err != nil {
    log.Fatal(err)
}

fmt.Println(response.Content)
```

### Using System Prompts

```go
ctx := context.Background()

messages := []provider.Message{
    {Role: "user", Content: "Explain quantum computing"},
}

response, err := provider.Chat(ctx, messages,
    provider.WithModel("gemini-pro"),
    provider.WithSystemPrompt("You are a helpful physics teacher. Explain concepts clearly and simply."),
    provider.WithTemperature(0.5),
)

if err != nil {
    log.Fatal(err)
}

fmt.Println(response.Content)
```

### Advanced Configuration

```go
// Custom base URL and HTTP client
httpClient := &http.Client{
    Timeout: 30 * time.Second,
}

config := gemini.Config{
    APIKey:     "your-api-key",
    BaseURL:    "https://custom.api.endpoint.com",
    HTTPClient: httpClient,
    Logger:     myLogger,
}

provider, err := gemini.NewGeminiProvider(config)
if err != nil {
    log.Fatal(err)
}
```

### Using TopK Parameter (Gemini-specific)

```go
messages := []provider.Message{
    {Role: "user", Content: "Generate creative ideas"},
}

response, err := provider.Chat(ctx, messages,
    provider.WithModel("gemini-pro"),
    provider.WithTemperature(0.9),
    provider.WithMetadata("topK", "40"), // Gemini-specific parameter
)
```

## Error Handling

The provider returns typed errors that can be inspected:

```go
response, err := provider.Chat(ctx, messages, provider.WithModel("gemini-pro"))
if err != nil {
    switch e := err.(type) {
    case *provider.AuthenticationError:
        log.Printf("Invalid API key: %v", e)
    case *provider.RateLimitError:
        log.Printf("Rate limited. Retry after %d seconds", e.RetryAfter)
    case *provider.ContextLengthError:
        log.Printf("Content too long: %d tokens (max: %d)", e.RequestedTokens, e.MaxTokens)
    case *provider.InvalidModelError:
        log.Printf("Invalid model. Supported models: %v", e.SupportedModels)
    default:
        log.Printf("Error: %v", err)
    }
    return
}
```

## Safety Settings

Gemini includes built-in safety filtering. If content is blocked:

```go
response, err := provider.Chat(ctx, messages, provider.WithModel("gemini-pro"))
if err != nil {
    if strings.Contains(err.Error(), "blocked") {
        log.Println("Content was blocked by safety settings")
        // Handle blocked content
    }
}
```

## Configuration Reference

### Config Options

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| APIKey | string | Google API key (required) | - |
| BaseURL | string | API base URL | https://generativelanguage.googleapis.com/v1beta |
| HTTPClient | *http.Client | Custom HTTP client | Standard client with 60s timeout |
| Logger | LoggerInterface | Logger for debugging | nil |

### Chat Options

| Option | Description | Default |
|--------|-------------|---------|
| WithModel() | Set the model to use | Required |
| WithMaxTokens() | Maximum tokens to generate | 1024 |
| WithTemperature() | Sampling temperature (0.0-1.0) | 0.7 |
| WithTopP() | Nucleus sampling parameter | 1.0 |
| WithStopSequences() | Sequences that stop generation | nil |
| WithSystemPrompt() | System instruction | "" |
| WithMetadata() | Custom metadata (e.g., topK) | {} |

## Best Practices

1. **API Key Security**: Never hardcode API keys. Use environment variables or secure vaults.
2. **Error Handling**: Always check for and handle errors appropriately.
3. **Context Management**: Use context with timeouts for production code.
4. **Resource Cleanup**: Always call `Close()` when done with the provider.
5. **Model Selection**: Choose the right model for your use case:
   - `gemini-pro`: General purpose tasks
   - `gemini-pro-vision`: When you need image understanding
   - `gemini-1.5-pro`: For maximum context and best performance
   - `gemini-1.5-flash`: For low-latency, cost-effective applications

## Testing

The provider includes comprehensive unit and integration tests:

```bash
# Run unit tests
go test ./internal/provider/gemini/...

# Run with coverage
go test -cover ./internal/provider/gemini/...

# Run integration tests
go test ./tests/integration/gemini_integration_test.go
```

## Embeddings Note

For vector embeddings and semantic search, use the **AINative platform APIs** instead of Google's embedding endpoints. The Gemini provider focuses on text generation only.

```go
// For embeddings, use:
import "github.com/AINative-studio/ainative-code/internal/embeddings/ainative"
```

## API Rate Limits

Google applies rate limits to Gemini API requests. The provider automatically handles rate limiting with exponential backoff retry logic. Monitor the `RateLimitError` for rate limit information.

## References

- [Gemini API Documentation](https://ai.google.dev/docs)
- [Supported Models](https://ai.google.dev/models/gemini)
- [Safety Settings](https://ai.google.dev/docs/safety_setting_gemini)
- [API Pricing](https://ai.google.dev/pricing)
