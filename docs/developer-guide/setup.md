# Development Setup Guide

## Overview

This guide walks you through setting up a complete development environment for contributing to AINative Code. By the end of this guide, you'll have a working development environment with all necessary tools and dependencies.

## Prerequisites

### Required Software

#### 1. Go Programming Language

AINative Code requires **Go 1.21 or higher**.

**macOS**:
```bash
# Using Homebrew
brew install go

# Verify installation
go version  # Should show go1.21 or higher
```

**Linux**:
```bash
# Download and install
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add to PATH (add to ~/.bashrc or ~/.zshrc)
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Verify installation
go version
```

**Windows**:
```powershell
# Using Chocolatey
choco install golang

# Or download installer from https://go.dev/dl/

# Verify installation
go version
```

#### 2. Git

**macOS**:
```bash
# Using Homebrew
brew install git

# Or use Xcode Command Line Tools
xcode-select --install
```

**Linux**:
```bash
# Debian/Ubuntu
sudo apt-get install git

# Fedora/RHEL
sudo dnf install git
```

**Windows**:
```powershell
# Using Chocolatey
choco install git

# Or download from https://git-scm.com/download/win
```

#### 3. Make (Optional but Recommended)

**macOS**: Included with Xcode Command Line Tools

**Linux**:
```bash
# Debian/Ubuntu
sudo apt-get install build-essential

# Fedora/RHEL
sudo dnf install make
```

**Windows**:
```powershell
# Using Chocolatey
choco install make

# Or use WSL (recommended for Windows development)
```

### Recommended Development Tools

#### golangci-lint (Linter)

```bash
# macOS/Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Verify installation
golangci-lint --version
```

#### delve (Debugger)

```bash
go install github.com/go-delve/delve/cmd/dlv@latest

# Verify installation
dlv version
```

#### gosec (Security Scanner)

```bash
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Verify installation
gosec --version
```

#### govulncheck (Vulnerability Scanner)

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest

# Verify installation
govulncheck -version
```

## Initial Setup

### 1. Fork and Clone Repository

**Fork the repository** on GitHub:
1. Visit https://github.com/AINative-studio/ainative-code
2. Click "Fork" button in the top-right
3. Select your GitHub account

**Clone your fork**:
```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/ainative-code.git
cd ainative-code

# Add upstream remote
git remote add upstream https://github.com/AINative-studio/ainative-code.git

# Verify remotes
git remote -v
```

### 2. Install Dependencies

```bash
# Download all Go dependencies
go mod download

# Verify dependencies
go mod verify

# Tidy up (removes unused dependencies)
go mod tidy
```

### 3. Verify Setup

```bash
# Run tests to verify everything works
make test

# Or without Make
go test -v -race ./...

# Build the project
make build

# Or without Make
go build -o build/ainative-code ./cmd/ainative-code
```

If all tests pass and the build succeeds, your setup is complete!

## Building from Source

### Development Build

```bash
# Quick build (no optimization)
make build

# Build with race detector (slower, for development)
go build -race -o build/ainative-code ./cmd/ainative-code

# Build with version information
make build VERSION=dev
```

### Production Build

```bash
# Optimized build with version info
make build LDFLAGS="-s -w"

# Build for all platforms
make build-all

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o build/ainative-code-linux-amd64 ./cmd/ainative-code
```

### Build Output

Binaries are created in the `build/` directory:
```
build/
├── ainative-code                    # Current platform
├── ainative-code-darwin-amd64      # macOS Intel
├── ainative-code-darwin-arm64      # macOS Apple Silicon
├── ainative-code-linux-amd64       # Linux x64
├── ainative-code-linux-arm64       # Linux ARM64
└── ainative-code-windows-amd64.exe # Windows x64
```

## Running in Development Mode

### Basic Execution

```bash
# Run directly with go run
go run ./cmd/ainative-code

# Run with arguments
go run ./cmd/ainative-code chat "Hello, world"

# Run with debug logging
go run ./cmd/ainative-code --verbose chat
```

### Using Built Binary

```bash
# Build first
make build

# Run the binary
./build/ainative-code chat

# Or install to GOPATH/bin
make install
ainative-code chat
```

### Development Configuration

Create a development configuration file:

```bash
# Create config directory
mkdir -p ~/.config/ainative-code

# Create dev config
cat > ~/.config/ainative-code/config.yaml <<EOF
# Development Configuration
providers:
  anthropic:
    api_key: "\${ANTHROPIC_API_KEY}"
    model: "claude-3-5-sonnet-20241022"
    max_tokens: 4096
    temperature: 0.7

  mock:
    # Mock provider for testing
    enabled: true

