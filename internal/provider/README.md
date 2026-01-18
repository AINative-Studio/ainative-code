# Provider Selection Package

Intelligent provider selection logic for AINative platform that determines which LLM provider to use based on user preferences, available credits, and provider capabilities.

## Overview

The `provider` package implements sophisticated provider selection with the following features:

- **User Preference Selection**: Respects user's preferred LLM provider
- **Credit-Aware Selection**: Warns users about low credits and prevents zero-credit requests
- **Capability-Based Selection**: Matches provider capabilities with request requirements
- **Automatic Fallback**: Falls back to alternative providers when preferred one doesn't meet requirements
- **Provider Availability Checking**: Validates provider availability before selection

## Package Structure

```
internal/provider/
├── selector.go              # Core provider selection logic
├── selector_test.go         # Comprehensive test suite (23 tests)
├── selector_example_test.go # Usage examples
├── types.go                 # Type definitions
├── config.go                # Provider capability configurations
└── README.md                # This file
```

## Core Types

### ProviderInfo

Represents an LLM provider with its capabilities:

```go
type ProviderInfo struct {
    Name                    string  // Provider identifier (e.g., "anthropic")
    DisplayName             string  // Human-readable name (e.g., "Anthropic Claude")
    SupportsVision          bool    // Vision/image analysis capability
    SupportsFunctionCalling bool    // Function/tool calling capability
    SupportsStreaming       bool    // Streaming response capability
    MaxTokens               int     // Maximum context window size
    LowCreditWarning        bool    // Set when user has low credits
}
```

### User

Represents a user with credit balance:

```go
type User struct {
    Email   string  // User email
    Credits int     // Available credits
    Tier    string  // Subscription tier (free, pro, etc.)
}
```

### SelectionRequest

Defines capability requirements for selection:

```go
type SelectionRequest struct {
    Model                   string  // Model identifier or "auto"
    RequiresVision          bool    // Requires vision capability
    RequiresFunctionCalling bool    // Requires function calling
    RequiresStreaming       bool    // Requires streaming
}
```

## Usage Examples

### Basic Selection by User Preference

```go
selector := provider.NewSelector(
    provider.WithProviders("anthropic", "openai", "google"),
    provider.WithUserPreference("anthropic"),
)

selectedProvider, err := selector.Select(context.Background(), nil)
if err != nil {
    // Handle error
}

fmt.Println(selectedProvider.Name) // "anthropic"
```

### Credit-Aware Selection

```go
user := &provider.User{
    Email:   "user@example.com",
    Credits: 10,
    Tier:    "free",
}

selector := provider.NewSelector(
    provider.WithProviders("anthropic", "openai"),
    provider.WithCreditThreshold(50), // Warn if credits < 50
)

selectedProvider, err := selector.Select(context.Background(), user)
if err != nil {
    // Handle insufficient credits
}

if selectedProvider.LowCreditWarning {
    // Show warning to user
    fmt.Println("Warning: Low credits remaining")
}
```

### Capability-Based Selection

```go
req := &provider.SelectionRequest{
    RequiresVision:          true,
    RequiresFunctionCalling: true,
    Model:                   "auto",
}

selector := provider.NewSelector(
    provider.WithProviders("anthropic", "openai", "google"),
)

selectedProvider, err := selector.Select(context.Background(), nil, req)
if err != nil {
    // Handle no matching provider
}

// Provider guaranteed to support vision and function calling
```

### Advanced Selection with Multiple Requirements

```go
user := &provider.User{
    Email:   "premium@example.com",
    Credits: 1000,
    Tier:    "pro",
}

req := &provider.SelectionRequest{
    RequiresVision:          true,
    RequiresFunctionCalling: true,
    RequiresStreaming:       true,
    Model:                   "auto",
}

selector := provider.NewSelector(
    provider.WithProviders("anthropic", "openai", "google"),
    provider.WithUserPreference("google"),
    provider.WithCreditThreshold(100),
)

selectedProvider, err := selector.Select(context.Background(), user, req)
// Returns Google if it meets all requirements, falls back otherwise
```

### Check Provider Availability

```go
selector := provider.NewSelector(
    provider.WithProviders("anthropic", "openai"),
)

if selector.IsAvailable("anthropic") {
    // Anthropic is available
}

if !selector.IsAvailable("google") {
    // Google is not in the available providers list
}
```

