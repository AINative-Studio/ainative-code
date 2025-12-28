# Code Style Guidelines

This document defines the coding standards and style guidelines for the AINative Code project.

## Table of Contents

- [General Principles](#general-principles)
- [Go Code Style](#go-code-style)
- [Naming Conventions](#naming-conventions)
- [Code Organization](#code-organization)
- [Error Handling](#error-handling)
- [Comments and Documentation](#comments-and-documentation)
- [Testing Standards](#testing-standards)
- [Best Practices](#best-practices)

## General Principles

### 1. Follow Standard Go Conventions

- Adhere to [Effective Go](https://go.dev/doc/effective_go)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for automatic formatting
- Use `goimports` for import organization

### 2. Code Philosophy

- **Simplicity**: Simple code is better than clever code
- **Readability**: Code is read more often than written
- **Consistency**: Follow existing patterns in the codebase
- **Explicitness**: Be explicit rather than implicit
- **Performance**: Optimize for clarity first, performance second

### 3. Code Quality

- Maintain 80%+ test coverage
- Pass all linters (golangci-lint)
- No compiler warnings
- Zero race conditions
- Security-conscious code

## Go Code Style

### Formatting

**Always use `gofmt`** (or `goimports`):

```bash
# Format all files
make fmt

# Check formatting
make fmt-check
```

### Line Length

- Prefer lines under 100 characters
- Break long lines at logical points
- Use multiple lines for function parameters if needed

```go
// Good: Readable line breaks
func CreateUser(
    ctx context.Context,
    email string,
    name string,
    preferences UserPreferences,
) (*User, error) {
    // Implementation
}

// Bad: Too long
func CreateUser(ctx context.Context, email string, name string, preferences UserPreferences) (*User, error) {
    // Implementation
}
```

### Imports

Group imports into three sections:
1. Standard library
2. External packages
3. Internal packages

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // External packages
    "github.com/rs/zerolog"
    "github.com/spf13/viper"

    // Internal packages
    "github.com/AINative-studio/ainative-code/internal/config"
    "github.com/AINative-studio/ainative-code/internal/logger"
)
```

### Package Names

- Use short, lowercase, single-word names
- No underscores or mixedCaps
- Match the directory name

```go
// Good
package logger
package config
package database

// Bad
package loggerUtils
package configHelper
package db_connection
```

## Naming Conventions

### Variables

```go
// Use camelCase
var userName string
var sessionID string  // Acronyms in variable names use consistent case

// Single letter for short-lived variables
for i := 0; i < len(items); i++ {
    // ...
}

// Descriptive names for longer-lived variables
var authenticatedUser *User
var configurationManager *ConfigManager
```

### Constants

```go
// Use MixedCaps (not SCREAMING_SNAKE_CASE)
const MaxRetries = 3
const DefaultTimeout = 30 * time.Second

// Group related constants
const (
    StatusPending  = "pending"
    StatusActive   = "active"
    StatusInactive = "inactive"
)
```

### Functions and Methods

```go
// Use MixedCaps
func GetUser(id string) (*User, error)
func ValidateEmail(email string) bool

// Exported functions start with capital letter
func NewLogger(config *Config) *Logger

// Unexported functions start with lowercase
func parseConfig(data []byte) (*Config, error)
```

### Interfaces

```go
// Single-method interfaces end in -er
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

// Multi-method interfaces use descriptive names
type UserStore interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, user *User) error
    UpdateUser(ctx context.Context, user *User) error
    DeleteUser(ctx context.Context, id string) error
}
```

### Types and Structs

```go
// Use MixedCaps
type UserConfig struct {
    Email    string
    Password string
}

// Acronyms are consistently cased
type HTTPClient struct {
    URL     string
    APIKey  string
}
```

### Receivers

```go
// Use short, consistent receiver names (not "this" or "self")
type User struct {
    Name string
}

// Good: Short, 1-2 letter abbreviation
func (u *User) GetName() string {
    return u.Name
}

// Bad: Too verbose
func (user *User) GetName() string {
    return user.Name
}

// Use pointer receivers for:
// - Modifying the receiver
// - Large structs (performance)
// - Consistency (if some methods need pointers)
func (u *User) SetName(name string) {
    u.Name = name
}
```

## Code Organization

### File Structure

```go
// 1. Package declaration
package logger

// 2. Imports
import (
    "context"
    "fmt"
)

// 3. Constants
const (
    DefaultLevel = InfoLevel
)

// 4. Package-level variables
var (
    globalLogger *Logger
)

// 5. Type definitions
type Logger struct {
    level Level
}

// 6. Constructor functions
func New(config *Config) (*Logger, error) {
    // ...
}

// 7. Methods (grouped by receiver)
func (l *Logger) Info(msg string) {
    // ...
}

func (l *Logger) Error(msg string) {
    // ...
}

// 8. Private helper functions
func parseLevel(level string) (Level, error) {
    // ...
}
```

### Package Organization

```
internal/
├── api/           # API clients and interfaces
├── auth/          # Authentication logic
├── config/        # Configuration management
│   ├── config.go      # Main configuration
│   ├── loader.go      # Config loading
│   └── validator.go   # Config validation
├── database/      # Database layer
├── logger/        # Logging system
└── tui/           # Terminal UI
```

### Grouping Related Code

```go
// Group related struct fields
type Config struct {
    // Server configuration
    Host string
    Port int

    // Database configuration
    DBHost     string
    DBPort     int
    DBName     string

    // Logging configuration
    LogLevel  string
    LogFormat string
}
```

## Error Handling

### Error Creation

```go
// Use errors.New for simple errors
import "errors"

var ErrNotFound = errors.New("user not found")

// Use fmt.Errorf for formatted errors
err := fmt.Errorf("failed to connect to %s: %w", host, originalErr)

// Use custom error types for complex errors
type ValidationError struct {
    Field string
    Value interface{}
    Msg   string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Msg)
}
```

### Error Handling Patterns

```go
// Always check errors
result, err := DoSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Don't ignore errors
// Bad
_ = file.Close()

// Good
if err := file.Close(); err != nil {
    logger.Errorf("failed to close file: %v", err)
}

// Use defer with error checking
defer func() {
    if err := file.Close(); err != nil {
        logger.Errorf("failed to close file: %v", err)
    }
}()
```

### Error Wrapping

```go
// Use %w to wrap errors (Go 1.13+)
if err := validateUser(user); err != nil {
    return fmt.Errorf("user validation failed: %w", err)
}

// Check wrapped errors with errors.Is
if errors.Is(err, ErrNotFound) {
    // Handle not found error
}

// Check error types with errors.As
var validationErr *ValidationError
if errors.As(err, &validationErr) {
    // Handle validation error
}
```

## Comments and Documentation

### Package Documentation

```go
// Package logger provides structured logging with multiple output formats
// and automatic log rotation.
//
// Basic usage:
//
//     logger := logger.New(logger.DefaultConfig())
//     logger.Info("application started")
//
// The logger supports multiple log levels (DEBUG, INFO, WARN, ERROR, FATAL)
// and can output in JSON or text format.
package logger
```

### Function Documentation

```go
// New creates a new logger instance with the provided configuration.
// If config is nil, default configuration is used.
//
// Example:
//
//     logger := logger.New(&Config{
//         Level:  InfoLevel,
//         Format: JSONFormat,
//     })
//
// Returns an error if the configuration is invalid or if the output
// file cannot be created.
func New(config *Config) (*Logger, error) {
    // Implementation
}
```

### Comment Style

```go
// Good: Clear, concise comments
// GetUser retrieves a user by ID from the database.
func GetUser(id string) (*User, error) {
    // ...
}

// Bad: Obvious or redundant comments
// This function gets a user
func GetUser(id string) (*User, error) {
    // ...
}

// Use comments to explain why, not what
// Retry with exponential backoff to handle rate limits
for i := 0; i < maxRetries; i++ {
    // ...
}

// Not: Loop through retries
for i := 0; i < maxRetries; i++ {
    // ...
}
```

### TODO Comments

```go
// TODO(username): Add support for custom retry logic
// FIXME(username): This breaks when input is empty
// NOTE: This is a temporary workaround for issue #123
```

### Exported vs Unexported

```go
// All exported types, functions, and methods must have documentation

// Config holds logger configuration options.
type Config struct {
    Level  LogLevel
    Format OutputFormat
}

// Unexported items should have comments if not obvious
// parseLevel converts a string to LogLevel.
func parseLevel(s string) (LogLevel, error) {
    // ...
}
```

## Testing Standards

### Test File Naming

```go
// Source file: logger.go
// Test file: logger_test.go
// Benchmark file: logger_bench_test.go
// Example file: logger_example_test.go
```

### Test Function Naming

```go
// Test function names should be descriptive
func TestNew(t *testing.T) {}
func TestLogger_Info(t *testing.T) {}
func TestValidateEmail_WithInvalidEmail_ReturnsError(t *testing.T) {}

// Use table-driven tests
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "test@example.com", false},
        {"invalid email", "invalid", true},
        {"empty string", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail(%q) error = %v, wantErr %v",
                    tt.email, err, tt.wantErr)
            }
        })
    }
}
```

### Test Organization

```go
func TestFunction(t *testing.T) {
    // Arrange: Set up test data
    input := "test"
    expected := "result"

    // Act: Execute the function
    result := Function(input)

    // Assert: Verify the result
    if result != expected {
        t.Errorf("Function(%q) = %q; want %q", input, result, expected)
    }
}
```

## Best Practices

### 1. Use Context

```go
// Always pass context as first parameter
func GetUser(ctx context.Context, id string) (*User, error) {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Continue with operation
    return fetchUser(ctx, id)
}
```

### 2. Avoid Global State

```go
// Bad: Global mutable state
var users = make(map[string]*User)

