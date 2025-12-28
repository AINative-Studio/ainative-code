# Error Handling Framework

A comprehensive error handling framework for AINative Code with custom error types, wrapping/unwrapping support, stack traces, and recovery strategies.

## Table of Contents

- [Overview](#overview)
- [Error Types](#error-types)
- [Features](#features)
- [Usage Examples](#usage-examples)
- [Error Recovery](#error-recovery)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)

## Overview

The error handling framework provides:

- **Custom Error Types**: Specialized errors for different domains (Config, Auth, Provider, Tool, Database)
- **Error Wrapping**: Chain errors while preserving context
- **Stack Traces**: Capture call stacks in debug mode
- **User-Friendly Messages**: Separate technical and user-facing error messages
- **Retry Strategies**: Built-in retry mechanisms with exponential backoff
- **Circuit Breaker**: Prevent cascading failures
- **Structured Errors**: JSON serialization for APIs

## Error Types

### ConfigError

Configuration-related errors for invalid, missing, or malformed configuration.

```go
// Missing configuration
err := errors.NewConfigMissingError("api_key")

// Invalid configuration
err := errors.NewConfigInvalidError("timeout", "must be a positive integer")

// Parse error
err := errors.NewConfigParseError("/path/to/config.yaml", cause)

// Validation error
err := errors.NewConfigValidationError("port", "must be between 1 and 65535")
```

### AuthenticationError

Authentication and authorization errors.

```go
// General auth failure
err := errors.NewAuthFailedError("openai", cause)

// Invalid token
err := errors.NewInvalidTokenError("anthropic")

// Expired token (retryable)
err := errors.NewExpiredTokenError("google")

// Permission denied
err := errors.NewPermissionDeniedError("/api/users", "delete")

// Invalid credentials
err := errors.NewInvalidCredentialsError("aws")
```

### ProviderError

Errors related to external AI provider interactions.

```go
// Provider unavailable (retryable)
err := errors.NewProviderUnavailableError("openai", cause)

// Timeout (retryable)
err := errors.NewProviderTimeoutError("anthropic", "claude-3", 30*time.Second)

// Rate limit (retryable with delay)
err := errors.NewProviderRateLimitError("openai", 60*time.Second)

// Invalid response (retryable)
err := errors.NewProviderInvalidResponseError("google", "malformed JSON", cause)

// Provider not found
err := errors.NewProviderNotFoundError("unknown-provider")
```

### ToolExecutionError

Errors during tool execution.

```go
// Tool not found
err := errors.NewToolNotFoundError("git")

// Execution failed
err := errors.NewToolExecutionFailedError("git", 128, output, cause)

// Timeout (retryable)
err := errors.NewToolTimeoutError("terraform", 30*time.Second)

// Invalid input
err := errors.NewToolInvalidInputError("docker", "image", "must not be empty")

// Permission denied
err := errors.NewToolPermissionDeniedError("kubectl", "/etc/kubernetes/config")
```

### DatabaseError

Database-related errors.

```go
// Connection error (retryable)
err := errors.NewDBConnectionError("postgres", cause)

// Query error
err := errors.NewDBQueryError("SELECT", "users", cause)

// Not found
err := errors.NewDBNotFoundError("products", "id=123")

// Duplicate entry
err := errors.NewDBDuplicateError("users", "email", "test@example.com")

// Constraint violation
err := errors.NewDBConstraintError("orders", "fk_user_id", cause)

// Transaction error (retryable)
err := errors.NewDBTransactionError("commit", cause)
```

## Features

### Error Wrapping and Unwrapping

```go
// Wrap an error with additional context
originalErr := errors.New("connection refused")
wrappedErr := errors.Wrap(originalErr, errors.ErrCodeDBConnection, "failed to connect to database")

// Wrap with formatted message
wrappedErr := errors.Wrapf(originalErr, errors.ErrCodeDBConnection, "failed to connect to %s", dbName)

// Unwrap to get the original error
var baseErr *errors.BaseError
if errors.As(wrappedErr, &baseErr) {
    original := baseErr.Unwrap()
}

// Get the root cause
rootErr := errors.RootCause(wrappedErr)
```

### Stack Traces (Debug Mode)

```go
// Enable debug mode to capture stack traces
errors.EnableDebugMode()

// Create an error - stack trace is automatically captured
err := errors.NewConfigInvalidError("api_key", "must not be empty")

// Format with stack trace
formatted := errors.Format(err)
// Output includes file, line, and function information

// Get just the stack trace
stackTrace := err.StackTrace()
```

### User-Friendly Messages

```go
// Errors have both technical and user-facing messages
err := errors.NewConfigMissingError("database_url")

// Technical message (for logs)
technical := err.Error()
// Output: [CONFIG_MISSING] Required configuration 'database_url' is missing

// User-friendly message (for UI)
userMsg := err.UserMessage()
// Output: Configuration error: Required setting 'database_url' is not configured. Please check your configuration file.

// Or use the helper
userMsg := errors.FormatUser(err)
```

### Error Metadata

```go
err := errors.NewProviderTimeoutError("openai", "gpt-4", 30*time.Second)

// Add metadata
err.WithMetadata("request_id", "req-123")
err.WithMetadata("user_id", "user-456")
err.WithMetadata("timestamp", time.Now())

// Retrieve metadata
metadata := err.Metadata()
requestID := metadata["request_id"]
```

### JSON Serialization

```go
err := errors.NewAuthFailedError("openai", cause)

// Convert to JSON (for API responses)
jsonData, _ := errors.ToJSON(err)

// Parse from JSON
deserializedErr := errors.FromJSON(jsonData)
```

## Error Recovery

### Retry with Exponential Backoff

```go
ctx := context.Background()
config := errors.NewRetryConfig()

// Customize retry strategy
config.Strategy = errors.NewExponentialBackoff()
config.OnRetry = func(attempt int, err error) {
    log.Printf("Retry attempt %d: %v", attempt, err)
}

// Execute with retry
err := errors.Retry(ctx, func() error {
    return callExternalAPI()
}, config)
```

### Linear Backoff

```go
// Retry with constant delay
config := errors.NewRetryConfig()
config.Strategy = errors.NewLinearBackoff(1*time.Second, 3)

err := errors.Retry(ctx, func() error {
    return performOperation()
}, config)
```

### Circuit Breaker

```go
// Create circuit breaker (max 3 failures, 30s reset timeout)
cb := errors.NewCircuitBreaker(3, 30*time.Second)

// Execute through circuit breaker
err := cb.Execute(func() error {
    return callUnreliableService()
})

// Check circuit state
if cb.GetState() == errors.StateOpen {
    log.Println("Circuit is open, requests blocked")
}

// Manual reset
cb.Reset()
```

### Fallback Strategies

```go
// Simple fallback
err := errors.Fallback(
    func() error {
        return primaryOperation()
    },
    func() error {
        return fallbackOperation()
    },
)

// Fallback with value
result, err := errors.FallbackWithValue(
    func() (string, error) {
        return fetchFromPrimary()
    },
    "default value",
)
```

## Usage Examples

### Complete Example: API Call with Retry and Circuit Breaker

```go
package main

import (
    "context"
    "time"
    "github.com/ainative/ainative-code/internal/errors"
)

type APIClient struct {
    cb *errors.CircuitBreaker
}

func NewAPIClient() *APIClient {
    return &APIClient{
        cb: errors.NewCircuitBreaker(5, 30*time.Second),
    }
}

func (c *APIClient) CallAPI(ctx context.Context, endpoint string) error {
    config := errors.NewRetryConfig()
    config.Strategy = errors.NewExponentialBackoff()

    config.OnRetry = func(attempt int, err error) {
        log.Printf("Retrying API call to %s (attempt %d): %v", endpoint, attempt, err)
    }

    return errors.Retry(ctx, func() error {
        return c.cb.Execute(func() error {
            resp, err := http.Get(endpoint)
            if err != nil {
                return errors.NewProviderUnavailableError("api", err)
            }
            defer resp.Body.Close()

            if resp.StatusCode == 429 {
                retryAfter := parseRetryAfter(resp.Header)
                return errors.NewProviderRateLimitError("api", retryAfter)
            }

            if resp.StatusCode >= 500 {
                return errors.NewProviderUnavailableError("api",
                    fmt.Errorf("server error: %d", resp.StatusCode))
            }

            return nil
        })
    }, config)
}
```

### Database Operation with Error Handling

```go
func GetUser(ctx context.Context, db *sql.DB, userID string) (*User, error) {
    var user User

    err := db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = $1", userID).Scan(&user)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.NewDBNotFoundError("users", fmt.Sprintf("id=%s", userID))
        }
        return nil, errors.Wrap(err, errors.ErrCodeDBQuery, "failed to fetch user")
    }

    return &user, nil
}

func CreateUser(ctx context.Context, db *sql.DB, user *User) error {
    _, err := db.ExecContext(ctx,
        "INSERT INTO users (id, email, name) VALUES ($1, $2, $3)",
        user.ID, user.Email, user.Name)

    if err != nil {
        // Check for duplicate key error
        if isDuplicateKeyError(err) {
            return errors.NewDBDuplicateError("users", "email", user.Email)
        }
        return errors.Wrap(err, errors.ErrCodeDBQuery, "failed to create user")
    }

    return nil
}
```

### HTTP API Error Response

```go
func errorResponse(w http.ResponseWriter, err error) {
    // Get error code and severity
    code := errors.GetCode(err)
    severity := errors.GetSeverity(err)

    // Determine HTTP status code
    var statusCode int
    switch code {
    case errors.ErrCodeDBNotFound, errors.ErrCodeProviderNotFound:
        statusCode = http.StatusNotFound
    case errors.ErrCodeAuthFailed, errors.ErrCodeAuthInvalidToken:
        statusCode = http.StatusUnauthorized
    case errors.ErrCodeAuthPermissionDenied:
        statusCode = http.StatusForbidden
    case errors.ErrCodeProviderRateLimit:
        statusCode = http.StatusTooManyRequests
    default:
        statusCode = http.StatusInternalServerError
    }

    // Convert to JSON
    jsonData, _ := errors.ToJSON(err)

    // Log technical details
    if severity == errors.SeverityCritical || severity == errors.SeverityHigh {
        log.Printf("Error: %s", errors.Format(err))
    }

    // Send user-friendly response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    w.Write(jsonData)
}
```

## Best Practices

### 1. Use Specific Error Types

```go
// Good: Use specific error constructors
err := errors.NewConfigMissingError("api_key")

// Avoid: Using generic errors
err := fmt.Errorf("api_key is missing")
```

### 2. Wrap Errors at Boundaries

```go
// Wrap errors when crossing layer boundaries
func (s *Service) GetUser(id string) (*User, error) {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, errors.Wrap(err, errors.ErrCodeDBQuery, "service: failed to get user")
    }
    return user, nil
}
```

### 3. Use Debug Mode in Development

```go
// In main.go or initialization
if os.Getenv("ENV") == "development" {
    errors.EnableDebugMode()
}
```

### 4. Always Check Retryability

```go
if errors.IsRetryable(err) {
    // Implement retry logic
} else {
    // Fail immediately
}
```

### 5. Provide User-Friendly Messages

```go
// Always set user-friendly messages for end-user facing errors
err := errors.NewAuthFailedError("provider", cause)
// The constructor already provides a good user message

// Display to user
fmt.Println(err.UserMessage())
```

### 6. Add Context with Metadata

```go
err := errors.NewProviderTimeoutError("openai", "gpt-4", 30*time.Second)
err.WithMetadata("request_id", requestID)
err.WithMetadata("user_id", userID)
err.WithMetadata("endpoint", endpoint)
```

## API Reference

### Error Codes

- **Config**: `ErrCodeConfigInvalid`, `ErrCodeConfigMissing`, `ErrCodeConfigParse`, `ErrCodeConfigValidation`
- **Auth**: `ErrCodeAuthFailed`, `ErrCodeAuthInvalidToken`, `ErrCodeAuthExpiredToken`, `ErrCodeAuthPermissionDenied`, `ErrCodeAuthInvalidCredentials`
- **Provider**: `ErrCodeProviderUnavailable`, `ErrCodeProviderTimeout`, `ErrCodeProviderRateLimit`, `ErrCodeProviderInvalidResponse`, `ErrCodeProviderNotFound`
- **Tool**: `ErrCodeToolNotFound`, `ErrCodeToolExecutionFailed`, `ErrCodeToolTimeout`, `ErrCodeToolInvalidInput`, `ErrCodeToolPermissionDenied`
- **Database**: `ErrCodeDBConnection`, `ErrCodeDBQuery`, `ErrCodeDBNotFound`, `ErrCodeDBDuplicate`, `ErrCodeDBConstraint`, `ErrCodeDBTransaction`

### Severity Levels

- `SeverityLow`: Minor issues, informational
- `SeverityMedium`: Standard errors
- `SeverityHigh`: Serious errors requiring attention
- `SeverityCritical`: Critical failures, system-level issues

### Key Functions

- `Wrap(err, code, message)`: Wrap error with context
- `Wrapf(err, code, format, args...)`: Wrap with formatted message
- `GetCode(err)`: Extract error code
- `IsRetryable(err)`: Check if error is retryable
- `GetSeverity(err)`: Get error severity
- `Format(err)`: Format for logging (includes stack in debug mode)
- `FormatUser(err)`: Get user-friendly message
- `ToJSON(err)`: Serialize to JSON
- `FromJSON(data)`: Deserialize from JSON
- `RootCause(err)`: Get the root cause of wrapped errors

### Retry Strategies

- `NewExponentialBackoff()`: Exponential backoff with sensible defaults
- `NewLinearBackoff(delay, maxRetries)`: Constant delay between retries
- `Retry(ctx, fn, config)`: Execute function with retry logic
- `NewCircuitBreaker(maxFailures, resetTimeout)`: Create circuit breaker
- `Fallback(primary, fallback)`: Execute with fallback

---

**Copyright © 2024 AINative Studio. All rights reserved.**
