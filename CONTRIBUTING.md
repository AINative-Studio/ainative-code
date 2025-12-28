# Contributing to AINative Code

Thank you for your interest in contributing to AINative Code! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, but recommended)
- golangci-lint (for linting)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/ainative-code.git
   cd ainative-code
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/AINative-studio/ainative-code.git
   ```

## Development Setup

### Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
make install-tools
```

### Build the Project

```bash
# Build the binary
make build

# Build for all platforms
make build-all
```

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run integration tests
make test-integration
```

## Development Workflow

### 1. Create a Feature Branch

Always create a new branch for your work:

```bash
git checkout -b feature/your-feature-name
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions or changes

### 2. Make Your Changes

- Write clean, readable code
- Follow the coding standards (see below)
- Add tests for new functionality
- Update documentation as needed
- Keep commits focused and atomic

### 3. Test Your Changes

Before submitting:

```bash
# Run all tests
make test

# Run linter
make lint

# Check code formatting
make fmt-check

# Verify build works
make build
```

### 4. Commit Your Changes

Follow our commit message guidelines (see below):

```bash
git add .
git commit -m "feat: add new feature"
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## Coding Standards

### Go Code Style

We follow standard Go conventions:

- Use `gofmt` for formatting (run `make fmt`)
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use meaningful variable and function names
- Keep functions small and focused
- Add comments for exported functions and types

### Code Organization

```
ainative-code/
├── cmd/              # Command-line entry points
├── internal/         # Private application code
│   ├── auth/        # Authentication logic
│   ├── llm/         # LLM provider implementations
│   ├── tui/         # Terminal UI components
│   ├── config/      # Configuration management
│   └── logger/      # Logging utilities
├── pkg/             # Public library code
├── docs/            # Documentation
├── tests/           # Integration and E2E tests
└── scripts/         # Build and utility scripts
```

### Error Handling

- Always handle errors explicitly
- Use meaningful error messages
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Return errors instead of panicking (except in truly unrecoverable situations)

Example:
```go
func ReadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
    }
    // ... rest of function
}
```

### Logging

Use the project's logger package:

```go
import "github.com/AINative-studio/ainative-code/internal/logger"

// Simple logging
logger.Info("User authenticated successfully")

// Structured logging
logger.InfoWithFields("Request processed", map[string]interface{}{
    "request_id": reqID,
    "duration_ms": duration.Milliseconds(),
})

// Context-aware logging
ctx := logger.WithRequestID(context.Background(), "req-123")
log := logger.WithContext(ctx)
log.Info("Processing request") // Automatically includes request_id
```

## Testing Guidelines

### Unit Tests

- Write unit tests for all new code
- Aim for >80% code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

Example:
```go
func TestAuthenticate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    bool
        wantErr bool
    }{
        {"valid token", "valid-token", true, false},
        {"invalid token", "invalid", false, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Authenticate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Authenticate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Authenticate() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

- Place integration tests in the `tests/` directory
- Use build tags to separate from unit tests:
  ```go
  //go:build integration
  ```
- Clean up resources after tests

### Benchmarks

Add benchmarks for performance-critical code:

```go
func BenchmarkLogger(b *testing.B) {
    log := logger.New(logger.Config{Level: logger.INFO})
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        log.Info("Benchmark message")
    }
}
```

## Commit Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semicolons, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes
- `ci`: CI/CD changes

### Examples

```
feat(auth): implement JWT token validation

Add support for RS256 JWT token validation with public key caching.
Includes fallback to API validation and offline mode support.

Closes #123
```

```
fix(logger): correct timestamp formatting in JSON output

The timestamp was using incorrect RFC3339 format. Changed to use
ISO8601 format for consistency with other services.
```

```
docs(readme): update installation instructions

Add Homebrew installation method and improve Docker instructions.
```

### Scope

Optional, but recommended. Common scopes:
- `auth`
- `llm`
- `tui`
- `config`
- `logger`
- `zerodb`
- `cli`

## Pull Request Process

### Before Submitting

1. Update documentation if needed
2. Add tests for new functionality
3. Ensure all tests pass
4. Run the linter and fix any issues
5. Update CHANGELOG.md (if applicable)
6. Rebase on latest main branch

### PR Title

Use the same format as commit messages:
```
feat(auth): implement OAuth 2.0 flow
```

### PR Description

Include:
- Summary of changes
- Motivation and context
- Related issues (use "Fixes #123" or "Closes #123")
- Screenshots (for UI changes)
- Testing instructions
- Checklist of completed items

Template:
```markdown
## Summary
Brief description of changes

## Motivation
Why is this change needed?

## Changes
- Change 1
- Change 2

## Testing
How to test these changes

## Related Issues
Fixes #123

## Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Changelog updated
- [ ] All tests passing
- [ ] Linter passing
```

### Review Process

1. Maintainers will review your PR
2. Address any requested changes
3. Once approved, a maintainer will merge your PR
4. Your changes will be included in the next release

## Documentation

### Code Documentation

- Add godoc comments for all exported functions, types, and packages
- Include examples in documentation where helpful
- Keep documentation up to date with code changes

Example:
```go
// AuthenticateUser validates user credentials and returns a JWT token.
// It returns an error if authentication fails or if the credentials are invalid.
//
// Example:
//   token, err := AuthenticateUser("user@example.com", "password")
//   if err != nil {
//       log.Fatal(err)
//   }
func AuthenticateUser(email, password string) (string, error) {
    // Implementation
}
```

### User Documentation

- Update relevant docs in `/docs` directory
- Include code examples and use cases
- Keep README.md up to date
- Add examples to `/docs/examples` for new features

## Community

### Getting Help

- GitHub Discussions for questions
- GitHub Issues for bug reports and feature requests
- Email: support@ainative.studio

### Communication Guidelines

- Be respectful and inclusive
- Provide constructive feedback
- Help others when you can
- Stay on topic in discussions

### Recognition

Contributors are recognized in:
- Release notes
- Contributors section (planned)
- Special recognition for significant contributions

## Questions?

If you have questions about contributing, please:
1. Check existing documentation
2. Search GitHub Issues and Discussions
3. Create a new Discussion if needed
4. Reach out to maintainers

Thank you for contributing to AINative Code!

---

**Copyright 2024 AINative Studio. All rights reserved.**
