# Hosted Inference with AINative

## What is Hosted Inference?

AINative provides hosted access to multiple LLM providers through a unified API. Instead of managing API keys for each provider, you authenticate once with AINative and access all providers through a single interface.

## Supported Providers and Models

### Provider Overview

| Provider | Models | Max Context | Capabilities |
|----------|--------|-------------|--------------|
| **Anthropic** | Claude Sonnet 4.5, Claude Opus 4 | 200K tokens | Vision, Function Calling, Streaming |
| **OpenAI** | GPT-4, GPT-4 Turbo, GPT-3.5 Turbo | 128K tokens | Vision, Function Calling, Streaming |
| **Google** | Gemini Pro, Gemini Ultra | 1M tokens | Vision, Function Calling, Streaming |

### Model Details

**Anthropic Claude:**
- `claude-sonnet-4-5` - Balanced performance and speed
- `claude-opus-4` - Highest capability model

**OpenAI GPT:**
- `gpt-4` - Advanced reasoning and analysis
- `gpt-4-turbo` - Faster GPT-4 with vision
- `gpt-3.5-turbo` - Fast, cost-effective model

**Google Gemini:**
- `gemini-pro` - Versatile model for various tasks
- `gemini-ultra` - Most capable Gemini model

## Basic Usage

### Send a Chat Message

Send a single message and receive a response:

```bash
ainative-code chat-ainative \
  --message "Explain quantum computing in simple terms"
```

**Short form:**
```bash
ainative-code chat-ainative -m "Hello, world!"
```

### Specify a Model

Use a specific model for your request:

```bash
ainative-code chat-ainative \
  --message "Write a Python function to sort a list" \
  --model claude-sonnet-4-5
```

### Auto Provider Selection

Let AINative automatically select the best provider:

```bash
ainative-code chat-ainative \
  --message "What is machine learning?" \
  --auto-provider
```

The auto provider selection considers:
- Your configured provider preference
- Available credits (future feature)
- Model capabilities required
- Current provider availability
- Load balancing across providers

## Advanced Features

### Streaming Responses

Enable real-time streaming for long responses:

```bash
ainative-code chat-ainative \
  --message "Write a detailed explanation of REST APIs" \
  --stream
```

Benefits of streaming:
- See responses as they're generated
- Better user experience for long outputs
- Ability to interrupt if needed
- Same credit consumption as non-streaming

### Verbose Mode

Get detailed information about the request and response:

```bash
ainative-code chat-ainative \
  --message "Hello" \
  --verbose
```

**Verbose output includes:**
- Selected provider and model
- Request parameters
- Response time
- Token usage (prompt and completion)
- Credits consumed (when available)
- Full request/response headers

### Custom System Messages

Provide a system message to set context or behavior:

```bash
ainative-code chat-ainative \
  --message "Explain variables" \
  --system "You are a programming tutor for beginners. Use simple language and provide examples."
```

### Multi-Turn Conversations

For conversation context, you'll need to use session management (future feature) or pass conversation history programmatically.

## Provider Selection Logic

### Manual Provider Selection

Explicitly choose a provider:

```bash
ainative-code chat-ainative \
  --message "Hello" \
  --provider anthropic
```

### Automatic Provider Selection

When using `--auto-provider`, the selection logic follows this order:

1. **User Preference** - Your configured preferred provider (if available)
2. **Capability Matching** - Providers that support required features
3. **Availability** - Currently healthy and responsive providers
4. **Load Balancing** - Distribute load across providers
5. **Fallback** - Alternative provider if primary fails

### Configuration

Set your preferred provider in `~/.config/ainative-code/config.yaml`:

```yaml
ainative:
  preferred_provider: anthropic  # Options: anthropic, openai, google
  fallback_enabled: true
```

## Credit Management

### How Credits Work

AINative uses a credit-based system for billing:
- Different models consume different amounts of credits
- Credits are deducted based on token usage
- Both input and output tokens consume credits
- Streaming uses the same credits as non-streaming

### Check Credit Balance

View your current credit balance:

```bash
ainative-code auth whoami
```

**Note:** Credit display in `whoami` is a future feature. Currently, this command shows authentication status only.

### Credit Consumption

Approximate credit consumption:

