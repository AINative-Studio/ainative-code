# Configuration Guide

This document provides comprehensive information about configuring the AINative Code application.

## Table of Contents

1. [Overview](#overview)
2. [Configuration Sources](#configuration-sources)
3. [Configuration Schema](#configuration-schema)
4. [LLM Provider Configuration](#llm-provider-configuration)
5. [Platform Configuration](#platform-configuration)
6. [Service Endpoints](#service-endpoints)
7. [Tool Configuration](#tool-configuration)
8. [Performance Settings](#performance-settings)
9. [Security](#security)
10. [Environment Variables](#environment-variables)
11. [Validation](#validation)
12. [Best Practices](#best-practices)

## Overview

AINative Code uses a hierarchical configuration system that supports multiple sources:
- Configuration files (YAML)
- Environment variables
- Command-line flags
- Default values

Configuration is loaded in the following order of precedence (highest to lowest):
1. Command-line flags
2. Environment variables
3. Configuration file
4. Default values

## Configuration Sources

### Configuration File

The application looks for configuration files in the following locations (in order):
1. Current directory: `./config.yaml`
2. Config directory: `./configs/config.yaml`
3. User home: `$HOME/.ainative/config.yaml`
4. System: `/etc/ainative/config.yaml`

You can also specify a custom configuration file:
```bash
ainative-code --config /path/to/config.yaml
```

### Environment Variables

All configuration values can be set via environment variables using the `AINATIVE_` prefix. Nested keys use underscores:

```bash
# Example environment variables
export AINATIVE_APP_ENVIRONMENT=production
export AINATIVE_LLM_DEFAULT_PROVIDER=anthropic
export AINATIVE_LLM_ANTHROPIC_API_KEY=sk-ant-...
export AINATIVE_SERVICES_ZERODB_ENDPOINT=postgresql://localhost:5432
```

### Example Configuration

A complete example configuration file is available at [`examples/config.yaml`](../examples/config.yaml).

## Configuration Schema

### Application Settings

```yaml
app:
  name: ainative-code              # Application name
  version: 0.1.0                   # Application version
  environment: development         # development, staging, production
  debug: true                      # Enable debug mode
```

**Environment Values:**
- `development`: Development mode with verbose logging
- `staging`: Staging environment
- `production`: Production mode with optimized settings

## LLM Provider Configuration

### Overview

AINative Code supports multiple LLM providers with automatic fallback:

- **Anthropic Claude**: Advanced reasoning and code generation
- **OpenAI**: GPT-4 and other OpenAI models
- **Google Gemini**: Google's multimodal AI
- **AWS Bedrock**: Managed AI service on AWS
- **Azure OpenAI**: OpenAI models on Azure
- **Ollama**: Local open-source models

### Default Provider

```yaml
llm:
  default_provider: anthropic  # Primary provider to use
```

### Anthropic Claude

```yaml
llm:
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}           # Required: API key
    model: claude-3-5-sonnet-20241022       # Model to use
    max_tokens: 4096                        # Maximum response tokens
    temperature: 0.7                        # Randomness (0.0-1.0)
    top_p: 1.0                              # Nucleus sampling (0.0-1.0)
    top_k: 0                                # Top-K sampling
    timeout: 30s                            # Request timeout
    retry_attempts: 3                       # Number of retry attempts
    api_version: "2023-06-01"               # API version
    base_url: https://api.anthropic.com     # Optional: custom base URL
```

**Available Models:**
- `claude-3-5-sonnet-20241022`: Latest Sonnet model (recommended)
- `claude-3-opus-20240229`: Most capable model
- `claude-3-sonnet-20240229`: Balanced performance
- `claude-3-haiku-20240307`: Fastest model

### OpenAI

```yaml
llm:
  openai:
    api_key: ${OPENAI_API_KEY}              # Required: API key
    model: gpt-4-turbo-preview              # Model to use
    organization: ${OPENAI_ORG_ID}          # Optional: organization ID
    max_tokens: 4096                        # Maximum response tokens
    temperature: 0.7                        # Randomness (0.0-2.0)
    top_p: 1.0                              # Nucleus sampling (0.0-1.0)
    frequency_penalty: 0.0                  # Frequency penalty (-2.0-2.0)
    presence_penalty: 0.0                   # Presence penalty (-2.0-2.0)
    timeout: 30s                            # Request timeout
    retry_attempts: 3                       # Number of retry attempts
    base_url: https://api.openai.com/v1     # Optional: custom base URL
```

**Available Models:**
- `gpt-4-turbo-preview`: Latest GPT-4 Turbo
- `gpt-4`: Standard GPT-4
- `gpt-3.5-turbo`: Faster, more affordable

### Google Gemini

```yaml
llm:
  google:
    api_key: ${GOOGLE_API_KEY}              # Required: API key
    model: gemini-pro                       # Model to use
    project_id: ${GOOGLE_PROJECT_ID}        # Optional: for Vertex AI
    location: us-central1                   # Optional: for Vertex AI
    max_tokens: 4096                        # Maximum response tokens
    temperature: 0.7                        # Randomness (0.0-1.0)
    top_p: 1.0                              # Nucleus sampling (0.0-1.0)
    top_k: 40                               # Top-K sampling
    timeout: 30s                            # Request timeout
    retry_attempts: 3                       # Number of retry attempts
```

**Available Models:**
- `gemini-pro`: Best for text tasks
- `gemini-pro-vision`: Multimodal capabilities

### AWS Bedrock

```yaml
llm:
  bedrock:
    region: us-east-1                                           # AWS region
    model: anthropic.claude-3-sonnet-20240229-v1:0             # Model ARN
    # Option 1: Use explicit credentials
    access_key_id: ${AWS_ACCESS_KEY_ID}
    secret_access_key: ${AWS_SECRET_ACCESS_KEY}
    session_token: ${AWS_SESSION_TOKEN}                         # Optional
    # Option 2: Use AWS profile
    profile: default
    max_tokens: 4096
    temperature: 0.7
    top_p: 1.0
    timeout: 60s
    retry_attempts: 3
```

**Available Models:**
- `anthropic.claude-3-sonnet-20240229-v1:0`: Claude 3 Sonnet
- `anthropic.claude-3-haiku-20240307-v1:0`: Claude 3 Haiku
- `meta.llama2-70b-chat-v1`: Llama 2 70B

### Azure OpenAI

```yaml
llm:
  azure:
    api_key: ${AZURE_OPENAI_API_KEY}        # Required: API key
    endpoint: ${AZURE_OPENAI_ENDPOINT}      # Required: endpoint URL
    deployment_name: gpt-4-deployment       # Required: deployment name
    api_version: "2023-05-15"               # API version
    max_tokens: 4096
    temperature: 0.7
    top_p: 1.0
    timeout: 30s
    retry_attempts: 3
```

### Ollama (Local Models)

```yaml
llm:
  ollama:
    base_url: http://localhost:11434        # Ollama server URL
    model: llama2                           # Model name
    max_tokens: 4096
    temperature: 0.7
    top_p: 1.0
    top_k: 40
    timeout: 120s                           # Longer timeout for local models
    retry_attempts: 1
    keep_alive: 5m                          # Model keep-alive duration
```

**Popular Models:**
- `llama2`: Meta's Llama 2
- `codellama`: Code-specialized Llama
- `mistral`: Mistral 7B
- `mixtral`: Mixtral 8x7B

### Fallback Configuration

Enable automatic fallback to alternative providers:

```yaml
llm:
  fallback:
    enabled: true
    providers:                              # Ordered list of fallback providers
      - anthropic
      - openai
      - ollama
    max_retries: 2                          # Retries per provider
    retry_delay: 1s                         # Delay between retries
```

## Platform Configuration

### Authentication

AINative Code supports multiple authentication methods:

```yaml
platform:
  authentication:
    method: api_key                         # api_key, jwt, oauth2
    timeout: 10s
```

#### API Key Authentication

```yaml
platform:
  authentication:
    method: api_key
    api_key: ${AINATIVE_API_KEY}
```

#### JWT Authentication

```yaml
platform:
  authentication:
    method: jwt
    token: ${AINATIVE_TOKEN}
    refresh_token: ${AINATIVE_REFRESH_TOKEN}
```

#### OAuth2 Authentication

```yaml
platform:
  authentication:
    method: oauth2
    client_id: ${AINATIVE_CLIENT_ID}
    client_secret: ${AINATIVE_CLIENT_SECRET}
    token_url: https://auth.ainative.studio/oauth/token
    scopes:
      - read
      - write
```

### Organization Settings

```yaml
platform:
  organization:
    id: ${AINATIVE_ORG_ID}
    name: My Organization
    workspace: default
```

## Service Endpoints

### ZeroDB

Encrypted database service:

```yaml
services:
  zerodb:
    enabled: true
    endpoint: ${ZERODB_ENDPOINT}
    database: ainative_code
    username: ${ZERODB_USERNAME}
    password: ${ZERODB_PASSWORD}
    ssl: true
    ssl_mode: require                       # disable, require, verify-ca, verify-full
    max_connections: 10
    idle_connections: 2
    conn_max_lifetime: 1h
    timeout: 5s
    retry_attempts: 3
    retry_delay: 1s
```

**SSL Modes:**
- `disable`: No SSL
- `require`: SSL required but no verification
- `verify-ca`: Verify certificate authority
- `verify-full`: Full certificate verification

### AINative Design Service

```yaml
services:
  design:
    enabled: true
    endpoint: https://design.ainative.studio/api
    api_key: ${DESIGN_API_KEY}
    timeout: 30s
    retry_attempts: 3
```

### Strapi CMS

```yaml
services:
  strapi:
    enabled: false
    endpoint: ${STRAPI_ENDPOINT}
    api_key: ${STRAPI_API_KEY}
    timeout: 30s
    retry_attempts: 3
```

### RLHF Service

Reinforcement Learning from Human Feedback:

```yaml
services:
  rlhf:
    enabled: false
    endpoint: ${RLHF_ENDPOINT}
    api_key: ${RLHF_API_KEY}
    timeout: 60s
    retry_attempts: 3
    model_id: ${RLHF_MODEL_ID}
```

## Tool Configuration

### File System Tool

Controls file system access:

```yaml
tools:
  filesystem:
    enabled: true
    allowed_paths:
      - /Users/aideveloper/projects
      - /tmp/ainative
    blocked_paths:
      - /Users/aideveloper/.ssh
      - /etc/passwd
    max_file_size: 104857600                # 100MB in bytes
    allowed_extensions:
      - .go
      - .py
      - .js
      - .ts
```

**Security Note:** Always use absolute paths and carefully review allowed paths.

### Terminal Tool

Controls command execution:

```yaml
tools:
  terminal:
    enabled: true
    allowed_commands:
      - git
      - npm
      - go
      - python
    blocked_commands:
      - rm -rf /
      - mkfs
    timeout: 5m
    working_dir: /Users/aideveloper/projects
```

**Security Note:** Use allowlists for commands and always block dangerous operations.

### Browser Automation Tool

```yaml
tools:
  browser:
    enabled: false
    headless: true
    timeout: 30s
    user_agent: AINative-Code/0.1.0
```

### Code Analysis Tool

```yaml
tools:
  code_analysis:
    enabled: true
    languages:
      - go
      - python
      - javascript
      - typescript
    max_file_size: 10485760                 # 10MB
    include_tests: true
```

## Performance Settings

### Caching

```yaml
performance:
  cache:
    enabled: true
    type: memory                            # memory, redis, memcached
    ttl: 1h
    max_size: 100                           # MB
    # For Redis:
    redis_url: redis://localhost:6379/0
    # For Memcached:
    memcached_url: localhost:11211
```

**Cache Types:**
- `memory`: In-memory cache (fastest, lost on restart)
- `redis`: Redis-backed cache (persistent, distributed)
- `memcached`: Memcached-backed cache (distributed)

### Rate Limiting

```yaml
performance:
  rate_limit:
    enabled: true
    requests_per_minute: 60
    burst_size: 10
    time_window: 1m
```

### Concurrency

```yaml
performance:
  concurrency:
    max_workers: 10
    max_queue_size: 100
    worker_timeout: 5m
```

### Circuit Breaker

Prevents cascading failures:

```yaml
performance:
  circuit_breaker:
    enabled: true
    failure_threshold: 5                    # Open after 5 failures
    success_threshold: 2                    # Close after 2 successes
    timeout: 60s                            # Request timeout
    reset_timeout: 30s                      # Time before retry
```

## Security

### Encryption

```yaml
security:
  encrypt_config: false
  encryption_key: ${CONFIG_ENCRYPTION_KEY}  # 32+ characters for AES-256
```

### CORS

```yaml
security:
  allowed_origins:
    - http://localhost:3000
    - https://app.ainative.studio
```

### TLS/SSL

```yaml
security:
  tls_enabled: false
  tls_cert_path: /etc/ssl/certs/ainative.crt
  tls_key_path: /etc/ssl/private/ainative.key
```

### Secret Rotation

```yaml
security:
  secret_rotation: 90d                      # Rotate secrets every 90 days
```

## Logging

```yaml
logging:
  level: info                               # debug, info, warn, error
  format: json                              # json, console
  output: stdout                            # stdout, file
  file_path: /var/log/ainative-code/app.log
  max_size: 100                             # MB per file
  max_backups: 3                            # Number of old log files
  max_age: 7                                # Days to retain logs
  compress: true                            # Compress rotated logs
  sensitive_keys:                           # Keys to redact from logs
    - api_key
    - password
    - token
    - secret
```

**Log Levels:**
- `debug`: Detailed debugging information
- `info`: General informational messages
- `warn`: Warning messages
- `error`: Error messages only

## Environment Variables

### Naming Convention

Environment variables use the `AINATIVE_` prefix with underscores for nesting:

```
AINATIVE_<SECTION>_<SUBSECTION>_<KEY>
```

### Common Environment Variables

```bash
# Application
export AINATIVE_APP_ENVIRONMENT=production
export AINATIVE_APP_DEBUG=false

# LLM Providers
export AINATIVE_LLM_DEFAULT_PROVIDER=anthropic
export AINATIVE_LLM_ANTHROPIC_API_KEY=sk-ant-...
export AINATIVE_LLM_OPENAI_API_KEY=sk-...

# Platform
export AINATIVE_PLATFORM_AUTHENTICATION_API_KEY=your-api-key

# Services
export AINATIVE_SERVICES_ZERODB_ENDPOINT=postgresql://localhost:5432
export AINATIVE_SERVICES_ZERODB_USERNAME=user
export AINATIVE_SERVICES_ZERODB_PASSWORD=pass

# Security
export AINATIVE_SECURITY_ENCRYPTION_KEY=your-32-char-encryption-key-here
```

### Loading from .env File

Create a `.env` file in your project root:

```bash
# .env
AINATIVE_LLM_ANTHROPIC_API_KEY=sk-ant-...
AINATIVE_SERVICES_ZERODB_PASSWORD=secret
```

The application automatically loads `.env` files if present.

## Validation

The configuration system performs comprehensive validation on load:

### Validation Rules

1. **Required Fields**: Certain fields must be present
2. **Type Validation**: Values must match expected types
3. **Range Validation**: Numeric values must be within valid ranges
4. **Format Validation**: URLs, emails, paths must be properly formatted
5. **Dependency Validation**: Related fields must be consistent

### Validation Examples

```yaml
# Invalid: temperature out of range
llm:
  anthropic:
    temperature: 1.5  # Error: must be between 0 and 1

# Invalid: missing required field
services:
  zerodb:
    enabled: true
    # Error: endpoint is required when enabled
```

### Validation Errors

Validation errors provide clear, actionable messages:

```
Configuration validation failed:
  - llm.anthropic.api_key: Anthropic API key is required
  - services.zerodb.endpoint: endpoint is required
  - tools.filesystem.allowed_paths: at least one allowed path must be specified
```

## Best Practices

### 1. Use Environment Variables for Secrets

Never commit secrets to configuration files:

```yaml
# Good
llm:
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}

# Bad
llm:
  anthropic:
    api_key: sk-ant-your-actual-key  # Don't do this!
```

### 2. Separate Configurations by Environment

Use different configuration files for each environment:

```
configs/
  ├── development.yaml
  ├── staging.yaml
  └── production.yaml
```

Load with:
```bash
ainative-code --config configs/production.yaml
```

### 3. Enable Security Features in Production

```yaml
app:
  environment: production
  debug: false

security:
  encrypt_config: true
  tls_enabled: true

logging:
  level: info
  format: json
```

### 4. Configure Appropriate Timeouts

Adjust timeouts based on your use case:

```yaml
llm:
  anthropic:
    timeout: 30s      # Standard for most operations

  ollama:
    timeout: 120s     # Longer for local models
```

### 5. Use Fallback Providers

Configure fallback for high availability:

```yaml
llm:
  default_provider: anthropic
  fallback:
    enabled: true
    providers:
      - anthropic
      - openai
      - ollama  # Local fallback
```

### 6. Implement Rate Limiting

Protect against abuse and control costs:

```yaml
performance:
  rate_limit:
    enabled: true
    requests_per_minute: 60
```

### 7. Enable Circuit Breakers

Prevent cascading failures:

```yaml
performance:
  circuit_breaker:
    enabled: true
    failure_threshold: 5
```

### 8. Configure Logging Appropriately

```yaml
# Development
logging:
  level: debug
  format: console

# Production
logging:
  level: info
  format: json
  output: file
  compress: true
```

### 9. Validate Configuration Before Deployment

Test your configuration:

```bash
ainative-code validate-config --config configs/production.yaml
```

### 10. Document Custom Settings

Add comments to your configuration:

```yaml
llm:
  anthropic:
    # Increased timeout for long-running code generation tasks
    timeout: 60s
    # Higher temperature for creative code suggestions
    temperature: 0.9
```

## Troubleshooting

### Configuration Not Loading

1. Check file exists at expected location
2. Verify YAML syntax is valid
3. Check file permissions
4. Review error messages

### Environment Variables Not Working

1. Ensure proper naming with `AINATIVE_` prefix
2. Check for typos in variable names
3. Restart application after setting variables
4. Use `export` in shell

### Validation Errors

1. Read error message carefully
2. Check data types match expectations
3. Verify required fields are present
4. Ensure values are within valid ranges

### Connection Failures

1. Verify endpoint URLs are correct
2. Check network connectivity
3. Validate credentials
4. Review timeout settings
5. Check firewall rules

## Further Reading

- [Viper Documentation](https://github.com/spf13/viper) - Configuration library
- [Security Best Practices](./security.md) - Security guidelines
- [API Documentation](./api.md) - API reference
- [Deployment Guide](./deployment.md) - Production deployment

## Support

For issues or questions:
- GitHub Issues: https://github.com/AINative-studio/ainative-code/issues
- Documentation: https://docs.ainative.studio
- Community: https://community.ainative.studio
