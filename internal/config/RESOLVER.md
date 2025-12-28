# Dynamic API Key Resolution

The AINative Code configuration system includes a powerful dynamic API key resolution feature that allows you to store and retrieve API keys and secrets from multiple sources securely.

## Table of Contents

- [Overview](#overview)
- [Supported Formats](#supported-formats)
- [Configuration](#configuration)
- [Usage Examples](#usage-examples)
- [Security Considerations](#security-considerations)
- [Error Handling](#error-handling)
- [Testing](#testing)

## Overview

The resolver system provides a unified way to reference API keys and secrets in your configuration files without hardcoding sensitive values. This improves security by:

1. **Avoiding secret exposure** in configuration files
2. **Supporting multiple secret management tools** (pass, 1Password, etc.)
3. **Enabling environment-specific configurations** via environment variables
4. **Centralizing secret access** through file-based storage

## Supported Formats

The resolver automatically detects and processes the following formats:

### 1. Direct String (Hardcoded)

```yaml
llm:
  anthropic:
    api_key: "sk-ant-api-key-123456"
```

**Use case**: Development or testing (not recommended for production)

### 2. Environment Variable

```yaml
llm:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
```

**Environment setup**:
```bash
export ANTHROPIC_API_KEY="sk-ant-api-key-123456"
```

**Use case**:
- Container deployments (Docker, Kubernetes)
- CI/CD pipelines
- Cloud platforms (AWS, Azure, GCP)

### 3. File Path

```yaml
llm:
  anthropic:
    api_key: "~/secrets/anthropic-key.txt"
```

**Supported path formats**:
- Home directory: `~/secrets/api-key.txt`
- Absolute path: `/etc/secrets/api-key.txt`
- Relative path: `./secrets/api-key.txt`
- Parent directory: `../secrets/api-key.txt`

**File extensions recognized as key files**:
- `.txt`
- `.key`
- `.pem`
- `.secret`
- `.env`

**Use case**:
- Local development
- Mounted secrets in containers
- File-based secret management

### 4. Command Execution

```yaml
llm:
  anthropic:
    api_key: "$(pass show anthropic/api-key)"
```

**Supported commands**:
- `pass` - Password manager
- `1password` - 1Password CLI
- `aws` - AWS Secrets Manager
- `gcloud` - Google Cloud Secret Manager
- `vault` - HashiCorp Vault
- Custom scripts

**Use case**:
- Integration with password managers
- Cloud secret management services
- Custom secret retrieval scripts

## Configuration

### Basic Usage

The resolver is automatically initialized when you create a configuration loader:

```go
loader := config.NewLoader()
cfg, err := loader.Load()
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}
// API keys are automatically resolved
```

### Custom Resolver Configuration

You can customize the resolver behavior:

```go
import (
    "time"
    "github.com/AINative-studio/ainative-code/internal/config"
)

// Create a custom resolver
resolver := config.NewResolver(
    // Set command execution timeout
    config.WithCommandTimeout(10 * time.Second),

    // Whitelist allowed commands (security)
    config.WithAllowedCommands("pass", "1password", "aws"),

    // Disable command execution entirely
    config.WithCommandExecution(false),
)

// Use custom resolver with loader
loader := config.NewLoader(
    config.WithResolver(resolver),
)

cfg, err := loader.Load()
```

## Usage Examples

### Example 1: Mixed Sources Configuration

```yaml
# config.yaml
llm:
  anthropic:
    api_key: "$(pass show anthropic/api-key)"  # From password manager

  openai:
    api_key: "${OPENAI_API_KEY}"  # From environment variable

  google:
    api_key: "~/secrets/google-api-key.txt"  # From file

platform:
  authentication:
    api_key: "sk-platform-direct-key"  # Direct string (dev only)
```

### Example 2: AWS Secrets Manager

```yaml
llm:
  anthropic:
    api_key: "$(aws secretsmanager get-secret-value --secret-id anthropic-api-key --query SecretString --output text)"
```

**Prerequisites**:
```bash
# Configure AWS CLI
aws configure

# Verify access
aws secretsmanager get-secret-value --secret-id anthropic-api-key
```

### Example 3: HashiCorp Vault

```yaml
llm:
  anthropic:
    api_key: "$(vault kv get -field=api_key secret/anthropic)"
```

**Prerequisites**:
```bash
# Login to Vault
vault login

# Verify access
vault kv get secret/anthropic
```

### Example 4: 1Password CLI

```yaml
llm:
  anthropic:
    api_key: "$(op read op://Private/Anthropic/api_key)"
```

**Prerequisites**:
```bash
# Install 1Password CLI
brew install --cask 1password-cli

# Sign in
eval $(op signin)

# Verify access
op read op://Private/Anthropic/api_key
```

### Example 5: Docker Secrets

```yaml
llm:
  anthropic:
    api_key: "/run/secrets/anthropic_api_key"
```

**Docker Compose**:
```yaml
version: '3.8'
services:
  app:
    image: ainative-code:latest
    secrets:
      - anthropic_api_key

secrets:
  anthropic_api_key:
    file: ./secrets/anthropic_api_key.txt
```

### Example 6: Kubernetes Secrets

```yaml
llm:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
```

**Kubernetes Deployment**:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: api-keys
type: Opaque
stringData:
  anthropic-api-key: "sk-ant-api-key-123456"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ainative-code
spec:
  template:
    spec:
      containers:
      - name: app
        env:
        - name: ANTHROPIC_API_KEY
          valueFrom:
            secretKeyRef:
              name: api-keys
              key: anthropic-api-key
```

## Security Considerations

### Command Execution Security

**By default, command execution is enabled but should be restricted in production:**

```go
// Recommended: Whitelist specific commands
resolver := config.NewResolver(
    config.WithAllowedCommands(
        "pass",
        "1password",
        "aws",
        "vault",
    ),
)
```

**Alternative: Disable command execution:**

```go
resolver := config.NewResolver(
    config.WithCommandExecution(false),
)
```

### File Permissions

Ensure secret files have appropriate permissions:

```bash
# Set restrictive permissions
chmod 600 ~/secrets/api-key.txt
chown $USER:$USER ~/secrets/api-key.txt

# Verify
ls -la ~/secrets/api-key.txt
# Output: -rw------- 1 user user 64 Dec 27 10:00 api-key.txt
```

### Environment Variable Security

- **Never log environment variables** containing secrets
- Use **secret management in CI/CD** (GitHub Secrets, GitLab CI/CD Variables)
- **Rotate secrets regularly**
- **Limit access** to environment variables

### Best Practices

1. **Never commit secrets** to version control
2. **Use .gitignore** for secret files:
   ```gitignore
   secrets/
   *.key
   *.pem
   .env.local
   ```
3. **Use different secrets** for each environment
4. **Implement secret rotation** policies
5. **Audit secret access** regularly
6. **Use read-only permissions** where possible

## Error Handling

The resolver provides detailed error messages for troubleshooting:

### Common Errors

#### 1. Missing Environment Variable

```
Error: environment variable ANTHROPIC_API_KEY
Configuration error: Required setting 'environment variable ANTHROPIC_API_KEY' is not configured.
```

**Solution**: Set the environment variable
```bash
export ANTHROPIC_API_KEY="your-key-here"
```

#### 2. File Not Found

```
Error: file does not exist: /path/to/key.txt
Invalid configuration for 'api_key_file': file does not exist: /path/to/key.txt
```

**Solution**: Verify file path and existence
```bash
ls -la /path/to/key.txt
```

#### 3. Command Execution Failed

```
Error: command execution failed: exec: "pass": executable file not found in $PATH
```

**Solution**: Install the required command
```bash
# For pass
brew install pass  # macOS
apt-get install pass  # Ubuntu/Debian
```

#### 4. Command Timeout

```
Error: command execution timed out after 5s
```

**Solution**: Increase timeout or optimize command
```go
resolver := config.NewResolver(
    config.WithCommandTimeout(30 * time.Second),
)
```

#### 5. Command Not Allowed

```
Error: command 'curl' is not in the allowed commands list
Security error: The requested operation is not permitted.
```

**Solution**: Add command to whitelist
```go
resolver := config.NewResolver(
    config.WithAllowedCommands("pass", "1password", "curl"),
)
```

## Testing

### Unit Testing with Mocked Secrets

```go
func TestConfigWithSecrets(t *testing.T) {
    // Set test environment
    os.Setenv("TEST_API_KEY", "sk-test-key-123")
    defer os.Unsetenv("TEST_API_KEY")

    // Create config with env var
    cfg := &config.Config{
        LLM: config.LLMConfig{
            Anthropic: &config.AnthropicConfig{
                APIKey: "${TEST_API_KEY}",
            },
        },
    }

    // Load and resolve
    loader := config.NewLoader()
    err := loader.resolveAPIKeys(cfg)
    require.NoError(t, err)

    // Verify resolution
    assert.Equal(t, "sk-test-key-123", cfg.LLM.Anthropic.APIKey)
}
```

### Integration Testing

```go
func TestResolverIntegration(t *testing.T) {
    // Create temporary secret file
    tmpDir := t.TempDir()
    keyFile := filepath.Join(tmpDir, "api-key.txt")
    err := os.WriteFile(keyFile, []byte("sk-integration-key"), 0600)
    require.NoError(t, err)

    // Test resolution
    resolver := config.NewResolver()
    result, err := resolver.Resolve(keyFile)
    require.NoError(t, err)
    assert.Equal(t, "sk-integration-key", result)
}
```

### Testing All Resolution Methods

```go
func TestAllResolutionMethods(t *testing.T) {
    resolver := config.NewResolver()

    // Test direct string
    result, err := resolver.Resolve("sk-direct-key")
    assert.NoError(t, err)
    assert.Equal(t, "sk-direct-key", result)

    // Test environment variable
    os.Setenv("MY_KEY", "sk-env-key")
    defer os.Unsetenv("MY_KEY")
    result, err = resolver.Resolve("${MY_KEY}")
    assert.NoError(t, err)
    assert.Equal(t, "sk-env-key", result)

    // Test command
    result, err = resolver.Resolve("$(/bin/echo sk-cmd-key)")
    assert.NoError(t, err)
    assert.Equal(t, "sk-cmd-key", result)

    // Test file
    tmpFile := createTempKeyFile(t, "sk-file-key")
    result, err = resolver.Resolve(tmpFile)
    assert.NoError(t, err)
    assert.Equal(t, "sk-file-key", result)
}
```

## Resolution Priority

When the resolver processes a value, it checks in this order:

1. **Command Execution** - Pattern: `$(command)`
2. **Environment Variable** - Pattern: `${VAR_NAME}`
3. **File Path** - Detected by path characteristics
4. **Direct String** - Fallback for all other values

This ensures that the most dynamic and secure methods are tried first.

## Limitations

1. **File Size**: Maximum 1MB for security files
2. **Command Timeout**: Default 5 seconds (configurable)
3. **Environment Variable Names**: Must match `[A-Za-z_][A-Za-z0-9_]*`
4. **Command Output**: Must be non-empty after trimming whitespace

## Advanced Topics

### Custom Resolution Logic

For advanced use cases, you can implement custom resolution logic:

```go
type CustomResolver struct {
    *config.Resolver
    cache map[string]string
}

func (r *CustomResolver) Resolve(value string) (string, error) {
    // Check cache first
    if cached, ok := r.cache[value]; ok {
        return cached, nil
    }

    // Fall back to standard resolution
    result, err := r.Resolver.Resolve(value)
    if err != nil {
        return "", err
    }

    // Cache the result
    r.cache[value] = result
    return result, nil
}
```

### Audit Logging

Track secret access for security auditing:

```go
type AuditingResolver struct {
    *config.Resolver
    logger *log.Logger
}

func (r *AuditingResolver) Resolve(value string) (string, error) {
    start := time.Now()
    result, err := r.Resolver.Resolve(value)

    r.logger.Printf(
        "Secret resolution: pattern=%s, success=%v, duration=%v",
        detectPattern(value),
        err == nil,
        time.Since(start),
    )

    return result, err
}
```

## Related Documentation

- [Configuration Schema](./README.md)
- [Security Best Practices](../security/README.md)
- [Error Handling](../errors/README.md)
- [Testing Guide](../../docs/testing.md)

## Support

For issues or questions about API key resolution:

1. Check the [Error Handling](#error-handling) section
2. Review the [Examples](#usage-examples)
3. Consult the [Security Considerations](#security-considerations)
4. Open an issue on GitHub
