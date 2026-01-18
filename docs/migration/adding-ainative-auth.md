# Migration Guide: Adding AINative Authentication

## Overview

This guide helps existing AINative Code users migrate from direct API key authentication to AINative cloud authentication and hosted inference.

## Why Migrate?

### Benefits of AINative Cloud Authentication

| Feature | Before (API Keys) | After (AINative Cloud) |
|---------|-------------------|------------------------|
| **Authentication** | Separate API keys per provider | Single login for all providers |
| **Billing** | Multiple invoices from different providers | Unified credit-based billing |
| **Key Management** | Manually manage multiple API keys | No API keys to manage |
| **Provider Selection** | Manual provider switching | Automatic intelligent routing |
| **Security** | API keys in config/env files | Tokens with auto-refresh |
| **Credit Tracking** | Monitor across multiple dashboards | Unified dashboard |
| **Flexibility** | Locked to paid provider plans | Pay-as-you-go credits |

### What You Gain

- **Simplified Authentication**: One login for all providers
- **Cost Optimization**: Pay only for tokens used, no monthly subscriptions
- **Automatic Failover**: Switch providers when one is unavailable
- **Credit Monitoring**: Real-time visibility into consumption
- **Enhanced Security**: JWT tokens with automatic refresh

## Migration Path

### Prerequisites

1. **AINative Account**: Sign up at https://ainative.studio
2. **Python Backend**: Ensure backend is running
3. **Backup Config**: Save your current config file

```bash
cp ~/.config/ainative-code/config.yaml ~/.config/ainative-code/config.yaml.backup
```

### Step-by-Step Migration

#### Step 1: Sign Up for AINative Account

1. Visit https://ainative.studio
2. Click "Sign Up"
3. Verify your email
4. Add initial credits (if prompted)

#### Step 2: Start Python Backend

The AINative integration requires the Python backend:

```bash
cd /Users/aideveloper/AINative-Code/python-backend
uvicorn app.main:app --reload
```

Verify it's running:
```bash
curl http://localhost:8000/health
```

Expected response:
```json
{"status": "healthy"}
```

#### Step 3: Login via CLI

Authenticate with your AINative account:

```bash
ainative-code auth login-backend \
  --email your-email@example.com \
  --password your-password
```

Response:
```
Successfully logged in as your-email@example.com
```

#### Step 4: Test the Integration

Send a test message:

```bash
ainative-code chat-ainative \
  --message "Hello! This is a test." \
  --auto-provider
```

#### Step 5: Update Your Workflows

Replace old commands with new ones:

**Before:**
```bash
ainative-code chat \
  --message "Hello" \
  --provider anthropic \
  --api-key sk-ant-...
```

**After:**
```bash
ainative-code chat-ainative \
  --message "Hello" \
  --auto-provider
```

#### Step 6: Configure Provider Preferences

Set your preferred provider in config:

```yaml
# ~/.config/ainative-code/config.yaml
ainative:
  preferred_provider: anthropic  # or openai, google
  fallback_enabled: true
```

#### Step 7: Remove API Keys (Optional)

Once you've verified the integration works, you can remove API keys:

**From config file:**
```yaml
# Remove these sections
providers:
  anthropic:
    api_key: sk-ant-...  # Remove
  openai:
    api_key: sk-...      # Remove
```

**From environment:**
```bash
unset AINATIVE_CODE_LLM_ANTHROPIC_API_KEY
unset AINATIVE_CODE_LLM_OPENAI_API_KEY
```

**Keep your API keys backed up** in case you need to revert.

---

## Command Migration Reference

### Authentication Commands

| Old Command | New Command | Notes |
|-------------|-------------|-------|
| N/A (direct API keys) | `ainative-code auth login-backend` | New authentication method |
| N/A | `ainative-code auth logout-backend` | Clear stored tokens |
| N/A | `ainative-code auth refresh-backend` | Refresh access token |
| N/A | `ainative-code auth whoami` | Check auth status |

### Chat Commands

| Old Command | New Command | Key Changes |
|-------------|-------------|-------------|
| `ainative-code chat` | `ainative-code chat-ainative` | New backend integration |
| `--api-key sk-ant-...` | (none) | No longer needed |
| `--provider anthropic` | `--provider anthropic` or `--auto-provider` | Auto selection available |

### Example Migrations

**Example 1: Simple Chat**

**Before:**
```bash
ainative-code chat \
  --message "Explain REST APIs" \
  --provider anthropic \
  --api-key sk-ant-api03-...
```

**After:**
```bash
ainative-code chat-ainative \
  --message "Explain REST APIs" \
  --auto-provider
```

**Example 2: Specific Model**

**Before:**
```bash
ainative-code chat \
  --message "Write Python code" \
  --provider openai \
  --model gpt-4 \
  --api-key sk-...
```

**After:**
```bash
ainative-code chat-ainative \
  --message "Write Python code" \
  --model gpt-4 \
  --provider openai
```

**Example 3: Streaming**

**Before:**
```bash
ainative-code chat \
  --message "Long response" \
  --provider anthropic \
  --api-key sk-ant-... \
  --stream
```

**After:**
```bash
ainative-code chat-ainative \
  --message "Long response" \
  --stream \
  --auto-provider
```

---

## Configuration Migration

### Old Configuration Format

```yaml
# Old config.yaml
providers:
  anthropic:
    api_key: "sk-ant-api03-..."
    model: "claude-3-5-sonnet-20241022"
    max_tokens: 4096
    temperature: 0.7

  openai:
    api_key: "sk-..."
    model: "gpt-4"
    max_tokens: 4096
```

### New Configuration Format

