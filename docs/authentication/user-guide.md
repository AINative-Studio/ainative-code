# Authentication User Guide

## Overview

This guide explains how to authenticate with AINative Code CLI and manage your authentication tokens. AINative Code uses OAuth 2.0 with PKCE for secure authentication with the AINative platform.

## Quick Start

### First-Time Setup

1. **Initialize configuration**:
   ```bash
   ainative-code setup
   ```

2. **Log in to AINative platform**:
   ```bash
   ainative-code auth login
   ```

3. **Verify authentication**:
   ```bash
   ainative-code auth whoami
   ```

That's it! You're now authenticated and can use all AINative platform features.

## Authentication Commands

### Login

Authenticate with the AINative platform using your browser.

```bash
ainative-code auth login
```

**What happens**:
1. CLI opens your default web browser
2. You log in with your AINative credentials
3. You authorize the CLI application
4. Tokens are stored securely in your OS keychain
5. CLI confirms successful authentication

**Example Output**:
```
Initiating authentication flow...
Opening browser for authentication...
Please complete the authorization in your browser

✓ Authentication successful!
Tokens stored securely in OS keychain
Access token expires in: 3600 seconds

To view your authentication status, run: ainative-code auth whoami
```

**Flags**:

| Flag | Description | Default |
|------|-------------|---------|
| `--auth-url` | Authorization endpoint URL | `https://auth.ainative.studio/oauth/authorize` |
| `--token-url` | Token endpoint URL | `https://auth.ainative.studio/oauth/token` |
| `--client-id` | OAuth client ID | `ainative-code-cli` |
| `--scopes` | OAuth scopes to request | `read,write,offline_access` |

**Example with Custom Scopes**:
```bash
ainative-code auth login --scopes read,write,admin
```

### Check Authentication Status

View information about the currently authenticated user.

```bash
ainative-code auth whoami
```

**Example Output**:
```
Authenticated User:
  Email: user@example.com
  Token Type: Bearer
  Expires In: 45m 23s
```

**If Not Authenticated**:
```
Not authenticated

Run 'ainative-code auth login' to authenticate
```

**If Token Expiring Soon**:
```
Authenticated User:
  Email: user@example.com
  Token Type: Bearer
  Expires In: 2m 15s
  ⚠️  Token expiring soon!
```

### Logout

Remove all stored credentials from your system.

```bash
ainative-code auth logout
```

**What happens**:
1. Deletes access tokens from keychain
2. Deletes refresh tokens from keychain
3. Clears user information
4. Confirms successful logout

**Example Output**:
```
✓ Successfully logged out
All credentials have been removed from OS keychain
```

**Note**: This clears local credentials only. To revoke server-side sessions, visit your account settings at https://ainative.studio/settings/sessions

## Token Management

### View Token Status

Display detailed information about your authentication tokens.

```bash
ainative-code auth token status
```

**Example Output**:
```
Token Status:
─────────────────────────────────────────
Access Token:  eyJhbGciOiJSUzI1Ni...
Refresh Token: eyJhbGciOiJSUzI1Ni...
Token Type:    Bearer
Expires At:    Mon, 06 Jan 2025 14:30:45 PST
Time Until Expiry: 45m 23s

Status: ✓ VALID

Auto-Refresh: Managed by background service
```

**Status Indicators**:

| Status | Meaning | Action |
|--------|---------|--------|
| ✓ VALID | Token is valid and not expiring soon | No action needed |
| ⚠️ EXPIRING SOON | Token expires in less than 5 minutes | Consider refreshing |
| ❌ EXPIRED | Token has expired | Refresh or re-authenticate |

**If No Tokens Found**:
```
No tokens found

Run 'ainative-code auth login' to authenticate
```

### Refresh Access Token

Manually refresh your access token using the stored refresh token.

```bash
ainative-code auth token refresh
```

**What happens**:
1. Retrieves refresh token from keychain
2. Sends refresh request to token server
3. Receives new access token
4. Updates keychain with new tokens
5. Confirms successful refresh

**Example Output**:
```
Refreshing access token...
✓ Token refreshed successfully
New token expires in: 3600 seconds
```

