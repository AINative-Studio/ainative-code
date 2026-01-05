# Extending Providers Guide

## Overview

This guide explains how to create custom LLM provider integrations for AINative Code. By implementing the Provider interface, you can add support for any LLM API.

## Provider Interface

### Interface Definition

```go
// Location: internal/provider/provider.go
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
```

### Core Types

```go
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
```

## Creating a Custom Provider

### Step 1: Define Configuration

Create a configuration struct for your provider:

```go
package myprovider

// Config contains configuration for MyProvider
type Config struct {
    APIKey     string
    BaseURL    string
    HTTPClient *http.Client
    Logger     logger.LoggerInterface
}
```

### Step 2: Create Provider Struct

```go
package myprovider

import (
    "github.com/AINative-studio/ainative-code/internal/provider"
)

// MyProvider implements the Provider interface
type MyProvider struct {
    *provider.BaseProvider  // Embed BaseProvider for common functionality
    apiKey  string
    baseURL string
}

// Supported models for this provider
var supportedModels = []string{
    "my-model-1",
    "my-model-2",
}
```

### Step 3: Implement Constructor

```go
func NewMyProvider(config Config) (*MyProvider, error) {
    // Validate configuration
    if config.APIKey == "" {
        return nil, provider.NewAuthenticationError("myprovider", "API key is required")
    }

    baseURL := config.BaseURL
    if baseURL == "" {
        baseURL = "https://api.myprovider.com/v1"
    }

    // Create base provider with common configuration
    baseProvider := provider.NewBaseProvider(provider.BaseProviderConfig{
        Name:       "myprovider",
        HTTPClient: config.HTTPClient,
        Logger:     config.Logger,
        RetryConfig: provider.DefaultRetryConfig(),
    })

    return &MyProvider{
        BaseProvider: baseProvider,
        apiKey:       config.APIKey,
        baseURL:      baseURL,
    }, nil
}
```

### Step 4: Implement Interface Methods

```go
// Name returns the provider name
func (p *MyProvider) Name() string {
    return p.BaseProvider.Name()
}

// Models returns supported models
func (p *MyProvider) Models() []string {
    models := make([]string, len(supportedModels))
    copy(models, supportedModels)
    return models
}

// Close releases resources
func (p *MyProvider) Close() error {
    return nil  // Cleanup if needed
}
```

### Step 5: Implement Chat Method

```go
func (p *MyProvider) Chat(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error) {
    // 1. Apply options
    options := provider.DefaultChatOptions()
    provider.ApplyChatOptions(options, opts...)

    // 2. Validate model
    if err := p.ValidateModel(options.Model, supportedModels); err != nil {
        return provider.Response{}, err
    }

    // 3. Build HTTP request
    req, err := p.buildRequest(ctx, messages, options, false)
    if err != nil {
        return provider.Response{}, provider.NewProviderError("myprovider", options.Model, err)
    }

    // 4. Execute request (uses BaseProvider's DoRequest with retry logic)
    resp, err := p.DoRequest(ctx, req)
    if err != nil {
        return provider.Response{}, err
    }
    defer resp.Body.Close()

    // 5. Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return provider.Response{}, provider.NewProviderError("myprovider", options.Model,
            fmt.Errorf("failed to read response: %w", err))
    }

    // 6. Handle API errors
    if resp.StatusCode != http.StatusOK {
        return provider.Response{}, p.handleAPIError(resp, body, options.Model)
    }

    // 7. Parse and return response
    return p.parseResponse(body, options.Model)
}
```

### Step 6: Implement Stream Method

```go
func (p *MyProvider) Stream(ctx context.Context, messages []provider.Message, opts ...provider.StreamOption) (<-chan provider.Event, error) {
    // 1. Apply options
    options := provider.DefaultChatOptions()
    provider.ApplyStreamOptions(options, opts...)

    // 2. Validate model
    if err := p.ValidateModel(options.Model, supportedModels); err != nil {
        return nil, err
    }

    // 3. Build request
    req, err := p.buildRequest(ctx, messages, options, true)
    if err != nil {
        return nil, provider.NewProviderError("myprovider", options.Model, err)
    }

    // 4. Execute request
    resp, err := p.DoRequest(ctx, req)
    if err != nil {
        return nil, err
    }

    // 5. Handle errors
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        resp.Body.Close()
        return nil, p.handleAPIError(resp, body, options.Model)
    }

    // 6. Create event channel
    events := make(chan provider.Event, 10)

    // 7. Start streaming goroutine
    go func() {
        defer resp.Body.Close()
        defer close(events)

        if err := p.streamResponse(ctx, resp.Body, events, options.Model); err != nil {
            events <- provider.Event{
                Type:  provider.EventTypeError,
                Error: err,
                Done:  true,
            }
        }
    }()

    return events, nil
}
```

## Helper Methods

### Build Request