| Model | Input (per 1K tokens) | Output (per 1K tokens) |
|-------|----------------------|------------------------|
| Claude Sonnet 4.5 | 3 credits | 15 credits |
| GPT-4 Turbo | 10 credits | 30 credits |
| GPT-3.5 Turbo | 1 credit | 2 credits |
| Gemini Pro | 2 credits | 6 credits |

**Note:** These are example rates. Check https://ainative.studio/pricing for current rates.

### Low Credit Warnings

When your credit balance is low:

```
Warning: Low credit balance (45 credits remaining)
Visit https://ainative.studio/billing to add credits
```

Actions:
1. Visit https://ainative.studio/billing
2. Purchase additional credits
3. Monitor usage with `--verbose` flag

## Use Cases and Examples

### Code Generation

```bash
ainative-code chat-ainative \
  --message "Write a REST API endpoint in Go for user authentication" \
  --model claude-sonnet-4-5 \
  --verbose
```

### Code Explanation

```bash
ainative-code chat-ainative \
  --message "Explain this code: func main() { fmt.Println(\"Hello\") }"
```

### Debugging Help

```bash
ainative-code chat-ainative \
  --message "Why am I getting a null pointer exception in this code?" \
  --system "You are an expert debugger. Analyze code and explain errors clearly."
```

### Documentation Writing

```bash
ainative-code chat-ainative \
  --message "Write API documentation for a user login endpoint" \
  --model gpt-4
```

### Learning and Education

```bash
ainative-code chat-ainative \
  --message "Teach me about Docker containers with examples" \
  --auto-provider
```

## Best Practices

### 1. Use Auto Provider for Flexibility

Enable `--auto-provider` to automatically adapt to provider availability:

```bash
ainative-code chat-ainative -m "Hello" --auto-provider
```

### 2. Enable Verbose Mode for Debugging

When troubleshooting or monitoring usage, use `--verbose`:

```bash
ainative-code chat-ainative -m "Test" --verbose
```

### 3. Choose Models Based on Task Complexity

- **Simple queries**: Use GPT-3.5 Turbo or Gemini Pro
- **Complex reasoning**: Use Claude Opus 4 or GPT-4
- **Code generation**: Use Claude Sonnet 4.5 or GPT-4 Turbo

### 4. Use Streaming for Long Responses

Enable streaming for better UX with lengthy outputs:

```bash
ainative-code chat-ainative \
  --message "Write a comprehensive guide..." \
  --stream
```

### 5. Set System Messages for Specialized Tasks

Provide context with system messages:

```bash
ainative-code chat-ainative \
  --message "Review this code" \
  --system "You are a senior software engineer conducting a code review"
```

## Limitations and Quotas

### Rate Limits

Current rate limits (subject to change):
- **Requests per minute**: 60
- **Tokens per minute**: 100,000
- **Concurrent requests**: 10

Rate limit errors will return a 429 status code.

### Context Windows

Respect maximum context windows:
- Anthropic Claude: 200K tokens
- OpenAI GPT-4: 128K tokens
- Google Gemini: 1M tokens

### Response Timeouts

Default timeout: 2 minutes

For longer operations, consider:
- Breaking down requests into smaller chunks
- Using streaming for incremental results
- Implementing retry logic with exponential backoff

## Troubleshooting

### Request Fails with "payment required" (402)

**Cause:** Insufficient credits

**Solution:**
1. Visit https://ainative.studio/billing
2. Purchase additional credits
3. Verify balance with `ainative-code auth whoami`

### Request Fails with "not authenticated"

**Cause:** No valid access token

**Solution:**
```bash
ainative-code auth login-backend -e your@email.com -p password
```

### Request Fails with "provider not available"

**Cause:** Specified provider is unavailable

**Solution:**
- Use `--auto-provider` for automatic fallback
- Try a different provider manually
- Check provider status at https://status.ainative.studio

### Slow Response Times

**Causes:**
- Large context window
- Complex model (GPT-4, Claude Opus)
- High provider load

**Solutions:**
- Use faster models (GPT-3.5, Claude Sonnet)
- Reduce context/prompt length
- Enable `--stream` for incremental results
- Try `--auto-provider` for load balancing

## Next Steps

- [Provider Configuration Guide](provider-configuration.md) - Configure provider preferences
- [Authentication Guide](authentication.md) - Manage tokens and credentials
- [Troubleshooting Guide](troubleshooting.md) - Solve common issues
- [API Reference](../api/ainative-provider.md) - Detailed command reference
