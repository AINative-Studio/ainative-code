# Examples and Usage Patterns

This directory contains comprehensive examples and usage patterns for AINative Code.

## Overview

Learn AINative Code through practical examples covering common use cases, advanced features, and best practices.

## Quick Examples

### Basic Usage

**Simple Chat**:
```bash
ainative-code chat "Explain Go channels with examples"
```

**Resume Session**:
```bash
ainative-code chat --resume
```

**Use Specific Provider**:
```bash
ainative-code chat --provider openai "Generate a REST API in Go"
```

### Configuration

**Initialize**:
```bash
ainative-code init
```

**Set Provider**:
```bash
ainative-code config set llm.default_provider anthropic
export ANTHROPIC_API_KEY="your-key"
```

**View Config**:
```bash
ainative-code config show
```

## Comprehensive Guides

### [Basic Usage Examples](basic-usage.md)

Complete guide to common tasks and workflows.

**Topics Covered**:
- Getting started
- Code generation
- Code review and improvement
- Debugging
- Learning and documentation
- Testing
- Refactoring
- Session management
- Provider management
- Tips for effective usage

**Read this if**: You're new to AINative Code or want to learn effective usage patterns.

### [ZeroDB Usage Examples](zerodb-nosql-usage.md)

Working with ZeroDB for vector search and NoSQL operations.

**Topics Covered**:
- Vector operations (upsert, search, delete)
- NoSQL table management
- Agent memory storage
- Quantum-enhanced search
- Integration patterns

**Read this if**: You're using AINative platform's ZeroDB features.

## Configuration Examples

### Minimal Configuration

```yaml
# ~/.config/ainative-code/config.yaml
app:
  environment: development

llm:
  default_provider: anthropic
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    model: claude-3-5-sonnet-20241022
    max_tokens: 4096

logging:
  level: info
  format: console
```

### Multi-Provider Setup

```yaml
llm:
  default_provider: anthropic

  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    model: claude-3-5-sonnet-20241022
    max_tokens: 8192
    temperature: 0.7

  openai:
    api_key: "${OPENAI_API_KEY}"
    model: gpt-4-turbo-preview
    max_tokens: 4096

  ollama:
    endpoint: http://localhost:11434
    model: llama3
    max_tokens: 2048

  fallback:
    enabled: true
    providers:
      - anthropic
      - openai
      - ollama
```

### Full Platform Integration

```yaml
llm:
  default_provider: anthropic
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    model: claude-3-5-sonnet-20241022

platform:
  authentication:
    method: oauth2
    auto_login: true

services:
  zerodb:
    enabled: true
    endpoint: postgresql://zerodb.ainative.studio:5432
    default_project: my-project

  design:
    enabled: true
    api_url: https://design.ainative.studio

  strapi:
    enabled: true
    api_url: https://strapi.ainative.studio

  rlhf:
    enabled: true
    auto_submit: false

mcp_servers:
  - name: "github"
    transport: "stdio"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_TOKEN: "${GITHUB_TOKEN}"

tools:
  bash:
    enabled: true
    require_confirmation: true
    timeout: 30s

  file_operations:
    enabled: true
    require_confirmation: false
    max_file_size: 10485760 # 10MB

logging:
  level: info
  format: json
  rotation:
    enabled: true
    max_size: 100
    max_age: 30
    max_backups: 10
```

### API Key Resolution Examples

```yaml
# Using password managers
llm:
  anthropic:
    api_key: "$(pass show anthropic-api-key)"

  openai:
    api_key: "$(security find-generic-password -w -s openai)"

# Using environment variables
llm:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"

# Direct value (not recommended for security)
llm:
  anthropic:
    api_key: "sk-ant-..."
```

## Use Case Examples

### Code Generation

```bash
# Generate REST API
ainative-code chat "Create a REST API in Go with:
- User CRUD endpoints
- JWT authentication
- Input validation
- Error handling
- Unit tests"

# Generate database schema
ainative-code chat "Design a PostgreSQL schema for a blog with:
- Users
- Posts
- Comments
- Tags
- Proper indexes"

# Generate configuration
ainative-code chat "Create a Docker Compose file for:
- PostgreSQL database
- Redis cache
- Go application
- Nginx reverse proxy"
```

### Code Review

```bash
# Security review
ainative-code chat "Review this code for security vulnerabilities: [paste code]"

# Performance review
ainative-code chat "Analyze this function for performance issues: [paste code]"

# Best practices review
ainative-code chat "Review this code against Go best practices: [paste code]"
```

### Debugging

```bash
# Debug error
ainative-code chat "I'm getting this error: [error message]. Here's the code: [code]"

# Analyze stack trace
ainative-code chat "Explain this stack trace and help me fix it: [stack trace]"

# Memory leak investigation
ainative-code chat "I have a memory leak. Here's the pprof output: [pprof data]"
```

