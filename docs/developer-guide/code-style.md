# Code Style Guide

## Overview

This guide defines the coding standards and best practices for AINative Code. Following these guidelines ensures consistency, maintainability, and readability across the codebase.

## Go Standards

We follow standard Go conventions and best practices:

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

## Formatting

### Automatic Formatting

All code must be formatted with `gofmt`:

```bash
# Format all code
make fmt

# Or manually
gofmt -s -w .

# Check formatting
make fmt-check
```

### Line Length

- **Preferred**: 80-100 characters
- **Maximum**: 120 characters
- Break long lines at logical points

**Good**:
```go
response, err := provider.Chat(
    ctx,
    messages,
    WithMaxTokens(4096),
    WithTemperature(0.7),
)
```

**Bad**:
```go
response, err := provider.Chat(ctx, messages, WithMaxTokens(4096), WithTemperature(0.7), WithTopP(0.9))
```

### Indentation

- Use **tabs** for indentation (Go standard)
- Use **spaces** for alignment
- `gofmt` handles this automatically

## Naming Conventions

### Packages

**Rules**:
- Short, concise, lowercase
- Single word preferred
- No underscores or mixedCaps
- Should be noun, not verb

**Good**:
```go
package provider
package auth
package logger
package config
```

**Bad**:
```go
package llm_provider
package authentication_handler
package myPackage
```

### Files

**Rules**:
- Lowercase with underscores
- Match package or purpose
- Test files end with `_test.go`

**Examples**:
```
provider.go
provider_test.go
anthropic_provider.go
options.go
registry.go
```

### Types and Interfaces

**Rules**:
- PascalCase (exported) or camelCase (unexported)
- Interfaces: noun or adjective
- Interface names should describe what it does
- Single-method interfaces: verb + "er" suffix

**Good**:
```go
type Provider interface { }
type Message struct { }
type ChatOption func(*chatConfig)

// Single-method interface
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}
```

**Bad**:
```go
type IProvider interface { }  // No "I" prefix
type provider_type struct { } // No underscores
type ProviderInterface interface { } // No "Interface" suffix
```

### Functions and Methods

**Rules**:
- PascalCase (exported) or camelCase (unexported)
- Start with verb
- Be descriptive but concise

**Good**:
```go
func NewProvider(config Config) *Provider { }
func (p *Provider) Chat(ctx context.Context, messages []Message) (Response, error) { }
func validateMessage(msg Message) error { }
func parseToken(token string) (*Claims, error) { }
```

**Bad**:
```go
func provider(config Config) *Provider { } // Not descriptive
func (p *Provider) DoChat(ctx context.Context, messages []Message) (Response, error) { } // Unnecessary "Do"
func validate_message(msg Message) error { } // Underscores
```

### Variables

**Rules**:
- PascalCase (exported) or camelCase (unexported)
- Descriptive but not verbose
- Single letter for short-lived locals (i, j, k, v)
- Avoid generic names like "data", "info"

**Good**:
```go
var maxRetries = 3
var defaultTimeout = 30 * time.Second

func processMessages(ctx context.Context, messages []Message) {
    for i, msg := range messages {
        // ...
    }
}
```

**Bad**:
```go
var MAX_RETRIES = 3  // Use camelCase, not SCREAMING_SNAKE_CASE
var t = 30 * time.Second  // Not descriptive

func processMessages(ctx context.Context, messages []Message) {
    for index, message := range messages {  // Too verbose for simple loop
        // ...
    }
}
```

### Constants

**Rules**:
- PascalCase (exported) or camelCase (unexported)
- Group related constants in blocks
- Use iota for enumerations

**Good**:
```go
const (
    maxTokens        = 4096
    defaultModel     = "claude-3-5-sonnet-20241022"
    maxRetries       = 3
)

const (
    StatusPending Status = iota
    StatusInProgress
    StatusCompleted
    StatusFailed
)
```

**Bad**:
```go
const MAX_TOKENS = 4096  // Don't use SCREAMING_SNAKE_CASE
const default_model = "claude-3-5-sonnet-20241022"  // No underscores
```

