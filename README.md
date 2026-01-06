# AINative Code

[![CI](https://github.com/AINative-studio/ainative-code/workflows/CI/badge.svg)](https://github.com/AINative-studio/ainative-code/actions/workflows/ci.yml)
[![Release](https://github.com/AINative-studio/ainative-code/workflows/Release/badge.svg)](https://github.com/AINative-studio/ainative-code/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/AINative-studio/ainative-code/branch/main/graph/badge.svg)](https://codecov.io/gh/AINative-studio/ainative-code)
[![Go Report Card](https://goreportcard.com/badge/github.com/AINative-studio/ainative-code)](https://goreportcard.com/report/github.com/AINative-studio/ainative-code)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/AINative-studio/ainative-code)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/AINative-studio/ainative-code)](https://github.com/AINative-studio/ainative-code/releases/latest)

> AI-Native Development, Natively

A next-generation terminal-based AI coding assistant that combines the best features of open-source AI CLI tools with native integration to the AINative platform ecosystem.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
  - [macOS](#macos)
  - [Linux](#linux)
  - [Windows](#windows)
  - [Docker](#docker)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Usage](#usage)
- [Completed Features](#completed-features)
- [Development](#development)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)
- [Documentation](#documentation)
- [Support](#support)
- [Acknowledgments](#acknowledgments)

## Features

- **Multi-Provider AI Support**: Anthropic Claude, OpenAI GPT, Google Gemini, AWS Bedrock, Azure OpenAI, and Ollama
- **Beautiful TUI**: Sophisticated Bubble Tea-based terminal interface
- **AINative Platform Integration**: Native access to ZeroDB, Design Tokens, Strapi CMS, and RLHF systems
- **Hybrid Authentication**: JWT/OAuth 2.0 for AINative services, API keys for LLM providers
- **Streaming Responses**: Real-time AI responses with Server-Sent Events
- **Cross-Platform**: macOS, Linux, and Windows support

## Installation

### macOS

#### Homebrew (Recommended)
```bash
brew install ainative-studio/tap/ainative-code
```

#### Direct Download
```bash
# Intel
curl -LO https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-darwin-amd64
chmod +x ainative-code-darwin-amd64
sudo mv ainative-code-darwin-amd64 /usr/local/bin/ainative-code

# Apple Silicon
curl -LO https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-darwin-arm64
chmod +x ainative-code-darwin-arm64
sudo mv ainative-code-darwin-arm64 /usr/local/bin/ainative-code
```

### Linux

```bash
# AMD64
curl -LO https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-linux-amd64
chmod +x ainative-code-linux-amd64
sudo mv ainative-code-linux-amd64 /usr/local/bin/ainative-code

# ARM64
curl -LO https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-linux-arm64
chmod +x ainative-code-linux-arm64
sudo mv ainative-code-linux-arm64 /usr/local/bin/ainative-code
```

### Windows

```powershell
# Download the latest release
Invoke-WebRequest -Uri "https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-windows-amd64.exe" -OutFile "ainative-code.exe"

# Add to PATH or move to a directory in your PATH
Move-Item ainative-code.exe C:\Windows\System32\
```

### Docker

```bash
docker pull ghcr.io/ainative-studio/ainative-code:latest
docker run -it --rm ghcr.io/ainative-studio/ainative-code:latest
```

## Quick Start

1. **Initialize configuration**:
   ```bash
   ainative-code init
   ```

2. **Authenticate with AINative platform** (optional):
   ```bash
   ainative-code auth login
   ```

3. **Configure your preferred LLM provider**:
   ```bash
   ainative-code config set provider anthropic
   ainative-code config set anthropic.api_key "your-api-key"
   ```

4. **Start coding**:
   ```bash
   ainative-code chat
   ```

## Configuration

Configuration file location: `~/.config/ainative-code/config.yaml`

Example configuration:

```yaml
# LLM Provider Configuration
providers:
  anthropic:
    api_key: "$(pass show anthropic)"
    model: "claude-3-5-sonnet-20241022"
    max_tokens: 4096
    temperature: 0.7

  openai:
    api_key: "${OPENAI_API_KEY}"
    model: "gpt-4"
    max_tokens: 4096

# AINative Platform Configuration
ainative:
  auth:
    token_cache: "~/.config/ainative-code/tokens.json"
    auto_refresh: true

  zerodb:
    endpoint: "https://zerodb.ainative.studio"

  strapi:
    endpoint: "https://cms.ainative.studio"

# TUI Settings
ui:
  theme: "dark"
  colors:
    primary: "#6366F1"
    secondary: "#8B5CF6"
    success: "#10B981"
    error: "#EF4444"
```

## Usage

### Chat Mode
```bash
# Start interactive chat
ainative-code chat

# Chat with specific model
ainative-code chat --model claude-3-opus-20240229

# One-shot chat
ainative-code chat "Explain how to implement OAuth 2.0"
```

### Code Generation
```bash
# Generate code from prompt
ainative-code generate "Create a REST API handler for user authentication"

# Generate and save to file
ainative-code generate "Create a REST API handler" -o handler.go
```

### AINative Platform Operations
```bash
# Query ZeroDB
ainative-code zerodb query "SELECT * FROM users WHERE active = true"

# Extract design tokens
ainative-code design-tokens extract --format json

# Sync with Strapi CMS
ainative-code strapi sync
```

## Completed Features

### Logging System (TASK-009) ✅

The project includes a production-ready structured logging system with:

- **Structured Logging**: JSON and text output formats for easy parsing and debugging
- **Multiple Log Levels**: DEBUG, INFO, WARN, ERROR, FATAL with configurable minimum level
- **Context-Aware Logging**: Automatic inclusion of request IDs, session IDs, and user IDs from Go context
- **Log Rotation**: Automatic rotation based on file size, age, and backup count using lumberjack
- **High Performance**: ~2μs per log operation, zero allocations for disabled log levels
- **Thread-Safe**: Global logger with mutex protection for concurrent use
- **Flexible Configuration**: YAML-based or programmatic configuration

#### Logging Quick Start

```go
import "github.com/AINative-studio/ainative-code/internal/logger"

func main() {
    // Use global logger with default configuration
    logger.Info("Application started")

    // Structured logging with fields
    logger.InfoWithFields("User logged in", map[string]interface{}{
        "user_id": "user123",
        "email": "user@example.com",
    })

    // Context-aware logging
    ctx := logger.WithRequestID(context.Background(), "req-123")
    log := logger.WithContext(ctx)
    log.Info("Processing request") // Automatically includes request_id
}
```

**Performance Benchmarks** (Apple M3):

| Operation | Time/op | Allocations |
|-----------|---------|-------------|
| Simple message | 2.0 μs | 0 allocs |
| Formatted message | 2.2 μs | 1 allocs |
| Structured fields (5) | 2.9 μs | 10 allocs |
| Context-aware | 2.2 μs | 0 allocs |
| Disabled level | 0.7 ns | 0 allocs |

See [docs/logging.md](docs/logging.md) for complete logging documentation.

## Development

### Prerequisites

- Go 1.21 or higher
- Make (optional, for using Makefile)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/AINative-studio/ainative-code.git
cd ainative-code

# Build
make build

# Run tests
make test

# Run linter
make lint

# Install locally
make install
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run integration tests
make test-integration

# Run specific test
go test -v ./internal/llm/...
```

### Project Structure

```
ainative-code/
├── cmd/
│   └── ainative-code/      # Main CLI entry point
├── internal/               # Private application code
│   ├── auth/              # Authentication logic
│   ├── llm/               # LLM provider implementations
│   ├── tui/               # Terminal UI components
│   ├── config/            # Configuration management
│   ├── api/               # API clients (ZeroDB, Strapi, etc.)
│   └── database/          # Local SQLite database
├── pkg/                   # Public library code
├── configs/               # Configuration files
├── docs/                  # Documentation
├── scripts/               # Build and utility scripts
├── tests/                 # Integration and E2E tests
└── .github/
    └── workflows/         # CI/CD workflows
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on our development process, coding standards, and how to submit pull requests.

Before contributing, please:
1. Read our [Code of Conduct](CODE_OF_CONDUCT.md)
2. Check existing [issues](https://github.com/AINative-studio/ainative-code/issues) and [pull requests](https://github.com/AINative-studio/ainative-code/pulls)
3. Review the [Development Guide](docs/development/README.md)

Quick contribution steps:
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes and add tests
4. Ensure all tests pass (`make test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

**Copyright © 2024 AINative Studio. All rights reserved.**

## Brand

**AINative Code** - AI-Native Development, Natively

Brand Colors:
- Primary: #6366F1 (Indigo)
- Secondary: #8B5CF6 (Purple)
- Success: #10B981 (Green)
- Error: #EF4444 (Red)

## Documentation

Comprehensive documentation is available in the `/docs` directory:

- **[Architecture Guide](docs/architecture/README.md)** - System design and technical architecture
- **[User Guide](docs/user-guide/README.md)** - Getting started and usage instructions
- **[API Reference](docs/api/README.md)** - Detailed API documentation
- **[Development Guide](docs/development/README.md)** - Contributing and development setup
- **[Examples](docs/examples/README.md)** - Code examples and use cases

## Support

- **Documentation**: [https://docs.ainative.studio/code](https://docs.ainative.studio/code)
- **Issues**: [GitHub Issues](https://github.com/AINative-studio/ainative-code/issues)
- **Discussions**: [GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions)
- **Email**: support@ainative.studio

## Acknowledgments

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [zerolog](https://github.com/rs/zerolog) - High-performance logging

Inspired by projects like:
- Aider
- GitHub Copilot CLI
- Cursor

---

**AI-Native Development, Natively**
