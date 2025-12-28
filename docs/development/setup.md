# Development Environment Setup Guide

This guide will walk you through setting up your development environment for AINative Code.

## Table of Contents

- [Prerequisites](#prerequisites)
- [System Requirements](#system-requirements)
- [Initial Setup](#initial-setup)
- [Development Tools](#development-tools)
- [IDE Configuration](#ide-configuration)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)
- [Next Steps](#next-steps)

## Prerequisites

### Required Software

1. **Go 1.21 or higher**
   - Download from: https://go.dev/dl/
   - Verify installation: `go version`
   - Required for building and running the application

2. **Git**
   - Download from: https://git-scm.com/downloads
   - Verify installation: `git --version`
   - Required for version control

3. **C Compiler** (for CGO/SQLite support)
   - **macOS**: Install Xcode Command Line Tools
     ```bash
     xcode-select --install
     ```
   - **Linux (Debian/Ubuntu)**:
     ```bash
     sudo apt-get update
     sudo apt-get install build-essential
     ```
   - **Linux (RHEL/CentOS)**:
     ```bash
     sudo yum groupinstall "Development Tools"
     ```
   - **Windows**: Install MinGW-w64 or TDM-GCC
     - Download from: https://www.mingw-w64.org/

4. **Make** (optional, but recommended)
   - **macOS/Linux**: Usually pre-installed
   - **Windows**: Install via Chocolatey: `choco install make`

### Recommended Software

1. **golangci-lint** - For code linting
   ```bash
   # macOS
   brew install golangci-lint

   # Linux and Windows
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
   ```

2. **gosec** - For security scanning
   ```bash
   go install github.com/securego/gosec/v2/cmd/gosec@latest
   ```

3. **govulncheck** - For vulnerability checking
   ```bash
   go install golang.org/x/vuln/cmd/govulncheck@latest
   ```

4. **sqlc** - For type-safe SQL code generation
   ```bash
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   ```

## System Requirements

### Minimum Requirements

- **CPU**: 2 cores
- **RAM**: 4 GB
- **Disk**: 500 MB free space (for source code and dependencies)
- **OS**: macOS 10.15+, Linux (Ubuntu 20.04+, RHEL 8+), Windows 10+

### Recommended Requirements

- **CPU**: 4+ cores
- **RAM**: 8+ GB
- **Disk**: 2+ GB free space
- **OS**: Latest stable versions

## Initial Setup

### 1. Clone the Repository

```bash
# HTTPS
git clone https://github.com/AINative-studio/ainative-code.git
cd ainative-code

# SSH (if you have SSH keys configured)
git clone git@github.com:AINative-studio/ainative-code.git
cd ainative-code
```

### 2. Set Up Go Environment

Configure your Go environment variables (if not already set):

```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export CGO_ENABLED=1  # Required for SQLite support
```

Reload your shell configuration:
```bash
source ~/.bashrc  # or ~/.zshrc
```

### 3. Download Dependencies

```bash
# Download all Go dependencies
go mod download

# Verify dependencies
go mod verify

# Tidy up dependencies
go mod tidy
```

Or use the Makefile:
```bash
make deps
```

### 4. Verify Installation

Run the verification script:
```bash
./verify-deps.sh
```

Expected output:
```
=== Verifying AINative Code Dependencies ===

✓ Go is installed (go1.21.x)
✓ Git is installed
✓ C compiler is available
✓ All Go dependencies are present
✓ SQLC is installed

=== All required dependencies are installed! ===
```

### 5. Build the Application

```bash
# Build using Makefile (recommended)
make build

# Or build manually
go build -o build/ainative-code ./cmd/ainative-code
```

Verify the build:
```bash
./build/ainative-code version
```

## Development Tools

### golangci-lint Configuration

The project uses a comprehensive `.golangci.yml` configuration with:
- 40+ enabled linters
- Cyclomatic complexity checks
- Security scanning
- Code formatting validation

Run linting:
```bash
make lint
```

### SQLC Configuration

SQLC generates type-safe Go code from SQL queries. Configuration is in `sqlc.yaml`:

```yaml
version: "2"
sql:
  - engine: "sqlite"
    queries: "internal/database/queries/"
    schema: "internal/database/schema/"
    gen:
      go:
        package: "database"
        out: "internal/database"
```

Generate database code:
```bash
sqlc generate
# or
make sqlc-generate  # if available in Makefile
```

### Docker Setup (Optional)

For containerized development:

```bash
# Build Docker image
make docker-build

# Run in Docker container
make docker-run

# Or manually
docker build -t ainative-code:dev .
docker run -it --rm ainative-code:dev
```

## IDE Configuration

### Visual Studio Code

Recommended extensions:
- **Go** (golang.go) - Official Go extension
- **Go Test Explorer** - Test runner UI
- **Go Coverage Viewer** - Coverage visualization
- **Better Comments** - Enhanced comment highlighting

Settings (`.vscode/settings.json`):
```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.formatTool": "gofmt",
  "go.useLanguageServer": true,
  "go.buildOnSave": "workspace",
  "go.vetOnSave": "workspace",
  "editor.formatOnSave": true,
  "go.testFlags": ["-v", "-race"],
  "go.coverageOptions": "showCoveredCodeOnly"
}
```

### GoLand / IntelliJ IDEA

1. Open the project directory
2. Enable Go Modules: `Settings > Go > Go Modules > Enable Go modules integration`
3. Set GOROOT to your Go installation
4. Configure golangci-lint:
   - `Settings > Tools > File Watchers > + > golangci-lint`
   - Program: `$GOPATH/bin/golangci-lint`
   - Arguments: `run $FileDir$`

### Vim/Neovim

Install vim-go plugin:
```vim
" Using vim-plug
Plug 'fatih/vim-go', { 'do': ':GoUpdateBinaries' }

" Configuration
let g:go_fmt_command = "goimports"
let g:go_auto_type_info = 1
let g:go_metalinter_enabled = ['golangci-lint']
let g:go_metalinter_autosave = 1
```

## Verification

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/logger/...

# Run tests with race detection
go test -race ./...
```

### Check Code Quality

```bash
# Format code
make fmt

# Check formatting
make fmt-check

# Run vet
make vet

# Run linter
make lint

# Run security scan
make security

# Check for vulnerabilities
make vuln-check

# Run all CI checks
make ci
```

### Build All Platforms

```bash
# Build for all platforms (macOS, Linux, Windows)
make build-all

# Outputs will be in ./build/ directory:
# - ainative-code-darwin-amd64
# - ainative-code-darwin-arm64
# - ainative-code-linux-amd64
# - ainative-code-linux-arm64
# - ainative-code-windows-amd64.exe
```

## Troubleshooting

### Common Issues

#### 1. CGO/SQLite Errors

**Problem**: `fatal error: 'sqlite3.h' file not found`

**Solution**:
```bash
# macOS
xcode-select --install

# Linux
sudo apt-get install libsqlite3-dev  # Debian/Ubuntu
sudo yum install sqlite-devel        # RHEL/CentOS

# Verify CGO is enabled
go env CGO_ENABLED  # Should output: 1
```

#### 2. SQLC Command Not Found

**Problem**: `sqlc: command not found`

**Solution**:
```bash
# Install SQLC
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Add to PATH
export PATH="$GOPATH/bin:$PATH"

# Or use full path
~/go/bin/sqlc generate
```

#### 3. Module Errors

**Problem**: `cannot find module providing package`

**Solution**:
```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download
go mod tidy
```

#### 4. Build Errors on Windows

**Problem**: `gcc: command not found` on Windows

**Solution**:
- Install MinGW-w64 or TDM-GCC
- Add to PATH: `C:\MinGW\bin`
- Restart your terminal

#### 5. Permission Denied on Scripts

**Problem**: `permission denied: ./verify-deps.sh`

**Solution**:
```bash
chmod +x verify-deps.sh verify-branding.sh
./verify-deps.sh
```

### Getting Help

If you encounter issues not covered here:

1. Check the [project documentation](/docs)
2. Search [GitHub Issues](https://github.com/AINative-studio/ainative-code/issues)
3. Review the [Quick Start Guide](/QUICK-START.md)
4. Read the [Dependencies Documentation](/DEPENDENCIES.md)
5. Join our [Discord community](#) (if available)
6. Create a new [GitHub Issue](https://github.com/AINative-studio/ainative-code/issues/new)

## Next Steps

After setting up your development environment:

1. Read the [Build Instructions](build.md) for detailed build information
2. Review the [Testing Guide](testing.md) for writing and running tests
3. Check the [Code Style Guidelines](code-style.md) for coding standards
4. Understand the [Git Workflow](git-workflow.md) for contributing
5. Explore the [Debugging Guide](debugging.md) for troubleshooting tips
6. Review the [Architecture Documentation](/docs/architecture) to understand the codebase

## Environment Variables

### Required for Development

```bash
# Go configuration
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export CGO_ENABLED=1

# Optional: For testing with real APIs
export ANTHROPIC_API_KEY="your-key-here"
export OPENAI_API_KEY="your-key-here"
```

### Recommended Development Settings

```bash
# Enable Go module support
export GO111MODULE=on

# Use local module cache
export GOMODCACHE=$GOPATH/pkg/mod

# Increase test timeout for integration tests
export GOTEST_TIMEOUT=10m

# Enable race detector
export GORACE="log_path=/tmp/race halt_on_error=1"
```

## Quick Reference

### Essential Commands

```bash
# Build
make build              # Build binary
make build-all          # Build for all platforms
make clean              # Clean build artifacts

# Development
make run                # Build and run
make install            # Install to $GOPATH/bin

# Testing
make test               # Run tests
make test-coverage      # Test with coverage
make test-integration   # Run integration tests
make test-benchmark     # Run benchmarks

# Code Quality
make fmt                # Format code
make lint               # Run linter
make vet                # Run go vet
make security           # Security scan
make vuln-check         # Check vulnerabilities

# CI/CD Simulation
make ci                 # Run all CI checks
make pre-commit         # Pre-commit checks

# Dependencies
make deps               # Download dependencies
make deps-upgrade       # Upgrade dependencies
make deps-verify        # Verify dependencies

# Docker
make docker-build       # Build Docker image
make docker-run         # Run in Docker

# Information
make help               # Show all commands
make version            # Show version
make info               # Show build info
```

### Project Layout

```
ainative-code/
├── cmd/                    # Application entry points
│   └── ainative-code/     # Main CLI application
├── internal/              # Private application code
│   ├── api/              # API clients
│   ├── auth/             # Authentication
│   ├── config/           # Configuration
│   ├── database/         # Database layer
│   ├── logger/           # Logging
│   └── tui/              # Terminal UI
├── pkg/                   # Public libraries
├── docs/                  # Documentation
│   ├── development/      # Development guides
│   ├── architecture/     # Architecture docs
│   └── api/              # API documentation
├── configs/              # Configuration files
├── scripts/              # Build scripts
├── tests/                # Integration tests
└── .github/              # GitHub workflows
```

## Additional Resources

- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Bubble Tea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [SQLC Documentation](https://docs.sqlc.dev/)
- [golangci-lint Linters](https://golangci-lint.run/usage/linters/)

---

**Next**: [Build Instructions](build.md) | [Testing Guide](testing.md) | [Code Style](code-style.md)
