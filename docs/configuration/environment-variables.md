# Environment Variables

AINative Code supports configuration via environment variables with the `AINATIVE_CODE_` prefix. This is particularly useful for CI/CD environments, containerized deployments, and keeping secrets out of configuration files.

## Configuration Precedence

Configuration values are loaded in the following order (highest priority first):

1. **Command-line flags** (e.g., `--provider openai`)
2. **Environment variables** (e.g., `AINATIVE_CODE_PROVIDER=openai`)
3. **Configuration file** (e.g., `~/.ainative-code.yaml`)
4. **Default values**

## Environment Variable Naming

Environment variables follow this pattern:
```
AINATIVE_CODE_<CONFIG_KEY>
```

For nested configuration keys, dots (`.`) and dashes (`-`) are replaced with underscores (`_`):

| Config Key | Environment Variable |
|------------|---------------------|
| `provider` | `AINATIVE_CODE_PROVIDER` |
| `model` | `AINATIVE_CODE_MODEL` |
| `api_key` | `AINATIVE_CODE_API_KEY` |
| `llm.anthropic.api_key` | `AINATIVE_CODE_LLM_ANTHROPIC_API_KEY` |
| `llm.openai.model` | `AINATIVE_CODE_LLM_OPENAI_MODEL` |
| `max-tokens` | `AINATIVE_CODE_MAX_TOKENS` |

## Supported Environment Variables

### Basic Configuration

```bash
# Provider selection
export AINATIVE_CODE_PROVIDER=openai          # openai, anthropic, google, azure, bedrock, ollama
export AINATIVE_CODE_MODEL=gpt-4              # Model name
export AINATIVE_CODE_VERBOSE=true             # Enable verbose logging
export AINATIVE_CODE_API_KEY=sk-...           # Generic API key
```

### LLM Provider Configuration

#### Anthropic Claude
```bash
export AINATIVE_CODE_LLM_ANTHROPIC_API_KEY=sk-ant-...
export AINATIVE_CODE_LLM_ANTHROPIC_MODEL=claude-3-5-sonnet-20241022
export AINATIVE_CODE_LLM_ANTHROPIC_MAX_TOKENS=8192
export AINATIVE_CODE_LLM_ANTHROPIC_TEMPERATURE=0.7
export AINATIVE_CODE_LLM_ANTHROPIC_TOP_P=1.0
export AINATIVE_CODE_LLM_ANTHROPIC_TOP_K=0
export AINATIVE_CODE_LLM_ANTHROPIC_BASE_URL=https://api.anthropic.com
```

#### OpenAI
```bash
export AINATIVE_CODE_LLM_OPENAI_API_KEY=sk-...
export AINATIVE_CODE_LLM_OPENAI_MODEL=gpt-4-turbo-preview
export AINATIVE_CODE_LLM_OPENAI_ORGANIZATION=org-...
export AINATIVE_CODE_LLM_OPENAI_MAX_TOKENS=4096
export AINATIVE_CODE_LLM_OPENAI_TEMPERATURE=0.7
export AINATIVE_CODE_LLM_OPENAI_TOP_P=1.0
export AINATIVE_CODE_LLM_OPENAI_BASE_URL=https://api.openai.com/v1
```

#### Google Gemini
```bash
export AINATIVE_CODE_LLM_GOOGLE_API_KEY=...
export AINATIVE_CODE_LLM_GOOGLE_MODEL=gemini-pro
export AINATIVE_CODE_LLM_GOOGLE_PROJECT_ID=my-project
export AINATIVE_CODE_LLM_GOOGLE_LOCATION=us-central1
export AINATIVE_CODE_LLM_GOOGLE_MAX_TOKENS=4096
export AINATIVE_CODE_LLM_GOOGLE_TEMPERATURE=0.7
```

#### AWS Bedrock
```bash
export AINATIVE_CODE_LLM_BEDROCK_REGION=us-east-1
export AINATIVE_CODE_LLM_BEDROCK_MODEL=anthropic.claude-3-sonnet-20240229-v1:0
export AINATIVE_CODE_LLM_BEDROCK_ACCESS_KEY_ID=AKIA...
export AINATIVE_CODE_LLM_BEDROCK_SECRET_ACCESS_KEY=...
export AINATIVE_CODE_LLM_BEDROCK_SESSION_TOKEN=...    # Optional
export AINATIVE_CODE_LLM_BEDROCK_PROFILE=default     # Optional, for AWS CLI profiles
```

#### Azure OpenAI
```bash
export AINATIVE_CODE_LLM_AZURE_API_KEY=...
export AINATIVE_CODE_LLM_AZURE_ENDPOINT=https://your-resource.openai.azure.com
export AINATIVE_CODE_LLM_AZURE_DEPLOYMENT_NAME=gpt-4
export AINATIVE_CODE_LLM_AZURE_API_VERSION=2024-02-15-preview
export AINATIVE_CODE_LLM_AZURE_MAX_TOKENS=4096
```

