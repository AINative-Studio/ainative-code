# Development Documentation

Welcome to the AINative Code development documentation. This guide will help you set up your development environment, understand our workflows, and contribute effectively to the project.

## Quick Start

New to the project? Start here:

1. **[Setup Guide](setup.md)** - Set up your development environment
2. **[Build Instructions](build.md)** - Learn how to build the project
3. **[Testing Guide](testing.md)** - Understand our testing practices
4. **[Code Style](code-style.md)** - Follow our coding standards
5. **[Git Workflow](git-workflow.md)** - Learn our version control process

## Documentation Overview

### [Development Setup](setup.md)

Complete guide to setting up your development environment:
- Prerequisites and system requirements
- Installation instructions for all platforms
- IDE configuration
- Development tools setup
- Verification and troubleshooting

**Start here if**: You're setting up the project for the first time

### [Build Instructions](build.md)

Comprehensive build documentation:
- Quick build commands
- Cross-platform building
- Docker builds
- Release builds
- Build optimization techniques
- Troubleshooting build issues

**Start here if**: You need to build the application or create releases

### [Testing Guide](testing.md)

Testing practices and examples:
- Unit testing patterns
- Integration testing
- Test coverage requirements
- Benchmarking
- Mocking strategies
- Testing best practices

**Start here if**: You're writing or running tests

### [Debugging Guide](debugging.md)

Debugging tools and techniques:
- Using Delve debugger
- Logging for debugging
- Common issues and solutions
- Performance debugging
- Memory debugging
- IDE-specific debugging

**Start here if**: You're troubleshooting issues or debugging code

### [Code Style Guidelines](code-style.md)

Coding standards and conventions:
- Go code style
- Naming conventions
- Code organization
- Error handling
- Documentation standards
- Best practices

**Start here if**: You're writing or reviewing code

### [Git Workflow](git-workflow.md)

Version control and collaboration:
- Branch strategy
- Commit guidelines
- Pull request process
- Release process
- Common Git tasks
- Best practices

**Start here if**: You're contributing code or managing branches

## Development Workflow

### Typical Development Cycle

1. **Set Up Environment**
   ```bash
   # Clone repository
   git clone https://github.com/AINative-studio/ainative-code.git
   cd ainative-code

   # Install dependencies
   make deps

   # Verify setup
   ./verify-deps.sh
   ```

2. **Create Feature Branch**
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/123-my-feature
   ```

3. **Develop and Test**
   ```bash
   # Write code
   # Write tests
   make test

   # Check code quality
   make fmt
   make lint
   ```

4. **Commit and Push**
   ```bash
   git add .
   git commit -m "feat(scope): add new feature"
   git push origin feature/123-my-feature
   ```

5. **Create Pull Request**
   - Open PR on GitHub
   - Wait for CI checks
   - Address review comments
   - Merge when approved

### Essential Commands

```bash
# Build
make build              # Build for current platform
make build-all          # Build for all platforms

# Testing
make test               # Run all tests
make test-coverage      # Run tests with coverage
make test-integration   # Run integration tests

# Code Quality
make fmt                # Format code
make lint               # Run linters
make vet                # Run go vet

# Development
make run                # Build and run
make clean              # Clean build artifacts

