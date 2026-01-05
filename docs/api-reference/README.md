# AINative Code API Reference

Complete API reference documentation for the AINative Code CLI tool and libraries.

## Overview

AINative Code is a comprehensive CLI tool for AI-assisted software development, providing:

- **Multi-Provider LLM Integration**: Support for Anthropic Claude, OpenAI, Google Gemini, AWS Bedrock, Azure, and local models via Ollama
- **Secure Authentication**: OAuth 2.0 with PKCE flow, JWT token management, and OS keychain integration
- **Session Management**: Persistent conversation sessions with message history and export capabilities
- **Tool Execution Framework**: Extensible system for filesystem, network, system, and database operations
- **MCP Integration**: Model Context Protocol support for external tool servers
- **Database Layer**: SQLite-based storage with SQLC-generated type-safe queries
- **Configuration Management**: YAML-based config with environment variable overrides
- **Error Handling**: Comprehensive error types with recovery strategies and user-friendly messages

## How to Use This Reference

This API reference is organized by functional area:

1. **[Core Packages](core-packages.md)** - Main client, configuration, session management, and database APIs
2. **[Providers](providers.md)** - LLM provider interfaces and implementations
3. **[Tools](tools.md)** - Tool execution framework and built-in tools
4. **[Authentication](authentication.md)** - OAuth, JWT, and keychain integration
5. **[Configuration](configuration.md)** - Configuration structure and loading
6. **[Events](events.md)** - Event system and streaming (when implemented)
7. **[Errors](errors.md)** - Error types, codes, and recovery strategies

Each section includes:
- Function signatures with parameter types and return values
- Complete, working code examples
- Error conditions and handling
- Best practices and common use cases
- Links to related APIs

## Quick Start Guide

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/AINative-studio/ainative-code/internal/client"
    "github.com/AINative-studio/ainative-code/internal/config"
    "github.com/AINative-studio/ainative-code/internal/database"
)

func main() {
    ctx := context.Background()

    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize database
    db, err := database.Initialize(&database.ConnectionConfig{
        Driver: "sqlite3",
        DSN:    "ainative.db",
    })
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()

    // Create API client
    apiClient := client.New(
        client.WithBaseURL("https://api.ainative.studio"),
        client.WithTimeout(30 * time.Second),
    )

    // Make API request
    response, err := apiClient.Get(ctx, "/api/v1/health")
    if err != nil {
        log.Fatalf("API request failed: %v", err)
    }

    log.Printf("API Response: %s", response)
}
```

### Using Providers

```go
package main

import (
    "context"
    "log"

    "github.com/AINative-studio/ainative-code/internal/provider"
    "github.com/AINative-studio/ainative-code/internal/config"
)

func main() {
    ctx := context.Background()

    // Get provider from registry
    p, err := provider.GetRegistry().Get("anthropic")
    if err != nil {
        log.Fatalf("Failed to get provider: %v", err)
    }

    // Send chat request
    messages := []provider.Message{
        {Role: "user", Content: "Hello, how are you?"},
    }

    response, err := p.Chat(ctx, messages)
    if err != nil {
        log.Fatalf("Chat request failed: %v", err)
    }

    log.Printf("Response: %s", response.Content)
    log.Printf("Tokens used: %d", response.Usage.TotalTokens)
}
```

### Authentication

```go
package main

import (
    "context"
    "log"

    "github.com/AINative-studio/ainative-code/internal/auth"
)

func main() {
    ctx := context.Background()

    // Create auth client
    authClient, err := auth.NewClient(auth.DefaultClientOptions())
    if err != nil {
        log.Fatalf("Failed to create auth client: %v", err)
    }

    // Check for stored tokens
    tokens, err := authClient.GetStoredTokens(ctx)
    if err != nil || !tokens.IsValid() {
        // Authenticate user (opens browser)
        tokens, err = authClient.Authenticate(ctx)
        if err != nil {
            log.Fatalf("Authentication failed: %v", err)
        }
    }

    // Refresh if needed
    if tokens.NeedsRefresh() {
        tokens, err = authClient.RefreshToken(ctx, tokens.RefreshToken)
        if err != nil {
            log.Fatalf("Token refresh failed: %v", err)
        }
    }

    log.Printf("Authenticated as: %s", tokens.AccessToken.Email)
}
```

## Package Index

### Core Packages

| Package | Description | Documentation |
|---------|-------------|---------------|
| `internal/client` | HTTP client for AINative platform APIs | [Core Packages](core-packages.md#client) |
| `internal/config` | Configuration management and loading | [Configuration](configuration.md) |
| `internal/session` | Session and message management | [Core Packages](core-packages.md#session) |
| `internal/database` | Database layer with SQLC queries | [Core Packages](core-packages.md#database) |

### Provider Packages

| Package | Description | Documentation |
|---------|-------------|---------------|
| `internal/provider` | Provider interface and registry | [Providers](providers.md) |
| `internal/providers/*` | Provider implementations (Anthropic, OpenAI, etc.) | [Providers](providers.md#implementations) |

### Tool Packages

| Package | Description | Documentation |
|---------|-------------|---------------|
| `internal/tools` | Tool execution framework | [Tools](tools.md) |
| `internal/mcp` | Model Context Protocol client | [Tools](tools.md#mcp-integration) |

### Authentication Packages

| Package | Description | Documentation |
|---------|-------------|---------------|
| `internal/auth` | OAuth 2.0 and JWT authentication | [Authentication](authentication.md) |

### Support Packages

| Package | Description | Documentation |
|---------|-------------|---------------|
| `internal/errors` | Error types and handling | [Errors](errors.md) |
| `internal/logger` | Structured logging with zerolog | See source code |
| `internal/cache` | Response caching layer | See source code |
| `internal/ratelimit` | Rate limiting middleware | See source code |

## Common Patterns

### Error Handling

```go
import "github.com/AINative-studio/ainative-code/internal/errors"

// Check error type
if errors.GetCode(err) == errors.ErrCodeProviderTimeout {
    log.Println("Provider timed out, retrying...")
}

// Check if retryable
if errors.IsRetryable(err) {
    // Implement retry logic
}

// Get user-friendly message
if baseErr, ok := err.(*errors.BaseError); ok {
    fmt.Println(baseErr.UserMessage())
}
```

### Using Context

```go
import (
    "context"
    "time"
)

// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Use context in API calls
response, err := client.Get(ctx, "/api/endpoint")
```

### Configuration Override

```go
import "github.com/AINative-studio/ainative-code/internal/config"

// Load config from file
cfg, _ := config.LoadFromFile("config.yaml")

// Override with environment variables
cfg = config.ApplyEnvOverrides(cfg)

// Validate configuration
if err := config.Validate(cfg); err != nil {
    log.Fatalf("Invalid config: %v", err)
}
```

## Version Compatibility

This API reference documents the current version of AINative Code. For version-specific documentation, see the [releases documentation](../releases/).

## Contributing

For guidelines on contributing to the AINative Code project, see the main [README.md](../../README.md) and [development documentation](../development/).

## Support

- **Documentation**: [docs/](../)
- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions

## License

See [LICENSE](../../LICENSE) file for details.
