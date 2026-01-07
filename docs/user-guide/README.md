# AINative Code User Guide

Welcome to the comprehensive user guide for AINative Code - your AI-powered terminal coding assistant!

## Overview

AINative Code is a next-generation terminal-based AI coding assistant that combines the best features of open-source AI CLI tools with native integration to the AINative platform ecosystem. Whether you're writing code, debugging issues, learning new concepts, or managing your development workflow, AINative Code provides intelligent assistance right in your terminal.

### Key Features

- **Multi-Provider AI Support**: Use Anthropic Claude, OpenAI GPT, Google Gemini, AWS Bedrock, Azure OpenAI, or local models via Ollama
- **Intelligent Session Management**: Save, resume, and organize conversations with full context preservation
- **AINative Platform Integration**: Native access to ZeroDB, Design Tokens, Strapi CMS, and RLHF feedback systems
- **Advanced Tool System**: Execute commands, manage files, and integrate custom tools via MCP protocol
- **Streaming Responses**: Real-time AI responses for immediate feedback
- **Cross-Platform**: Works on macOS, Linux, and Windows

## Documentation Structure

This user guide is organized into the following sections:

### Getting Started

Perfect for new users who want to start using AINative Code quickly.

- **[Installation Guide](installation.md)** - Complete installation instructions for all platforms
- **[Getting Started](getting-started.md)** - Quick start tutorial and first steps
- **[FAQ](faq.md)** - Frequently asked questions and quick answers

### Configuration & Setup

Learn how to configure AINative Code for your workflow.

- **[Configuration Guide](configuration.md)** - Comprehensive configuration options and examples
- **[LLM Providers](providers.md)** - Setting up and using different AI providers
- **[Authentication](authentication.md)** - Platform and provider authentication setup

### Core Features

Master the core functionality of AINative Code.

- **[Session Management](sessions.md)** - Creating, managing, and organizing conversations
- **[Tools Usage](tools.md)** - Built-in tools, MCP integration, and custom tools
- **[AINative Integrations](ainative-integrations.md)** - ZeroDB, Design Tokens, Strapi, and RLHF

### Support & Reference

Resources for troubleshooting and getting help.

- **[Troubleshooting Guide](troubleshooting.md)** - Common issues and solutions
- **[FAQ](faq.md)** - Frequently asked questions

## Quick Links

### Installation

```bash
# macOS (Homebrew)
brew install ainative-studio/tap/ainative-code

# Linux (Direct download)
curl -LO https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-linux-amd64
chmod +x ainative-code-linux-amd64
sudo mv ainative-code-linux-amd64 /usr/local/bin/ainative-code

# Verify installation
ainative-code --version
```

### First Steps

```bash
# Initialize configuration
ainative-code setup

# Set up your LLM provider
export ANTHROPIC_API_KEY="your-api-key"
ainative-code config set llm.default_provider anthropic

# Start your first chat
ainative-code chat
```

### Common Commands

```bash
# Chat with AI
ainative-code chat "How do I use Go channels?"

# Resume last session
ainative-code chat --resume

# List sessions
ainative-code session list

# View configuration
ainative-code config show

# Get help
ainative-code --help
```

## Use Cases

### Code Generation

Generate production-ready code with AI assistance:

```bash
ainative-code chat "Create a REST API in Go with:
- User authentication
- CRUD operations
- Input validation
- Error handling
- Unit tests"
```

### Code Review

Get detailed code reviews and suggestions:

```bash
ainative-code chat "Review this code for security and performance:
[paste your code]"
```

### Learning & Exploration

Learn new concepts and technologies:

```bash
ainative-code chat "Explain microservices architecture with examples in Go"
```

### Debugging

Get help debugging issues:

```bash
ainative-code chat "I'm getting this error: [error message]
Here's my code: [code]
What's wrong?"
```

### Project Scaffolding

Generate project structures and boilerplate:

```bash
ainative-code chat "Create a project structure for a React app with TypeScript, including:
- Component organization
- State management with Redux
- Testing setup
- Build configuration"
```

### Documentation

Generate comprehensive documentation:

```bash
ainative-code chat "Generate API documentation for this code:
[paste code]"
```

## Features at a Glance

### Multi-Provider Support

| Provider | Best For | Cost | Speed |
|----------|----------|------|-------|
| Anthropic Claude | Complex reasoning, code generation | $$ | Fast |
| OpenAI GPT | General purpose | $$-$$$ | Fast |
| Google Gemini | Multimodal tasks | $$ | Fast |
| AWS Bedrock | Enterprise, AWS integration | $$$ | Fast |
| Azure OpenAI | Enterprise, Microsoft integration | $$$ | Fast |
| Ollama | Privacy, offline use | Free | Medium |

