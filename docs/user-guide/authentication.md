# Authentication Guide

This guide covers authentication with both AINative platform services and LLM providers.

## Table of Contents

1. [Overview](#overview)
2. [AINative Platform Authentication](#ainative-platform-authentication)
3. [LLM Provider Authentication](#llm-provider-authentication)
4. [Token Management](#token-management)
5. [Security Best Practices](#security-best-practices)
6. [Troubleshooting](#troubleshooting)

## Overview

AINative Code uses two types of authentication:

1. **Platform Authentication**: OAuth 2.0 for AINative platform services (ZeroDB, Strapi, RLHF)
2. **Provider Authentication**: API keys for LLM providers (Anthropic, OpenAI, Google, etc.)

## AINative Platform Authentication

### OAuth 2.0 Login

Login to access AINative platform features:

```bash
ainative-code auth login
```

This will:
1. Open your browser to the authentication page
2. Prompt you to authorize the application
3. Store tokens securely in your OS keychain
4. Enable automatic token refresh

### Login Flow

```
1. User runs: ainative-code auth login
2. CLI starts local callback server on port 8080
3. Opens browser to: https://auth.ainative.studio/authorize
4. User logs in and authorizes
5. Browser redirects to: http://localhost:8080/callback?code=...
6. CLI exchanges code for tokens
7. Tokens stored in OS keychain
8. Login complete
```

### Custom OAuth Configuration

For enterprise or custom deployments:

```bash
ainative-code auth login \
  --auth-url https://custom-auth.company.com/authorize \
  --token-url https://custom-auth.company.com/token \
  --client-id custom-client-id \
  --scopes read,write,admin
```

### Check Authentication Status

```bash
# View current user
ainative-code auth whoami

# Output:
# Authenticated User:
#   Email: user@example.com
#   Token Type: Bearer
#   Expires In: 2h 15m 30s
```

### Logout

Remove stored credentials:

```bash
ainative-code auth logout

# Output:
# Successfully logged out
# All credentials have been removed from OS keychain
```

## LLM Provider Authentication

### Anthropic Claude

**Get API Key:**
1. Visit [Anthropic Console](https://console.anthropic.com/)
2. Navigate to API Keys
3. Create a new key

**Configure:**

```bash
# Set API key
ainative-code config set llm.anthropic.api_key "sk-ant-api03-..."

# Or use environment variable (recommended)
export ANTHROPIC_API_KEY="sk-ant-api03-..."
```

**Verify:**

```bash
# Test the configuration
ainative-code chat --provider anthropic "Hello"
```

### OpenAI

**Get API Key:**
1. Visit [OpenAI Platform](https://platform.openai.com/api-keys)
2. Create new secret key
3. Copy the key (shown only once!)

**Configure:**

```bash
# Set API key
ainative-code config set llm.openai.api_key "sk-..."

# Or use environment variable
export OPENAI_API_KEY="sk-..."
```

**Organization (Optional):**

If you're part of multiple organizations:

```bash
ainative-code config set llm.openai.organization "org-..."
```

### Google Gemini

**Get API Key:**
1. Visit [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Create API key
3. Copy the key

**Configure:**

```bash
# Set API key
ainative-code config set llm.google.api_key "..."

# Or use environment variable
export GOOGLE_API_KEY="..."
```

**Vertex AI (Enterprise):**

For Vertex AI deployment:

```bash
# Authenticate with gcloud
gcloud auth application-default login

# Set project
ainative-code config set llm.google.project_id "your-gcp-project"
ainative-code config set llm.google.location "us-central1"
```

### AWS Bedrock

**Configure AWS Credentials:**

```bash
# Option 1: AWS CLI
aws configure

# Option 2: Environment variables
export AWS_ACCESS_KEY_ID="..."
export AWS_SECRET_ACCESS_KEY="..."
export AWS_REGION="us-east-1"

# Option 3: Config file
ainative-code config set llm.bedrock.region "us-east-1"
ainative-code config set llm.bedrock.profile "default"
```

**IAM Permissions:**

Ensure your IAM role has:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel",
        "bedrock:InvokeModelWithResponseStream"
      ],
      "Resource": "arn:aws:bedrock:*::foundation-model/*"
    }
  ]
}
```

### Azure OpenAI

**Get Credentials:**
1. Create Azure OpenAI resource in Azure Portal
2. Deploy a model
3. Get API key and endpoint

**Configure:**

```bash
# Set credentials
ainative-code config set llm.azure.api_key "..."
ainative-code config set llm.azure.endpoint "https://your-resource.openai.azure.com/"
ainative-code config set llm.azure.deployment_name "gpt-4"

# Or use environment variable
export AZURE_OPENAI_API_KEY="..."
```

**Managed Identity:**

For Azure VMs/services with managed identity:

```yaml
llm:
  azure:
    endpoint: https://your-resource.openai.azure.com/
    deployment_name: gpt-4
    # No api_key needed with managed identity
```

### Ollama (Local)

No authentication required for local Ollama:

```bash
# Just ensure Ollama is running
ollama serve

# Configure endpoint if non-default
ainative-code config set llm.ollama.base_url "http://localhost:11434"
```

## Token Management

### Access Tokens

**View Token Status:**

```bash
ainative-code auth token status

# Output:
# Token Status:
# ─────────────────────────────────────────
# Access Token:  eyJhbGciOiJSUzI1NiIs...
# Refresh Token: rt_1234567890abcdef...
# Token Type:    Bearer
# Expires At:    2024-01-20 14:30:00
# Time Until Expiry: 2h 15m 30s
#
# Status: VALID
```

### Refresh Tokens

**Manual Refresh:**

```bash
ainative-code auth token refresh

# Output:
# Token refreshed successfully
# New token expires in: 7200 seconds
```

**Automatic Refresh:**

Tokens are automatically refreshed:
- When they expire
- When they're within 5 minutes of expiring
- Before making API calls

**Disable Auto-Refresh:**

```yaml
platform:
  authentication:
    auto_refresh: false
```

### Token Storage

Tokens are stored securely in:

- **macOS**: Keychain
- **Linux**: Secret Service API (GNOME Keyring, KWallet)
- **Windows**: Credential Manager

**Location:**
- Platform tokens: OS keychain (service: `ainative-code`)
- LLM API keys: Config file or environment variables

### Token Security

**View Stored Tokens:**

```bash
# macOS
security find-generic-password -s ainative-code -w

# Linux (GNOME)
secret-tool lookup service ainative-code

# Windows
cmdkey /list | findstr ainative
```

**Delete Tokens:**

```bash
# Delete all stored credentials
ainative-code auth logout

# Delete specific provider tokens
ainative-code config unset llm.anthropic.api_key
```

## Security Best Practices

### 1. Never Commit API Keys

Add to `.gitignore`:

```
# .gitignore
.env
config.yaml
.ainative-code.yaml
*.key
tokens.json
```

### 2. Use Environment Variables

```bash
# .env file (add to .gitignore!)
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=...
```

Load before running:

```bash
source .env
ainative-code chat
```

Or use direnv:

```bash
# Install direnv
# Create .envrc
echo 'export ANTHROPIC_API_KEY=sk-ant-...' > .envrc

# Allow
direnv allow

# Automatically loaded when cd into directory
```

### 3. Use Separate Keys for Different Environments

```bash
# Development
export ANTHROPIC_API_KEY="${ANTHROPIC_DEV_KEY}"

# Production
export ANTHROPIC_API_KEY="${ANTHROPIC_PROD_KEY}"
```

### 4. Rotate Keys Regularly

```bash
# Every 90 days:
# 1. Generate new API key
# 2. Update configuration
# 3. Delete old key
```

### 5. Restrict API Key Permissions

When creating API keys:
- Use least privilege necessary
- Set usage limits
- Enable IP restrictions if available
- Set expiration dates

### 6. Monitor API Usage

```bash
# Check usage regularly
ainative-code analytics providers

# Set up alerts for unusual usage
```

### 7. Encrypt Configuration

For sensitive deployments:

```yaml
security:
  encrypt_config: true
  encryption_key: "${CONFIG_ENCRYPTION_KEY}"
```

### 8. Use OS Keychain

Let the OS manage secrets securely:

```bash
# Store in keychain instead of config file
ainative-code config set llm.anthropic.api_key "$(pass show anthropic-api-key)"
```

### 9. Audit Authentication

```bash
# Enable audit logging
logging:
  audit:
    enabled: true
    log_file: /var/log/ainative-code/audit.log
    include_auth_events: true
```

### 10. Secure Token Files

```bash
# Set restrictive permissions
chmod 600 ~/.config/ainative-code/config.yaml
chmod 700 ~/.config/ainative-code
```

## Troubleshooting

### Authentication Failed

**Check Credentials:**

```bash
# Verify API key is set
ainative-code config get llm.anthropic.api_key

# Or check environment
echo $ANTHROPIC_API_KEY
```

**Test API Key:**

```bash
# Test Anthropic key
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "max_tokens": 10,
    "messages": [{"role": "user", "content": "Hi"}]
  }'

# Test OpenAI key
curl https://api.openai.com/v1/chat/completions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hi"}],
    "max_tokens": 10
  }'
```

### Token Expired

**Refresh Token:**

```bash
ainative-code auth token refresh
```

**Re-login if Refresh Fails:**

```bash
ainative-code auth logout
ainative-code auth login
```

### OAuth Callback Failed

**Port Already in Use:**

```bash
# Use different port
ainative-code auth login --redirect-url http://localhost:8081/callback
```

**Firewall Blocking:**

```bash
# Temporarily allow port 8080
sudo ufw allow 8080/tcp  # Linux
# Or add Windows Firewall rule
```

**Browser Not Opening:**

```bash
# Manually open the URL shown in terminal
# Copy the callback URL after authorization
```

### Invalid API Key Format

**Check Key Format:**

- Anthropic: `sk-ant-api03-...`
- OpenAI: `sk-...`
- Google: Random alphanumeric string

**Verify No Extra Characters:**

```bash
# Remove whitespace
export ANTHROPIC_API_KEY=$(echo $ANTHROPIC_API_KEY | tr -d '[:space:]')
```

### Permission Denied (Keychain)

**macOS:**

```bash
# Grant terminal access to keychain
# System Preferences > Security & Privacy > Privacy > Automation
```

**Linux:**

```bash
# Install keyring
sudo apt install gnome-keyring  # Ubuntu
sudo dnf install gnome-keyring  # Fedora

# Or use alternative storage
ainative-code config set security.use_keychain false
```

**Windows:**

Run as administrator if permission issues persist.

### Rate Limited

**Check Rate Limits:**

```bash
# View current usage
ainative-code analytics providers

# Wait before retrying
# Or use different provider
ainative-code chat --provider openai
```

### Network/Proxy Issues

**Behind Corporate Proxy:**

```bash
# Set proxy
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=http://proxy.company.com:8080

# Or in config
ainative-code config set network.http_proxy "http://proxy.company.com:8080"
```

**SSL Certificate Issues:**

```bash
# Use custom CA bundle
export SSL_CERT_FILE=/path/to/ca-bundle.crt

# Or disable SSL verification (not recommended for production)
ainative-code config set security.tls_verify false
```

## Multiple Accounts

### Different Profiles

```bash
# Create profiles for different accounts
ainative-code auth login --profile work
ainative-code auth login --profile personal

# Use specific profile
ainative-code --profile work chat
ainative-code --profile personal chat
```

### Switching Accounts

```bash
# Logout current account
ainative-code auth logout

# Login with different account
ainative-code auth login
```

## Next Steps

- [Configuration Guide](configuration.md) - Configure authentication settings
- [Providers Guide](providers.md) - Provider-specific setup
- [AINative Integrations](ainative-integrations.md) - Platform features
- [Security Best Practices](../security.md) - Advanced security
