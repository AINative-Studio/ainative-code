# API Reference

This directory contains comprehensive API reference documentation for AINative Code.

## Overview

AINative Code provides several API layers:

1. **Provider Interface** - Abstract interface for LLM providers
2. **Internal APIs** - Core application services
3. **Platform APIs** - AINative platform integrations
4. **Tool APIs** - Built-in and custom tool interfaces

## Core Documentation

### [Provider Interface](provider-interface.md)

Complete specification for LLM provider implementations.

**Topics Covered**:
- Interface definition and contracts
- Message format and types
- Response structures
- Streaming implementation
- Error handling
- Testing requirements

**Read this if**: You're implementing a new provider or understanding how providers work.

## Provider Implementations

### Supported Providers

| Provider | Status | Documentation | Models |
|----------|--------|---------------|---------|
| **Anthropic Claude** | ✅ Production | [Anthropic Docs](../providers/anthropic.md) | Claude 3.5 Sonnet, Claude 3 Opus, Claude 3 Haiku |
| **OpenAI** | ✅ Production | [OpenAI Docs](../OPENAI_PROVIDER.md) | GPT-4, GPT-4 Turbo, GPT-3.5 Turbo |
| **Google Gemini** | ✅ Production | [Gemini Docs](../providers/gemini.md) | Gemini Pro, Gemini Ultra |
| **AWS Bedrock** | ✅ Production | [Bedrock Docs](../providers/bedrock.md) | Claude on Bedrock, Titan |
| **Azure OpenAI** | ✅ Production | [Azure Docs](../providers/azure.md) | GPT-4, GPT-3.5 (Azure-hosted) |
| **Ollama** | ✅ Production | [Ollama Docs](../ollama-provider.md) | Llama 3, Mistral, CodeLlama, etc. |

### Provider Features

| Feature | Anthropic | OpenAI | Gemini | Bedrock | Azure | Ollama |
|---------|-----------|--------|--------|---------|-------|--------|
| Streaming | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Tool Use | ✅ | ✅ | ✅ | ✅ | ✅ | ⚠️ |
| Vision | ✅ | ✅ | ✅ | ✅ | ✅ | ⚠️ |
| Prompt Caching | ✅ | ⚠️ | ❌ | ✅ | ⚠️ | ❌ |
| Extended Thinking | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |

✅ = Fully supported, ⚠️ = Partial support, ❌ = Not supported

## Internal APIs

### Configuration API

Hierarchical configuration system with dynamic resolution.

**Key Features**:
- YAML configuration files
- Environment variable support
- Dynamic API key resolution (password managers)
- Configuration validation
- Hot reload capabilities

**Documentation**: [Configuration Guide](../configuration.md)

### Database API

Type-safe SQLite database operations using SQLC.

**Key Features**:
- Session persistence
- Message storage
- Metadata management
- Migration support
- Transaction handling

**Documentation**: [Database Guide](../database-guide.md)

### Logging API

High-performance structured logging system.

**Key Features**:
- Multiple log levels (DEBUG, INFO, WARN, ERROR, FATAL)
- JSON and console formats
- Context-aware logging
- Log rotation
- Zero-allocation for disabled levels

**Documentation**: [Logging Guide](../logging.md)

**Performance**:
- Simple log: ~2μs
- Structured fields: ~3μs
- Disabled level: <1ns (zero allocation)

### Authentication API

Multi-tier authentication for platform and providers.

**Key Features**:
- OAuth 2.0 with PKCE for AINative platform
- API key management for LLM providers
- Token caching and refresh
- Secure keychain storage
- Three-tier validation (local RSA → API → offline)

**Documentation**: [Authentication Guide](../authentication/README.md)

### Session API

Conversation persistence and management.

**Key Features**:
- Create and manage sessions
- Message persistence
- Session metadata tracking
- Export/import (JSON, Markdown)
- Search and filtering

**Example**:
```go
// Create session
session, err := sessionMgr.Create("My Feature Implementation")

// Add message
err = sessionMgr.AddMessage(session.ID, Message{
    Role:    RoleUser,
    Content: "Help me implement OAuth",
})

// Load session
session, err = sessionMgr.Load(sessionID)

// Export session
data, err := sessionMgr.Export(sessionID, "markdown")
```

### Event API

Asynchronous event bus for system events.

**Event Types**:
- Message events (sent, received, streaming)
- Session events (created, updated, deleted)
- Provider events (switched, error)
- Auth events (login, logout, refresh)
- Tool events (executed, completed, error)

