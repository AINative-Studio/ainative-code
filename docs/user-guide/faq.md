# Frequently Asked Questions (FAQ)

Common questions and answers about AINative Code.

## Table of Contents

1. [General Questions](#general-questions)
2. [Installation & Setup](#installation--setup)
3. [Usage & Features](#usage--features)
4. [LLM Providers](#llm-providers)
5. [Sessions & History](#sessions--history)
6. [Performance & Cost](#performance--cost)
7. [Security & Privacy](#security--privacy)
8. [Platform Integration](#platform-integration)
9. [Troubleshooting](#troubleshooting)
10. [Best Practices](#best-practices)

## General Questions

### What is AINative Code?

AINative Code is a terminal-based AI coding assistant that combines the best features of open-source AI CLI tools with native integration to the AINative platform ecosystem. It provides interactive AI chat, code generation, and integrations with services like ZeroDB, design tokens, and Strapi CMS.

### Is AINative Code free?

The AINative Code CLI tool itself is free and open-source. However, you'll need API keys from LLM providers (Anthropic, OpenAI, etc.), which have their own pricing. You can also use free local models via Ollama.

### What are the system requirements?

- **OS**: macOS 10.15+, Linux (Ubuntu 20.04+), Windows 10+
- **Memory**: 2 GB RAM minimum (4 GB recommended)
- **Disk**: 100 MB for installation
- **Network**: Internet connection for cloud LLM providers

### How does AINative Code differ from ChatGPT or GitHub Copilot?

- **Terminal-native**: Works directly in your terminal, not in a browser or IDE
- **Multi-provider**: Use Anthropic, OpenAI, Google, AWS, Azure, or local models
- **Platform integration**: Native access to ZeroDB, design tokens, Strapi CMS
- **Session management**: Save and resume conversations with full context
- **Privacy options**: Use local models via Ollama for complete privacy

### Can I use AINative Code offline?

Partially. You can use local models via Ollama for offline operation, but cloud-based providers (Anthropic, OpenAI, etc.) require internet connectivity.

## Installation & Setup

### How do I install AINative Code?

**macOS (recommended):**
```bash
brew install ainative-studio/tap/ainative-code
```

**Linux:**
```bash
curl -LO https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-linux-amd64
chmod +x ainative-code-linux-amd64
sudo mv ainative-code-linux-amd64 /usr/local/bin/ainative-code
```

See the [Installation Guide](installation.md) for all installation methods.

### Do I need to install anything else?

No additional dependencies are required. AINative Code is a standalone binary. However, you'll need API keys from at least one LLM provider.

### How do I get an API key?

- **Anthropic Claude**: [console.anthropic.com](https://console.anthropic.com/)
- **OpenAI**: [platform.openai.com/api-keys](https://platform.openai.com/api-keys)
- **Google Gemini**: [makersuite.google.com/app/apikey](https://makersuite.google.com/app/apikey)
- **Ollama**: No API key needed (local models)

### Where should I store my API keys?

Use environment variables (recommended):
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

Or in config file with environment variable reference:
```yaml
llm:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
```

Never commit API keys to version control!

### How do I update AINative Code?

**Homebrew:**
```bash
brew upgrade ainative-code
```

**Manual installation:**
Download the latest release and replace the binary.

## Usage & Features

### How do I start a chat session?

```bash
ainative-code chat
```

Or with a specific message:
```bash
ainative-code chat "Explain how to use Go interfaces"
```

### Can I use multiple LLM providers?

Yes! Configure multiple providers and switch between them:

```bash
# Use Anthropic
ainative-code chat --provider anthropic "Question"

# Use OpenAI
ainative-code chat --provider openai "Question"

# Set default in config
ainative-code config set llm.default_provider anthropic
```

### How do I save conversations?

Sessions are automatically saved. You can:

```bash
# List sessions
ainative-code session list

# Export a session
ainative-code session export <session-id> -o session.json

# Resume a session
ainative-code chat --session-id <session-id>
```

### Can I use AINative Code in scripts?

Yes! Use one-shot mode for scripting:

```bash
# Get response and process
response=$(ainative-code chat "Generate unit test for this function: [code]")
echo "$response" > test.go
```

### Does AINative Code support multiple languages?

Yes, the AI models support all major programming languages:
- Go, Python, JavaScript/TypeScript, Java, Rust, C++, Ruby, PHP, etc.

You can also ask questions in different natural languages (English, Spanish, French, etc.), though English typically provides best results.

### Can I execute code directly?

Not automatically for security. However, the AI can use the bash tool to run commands when appropriate and with your permission.

### How do I customize the AI's behavior?

Use system messages or custom prompts:

```bash
ainative-code chat --system "You are an expert Go developer. Always provide idiomatic Go code."
```

Or in configuration:
```yaml
llm:
  anthropic:
    system_message: "You are an expert Go developer..."
```

## LLM Providers

### Which LLM provider is best for coding?

For coding tasks, we recommend:

1. **Anthropic Claude 3.5 Sonnet** - Best balance of quality and cost
2. **OpenAI GPT-4 Turbo** - Excellent general-purpose
3. **Claude 3 Opus** - Best for complex tasks (more expensive)
4. **Claude 3 Haiku** - Best for quick questions (cheapest)
5. **Ollama CodeLlama** - Best for privacy/offline (free)

### Can I use multiple providers simultaneously?

Yes, configure fallback providers:

```yaml
llm:
  fallback:
    enabled: true
    providers:
      - anthropic
      - openai
      - google
```

The system will automatically try the next provider if one fails.

### How much do LLM providers cost?

Approximate costs (per 1M tokens):

- **Claude 3.5 Sonnet**: $3 input / $15 output
- **Claude 3 Opus**: $15 input / $75 output
- **Claude 3 Haiku**: $0.25 input / $1.25 output
- **GPT-4 Turbo**: $10 input / $30 output
- **GPT-3.5 Turbo**: $0.50 input / $1.50 output
- **Ollama**: Free (local)

### Can I use local models only?

Yes! Use Ollama for completely local operation:

```bash
# Install Ollama
brew install ollama

# Pull a model
ollama pull codellama

# Configure AINative Code
ainative-code config set llm.default_provider ollama
ainative-code config set llm.ollama.model codellama

# Use it
ainative-code chat "Explain Go channels"
```

### Which model should I use for different tasks?

- **Complex reasoning**: Claude 3 Opus, GPT-4
- **Code generation**: Claude 3.5 Sonnet, GPT-4 Turbo
- **Code review**: Claude 3 Opus, GPT-4
- **Quick questions**: Claude 3 Haiku, GPT-3.5 Turbo
- **Privacy/offline**: Ollama CodeLlama, DeepSeek-Coder

## Sessions & History

### Where are my sessions stored?

Sessions are stored in:
- Local SQLite database: `~/.config/ainative-code/sessions.db`
- ZeroDB (if enabled): For cloud backup and sync

### How long are sessions kept?

Sessions are kept indefinitely by default. You can:

```bash
# Delete old sessions
ainative-code session delete --older-than 90d

# Delete specific session
ainative-code session delete <session-id>
```

### Can I share sessions with team members?

Yes! Export and share:

```bash
# Export session
ainative-code session export <session-id> -o session.json

# Share file with team
# They can import:
ainative-code session import -i session.json
```

### Do sessions sync across devices?

If ZeroDB is enabled and you're logged in, sessions automatically sync across devices.

### How do I search through old sessions?

```bash
# Search by title or content
ainative-code session search "OAuth implementation"

# Filter by date
ainative-code session list --since 2024-01-01

# Filter by tag
ainative-code session list --tag golang
```

## Performance & Cost

### How can I reduce costs?

1. **Use cheaper models**:
   ```bash
   ainative-code chat --model claude-3-haiku "simple question"
   ```

2. **Reduce max_tokens**:
   ```yaml
   llm:
     anthropic:
       max_tokens: 2048  # Lower for shorter responses
   ```

3. **Use local models** for testing:
   ```bash
   ainative-code config set llm.default_provider ollama
   ```

4. **Enable caching**:
   ```yaml
   performance:
     cache:
       enabled: true
   ```

### How do I track my usage and costs?

```bash
# View usage statistics
ainative-code analytics usage

# View costs by provider
ainative-code analytics providers

# Monthly cost report
ainative-code analytics cost --month current

# Export for analysis
ainative-code analytics cost --export costs.csv
```

### Why are responses slow?

Possible causes and solutions:

1. **Model is complex**: Use faster models (Haiku, GPT-3.5)
2. **Network issues**: Check connection, try different provider
3. **Large responses**: Reduce max_tokens
4. **Not streaming**: Enable streaming for progressive output

```bash
# Use faster model
ainative-code chat --model claude-3-haiku

# Enable streaming
ainative-code chat --stream
```

### Can I cache responses?

Yes, enable caching:

```yaml
performance:
  cache:
    enabled: true
    type: memory  # or: redis
    ttl: 1h
    max_size: 100  # MB
```

## Security & Privacy

### Are my conversations private?

- **Local storage**: Sessions stored locally on your device
- **API calls**: Sent to LLM providers (Anthropic, OpenAI, etc.)
- **ZeroDB**: Optional cloud backup (encrypted)
- **Local models**: Complete privacy with Ollama

### Does AINative send my data anywhere?

Only to:
1. **LLM providers** (Anthropic, OpenAI, etc.) - required for AI features
2. **AINative platform** (optional) - only if you use ZeroDB, Strapi, etc.
3. **MCP servers** - only servers you explicitly configure

With Ollama, everything stays local.

### How are API keys stored?

- **macOS**: Securely in Keychain
- **Linux**: GNOME Keyring or KWallet
- **Windows**: Credential Manager
- **Fallback**: Encrypted config file

### Can I use AINative Code in a corporate environment?

Yes! Enterprise features:

1. **Self-hosted LLMs**: Use Azure OpenAI, AWS Bedrock, or Ollama
2. **Private deployment**: All data stays in your infrastructure
3. **Audit logging**: Track all AI interactions
4. **Network isolation**: No external API calls required (with Ollama)

### Should I review AI-generated code?

**Always!** AI-generated code should be reviewed:
- Check for security vulnerabilities
- Verify correctness and logic
- Ensure it follows your coding standards
- Test thoroughly before deployment

### Can I prevent certain commands from running?

Yes, configure tool restrictions:

```yaml
tools:
  terminal:
    blocked_commands:
      - rm -rf
      - sudo
      - dd
    require_confirmation: true
```

## Platform Integration

### Do I need to use AINative platform features?

No, they're optional. You can use AINative Code with just LLM providers (Anthropic, OpenAI, etc.) without the AINative platform.

### What is ZeroDB?

ZeroDB is a vector database and NoSQL database optimized for AI applications. It provides:
- Vector embeddings and semantic search
- NoSQL tables for flexible data
- PostgreSQL compatibility
- Agent memory persistence

### How do I enable ZeroDB?

```bash
# Login to AINative platform
ainative-code auth login

# Enable in config
ainative-code config set services.zerodb.enabled true
```

### What are design tokens?

Design tokens are design system values (colors, spacing, typography) that can be:
- Extracted from Figma, Sketch, Adobe XD
- Generated as code (CSS, SCSS, JavaScript, Tailwind)
- Synced between design and development

### Do I need Strapi CMS?

No, it's optional. Only needed if you want to:
- Create blog posts via AI
- Manage content through the CLI
- Integrate with your Strapi instance

## Troubleshooting

### Command not found after installation

```bash
# Check PATH
echo $PATH

# Add to PATH
export PATH="$PATH:/usr/local/bin"

# Reload shell
source ~/.bashrc  # or ~/.zshrc
```

See [Troubleshooting Guide](troubleshooting.md) for more.

### "Invalid API key" error

```bash
# Verify API key is set
echo $ANTHROPIC_API_KEY

# Remove any whitespace
export ANTHROPIC_API_KEY=$(echo $ANTHROPIC_API_KEY | tr -d '[:space:]')

# Test directly
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{"model":"claude-3-haiku-20240307","max_tokens":10,"messages":[{"role":"user","content":"Hi"}]}'
```

### Rate limit exceeded

```bash
# Wait before retrying
sleep 60

# Use different provider
ainative-code chat --provider openai

# Enable fallback
llm:
  fallback:
    enabled: true
    providers: [anthropic, openai]
```

### Session not found

```bash
# List all sessions
ainative-code session list --all

# Check session ID (case-sensitive)
# Sync from cloud if enabled
ainative-code session sync --pull
```

### Configuration not loading

```bash
# Check which config is used
ainative-code --verbose config show

# Re-initialize
mv ~/.config/ainative-code/config.yaml ~/.config/ainative-code/config.yaml.bak
ainative-code init
```

## Best Practices

### What are the best practices for prompting?

1. **Be specific**: "Implement JWT authentication in Go with RSA signing" vs "Add auth"
2. **Provide context**: "I'm using Gin framework" helps the AI understand your environment
3. **Ask for explanations**: "Explain why" helps you learn
4. **Iterate**: Refine responses by asking follow-up questions
5. **Request tests**: "Include unit tests" ensures code quality

### How should I organize sessions?

1. **One session per topic/feature**: Keeps context focused
2. **Use descriptive titles**: Makes sessions easy to find
3. **Tag sessions**: Organize by project, language, type
4. **Export important sessions**: Backup valuable conversations
5. **Clean up old sessions**: Delete or archive completed work

### Should I use streaming or wait for complete responses?

- **Streaming** (default): See responses as they're generated, faster perceived latency
- **Complete**: Better for scripting, processing entire response at once

```bash
# Streaming (default)
ainative-code chat --stream

# Complete response
ainative-code chat --stream=false
```

### How often should I back up sessions?

```bash
# Weekly backup (cron job)
0 0 * * 0 ainative-code session export --all -o ~/backups/sessions-$(date +\%Y\%m\%d).json
```

### What information should I include when asking for help?

1. **Version**: `ainative-code --version`
2. **Operating system**: macOS 14.1, Ubuntu 22.04, etc.
3. **Configuration**: `ainative-code config show` (redact secrets!)
4. **Error messages**: Full error output
5. **Steps to reproduce**: What commands you ran
6. **Logs**: Recent log entries (if applicable)

## Getting More Help

### Where can I find more documentation?

- [Installation Guide](installation.md)
- [Getting Started](getting-started.md)
- [Configuration Guide](configuration.md)
- [Providers Guide](providers.md)
- [Sessions Guide](sessions.md)
- [Tools Guide](tools.md)
- [AINative Integrations](ainative-integrations.md)
- [Authentication Guide](authentication.md)
- [Troubleshooting Guide](troubleshooting.md)

### Where can I report bugs or request features?

- **Bug reports**: [GitHub Issues](https://github.com/AINative-studio/ainative-code/issues)
- **Feature requests**: [GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions)
- **Security issues**: security@ainative.studio

### How can I contribute?

We welcome contributions! See:
- [CONTRIBUTING.md](../../CONTRIBUTING.md)
- [Development Guide](../development/README.md)
- [GitHub Repository](https://github.com/AINative-studio/ainative-code)

### Where can I get commercial support?

For enterprise/commercial support:
- Email: support@ainative.studio
- Include your organization details
- Describe your use case

### Is there a community?

Yes!
- [GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions)
- Discord (coming soon)
- Twitter: [@ainativestudio](https://twitter.com/ainativestudio)

## Tips and Tricks

### Quick Tips

1. **Use aliases** for common commands:
   ```bash
   alias ac='ainative-code chat'
   alias acs='ainative-code session'
   ```

2. **Shell integration**:
   ```bash
   # Add to .bashrc or .zshrc
   function ask() {
     ainative-code chat "$*"
   }
   # Usage: ask "How do I use grep?"
   ```

3. **Pipe input**:
   ```bash
   cat error.log | ainative-code chat "Explain this error"
   ```

4. **Use with git**:
   ```bash
   git diff | ainative-code chat "Review these changes"
   ```

5. **Generate commit messages**:
   ```bash
   git diff --staged | ainative-code chat "Generate a commit message"
   ```

### Power User Features

1. **Custom tools**: Create scripts for common tasks
2. **MCP servers**: Build custom integrations
3. **Session templates**: Reuse conversation structures
4. **Batch operations**: Process multiple files
5. **Analytics**: Track and optimize usage

## Version-Specific Questions

### What version should I use?

Always use the latest stable release:

```bash
ainative-code version --check-updates
```

### Are there beta features?

Yes, enable experimental features:

```yaml
app:
  enable_experimental: true
```

Check release notes for current experimental features.

### How do I rollback to a previous version?

**Homebrew:**
```bash
brew install ainative-code@0.1.0
```

**Manual:**
Download specific version from [releases page](https://github.com/AINative-studio/ainative-code/releases).

---

**Have a question not answered here?**

- Check the [documentation](README.md)
- Search [GitHub Issues](https://github.com/AINative-studio/ainative-code/issues)
- Ask in [GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions)
- Contact support@ainative.studio
