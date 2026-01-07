# Getting Started Guide

Welcome to AINative Code! This guide will walk you through your first steps with the AI-powered development assistant.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Your First Conversation](#your-first-conversation)
4. [Basic Commands](#basic-commands)
5. [Common Workflows](#common-workflows)
6. [Understanding the Interface](#understanding-the-interface)
7. [Tips for Beginners](#tips-for-beginners)
8. [Next Steps](#next-steps)

## Prerequisites

Before getting started, ensure you have:

1. **Installed AINative Code** - See the [Installation Guide](installation.md)
2. **An API Key** - From at least one LLM provider:
   - [Anthropic Claude](https://console.anthropic.com/)
   - [OpenAI](https://platform.openai.com/api-keys)
   - [Google AI Studio](https://makersuite.google.com/app/apikey)
   - Or a local Ollama installation
3. **Internet Connection** - For cloud LLM providers
4. **Terminal/Command Line** - Basic familiarity recommended

## Quick Start

### 1. Verify Installation

First, verify that AINative Code is installed correctly:

```bash
ainative-code --version
```

You should see output like:
```
ainative-code version 0.1.0
```

### 2. Initialize Configuration

Create your configuration file:

```bash
ainative-code setup
```

This creates `~/.config/ainative-code/config.yaml` with default settings.

### 3. Configure Your Provider

Set up your preferred LLM provider. Here's an example using Anthropic Claude:

```bash
# Set Anthropic as your default provider
ainative-code config set llm.default_provider anthropic

# Set your API key
ainative-code config set llm.anthropic.api_key "your-api-key"

# Or use environment variable (recommended)
export ANTHROPIC_API_KEY="your-api-key"
```

For other providers, see the [Providers Guide](providers.md).

### 4. Start Your First Chat

Launch the interactive chat mode:

```bash
ainative-code chat
```

You'll see the AINative Code interface ready for your input.

## Your First Conversation

Let's try a simple coding task to get familiar with the tool.

### Example 1: Explain a Concept

```
You: Explain how to implement a binary search in Go

AI: I'll explain binary search implementation in Go...
[Detailed explanation with code examples]
```

### Example 2: Generate Code

```
You: Create a Go function that reverses a string

AI: Here's a function to reverse a string in Go:

```go
func reverseString(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}
```

This function...
[Explanation follows]
```

### Example 3: Debug Code

```
You: Why is this code giving me a nil pointer error?
[paste your code]

AI: The nil pointer error is occurring because...
[Analysis and solution]
```

### Example 4: Ask for Best Practices

```
You: What are best practices for error handling in Go?

AI: Here are key best practices for error handling in Go:

1. Always check errors...
2. Use custom error types...
[Comprehensive explanation]
```

## Basic Commands

### Chat Commands

```bash
# Start interactive chat
ainative-code chat

# Send a single message (one-shot)
ainative-code chat "Explain goroutines"

# Continue a previous session
ainative-code chat --session-id abc123

# Use a specific model
ainative-code chat --provider openai --model gpt-4

# Stream responses (default)
ainative-code chat --stream

# Disable streaming
ainative-code chat --stream=false
```

### Session Management

```bash
# List your recent sessions
ainative-code session list

# View a specific session
ainative-code session show <session-id>

# Export a session
ainative-code session export <session-id> -o session.json

# Delete a session
ainative-code session delete <session-id>
```

### Configuration

```bash
# View current configuration
ainative-code config show

# Set a configuration value
ainative-code config set key value

# Get a specific configuration value
ainative-code config get key

# List all configuration keys
ainative-code config list
```

### Authentication

```bash
# Login to AINative platform
ainative-code auth login

# Check authentication status
ainative-code auth whoami

# Refresh access token
ainative-code auth token refresh

# Logout
ainative-code auth logout
```

### Version and Help

```bash
# Show version
ainative-code --version

# Show help
ainative-code --help

# Show help for a specific command
ainative-code chat --help
ainative-code session --help
```

## Common Workflows

### Workflow 1: Code Review Assistant

Use AINative Code to review your code:

```bash
# Start a chat session
ainative-code chat

# Then in the chat:
You: Please review this code for potential issues:
[paste your code]

# AI will analyze and provide feedback
```

### Workflow 2: Learning New Concepts

Use it as a learning companion:

```bash
ainative-code chat "Teach me about Go interfaces with examples"
```

The AI will provide:
- Clear explanations
- Code examples
- Common use cases
- Best practices

### Workflow 3: Debugging Help

Get help debugging issues:

```bash
ainative-code chat
```

```
You: I'm getting this error: "panic: runtime error: index out of range"
Here's my code:
[paste code]

AI: [Analyzes the code and identifies the issue]
```

### Workflow 4: Project Scaffolding

Generate project structure and boilerplate:

```bash
ainative-code chat "Create a project structure for a REST API in Go with best practices"
```

### Workflow 5: Documentation Writing

Get help writing documentation:

```bash
ainative-code chat "Write API documentation for this function: [paste code]"
```

### Workflow 6: Test Generation

Generate unit tests for your code:

```bash
ainative-code chat "Generate comprehensive unit tests for this function: [paste code]"
```

## Understanding the Interface

### Chat Interface Elements

When you run `ainative-code chat`, you'll see:

```
┌─────────────────────────────────────────────────────────┐
│ AINative Code - AI Coding Assistant                    │
│ Provider: anthropic | Model: claude-3-5-sonnet-20241022│
├─────────────────────────────────────────────────────────┤
│                                                         │
│ You can ask me anything about coding!                  │
│                                                         │
├─────────────────────────────────────────────────────────┤
│ > Your input here_                                      │
└─────────────────────────────────────────────────────────┘
```

### Keyboard Shortcuts (in chat mode)

- `Ctrl+C` or `Cmd+C` - Exit chat
- `Ctrl+D` - End input (submit message)
- `Ctrl+L` - Clear screen
- `Up/Down arrows` - Navigate command history

### Message Formatting

The AI supports various code formatting:

````markdown
```go
// Go code with syntax highlighting
func main() {
    fmt.Println("Hello, World!")
}
```

```python
# Python code
def hello():
    print("Hello, World!")
```
````

## Tips for Beginners

### 1. Be Specific in Your Questions

Instead of:
```
"How do I use Go?"
```

Try:
```
"How do I read a JSON file in Go and unmarshal it into a struct?"
```

### 2. Provide Context

Give the AI context about your project:

```
"I'm building a REST API using Go and Gin framework. How should I structure my error handling middleware?"
```

### 3. Ask for Explanations

Don't just ask for code - ask for understanding:

```
"Explain how channels work in Go and show me examples of common patterns"
```

### 4. Iterate on Responses

If the first response isn't quite what you need:

```
You: Create a user authentication function in Go
AI: [Provides basic implementation]
You: Can you add JWT token generation to that?
AI: [Adds JWT functionality]
You: Also add password hashing with bcrypt
AI: [Updates with bcrypt]
```

### 5. Use Sessions for Complex Tasks

For multi-step projects, use sessions to maintain context:

```bash
# Start a new session
ainative-code chat

# Work on your project over multiple interactions
# The AI remembers the conversation context
```

### 6. Save Important Sessions

Export sessions you want to keep:

```bash
ainative-code session list
ainative-code session export <session-id> -o my-project-session.json
```

### 7. Leverage Code Examples

Ask for multiple examples:

```
"Show me 3 different ways to implement a singleton pattern in Go"
```

### 8. Request Best Practices

Ask about conventions and best practices:

```
"What are the Go community's best practices for project structure?"
```

### 9. Get Code Reviews

Paste your code and ask for review:

```
"Review this code for performance issues and suggest improvements:
[paste code]"
```

### 10. Learn by Example

Ask the AI to explain existing code:

```
"Explain what this code does line by line:
[paste code]"
```

## Next Steps

Now that you're familiar with the basics, explore more advanced features:

### Configure Advanced Settings

- [Configuration Guide](configuration.md) - Customize your setup
- [Providers Guide](providers.md) - Configure multiple LLM providers
- [Authentication Guide](authentication.md) - Set up AINative platform access

### Explore Platform Integrations

- [AINative Integrations](ainative-integrations.md) - ZeroDB, Design Tokens, Strapi
- [Tools Guide](tools.md) - MCP servers and custom tools

### Master Session Management

- [Sessions Guide](sessions.md) - Advanced session management techniques

### Try Advanced Use Cases

1. **Multi-file Projects**: Ask the AI to help you design and implement entire projects
2. **Code Refactoring**: Get suggestions for improving existing codebases
3. **Architecture Design**: Discuss system design and architecture decisions
4. **Performance Optimization**: Get help identifying and fixing performance bottlenecks
5. **Security Review**: Ask about security best practices and vulnerabilities

### Example Advanced Session

```bash
ainative-code chat
```

```
You: I'm designing a microservices architecture for an e-commerce platform.
I need:
1. User service (authentication, profiles)
2. Product catalog service
3. Order management service
4. Payment processing service

Can you help me design the architecture with Go, including:
- Service communication patterns
- Database choices
- API design
- Deployment strategy

AI: [Provides comprehensive architecture design]

You: Let's start with the user service. Show me the project structure
and implementation of the authentication endpoints.

AI: [Provides detailed implementation]

[Continue building your project with AI assistance]
```

## Practice Exercises

Try these exercises to get comfortable with AINative Code:

### Exercise 1: Hello World Plus

```bash
ainative-code chat "Create a Go program that:
1. Accepts a name as a command-line argument
2. Greets the user with their name
3. Includes proper error handling
4. Has unit tests"
```

### Exercise 2: Data Structure Practice

```bash
ainative-code chat "Implement a thread-safe LRU cache in Go with:
1. Get(key) and Put(key, value) methods
2. Configurable capacity
3. O(1) time complexity for both operations
4. Full unit test coverage"
```

### Exercise 3: Real-World API

```bash
ainative-code chat "Create a REST API in Go that:
1. Uses the Gin framework
2. Implements CRUD operations for a Todo list
3. Includes input validation
4. Has proper error handling
5. Uses PostgreSQL for storage
6. Includes comprehensive tests"
```

## Troubleshooting Common Issues

### Issue: API Key Not Working

```bash
# Verify your API key is set
ainative-code config get llm.anthropic.api_key

# Or check environment variable
echo $ANTHROPIC_API_KEY
```

### Issue: Slow Responses

- Try a faster model (e.g., claude-3-haiku instead of opus)
- Check your internet connection
- Use `--stream` flag for progressive responses

### Issue: Session Not Found

```bash
# List all sessions
ainative-code session list

# Verify session ID
ainative-code session show <session-id>
```

### Issue: Configuration Not Loading

```bash
# Check configuration file location
ainative-code config show

# Re-initialize if needed
ainative-code setup --force
```

## Getting Help

If you need help:

1. **Built-in Help**: `ainative-code --help`
2. **Command Help**: `ainative-code [command] --help`
3. **Documentation**: [https://docs.ainative.studio/code](https://docs.ainative.studio/code)
4. **GitHub Issues**: [https://github.com/AINative-studio/ainative-code/issues](https://github.com/AINative-studio/ainative-code/issues)
5. **Community**: [GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions)
6. **Support Email**: support@ainative.studio

## Further Reading

- [Configuration Guide](configuration.md) - Detailed configuration options
- [Providers Guide](providers.md) - LLM provider setup
- [Sessions Guide](sessions.md) - Advanced session management
- [Tools Guide](tools.md) - MCP and custom tools
- [FAQ](faq.md) - Frequently asked questions
- [Troubleshooting Guide](troubleshooting.md) - Common problems and solutions

Welcome to AI-assisted development with AINative Code!