**Example**:
```go
// Subscribe to events
eventBus.Subscribe(EventMessageReceived, func(event Event) {
    log.Info("Message received", "session", event.Data)
})

// Publish event
eventBus.Publish(Event{
    Type: EventMessageReceived,
    Data: message,
})
```

## Platform APIs

### ZeroDB Client

Integration with ZeroDB for vector search, NoSQL tables, and PostgreSQL.

**Features**:
- Vector operations (upsert, search, delete)
- NoSQL table management
- Agent memory storage
- Quantum-enhanced search (advanced)
- PostgreSQL operations

**Documentation**: [ZeroDB Guide](../zerodb/README.md)

### Design Token Client

Extract and manage design tokens.

**Features**:
- Extract tokens from CSS/SCSS/JSON
- Upload to Design platform
- Generate themes
- Component analysis
- Multiple output formats

**Documentation**: [Design Token Guide](../design-token-upload.md)

### Strapi Client

Content management via Strapi CMS.

**Features**:
- Blog post management
- Tutorial management
- Event management
- Content querying
- Markdown support

**Documentation**: [Strapi Guide](../strapi-blog.md)

### RLHF Client

Collect and submit feedback for model improvement.

**Features**:
- Interaction feedback
- Agent feedback
- Workflow feedback
- Error reporting
- Anonymous collection

## Tool APIs

### Built-in Tools

#### Bash Tool
Execute shell commands with sandboxing.

```go
type BashTool interface {
    Execute(ctx context.Context, command string) (*ToolResult, error)
}
```

#### File Operations Tool
Read, write, and manage files.

```go
type FileOpsTool interface {
    Read(path string) ([]byte, error)
    Write(path string, content []byte) error
    Delete(path string) error
}
```

#### Grep Tool
Search for patterns in files.

```go
type GrepTool interface {
    Search(pattern string, paths []string) ([]Match, error)
}
```

#### Web Fetch Tool
Retrieve and analyze web content.

```go
type WebFetchTool interface {
    Fetch(url string) (string, error)
}
```

### MCP (Model Context Protocol) Integration

Integrate custom tools via MCP servers.

**Supported Transports**:
- stdio (process communication)
- HTTP (REST API)
- SSE (Server-Sent Events)

**Example Configuration**:
```yaml
mcp_servers:
  - name: "github"
    transport: "stdio"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_TOKEN: "${GITHUB_TOKEN}"
```

**Documentation**: [MCP Integration Guide](../development/mcp-integration.md)

## Error Handling

All APIs follow consistent error handling patterns:

```go
// Standard errors
var (
    ErrInvalidInput   = errors.New("invalid input")
    ErrNotFound       = errors.New("not found")
    ErrUnauthorized   = errors.New("unauthorized")
    ErrRateLimited    = errors.New("rate limited")
    ErrServerError    = errors.New("server error")
)

// Wrapped errors with context
err := fmt.Errorf("failed to create session: %w", ErrInvalidInput)

// Check error types
if errors.Is(err, ErrNotFound) {
    // Handle not found
}
```

## API Versioning

### Current Version: v1

APIs follow semantic versioning:
- **Major version** (v1, v2): Breaking changes
- **Minor version** (v1.1): New features, backward compatible
- **Patch version** (v1.1.1): Bug fixes, backward compatible

### Compatibility Promise

- v1.x releases maintain backward compatibility
- Deprecated features have 2 minor version grace period
- Breaking changes only in major versions

## Testing

All APIs include comprehensive tests:

```bash
# Run all API tests
make test

# Run specific API tests
go test ./internal/provider/...
go test ./internal/session/...
go test ./internal/auth/...

# Run with coverage
make test-coverage
```

## Development

### Adding a New Provider

1. Implement `Provider` interface
2. Add configuration schema
3. Implement message formatting
4. Add streaming support
5. Write comprehensive tests
6. Update documentation

See [Provider Interface](provider-interface.md) for details.

### Adding a New Tool

1. Implement `Tool` interface
2. Add parameter definitions
3. Implement safety checks
4. Add permission handling
5. Write tests
6. Document usage

## References

### Internal Documentation
- [Architecture](../architecture/README.md) - System architecture
- [Development Guide](../development/README.md) - Development setup
- [User Guide](../user-guide/README.md) - End-user documentation

### External Resources
- [Go Documentation](https://go.dev/doc/)
- [Anthropic API](https://docs.anthropic.com/)
- [OpenAI API](https://platform.openai.com/docs/api-reference)
- [Google Gemini API](https://ai.google.dev/docs)

---

**Last Updated**: January 2025
**Maintainer**: AINative Studio Engineering Team
