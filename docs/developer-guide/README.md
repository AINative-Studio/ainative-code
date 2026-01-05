# Developer Guide

## Overview

Welcome to the AINative Code developer guide! This comprehensive documentation is designed for contributors and integrators who want to understand, extend, or integrate with the AINative Code platform.

AINative Code is a next-generation terminal-based AI coding assistant built in Go that combines multi-provider AI support with native integration to the AINative platform ecosystem. This guide will help you navigate the codebase, understand the architecture, and contribute effectively to the project.

## Purpose

This developer guide serves multiple audiences:

- **Contributors**: Developers who want to fix bugs, add features, or improve the codebase
- **Integrators**: Teams building on top of AINative Code or integrating it into their workflows
- **Custom Provider Developers**: Those implementing new LLM provider integrations
- **Tool Developers**: Developers creating custom tools and MCP servers
- **Platform Developers**: Teams extending the AINative platform integration

## Quick Links

### Getting Started
- [Development Setup](setup.md) - Set up your development environment
- [Architecture Overview](architecture.md) - Understand the system architecture
- [Contributing Guide](contributing.md) - Learn the contribution workflow

### Core Documentation
- [Code Style Guide](code-style.md) - Coding standards and best practices
- [Testing Guide](testing.md) - Writing and running tests
- [Debugging Guide](debugging.md) - Debugging techniques and tools

### Extension Guides
- [Extending Providers](extending-providers.md) - Create custom LLM provider integrations
- [Creating Tools](creating-tools.md) - Build custom tools and MCP servers

## Table of Contents

### 1. [Architecture](architecture.md)
Comprehensive overview of the system architecture, including:
- High-level architecture diagrams
- Component interaction flows
- Package structure and responsibilities
- Design patterns and principles
- Data flow diagrams

### 2. [Development Setup](setup.md)
Everything you need to get started developing:
- Prerequisites and requirements
- Development environment setup
- Building from source
- Running in development mode
- IDE configuration (VSCode, GoLand)
- Troubleshooting common setup issues

### 3. [Testing](testing.md)
Complete testing guide covering:
- Running unit, integration, and E2E tests
- Writing new tests with examples
- Test coverage requirements (80% minimum)
- Mock setup and usage patterns
- Benchmarking and performance testing
- Test organization and best practices

### 4. [Contributing](contributing.md)
Contribution workflow and standards:
- Git workflow (branching, commits, PRs)
- Code review process and expectations
- Issue triage and management
- Documentation standards
- Release process and versioning
- Community guidelines

### 5. [Code Style](code-style.md)
Coding standards and conventions:
- Go coding standards and idiomatic patterns
- Naming conventions for packages, types, and functions
- Error handling patterns and best practices
- Comment and documentation guidelines
- Package organization principles
- Import ordering and grouping

### 6. [Extending Providers](extending-providers.md)
Guide to creating custom LLM provider integrations:
- Provider interface documentation
- Implementing the Provider interface
- Streaming response handling
- Error handling and recovery
- Authentication and configuration
- Testing providers thoroughly
- Complete code examples

### 7. [Creating Tools](creating-tools.md)
Guide to building custom tools and MCP servers:
- Tool interface documentation
- Implementing custom tools
- MCP server development patterns
- Tool registration and discovery
- Input validation and schemas
- Testing tools and servers
- Real-world examples

### 8. [Debugging](debugging.md)
Debugging techniques and troubleshooting:
- Debugging with delve
- Logging best practices
- Profiling (CPU, memory, goroutines)
- Performance analysis
- Common issues and solutions
- Remote debugging
- Production debugging strategies

## Key Concepts

### Project Structure

```
ainative-code/
├── cmd/                    # Command-line entry points
│   └── ainative-code/     # Main CLI application
├── internal/              # Private application code
│   ├── auth/             # Authentication (JWT, OAuth, keychain)
│   ├── provider/         # LLM provider interface and implementations
│   ├── tools/            # Tool interface and built-in tools
│   ├── tui/              # Terminal UI components (Bubble Tea)
│   ├── config/           # Configuration management (Viper)
│   ├── logger/           # Structured logging (zerolog)
│   ├── client/           # AINative platform clients
│   ├── database/         # Local SQLite database
│   └── mcp/              # Model Context Protocol implementation
├── pkg/                   # Public library code
├── tests/                # Integration and E2E tests
├── docs/                 # Documentation
└── scripts/              # Build and utility scripts
```

### Technology Stack

- **Language**: Go 1.21+
- **CLI Framework**: Cobra
- **Configuration**: Viper
- **TUI Framework**: Bubble Tea
- **Logging**: zerolog
- **Database**: SQLite (modernc.org/sqlite)
- **Testing**: Go testing, testify
- **Build**: Make, Docker

### Development Workflow

1. **Fork and Clone** - Fork the repository and clone it locally
2. **Create Branch** - Create a feature branch from `main`
3. **Develop** - Write code following our style guide
4. **Test** - Add tests and ensure coverage meets requirements
5. **Lint** - Run linter and fix any issues
6. **Commit** - Use conventional commit messages
7. **Push** - Push to your fork
8. **Pull Request** - Open a PR with detailed description

### Code Quality Standards

- **Test Coverage**: Minimum 80% coverage required
- **Linting**: Must pass golangci-lint with project configuration
- **Formatting**: All code must be gofmt'd
- **Documentation**: All exported symbols must be documented
- **Error Handling**: All errors must be handled explicitly
- **Security**: Must pass gosec security checks

## Getting Help

### Resources

- **Documentation**: [https://docs.ainative.studio/code](https://docs.ainative.studio/code)
- **API Reference**: [/docs/api-reference](../api-reference)
- **Examples**: [/docs/examples](../examples)
- **Architecture**: [/docs/architecture](../architecture)

### Community

- **GitHub Issues**: [Bug reports and feature requests](https://github.com/AINative-studio/ainative-code/issues)
- **GitHub Discussions**: [Questions and community help](https://github.com/AINative-studio/ainative-code/discussions)
- **Email**: support@ainative.studio

### Support Channels

1. **Check Documentation** - Search existing docs first
2. **Search Issues** - Look for similar issues or discussions
3. **Ask in Discussions** - Post questions to the community
4. **Create Issue** - Report bugs with detailed reproduction steps
5. **Email Support** - For private or security concerns

## Version Information

This guide is maintained for:
- **AINative Code**: v1.0.0+
- **Go**: 1.21+
- **Last Updated**: 2025-01-05

## Contributing to This Guide

The developer guide is itself open source! If you find errors, unclear explanations, or missing information, please:

1. Open an issue describing the problem
2. Submit a PR with improvements
3. Suggest new sections or topics in Discussions

All documentation follows the same contribution guidelines as code.

## Next Steps

1. **New Contributors**: Start with [Development Setup](setup.md)
2. **Architecture Overview**: Read [Architecture](architecture.md)
3. **First Contribution**: Follow [Contributing Guide](contributing.md)
4. **Provider Development**: See [Extending Providers](extending-providers.md)
5. **Tool Development**: See [Creating Tools](creating-tools.md)

---

**Happy Coding!**

Copyright © 2024-2025 AINative Studio. All rights reserved.
