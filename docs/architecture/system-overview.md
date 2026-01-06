# AINative Code - System Architecture Overview

## Executive Summary

AINative Code is a terminal-based AI coding assistant built with Go that provides multi-provider LLM support, sophisticated terminal UI, and deep integration with the AINative platform ecosystem. This document describes the high-level architecture, key components, and design decisions.

## Architecture Principles

### 1. Modularity
- Clean separation of concerns across packages
- Well-defined interfaces between components
- Pluggable provider architecture for extensibility

### 2. Performance
- Efficient streaming of LLM responses
- Minimal memory footprint (<100MB idle)
- Fast startup time (<100ms)
- Asynchronous operations where possible

### 3. Reliability
- Comprehensive error handling
- Graceful degradation when services unavailable
- Automatic retry with exponential backoff
- Session persistence and recovery

### 4. Security
- Secure credential storage using OS keychains
- No logging of sensitive data
- TLS for all network communication
- Sandboxed command execution

### 5. Usability
- Intuitive terminal interface
- Clear error messages
- Comprehensive documentation
- Sensible defaults with customization options

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Interface                         │
│                    (Cobra Commands)                          │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────┴────────────────────────────────────┐
│                   Terminal UI Layer                          │
│                   (Bubble Tea TUI)                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │Chat Interface│  │Config Manager│  │Session Viewer│     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────┴────────────────────────────────────┐
│                    Core Business Logic                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │Provider Layer│  │Session Mgmt  │  │Tool Executor │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │Auth Manager  │  │Event System  │  │Cache Layer   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────┴────────────────────────────────────┐
│                   Infrastructure Layer                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │Config System │  │Logger        │  │Database (SQL)│     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
                         │
┌────────────────────────┴────────────────────────────────────┐
│                    External Services                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │LLM Providers │  │AINative Auth │  │ZeroDB        │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │Strapi CMS    │  │Design Tokens │                        │
│  └──────────────┘  └──────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. CLI Interface (`internal/cmd/`)

**Purpose**: Command-line interface and user interaction

**Key Features**:
- Command parsing and routing (Cobra)
- Input validation
- Help text generation
- Flag management

**Main Commands**:
- `chat` - Interactive conversation
- `config` - Configuration management
- `session` - Session operations
- `auth` - Authentication management
- `zerodb` - ZeroDB operations
- `design` - Design token operations
- `strapi` - CMS operations

### 2. Terminal UI Layer (`internal/tui/`)

**Purpose**: Rich terminal user interface

**Technology**: Bubble Tea (charmbracelet/bubbletea)

**Components**:
- Chat view with streaming responses
- Code block rendering with syntax highlighting
- Session list and navigation
- Configuration editor
- Status indicators and progress bars

**Key Features**:
- Real-time streaming display
- Keyboard shortcuts
- Mouse support
- Responsive layout
- Theming support

### 3. Provider Layer (`internal/provider/`, `internal/providers/`)

**Purpose**: Abstract LLM provider implementations

**Interface**:
```go
type Provider interface {
    Name() string
    Chat(ctx context.Context, messages []Message, opts ...Option) (*Response, error)
    Stream(ctx context.Context, messages []Message, opts ...Option) (<-chan StreamEvent, error)
    ValidateConfig(config *Config) error
}
```

**Implementations**:
- Anthropic (Claude 3.x)
- OpenAI (GPT-4, GPT-3.5)
- Google Gemini
- AWS Bedrock
- Azure OpenAI
- Ollama

**Features**:
- Automatic format translation
- Rate limiting and retry logic
- Error normalization
- Streaming support
- Cost tracking

### 4. Session Management (`internal/session/`)

**Purpose**: Conversation persistence and management

**Storage**: SQLite database

**Schema**:
- Sessions table (id, title, created_at, updated_at, metadata)
- Messages table (id, session_id, role, content, timestamp)
- Attachments table (id, message_id, type, data)

**Features**:
- Auto-save conversations
- Session resume
- Export/import (JSON, Markdown)
- Search and filtering
- Metadata tracking (tokens, cost, duration)

### 5. Authentication (`internal/auth/`)

**Purpose**: Multi-tier authentication system

**Components**:

#### OAuth 2.0 Flow (AINative Platform)
- Authorization Code flow with PKCE
- JWT access/refresh tokens
- Automatic token refresh
- Three-tier validation:
  1. Local RSA validation (cached public keys)
  2. API validation fallback
  3. Local auth fallback (offline)

#### API Key Management (LLM Providers)
- Dynamic resolution from password managers
- Environment variable support
- Secure storage in OS keychain
- Automatic re-resolution on 401

**Security Features**:
- HttpOnly cookies
- CSRF protection
- Rate limiting
- Audit logging

### 6. Tool Execution (`internal/tools/`)

**Purpose**: Execute tools requested by LLM