func GetUser(id string) *User {
    return users[id]
}

// Good: Dependency injection
type UserStore struct {
    users map[string]*User
}

func (s *UserStore) GetUser(id string) *User {
    return s.users[id]
}
```

### 3. Use Interfaces for Abstraction

```go
// Define interfaces in consumer packages, not provider packages
type UserGetter interface {
    GetUser(ctx context.Context, id string) (*User, error)
}

func ProcessUser(ctx context.Context, store UserGetter, id string) error {
    user, err := store.GetUser(ctx, id)
    // ...
}
```

### 4. Initialize Structs Explicitly

```go
// Good: Explicit initialization
config := &Config{
    Level:  InfoLevel,
    Format: JSONFormat,
    Output: "stdout",
}

// Acceptable for zero values
var user User

// Bad: Implicit initialization
config := &Config{InfoLevel, JSONFormat, "stdout"}
```

### 5. Use Defer for Cleanup

```go
func ProcessFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()  // Cleanup happens even if errors occur

    // Process file
    return nil
}
```

### 6. Avoid Naked Returns

```go
// Bad: Naked return
func GetUser(id string) (user *User, err error) {
    user, err = fetchUser(id)
    return  // Which variables are being returned?
}

// Good: Explicit return
func GetUser(id string) (*User, error) {
    user, err := fetchUser(id)
    return user, err  // Clear what's being returned
}
```

### 7. Use Meaningful Variable Names

```go
// Bad: Unclear abbreviations
func proc(u *U, cfg *C) error {
    // ...
}