### Learning

```bash
# Learn concepts
ainative-code chat "Explain Go interfaces with practical examples"

# Compare approaches
ainative-code chat "What are the differences between channels and mutexes in Go?"

# Best practices
ainative-code chat "What are the current best practices for error handling in Go?"
```

### Testing

```bash
# Generate unit tests
ainative-code chat "Generate comprehensive unit tests for: [paste code]"

# Generate table-driven tests
ainative-code chat "Create table-driven tests for this function: [paste code]"

# Generate mocks
ainative-code chat "Create mock implementation for this interface: [paste interface]"
```

## Advanced Patterns

### Multi-Session Workflow

```bash
# Session 1: Research
ainative-code chat --new --title "Research: OAuth 2.0"
# Research OAuth 2.0 implementations

# Session 2: Implementation
ainative-code chat --new --title "Implement: OAuth"
# Generate OAuth implementation

# Session 3: Testing
ainative-code chat --new --title "Test: OAuth"
# Generate tests

# Session 4: Documentation
ainative-code chat --new --title "Docs: OAuth"
# Generate documentation
```

### Iterative Refinement

```bash
# Start with basic implementation
ainative-code chat "Create a basic HTTP server in Go"

# Add features iteratively
ainative-code chat --resume "Add middleware for logging"
ainative-code chat --resume "Add graceful shutdown"
ainative-code chat --resume "Add metrics collection"
ainative-code chat --resume "Add comprehensive tests"
```

### Context Management

```bash
# Provide multiple files as context
ainative-code chat "Review these files for consistency:

File: main.go
[paste content]

File: handler.go
[paste content]

File: middleware.go
[paste content]

Are the error handling patterns consistent across all files?"
```

## Platform Integration Examples

### ZeroDB Vector Search

```bash
ainative-code chat "Store this documentation in ZeroDB for semantic search:
[paste documentation]

Then create a function that searches for relevant docs based on user queries"
```

### Design Token Extraction

```bash
ainative-code chat "Extract design tokens from our CSS files and generate:
1. Tailwind config
2. TypeScript theme object
3. Documentation"
```

### Strapi CMS Content

```bash
ainative-code chat "Create a blog post in Strapi about today's product updates"
```

## Project Examples

Full example projects are available in the repository:

- [Example Configuration](../../examples/config.yaml) - Complete configuration file
- [Resolver Examples](../../examples/config-with-resolver.yaml) - Dynamic API key resolution

## Best Practices

### 1. Clear and Specific Prompts

**Good**:
```bash
ainative-code chat "Create a user authentication system with:
- Email/password login
- JWT tokens
- Password hashing with bcrypt
- Rate limiting
- Unit tests"
```

**Avoid**:
```bash
ainative-code chat "Make me an auth system"
```

### 2. Provide Context

**Good**:
```bash
ainative-code chat "I'm building a REST API in Go using Chi router.
I need to add authentication middleware that:
- Validates JWT tokens
- Handles expired tokens
- Returns 401 for invalid tokens
Here's my current middleware structure: [paste code]"
```

**Avoid**:
```bash
ainative-code chat "Add auth middleware"
```

### 3. Use Code Blocks

Always use proper markdown code blocks:

````
```go
func main() {
    // Your code here
}
```
````

### 4. Iterate and Refine

Don't try to get everything perfect in one prompt:
1. Start with basic implementation
2. Add error handling
3. Add tests
4. Add documentation
5. Optimize performance

### 5. Ask for Explanations

Request understanding, not just code:
```bash
ainative-code chat "Implement a rate limiter and explain:
- Why this approach is used
- What the trade-offs are
- When to use alternatives"
```

## Related Documentation

- [User Guide](../user-guide/README.md) - Complete user guide
- [Configuration Guide](../user-guide/configuration.md) - Configuration options
- [Provider Guide](../user-guide/providers.md) - LLM provider details
- [Tools Guide](../user-guide/tools.md) - Built-in tools usage
- [API Reference](../api/README.md) - API documentation

## Additional Resources

### Internal Documentation
- [Logging Guide](../logging.md) - Logging examples
- [Database Guide](../database-guide.md) - Database usage
- [Authentication Guide](../authentication/README.md) - Auth examples

### Configuration Files
- [Config Reference](../configuration.md) - Complete configuration reference
- [Environment Variables](../environment-variables.md) - Environment variable guide

### Video Tutorials

(Coming soon: Video tutorials demonstrating common workflows)

## Getting Help

If you have questions about examples:

- Check the [FAQ](../user-guide/faq.md)
- Browse [GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions)
- Open an [Issue](https://github.com/AINative-studio/ainative-code/issues)
- Email: support@ainative.studio

---

**Last Updated**: January 2025
**Maintainer**: AINative Studio Documentation Team