**Built-in Tools**:
- Bash execution (sandboxed)
- File operations (read, write, edit)
- Code search (grep, find)
- Web fetch
- Database queries

**Safety Features**:
- Permission prompts
- Timeout enforcement
- Working directory restrictions
- Command allowlist/blocklist

**MCP Integration**:
- Stdio transport
- HTTP transport
- Tool discovery
- Context injection

### 7. Configuration System (`internal/config/`)

**Purpose**: Hierarchical configuration management

**Technology**: Viper

**Configuration Sources** (priority order):
1. Command-line flags
2. Environment variables
3. Config file (~/.config/ainative-code/config.yaml)
4. Defaults

**Features**:
- YAML configuration
- Dynamic API key resolution
- Environment variable expansion
- Configuration validation
- Hot reload support

### 8. Event System (`internal/events/`)

**Purpose**: Asynchronous event handling

**Event Types**:
- Message events (sent, received, streaming)
- Session events (created, updated, deleted)
- Provider events (switched, error)
- Auth events (login, logout, refresh)
- Tool events (executed, completed, error)

**Features**:
- Pub/sub pattern
- Event filtering
- Async processing
- Event persistence (optional)

### 9. Database Layer (`internal/database/`)

**Purpose**: Local data persistence

**Technology**: SQLite with SQLC

**Features**:
- Type-safe queries (generated by SQLC)
- Migration management
- Connection pooling
- Transaction support
- Backup/restore

**Tables**:
- sessions
- messages
- attachments
- config_cache
- auth_tokens

### 10. Logging (`internal/logger/`)

**Purpose**: Structured logging system

**Technology**: zerolog with lumberjack rotation

**Features**:
- Multiple log levels (DEBUG, INFO, WARN, ERROR, FATAL)
- JSON and console formats
- Context-aware logging
- Automatic rotation
- Performance optimized (~2μs per operation)

**Log Destinations**:
- Console (stdout/stderr)
- File with rotation
- Remote logging (optional)

## Data Flow

### Chat Request Flow

```
1. User Input (TUI)
   │
   ├─→ 2. Command Parser (CLI)
   │
   ├─→ 3. Session Manager (load context)
   │
   ├─→ 4. Provider Layer
   │     │
   │     ├─→ 5a. Format messages for provider
   │     ├─→ 5b. Add system prompts
   │     ├─→ 5c. Inject tool definitions
   │     │
   │     └─→ 6. HTTP Client → LLM Provider API
   │
   ├─→ 7. Stream Response Handler
   │     │
   │     ├─→ 8a. Update TUI (real-time)
   │     ├─→ 8b. Parse tool calls
   │     ├─→ 8c. Buffer response
   │     │
   │     └─→ 9. Execute Tools (if requested)
   │           │
   │           └─→ 10. Tool Executor
   │
   └─→ 11. Session Manager (save messages)
        │
        └─→ 12. Database (persist)
```

### Authentication Flow (OAuth)

```
1. User: ainative-code auth login
   │
   ├─→ 2. Generate PKCE code verifier/challenge
   │
   ├─→ 3. Open browser to authorization URL
   │
   ├─→ 4. User authorizes in browser
   │
   ├─→ 5. Callback to local server (localhost)
   │
   ├─→ 6. Exchange auth code for tokens
   │
   ├─→ 7. Store tokens in OS keychain
   │
   └─→ 8. Cache tokens for current session
```

### Tool Execution Flow

```
1. LLM requests tool execution
   │
   ├─→ 2. Parse tool call (name, arguments)
   │
   ├─→ 3. Permission Check
   │     │
   │     ├─→ [Denied] → Return error to LLM
   │     │
   │     └─→ [Approved] → Continue
   │
   ├─→ 4. Tool Executor
   │     │
   │     ├─→ 5. Validate arguments
   │     ├─→ 6. Set timeout
   │     ├─→ 7. Execute (sandboxed)
   │     │
   │     └─→ 8. Capture output/error
   │
   └─→ 9. Return result to LLM
```

## Key Design Decisions

### 1. Why Go?

**Rationale**:
- Fast compilation and execution
- Excellent concurrency primitives (goroutines, channels)
- Static typing with type inference
- Strong standard library
- Easy cross-platform builds
- Good performance for CLI tools

**Trade-offs**:
- Larger binary size than scripting languages
- Less dynamic than Python
- Fewer ML/AI libraries than Python

### 2. Why Bubble Tea for TUI?

**Rationale**:
- Pure Go implementation
- Elm architecture (model-update-view)
- Excellent performance
- Rich ecosystem (lipgloss, bubbles)
- Active maintenance

**Alternatives Considered**:
- termui (less maintained)
- tview (different paradigm)
- Raw terminal manipulation (too complex)

### 3. Why SQLite for Storage?