## Error Handling

### Error Creation

**Use `fmt.Errorf` with `%w` for wrapping**:
```go
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}
```

**Custom error types for specific cases**:
```go
type ProviderError struct {
    Provider string
    Code     string
    Message  string
    Err      error
}

func (e *ProviderError) Error() string {
    return fmt.Sprintf("%s provider error [%s]: %s", e.Provider, e.Code, e.Message)
}

func (e *ProviderError) Unwrap() error {
    return e.Err
}
```

### Error Checking

**Good**:
```go
response, err := provider.Chat(ctx, messages)
if err != nil {
    return fmt.Errorf("chat failed: %w", err)
}
return response, nil
```

**Bad**:
```go
response, _ := provider.Chat(ctx, messages)  // Never ignore errors
return response, nil

// Or
response, err := provider.Chat(ctx, messages)
if err != nil {
    panic(err)  // Don't panic in library code
}
```

### Error Messages

**Rules**:
- Lowercase (except proper nouns)
- No punctuation at the end
- Include context
- Use %w for error wrapping

**Good**:
```go
return fmt.Errorf("failed to parse JWT token: %w", err)
return fmt.Errorf("invalid API key for provider %s", providerName)
return fmt.Errorf("database query failed: %w", err)
```

**Bad**:
```go
return fmt.Errorf("Error: Failed to parse JWT token: %v.", err)  // Capitalized, punctuation, %v
return fmt.Errorf("error")  // Not descriptive
return errors.New("Invalid API key")  // Capitalized, no context
```

## Comments and Documentation

### Package Documentation

Every package must have a package comment:

```go
// Package provider defines the interface for LLM providers and includes
// implementations for various AI providers like Anthropic, OpenAI, and others.
//
// The main interface is Provider, which defines methods for chat completion
// and streaming responses. Each provider implementation handles authentication,
// API communication, and error handling specific to that provider.
//
// Example usage:
//
//	provider := anthropic.NewProvider(config)
//	response, err := provider.Chat(ctx, messages)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(response.Content)
package provider
```

### Function Documentation

All exported functions must have documentation:

```go
// NewProvider creates a new Anthropic provider instance with the given configuration.
// It returns an error if the configuration is invalid or if the API key is missing.
//
// The provider will use the configured model, or default to claude-3-5-sonnet-20241022
// if no model is specified.
//
// Example:
//
//	config := Config{
//	    APIKey: "sk-ant-...",
//	    Model: "claude-3-opus-20240229",
//	}
//	provider, err := NewProvider(config)
//	if err != nil {
//	    return err
//	}
func NewProvider(config Config) (*Provider, error) {
    // Implementation
}
```

### Type Documentation

All exported types must be documented:

```go
// Provider defines the interface for LLM providers.
// All provider implementations must satisfy this interface.
type Provider interface {
    // Chat sends a complete chat request and waits for the full response.
    // It returns an error if the request fails or if the context is cancelled.
    Chat(ctx context.Context, messages []Message, opts ...ChatOption) (Response, error)

    // Stream sends a streaming chat request and returns a channel for events.
    // The channel will be closed when the stream completes or encounters an error.
    Stream(ctx context.Context, messages []Message, opts ...StreamOption) (<-chan Event, error)
}

// Message represents a chat message with a role and content.
type Message struct {
    // Role is the message sender: "user", "assistant", or "system"
    Role string

    // Content is the message text
    Content string
}
```

### Inline Comments

Use inline comments for complex logic:

```go
func parseToken(token string) (*Claims, error) {
    // Split token into parts: header.payload.signature
    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        return nil, errors.New("invalid token format")
    }

    // Decode base64 payload
    payload, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return nil, fmt.Errorf("failed to decode payload: %w", err)
    }

    // Parse JSON claims
    var claims Claims
    if err := json.Unmarshal(payload, &claims); err != nil {
        return nil, fmt.Errorf("failed to parse claims: %w", err)
    }

    return &claims, nil
}
```