### Session Management

- **Auto-save**: Conversations automatically saved locally
- **Resume**: Pick up where you left off with full context
- **Export/Import**: Share sessions with team members
- **Search**: Find past conversations by content or metadata
- **Sync**: Optional cloud backup via ZeroDB

### Platform Integrations

- **ZeroDB**: Vector search, NoSQL tables, PostgreSQL
- **Design Tokens**: Extract from Figma, generate code
- **Strapi CMS**: Content management integration
- **RLHF**: Provide feedback to improve AI responses
- **Analytics**: Track usage and costs

### Tool System

- **Bash**: Execute terminal commands
- **File Operations**: Read, write, and manage files
- **Web Fetch**: Retrieve and analyze web content
- **Code Analysis**: Analyze code structure and patterns
- **MCP Integration**: Custom tools from external servers

## Configuration Examples

### Minimal Setup (Anthropic Only)

```yaml
app:
  environment: development

llm:
  default_provider: anthropic
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    model: claude-3-5-sonnet-20241022
    max_tokens: 4096

logging:
  level: info
  format: console
```

### Multi-Provider with Fallback

```yaml
llm:
  default_provider: anthropic

  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    model: claude-3-5-sonnet-20241022

  openai:
    api_key: "${OPENAI_API_KEY}"
    model: gpt-4-turbo-preview

  fallback:
    enabled: true
    providers:
      - anthropic
      - openai
```

### Full Platform Integration

```yaml
llm:
  default_provider: anthropic
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"

platform:
  authentication:
    method: oauth2

services:
  zerodb:
    enabled: true
    endpoint: postgresql://zerodb.ainative.studio:5432

  design:
    enabled: true

  strapi:
    enabled: true

  rlhf:
    enabled: true
```

## Best Practices

### 1. Security

- Use environment variables for API keys
- Never commit secrets to version control
- Rotate API keys regularly
- Enable encryption for sensitive deployments

### 2. Cost Management

- Use appropriate models for tasks (Haiku for simple, Opus for complex)
- Monitor usage with analytics
- Enable caching for repeated queries
- Consider local models (Ollama) for development

### 3. Session Organization

- One session per topic or feature
- Use descriptive titles
- Tag sessions for easy filtering
- Export important sessions for backup
- Clean up old sessions regularly

### 4. Effective Prompting

- Be specific in your requests
- Provide relevant context
- Ask for explanations to learn
- Iterate and refine responses
- Request tests and documentation

### 5. Performance

- Use streaming for long responses
- Enable caching where appropriate
- Choose faster models for simple tasks
- Monitor and optimize token usage

## Getting Help

### Documentation

- **User Guide**: You're reading it!
- **API Reference**: [../api/README.md](../api/README.md)
- **Development Guide**: [../development/README.md](../development/README.md)
- **Architecture Guide**: [../architecture/README.md](../architecture/README.md)

### Community & Support

- **GitHub Issues**: [Report bugs](https://github.com/AINative-studio/ainative-code/issues)
- **GitHub Discussions**: [Ask questions](https://github.com/AINative-studio/ainative-code/discussions)
- **Email Support**: support@ainative.studio
- **Documentation**: [https://docs.ainative.studio/code](https://docs.ainative.studio/code)

### Contributing

We welcome contributions! See:

- [CONTRIBUTING.md](../../CONTRIBUTING.md)
- [Development Guide](../development/README.md)
- [GitHub Repository](https://github.com/AINative-studio/ainative-code)

## Version Information

This documentation is for **AINative Code v0.1.0**.

Check your version:
```bash
ainative-code --version
```

Check for updates:
```bash
ainative-code version --check-updates
```

## License

AINative Code is released under the MIT License. See [LICENSE](../../LICENSE) for details.

**Copyright © 2024 AINative Studio. All rights reserved.**

## Next Steps

Ready to get started? Here's what to do next:

1. **[Install AINative Code](installation.md)** - Choose your platform and install
2. **[Complete the Getting Started Guide](getting-started.md)** - Learn the basics in 15 minutes
3. **[Configure Your Setup](configuration.md)** - Customize for your workflow
4. **[Set Up Providers](providers.md)** - Configure your preferred AI models
5. **[Start Coding!](getting-started.md#your-first-conversation)** - Begin your AI-assisted development journey

---

**Happy Coding with AI Assistance!**

Need help? Check the [FAQ](faq.md) or visit [GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions).