// Good: Clear, descriptive names
func processUser(user *User, config *Config) error {
    // ...
}

// Acceptable: Well-known abbreviations in limited scope
for i, item := range items {
    // i and item are clear in this context
}
```

### 8. Keep Functions Small

```go
// Good: Single responsibility
func ValidateUser(user *User) error {
    if err := validateEmail(user.Email); err != nil {
        return err
    }
    if err := validatePassword(user.Password); err != nil {
        return err
    }
    return nil
}

// Supporting functions
func validateEmail(email string) error { /* ... */ }
func validatePassword(password string) error { /* ... */ }
```

### 9. Use Constants for Magic Values

```go
// Bad: Magic numbers
if len(password) < 8 {
    return errors.New("password too short")
}

// Good: Named constants
const MinPasswordLength = 8

if len(password) < MinPasswordLength {
    return errors.New("password too short")
}
```

### 10. Prefer Composition Over Inheritance

```go
// Good: Composition
type LoggedDatabase struct {
    db     *Database
    logger *Logger
}

func (ld *LoggedDatabase) Query(sql string) (Result, error) {
    ld.logger.Debugf("Executing query: %s", sql)
    return ld.db.Query(sql)
}
```

## Linting Configuration

The project uses golangci-lint with comprehensive checks defined in `.golangci.yml`.

### Run Linters

```bash
# Run all linters
make lint

# Auto-fix issues (where possible)
golangci-lint run --fix

# Run specific linter
golangci-lint run --disable-all --enable=errcheck
```

### Key Linters

- **errcheck**: Check for unchecked errors
- **gosimple**: Simplify code
- **govet**: Vet examines Go source code
- **staticcheck**: Advanced static analysis
- **gosec**: Security checks
- **gofmt**: Code formatting
- **revive**: Fast, configurable linter

### Disable Linter (When Necessary)

```go
// Disable specific linter for a line
//nolint:errcheck
file.Close()

// Disable with reason
//nolint:gosec // This is safe because input is sanitized
hash := md5.Sum(data)

// Disable for entire function
//nolint:gocyclo
func complexFunction() {
    // ...
}
```

## Pre-Commit Checklist

Before committing code:

- [ ] Code is formatted (`make fmt`)
- [ ] All tests pass (`make test`)
- [ ] Linters pass (`make lint`)
- [ ] No race conditions (`go test -race ./...`)
- [ ] Code is documented
- [ ] Commit message is clear and descriptive
- [ ] No debugging code or commented-out code
- [ ] No secrets or sensitive data

```bash
# Run all pre-commit checks
make pre-commit
```

## Tools

### Install Development Tools

```bash
# goimports (better than gofmt)
go install golang.org/x/tools/cmd/goimports@latest

# golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# gosec (security scanner)
go install github.com/securego/gosec/v2/cmd/gosec@latest

# staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest
```

### Editor Integration

**VSCode** (`.vscode/settings.json`):
```json
{
  "go.formatTool": "goimports",
  "editor.formatOnSave": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"]
}
```

**GoLand**: Configure in Preferences > Go > Code Style

## Quick Reference

### Naming

```
Variables:     camelCase (userID, sessionToken)
Constants:     MixedCaps (MaxRetries, DefaultTimeout)
Functions:     MixedCaps (GetUser, CreateSession)
Types:         MixedCaps (UserConfig, HTTPClient)
Packages:      lowercase (logger, config)
Receivers:     Short (u *User, c *Config)
```

### Code Organization

```
1. Package declaration
2. Imports (stdlib, external, internal)
3. Constants
4. Variables
5. Types
6. Constructors
7. Methods
8. Helper functions
```

### Common Commands

```bash
make fmt          # Format code
make fmt-check    # Check formatting
make lint         # Run linters
make test         # Run tests
make pre-commit   # Run all checks
```

---

**Next**: [Git Workflow](git-workflow.md) | [Testing Guide](testing.md) | [Build Instructions](build.md)
