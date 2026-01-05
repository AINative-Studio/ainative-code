# API Key Resolution Security Guide

## Overview

The AINative Code platform provides a secure, flexible API key resolution system that supports multiple input formats while maintaining strict security controls. This document outlines the resolution methods, security considerations, configuration examples, and best practices.

## Resolution Methods

The API key resolver supports four distinct input formats, processed in priority order:

### 1. Command Execution (`$(command)`)

Execute a shell command to retrieve an API key from external sources like password managers or secret vaults.

**Format:**
```
$(command args...)
```

**Examples:**
```yaml
# Password Store (pass)
anthropic:
  api_key: $(pass show anthropic/api-key)

# macOS Keychain
openai:
  api_key: $(security find-generic-password -s openai -w)

# HashiCorp Vault
google:
  api_key: $(vault kv get -field=key secret/api-keys/google)

# AWS Secrets Manager
bedrock:
  access_key_id: $(aws secretsmanager get-secret-value --secret-id bedrock/access-key --query SecretString --output text)
```

**Security Features:**
- **Timeout Enforcement:** Commands timeout after 5 seconds (configurable)
- **No Shell Interpretation:** Commands execute directly without shell expansion
- **Whitelist Support:** Optional command whitelist for strict control
- **Output Sanitization:** Stdout is captured and trimmed of whitespace
- **Error Handling:** Command failures return clear error messages

**Security Considerations:**
- Command injection is prevented by using `exec.CommandContext` with separate arguments
- Shell metacharacters (`;`, `&&`, `||`, `|`, etc.) are treated as literal arguments
- Commands can be disabled entirely via configuration
- Whitelist enforcement ensures only approved commands execute

### 2. Environment Variables (`${VAR_NAME}` or `$VAR_NAME`)

Resolve API keys from environment variables with support for both braced and unbraced syntax.

**Format:**
```
${VARIABLE_NAME}  # Preferred syntax
$VARIABLE_NAME    # Also supported
```

**Examples:**
```yaml
# With braces (recommended)
anthropic:
  api_key: ${ANTHROPIC_API_KEY}

# Without braces
openai:
  api_key: $OPENAI_API_KEY

# Platform authentication
platform:
  authentication:
    api_key: ${AINATIVE_PLATFORM_KEY}
```

**Security Features:**
- **Pattern Validation:** Only valid variable names are accepted (`[A-Za-z_][A-Za-z0-9_]*`)
- **No Default Values:** Missing variables return explicit errors
- **No Recursive Expansion:** Variable values are treated as literals
- **Injection Prevention:** Variable values containing commands are not executed

**Security Considerations:**
- Environment variables should be set with restricted permissions
- Never commit `.env` files with secrets to version control
- Use environment-specific variable names to prevent conflicts
- Consider using secret management tools to populate environment variables

### 3. File Paths (`~/path/to/file` or `/absolute/path`)

Read API keys from files with comprehensive security controls.

**Format:**
```
~/relative/path/to/key.txt
/absolute/path/to/key.txt
./relative/path/to/key.txt
../parent/path/to/key.txt
```

**Examples:**
```yaml
# Home directory reference
anthropic:
  api_key: ~/.secrets/anthropic-api-key.txt

# Absolute path
openai:
  api_key: /etc/secrets/openai/api-key

# Relative path
google:
  api_key: ./config/secrets/google.key
```

**Security Features:**
- **Size Limit:** Files must be ≤ 1KB (API keys should be small)
- **Path Traversal Prevention:** Paths are cleaned via `filepath.Clean`
- **Symlink Resolution:** Symlinks are evaluated via `filepath.EvalSymlinks`
- **Null Byte Detection:** Paths containing `\x00` are rejected
- **Permission Validation:** Permission errors are clearly reported
- **Directory Detection:** Directories are rejected (files only)

**File Permissions Best Practices:**
```bash
# Create secret file with restricted permissions
touch ~/.secrets/api-key.txt
chmod 600 ~/.secrets/api-key.txt  # Owner read/write only

# Verify permissions
ls -la ~/.secrets/api-key.txt
# -rw------- 1 user user 64 Jan 01 12:00 api-key.txt
```

**Security Considerations:**
- Store API key files outside the project directory
- Use restrictive file permissions (600 or 400)
- Never include secret files in version control
- Consider encrypting files at rest
- Regularly rotate API keys and update files
- Use separate files per environment (dev, staging, prod)

### 4. Direct Strings (Fallback)

Provide API keys directly as strings in configuration files.

**Format:**
```
sk-ant-api03-...
sk-proj-...
AIzaSy...
```

**Examples:**
```yaml
# Direct API key (NOT RECOMMENDED for production)
anthropic:
  api_key: sk-ant-api03-1234567890abcdef

# Better for local development only
ollama:
  base_url: http://localhost:11434
```

