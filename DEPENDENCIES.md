# AINative Code - Dependencies Documentation

This document provides a comprehensive overview of all dependencies installed for the AINative Code project.

## Core Dependencies

### UI/TUI Framework
- **github.com/charmbracelet/bubbletea** v1.3.10
  - Purpose: Terminal User Interface (TUI) framework based on The Elm Architecture
  - Used for: Building interactive command-line interfaces
  - Documentation: https://github.com/charmbracelet/bubbletea

- **github.com/charmbracelet/lipgloss** v1.1.0
  - Purpose: Style definitions for terminal rendering
  - Used for: Styling TUI components (colors, borders, layouts)
  - Documentation: https://github.com/charmbracelet/lipgloss

### CLI Framework
- **github.com/spf13/cobra** v1.10.2
  - Purpose: Modern CLI application framework
  - Used for: Command-line interface structure, flags, and subcommands
  - Documentation: https://github.com/spf13/cobra

- **github.com/spf13/pflag** v1.0.10
  - Purpose: Drop-in replacement for Go's flag package (POSIX compliant)
  - Used for: Command-line flag parsing
  - Documentation: https://github.com/spf13/pflag

### Configuration Management
- **github.com/spf13/viper** v1.21.0
  - Purpose: Complete configuration solution for Go applications
  - Used for: Managing application configuration from files, environment variables, and flags
  - Supports: JSON, TOML, YAML, HCL, envfile and Java properties config files
  - Documentation: https://github.com/spf13/viper

- **github.com/spf13/afero** v1.15.0
  - Purpose: Filesystem abstraction for Go
  - Used for: File system operations (dependency of Viper)

- **github.com/fsnotify/fsnotify** v1.9.0
  - Purpose: Cross-platform file system notifications
  - Used for: Watching configuration file changes

### Authentication & Security
- **github.com/golang-jwt/jwt/v5** v5.3.0
  - Purpose: JWT (JSON Web Tokens) implementation
  - Used for: Token-based authentication for Claude API integration
  - Documentation: https://github.com/golang-jwt/jwt

### Database
- **github.com/mattn/go-sqlite3** v1.14.32
  - Purpose: SQLite3 driver for Go's database/sql package
  - Used for: Local database storage (conversation history, settings, cache)
  - Note: Requires CGO to be enabled
  - Documentation: https://github.com/mattn/go-sqlite3

### HTTP Client
- **github.com/go-resty/resty/v2** v2.17.1
  - Purpose: Simple HTTP and REST client library
  - Used for: Making HTTP requests to Claude API and other external services
  - Features: Automatic retry, request/response middleware, debugging
  - Documentation: https://github.com/go-resty/resty

### Logging
- **github.com/rs/zerolog** v1.34.0
  - Purpose: Zero-allocation JSON logger
  - Used for: Structured logging throughout the application
  - Documentation: https://github.com/rs/zerolog

- **gopkg.in/natefinch/lumberjack.v2** v2.2.1
  - Purpose: Log rotation library
  - Used for: Managing log file size and rotation
  - Documentation: https://github.com/natefinch/lumberjack

## Development Tools

### SQLC
- **SQLC** v1.30.0 (installed via `go install`)
  - Purpose: Generate type-safe Go code from SQL
  - Installation location: ~/go/bin/sqlc
  - Used for: Creating type-safe database query functions
  - Configuration: sqlc.yaml
  - Documentation: https://docs.sqlc.dev/

## Supporting Dependencies

### Terminal & Color Support
- **github.com/mattn/go-isatty** v0.0.20
  - Checks if a file descriptor is a terminal

- **github.com/mattn/go-colorable** v0.1.13
  - Makes ANSI colors work on Windows

- **github.com/mattn/go-runewidth** v0.0.16
  - Provides character width calculation for terminal display

- **github.com/lucasb-eyer/go-colorful** v1.2.0
  - Color manipulation library

### Text Processing & Parsing
- **github.com/pelletier/go-toml/v2** v2.2.4
  - TOML parser for configuration files

- **github.com/subosito/gotenv** v1.6.0
  - Environment file (.env) parser

- **github.com/spf13/cast** v1.10.0
  - Safe type conversion utilities

- **github.com/go-viper/mapstructure/v2** v2.4.0
  - Go library for decoding generic map values into native Go structures

### Utility Libraries
- **github.com/sourcegraph/conc** v0.3.1
  - Better structured concurrency for Go

- **golang.org/x/net** v0.43.0
  - Networking extensions (HTTP/2, etc.)

- **golang.org/x/sys** v0.36.0
  - System-level operations

- **golang.org/x/text** v0.28.0
  - Text processing, including encoding and transformation

- **golang.org/x/sync** v0.16.0
  - Additional concurrency primitives

- **golang.org/x/crypto** v0.41.0
  - Cryptographic operations

## Installation Commands

All dependencies were installed using the following commands:

```bash
# Initialize Go module
go mod init github.com/AINative-studio/ainative-code

# Install core dependencies
go get github.com/charmbracelet/bubbletea
go get github.com/spf13/cobra
go get github.com/spf13/viper
go get github.com/golang-jwt/jwt/v5
go get github.com/mattn/go-sqlite3
go get github.com/go-resty/resty/v2

# Install SQLC for type-safe SQL queries
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## Dependency Management

Dependencies are managed using Go modules (go.mod and go.sum files).

### Updating Dependencies
```bash
# Update all dependencies to latest minor/patch versions
go get -u ./...

# Update a specific dependency
go get -u github.com/charmbracelet/bubbletea

# Tidy up go.mod and go.sum
go mod tidy
```

### Verifying Dependencies
```bash
# List all dependencies
go list -m all

# Check for available updates
go list -u -m all

# Verify dependencies haven't been modified
go mod verify
```

## Security Considerations

1. **JWT Tokens**: Store securely, never commit to version control
2. **SQLite Database**: Sensitive data should be encrypted at rest
3. **API Keys**: Use environment variables or secure configuration files
4. **Dependency Updates**: Regularly check for security updates using `go list -u -m all`

## Notes

- All dependencies are compatible with Go 1.25.5
- The project uses indirect dependencies (marked with `// indirect` in go.mod) which are pulled in automatically by direct dependencies
- SQLC is installed as a development tool and is not included in the application binary
- CGO is required for SQLite3 support; ensure a C compiler is available on the build system

## Version Compatibility

This dependency configuration has been tested with:
- Go version: 1.25.5
- Platform: darwin (macOS)
- Date: 2025-12-27
