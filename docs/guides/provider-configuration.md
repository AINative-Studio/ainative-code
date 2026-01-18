# Provider Configuration Guide

## Overview

AINative's provider configuration system allows you to customize how the platform selects and routes requests to different LLM providers (Anthropic, OpenAI, Google, etc.).

## Configuration File

### Location

**Config File:** `~/.config/ainative-code/config.yaml`

### Basic Configuration

```yaml
# Backend Configuration
backend_url: "http://localhost:8000"

# AINative Platform Settings
ainative:
  preferred_provider: anthropic  # Your preferred LLM provider
  fallback_enabled: true         # Enable automatic fallback
```

### Full Configuration Example

```yaml
# Backend API Configuration
backend_url: "http://localhost:8000"

# Authentication
access_token: "eyJhbGc..."
refresh_token: "eyJhbGc..."
user_email: "user@example.com"
user_id: "user123"

# AINative Configuration
ainative:
  # Provider Selection
  preferred_provider: anthropic
  fallback_enabled: true

  # Provider Priority List
  provider_priority:
    - anthropic
    - openai
    - google

  # Capability Requirements (future feature)
  require_streaming: false
  require_vision: false
  require_function_calling: false
```

## Provider Preferences

### Setting Preferred Provider

Your preferred provider is the default choice when using `--auto-provider`:

**Options:**
- `anthropic` - Claude models
- `openai` - GPT models
- `google` - Gemini models

**In config file:**
```yaml
ainative:
  preferred_provider: anthropic
```

**Via environment variable:**
```bash
export AINATIVE_PREFERRED_PROVIDER=anthropic
```

### Provider Priority

Set the order of providers to try when fallback is enabled:

```yaml
ainative:
  provider_priority:
    - anthropic  # Try Anthropic first
    - openai     # Fall back to OpenAI
    - google     # Finally try Google
```

## Provider Selection Logic

### Automatic Provider Selection

When using `--auto-provider`, the system follows this logic:

```
1. Check User Preference
   └─ If preferred_provider is set and available → Use it

2. Check Capability Requirements
   └─ Filter providers by required capabilities

3. Check Provider Health
   └─ Remove unavailable/unhealthy providers

4. Apply Priority Order
   └─ Select highest priority provider from remaining

5. Fallback Logic
   └─ If primary fails, try next in priority list
```

### Selection Examples

**Example 1: Simple Preference**
```yaml
ainative:
  preferred_provider: anthropic
```
Result: Always uses Anthropic when available

**Example 2: With Fallback**
```yaml
ainative:
  preferred_provider: anthropic
  fallback_enabled: true
  provider_priority:
    - anthropic
    - openai
```
Result: Uses Anthropic, falls back to OpenAI if unavailable

**Example 3: Capability-Based (Future)**
```yaml
ainative:
  preferred_provider: anthropic
  require_vision: true
```
Result: Only uses providers with vision support

## Manual Provider Override

### Per-Request Override

Override provider selection for a specific request:

```bash
ainative-code chat-ainative \
  --message "Hello" \
  --provider openai
```

This ignores your configured preference and uses the specified provider.

### Model-Specific Selection

Specify both provider and model:

```bash
ainative-code chat-ainative \
  --message "Hello" \
  --provider anthropic \
  --model claude-sonnet-4-5
```

## Provider Capabilities

### Understanding Capabilities

Different providers support different features:

| Capability | Anthropic | OpenAI | Google |
|------------|-----------|--------|--------|
| **Streaming** | Yes | Yes | Yes |
| **Vision** | Yes | Yes | Yes |
| **Function Calling** | Yes | Yes | Yes |
| **Large Context** | 200K | 128K | 1M |
| **JSON Mode** | Yes | Yes | Yes |

### Capability-Based Selection (Future Feature)

Request specific capabilities and let the system choose:

```bash
ainative-code chat-ainative \
  --message "Analyze this image" \
  --require-vision \
  --auto-provider
```

Configuration:
```yaml
ainative:
  require_vision: true
  # Only vision-capable providers will be selected
```

## Fallback Behavior

### Enabling Fallback

```yaml
ainative:
  fallback_enabled: true
  provider_priority:
    - anthropic
    - openai
    - google
```

### Fallback Flow

```
Request with --auto-provider
  ↓
Try Anthropic (preferred)
  ↓
[Anthropic Error/Unavailable]
  ↓
Try OpenAI (next priority)
  ↓
[OpenAI Success]
  ↓
Return Response
```

### Fallback Scenarios

Fallback occurs when:
- Provider returns 503 (Service Unavailable)
- Provider returns 429 (Rate Limited)
- Connection timeout
- Provider explicitly unavailable

Fallback does NOT occur when:
- Invalid authentication (401)
- Insufficient credits (402)
- Invalid request (400)

## Advanced Configuration

### Custom Backend URL

For production or custom deployments:

```yaml
backend_url: "https://api.ainative.studio"
```