#### Ollama (Local)
```bash
export AINATIVE_CODE_LLM_OLLAMA_BASE_URL=http://localhost:11434
export AINATIVE_CODE_LLM_OLLAMA_MODEL=llama3
export AINATIVE_CODE_LLM_OLLAMA_MAX_TOKENS=4096
export AINATIVE_CODE_LLM_OLLAMA_KEEP_ALIVE=5m
```

### Default Provider
```bash
export AINATIVE_CODE_LLM_DEFAULT_PROVIDER=anthropic
```

### AINative Platform Services

#### Authentication
```bash
export AINATIVE_CODE_PLATFORM_AUTHENTICATION_METHOD=api_key    # api_key, jwt, oauth2
export AINATIVE_CODE_PLATFORM_AUTHENTICATION_API_KEY=...
export AINATIVE_CODE_PLATFORM_AUTHENTICATION_TOKEN=...
export AINATIVE_CODE_PLATFORM_AUTHENTICATION_CLIENT_ID=...
export AINATIVE_CODE_PLATFORM_AUTHENTICATION_CLIENT_SECRET=...
```

#### ZeroDB
```bash
export AINATIVE_CODE_SERVICES_ZERODB_ENABLED=true
export AINATIVE_CODE_SERVICES_ZERODB_PROJECT_ID=...
export AINATIVE_CODE_SERVICES_ZERODB_ENDPOINT=https://api.zerodb.ainative.studio
export AINATIVE_CODE_SERVICES_ZERODB_DATABASE=mydb
export AINATIVE_CODE_SERVICES_ZERODB_USERNAME=...
export AINATIVE_CODE_SERVICES_ZERODB_PASSWORD=...
export AINATIVE_CODE_SERVICES_ZERODB_SSL=true
```

#### Design Tokens
```bash
export AINATIVE_CODE_SERVICES_DESIGN_ENABLED=true
export AINATIVE_CODE_SERVICES_DESIGN_ENDPOINT=https://design.ainative.studio
export AINATIVE_CODE_SERVICES_DESIGN_API_KEY=...
```

#### Strapi CMS
```bash
export AINATIVE_CODE_SERVICES_STRAPI_ENABLED=true
export AINATIVE_CODE_SERVICES_STRAPI_ENDPOINT=https://cms.ainative.studio
export AINATIVE_CODE_SERVICES_STRAPI_API_KEY=...
```

#### RLHF Service
```bash
export AINATIVE_CODE_SERVICES_RLHF_ENABLED=true
export AINATIVE_CODE_SERVICES_RLHF_ENDPOINT=https://rlhf.ainative.studio
export AINATIVE_CODE_SERVICES_RLHF_API_KEY=...
export AINATIVE_CODE_SERVICES_RLHF_MODEL_ID=...
```

### Application Settings
```bash
export AINATIVE_CODE_APP_NAME=ainative-code
export AINATIVE_CODE_APP_VERSION=0.1.0
export AINATIVE_CODE_APP_ENVIRONMENT=production    # development, staging, production
export AINATIVE_CODE_APP_DEBUG=false
```

### Logging
```bash
export AINATIVE_CODE_LOGGING_LEVEL=info           # debug, info, warn, error
export AINATIVE_CODE_LOGGING_FORMAT=json          # json, console
export AINATIVE_CODE_LOGGING_OUTPUT=stdout        # stdout, file
export AINATIVE_CODE_LOGGING_FILE_PATH=/var/log/ainative-code.log
export AINATIVE_CODE_LOGGING_MAX_SIZE=100         # MB
export AINATIVE_CODE_LOGGING_MAX_BACKUPS=3
export AINATIVE_CODE_LOGGING_MAX_AGE=7            # days
export AINATIVE_CODE_LOGGING_COMPRESS=true
```

### Performance Settings
```bash
# Cache
export AINATIVE_CODE_PERFORMANCE_CACHE_ENABLED=true
export AINATIVE_CODE_PERFORMANCE_CACHE_TYPE=memory    # memory, redis
export AINATIVE_CODE_PERFORMANCE_CACHE_TTL=1h
export AINATIVE_CODE_PERFORMANCE_CACHE_MAX_SIZE=100   # MB

# Rate Limiting
export AINATIVE_CODE_PERFORMANCE_RATE_LIMIT_ENABLED=true
export AINATIVE_CODE_PERFORMANCE_RATE_LIMIT_REQUESTS_PER_MINUTE=60
export AINATIVE_CODE_PERFORMANCE_RATE_LIMIT_BURST_SIZE=10

# Concurrency
export AINATIVE_CODE_PERFORMANCE_CONCURRENCY_MAX_WORKERS=10
export AINATIVE_CODE_PERFORMANCE_CONCURRENCY_MAX_QUEUE_SIZE=100
```