**Security Considerations:**
- **AVOID in production:** Direct strings expose secrets in configuration files
- **Acceptable for local development:** Use with local-only services (e.g., Ollama)
- **Version Control Risk:** Easy to accidentally commit secrets
- **Use alternatives:** Prefer environment variables, files, or commands for real API keys

## Resolution Priority

The resolver processes inputs in the following order:

1. **Command Execution** - `$(command)` pattern matched
2. **Environment Variable** - `${VAR}` or `$VAR` pattern matched
3. **File Path** - Path-like patterns matched
4. **Direct String** - Fallback for unmatched inputs

This priority ensures that dynamic sources are tried before static values.

## Configuration Options

### Global Resolver Configuration

```go
import "github.com/AINative-studio/ainative-code/internal/config"

// Create resolver with custom options
resolver := config.NewResolver(
    // Set command timeout (default: 5 seconds)
    config.WithCommandTimeout(10 * time.Second),

    // Enable command whitelist
    config.WithAllowedCommands("pass", "security", "vault"),

    // Disable command execution entirely
    config.WithCommandExecution(false),
)

// Use resolver
apiKey, err := resolver.Resolve("$(pass show anthropic)")
```

### Loader Integration

The resolver is automatically integrated into the configuration loader:

```go
// Create loader with custom resolver
loader := config.NewLoader(
    config.WithResolver(resolver),
)

// Load configuration (API keys are resolved automatically)
cfg, err := loader.Load()
```

## Security Best Practices

### 1. Principle of Least Privilege

- Grant minimal permissions to API key files and commands
- Use dedicated service accounts for production deployments
- Restrict file system access to necessary directories only

### 2. Secret Rotation

```bash
# Rotate API keys regularly
# 1. Generate new API key from provider
# 2. Update secret storage
pass edit anthropic/api-key  # Update password store

# 3. Restart application to load new key
# 4. Revoke old API key from provider
```

### 3. Environment Separation

```yaml
# Development: Use environment variables
# .env.development
ANTHROPIC_API_KEY=sk-ant-dev-key

# Staging: Use command execution
# config.staging.yaml
anthropic:
  api_key: $(vault kv get -field=key secret/staging/anthropic)

# Production: Use external secret management
# config.production.yaml
anthropic:
  api_key: $(aws secretsmanager get-secret-value --secret-id prod/anthropic --query SecretString --output text)
```

### 4. Audit and Monitoring

- Log API key resolution attempts (without logging actual keys)
- Monitor for unusual command execution patterns
- Track file access to secret storage locations
- Alert on permission errors or timeout failures

### 5. Encryption at Rest

```bash
# Use encrypted filesystems for secret storage
# macOS: FileVault
# Linux: LUKS, dm-crypt
# Cloud: KMS-encrypted volumes

# Encrypt individual secret files
gpg --encrypt --recipient you@example.com api-key.txt
# Resolve: $(gpg --decrypt ~/.secrets/api-key.txt.gpg)
```

### 6. CI/CD Integration

```yaml
# GitHub Actions example
env:
  ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}

# GitLab CI example
variables:
  ANTHROPIC_API_KEY:
    vault: secret/data/ci/anthropic/api_key@key

# Use in config
anthropic:
  api_key: ${ANTHROPIC_API_KEY}
```

### 7. Development Best Practices

```bash
# Use direnv for local development
# .envrc
export ANTHROPIC_API_KEY=$(pass show anthropic/api-key)
export OPENAI_API_KEY=$(pass show openai/api-key)

# Load with direnv allow
direnv allow

# Keys are available as environment variables
# config.yaml
anthropic:
  api_key: ${ANTHROPIC_API_KEY}
```

## Error Handling

The resolver provides clear, actionable error messages:

### Command Execution Errors

```
Error: command execution failed: command not found
Resolution: Verify the command exists and is in PATH

Error: command execution timed out after 5s
Resolution: Increase timeout or optimize command performance

Error: command 'wget' is not in the allowed commands list
Resolution: Add command to whitelist or use approved alternative
```

### Environment Variable Errors

```
Error: missing required configuration: environment variable ANTHROPIC_API_KEY
Resolution: Set the environment variable before running

Error: environment variable API_KEY is empty
Resolution: Ensure the variable is set to a non-empty value
```

### File Path Errors

```
Error: file does not exist: ~/.secrets/api-key.txt
Resolution: Create the file or verify the path is correct

Error: permission denied reading file: /etc/secrets/api.key
Resolution: Grant read permissions or run with appropriate user

Error: file is too large (max 1024 bytes, got 2048 bytes)
Resolution: API keys should be small; verify file contains only the key

Error: file path contains null byte
Resolution: Security violation detected; use clean file paths

Error: path is a directory, not a file: ~/.secrets/
Resolution: Specify the full path to the key file
```

## Testing

### Unit Tests