## Configuration Options

### WithProviders

Sets the list of available providers:

```go
provider.WithProviders("anthropic", "openai", "google")
```

### WithUserPreference

Sets the user's preferred provider:

```go
provider.WithUserPreference("anthropic")
```

### WithCreditThreshold

Sets the credit threshold for low credit warnings (default: 50):

```go
provider.WithCreditThreshold(100)
```

### WithFallback

Enables/disables fallback to alternative providers (default: true):

```go
provider.WithFallback(false) // Disable fallback
```

## Provider Capabilities

The package currently supports three providers with the following capabilities:

| Provider | Vision | Function Calling | Streaming | Max Tokens |
|----------|--------|------------------|-----------|------------|
| Anthropic Claude | Yes | Yes | Yes | 200,000 |
| OpenAI GPT | Yes | Yes | Yes | 128,000 |
| Google Gemini | Yes | Yes | Yes | 1,000,000 |

## Error Handling

The package defines two main errors:

```go
var (
    ErrInsufficientCredits = errors.New("insufficient credits")
    ErrNoProviderAvailable = errors.New("no provider available")
)
```

Example error handling:

```go
selectedProvider, err := selector.Select(ctx, user, req)
if err != nil {
    if errors.Is(err, provider.ErrInsufficientCredits) {
        // Show upgrade prompt to user
    } else if errors.Is(err, provider.ErrNoProviderAvailable) {
        // No provider meets requirements
    }
    return err
}
```

## Selection Logic Flow

1. **Credit Check**: If user has zero credits, return `ErrInsufficientCredits`
2. **Provider Availability**: Check if any providers are configured
3. **User Preference**: Try preferred provider first if set
4. **Capability Matching**: Verify provider meets request requirements
5. **Fallback**: Select alternative provider if preferred doesn't match
6. **Credit Warning**: Set `LowCreditWarning` if credits below threshold

## Test Coverage

The package has comprehensive test coverage:

- **Overall Coverage**: 93.4%
- **Total Tests**: 23 unit tests + 5 example tests
- **Test Categories**:
  - User preference selection (2 tests)
  - Credit-aware selection (3 tests)
  - Capability matching (5 tests)
  - Fallback logic (1 test)
  - Availability checking (3 tests)
  - Default behavior (2 tests)
  - Edge cases (3 tests)
  - Configuration (2 tests)
  - Validation (2 tests)

Run tests:

```bash
go test -v ./internal/provider/
go test -cover ./internal/provider/
```

## Implementation Details

### Test-Driven Development (TDD)

This package was developed using strict TDD methodology:

1. **RED Phase**: Wrote 23 failing tests first
2. **GREEN Phase**: Implemented minimal code to make tests pass
3. **REFACTOR Phase**: Cleaned up code while maintaining green tests
4. **Result**: 93.4% coverage, all tests passing

### Thread Safety

The `Selector` is read-only after construction, making it safe for concurrent use across goroutines.

### Performance

Provider selection is O(n) where n is the number of configured providers. Capability lookups use maps for O(1) access.

## Integration with Backend

The selector integrates with the Python backend for actual provider communication:

```go
import (
    "github.com/AINative-studio/ainative-code/internal/backend"
    "github.com/AINative-studio/ainative-code/internal/provider"
)

// Select provider
selector := provider.NewSelector(/* ... */)
selectedProvider, err := selector.Select(ctx, user, req)

// Use with backend client
client := backend.NewClient(backendURL)
response, err := client.Chat(ctx, &backend.ChatRequest{
    Provider: selectedProvider.Name,
    Messages: messages,
})
```

## Future Enhancements

Potential improvements for future iterations:

- [ ] Provider health checking and circuit breakers
- [ ] Cost-based provider selection
- [ ] Load balancing across providers
- [ ] Provider-specific model selection
- [ ] Dynamic capability discovery from backend
- [ ] Rate limit awareness
- [ ] Provider performance metrics

## Contributing

When extending this package:

1. **Write tests first** following TDD methodology
2. **Maintain >80% coverage** for all new code
3. **Run `go fmt` and `go vet`** before committing
4. **Update README** with new features
5. **Add examples** for new functionality

## License

Copyright 2026 AINative Studio. All rights reserved.