# CI Simulation
make ci                 # Run all CI checks
make pre-commit         # Pre-commit checks
```

## Project Structure

```
ainative-code/
├── cmd/                    # Application entry points
│   └── ainative-code/     # Main CLI application
├── internal/              # Private application code
│   ├── api/              # API clients
│   ├── auth/             # Authentication
│   ├── branding/         # Brand constants
│   ├── cmd/              # Command implementations
│   ├── config/           # Configuration management
│   ├── database/         # Database layer
│   ├── errors/           # Error handling
│   └── logger/           # Logging system
├── pkg/                   # Public libraries
├── docs/                  # Documentation
│   ├── development/      # This directory
│   ├── architecture/     # Architecture docs
│   ├── api/              # API documentation
│   └── user-guide/       # User documentation
├── configs/              # Configuration files
├── scripts/              # Build and utility scripts
├── tests/                # Integration tests
├── .github/              # GitHub workflows
│   └── workflows/        # CI/CD pipelines
├── go.mod                # Go module definition
├── go.sum                # Dependency checksums
├── Makefile              # Build automation
├── Dockerfile            # Container definition
└── README.md             # Project overview
```

## Key Technologies

### Core Technologies
- **Go 1.21+**: Primary programming language
- **SQLite**: Local database storage
- **SQLC**: Type-safe SQL code generation

### UI/CLI Frameworks
- **Bubble Tea**: Terminal UI framework
- **Cobra**: CLI command framework
- **Viper**: Configuration management
- **Lipgloss**: Terminal styling

### Development Tools
- **golangci-lint**: Comprehensive linting
- **Delve**: Go debugger
- **gosec**: Security scanning
- **govulncheck**: Vulnerability checking

### Libraries
- **zerolog**: High-performance logging
- **lumberjack**: Log rotation
- **resty**: HTTP client
- **jwt-go**: JWT authentication

## Development Standards

### Quality Requirements
- ✅ 80%+ test coverage
- ✅ All tests passing
- ✅ No linter errors
- ✅ No race conditions
- ✅ Code documented
- ✅ Security reviewed

### Code Review Criteria
- Follows code style guidelines
- Includes tests
- Updates documentation
- No breaking changes (or documented)
- Passes all CI checks
- Has clear commit messages

### CI/CD Pipeline

The project uses GitHub Actions for continuous integration:

**On Push/PR**:
1. Code formatting check
2. Linting (golangci-lint)
3. Unit tests (with race detection)
4. Integration tests
5. Code coverage check (80% minimum)
6. Security scan (gosec)
7. Vulnerability check (govulncheck)
8. Build verification (all platforms)

**On Release Tag**:
1. Build for all platforms
2. Run comprehensive tests
3. Create release packages
4. Generate checksums
5. Create GitHub release
6. Publish Docker images

## Getting Help

### Documentation Resources
- [Main README](../../README.md) - Project overview
- [Quick Start Guide](../../QUICK-START.md) - Get started quickly
- [Dependencies](../../DEPENDENCIES.md) - Dependency information
- [PRD](../../PRD.md) - Product requirements
- [Architecture Docs](../architecture/) - System architecture
- [User Guide](../user-guide/) - End-user documentation

### Communication Channels
- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and general discussion
- **Pull Requests**: Code contributions and reviews

### Common Questions

**Q: How do I run the application locally?**
```bash
make build
./build/ainative-code --help
```

**Q: How do I run tests?**
```bash
make test
```

**Q: How do I check my code before committing?**
```bash
make pre-commit
```

**Q: Where should I put new code?**
- Commands: `internal/cmd/`
- Business logic: `internal/`
- Public APIs: `pkg/`
- Tests: `*_test.go` alongside source files

**Q: How do I add a new dependency?**
```bash
go get github.com/package/name
go mod tidy
```

**Q: How do I update dependencies?**
```bash
make deps-upgrade
```

## Contributing

We welcome contributions! Here's how to get started:

1. **Fork the repository** on GitHub
2. **Clone your fork** locally
3. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
4. **Make your changes** following our guidelines
5. **Write or update tests**
6. **Update documentation**
7. **Run all checks** (`make pre-commit`)
8. **Commit your changes** with clear messages
9. **Push to your fork**
10. **Open a Pull Request**

See [Git Workflow](git-workflow.md) for detailed contribution guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.

## Additional Resources

### External Documentation
- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Bubble Tea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [SQLC Documentation](https://docs.sqlc.dev/)

### Tools Documentation
- [golangci-lint](https://golangci-lint.run/)
- [Delve Debugger](https://github.com/go-delve/delve)
- [Docker](https://docs.docker.com/)
- [GitHub Actions](https://docs.github.com/en/actions)

---

**Need help?** Check the documentation above or open an issue on GitHub.

**Ready to contribute?** Start with the [Setup Guide](setup.md) and [Git Workflow](git-workflow.md).