**When to use**:
- Token is expiring soon (< 5 minutes)
- Want to ensure fresh token before long-running operation
- Testing token refresh functionality

**Note**: The CLI automatically refreshes tokens when needed. Manual refresh is usually not required.

## Environment Variables

Configure authentication using environment variables:

### OAuth Configuration

```bash
# OAuth endpoints
export AINATIVE_AUTH_URL="https://auth.ainative.studio/oauth/authorize"
export AINATIVE_TOKEN_URL="https://auth.ainative.studio/oauth/token"

# OAuth client configuration
export AINATIVE_CLIENT_ID="ainative-code-cli"
export AINATIVE_REDIRECT_URI="http://localhost:8080/callback"

# OAuth scopes (comma-separated)
export AINATIVE_SCOPES="read,write,offline_access"

# Request timeout
export AINATIVE_AUTH_TIMEOUT="30s"
```

### JWT Configuration

```bash
# JWT validation
export AINATIVE_JWT_ISSUER="ainative-auth"
export AINATIVE_JWT_AUDIENCE="ainative-code"

# Public key for signature verification
export AINATIVE_JWT_PUBLIC_KEY_PATH="~/.ainative/jwt-public.pem"
```

### Keychain Configuration

```bash
# Keychain service name
export AINATIVE_KEYCHAIN_SERVICE="ainative-code"

# Auto-refresh settings
export AINATIVE_AUTO_REFRESH="true"
export AINATIVE_REFRESH_THRESHOLD="5m"
```

## Configuration File

You can also configure authentication in `~/.config/ainative-code/config.yaml`:

```yaml
platform:
  # OAuth 2.0 configuration
  oauth:
    client_id: "ainative-code-cli"
    auth_endpoint: "https://auth.ainative.studio/oauth/authorize"
    token_endpoint: "https://auth.ainative.studio/oauth/token"
    redirect_uri: "http://localhost:8080/callback"
    scopes:
      - "read"
      - "write"
      - "offline_access"
    timeout: 30s

  # JWT configuration
  jwt:
    issuer: "ainative-auth"
    audience: "ainative-code"
    public_key_path: "~/.ainative/jwt-public.pem"

  # Keychain configuration
  keychain:
    service_name: "ainative-code"
    auto_refresh: true
    refresh_threshold: 5m

  # Platform endpoints
  endpoints:
    api: "https://api.ainative.studio"
    zerodb: "https://zerodb.ainative.studio"
    strapi: "https://cms.ainative.studio"
```

## Session Management

### Automatic Token Refresh

AINative Code automatically refreshes your access token when:

- Access token has expired
- Access token will expire within 5 minutes
- Making an API request with expired token

**No manual intervention required** - the CLI handles refresh automatically.

### Token Lifetime

| Token Type | Lifetime | Purpose |
|------------|----------|---------|
| Access Token | 1 hour | API authentication |
| Refresh Token | 30 days | Obtain new access tokens |

**Refresh Strategy**:
- Access tokens are short-lived for security
- Refresh tokens allow getting new access tokens without re-login
- If refresh token expires, you must log in again

### Session Expiration

**When refresh token expires** (after 30 days of inactivity):
1. CLI detects refresh token expired
2. CLI prompts you to log in again
3. You complete login flow in browser
4. New tokens are issued and stored

**Example**:
```
Error: refresh token has expired

Run 'ainative-code auth login' to authenticate
```

## Security Best Practices

### Secure Token Storage

Tokens are automatically stored in OS-native secure storage:

- **macOS**: Keychain Access
- **Linux**: Secret Service (gnome-keyring, kwallet)
- **Windows**: Credential Manager

**Benefits**:
- Encrypted at rest
- Protected by OS access controls
- Isolated per user account
- No plaintext token files

### Access Token Security

**Do**:
- ✓ Let CLI manage tokens automatically
- ✓ Use `auth logout` when done on shared computers
- ✓ Keep your refresh token secret
- ✓ Log out if you suspect token compromise

**Don't**:
- ✗ Share access tokens with others
- ✗ Copy tokens to plaintext files
- ✗ Commit tokens to version control
- ✗ Send tokens in unencrypted channels

### Multi-Device Usage

**Each device should authenticate separately**:

