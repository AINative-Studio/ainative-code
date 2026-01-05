# Ollama Provider Documentation

## Overview

The Ollama provider enables local model inference using [Ollama](https://ollama.ai/), supporting Meta's LLAMA models (llama2, llama3, codellama) and other popular open-source models. This provider is part of the multi-provider architecture alongside Anthropic, OpenAI, Google Gemini, and AWS Bedrock.

## Features

- **Local Model Inference**: Run models locally without cloud dependencies
- **LLAMA Model Support**: Full support for Meta's LLAMA 2, LLAMA 3, and Code LLAMA
- **Streaming Support**: Real-time streaming responses
- **Model Management**: List, check availability, and get information about installed models
- **Health Checking**: Verify Ollama server connectivity
- **Comprehensive Error Handling**: Specific error types for connection, model not found, and out of memory issues
- **80.6% Test Coverage**: Thoroughly tested with unit and integration tests

## Supported Models

### LLAMA Models (Meta)

- **llama2** - LLAMA 2 7B (default)
- **llama2:7b** - LLAMA 2 7B
- **llama2:13b** - LLAMA 2 13B
- **llama2:70b** - LLAMA 2 70B
- **llama3** - LLAMA 3 8B (default)
- **llama3:8b** - LLAMA 3 8B
- **llama3:70b** - LLAMA 3 70B
- **codellama** - Code LLAMA 7B (default)
- **codellama:7b** - Code LLAMA 7B
- **codellama:13b** - Code LLAMA 13B
- **codellama:34b** - Code LLAMA 34B
- **codellama:70b** - Code LLAMA 70B

### Other Popular Models

- **mistral** - Mistral 7B
- **mixtral** - Mixtral 8x7B
- **phi** - Microsoft Phi
- **neural-chat** - Neural Chat 7B
- **orca-mini** - Orca Mini
- **vicuna** - Vicuna
- **gemma** - Google Gemma

## Installation & Setup

### 1. Install Ollama

```bash
# macOS / Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Or visit https://ollama.ai/download for other platforms
```

### 2. Start Ollama Server

```bash
ollama serve
```

The server will start on `http://localhost:11434` by default.

### 3. Download Models

```bash
# Download LLAMA 2 (7B)
ollama pull llama2

# Download LLAMA 3 (8B)
ollama pull llama3

# Download Code LLAMA (7B)
ollama pull codellama

# Download larger models
ollama pull llama2:13b
ollama pull llama3:70b

# List all available models
ollama list
```

## Usage Examples

### Basic Chat Request

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AINative-studio/ainative-code/internal/provider"
    "github.com/AINative-studio/ainative-code/internal/provider/ollama"
)

