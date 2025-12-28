# Configuration Package

The `config` package provides comprehensive configuration management for AINative Code with support for multiple LLM providers, service endpoints, and performance settings.

## Features

- **Multi-Source Configuration**: Load from files, environment variables, and defaults
- **Dynamic API Key Resolution**: Securely resolve API keys from multiple sources (see [RESOLVER.md](./RESOLVER.md))
  - Environment variables: `${OPENAI_API_KEY}`
  - File paths: `~/secrets/api-key.txt`
  - Command execution: `$(pass show anthropic)`
  - Direct strings: `sk-ant-api-key-123`
- **Comprehensive Validation**: Validate all configuration values with clear error messages
- **Multiple LLM Providers**: Support for Anthropic, OpenAI, Google, AWS Bedrock, Azure, and Ollama
- **Service Integration**: Configuration for ZeroDB, Design, Strapi, and RLHF services
- **Security**: Support for encryption, TLS, and secure credential management
- **Performance Tuning**: Built-in caching, rate limiting, and circuit breaker configuration

## Quick Start

```go
package main

import (
    "log"
    "github.com/AINative-studio/ainative-code/internal/config"
)

func main() {
    // Load configuration
    loader := config.NewLoader()
    cfg, err := loader.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Use configuration
    log.Printf("App: %s (env: %s)", cfg.App.Name, cfg.App.Environment)
    log.Printf("LLM Provider: %s", cfg.LLM.DefaultProvider)
}
```

## Configuration Structure

The configuration is organized into the following sections:

- `App`: General application settings
- `LLM`: Language model provider configurations
- `Platform`: AINative platform authentication and organization
- `Services`: External service endpoints (ZeroDB, Design, Strapi, RLHF)
- `Tools`: Tool-specific configurations (filesystem, terminal, browser, code analysis)
- `Performance`: Performance optimization settings (cache, rate limit, concurrency)
- `Logging`: Logging configuration
- `Security`: Security settings (encryption, TLS, CORS)

## Loading Configuration

### From Default Locations

```go
loader := config.NewLoader()
cfg, err := loader.Load()
```

Default search paths:
1. `./config.yaml`
2. `./configs/config.yaml`
3. `$HOME/.ainative/config.yaml`
4. `/etc/ainative/config.yaml`

### From Specific File

```go
loader := config.NewLoader()
cfg, err := loader.LoadFromFile("/path/to/config.yaml")
```

### With Custom Options

```go
loader := config.NewLoader(
    config.WithConfigName("myconfig"),
    config.WithConfigType("yaml"),
    config.WithConfigPaths("/custom/path"),
    config.WithEnvPrefix("MYAPP"),
)
```

## Environment Variables

All configuration values can be set via environment variables with the `AINATIVE_` prefix:

```bash
export AINATIVE_APP_ENVIRONMENT=production
export AINATIVE_LLM_DEFAULT_PROVIDER=anthropic
export AINATIVE_LLM_ANTHROPIC_API_KEY=sk-ant-...
export AINATIVE_SERVICES_ZERODB_ENDPOINT=postgresql://localhost:5432
```

## Validation

All configurations are automatically validated on load. Common validation checks:

- Required fields must be present
- Values must be within valid ranges
- URLs and paths must be properly formatted
- Dependencies between fields must be consistent

Example validation error:

```
Configuration validation failed:
  - llm.anthropic.api_key: Anthropic API key is required
  - services.zerodb.endpoint: endpoint is required
```

## LLM Provider Configuration

### Anthropic Claude

```yaml
llm:
  default_provider: anthropic
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    model: claude-3-5-sonnet-20241022
    max_tokens: 4096
    temperature: 0.7
```

### OpenAI

```yaml
llm:
  default_provider: openai
  openai:
    api_key: ${OPENAI_API_KEY}
    model: gpt-4-turbo-preview
    max_tokens: 4096
```

### Fallback Support

```yaml
llm:
  fallback:
    enabled: true
    providers:
      - anthropic
      - openai
      - ollama
```

## Dynamic API Key Resolution

The configuration system supports dynamic resolution of API keys and secrets from multiple sources. This allows you to avoid hardcoding sensitive values in configuration files.

### Supported Formats

```yaml
llm:
  anthropic:
    # Environment variable
    api_key: "${ANTHROPIC_API_KEY}"

  openai:
    # Command execution (password manager)
    api_key: "$(pass show openai/api-key)"

  google:
    # File path
    api_key: "~/secrets/google-api-key.txt"

  azure:
    # Direct string (not recommended for production)
    api_key: "sk-ant-api-key-123456"
```

### Security Options

Configure the resolver for enhanced security:

```go
import (
    "time"
    "github.com/AINative-studio/ainative-code/internal/config"
)

// Restrict allowed commands (recommended for production)
resolver := config.NewResolver(
    config.WithAllowedCommands("pass", "1password", "aws"),
    config.WithCommandTimeout(10 * time.Second),
)

loader := config.NewLoader(
    config.WithResolver(resolver),
)

cfg, err := loader.Load()
```

For comprehensive documentation on API key resolution, see [RESOLVER.md](./RESOLVER.md).

## Service Configuration

### ZeroDB

```yaml
services:
  zerodb:
    enabled: true
    endpoint: ${ZERODB_ENDPOINT}
    database: ainative_code
    username: ${ZERODB_USERNAME}
    password: ${ZERODB_PASSWORD}
    ssl: true
    max_connections: 10
```

## Performance Configuration

### Caching

```yaml
performance:
  cache:
    enabled: true
    type: memory
    ttl: 1h
    max_size: 100
```

### Rate Limiting

```yaml
performance:
  rate_limit:
    enabled: true
    requests_per_minute: 60
    burst_size: 10
```

## Testing

Run the configuration package tests:

```bash
go test ./internal/config/... -v
```

Check test coverage:

```bash
go test ./internal/config/... -cover
```

Current coverage: **63.8%**

## Documentation

For complete configuration documentation, see:
- [Configuration Guide](../../docs/configuration.md)
- [Example Config](../../examples/config.yaml)

## Security Best Practices

1. **Never commit secrets** - Use environment variables for API keys and passwords
2. **Enable encryption** - Set `security.encrypt_config: true` in production
3. **Use TLS** - Enable TLS for network communications
4. **Rotate secrets** - Configure automatic secret rotation
5. **Restrict tool access** - Carefully configure filesystem and terminal tool paths

## File Structure

```
internal/config/
├── README.md           # This file
├── doc.go             # Package documentation
├── types.go           # Configuration type definitions
├── validator.go       # Configuration validation logic
├── validator_test.go  # Validation tests
├── loader.go          # Configuration loader
└── loader_test.go     # Loader tests
```

## Dependencies

- `github.com/spf13/viper` - Configuration management
- `github.com/AINative-studio/ainative-code/internal/errors` - Error handling

## Contributing

When adding new configuration options:

1. Add the field to the appropriate struct in `types.go`
2. Add validation logic in `validator.go`
3. Add default value in `loader.go` `setDefaults()` method
4. Update `examples/config.yaml` with example usage
5. Document in `docs/configuration.md`
6. Add tests in `*_test.go` files

## License

Copyright (c) 2025 AINative Studio