```bash
# Device 1
device1$ ainative-code auth login
# Authenticate in browser

# Device 2
device2$ ainative-code auth login
# Authenticate in browser again
```

**Benefits**:
- Independent sessions per device
- Can revoke individual device sessions
- Better security and auditability

## Troubleshooting

### Browser Doesn't Open

**Problem**: `ainative-code auth login` doesn't open browser

**Solutions**:
1. Manually open the authorization URL printed in the terminal
2. Check default browser is set correctly
3. Try with `--auth-url` flag to specify custom endpoint

### Port Already in Use

**Problem**: `Error: failed to start callback server: port 8080 in use`

**Solution**: Kill process using port 8080 or use custom redirect URI:
```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill <PID>

# Or use custom redirect URI (requires server configuration)
ainative-code auth login --redirect-uri http://localhost:8081/callback
```

### Token Expired

**Problem**: `Error: token has expired`

**Solution**: Refresh token or re-authenticate:
```bash
# Try refresh first
ainative-code auth token refresh

# If refresh fails, login again
ainative-code auth login
```

### Keychain Access Denied

**Problem**: `Error: keychain access denied`

**Solutions**:

**macOS**:
1. Open Keychain Access app
2. Right-click "ainative-code" entry
3. Select "Get Info"
4. Update access permissions

**Linux**:
```bash
# Ensure gnome-keyring is running
ps aux | grep gnome-keyring

# Start if not running
gnome-keyring-daemon --start --daemonize
```

**Windows**:
1. Open Credential Manager
2. Check Windows Credentials
3. Verify no permission errors

For more troubleshooting, see [Troubleshooting Guide](troubleshooting.md).

## Common Workflows

### First-Time Setup

```bash
# 1. Initialize configuration
ainative-code setup

# 2. Authenticate
ainative-code auth login

# 3. Verify authentication
ainative-code auth whoami

# 4. Start using platform features
ainative-code zerodb query "SELECT * FROM users LIMIT 10"
```

### Daily Usage

```bash
# Check if still authenticated
ainative-code auth whoami

# If expired, login again
ainative-code auth login

# Use platform features
ainative-code chat "Help me write a REST API"
ainative-code design-tokens extract
```

### Switching Accounts

```bash
# Log out of current account
ainative-code auth logout

# Log in with different account
ainative-code auth login
# (Use different credentials in browser)

# Verify new account
ainative-code auth whoami
```

### Before Long-Running Operation

```bash
# Ensure fresh token
ainative-code auth token refresh

# Check token status
ainative-code auth token status

# Run long operation
ainative-code generate --complexity high "Build e-commerce platform"
```

## Integration with Platform Services

### Using Authentication with APIs

Once authenticated, access tokens are automatically included in requests:

```bash
# ZeroDB operations
ainative-code zerodb query "SELECT * FROM users"
# → Uses access token automatically

# Strapi CMS operations
ainative-code strapi sync
# → Uses access token automatically

# Design tokens
ainative-code design-tokens extract
# → Uses access token automatically
```

### Programmatic Access

For programmatic access from custom scripts:

```go
package main

import (
    "context"
    "fmt"

    "github.com/AINative-studio/ainative-code/internal/auth/keychain"
)

func main() {
    ctx := context.Background()

    // Get stored tokens
    kc := keychain.Get()
    tokens, err := kc.GetTokenPair()
    if err != nil {
        fmt.Println("Not authenticated. Run: ainative-code auth login")
        return
    }

    // Use access token for API requests
    accessToken := tokens.AccessToken

    // Make authenticated API request
    // req.Header.Set("Authorization", "Bearer " + accessToken)
}
```

## Next Steps

- [OAuth Flow Details](oauth-flow.md) - Deep dive into OAuth 2.0 PKCE flow
- [API Reference](api-reference.md) - Authentication API documentation
- [Security Best Practices](security-best-practices.md) - Advanced security guidelines
- [Troubleshooting](troubleshooting.md) - Solutions to common problems

## Support

If you encounter issues:

1. Check [Troubleshooting Guide](troubleshooting.md)
2. View detailed logs: `ainative-code --log-level debug auth login`
3. Report issues: https://github.com/AINative-studio/ainative-code/issues
4. Email support: support@ainative.studio