```yaml
# New config.yaml
backend_url: "http://localhost:8000"

# Authentication (auto-populated after login)
access_token: "eyJhbGc..."
refresh_token: "eyJhbGc..."
user_email: "user@example.com"
user_id: "user123"

# AINative settings
ainative:
  preferred_provider: anthropic
  fallback_enabled: true
  provider_priority:
    - anthropic
    - openai
    - google

# Optional: Keep old provider configs for backward compatibility
providers:
  anthropic:
    # api_key no longer needed
    model: "claude-sonnet-4-5"
    max_tokens: 4096
    temperature: 0.7

  openai:
    # api_key no longer needed
    model: "gpt-4"
    max_tokens: 4096
```

---

## Backward Compatibility

### Both Methods Work

AINative Code supports BOTH authentication methods:

**Method 1: AINative Cloud (Recommended)**
```bash
ainative-code chat-ainative \
  --message "Hello" \
  --auto-provider
```

**Method 2: Direct API Keys (Still Supported)**
```bash
ainative-code chat \
  --message "Hello" \
  --provider anthropic \
  --api-key sk-ant-...
```

### Gradual Migration

You can migrate gradually:

1. Start using `chat-ainative` for new work
2. Keep existing scripts using `chat` with API keys
3. Migrate scripts one by one
4. Eventually remove API keys when fully migrated

---

## Rollback Plan

If you need to revert to API keys:

### Step 1: Restore Config Backup

```bash
cp ~/.config/ainative-code/config.yaml.backup ~/.config/ainative-code/config.yaml
```

### Step 2: Logout from AINative

```bash
ainative-code auth logout-backend
```

### Step 3: Use Original Commands

Return to using direct API keys:

```bash
ainative-code chat \
  --message "Hello" \
  --provider anthropic \
  --api-key sk-ant-...
```

---

## Troubleshooting Migration

### Issue: "Backend not running"

**Error:**
```
Error: connection refused
```

**Solution:**
```bash
cd /Users/aideveloper/AINative-Code/python-backend
uvicorn app.main:app --reload
```

### Issue: "Authentication failed"

**Error:**
```
Error: login failed: invalid credentials
```

**Solutions:**
1. Verify email and password
2. Check account is activated
3. Reset password at https://ainative.studio if needed

### Issue: "Commands not found"

**Error:**
```
Error: unknown command "chat-ainative"
```

**Solution:** Update to latest version:
```bash
brew upgrade ainative-code
# OR
curl -fsSL https://raw.githubusercontent.com/AINative-Studio/ainative-code/main/install.sh | bash
```

### Issue: "Old commands not working"

**Error:**
```
Old chat command broken after migration
```

**Solution:** Restore API keys in config or use `--api-key` flag

---

## What Changes?

### Architecture Changes

**Before:**
```
CLI → Direct to LLM Provider (Anthropic/OpenAI/Google)
    ↓
  API Key Authentication
```

**After:**
```
CLI → Python Backend → LLM Provider
    ↓         ↓
JWT Tokens  Provider Selection + Credit Management
```

### Workflow Changes

| Aspect | Before | After |
|--------|--------|-------|
| **Authentication** | API keys per provider | Single AINative login |
| **Provider Selection** | Manual via `--provider` | Automatic with `--auto-provider` |
| **Billing** | Per provider | Unified credits |
| **Token Management** | N/A (API keys don't expire) | Auto-refresh tokens |
| **Commands** | `chat` | `chat-ainative` |

---

## Best Practices After Migration

### 1. Set Preferred Provider

Always configure your preference:

```yaml
ainative:
  preferred_provider: anthropic
```

### 2. Enable Fallback

For reliability:

```yaml
ainative:
  fallback_enabled: true
```

### 3. Monitor Credits

Check regularly:

```bash
ainative-code auth whoami
```

### 4. Use Auto Provider

Let the system choose:

```bash
ainative-code chat-ainative -m "message" --auto-provider
```

### 5. Keep Backend Running

For development, keep backend running:

```bash
# Add to your shell profile
alias start-ainative-backend="cd /Users/aideveloper/AINative-Code/python-backend && uvicorn app.main:app --reload"
```

---

## Frequently Asked Questions

### Q: Do I need to remove my API keys?

**A:** No, both methods work. You can keep API keys for backward compatibility and use AINative cloud for new work.

### Q: Will my existing scripts break?

**A:** No, existing commands with API keys still work. The old `chat` command is still supported.

### Q: What if I run out of credits?

**A:** Add credits at https://ainative.studio/billing. The system will show warnings when credits are low.

### Q: Can I switch back to API keys?

**A:** Yes, just use the old `chat` command with `--api-key` flag or logout and restore your backup config.

### Q: Is the migration required?

**A:** No, it's optional. However, AINative cloud provides better developer experience with unified authentication and billing.

### Q: What about rate limits?

**A:** AINative cloud has its own rate limits, separate from provider rate limits. Check your account dashboard for details.

---

## Next Steps

After migrating:

1. **Read the Guides**:
   - [Getting Started Guide](../guides/ainative-getting-started.md)
   - [Authentication Guide](../guides/authentication.md)
   - [Hosted Inference Guide](../guides/hosted-inference.md)

2. **Configure Your Preferences**:
   - [Provider Configuration Guide](../guides/provider-configuration.md)

3. **Learn Troubleshooting**:
   - [Troubleshooting Guide](../guides/troubleshooting.md)

4. **Explore the API**:
   - [API Reference](../api/ainative-provider.md)

---

## Support

Need help with migration?

- **Documentation**: https://github.com/AINative-Studio/ainative-code/docs
- **GitHub Issues**: https://github.com/AINative-Studio/ainative-code/issues
- **Email**: support@ainative.studio