**Rationale**:
- No external dependencies
- Single file database
- ACID compliance
- Good performance for read-heavy workloads
- Cross-platform

**Alternatives Considered**:
- JSON files (no transactions, no queries)
- PostgreSQL (requires server)
- Embedded key-value stores (less query flexibility)

### 4. Why Provider Abstraction?

**Rationale**:
- Future-proof for new LLM providers
- Consistent interface across providers
- Easy to test with mocks
- Allows fallback strategies

**Implementation**:
- Common message format
- Streaming abstraction
- Error normalization
- Feature flags for provider-specific capabilities

### 5. Why Multi-Tier Authentication?

**Rationale**:
- Optimize for common case (local validation)
- Fallback for edge cases (API validation)
- Support offline operation (local auth)
- Balance security and performance

**Trade-offs**:
- More complex implementation
- Requires key caching logic
- Potential consistency issues

## Security Architecture

### Threat Model

**Assets to Protect**:
- User API keys (LLM providers)
- OAuth tokens (AINative platform)
- Session data (conversation history)
- User code and files (executed tools)

**Threats**:
- API key exposure (logs, config files)
- Token theft (network sniffing, local access)
- Arbitrary code execution (malicious tool calls)
- Data exfiltration (via LLM)

### Security Controls

#### 1. Credential Protection
- Store in OS keychain (macOS Keychain, Linux Secret Service)
- Never log full credentials
- Mask in UI and logs
- Encrypt at rest

#### 2. Network Security
- TLS 1.2+ for all external communication
- Certificate validation
- No insecure HTTP connections
- Proxy support with auth

#### 3. Tool Execution Safety
- Sandboxed execution environment
- Timeout enforcement (default 30s)
- Permission prompts for dangerous operations
- Working directory restrictions
- Command allowlist/blocklist

#### 4. Input Validation
- Validate all user inputs
- Sanitize file paths
- Validate API responses
- Prevent injection attacks

#### 5. Rate Limiting
- Respect provider rate limits
- Exponential backoff on errors
- Circuit breaker pattern
- Request throttling

#### 6. Audit Logging
- Log all authentication events
- Log all tool executions
- Log API calls (without sensitive data)
- Tamper-evident logs (optional)

## Performance Characteristics

### Benchmarks (Apple M3)

| Operation | Time | Memory | Notes |
|-----------|------|--------|-------|
| Startup | <100ms | 50MB | Cold start |
| First token | <2s | 80MB | Includes network |
| Streaming | 1000+/s | 100MB | Token throughput |
| Session load | <50ms | +10MB | From SQLite |
| Config parse | <10ms | +5MB | YAML parsing |
| Log operation | 2μs | 0 allocs | Disabled level |

### Scalability

**Session Limits**:
- Max messages per session: 10,000
- Max sessions: Limited by disk space
- Concurrent sessions: 1 (current), N (future)

**Provider Limits**:
- Simultaneous providers: All (for fallback)
- Requests/second: Provider-dependent
- Max context window: Provider-dependent (up to 200K tokens)

**Database**:
- SQLite performs well up to 1M records
- Auto-vacuum prevents fragmentation
- Index optimization for common queries

## Deployment Architecture

### Single Binary Distribution

```
ainative-code (binary)
│
├─→ Embedded assets (if any)
├─→ Default config template
└─→ Version info
```

**Runtime Dependencies**:
- None (statically linked)

**User Data Locations**:
- Config: `~/.config/ainative-code/config.yaml`
- Data: `~/.local/share/ainative-code/sessions.db`
- Logs: `~/.local/share/ainative-code/logs/`
- Cache: `~/.cache/ainative-code/`

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
# Build binary

FROM alpine:latest
# Copy binary and run
```

**Volumes**:
- `/root/.config/ainative-code` - Configuration
- `/root/.local/share/ainative-code` - Data
- `/workspace` - Working directory

## Future Architecture Considerations

### Planned Enhancements

1. **Distributed Sessions**
   - Sync sessions across devices via ZeroDB
   - Conflict resolution strategies
   - End-to-end encryption

2. **Plugin System**
   - WASM-based plugins
   - Plugin marketplace
   - Sandboxed execution

3. **Web UI Companion**
   - Next.js frontend
   - WebSocket for real-time updates
   - Shared sessions with CLI

4. **Team Features**
   - Shared sessions
   - Access control
   - Usage analytics
   - Centralized billing

5. **Enhanced Caching**
   - Redis for distributed cache
   - Prompt caching optimization
   - Response deduplication

## References

- [Configuration Guide](../configuration.md)
- [Database Guide](../database-guide.md)
- [Logging Guide](../logging.md)
- [Development Guide](../development/README.md)
- [API Reference](../api/README.md)

---

**Document Version**: 1.0
**Last Updated**: January 2025
**Author**: AINative Studio Engineering Team