Or via environment:
```bash
export AINATIVE_BACKEND_URL="https://api.ainative.studio"
```

### Timeout Configuration

Set request timeout (default: 120 seconds):

```yaml
ainative:
  timeout: 60  # seconds
```

### Retry Configuration

Configure retry behavior:

```yaml
ainative:
  max_retries: 3
  retry_delay: 1  # seconds
  exponential_backoff: true
```

## Provider-Specific Settings

### Model Defaults

Set default models per provider:

```yaml
providers:
  anthropic:
    default_model: claude-sonnet-4-5
  openai:
    default_model: gpt-4-turbo
  google:
    default_model: gemini-pro
```

### Provider Parameters

Configure provider-specific parameters:

```yaml
providers:
  anthropic:
    temperature: 0.7
    max_tokens: 4096
  openai:
    temperature: 0.8
    max_tokens: 2048
```

## Environment Variables

All configuration options can be set via environment variables:

### Provider Selection
```bash
export AINATIVE_PREFERRED_PROVIDER=anthropic
export AINATIVE_FALLBACK_ENABLED=true
```

### Backend Configuration
```bash
export AINATIVE_BACKEND_URL="http://localhost:8000"
export AINATIVE_TIMEOUT=60
```

### Authentication
```bash
export AINATIVE_ACCESS_TOKEN="eyJhbGc..."
export AINATIVE_REFRESH_TOKEN="eyJhbGc..."
```

### Priority Order

Configuration sources (highest to lowest):
1. Command-line flags (`--provider anthropic`)
2. Environment variables (`AINATIVE_PREFERRED_PROVIDER`)
3. Config file (`~/.config/ainative-code/config.yaml`)
4. Default values

## Best Practices

### 1. Set a Preferred Provider

Always configure a preferred provider:

```yaml
ainative:
  preferred_provider: anthropic  # Your preferred choice
```

### 2. Enable Fallback for Production

Ensure high availability with fallback:

```yaml
ainative:
  fallback_enabled: true
  provider_priority:
    - anthropic
    - openai
```

### 3. Use Environment Variables for Secrets

Keep tokens in environment variables, not config files:

```bash
export AINATIVE_ACCESS_TOKEN="..."
export AINATIVE_REFRESH_TOKEN="..."
```

### 4. Set Reasonable Timeouts

Balance responsiveness and reliability:

```yaml
ainative:
  timeout: 120  # 2 minutes for complex requests
```

### 5. Monitor Provider Performance

Use `--verbose` to see which provider is selected:

```bash
ainative-code chat-ainative -m "test" --auto-provider --verbose
```

## Troubleshooting

### Provider Not Being Selected

**Issue:** Preferred provider is not being used

**Check:**
1. Is the provider name spelled correctly?
2. Is `--auto-provider` flag enabled?
3. Is the backend running and accessible?

**Debug:**
```bash
ainative-code chat-ainative -m "test" --auto-provider --verbose
```

### Fallback Not Working

**Issue:** Fallback is not occurring when expected

**Check:**
1. Is `fallback_enabled: true` in config?
2. Is `provider_priority` list defined?
3. Are all providers in the list available?

**Solution:**
```yaml
ainative:
  fallback_enabled: true
  provider_priority:
    - anthropic
    - openai
    - google
```

### Configuration Not Loading

**Issue:** Config changes are not taking effect

**Check:**
1. Config file location: `~/.config/ainative-code/config.yaml`
2. YAML syntax is correct
3. No conflicting environment variables

**Verify:**
```bash
ainative-code config get ainative.preferred_provider
```

## Examples

### Example 1: Anthropic-First Strategy

```yaml
ainative:
  preferred_provider: anthropic
  fallback_enabled: true
  provider_priority:
    - anthropic
    - openai
```

**Usage:**
```bash
ainative-code chat-ainative -m "Hello" --auto-provider
```

### Example 2: Cost-Optimized Strategy

```yaml
ainative:
  preferred_provider: google  # Gemini is cost-effective
  fallback_enabled: true
  provider_priority:
    - google
    - openai
    - anthropic
```

### Example 3: OpenAI-Only Strategy

```yaml
ainative:
  preferred_provider: openai
  fallback_enabled: false  # No fallback, OpenAI only
```

### Example 4: Development Setup

```yaml
backend_url: "http://localhost:8000"
ainative:
  preferred_provider: anthropic
  fallback_enabled: false  # Fail fast for debugging
  timeout: 30
```

### Example 5: Production Setup

```yaml
backend_url: "https://api.ainative.studio"
ainative:
  preferred_provider: anthropic
  fallback_enabled: true
  provider_priority:
    - anthropic
    - openai
    - google
  timeout: 120
  max_retries: 3
```

## Next Steps

- [Hosted Inference Guide](hosted-inference.md) - Learn about using different models
- [Authentication Guide](authentication.md) - Manage credentials
- [Troubleshooting Guide](troubleshooting.md) - Solve common issues
- [API Reference](../api/ainative-provider.md) - Detailed command documentation
