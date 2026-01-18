# Authentication with AINative Cloud

## Overview

AINative uses JWT-based authentication for secure access to hosted LLM services. This guide explains how authentication works and how to manage your credentials.

## How Authentication Works

The authentication flow uses JSON Web Tokens (JWT) with access and refresh tokens:

1. **Login**: Exchange email/password for JWT tokens
2. **Access Token**: Short-lived token (typically 15-30 minutes) used for API requests
3. **Refresh Token**: Long-lived token (typically 7 days) used to obtain new access tokens
4. **Auto-Refresh**: The CLI can automatically refresh tokens when needed

### Token Lifecycle

```
User Login
    ↓
[Email/Password] → AINative Backend
    ↓
[Access Token + Refresh Token]
    ↓
Stored in Config File
    ↓
Used for API Requests
    ↓
Token Expires → Auto Refresh → New Access Token
```

## Authentication Commands

### Login

Authenticate with your AINative account:

```bash
ainative-code auth login-backend \
  --email user@example.com \
  --password yourpassword
```

**Short form:**
```bash
ainative-code auth login-backend -e user@example.com -p yourpassword
```

**Response:**
```
Successfully logged in as user@example.com
```

**What happens:**
- Validates credentials with AINative backend
- Receives access and refresh tokens
- Saves tokens to `~/.config/ainative-code/config.yaml`
- Saves user email and ID to config

### Logout

Clear your stored credentials:

```bash
ainative-code auth logout-backend
```

**What happens:**
- Notifies backend of logout (if token is valid)
- Clears access token from config
- Clears refresh token from config
- Removes user email and ID from config

**Response:**
```
Successfully logged out
```

### Refresh Token

Manually refresh your access token:

```bash
ainative-code auth refresh-backend
```

**When to use:**
- Your access token has expired
- You want to ensure a fresh token before a critical operation
- Testing token refresh functionality

**Response:**
```
Token refreshed successfully
```

### Check Authentication Status

View your current authentication status:

```bash
ainative-code auth whoami
```

**Response:**
```
Authenticated User:
  Email: user@example.com
  Token Type: Bearer
  Expires In: 15m 30s
```

## Token Storage

Tokens are stored in the configuration file for security and convenience.

### Storage Location

**Config File:** `~/.config/ainative-code/config.yaml`

**Stored Information:**
```yaml
access_token: "eyJhbGc..."
refresh_token: "eyJhbGc..."
user_email: "user@example.com"
user_id: "user123"
```

### Security Considerations

1. **File Permissions**: The config file should have restricted permissions (600)
2. **Never Commit Tokens**: Add config files to `.gitignore`
3. **Rotate Regularly**: Tokens should be refreshed or re-issued regularly
4. **Secure Environment**: Use secure systems for storing credentials

## Configuration

### Backend URL

Specify the backend URL in your config file:

```yaml
backend_url: "http://localhost:8000"
```

Or use an environment variable:

```bash
export AINATIVE_BACKEND_URL="http://localhost:8000"
```

### Auto-Refresh (Future Feature)

Token auto-refresh will be available in a future release. Currently, you must manually refresh tokens using `auth refresh-backend` when they expire.

## Token Expiration

### Access Token Expiration

Access tokens typically expire after 15-30 minutes. When an access token expires:

1. API requests will return a 401 Unauthorized error
2. You'll see an error message: "unauthorized" or "token expired"
3. Run `ainative-code auth refresh-backend` to get a new access token

### Refresh Token Expiration

Refresh tokens typically expire after 7 days. When a refresh token expires:

1. Token refresh will fail
2. You must log in again with your email and password
3. Run `ainative-code auth login-backend` to authenticate again

## Security Best Practices

### 1. Never Share Your Tokens

Access tokens provide full access to your account. Never:
- Share tokens in chat messages or emails
- Commit tokens to version control
- Expose tokens in screenshots or logs

### 2. Use Environment Variables for CI/CD

For automated systems, use environment variables instead of config files:

```bash
export AINATIVE_ACCESS_TOKEN="your-token"
export AINATIVE_REFRESH_TOKEN="your-refresh-token"
```

### 3. Rotate Tokens Regularly

- Refresh access tokens before they expire
- Log out and log in periodically to get new refresh tokens
- Immediately refresh tokens if you suspect compromise

### 4. Enable 2FA on Your Account

Enable two-factor authentication on your AINative account at https://ainative.studio for enhanced security.

### 5. Use Separate Accounts for Dev/Prod

Use different accounts for:
- Development and testing
- Production deployments
- Team collaboration

## Troubleshooting

### Error: "not authenticated"

**Cause:** No valid access token found in config

**Solution:**
```bash
ainative-code auth login-backend --email your@email.com --password yourpass
```

### Error: "unauthorized" (401)

**Cause:** Access token has expired or is invalid

**Solution:**
```bash
ainative-code auth refresh-backend
```

If refresh fails, log in again:
```bash
ainative-code auth login-backend --email your@email.com --password yourpass
```

### Error: "invalid credentials"

**Cause:** Incorrect email or password

**Solution:**
1. Verify your email address
2. Verify your password
3. Reset password at https://ainative.studio if needed

### Error: "connection refused"

**Cause:** Python backend is not running

**Solution:**
```bash
cd /path/to/python-backend
uvicorn app.main:app --reload
```

### Token Refresh Fails

**Cause:** Refresh token has expired

**Solution:** Log in again to get new tokens:
```bash
ainative-code auth login-backend --email your@email.com --password yourpass
```

## Advanced Usage

### Custom Backend URL

Specify a custom backend URL for testing or production:

```bash
# In config file
backend_url: "https://api.ainative.studio"

# Or via environment variable
export AINATIVE_BACKEND_URL="https://api.ainative.studio"
```

### Programmatic Authentication

For integrations and scripts, you can authenticate programmatically:

```bash
# Non-interactive login (for scripts)
ainative-code auth login-backend -e "$EMAIL" -p "$PASSWORD"

# Check if authenticated
if ainative-code auth whoami > /dev/null 2>&1; then
  echo "Authenticated"
else
  echo "Not authenticated"
  exit 1
fi
```

### Multiple Accounts

To switch between accounts:

```bash
# Logout current account
ainative-code auth logout-backend

# Login with different account
ainative-code auth login-backend -e other@email.com -p password
```

## Next Steps

- [Hosted Inference Guide](hosted-inference.md) - Learn how to use chat and completions
- [Provider Configuration Guide](provider-configuration.md) - Configure provider preferences
- [Troubleshooting Guide](troubleshooting.md) - Solve common issues
