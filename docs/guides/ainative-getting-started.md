# Getting Started with AINative Cloud

## Overview

AINative Code now supports cloud-based authentication and hosted inference through the AINative platform. This guide will help you set up and start using AINative's hosted LLM services.

## Benefits

- **No API Keys to Manage**: Single authentication for all providers
- **Unified Billing**: Pay-as-you-go credit system across all providers
- **Automatic Provider Selection**: Intelligent routing based on capabilities and availability
- **Credit-Based Usage Tracking**: Real-time visibility into consumption
- **Enterprise-Grade Security**: Secure token storage in OS keychain

## Prerequisites

- **AINative Code CLI installed**: See [Installation Guide](../user-guide/installation.md)
- **AINative account**: Sign up at https://ainative.studio
- **Python Backend running**: The backend service must be accessible

## Quick Start

### 1. Start the Python Backend

The AINative integration requires the Python backend service to be running:

```bash
cd /Users/aideveloper/AINative-Code/python-backend
uvicorn app.main:app --reload
```

The backend will start on `http://localhost:8000` by default.

### 2. Login to AINative

Authenticate with your AINative account credentials:

```bash
ainative-code auth login-backend \
  --email your-email@example.com \
  --password your-password
```

**Expected Response:**
```
Successfully logged in as your-email@example.com
```

Your tokens are automatically saved to the configuration file and will be used for subsequent requests.

### 3. Send Your First Chat Message

Now you can send chat messages using AINative's hosted inference:

```bash
ainative-code chat-ainative \
  --message "Hello! Tell me about AINative" \
  --auto-provider
```

The `--auto-provider` flag enables intelligent provider selection based on:
- Your configured preferences
- Available credits
- Model capabilities
- Provider availability

### 4. Check Your Authentication Status

Verify your login and check configuration:

```bash
ainative-code auth whoami
```

This displays your current authentication status and user information.

## Configuration

### Backend URL

By default, the CLI expects the backend at `http://localhost:8000`. You can customize this in your config file:

**Location:** `~/.config/ainative-code/config.yaml`

```yaml
backend_url: "http://localhost:8000"  # or your custom URL
```

### Provider Preferences

Set your preferred LLM provider:

```yaml
ainative:
  preferred_provider: anthropic  # Options: anthropic, openai, google
  fallback_enabled: true
```

## What's Next?

Now that you're set up, explore these guides:

- **[Authentication Guide](authentication.md)** - Learn about token management and refresh flows
- **[Hosted Inference Guide](hosted-inference.md)** - Explore chat features and model capabilities
- **[Provider Configuration Guide](provider-configuration.md)** - Configure provider preferences and selection logic
- **[Troubleshooting Guide](troubleshooting.md)** - Solve common issues

## Common Use Cases

### One-Shot Questions

Ask a quick question without entering interactive mode:

```bash
ainative-code chat-ainative --message "Explain async/await in Python"
```

### Specific Model Selection

Use a specific model for your request:

```bash
ainative-code chat-ainative \
  --message "Write a REST API in Go" \
  --model claude-sonnet-4-5
```

### Verbose Output

See detailed information about the request and response:

```bash
ainative-code chat-ainative \
  --message "Hello" \
  --verbose
```

This shows:
- Selected provider
- Model used
- Response time
- Token usage
- Credits consumed (when available)

## Support

If you encounter issues:

1. Check the [Troubleshooting Guide](troubleshooting.md)
2. Ensure the Python backend is running
3. Verify your authentication with `ainative-code auth whoami`
4. Review logs in verbose mode with `--verbose`
5. Open an issue at https://github.com/AINative-Studio/ainative-code/issues