**Don't state the obvious**:
```go
// Bad: Comment just restates the code
// Set i to 0
i := 0

// Good: Comment explains why
// Start from beginning of slice to process all items
i := 0
```

## Function Design

### Function Signature

**Keep parameters to a minimum**:
```go
// Good: Uses options pattern
func Chat(ctx context.Context, messages []Message, opts ...ChatOption) (Response, error)

// Bad: Too many parameters
func Chat(ctx context.Context, messages []Message, maxTokens int, temperature float64, topP float64, model string) (Response, error)
```

### Return Values

**Rules**:
- Return errors as the last return value
- Use named returns sparingly (only for documentation)
- Don't mix named and unnamed returns

**Good**:
```go
func LoadConfig(path string) (*Config, error) {
    // Implementation
}

// Named returns for documentation
func ParseToken(token string) (claims *Claims, err error) {
    // Implementation
}
```

**Bad**:
```go
func LoadConfig(path string) (error, *Config) {  // Error should be last
    // Implementation
}

func ParseToken(token string) (claims *Claims, error) {  // Mixed named/unnamed
    // Implementation
}
```

### Function Length

- **Target**: 20-50 lines
- **Maximum**: 100 lines
- Break large functions into smaller helpers

**Good**:
```go
func ProcessMessage(msg Message) (Response, error) {
    if err := validateMessage(msg); err != nil {
        return Response{}, err
    }

    content := sanitizeContent(msg.Content)
    tokens := tokenizeContent(content)

    return generateResponse(tokens)
}

func validateMessage(msg Message) error { /* ... */ }
func sanitizeContent(content string) string { /* ... */ }
func tokenizeContent(content string) []Token { /* ... */ }
```

## Package Organization

### Internal vs Pkg

