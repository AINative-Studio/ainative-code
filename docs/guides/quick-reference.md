# AINative Cloud Quick Reference

## Quick Start (3 Steps)

```bash
# 1. Start backend
cd python-backend && uvicorn app.main:app --reload

# 2. Login
ainative-code auth login-backend -e your@email.com -p password

# 3. Chat
ainative-code chat-ainative -m "Hello!" --auto-provider
```

---

## Authentication Commands

| Command | Description | Example |
|---------|-------------|---------|
| `auth login-backend` | Login to AINative | `ainative-code auth login-backend -e user@email.com -p pass` |
| `auth logout-backend` | Logout and clear tokens | `ainative-code auth logout-backend` |
| `auth refresh-backend` | Refresh access token | `ainative-code auth refresh-backend` |
| `auth whoami` | Check auth status | `ainative-code auth whoami` |

---

## Chat Commands

| Command | Description | Example |
|---------|-------------|---------|
| `chat-ainative -m "text"` | Send message | `ainative-code chat-ainative -m "Hello"` |
| `--auto-provider` | Auto provider selection | `ainative-code chat-ainative -m "Hi" --auto-provider` |
| `--model MODEL` | Specify model | `ainative-code chat-ainative -m "Hi" --model claude-sonnet-4-5` |
| `--provider PROVIDER` | Specify provider | `ainative-code chat-ainative -m "Hi" --provider anthropic` |
| `--stream` | Enable streaming | `ainative-code chat-ainative -m "Count to 10" --stream` |
| `--verbose` | Show details | `ainative-code chat-ainative -m "Hi" --verbose` |

---

## Common Workflows

### First Time Setup
```bash
# Start backend
cd python-backend
uvicorn app.main:app --reload

# Login (in another terminal)
ainative-code auth login-backend -e your@email.com -p password

# Test
ainative-code chat-ainative -m "Hello!" --auto-provider
```

### Daily Usage
```bash
# Check auth
ainative-code auth whoami

# If expired, refresh
ainative-code auth refresh-backend

# Chat
ainative-code chat-ainative -m "Your question" --auto-provider
```

### Provider Selection
```bash
# Let system choose
ainative-code chat-ainative -m "Question" --auto-provider

# Choose specific provider
ainative-code chat-ainative -m "Question" --provider anthropic

# Choose specific model
ainative-code chat-ainative -m "Question" --model claude-sonnet-4-5
```

---

## Configuration

### File Location
`~/.config/ainative-code/config.yaml`

### Basic Config
```yaml
backend_url: "http://localhost:8000"
ainative:
  preferred_provider: anthropic
  fallback_enabled: true
```

### Set Config via CLI
```bash
ainative-code config set ainative.preferred_provider anthropic
ainative-code config get ainative.preferred_provider
```

---

## Supported Providers & Models

| Provider | Models | Context |
|----------|--------|---------|
| **Anthropic** | claude-sonnet-4-5, claude-opus-4 | 200K |
| **OpenAI** | gpt-4, gpt-4-turbo, gpt-3.5-turbo | 128K |
| **Google** | gemini-pro, gemini-ultra | 1M |

---

## Common Errors & Solutions

| Error | Solution |
|-------|----------|
| "not authenticated" | `ainative-code auth login-backend -e email -p pass` |
| "unauthorized" (401) | `ainative-code auth refresh-backend` |
| "connection refused" | Start backend: `cd python-backend && uvicorn app.main:app --reload` |
| "payment required" (402) | Add credits at https://ainative.studio/billing |
| "no provider available" | Use `--auto-provider` or set preferred provider |

---

## Environment Variables

```bash
export AINATIVE_BACKEND_URL="http://localhost:8000"
export AINATIVE_PREFERRED_PROVIDER="anthropic"
export AINATIVE_ACCESS_TOKEN="your-token"
export AINATIVE_REFRESH_TOKEN="your-refresh-token"
```

---

## Troubleshooting Commands

```bash
# Check auth status
ainative-code auth whoami

# Test with verbose
ainative-code chat-ainative -m "test" --verbose

# Check backend
curl http://localhost:8000/health

# Check config
cat ~/.config/ainative-code/config.yaml
```

---

## Best Practices

1. **Always use auto-provider** for flexibility
2. **Enable fallback** in production
3. **Use environment variables** for tokens in CI/CD
4. **Check auth** before long sessions
5. **Enable verbose** when debugging

---

## Getting Help

- **Docs**: [docs/guides/](.)
- **Issues**: https://github.com/AINative-Studio/ainative-code/issues
- **Email**: support@ainative.studio

---

## Full Documentation

- [Getting Started Guide](ainative-getting-started.md)
- [Authentication Guide](authentication.md)
- [Hosted Inference Guide](hosted-inference.md)
- [Provider Configuration Guide](provider-configuration.md)
- [Troubleshooting Guide](troubleshooting.md)
- [API Reference](../api/ainative-provider.md)
- [Migration Guide](../migration/adding-ainative-auth.md)