# AINative Platform (optional for development)
ainative:
  auth:
    token_cache: "~/.config/ainative-code/tokens.json"
    auto_refresh: true

# Development settings
logging:
  level: "debug"
  format: "text"  # More readable than JSON
  output: "stdout"

# TUI Settings
ui:
  theme: "dark"
  colors:
    primary: "#6366F1"
    secondary: "#8B5CF6"
EOF
```

### Environment Variables

Set up development environment variables:

```bash
# Add to ~/.bashrc, ~/.zshrc, or ~/.profile

# LLM Provider API Keys
export ANTHROPIC_API_KEY="your-key-here"
export OPENAI_API_KEY="your-key-here"

# Development settings
export AINATIVE_ENV="development"
export AINATIVE_LOG_LEVEL="debug"

# Optional: AINative Platform
export AINATIVE_PLATFORM_URL="https://api.ainative.studio"
```

## IDE Configuration

### Visual Studio Code

#### Recommended Extensions

Install the following extensions:

1. **Go** (`golang.go`) - Official Go extension
2. **Go Test Explorer** (`ethan-reesor.vscode-go-test-adapter`)
3. **Error Lens** (`usernamehw.errorlens`)
4. **GitLens** (`eamodio.gitlens`)

#### Configuration

Create `.vscode/settings.json`:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "gofmt",
  "go.formatOnSave": true,
  "go.vetOnSave": "workspace",
  "go.testOnSave": false,
  "go.coverOnSave": false,
  "go.buildOnSave": "off",
  "go.installDependenciesWhenBuilding": true,
  "go.testFlags": ["-v", "-race"],
  "go.testTimeout": "60s",
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  },
  "gopls": {
    "ui.semanticTokens": true,
    "ui.codelenses": {
      "generate": true,
      "test": true
    }
  }
}
```

Create `.vscode/launch.json` for debugging:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Chat Command",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/ainative-code",
      "args": ["chat"],
      "env": {
        "AINATIVE_LOG_LEVEL": "debug"
      }
    },
    {
      "name": "Launch Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/internal/provider",
      "args": ["-v"]
    },
    {
      "name": "Attach to Running Process",
      "type": "go",
      "request": "attach",
      "mode": "local",
      "processId": "${command:pickProcess}"
    }
  ]
}
```

Create `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build",
      "type": "shell",
      "command": "make build",
      "group": {
        "kind": "build",
        "isDefault": true
      }
    },
    {
      "label": "Test",
      "type": "shell",
      "command": "make test",
      "group": {
        "kind": "test",
        "isDefault": true
      }
    },
    {
      "label": "Lint",
      "type": "shell",
      "command": "make lint"
    },
    {
      "label": "Coverage",
      "type": "shell",
      "command": "make test-coverage"
    }
  ]
}
```

### GoLand / IntelliJ IDEA

#### Configuration

1. **Open Project**: File → Open → Select `ainative-code` directory

2. **Configure Go SDK**:
   - File → Settings → Languages & Frameworks → Go → GOROOT
   - Select Go 1.21+ installation

3. **Enable Go Modules**:
   - File → Settings → Languages & Frameworks → Go → Go Modules
   - Check "Enable Go modules integration"
   - Set "GOPROXY" to default

4. **Configure File Watchers**:
   - File → Settings → Tools → File Watchers
   - Add watcher for `gofmt` on save

5. **Configure Run Configurations**:

**Chat Command**:
- Run → Edit Configurations → Add New → Go Build
- Name: "Chat Command"
- Run kind: "Package"
- Package path: `github.com/AINative-studio/ainative-code/cmd/ainative-code`
- Program arguments: `chat`
- Environment: `AINATIVE_LOG_LEVEL=debug`

**Run Tests**:
- Run → Edit Configurations → Add New → Go Test
- Name: "All Tests"
- Test kind: "Directory"
- Directory: `$ProjectFileDir$`
- Pattern: `.*`
- Environment: `AINATIVE_ENV=test`

#### Debugging

1. Set breakpoints by clicking in the gutter
2. Right-click on run configuration
3. Select "Debug 'Chat Command'"

### Vim/Neovim

#### vim-go Configuration

Add to `.vimrc` or `init.vim`:

```vim
" Install vim-go
Plug 'fatih/vim-go', { 'do': ':GoUpdateBinaries' }

" Go syntax highlighting
let g:go_highlight_functions = 1
let g:go_highlight_methods = 1
let g:go_highlight_fields = 1
let g:go_highlight_types = 1
let g:go_highlight_operators = 1
let g:go_highlight_build_constraints = 1