**internal/**: Private application code, not importable by other projects
```go
internal/
├── auth/         // Authentication logic
├── provider/     // LLM providers
├── config/       // Configuration
└── database/     // Database operations
```

**pkg/**: Public library code, importable by other projects
```go
pkg/
└── types/        // Shared types
```

### File Organization

Group related functionality in the same file:

```go
// provider.go - Interface and core types
type Provider interface { }
type Message struct { }
type Response struct { }

// options.go - Configuration options
type ChatOption func(*chatConfig)
func WithMaxTokens(tokens int) ChatOption { }
func WithTemperature(temp float64) ChatOption { }

// registry.go - Provider registry
type Registry struct { }
func NewRegistry() *Registry { }
func (r *Registry) Register(name string, provider Provider) { }

// anthropic/provider.go - Anthropic implementation
type AnthropicProvider struct { }
func NewProvider(config Config) *AnthropicProvider { }
func (p *AnthropicProvider) Chat(...) { }
```

### Import Grouping

Group imports into three blocks:

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // External dependencies
    "github.com/anthropics/anthropic-sdk-go"
    "github.com/spf13/viper"

    // Internal packages
    "github.com/AINative-studio/ainative-code/internal/logger"
    "github.com/AINative-studio/ainative-code/internal/provider"
)
```

## Struct Design

### Field Order

1. Exported fields first
2. Unexported fields second
3. Group related fields
4. Most important fields first

```go
type Provider struct {
    // Exported fields
    Name   string
    Models []string

    // Configuration
    APIKey      string
    Endpoint    string
    MaxRetries  int

    // Internal state
    client      *http.Client
    rateLimiter *rate.Limiter
    mu          sync.Mutex
}
```

### Embedding

Use embedding for composition:

```go
type BaseProvider struct {
    name   string
    models []string
}

type AnthropicProvider struct {
    BaseProvider  // Embed base

    apiKey string
    client *http.Client
}
```

### Tags

Use tags for JSON, YAML, validation:

```go
type Config struct {
    Provider    string  `yaml:"provider" json:"provider" validate:"required"`
    APIKey      string  `yaml:"api_key" json:"api_key" validate:"required"`
    Model       string  `yaml:"model" json:"model"`
    MaxTokens   int     `yaml:"max_tokens" json:"max_tokens" validate:"min=1,max=100000"`
    Temperature float64 `yaml:"temperature" json:"temperature" validate:"min=0,max=2"`
}
```

## Error Handling Patterns

### Sentinel Errors

```go
var (
    ErrInvalidToken   = errors.New("invalid token")
    ErrTokenExpired   = errors.New("token expired")
    ErrUnauthorized   = errors.New("unauthorized")
)

// Usage
if err := validateToken(token); err != nil {
    if errors.Is(err, ErrTokenExpired) {
        // Handle expired token
    }
    return err
}
```

### Error Wrapping

```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    if err := validateConfig(&config); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }

    return &config, nil
}
```

### Error Types

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error for field %q: %s", e.Field, e.Message)
}

// Usage with errors.As
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Printf("Field %s failed validation\n", valErr.Field)
}
```

## Concurrency Patterns

### Goroutine Management

```go
func Stream(ctx context.Context, messages []Message) (<-chan Event, error) {
    events := make(chan Event, 10)  // Buffered channel

    go func() {
        defer close(events)  // Always close channel

        for {
            select {
            case <-ctx.Done():
                events <- Event{Type: EventTypeError, Error: ctx.Err()}
                return
            case event := <-source:
                events <- event
            }
        }
    }()

    return events, nil
}
```

### Mutexes

```go
type Registry struct {
    mu        sync.RWMutex
    providers map[string]Provider
}

func (r *Registry) Get(name string) (Provider, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    p, ok := r.providers[name]
    return p, ok
}

func (r *Registry) Register(name string, provider Provider) {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.providers[name] = provider
}
```

### Context Usage

```go
func (p *Provider) Chat(ctx context.Context, messages []Message) (Response, error) {
    // Create request
    req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint, body)
    if err != nil {
        return Response{}, err
    }

    // Execute with context
    resp, err := p.client.Do(req)
    if err != nil {
        return Response{}, err
    }
    defer resp.Body.Close()

    // Check for cancellation
    select {
    case <-ctx.Done():
        return Response{}, ctx.Err()
    default:
    }

    return parseResponse(resp)
}
```

## Testing Style

See [Testing Guide](testing.md) for comprehensive testing guidelines.

### Test Naming

```go
func TestProviderChat(t *testing.T) { }
func TestProvider_Chat_ReturnsError_WhenAPIKeyInvalid(t *testing.T) { }
func TestConfigLoader_LoadFromFile_ParsesYAML(t *testing.T) { }
```

### Test Organization

```go
func TestFunction(t *testing.T) {
    // Arrange
    setup()

    // Act
    result := doSomething()

    // Assert
    assert.Equal(t, expected, result)
}
```

## Linter Configuration

We use golangci-lint with strict settings. See `.golangci.yml` for full configuration.

### Running Linter

```bash
# Run linter
make lint

# Auto-fix issues
golangci-lint run --fix
```

### Common Linter Issues

**Unchecked errors** (errcheck):
```go
// Bad
file.Close()

// Good
defer file.Close()

// Or explicitly
if err := file.Close(); err != nil {
    log.Printf("failed to close file: %v", err)
}
```

**Shadowing** (govet):
```go
// Bad
if err := doSomething(); err != nil {
    err := doSomethingElse()  // Shadows err
}

// Good
if err := doSomething(); err != nil {
    if err2 := doSomethingElse(); err2 != nil {
        // Handle
    }
}
```

## Best Practices

### 1. Prefer Composition Over Inheritance

Use embedding and interfaces instead of classical inheritance.

### 2. Accept Interfaces, Return Structs

```go
// Good
func ProcessMessages(reader io.Reader) (*Result, error) { }

// Not ideal
func ProcessMessages(reader *os.File) (*Result, error) { }
```

### 3. Use context.Context

Always pass context.Context as the first parameter for operations that can be cancelled or have deadlines.

### 4. Handle Errors Immediately

Don't accumulate errors; handle them as they occur.

### 5. Keep Code Simple

Prefer simple, readable code over clever optimizations unless profiling shows a real need.

---

**Last Updated**: 2025-01-05