```go
func TestAPIKeyResolution(t *testing.T) {
    resolver := config.NewResolver()

    // Test environment variable
    os.Setenv("TEST_KEY", "sk-test-123")
    defer os.Unsetenv("TEST_KEY")

    result, err := resolver.Resolve("${TEST_KEY}")
    require.NoError(t, err)
    assert.Equal(t, "sk-test-123", result)
}
```

### Security Tests

Comprehensive security tests are located in `tests/security/api_key_resolution_test.go`:

- Command injection prevention
- Path traversal prevention
- Symlink handling
- File size limits
- Permission validation
- Timeout enforcement
- Whitelist enforcement
- Concurrent access safety

### Integration Tests

```bash
# Test with real password manager
echo "sk-test-key" | pass insert -e test/api-key
go test -v -run TestRealPasswordManager

# Test with macOS Keychain
security add-generic-password -s test-api -a user -w sk-test-key
go test -v -run TestRealKeychain

# Cleanup
pass rm test/api-key
security delete-generic-password -s test-api
```

## Troubleshooting

### Common Issues

**Issue:** API key not found
```
Solution:
1. Verify the resolution method syntax
2. Check environment variables: env | grep KEY
3. Verify file exists: ls -la ~/.secrets/api-key.txt
4. Test command manually: pass show anthropic
```

**Issue:** Permission denied
```
Solution:
1. Check file permissions: ls -la /path/to/key
2. Fix permissions: chmod 600 /path/to/key
3. Verify user ownership: chown user:user /path/to/key
```

**Issue:** Command timeout
```
Solution:
1. Increase timeout in resolver configuration
2. Optimize command performance
3. Verify network connectivity for remote commands
```

**Issue:** Security violation
```
Solution:
1. Review path for traversal attempts
2. Check for null bytes in input
3. Verify command is whitelisted
4. Ensure file size is within limits
```

## Recommended Tools

### Password Managers

- **pass (Password Store):** Unix password manager using GPG
  ```bash
  pass insert anthropic/api-key
  pass show anthropic/api-key
  ```

- **1Password CLI:** 1Password command-line interface
  ```bash
  op read "op://Private/Anthropic/api_key"
  ```

- **Bitwarden CLI:** Open-source password manager
  ```bash
  bw get password anthropic-api-key
  ```

### Secret Management Systems

- **HashiCorp Vault:** Enterprise secret management
  ```bash
  vault kv get -field=key secret/api-keys/anthropic
  ```

- **AWS Secrets Manager:** Cloud secret storage
  ```bash
  aws secretsmanager get-secret-value --secret-id anthropic --query SecretString --output text
  ```

- **Google Secret Manager:** GCP secret storage
  ```bash
  gcloud secrets versions access latest --secret=anthropic-api-key
  ```

- **Azure Key Vault:** Azure secret management
  ```bash
  az keyvault secret show --vault-name myvault --name anthropic-api-key --query value -o tsv
  ```

### macOS Keychain

```bash
# Store API key
security add-generic-password -s anthropic-api -a user -w "sk-ant-key-123"

# Retrieve API key
security find-generic-password -s anthropic-api -w

# Delete API key
security delete-generic-password -s anthropic-api
```

## Migration Guide

### Migrating from Direct Strings

**Before:**
```yaml
anthropic:
  api_key: sk-ant-api03-hardcoded-key-123
```

**After (using environment variable):**
```yaml
anthropic:
  api_key: ${ANTHROPIC_API_KEY}
```
```bash
export ANTHROPIC_API_KEY="sk-ant-api03-hardcoded-key-123"
```

**After (using password manager):**
```yaml
anthropic:
  api_key: $(pass show anthropic/api-key)
```
```bash
echo "sk-ant-api03-hardcoded-key-123" | pass insert -e anthropic/api-key
```

### Migrating from Custom Scripts

If you have existing custom secret retrieval scripts, integrate them:

```yaml
# Old approach: Custom script loaded separately
# New approach: Direct integration
anthropic:
  api_key: $(./scripts/get-api-key.sh anthropic)
```

## Compliance Considerations

### PCI DSS

- API keys are not stored in plaintext in configuration files
- Access to key material is logged and auditable
- Keys are encrypted at rest when using file-based storage
- Timeout controls prevent long-running credential theft

### SOC 2

- Segregation of duties via command whitelisting
- Access control through file permissions
- Audit trail via error logging
- Incident response via clear error messages

### GDPR

- API keys can be rotated without code changes
- Secret access is controlled and traceable
- No unnecessary retention of key material

## Support

For questions or issues:

1. Review this documentation
2. Check the troubleshooting section
3. Review security tests for examples
4. File an issue with reproduction steps

## References

- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
- [NIST Secret Management Guidelines](https://csrc.nist.gov/publications/)
- [CIS Controls for Secret Management](https://www.cisecurity.org/controls/)
- [Cloud Security Alliance Secrets Management](https://cloudsecurityalliance.org/)