### Security Settings
```bash
export AINATIVE_CODE_SECURITY_ENCRYPT_CONFIG=true
export AINATIVE_CODE_SECURITY_ENCRYPTION_KEY=...
export AINATIVE_CODE_SECURITY_TLS_ENABLED=true
export AINATIVE_CODE_SECURITY_TLS_CERT_PATH=/path/to/cert.pem
export AINATIVE_CODE_SECURITY_TLS_KEY_PATH=/path/to/key.pem
```

## Usage Examples

### Basic Usage
```bash
# Set provider and model via environment variables
export AINATIVE_CODE_PROVIDER=openai
export AINATIVE_CODE_MODEL=gpt-4

# Run the application
ainative-code chat
```

### CI/CD Integration
```bash
# GitHub Actions example
- name: Run AINative Code
  env:
    AINATIVE_CODE_PROVIDER: anthropic
    AINATIVE_CODE_LLM_ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
    AINATIVE_CODE_LOGGING_LEVEL: debug
  run: |
    ainative-code config show
```

### Docker/Container Usage
```dockerfile
# Dockerfile
ENV AINATIVE_CODE_PROVIDER=openai
ENV AINATIVE_CODE_LLM_OPENAI_API_KEY=${OPENAI_API_KEY}
ENV AINATIVE_CODE_LOGGING_FORMAT=json
ENV AINATIVE_CODE_LOGGING_OUTPUT=stdout
```

```bash
# docker-compose.yml
services:
  ainative-code:
    image: ainative-code:latest
    environment:
      - AINATIVE_CODE_PROVIDER=anthropic
      - AINATIVE_CODE_LLM_ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - AINATIVE_CODE_SERVICES_ZERODB_ENABLED=true
      - AINATIVE_CODE_SERVICES_ZERODB_ENDPOINT=https://api.zerodb.dev
```

### Override Config File Values
```bash
# Config file has provider=openai
# Override with environment variable
export AINATIVE_CODE_PROVIDER=anthropic

# Check effective configuration
ainative-code config show
# Output: provider: anthropic (from env var, not config file)
```

### Testing Different Providers
```bash
# Test with OpenAI
AINATIVE_CODE_PROVIDER=openai \
AINATIVE_CODE_LLM_OPENAI_API_KEY=sk-... \
AINATIVE_CODE_LLM_OPENAI_MODEL=gpt-4 \
  ainative-code chat "Hello world"

# Test with Anthropic
AINATIVE_CODE_PROVIDER=anthropic \
AINATIVE_CODE_LLM_ANTHROPIC_API_KEY=sk-ant-... \
AINATIVE_CODE_LLM_ANTHROPIC_MODEL=claude-3-opus \
  ainative-code chat "Hello world"
```

## Variable Substitution in Config Files

In addition to direct environment variables, you can use `${VAR_NAME}` syntax in config files to reference environment variables:

```yaml
# config.yaml
llm:
  openai:
    api_key: "${OPENAI_API_KEY}"       # Resolved from environment
    model: "gpt-4"

  anthropic:
    api_key: "$(pass show anthropic)"  # Resolved from password manager
    model: "claude-3-5-sonnet"
```

This allows you to:
- Keep secrets out of config files
- Use different credentials per environment
- Integrate with password managers and secret vaults

## Security Best Practices

1. **Never commit API keys to version control**
   ```bash
   # Use environment variables instead
   export AINATIVE_CODE_LLM_OPENAI_API_KEY=sk-...
   ```

2. **Use secret management in CI/CD**
   ```yaml
   # GitHub Actions
   env:
     AINATIVE_CODE_LLM_ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
   ```

3. **Restrict environment variable access**
   ```bash
   # Use .env files with restricted permissions
   chmod 600 .env
   ```

4. **Rotate credentials regularly**
   ```bash
   # Use the AINATIVE_CODE_SECURITY_SECRET_ROTATION setting
   export AINATIVE_CODE_SECURITY_SECRET_ROTATION=30d
   ```

## Troubleshooting

### Environment variables not being recognized
```bash
# Check if env vars are set
env | grep AINATIVE_CODE_

# Verify precedence
ainative-code config show

# Get specific value
ainative-code config get llm.openai.api_key
```

### Debugging configuration loading
```bash
# Enable verbose output
export AINATIVE_CODE_VERBOSE=true
ainative-code config show

# Or use the -v flag
ainative-code -v config show
```

### Testing environment variable expansion
```bash
# Set a test variable
export TEST_VALUE="hello-world"

# Reference it in config
echo 'test_key: "${TEST_VALUE}"' > test-config.yaml

# Verify it's expanded
ainative-code --config test-config.yaml config get test_key
```

## See Also

- [Configuration Guide](../configuration/README.md)
- [Security Best Practices](../security/README.md)
- [CI/CD Integration](../ci-cd/README.md)