```go
func (p *MyProvider) buildRequest(ctx context.Context, messages []provider.Message, options *provider.ChatOptions, stream bool) (*http.Request, error) {
    // Convert messages to API format
    apiMessages := make([]map[string]string, len(messages))
    for i, msg := range messages {
        apiMessages[i] = map[string]string{
            "role":    msg.Role,
            "content": msg.Content,
        }
    }

    // Build request body
    reqBody := map[string]interface{}{
        "model":       options.Model,
        "messages":    apiMessages,
        "max_tokens":  options.MaxTokens,
        "temperature": options.Temperature,
        "stream":      stream,
    }

    // Encode JSON
    bodyBytes, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("failed to encode request: %w", err)
    }

    // Create HTTP request
    req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(bodyBytes))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    // Set headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+p.apiKey)
    req.Header.Set("User-Agent", "AINative-Code/1.0")

    return req, nil
}
```

### Parse Response

```go
type apiResponse struct {
    ID      string `json:"id"`
    Object  string `json:"object"`
    Created int64  `json:"created"`
    Model   string `json:"model"`
    Choices []struct {
        Index   int `json:"index"`
        Message struct {
            Role    string `json:"role"`
            Content string `json:"content"`
        } `json:"message"`
        FinishReason string `json:"finish_reason"`
    } `json:"choices"`
    Usage struct {
        PromptTokens     int `json:"prompt_tokens"`
        CompletionTokens int `json:"completion_tokens"`
        TotalTokens      int `json:"total_tokens"`
    } `json:"usage"`
}

func (p *MyProvider) parseResponse(body []byte, model string) (provider.Response, error) {
    var apiResp apiResponse
    if err := json.Unmarshal(body, &apiResp); err != nil {
        return provider.Response{}, fmt.Errorf("failed to parse response: %w", err)
    }

    if len(apiResp.Choices) == 0 {
        return provider.Response{}, fmt.Errorf("no choices in response")
    }

    return provider.Response{
        Content: apiResp.Choices[0].Message.Content,
        Usage: provider.Usage{
            PromptTokens:     apiResp.Usage.PromptTokens,
            CompletionTokens: apiResp.Usage.CompletionTokens,
            TotalTokens:      apiResp.Usage.TotalTokens,
        },
        Model: apiResp.Model,
    }, nil
}
```

### Handle API Errors

```go
type apiError struct {
    Error struct {
        Message string `json:"message"`
        Type    string `json:"type"`
        Code    string `json:"code"`
    } `json:"error"`
}

func (p *MyProvider) handleAPIError(resp *http.Response, body []byte, model string) error {
    var apiErr apiError
    if err := json.Unmarshal(body, &apiErr); err != nil {
        return provider.NewProviderError("myprovider", model,
            fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)))
    }

    // Map API errors to provider error types
    switch resp.StatusCode {
    case http.StatusUnauthorized:
        return provider.NewAuthenticationError("myprovider", apiErr.Error.Message)
    case http.StatusTooManyRequests:
        return provider.NewRateLimitError("myprovider", apiErr.Error.Message)
    case http.StatusBadRequest:
        return provider.NewValidationError("myprovider", apiErr.Error.Message)
    default:
        return provider.NewProviderError("myprovider", model,
            fmt.Errorf("%s: %s", apiErr.Error.Type, apiErr.Error.Message))
    }
}
```

### Stream Response

```go
func (p *MyProvider) streamResponse(ctx context.Context, body io.Reader, events chan<- provider.Event, model string) error {
    scanner := bufio.NewScanner(body)

    // Send start event
    events <- provider.Event{
        Type: provider.EventTypeContentStart,
    }

    for scanner.Scan() {
        line := scanner.Text()

        // Check for context cancellation
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        // Parse SSE format: "data: {...}"
        if !strings.HasPrefix(line, "data: ") {
            continue
        }

        data := strings.TrimPrefix(line, "data: ")

        // Check for stream end
        if data == "[DONE]" {
            break
        }

        // Parse chunk
        var chunk struct {
            Choices []struct {
                Delta struct {
                    Content string `json:"content"`
                } `json:"delta"`
            } `json:"choices"`
        }

        if err := json.Unmarshal([]byte(data), &chunk); err != nil {
            continue  // Skip malformed chunks
        }

        if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
            events <- provider.Event{
                Type:    provider.EventTypeContentDelta,
                Content: chunk.Choices[0].Delta.Content,
            }
        }
    }

    if err := scanner.Err(); err != nil {
        return fmt.Errorf("stream error: %w", err)
    }

    // Send end event
    events <- provider.Event{
        Type: provider.EventTypeContentEnd,
        Done: true,
    }

    return nil
}
```

## Testing Your Provider

### Unit Tests

