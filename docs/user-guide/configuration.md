# Configuration Guide

This guide covers all configuration options for AINative Code, helping you customize the tool to fit your workflow.

## Table of Contents

1. [Configuration Overview](#configuration-overview)
2. [Configuration File](#configuration-file)
3. [Environment Variables](#environment-variables)
4. [Command-Line Flags](#command-line-flags)
5. [Configuration Examples](#configuration-examples)
6. [Advanced Settings](#advanced-settings)
7. [Security Best Practices](#security-best-practices)

## Configuration Overview

AINative Code uses a hierarchical configuration system with multiple sources:

### Configuration Precedence

Settings are loaded in this order (highest to lowest priority):

1. **Command-line flags** - `--provider anthropic`
2. **Environment variables** - `AINATIVE_LLM_DEFAULT_PROVIDER=anthropic`
3. **Configuration file** - `~/.config/ainative-code/config.yaml`
4. **Default values** - Built-in defaults

### Configuration File Locations

The application searches for `config.yaml` in these locations (in order):

1. Path specified by `--config` flag
2. Current directory: `./config.yaml`
3. Config subdirectory: `./configs/config.yaml`
4. User config directory: `~/.config/ainative-code/config.yaml` (Linux/macOS)
5. User config directory: `%APPDATA%\ainative-code\config.yaml` (Windows)
6. System config: `/etc/ainative/config.yaml` (Linux/macOS)

## Configuration File

### Creating the Configuration File

Initialize a default configuration file:

```bash
ainative-code setup
```

This creates `~/.config/ainative-code/config.yaml` with sensible defaults.

### Complete Configuration Example

Here's a comprehensive example configuration file:

```yaml
# Application Settings
app:
  name: ainative-code
  version: 0.1.0
  environment: development  # development, staging, production
  debug: false

# LLM Provider Configuration
llm:
  default_provider: anthropic

  # Anthropic Claude Configuration
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"  # Use environment variable
    model: claude-3-5-sonnet-20241022
    max_tokens: 4096
    temperature: 0.7
    top_p: 0.9
    top_k: 40
    timeout: 300s
    retry_attempts: 3
    api_version: "2023-06-01"
    extended_thinking:
      enabled: true
      auto_expand: false
      max_depth: 5

  # OpenAI Configuration
  openai:
    api_key: "${OPENAI_API_KEY}"
    model: gpt-4-turbo-preview
    organization: ""  # Optional
    max_tokens: 4096
    temperature: 0.7
    top_p: 1.0
    frequency_penalty: 0.0
    presence_penalty: 0.0
    timeout: 300s
    retry_attempts: 3

  # Google Gemini Configuration
  google:
    api_key: "${GOOGLE_API_KEY}"
    model: gemini-pro
    project_id: ""  # Optional for Vertex AI
    location: us-central1
    max_tokens: 2048
    temperature: 0.7
    top_p: 0.95
    top_k: 40
    timeout: 300s
    retry_attempts: 3

  # AWS Bedrock Configuration
  bedrock:
    region: us-east-1
    model: anthropic.claude-3-sonnet-20240229-v1:0
    profile: default  # AWS profile name
    max_tokens: 4096
    temperature: 0.7
    top_p: 0.9
    timeout: 300s
    retry_attempts: 3

  # Azure OpenAI Configuration
  azure:
    api_key: "${AZURE_OPENAI_API_KEY}"
    endpoint: https://your-resource.openai.azure.com/
    deployment_name: gpt-4
    api_version: "2024-02-15-preview"
    max_tokens: 4096
    temperature: 0.7
    top_p: 1.0
    timeout: 300s
    retry_attempts: 3

  # Ollama (Local LLM) Configuration
  ollama:
    base_url: http://localhost:11434
    model: codellama
    max_tokens: 2048
    temperature: 0.7
    top_p: 0.9
    top_k: 40
    timeout: 600s
    retry_attempts: 3
    keep_alive: 5m

  # Fallback Configuration
  fallback:
    enabled: true
    providers:
      - anthropic
      - openai
      - google
    max_retries: 2
    retry_delay: 5s

# AINative Platform Configuration
platform:
  authentication:
    method: oauth2  # jwt, api_key, oauth2
    client_id: ainative-code-cli
    token_url: https://auth.ainative.studio/token
    scopes:
      - read
      - write
      - offline_access
    timeout: 30s

  organization:
    id: ""
    name: ""
    workspace: ""

# Service Endpoints
services:
  # ZeroDB Configuration
  zerodb:
    enabled: true
    endpoint: postgresql://localhost:5432
    database: zerodb
    ssl: true
    ssl_mode: require
    max_connections: 25
    idle_connections: 5
    conn_max_lifetime: 1h
    timeout: 30s
    retry_attempts: 3
    retry_delay: 1s

  # Design Tokens Service
  design:
    enabled: true
    endpoint: https://design.ainative.studio
    timeout: 30s
    retry_attempts: 3

  # Strapi CMS
  strapi:
    enabled: true
    endpoint: https://cms.ainative.studio
    timeout: 30s
    retry_attempts: 3

  # RLHF Service
  rlhf:
    enabled: true
    endpoint: https://rlhf.ainative.studio
    timeout: 30s
    retry_attempts: 3
    model_id: ""

# Tool Configuration
tools:
  filesystem:
    enabled: true
    allowed_paths:
      - /home/user/projects
      - /workspace
    blocked_paths:
      - /etc
      - /sys
      - /proc
    max_file_size: 10485760  # 10 MB in bytes
    allowed_extensions:
      - .go
      - .js
      - .ts
      - .py
      - .java
      - .rs
      - .md
      - .json
      - .yaml

  terminal:
    enabled: true
    allowed_commands:
      - git
      - npm
      - go
      - python
      - docker
    blocked_commands:
      - rm -rf /
      - dd
      - mkfs
    timeout: 300s
    working_dir: /workspace

  browser:
    enabled: false
    headless: true
    timeout: 30s

  code_analysis:
    enabled: true
    languages:
      - go
      - javascript
      - typescript
      - python
    max_file_size: 5242880  # 5 MB
    include_tests: true

# Performance Settings
performance:
  cache:
    enabled: true
    type: memory  # memory, redis, memcached
    ttl: 1h
    max_size: 100  # MB

  rate_limit:
    enabled: true
    requests_per_minute: 60
    burst_size: 10
    time_window: 1m
    per_user: true

  concurrency:
    max_workers: 10
    max_queue_size: 100
    worker_timeout: 5m

  circuit_breaker:
    enabled: true
    failure_threshold: 5
    success_threshold: 2
    timeout: 60s
    reset_timeout: 30s

# Logging Configuration
logging:
  level: info  # debug, info, warn, error
  format: json  # json, console
  output: stdout  # stdout, file
  file_path: /var/log/ainative-code/app.log
  max_size: 100  # MB
  max_backups: 3
  max_age: 7  # days
  compress: true
  sensitive_keys:
    - password
    - api_key
    - token
    - secret

# Security Settings
security:
  encrypt_config: false
  tls_enabled: true
  allowed_origins:
    - https://ainative.studio
    - http://localhost:*
  secret_rotation: 90d
```

## Environment Variables

All configuration values can be set via environment variables using the `AINATIVE_` prefix.

### Naming Convention

Convert configuration keys to environment variables:

- Add `AINATIVE_` prefix
- Convert nested keys with underscores
- Use uppercase letters

**Examples:**

| Configuration Key | Environment Variable |
|------------------|---------------------|
| `llm.default_provider` | `AINATIVE_LLM_DEFAULT_PROVIDER` |
| `llm.anthropic.api_key` | `AINATIVE_LLM_ANTHROPIC_API_KEY` |
| `services.zerodb.endpoint` | `AINATIVE_SERVICES_ZERODB_ENDPOINT` |
| `logging.level` | `AINATIVE_LOGGING_LEVEL` |

### Common Environment Variables

```bash
# LLM Provider Settings
export AINATIVE_LLM_DEFAULT_PROVIDER=anthropic
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."
export GOOGLE_API_KEY="..."

# Application Settings
export AINATIVE_APP_ENVIRONMENT=production
export AINATIVE_APP_DEBUG=false

# ZeroDB Settings
export AINATIVE_SERVICES_ZERODB_ENDPOINT="postgresql://host:5432/db"
export AINATIVE_SERVICES_ZERODB_SSL=true

# Logging
export AINATIVE_LOGGING_LEVEL=info
export AINATIVE_LOGGING_FORMAT=json
```

### Loading Environment Variables from File

Create a `.env` file:

```bash
# .env file
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
AINATIVE_LLM_DEFAULT_PROVIDER=anthropic
AINATIVE_APP_ENVIRONMENT=development
```

Load it before running:

```bash
# Using source (bash/zsh)
source .env
ainative-code chat

# Using export
export $(cat .env | xargs)
ainative-code chat

# Using direnv (recommended)
direnv allow
ainative-code chat
```

## Command-Line Flags

Override configuration with command-line flags:

### Global Flags

```bash
# Specify config file
ainative-code --config /path/to/config.yaml chat

# Set provider
ainative-code --provider anthropic chat

# Set model
ainative-code --model claude-3-opus-20240229 chat

# Enable verbose logging
ainative-code --verbose chat
ainative-code -v chat
```

### Command-Specific Flags

```bash
# Chat flags
ainative-code chat --session-id abc123
ainative-code chat --system "You are a Go expert"
ainative-code chat --stream=false

# Session flags
ainative-code session list --all
ainative-code session list --limit 20

# Auth flags
ainative-code auth login --auth-url https://custom-auth.example.com
```

## Configuration Examples

### Example 1: Anthropic Claude Only

```yaml
app:
  environment: production
  debug: false

llm:
  default_provider: anthropic
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    model: claude-3-5-sonnet-20241022
    max_tokens: 8192
    temperature: 0.7

logging:
  level: info
  format: console
```

### Example 2: Multi-Provider with Fallback

```yaml
llm:
  default_provider: anthropic

  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    model: claude-3-5-sonnet-20241022
    max_tokens: 4096

  openai:
    api_key: "${OPENAI_API_KEY}"
    model: gpt-4-turbo-preview
    max_tokens: 4096

  fallback:
    enabled: true
    providers:
      - anthropic
      - openai
    max_retries: 2
```

### Example 3: Local Development with Ollama

```yaml
llm:
  default_provider: ollama

  ollama:
    base_url: http://localhost:11434
    model: codellama
    max_tokens: 4096
    temperature: 0.7
    keep_alive: 10m

tools:
  filesystem:
    enabled: true
    allowed_paths:
      - ${HOME}/projects

  terminal:
    enabled: true
    working_dir: ${HOME}/projects

logging:
  level: debug
  format: console
```

### Example 4: Enterprise Production Setup

```yaml
app:
  environment: production
  debug: false

llm:
  default_provider: azure

  azure:
    api_key: "${AZURE_OPENAI_API_KEY}"
    endpoint: https://company.openai.azure.com/
    deployment_name: gpt-4-production
    api_version: "2024-02-15-preview"
    max_tokens: 4096
    timeout: 300s
    retry_attempts: 5

services:
  zerodb:
    enabled: true
    endpoint: postgresql://prod-db.company.com:5432
    database: ainative_prod
    ssl: true
    ssl_mode: require
    max_connections: 50

performance:
  cache:
    enabled: true
    type: redis
    redis_url: redis://cache.company.com:6379

  rate_limit:
    enabled: true
    requests_per_minute: 120
    per_user: true

logging:
  level: warn
  format: json
  output: file
  file_path: /var/log/ainative-code/app.log
  max_size: 100
  max_backups: 7
  compress: true

security:
  encrypt_config: true
  tls_enabled: true
  allowed_origins:
    - https://company.com
  secret_rotation: 30d
```

### Example 5: Development with All Features

```yaml
app:
  environment: development
  debug: true

llm:
  default_provider: anthropic

  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    model: claude-3-5-sonnet-20241022
    extended_thinking:
      enabled: true
      auto_expand: true

services:
  zerodb:
    enabled: true
    endpoint: postgresql://localhost:5432

  design:
    enabled: true
    endpoint: http://localhost:3000

  strapi:
    enabled: true
    endpoint: http://localhost:1337

  rlhf:
    enabled: true
    endpoint: http://localhost:8080

tools:
  filesystem:
    enabled: true
    allowed_paths:
      - ${PWD}

  terminal:
    enabled: true

  browser:
    enabled: true
    headless: false

logging:
  level: debug
  format: console
  output: stdout
```

## Advanced Settings

### Extended Thinking (Anthropic)

Enable Claude's extended thinking for complex reasoning:

```yaml
llm:
  anthropic:
    extended_thinking:
      enabled: true
      auto_expand: false  # Expand thinking blocks automatically
      max_depth: 5  # Maximum nested thinking depth
```

### Fallback Configuration

Configure automatic fallback between providers:

```yaml
llm:
  fallback:
    enabled: true
    providers:
      - anthropic  # Try first
      - openai     # Then this
      - google     # Finally this
    max_retries: 2
    retry_delay: 5s
```

### Circuit Breaker

Prevent cascading failures:

```yaml
performance:
  circuit_breaker:
    enabled: true
    failure_threshold: 5    # Open after 5 failures
    success_threshold: 2    # Close after 2 successes
    timeout: 60s           # Request timeout
    reset_timeout: 30s     # Time before retry
```

### Rate Limiting

Control request rates:

```yaml
performance:
  rate_limit:
    enabled: true
    requests_per_minute: 60
    burst_size: 10
    time_window: 1m
    per_user: true
    per_endpoint: false
    endpoint_limits:
      /chat: 30
      /session: 120
```

### Caching

Configure response caching:

```yaml
performance:
  cache:
    enabled: true
    type: redis
    ttl: 1h
    max_size: 100  # MB for memory cache
    redis_url: redis://localhost:6379
```

### Tool Restrictions

Restrict tool capabilities for security:

```yaml
tools:
  filesystem:
    enabled: true
    allowed_paths:
      - /home/user/projects
    blocked_paths:
      - /etc
      - /sys
    max_file_size: 10485760  # 10 MB

  terminal:
    enabled: true
    allowed_commands:
      - git
      - npm
    blocked_commands:
      - rm -rf
      - sudo
    timeout: 300s
```

## Security Best Practices

### 1. Protect API Keys

Never commit API keys to version control:

```yaml
# Good: Use environment variables
llm:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"

# Bad: Hardcoded keys
llm:
  anthropic:
    api_key: "sk-ant-..." # Don't do this!
```

### 2. Use .gitignore

Exclude sensitive files:

```bash
# .gitignore
config.yaml
.env
*.key
tokens.json
```

### 3. Encrypt Configuration

Enable config encryption for sensitive environments:

```yaml
security:
  encrypt_config: true
  encryption_key: "${ENCRYPTION_KEY}"
```

### 4. Restrict File Access

Limit filesystem access:

```yaml
tools:
  filesystem:
    allowed_paths:
      - /workspace
    blocked_paths:
      - /etc
      - /root
```

### 5. Use Least Privilege

Only enable needed services:

```yaml
services:
  zerodb:
    enabled: false  # Disable if not needed
  strapi:
    enabled: false
```

### 6. Rotate Secrets

Configure automatic rotation:

```yaml
security:
  secret_rotation: 30d  # Rotate every 30 days
```

### 7. Use TLS

Always use encrypted connections:

```yaml
services:
  zerodb:
    ssl: true
    ssl_mode: require

security:
  tls_enabled: true
```

## Configuration Management

### View Current Configuration

```bash
# Show all configuration
ainative-code config show

# Get specific value
ainative-code config get llm.default_provider

# List all keys
ainative-code config list
```

### Update Configuration

```bash
# Set a value
ainative-code config set llm.default_provider anthropic

# Set nested value
ainative-code config set llm.anthropic.model claude-3-opus-20240229

# Unset a value
ainative-code config unset llm.anthropic.temperature
```

### Validate Configuration

```bash
# Validate config file
ainative-code config validate

# Validate with custom file
ainative-code config validate --config /path/to/config.yaml
```

### Export/Import Configuration

```bash
# Export configuration
ainative-code config export -o my-config.yaml

# Import configuration
ainative-code config import -i my-config.yaml
```

## Troubleshooting Configuration

### Configuration Not Loading

```bash
# Check which config file is being used
ainative-code --verbose config show

# Re-initialize configuration
ainative-code setup --force
```

### Environment Variables Not Working

```bash
# Verify environment variables are set
env | grep AINATIVE

# Test with explicit variable
AINATIVE_LOGGING_LEVEL=debug ainative-code chat
```

### Permission Issues

```bash
# Fix config file permissions
chmod 600 ~/.config/ainative-code/config.yaml

# Fix directory permissions
chmod 700 ~/.config/ainative-code
```

## Next Steps

- [Providers Guide](providers.md) - Configure LLM providers
- [Authentication Guide](authentication.md) - Set up platform authentication
- [Tools Guide](tools.md) - Configure tools and MCP servers
- [Troubleshooting Guide](troubleshooting.md) - Common configuration issues
