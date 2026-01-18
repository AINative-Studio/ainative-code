# AINative CLI User Guide

## Overview

The AINative CLI now supports direct integration with the AINative backend platform for authentication and AI chat completions. This guide explains how to use the new commands.

## Table of Contents

1. [Authentication](#authentication)
2. [Chat Completions](#chat-completions)
3. [Provider Selection](#provider-selection)
4. [Configuration](#configuration)
5. [Troubleshooting](#troubleshooting)

## Authentication

### Login

Authenticate with your AINative account using email and password:

```bash
ainative-code auth login-backend --email your@email.com --password yourpassword
```

**What happens:**
1. Sends credentials to AINative backend
2. Receives access and refresh tokens
3. Stores tokens securely in configuration
4. Saves user information

**Success Output:**
```
Successfully logged in as your@email.com
```

**Flags:**
- `--email, -e`: Your AINative account email (required)
- `--password, -p`: Your account password (required)

**Example:**
```bash
ainative-code auth login-backend \
  -e developer@example.com \
  -p MySecurePassword123
```

### Logout

Clear stored credentials and notify the backend:

```bash
ainative-code auth logout-backend
```

**What happens:**
1. Calls backend logout endpoint (if token exists)
2. Clears access token
3. Clears refresh token
4. Removes user information

**Success Output:**
```
Successfully logged out
```

**Note:** Logout always clears local tokens, even if backend call fails.

### Refresh Token

Manually refresh your access token using the stored refresh token:

```bash
ainative-code auth refresh-backend
```

**What happens:**
1. Sends refresh token to backend
2. Receives new access and refresh tokens
3. Updates stored tokens

**Success Output:**
```
Token refreshed successfully
```

**When to use:**
- Access token expired
- Proactive refresh before long operations
- After receiving "unauthorized" errors

## Chat Completions

### Basic Chat

Send a message to an AI assistant:

```bash
ainative-code chat-ainative --message "Explain Test-Driven Development"
```

**What happens:**
1. Validates authentication
2. Sends message to AINative backend
3. Returns AI response

**Success Output:**
```
Test-Driven Development (TDD) is a software development approach where you write tests before writing the actual code...
```

**Flags:**
- `--message, -m`: Your message (required)
- `--model`: Specific model to use (optional)
- `--auto-provider`: Enable intelligent provider selection (optional)
- `--verbose`: Show usage statistics (optional)

### With Auto Provider Selection

Let the system intelligently select the best provider:

```bash
ainative-code chat-ainative \
  --message "Hello, AI!" \
  --auto-provider
```

**What happens:**
1. Checks your preferred provider setting
2. Verifies credit balance
3. Selects optimal provider based on:
   - User preferences
   - Credit availability
   - Provider capabilities
   - Fallback rules
4. Displays low credit warning if needed

**Low Credit Warning Example:**
```
Warning: Low credit balance (10 credits remaining)
Hello! How can I help you today?
```

### With Specific Model

Use a specific AI model:

```bash
ainative-code chat-ainative \
  --message "Explain goroutines in Go" \
  --model claude-sonnet-4-5
```

**Available Models:**
- `claude-sonnet-4-5` (default for Anthropic)
- `gpt-4` (OpenAI)
- `gemini-pro` (Google)

### With Verbose Output

Display detailed usage statistics:

```bash
ainative-code chat-ainative \
  --message "What is TDD?" \
  --verbose
```

**Verbose Output Example:**
```
TDD stands for Test-Driven Development...

Model: claude-sonnet-4-5
Tokens - Prompt: 8, Completion: 120, Total: 128
```

## Provider Selection

### How It Works

The provider selector intelligently routes your requests based on:

1. **User Preferences**: Your configured preferred provider
2. **Credit Balance**: Available credits in your account
3. **Capabilities**: Provider features (vision, function calling, etc.)
4. **Fallback Logic**: Automatic failover to alternative providers

### Credit Management

**Credit Threshold:**
- Default warning threshold: 50 credits
- Warning displayed when below threshold
- Requests blocked when credits reach zero

**Low Credit Warning:**
```
Warning: Low credit balance (25 credits remaining)
```

**Insufficient Credits Error:**
```
Error: provider selection failed: insufficient credits
```

### Provider Preferences

Configure your preferred provider in the config file:

```yaml
preferred_provider: anthropic
fallback_enabled: true
credits: 100
tier: pro
```

**Available Providers:**
- `anthropic` - Anthropic Claude (200k token limit)
- `openai` - OpenAI GPT (128k token limit)
- `google` - Google Gemini (1M token limit)

### Fallback Behavior

When `fallback_enabled: true`:
1. Try preferred provider first
2. If unavailable, try next capable provider
3. Continue until successful or all exhausted

## Configuration

### Configuration File

The CLI stores configuration in: `$HOME/.ainative-code.yaml`

**Example Configuration:**
```yaml
backend_url: http://localhost:8000
access_token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
refresh_token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
user_email: developer@example.com
user_id: user_123
preferred_provider: anthropic
credits: 150
tier: pro
fallback_enabled: true
```

### Environment Variables

Override defaults with environment variables:

- `AINATIVE_BACKEND_URL`: Backend API URL
- `AINATIVE_ACCESS_TOKEN`: Access token (for testing)
- `AINATIVE_REFRESH_TOKEN`: Refresh token (for testing)

**Example:**
```bash
export AINATIVE_BACKEND_URL=https://api.ainative.studio
ainative-code chat-ainative -m "Hello"
```

### View Configuration

Show current configuration:

```bash
ainative-code config show
```

**Note:** Sensitive values are masked by default. Use `--show-secrets` to display full values.

### Set Configuration Values

Update configuration values:

```bash
ainative-code config set preferred_provider anthropic
ainative-code config set fallback_enabled true
```

## Troubleshooting

### Authentication Issues

**Problem: "not authenticated"**
```
Error: not authenticated. Please run 'ainative-code auth login' first
```

**Solution:**
```bash
ainative-code auth login-backend --email your@email.com --password yourpassword
```

**Problem: "unauthorized"**
```
Error: login failed: unauthorized
```

**Solutions:**
1. Verify email and password are correct
2. Check backend URL in configuration
3. Ensure backend server is running

**Problem: Token expired**
```
Error: chat request failed: unauthorized
```

**Solution:**
```bash
ainative-code auth refresh-backend
```

### Chat Issues

**Problem: Empty message error**
```
Error: Error: message cannot be empty
```

**Solution:** Provide a non-empty message:
```bash
ainative-code chat-ainative -m "Your message here"
```

**Problem: Insufficient credits**
```
Error: chat request failed: payment required
```

**Solutions:**
1. Add credits to your account
2. Check credit balance in configuration
3. Contact support to top up credits

**Problem: Server error**
```
Error: chat request failed: server error
```

**Solutions:**
1. Check backend server is running
2. Verify backend URL is correct
3. Check server logs for details
4. Retry the request

### Provider Selection Issues

**Problem: No provider available**
```
Error: provider selection failed: no provider available
```

**Solutions:**
1. Verify at least one provider is configured
2. Check provider capabilities match requirements
3. Enable fallback: `ainative-code config set fallback_enabled true`

**Problem: Provider doesn't meet requirements**
```
Error: no provider meets requirements: no provider available
```

**Solutions:**
1. Remove specific capability requirements
2. Use a different provider manually
3. Update provider capabilities in configuration

### Configuration Issues

**Problem: Config file not found**
```
Warning: Could not save config: config file not found
```

**Solution:** Initialize configuration:
```bash
ainative-code config init
```

**Problem: Invalid backend URL**
```
Error: request failed: ...connection refused
```

**Solutions:**
1. Verify backend URL: `ainative-code config get backend_url`
2. Update backend URL: `ainative-code config set backend_url http://localhost:8000`
3. Ensure backend server is running

## Common Workflows

### First-Time Setup

```bash
# 1. Initialize configuration
ainative-code config init

# 2. Login to AINative
ainative-code auth login-backend \
  --email your@email.com \
  --password yourpassword

# 3. Set preferred provider
ainative-code config set preferred_provider anthropic

# 4. Send first chat message
ainative-code chat-ainative \
  --message "Hello, AI!" \
  --auto-provider
```

### Daily Usage

```bash
# Send chat messages with auto provider selection
ainative-code chat-ainative \
  -m "Your question here" \
  --auto-provider \
  --verbose
```

### Switching Providers

```bash
# Update preferred provider
ainative-code config set preferred_provider openai

# Send message with new provider
ainative-code chat-ainative -m "Hello" --auto-provider
```

### Token Management

```bash
# Check if token is still valid
ainative-code auth whoami

# Refresh token if expired
ainative-code auth refresh-backend

# Re-login if refresh fails
ainative-code auth login-backend -e your@email.com -p yourpassword
```

## Advanced Usage

### Scripting

Use in shell scripts with error handling:

```bash
#!/bin/bash

# Chat with error handling
if ! ainative-code chat-ainative -m "Explain TDD" --auto-provider > output.txt; then
  echo "Chat failed, attempting token refresh..."
  if ainative-code auth refresh-backend; then
    echo "Token refreshed, retrying..."
    ainative-code chat-ainative -m "Explain TDD" --auto-provider > output.txt
  else
    echo "Refresh failed, please login again"
    exit 1
  fi
fi

cat output.txt
```

### Batch Processing

Process multiple messages:

```bash
#!/bin/bash

messages=(
  "What is TDD?"
  "Explain goroutines"
  "Best practices for Go"
)

for msg in "${messages[@]}"; do
  echo "Question: $msg"
  ainative-code chat-ainative -m "$msg" --auto-provider
  echo "---"
done
```

### CI/CD Integration

Use in continuous integration:

```bash
# .github/workflows/ai-review.yml
- name: Login to AINative
  run: |
    ainative-code auth login-backend \
      --email ${{ secrets.AINATIVE_EMAIL }} \
      --password ${{ secrets.AINATIVE_PASSWORD }}

- name: Get AI Code Review
  run: |
    git diff main...HEAD > changes.diff
    ainative-code chat-ainative \
      -m "Review these code changes: $(cat changes.diff)" \
      --auto-provider > review.md

- name: Post Review
  run: cat review.md >> $GITHUB_STEP_SUMMARY
```

## Best Practices

1. **Security:**
   - Never commit credentials to version control
   - Use environment variables for automation
   - Rotate tokens regularly
   - Use `config show` instead of `--show-secrets` in shared environments

2. **Cost Management:**
   - Monitor credit balance regularly
   - Use `--verbose` to track token usage
   - Set appropriate credit thresholds
   - Enable fallback to cheaper providers when appropriate

3. **Error Handling:**
   - Always check command exit codes in scripts
   - Implement token refresh logic
   - Log errors for debugging
   - Use meaningful error messages

4. **Performance:**
   - Refresh tokens proactively before long operations
   - Use specific models when requirements are known
   - Cache responses when appropriate
   - Monitor API latency with verbose mode

## Support

For additional help:

1. Check command help: `ainative-code auth login-backend --help`
2. View configuration: `ainative-code config show`
3. Check logs: `ainative-code --verbose`
4. Review documentation: `/Users/aideveloper/AINative-Code/docs/`
5. Report issues: GitHub Issues

## Summary

The new AINative CLI commands provide:
- Seamless backend authentication
- Intelligent provider selection
- Credit management and warnings
- Comprehensive error handling
- Flexible configuration options

Get started now:
```bash
ainative-code auth login-backend --email your@email.com --password yourpassword
ainative-code chat-ainative -m "Hello, AI!" --auto-provider
```

Happy coding with AINative!