func main() {
    // Create Ollama provider for LLAMA 2
    config := ollama.Config{
        Model: "llama2",
    }

    provider, err := ollama.NewOllamaProvider(config)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Close()

    // Create messages
    messages := []provider.Message{
        {Role: "user", Content: "What is machine learning?"},
    }

    // Send chat request
    ctx := context.Background()
    response, err := provider.Chat(ctx, messages,
        provider.WithModel("llama2"),
        provider.WithMaxTokens(2048),
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\n", response.Content)
    fmt.Printf("Tokens - Prompt: %d, Completion: %d\n",
        response.Usage.PromptTokens,
        response.Usage.CompletionTokens,
    )
}
```

### Streaming Responses

```go
func streamExample() {
    config := ollama.Config{
        Model: "llama3",
    }

    provider, err := ollama.NewOllamaProvider(config)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Close()

    messages := []provider.Message{
        {Role: "user", Content: "Write a short poem about coding"},
    }

    ctx := context.Background()
    eventChan, err := provider.Stream(ctx, messages,
        provider.StreamWithModel("llama3"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Process streaming events
    for event := range eventChan {
        switch event.Type {
        case provider.EventTypeContentStart:
            fmt.Println("Stream started...")
        case provider.EventTypeContentDelta:
            fmt.Print(event.Content)
        case provider.EventTypeContentEnd:
            fmt.Println("\n\nStream completed!")
        case provider.EventTypeError:
            log.Printf("Error: %v\n", event.Error)
        }
    }
}
```

### Code Generation with Code LLAMA

```go
func codeGenerationExample() {
    config := ollama.Config{
        Model: "codellama",
        Temperature: 0.2, // Lower temperature for more deterministic code
    }

    provider, err := ollama.NewOllamaProvider(config)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Close()

    messages := []provider.Message{
        {
            Role: "system",
            Content: "You are an expert programmer. Provide clean, well-documented code.",
        },
        {
            Role: "user",
            Content: "Write a Go function to calculate fibonacci numbers recursively",
        },
    }

    ctx := context.Background()
    response, err := provider.Chat(ctx, messages,
        provider.WithModel("codellama"),
        provider.WithTemperature(0.2),
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Generated Code:")
    fmt.Println(response.Content)
}
```

### Model Management

```go
func modelManagementExample() {
    config := ollama.Config{
        Model: "llama2",
    }

    provider, err := ollama.NewOllamaProvider(config)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Close()

    ctx := context.Background()

    // Check health
    if err := provider.HealthCheck(ctx); err != nil {
        log.Fatal("Ollama server not reachable:", err)
    }

    // List available models
    models, err := provider.ListAvailableModels(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Available Models:")
    for _, model := range models {
        fmt.Printf("- %s (%s, %s)\n",
            model.Name,
            model.ParameterSize,
            model.Family,
        )
    }

    // Check if specific model is available
    available, err := provider.IsModelAvailable(ctx, "llama3")
    if err != nil {
        log.Fatal(err)
    }

    if !available {
        fmt.Println("LLAMA 3 not found. Download with: ollama pull llama3")
    }

    // Get model information
    info, err := provider.GetModelInfo(ctx, "llama2")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Model: %s\n", info.Name)
    fmt.Printf("Family: %s\n", info.Family)
    fmt.Printf("Parameters: %s\n", info.ParameterSize)
    fmt.Printf("Size: %d bytes\n", info.Size)
}
```

### Conversation with Context

```go
func conversationExample() {
    config := ollama.Config{
        Model: "llama2",
        NumCtx: 4096, // Larger context window
    }

    provider, err := ollama.NewOllamaProvider(config)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Close()

    // Multi-turn conversation
    messages := []provider.Message{
        {Role: "user", Content: "My favorite programming language is Go"},
        {Role: "assistant", Content: "That's great! Go is known for its simplicity, performance, and excellent concurrency support."},
        {Role: "user", Content: "What are its best features?"},
    }

    ctx := context.Background()
    response, err := provider.Chat(ctx, messages)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response.Content)
}
```

### Custom Configuration

```go
func customConfigExample() {
    config := ollama.Config{
        BaseURL:     "http://custom-host:11434",
        Model:       "mistral",
        NumCtx:      8192,       // Context window
        Temperature: 0.9,        // Creativity (0.0 - 2.0)
        TopK:        50,         // Top-k sampling
        TopP:        0.95,       // Nucleus sampling
        Timeout:     120 * time.Second,
    }

    provider, err := ollama.NewOllamaProvider(config)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Close()

    // Use provider...
}
```

### Error Handling

```go
func errorHandlingExample() {
    config := ollama.Config{
        Model: "unknown-model",
    }

    provider, err := ollama.NewOllamaProvider(config)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Close()

    ctx := context.Background()
    messages := []provider.Message{
        {Role: "user", Content: "Hello"},
    }

    response, err := provider.Chat(ctx, messages)
    if err != nil {
        // Check specific error types
        if ollama.IsConnectionError(err) {
            fmt.Println("Ollama server is not running. Start with: ollama serve")
        } else if ollama.IsModelNotFoundError(err) {
            fmt.Println("Model not found. Download with: ollama pull <model>")
        } else if ollama.IsOutOfMemoryError(err) {
            fmt.Println("Out of memory. Try a smaller model.")
        } else {
            log.Printf("Error: %v\n", err)
        }
        return
    }

    fmt.Println(response.Content)
}
```

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `BaseURL` | string | `http://localhost:11434` | Ollama server endpoint |
| `Model` | string | *required* | Model name (e.g., "llama2") |
| `NumCtx` | int | 2048 | Context window size |
| `Temperature` | float64 | 0.8 | Sampling temperature (0.0 - 2.0) |
| `TopK` | int | 40 | Top-k sampling |
| `TopP` | float64 | 0.9 | Nucleus sampling (0.0 - 1.0) |
| `Timeout` | time.Duration | 120s | Request timeout |
| `HTTPClient` | *http.Client | nil | Custom HTTP client |
| `Logger` | logger.LoggerInterface | nil | Custom logger |

## Error Types

The provider includes specific error types for better error handling:

- **`OllamaConnectionError`**: Ollama server not reachable
- **`OllamaOutOfMemoryError`**: Insufficient memory for model
- **`InvalidModelError`**: Model not found or invalid
- **`ContextLengthError`**: Input exceeds context window
- **`ProviderError`**: Generic provider errors

## Embeddings

**Important**: For vector embeddings and semantic search, use the **AINative Platform APIs** instead of Ollama's embedding endpoint. This ensures consistent embedding generation across the platform.

## Testing

The provider includes comprehensive tests with 80.6% code coverage:

```bash
# Run all tests
go test ./internal/provider/ollama/...

# Run with coverage
go test -cover ./internal/provider/ollama/...

# Generate coverage report
go test -coverprofile=coverage.out ./internal/provider/ollama/...
go tool cover -html=coverage.out
```

## Performance Considerations

### Model Sizes

- **7B models**: Require ~8GB RAM
- **13B models**: Require ~16GB RAM
- **70B models**: Require ~64GB RAM or GPU acceleration

### Optimization Tips

1. **Use appropriate model sizes** for your hardware
2. **Adjust context window** (`NumCtx`) based on needs
3. **Enable GPU acceleration** if available (Ollama auto-detects)
4. **Use streaming** for better user experience with large responses
5. **Lower temperature** (0.2-0.4) for deterministic outputs

## Troubleshooting

### Ollama Server Not Running

```bash
# Error: connection refused
# Solution: Start Ollama server
ollama serve
```

### Model Not Found

```bash
# Error: model 'llama3' not found
# Solution: Download the model
ollama pull llama3
```

### Out of Memory

```bash
# Error: out of memory loading model
# Solutions:
# 1. Use a smaller model (e.g., llama2:7b instead of llama2:70b)
# 2. Close other applications to free RAM
# 3. Use GPU acceleration
```

### Context Length Exceeded

```go
# Error: context length exceeded
# Solution: Reduce context or use a model with larger context window
config := ollama.Config{
    Model: "llama2",
    NumCtx: 8192, // Increase from default 2048
}
```

## Integration with Provider Registry

Register Ollama provider with the multi-provider registry:

```go
import (
    "github.com/AINative-studio/ainative-code/internal/provider"
    "github.com/AINative-studio/ainative-code/internal/provider/ollama"
)

func main() {
    registry := provider.NewRegistry()

    // Register Ollama provider
    ollamaProvider, err := ollama.NewOllamaProviderForModel("llama2")
    if err != nil {
        log.Fatal(err)
    }

    if err := registry.Register("ollama", ollamaProvider); err != nil {
        log.Fatal(err)
    }

    // Use provider from registry
    p, err := registry.Get("ollama")
    if err != nil {
        log.Fatal(err)
    }

    // Use p for chat...
}
```

## Comparison with Cloud Providers

| Feature | Ollama | Anthropic | OpenAI |
|---------|--------|-----------|--------|
| **Cost** | Free (local) | Pay per token | Pay per token |
| **Privacy** | Complete (local) | Cloud-based | Cloud-based |
| **Speed** | Depends on hardware | Fast | Fast |
| **Model Size** | Limited by RAM | Large models | Large models |
| **Internet** | Not required | Required | Required |
| **Scaling** | Single machine | Unlimited | Unlimited |

## Best Practices

1. **Check health** before making requests
2. **Verify model availability** before use
3. **Handle connection errors** gracefully
4. **Use streaming** for better UX
5. **Implement retries** for transient failures
6. **Monitor memory usage** with large models
7. **Close provider** when done to free resources

## References

- [Ollama Official Website](https://ollama.ai/)
- [Ollama GitHub](https://github.com/ollama/ollama)
- [Meta LLAMA](https://ai.meta.com/llama/)
- [Ollama Model Library](https://ollama.ai/library)

## License

This provider implementation is part of the AINative-Code project.
