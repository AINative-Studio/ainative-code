# Errors API Reference

**Import Path**: `github.com/AINative-studio/ainative-code/internal/errors`

The errors package provides a comprehensive error handling framework with custom error types, error codes, severity levels, stack traces, and recovery strategies.

## Table of Contents

- [Error Types](#error-types)
- [Error Codes](#error-codes)
- [Severity Levels](#severity-levels)
- [Error Functions](#error-functions)
- [Provider Errors](#provider-errors)
- [Authentication Errors](#authentication-errors)
- [Database Errors](#database-errors)
- [Tool Errors](#tool-errors)
- [Error Recovery](#error-recovery)
- [Usage Examples](#usage-examples)

## Error Types

### BaseError

```go
type BaseError struct {
    // Contains filtered or unexported fields
}
```

The foundation for all custom errors in the system.

**Methods**:

```go
func (e *BaseError) Error() string
func (e *BaseError) Unwrap() error
func (e *BaseError) Code() ErrorCode
func (e *BaseError) UserMessage() string
func (e *BaseError) Severity() Severity
func (e *BaseError) Stack() []StackFrame
func (e *BaseError) IsRetryable() bool
func (e *BaseError) Metadata() map[string]interface{}
func (e *BaseError) WithMetadata(key string, value interface{}) *BaseError
func (e *BaseError) StackTrace() string
```

**Example**:

```go
if baseErr, ok := err.(*errors.BaseError); ok {
    fmt.Printf("Error code: %s\n", baseErr.Code())
    fmt.Printf("Severity: %s\n", baseErr.Severity())
    fmt.Printf("User message: %s\n", baseErr.UserMessage())
    fmt.Printf("Retryable: %v\n", baseErr.IsRetryable())

    if baseErr.Severity() == errors.SeverityCritical {
        log.Fatal(baseErr.StackTrace())
    }
}
```

## Error Codes

```go
type ErrorCode string

const (
    // Configuration errors
    ErrCodeConfigInvalid    ErrorCode = "CONFIG_INVALID"
    ErrCodeConfigMissing    ErrorCode = "CONFIG_MISSING"
    ErrCodeConfigParse      ErrorCode = "CONFIG_PARSE"
    ErrCodeConfigValidation ErrorCode = "CONFIG_VALIDATION"

    // Authentication errors
    ErrCodeAuthFailed             ErrorCode = "AUTH_FAILED"
    ErrCodeAuthInvalidToken       ErrorCode = "AUTH_INVALID_TOKEN"
    ErrCodeAuthExpiredToken       ErrorCode = "AUTH_EXPIRED_TOKEN"
    ErrCodeAuthPermissionDenied   ErrorCode = "AUTH_PERMISSION_DENIED"
    ErrCodeAuthInvalidCredentials ErrorCode = "AUTH_INVALID_CREDENTIALS"

    // Provider errors
    ErrCodeProviderUnavailable    ErrorCode = "PROVIDER_UNAVAILABLE"
    ErrCodeProviderTimeout        ErrorCode = "PROVIDER_TIMEOUT"
    ErrCodeProviderRateLimit      ErrorCode = "PROVIDER_RATE_LIMIT"
    ErrCodeProviderInvalidResponse ErrorCode = "PROVIDER_INVALID_RESPONSE"
    ErrCodeProviderNotFound       ErrorCode = "PROVIDER_NOT_FOUND"

    // Tool errors
    ErrCodeToolNotFound         ErrorCode = "TOOL_NOT_FOUND"
    ErrCodeToolExecutionFailed  ErrorCode = "TOOL_EXECUTION_FAILED"
    ErrCodeToolTimeout          ErrorCode = "TOOL_TIMEOUT"
    ErrCodeToolInvalidInput     ErrorCode = "TOOL_INVALID_INPUT"
    ErrCodeToolPermissionDenied ErrorCode = "TOOL_PERMISSION_DENIED"

    // Database errors
    ErrCodeDBConnection   ErrorCode = "DB_CONNECTION_FAILED"
    ErrCodeDBQuery        ErrorCode = "DB_QUERY_FAILED"
    ErrCodeDBNotFound     ErrorCode = "DB_NOT_FOUND"
    ErrCodeDBDuplicate    ErrorCode = "DB_DUPLICATE"
    ErrCodeDBConstraint   ErrorCode = "DB_CONSTRAINT_VIOLATION"
    ErrCodeDBTransaction  ErrorCode = "DB_TRANSACTION_FAILED"

    // Security errors
    ErrCodeSecurityViolation  ErrorCode = "SECURITY_VIOLATION"
    ErrCodeSecurityInvalidKey ErrorCode = "SECURITY_INVALID_KEY"
    ErrCodeSecurityEncryption ErrorCode = "SECURITY_ENCRYPTION_FAILED"
    ErrCodeSecurityDecryption ErrorCode = "SECURITY_DECRYPTION_FAILED"
)
```

## Severity Levels

```go
type Severity string

const (
    SeverityLow      Severity = "low"       // Informational, no action required
    SeverityMedium   Severity = "medium"    // Warning, may require attention
    SeverityHigh     Severity = "high"      // Error, requires action
    SeverityCritical Severity = "critical"  // Critical, immediate action required
)
```

## Error Functions

### Wrap

```go
func Wrap(err error, code ErrorCode, message string) error
```

Wraps an existing error with additional context.

**Example**:

```go
file, err := os.Open("config.yaml")
if err != nil {
    return errors.Wrap(err, errors.ErrCodeConfigParse, "failed to open config file")
}
```

### Wrapf

```go
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) error
```

Wraps an error with a formatted message.

**Example**:

```go
if err != nil {
    return errors.Wrapf(err, errors.ErrCodeDBQuery,
        "failed to query user with ID %s", userID)
}
```

### Is

```go
func Is(err, target error) bool
```

Reports whether any error in err's chain matches target.

**Example**:

```go
if errors.Is(err, sql.ErrNoRows) {
    return errors.NewDBNotFoundError("user", userID)
}
```

### As

```go
func As(err error, target interface{}) bool
```

Finds the first error in err's chain that matches target.

**Example**:

```go
var providerErr *errors.ProviderError
if errors.As(err, &providerErr) {
    log.Printf("Provider: %s, Status: %d", providerErr.ProviderName, providerErr.StatusCode)
}
```

### GetCode

```go
func GetCode(err error) ErrorCode
```

Extracts the error code from an error.

**Example**:

```go
code := errors.GetCode(err)
if code == errors.ErrCodeProviderRateLimit {
    time.Sleep(30 * time.Second)
    // Retry
}
```

### IsRetryable

```go
func IsRetryable(err error) bool
```

Checks if an error is retryable.

**Example**:

```go
if errors.IsRetryable(err) {
    for attempt := 0; attempt < 3; attempt++ {
        time.Sleep(calculateBackoff(attempt))
        err = operation()
        if err == nil {
            break
        }
    }
}
```

### GetSeverity

```go
func GetSeverity(err error) Severity
```

Extracts the severity from an error.

**Example**:

```go
severity := errors.GetSeverity(err)
if severity == errors.SeverityCritical {
    log.Fatal(err)
} else if severity == errors.SeverityHigh {
    log.Error(err)
} else {
    log.Warn(err)
}
```

## Provider Errors

### ProviderError

```go
type ProviderError struct {
    *BaseError
    ProviderName string
    Model        string
    RequestID    string
    StatusCode   int
    RetryAfter   *time.Duration
}
```

**Constructors**:

```go
func NewProviderUnavailableError(providerName string, cause error) *ProviderError
func NewProviderTimeoutError(providerName, model string, timeout time.Duration) *ProviderError
func NewProviderRateLimitError(providerName string, retryAfter time.Duration) *ProviderError
func NewProviderInvalidResponseError(providerName string, reason string, cause error) *ProviderError
func NewProviderNotFoundError(providerName string) *ProviderError
```

**Methods**:

```go
func (e *ProviderError) WithModel(model string) *ProviderError
func (e *ProviderError) WithRequestID(requestID string) *ProviderError
func (e *ProviderError) WithStatusCode(statusCode int) *ProviderError
func (e *ProviderError) ShouldRetry() bool
func (e *ProviderError) GetRetryDelay() time.Duration
```

**Example**:

```go
response, err := provider.Chat(ctx, messages)
if err != nil {
    if providerErr, ok := err.(*errors.ProviderError); ok {
        if providerErr.ShouldRetry() {
            delay := providerErr.GetRetryDelay()
            if delay > 0 {
                time.Sleep(delay)
            }
            // Retry
        }
    }
}
```

## Authentication Errors

### AuthenticationError

```go
type AuthenticationError struct {
    *BaseError
    Provider string
    UserID   string
    Resource string
}
```

**Constructors**:

```go
func NewAuthFailedError(provider string, cause error) *AuthenticationError
func NewInvalidTokenError(provider string) *AuthenticationError
func NewExpiredTokenError(provider string) *AuthenticationError
func NewPermissionDeniedError(resource, action string) *AuthenticationError
func NewInvalidCredentialsError(provider string) *AuthenticationError
```

**Methods**:

```go
func (e *AuthenticationError) WithProvider(provider string) *AuthenticationError
func (e *AuthenticationError) WithUserID(userID string) *AuthenticationError
func (e *AuthenticationError) WithResource(resource string) *AuthenticationError
```

**Example**:

```go
tokens, err := authClient.Authenticate(ctx)
if err != nil {
    if authErr, ok := err.(*errors.AuthenticationError); ok {
        fmt.Printf("Auth failed for provider: %s\n", authErr.Provider)
        fmt.Printf("User message: %s\n", authErr.UserMessage())
    }
}
```

## Database Errors

**Constructors**:

```go
func NewDBConnectionError(dsn string, cause error) error
func NewDBQueryError(query string, cause error) error
func NewDBNotFoundError(resource string, id string) error
func NewDBDuplicateError(resource string, constraint string) error
func NewDBConstraintError(constraint string, cause error) error
func NewDBTransactionError(operation string, cause error) error
```

**Example**:

```go
user, err := db.GetUser(ctx, userID)
if err != nil {
    if errors.GetCode(err) == errors.ErrCodeDBNotFound {
        return errors.NewDBNotFoundError("user", userID)
    }
    return errors.Wrap(err, errors.ErrCodeDBQuery, "failed to get user")
}
```

## Tool Errors

**Constructors**:

```go
func NewToolNotFoundError(toolName string) error
func NewToolExecutionError(toolName string, cause error) error
func NewToolTimeoutError(toolName string, timeout time.Duration) error
func NewToolInvalidInputError(toolName string, field string, reason string) error
func NewToolPermissionDeniedError(toolName string, reason string) error
```

**Example**:

```go
result, err := toolRegistry.Execute(ctx, "file_read", input, execCtx)
if err != nil {
    code := errors.GetCode(err)
    switch code {
    case errors.ErrCodeToolNotFound:
        log.Printf("Tool not found")
    case errors.ErrCodeToolTimeout:
        log.Printf("Tool execution timed out")
    case errors.ErrCodeToolPermissionDenied:
        log.Printf("Permission denied")
    default:
        log.Printf("Tool execution failed: %v", err)
    }
}
```

## Error Recovery

### Recovery Strategies

```go
// Automatic retry with exponential backoff
func retryWithBackoff(operation func() error, maxRetries int) error {
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }

        lastErr = err

        // Check if retryable
        if !errors.IsRetryable(err) {
            return err
        }

        // Calculate backoff
        backoff := time.Duration(1<<uint(attempt)) * time.Second
        log.Printf("Attempt %d failed, retrying in %v: %v", attempt+1, backoff, err)
        time.Sleep(backoff)
    }

    return fmt.Errorf("operation failed after %d attempts: %w", maxRetries, lastErr)
}
```

### Circuit Breaker

```go
type CircuitBreaker struct {
    maxFailures  int
    resetTimeout time.Duration
    failures     int
    lastFailTime time.Time
    state        string // "closed", "open", "half-open"
}

func (cb *CircuitBreaker) Execute(operation func() error) error {
    if cb.state == "open" {
        if time.Since(cb.lastFailTime) > cb.resetTimeout {
            cb.state = "half-open"
        } else {
            return fmt.Errorf("circuit breaker is open")
        }
    }

    err := operation()
    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()

        if cb.failures >= cb.maxFailures {
            cb.state = "open"
        }

        return err
    }

    // Success
    cb.failures = 0
    cb.state = "closed"
    return nil
}
```

## Usage Examples

### Basic Error Handling

```go
package main

import (
    "log"
    "github.com/AINative-studio/ainative-code/internal/errors"
)

func main() {
    err := doSomething()
    if err != nil {
        // Get error code
        code := errors.GetCode(err)
        log.Printf("Error code: %s\n", code)

        // Get severity
        severity := errors.GetSeverity(err)
        if severity == errors.SeverityCritical {
            log.Fatal(err)
        }

        // Check if retryable
        if errors.IsRetryable(err) {
            log.Println("Error is retryable, will retry...")
        }

        // Get user-friendly message
        if baseErr, ok := err.(*errors.BaseError); ok {
            log.Printf("User message: %s\n", baseErr.UserMessage())
        }
    }
}
```

### Error Categorization

```go
func handleError(err error) {
    code := errors.GetCode(err)

    switch code {
    case errors.ErrCodeProviderRateLimit:
        handleRateLimit(err)
    case errors.ErrCodeProviderTimeout:
        handleTimeout(err)
    case errors.ErrCodeAuthExpiredToken:
        handleExpiredToken(err)
    case errors.ErrCodeDBNotFound:
        handleNotFound(err)
    default:
        handleGenericError(err)
    }
}
```

### Retryable Operations

```go
func executeWithRetry(operation func() error) error {
    maxRetries := 3
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }

        lastErr = err

        // Check if we should retry
        if !errors.IsRetryable(err) {
            return err
        }

        // Provider-specific retry delay
        if providerErr, ok := err.(*errors.ProviderError); ok {
            delay := providerErr.GetRetryDelay()
            if delay > 0 {
                log.Printf("Rate limited, waiting %v", delay)
                time.Sleep(delay)
                continue
            }
        }

        // Exponential backoff
        backoff := time.Duration(1<<uint(attempt)) * time.Second
        log.Printf("Attempt %d failed, retrying in %v", attempt+1, backoff)
        time.Sleep(backoff)
    }

    return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

### Error Logging

```go
func logError(err error) {
    if baseErr, ok := err.(*errors.BaseError); ok {
        log.Printf("[%s] %s (severity: %s, retryable: %v)",
            baseErr.Code(),
            baseErr.Error(),
            baseErr.Severity(),
            baseErr.IsRetryable(),
        )

        // Log metadata
        for key, value := range baseErr.Metadata() {
            log.Printf("  %s: %v", key, value)
        }

        // Log stack trace for high severity
        if baseErr.Severity() == errors.SeverityHigh ||
           baseErr.Severity() == errors.SeverityCritical {
            log.Printf("Stack trace:\n%s", baseErr.StackTrace())
        }
    } else {
        log.Printf("Error: %v", err)
    }
}
```

### Custom Error Creation

```go
func validateInput(input string) error {
    if input == "" {
        return errors.NewError(
            errors.ErrCodeToolInvalidInput,
            "input cannot be empty",
            errors.SeverityMedium,
            false, // not retryable
        ).WithMetadata("field", "input")
    }

    if len(input) > 1000 {
        return errors.NewError(
            errors.ErrCodeToolInvalidInput,
            "input exceeds maximum length",
            errors.SeverityMedium,
            false,
        ).WithMetadata("field", "input").
          WithMetadata("max_length", 1000).
          WithMetadata("actual_length", len(input))
    }

    return nil
}
```

## Best Practices

### 1. Always Use Error Codes

```go
// Good
return errors.Wrap(err, errors.ErrCodeProviderTimeout, "request timed out")

// Bad
return fmt.Errorf("request timed out: %w", err)
```

### 2. Provide User-Friendly Messages

```go
err := errors.NewProviderTimeoutError("anthropic", "claude-3-sonnet", 30*time.Second)
// err.UserMessage() returns user-friendly message
// err.Error() returns technical message
```

### 3. Add Metadata for Debugging

```go
return errors.NewDBQueryError(query, err).
    WithMetadata("user_id", userID).
    WithMetadata("timestamp", time.Now()).
    WithMetadata("query_params", params)
```

### 4. Handle Errors at the Right Level

```go
// Low level: wrap with context
func getUser(id string) (*User, error) {
    user, err := db.Query(...)
    if err != nil {
        return nil, errors.Wrap(err, errors.ErrCodeDBQuery, "failed to get user")
    }
    return user, nil
}

// High level: check and handle
func handleRequest(id string) {
    user, err := getUser(id)
    if err != nil {
        code := errors.GetCode(err)
        if code == errors.ErrCodeDBNotFound {
            http.Error(w, "User not found", 404)
        } else {
            http.Error(w, "Internal error", 500)
        }
    }
}
```

### 5. Use Severity Appropriately

```go
severity := errors.GetSeverity(err)
switch severity {
case errors.SeverityLow:
    log.Info(err)
case errors.SeverityMedium:
    log.Warn(err)
case errors.SeverityHigh:
    log.Error(err)
case errors.SeverityCritical:
    log.Fatal(err)
}
```

## Related Documentation

- [Core Packages](core-packages.md) - Using errors in core packages
- [Providers](providers.md) - Provider error handling
- [Authentication](authentication.md) - Auth error handling
- [Tools](tools.md) - Tool error handling