" Auto formatting
let g:go_fmt_command = "gofmt"
let g:go_fmt_autosave = 1

" Linting
let g:go_metalinter_autosave = 1
let g:go_metalinter_command = "golangci-lint"

" Testing
let g:go_test_timeout = '60s'

" Key mappings
autocmd FileType go nmap <leader>b  <Plug>(go-build)
autocmd FileType go nmap <leader>r  <Plug>(go-run)
autocmd FileType go nmap <leader>t  <Plug>(go-test)
autocmd FileType go nmap <leader>c  <Plug>(go-coverage-toggle)
```

## Development Workflow

### Daily Development Flow

```bash
# 1. Sync with upstream
git checkout main
git pull upstream main
git push origin main

# 2. Create feature branch
git checkout -b feature/my-feature

# 3. Make changes
# ... edit code ...

# 4. Run tests frequently
make test

# 5. Run linter
make lint

# 6. Commit changes
git add .
git commit -m "feat: add new feature"

# 7. Push to fork
git push origin feature/my-feature

# 8. Create pull request on GitHub
```

### Running Tests During Development

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./internal/provider/...

# Run specific test
go test -v -run TestProviderChat ./internal/provider

# Run with race detector
go test -race ./...

# Watch for changes and re-run tests
# Install: go install github.com/cespare/reflex@latest
reflex -r '\.go$' -s -- make test
```

### Code Quality Checks

```bash
# Format code
make fmt

# Check formatting
make fmt-check

# Run linter
make lint

# Run security scanner
make security

# Check for vulnerabilities
make vuln-check

# Run all CI checks locally
make ci
```

## Troubleshooting

### Common Issues

#### 1. "Cannot find package"

**Problem**: `go build` fails with package import errors

**Solution**:
```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download
go mod tidy
```

#### 2. "golangci-lint not found"

**Problem**: `make lint` fails

**Solution**:
```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Verify installation
golangci-lint --version

# Add to PATH if needed
export PATH=$PATH:$(go env GOPATH)/bin
```

#### 3. Tests fail with "database locked"

**Problem**: SQLite tests fail intermittently

**Solution**:
```bash
# Run tests sequentially
go test -p 1 ./...

# Or increase timeout
go test -timeout 5m ./...
```

#### 4. "permission denied" when running binary

**Problem**: Built binary is not executable

**Solution**:
```bash
chmod +x build/ainative-code
```

#### 5. CGO errors on macOS

**Problem**: Build fails with CGO-related errors

**Solution**:
```bash
# Install Xcode Command Line Tools
xcode-select --install

# Or use pure Go SQLite driver (already configured)
# The project uses modernc.org/sqlite which doesn't require CGO
```

### Getting Help

If you encounter issues not covered here:

1. Check [GitHub Issues](https://github.com/AINative-studio/ainative-code/issues)
2. Search [GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions)
3. Ask in Discussions with:
   - Your OS and version
   - Go version (`go version`)
   - Steps to reproduce
   - Error messages
   - What you've tried

## Next Steps

Now that your development environment is set up:

1. **Understand the Architecture**: Read [Architecture Guide](architecture.md)
2. **Learn Testing Practices**: Read [Testing Guide](testing.md)
3. **Review Code Style**: Read [Code Style Guide](code-style.md)
4. **Make Your First Contribution**: Read [Contributing Guide](contributing.md)

## Quick Reference

### Essential Commands

```bash
# Build
make build              # Build for current platform
make build-all          # Build for all platforms

# Test
make test              # Run tests
make test-coverage     # Run tests with coverage
make test-integration  # Run integration tests

# Code Quality
make fmt               # Format code
make lint              # Run linter
make vet               # Run go vet
make ci                # Run all CI checks

# Run
make run               # Build and run
make install           # Install to GOPATH/bin

# Clean
make clean             # Remove build artifacts

# Docker
make docker-build      # Build Docker image
make docker-run        # Run Docker container

# Help
make help              # Show all available commands
```

### Useful Go Commands

```bash
# Module management
go mod download        # Download dependencies
go mod tidy           # Remove unused dependencies
go mod verify         # Verify dependencies

# Testing
go test ./...         # Run all tests
go test -v ./...      # Verbose output
go test -race ./...   # Run with race detector
go test -cover ./...  # Show coverage

# Building
go build              # Build current package
go install            # Build and install
go clean              # Remove build artifacts

# Analysis
go vet ./...          # Run go vet
go fmt ./...          # Format code
```

---

**Last Updated**: 2025-01-05
