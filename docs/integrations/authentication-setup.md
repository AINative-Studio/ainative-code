# Authentication Setup Guide

## Overview

This comprehensive guide covers all authentication methods in AINative Code, including OAuth 2.0, API keys, and service-specific authentication for LLM providers and platform integrations.

## Table of Contents

1. [Quick Start](#quick-start)
2. [OAuth 2.0 Authentication](#oauth-20-authentication)
3. [API Key Management](#api-key-management)
4. [LLM Provider Authentication](#llm-provider-authentication)
5. [Platform Service Authentication](#platform-service-authentication)
6. [Security Best Practices](#security-best-practices)
7. [Troubleshooting](#troubleshooting)

## Quick Start

### First-Time Setup

```bash
# 1. Login to AINative platform
ainative-code auth login

# 2. Configure LLM provider (Anthropic example)
export ANTHROPIC_API_KEY="sk-ant-api03-..."

# 3. Verify authentication
ainative-code auth whoami

# 4. Test with a chat command
ainative-code chat "Hello, world!"
```

That's it! You're authenticated and ready to use AINative Code.

## OAuth 2.0 Authentication

OAuth 2.0 with PKCE is used for authenticating with the AINative platform services (ZeroDB, Design System, Strapi, etc.).

### Login Process

**Interactive Login:**

```bash
ainative-code auth login
```

**What happens:**

1. CLI starts local callback server on port 8080
2. Opens your default browser to AINative login page
3. You enter your credentials
4. You authorize the application
5. Browser redirects back to CLI
6. CLI receives and stores tokens securely

**Expected Output:**

```
🔐 Initiating authentication flow...
🌐 Opening browser for login...
⏳ Waiting for authorization...

✓ Authentication successful!
📧 Authenticated as: john@example.com
⏰ Access token expires in: 59 minutes

Tokens stored securely in OS keychain.
```

### Custom Port

If port 8080 is already in use:

```bash
ainative-code auth login --port 8081
```

### Headless/Server Environment

For environments without a browser:

```bash
# 1. Generate auth URL
ainative-code auth login --print-url

# Output:
# Please visit this URL to authenticate:
# https://auth.ainative.studio/oauth/authorize?...

# 2. Copy URL and open in browser on another machine
# 3. After authorization, copy the code parameter
# 4. Complete authentication
ainative-code auth login --code "authorization-code-here"
```

### Check Authentication Status

```bash
ainative-code auth whoami
```

**Output:**

```
Authentication Status
═══════════════════

User: john@example.com
Token Status: Valid ✓
Expires At: 2024-01-15 15:30:00 (in 45 minutes)
Refresh Token: Valid ✓
Scopes: read, write, offline_access

Platform Services:
  ZeroDB: Connected ✓
  Strapi: Connected ✓
  Design System: Connected ✓
```

### Token Refresh

Tokens are automatically refreshed when they expire. Manual refresh:

```bash
ainative-code auth token refresh
```

### Logout

```bash
ainative-code auth logout
```

This removes tokens from the keychain and revokes them on the server.

## API Key Management

API keys are used for LLM providers (Anthropic, OpenAI, Google) and some platform services.

### Environment Variables

**Recommended Method:**

```bash
# Add to ~/.bashrc or ~/.zshrc
export ANTHROPIC_API_KEY="sk-ant-api03-..."
export OPENAI_API_KEY="sk-..."
export GOOGLE_API_KEY="..."
export ZERODB_API_KEY="..."
export STRAPI_API_KEY="..."
export FIGMA_TOKEN="..."
```

**Load environment file:**

```bash
# Create .env file
cat > .env << 'EOF'
ANTHROPIC_API_KEY=sk-ant-api03-...
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=...
EOF

# Load variables
source .env
# or
export $(cat .env | xargs)
```

### Configuration File

```yaml
# ~/.config/ainative-code/config.yaml
llm:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"  # References env var

  openai:
    api_key: "${OPENAI_API_KEY}"

  google:
    api_key: "${GOOGLE_API_KEY}"

services:
  zerodb:
    api_key: "${ZERODB_API_KEY}"

  strapi:
    api_key: "${STRAPI_API_KEY}"

  design:
    figma_token: "${FIGMA_TOKEN}"
```

### Keychain Storage

**Store API key in OS keychain:**

```bash
ainative-code config set llm.anthropic.api_key --secure
# Prompts for API key (hidden input)
# Stores securely in OS keychain
```

**Retrieve from keychain:**

```bash
ainative-code config get llm.anthropic.api_key
# Output: sk-ant-****...****
```

### API Key Validation

```bash
# Test Anthropic API key
ainative-code test provider anthropic

# Test OpenAI API key
ainative-code test provider openai

# Test all providers
ainative-code test providers
```

**Output:**

```
Testing LLM Providers
═══════════════════

Anthropic:
  API Key: ✓ Valid
  Model: claude-3-5-sonnet-20241022
  Status: ✓ Available

OpenAI:
  API Key: ✓ Valid
  Model: gpt-4-turbo-preview
  Status: ✓ Available

Google:
  API Key: ✗ Not configured
```

## LLM Provider Authentication

### Anthropic Claude

**Get API Key:**

1. Visit https://console.anthropic.com/
2. Sign up or log in
3. Navigate to Settings > API Keys
4. Create new API key
5. Copy the key (starts with `sk-ant-api03-`)

**Configure:**

```bash
export ANTHROPIC_API_KEY="sk-ant-api03-..."

# Or use secure storage
ainative-code config set llm.anthropic.api_key --secure
```

**Test:**

```bash
ainative-code chat --provider anthropic "Hello, Claude!"
```

### OpenAI GPT

**Get API Key:**

1. Visit https://platform.openai.com/
2. Sign up or log in
3. Navigate to API Keys
4. Create new secret key
5. Copy the key (starts with `sk-`)

**Configure:**

```bash
export OPENAI_API_KEY="sk-..."

# Or use secure storage
ainative-code config set llm.openai.api_key --secure
```

**Test:**

```bash
ainative-code chat --provider openai "Hello, GPT!"
```

### Google Gemini

**Get API Key:**

1. Visit https://makersuite.google.com/app/apikey
2. Sign in with Google account
3. Create API key
4. Copy the key

**Configure:**

```bash
export GOOGLE_API_KEY="..."

# Or use secure storage
ainative-code config set llm.google.api_key --secure
```

**Test:**

```bash
ainative-code chat --provider google "Hello, Gemini!"
```

### Azure OpenAI

**Get Credentials:**

1. Create Azure OpenAI resource
2. Get endpoint URL
3. Get API key from Keys and Endpoint section

**Configure:**

```yaml
# config.yaml
llm:
  azure_openai:
    endpoint: "https://your-resource.openai.azure.com/"
    api_key: "${AZURE_OPENAI_API_KEY}"
    api_version: "2024-02-15-preview"
    deployment_name: "gpt-4"
```

**Test:**

```bash
export AZURE_OPENAI_API_KEY="..."
ainative-code chat --provider azure_openai "Hello!"
```

## Platform Service Authentication

### ZeroDB

**Get Credentials:**

1. Login to https://zerodb.ainative.studio
2. Create or select project
3. Copy Project ID and API Key

**Configure:**

```bash
export ZERODB_PROJECT_ID="your-project-id"
export ZERODB_API_KEY="your-api-key"
```

**Test:**

```bash
/zerodb-project-info
```

### Strapi CMS

**Get API Token:**

1. Login to your Strapi instance
2. Navigate to Settings > API Tokens
3. Create new token with appropriate permissions
4. Copy the token

**Configure:**

```bash
export STRAPI_URL="https://cms.ainative.studio"
export STRAPI_API_KEY="your-api-token"

# Or use CLI
ainative-code strapi config --url "https://cms.ainative.studio"
ainative-code strapi config --token "your-api-token"
```

**Test:**

```bash
ainative-code strapi test
```

### Figma (Design Tokens)

**Get Personal Access Token:**

1. Login to Figma
2. Go to Settings > Account
3. Scroll to Personal Access Tokens
4. Generate new token
5. Copy the token

**Configure:**

```bash
export FIGMA_TOKEN="your-figma-token"
```

**Test:**

```bash
ainative-code design extract \
  --source figma \
  --file-id "ABC123" \
  --output test-tokens.json
```

## Security Best Practices

### 1. Never Commit API Keys

```bash
# Add to .gitignore
echo ".env" >> .gitignore
echo "config.yaml" >> .gitignore
echo "*.key" >> .gitignore
```

### 2. Use Environment Variables

```bash
# Good: Environment variable
export ANTHROPIC_API_KEY="sk-ant-..."

# Bad: Hardcoded in scripts
API_KEY="sk-ant-..."  # Never do this
```

### 3. Rotate Keys Regularly

```bash
# Rotate every 90 days
# 1. Generate new key in provider console
# 2. Update environment variable
# 3. Test with new key
# 4. Revoke old key
```

### 4. Use Separate Keys for Different Environments

```bash
# Development
export ANTHROPIC_API_KEY="sk-ant-dev-..."

# Production
export ANTHROPIC_API_KEY="sk-ant-prod-..."

# CI/CD
export ANTHROPIC_API_KEY="sk-ant-ci-..."
```

### 5. Limit Key Permissions

- Create keys with minimum required permissions
- Use read-only keys where possible
- Create separate keys for different services

### 6. Monitor Key Usage

```bash
# Check usage and costs
ainative-code analytics providers

# Monitor for unusual activity
ainative-code analytics usage --provider anthropic
```

### 7. Secure Storage

```bash
# Use OS keychain
ainative-code config set llm.anthropic.api_key --secure

# Or use secrets manager
# AWS Secrets Manager
# Google Secret Manager
# HashiCorp Vault
```

### 8. Environment Separation

```bash
# ~/.bashrc (development)
export ANTHROPIC_API_KEY="dev-key"

# CI/CD pipeline (production)
# Use secrets management from CI platform
# GitHub Actions: ${{ secrets.ANTHROPIC_API_KEY }}
# GitLab CI: $ANTHROPIC_API_KEY
```

## Troubleshooting

### Invalid API Key

**Problem:** "Invalid API key" error

**Solutions:**

```bash
# 1. Verify key format
echo $ANTHROPIC_API_KEY | wc -c  # Should be ~50+ characters

# 2. Check for whitespace
export ANTHROPIC_API_KEY=$(echo $ANTHROPIC_API_KEY | tr -d '[:space:]')

# 3. Test key directly
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{"model": "claude-3-haiku-20240307", "max_tokens": 10, "messages": [{"role": "user", "content": "Hi"}]}'

# 4. Regenerate key if invalid
```

### OAuth Login Fails

**Problem:** Browser doesn't open or callback fails

**Solutions:**

```bash
# 1. Check port availability
lsof -i :8080

# 2. Use different port
ainative-code auth login --port 8081

# 3. Manual URL method
ainative-code auth login --print-url

# 4. Check firewall
sudo ufw status  # Linux
# Ensure port 8080 is allowed

# 5. Clear browser cache
# Try incognito/private mode
```

### Token Expired

**Problem:** "Token expired" error

**Solutions:**

```bash
# 1. Refresh token
ainative-code auth token refresh

# 2. Check token status
ainative-code auth whoami

# 3. Re-login if refresh fails
ainative-code auth logout
ainative-code auth login
```

### Keychain Access Denied

**Problem:** Cannot access OS keychain

**Solutions:**

**macOS:**

```bash
# Grant access in System Preferences
# Security & Privacy > Privacy > Automation

# Or use file-based storage
ainative-code config set security.use_keychain false
```

**Linux:**

```bash
# Install keyring
sudo apt install gnome-keyring  # Ubuntu
sudo dnf install gnome-keyring  # Fedora

# Start keyring daemon
eval $(gnome-keyring-daemon --start)
```

### Provider Not Available

**Problem:** "Provider not available" error

**Solutions:**

```bash
# 1. Check provider is configured
ainative-code config get llm.anthropic

# 2. Set default provider
ainative-code config set llm.default_provider anthropic

# 3. Verify API key
ainative-code test provider anthropic

# 4. Check for outages
# Visit provider status page
```

### Environment Variables Not Loading

**Problem:** Environment variables not recognized

**Solutions:**

```bash
# 1. Verify variables are exported
export ANTHROPIC_API_KEY="..."

# 2. Check if set
env | grep ANTHROPIC

# 3. Source env file
source ~/.bashrc  # or ~/.zshrc

# 4. Use full path
ANTHROPIC_API_KEY=sk-ant-... ainative-code chat "test"
```

## Next Steps

- [ZeroDB Integration](zerodb-integration.md)
- [Design Token Integration](design-token-integration.md)
- [Strapi CMS Integration](strapi-integration.md)
- [RLHF Feedback System](rlhf-integration.md)

## Resources

- [OAuth 2.0 Specification](https://oauth.net/2/)
- [PKCE RFC 7636](https://tools.ietf.org/html/rfc7636)
- [Anthropic API Documentation](https://docs.anthropic.com/)
- [OpenAI API Documentation](https://platform.openai.com/docs/)
- [Security Best Practices](../security/security-best-practices.md)