```go
package myprovider

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/AINative-studio/ainative-code/internal/provider"
)

func TestMyProvider_Chat(t *testing.T) {
    // Create test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "/chat/completions", r.URL.Path)
        assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

        // Send mock response
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{
            "id": "test-123",
            "choices": [{
                "message": {
                    "role": "assistant",
                    "content": "Hello, world!"
                }
            }],
            "usage": {
                "prompt_tokens": 10,
                "completion_tokens": 5,
                "total_tokens": 15
            }
        }`))
    }))
    defer server.Close()

    // Create provider
    p, err := NewMyProvider(Config{
        APIKey:  "test-key",
        BaseURL: server.URL,
    })
    require.NoError(t, err)

    // Test chat
    ctx := context.Background()
    messages := []provider.Message{
        {Role: "user", Content: "Hello"},
    }

    response, err := p.Chat(ctx, messages)
    require.NoError(t, err)
    assert.Equal(t, "Hello, world!", response.Content)
    assert.Equal(t, 15, response.Usage.TotalTokens)
}

func TestMyProvider_Stream(t *testing.T) {
    // Create test server that sends SSE
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/event-stream")
        w.WriteHeader(http.StatusOK)

        // Send chunks
        chunks := []string{
            `data: {"choices":[{"delta":{"content":"Hello"}}]}`,
            `data: {"choices":[{"delta":{"content":", world"}}]}`,
            `data: {"choices":[{"delta":{"content":"!"}}]}`,
            `data: [DONE]`,
        }

        for _, chunk := range chunks {
            w.Write([]byte(chunk + "\n"))
            if f, ok := w.(http.Flusher); ok {
                f.Flush()
            }
        }
    }))
    defer server.Close()

    // Create provider
    p, err := NewMyProvider(Config{
        APIKey:  "test-key",
        BaseURL: server.URL,
    })
    require.NoError(t, err)

    // Test stream
    ctx := context.Background()
    messages := []provider.Message{
        {Role: "user", Content: "Hello"},
    }

    events, err := p.Stream(ctx, messages)
    require.NoError(t, err)

    // Collect content
    var content string
    for event := range events {
        if event.Type == provider.EventTypeContentDelta {
            content += event.Content
        }
    }

    assert.Equal(t, "Hello, world!", content)
}
```

### Integration Tests

```go
//go:build integration

package myprovider

import (
    "context"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMyProvider_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    apiKey := os.Getenv("MYPROVIDER_API_KEY")
    if apiKey == "" {
        t.Skip("MYPROVIDER_API_KEY not set")
    }

    p, err := NewMyProvider(Config{
        APIKey: apiKey,
    })
    require.NoError(t, err)

    ctx := context.Background()
    messages := []provider.Message{
        {Role: "user", Content: "Say 'test' if you receive this"},
    }

    response, err := p.Chat(ctx, messages)
    require.NoError(t, err)
    assert.NotEmpty(t, response.Content)
    assert.Greater(t, response.Usage.TotalTokens, 0)
}
```

## Registering Your Provider

### Add to Registry

```go
// internal/cmd/root.go or similar initialization code
import (
    "github.com/AINative-studio/ainative-code/internal/provider"
    "your/package/myprovider"
)

func init() {
    // Register provider factory
    provider.RegisterFactory("myprovider", func(config interface{}) (provider.Provider, error) {
        cfg := config.(myprovider.Config)
        return myprovider.NewMyProvider(cfg)
    })
}
```

### Configuration

Add your provider to the configuration schema:

```yaml
# config.yaml
providers:
  myprovider:
    api_key: "${MYPROVIDER_API_KEY}"
    model: "my-model-1"
    max_tokens: 4096
    temperature: 0.7
```

## Best Practices

### 1. Use BaseProvider

Leverage `BaseProvider` for common functionality:
- HTTP client management
- Retry logic with exponential backoff
- Request/response logging
- Model validation

### 2. Handle Context Properly

Always respect context cancellation and timeouts:

```go
select {
case <-ctx.Done():
    return ctx.Err()
default:
    // Continue processing
}
```

### 3. Implement Proper Error Handling

Use appropriate error types:

```go
provider.NewAuthenticationError()  // For auth issues
provider.NewRateLimitError()       // For rate limiting
provider.NewValidationError()      // For invalid input
provider.NewProviderError()        // For general errors
```

### 4. Log Appropriately

Use the provided logger:

```go
p.Logger().InfoWithFields("Chat request", map[string]interface{}{
    "model":    options.Model,
    "messages": len(messages),
})
```

### 5. Clean Up Resources

Always close response bodies and channels:

```go
defer resp.Body.Close()
defer close(events)
```

### 6. Handle Streaming Edge Cases

- Empty chunks
- Malformed JSON
- Connection interruptions
- Context cancellation

### 7. Document Your Code

Provide clear documentation for:
- Configuration options
- Supported models
- Error conditions
- Usage examples

## Example: Complete Provider

See the complete example in `/internal/provider/anthropic/` for reference:

- `anthropic.go` - Main provider implementation
- `types.go` - API types and structures
- `sse.go` - Server-Sent Events parsing
- `thinking.go` - Extended thinking mode support

## Resources

- [Provider Interface](../../internal/provider/provider.go)
- [Base Provider](../../internal/provider/base.go)
- [Anthropic Implementation](../../internal/provider/anthropic/)
- [Provider Options](../../internal/provider/options.go)

---

**Last Updated**: 2025-01-05
