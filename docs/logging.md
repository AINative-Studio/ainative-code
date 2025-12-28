# Logging System Documentation

## Overview

The AINative Code logging system provides structured logging with configurable log levels, output formats, rotation, and context-aware capabilities. It is built on top of [zerolog](https://github.com/rs/zerolog), a zero-allocation JSON logger, ensuring high performance and minimal overhead.

## Features

- **Structured Logging**: JSON and text output formats
- **Log Levels**: DEBUG, INFO, WARN, ERROR, FATAL
- **Context-Aware**: Automatic inclusion of request IDs, session IDs, and user IDs
- **Log Rotation**: Automatic rotation based on file size, age, and backup count
- **Performance Optimized**: Zero-allocation logging with minimal overhead
- **Thread-Safe**: Concurrent-safe global logger with mutex protection
- **Flexible Configuration**: YAML-based or programmatic configuration

## Quick Start

### Basic Usage

```go
package main

import (
    "github.com/AINative-studio/ainative-code/internal/logger"
)

func main() {
    // Use global logger with default configuration
    logger.Info("Application started")
    logger.Infof("Server listening on port %d", 8080)

    // Structured logging
    logger.InfoWithFields("User logged in", map[string]interface{}{
        "user_id": "user123",
        "email": "user@example.com",
        "ip": "192.168.1.1",
    })
}
```

### Creating a Custom Logger

```go
package main

import (
    "github.com/AINative-studio/ainative-code/internal/logger"
)

func main() {
    // Create custom logger configuration
    config := &logger.Config{
        Level:          logger.InfoLevel,
        Format:         logger.JSONFormat,
        Output:         "/var/log/ainative-code/app.log",
        EnableRotation: true,
        MaxSize:        100,  // 100 MB
        MaxBackups:     5,
        MaxAge:         30,   // 30 days
        Compress:       true,
        EnableCaller:   true,
        EnableStackTrace: true,
    }

    log, err := logger.New(config)
    if err != nil {
        panic(err)
    }

    // Set as global logger
    logger.SetGlobalLogger(log)

    log.Info("Custom logger initialized")
}
```

## Configuration

### Config Structure

```go
type Config struct {
    // Level sets the minimum log level that will be output
    Level LogLevel  // "debug", "info", "warn", "error"

    // Format specifies the output format
    Format OutputFormat  // "json" or "text"

    // Output specifies where logs should be written
    // Can be "stdout", "stderr", or a file path
    Output string

    // EnableRotation enables log file rotation
    EnableRotation bool

    // MaxSize is the maximum size in megabytes before rotation
    MaxSize int

    // MaxBackups is the maximum number of old log files to retain
    MaxBackups int

    // MaxAge is the maximum number of days to retain old log files
    MaxAge int

    // Compress determines if rotated files should be gzipped
    Compress bool

    // EnableCaller adds file and line number to logs
    EnableCaller bool

    // EnableStackTrace adds stack traces for error level logs
    EnableStackTrace bool
}
```

### Default Configuration

```go
config := logger.DefaultConfig()
// Level: InfoLevel
// Format: TextFormat
// Output: "stdout"
// EnableRotation: false
// MaxSize: 100 MB
// MaxBackups: 3
// MaxAge: 28 days
// Compress: true
// EnableCaller: false
// EnableStackTrace: false
```

## Log Levels

### Available Levels

- **DEBUG**: Detailed information for debugging (usually disabled in production)
- **INFO**: General informational messages (default level)
- **WARN**: Warning messages for potentially harmful situations
- **ERROR**: Error messages for serious problems
- **FATAL**: Critical errors that cause the application to exit

### Using Log Levels

```go
// Debug level
logger.Debug("Debugging information")
logger.Debugf("Variable value: %v", value)
logger.DebugWithFields("Debug event", map[string]interface{}{
    "component": "auth",
    "step": "validation",
})

// Info level
logger.Info("User action completed")
logger.Infof("Processing %d items", count)

// Warn level
logger.Warn("Deprecated API usage")
logger.Warnf("High memory usage: %d MB", memoryMB)

// Error level
logger.Error("Failed to process request")
logger.Errorf("Database error: %v", err)
logger.ErrorWithErr("Query failed", err)

// Fatal level (exits the application)
logger.Fatal("Critical system failure")
logger.Fatalf("Cannot connect to database: %v", err)
```

## Output Formats

### JSON Format

JSON format is ideal for production environments and log aggregation systems.

```json
{
  "level": "info",
  "time": "2025-12-27T10:30:45Z",
  "message": "User logged in",
  "user_id": "user123",
  "session_id": "sess456"
}
```

```go
config := &logger.Config{
    Format: logger.JSONFormat,
    Output: "/var/log/app.log",
}
```

### Text Format

Text format is human-readable and ideal for development and console output.

```
2025-12-27T10:30:45Z INF User logged in user_id=user123 session_id=sess456
```

```go
config := &logger.Config{
    Format: logger.TextFormat,
    Output: "stdout",
}
```

## Context-Aware Logging

### Adding Context IDs

The logging system supports automatic extraction of request IDs, session IDs, and user IDs from Go context.

```go
package main

import (
    "context"
    "github.com/AINative-studio/ainative-code/internal/logger"
)

func handleRequest(ctx context.Context) {
    // Add IDs to context
    ctx = logger.WithRequestID(ctx, "req-abc123")
    ctx = logger.WithSessionID(ctx, "sess-xyz789")
    ctx = logger.WithUserID(ctx, "user-456")

    // Create context-aware logger
    log := logger.WithContext(ctx)

    // All logs will automatically include the IDs
    log.Info("Processing request")
    // Output: {"level":"info","time":"...","request_id":"req-abc123","session_id":"sess-xyz789","user_id":"user-456","message":"Processing request"}
}
```

### Retrieving Context IDs

```go
requestID, ok := logger.GetRequestID(ctx)
if ok {
    fmt.Printf("Request ID: %s\n", requestID)
}

sessionID, ok := logger.GetSessionID(ctx)
userID, ok := logger.GetUserID(ctx)
```

## Structured Logging

### Basic Structured Logging

```go
logger.InfoWithFields("Database query executed", map[string]interface{}{
    "query": "SELECT * FROM users",
    "duration_ms": 45,
    "rows_returned": 150,
})
```

### Advanced Structured Logging

```go
// Using the zerolog context directly for more control
log := logger.GetGlobalLogger()
log.With().
    Str("component", "api").
    Str("method", "POST").
    Str("path", "/api/users").
    Int("status_code", 201).
    Dur("duration", duration).
    Logger().
    Info("API request completed")
```

## Log Rotation

Log rotation prevents log files from consuming excessive disk space.

### Configuration

```go
config := &logger.Config{
    Output:         "/var/log/ainative-code/app.log",
    EnableRotation: true,
    MaxSize:        100,   // Rotate after 100 MB
    MaxBackups:     5,     // Keep 5 old log files
    MaxAge:         30,    // Delete files older than 30 days
    Compress:       true,  // Compress rotated files with gzip
}
```

### Rotation Behavior

- **MaxSize**: When the current log file reaches this size (in MB), it is rotated
- **MaxBackups**: Maximum number of old log files to retain (0 = retain all)
- **MaxAge**: Maximum number of days to retain old log files (0 = never delete)
- **Compress**: Rotated files are compressed using gzip (.gz extension)

### Rotated File Naming

```
app.log           # Current log file
app-2025-12-26.log
app-2025-12-25.log.gz
app-2025-12-24.log.gz
```

## Performance Benchmarks

The logging system is designed for minimal performance overhead. Here are typical benchmark results:

### Simple Message Logging
```
BenchmarkLoggerSimpleMessage-8          5000000    250 ns/op    0 B/op    0 allocs/op
```

### Formatted Message Logging
```
BenchmarkLoggerFormattedMessage-8       3000000    450 ns/op    128 B/op  2 allocs/op
```

### Structured Field Logging
```
BenchmarkLoggerStructuredFields-8       2000000    600 ns/op    256 B/op  5 allocs/op
```

### Context-Aware Logging
```
BenchmarkLoggerContextAware-8           4000000    300 ns/op    64 B/op   1 allocs/op
```

### Disabled Log Level (No-op)
```
BenchmarkLoggerDisabledLevel-8         50000000     30 ns/op    0 B/op    0 allocs/op
```

### Running Benchmarks

```bash
cd internal/logger
go test -bench=. -benchmem
```

## Best Practices

### 1. Use Appropriate Log Levels

```go
// Good: Use DEBUG for development details
logger.Debugf("Cache hit for key: %s", key)

// Good: Use INFO for important events
logger.Info("User authentication successful")

// Good: Use WARN for recoverable issues
logger.Warn("Rate limit approaching threshold")

// Good: Use ERROR for serious problems
logger.ErrorWithErr("Failed to save to database", err)

// Bad: Don't use INFO for debugging details
logger.Info("Variable x = 42")  // Use DEBUG instead

// Bad: Don't use ERROR for expected conditions
logger.Error("User not found")  // Use WARN or INFO instead
```

### 2. Include Relevant Context

```go
// Good: Include all relevant information
logger.InfoWithFields("Payment processed", map[string]interface{}{
    "user_id": userID,
    "amount": amount,
    "currency": "USD",
    "payment_method": "credit_card",
    "transaction_id": txnID,
})

// Bad: Vague message without context
logger.Info("Payment processed")
```

### 3. Use Context-Aware Logging

```go
// Good: Use context for request tracking
func handleAPIRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    ctx = logger.WithRequestID(ctx, generateRequestID())
    ctx = logger.WithUserID(ctx, getUserID(r))

    log := logger.WithContext(ctx)
    log.Info("API request started")

    // All subsequent logs will include request_id and user_id
    processRequest(ctx)

    log.Info("API request completed")
}
```

### 4. Don't Log Sensitive Information

```go
// Bad: Logging sensitive data
logger.Infof("User password: %s", password)
logger.InfoWithFields("Auth", map[string]interface{}{
    "api_key": apiKey,
    "secret": secret,
})

// Good: Redact or omit sensitive data
logger.Info("User authenticated successfully")
logger.InfoWithFields("Auth", map[string]interface{}{
    "user_id": userID,
    "auth_method": "api_key",
})
```

### 5. Use Structured Logging for Machine Parsing

```go
// Good: Structured logging for metrics
logger.InfoWithFields("HTTP request", map[string]interface{}{
    "method": r.Method,
    "path": r.URL.Path,
    "status": statusCode,
    "duration_ms": duration.Milliseconds(),
    "bytes_sent": bytesSent,
})

// Less ideal: Formatted string (harder to parse)
logger.Infof("HTTP %s %s returned %d in %dms", r.Method, r.URL.Path, statusCode, duration.Milliseconds())
```

### 6. Configure for Environment

```go
// Development: Text format, DEBUG level, stdout
devConfig := &logger.Config{
    Level:  logger.DebugLevel,
    Format: logger.TextFormat,
    Output: "stdout",
    EnableCaller: true,
}

// Production: JSON format, INFO level, file with rotation
prodConfig := &logger.Config{
    Level:          logger.InfoLevel,
    Format:         logger.JSONFormat,
    Output:         "/var/log/ainative-code/app.log",
    EnableRotation: true,
    MaxSize:        100,
    MaxBackups:     10,
    MaxAge:         30,
    Compress:       true,
    EnableStackTrace: true,
}
```

## Examples

### Example 1: HTTP Server Logging

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/AINative-studio/ainative-code/internal/logger"
)

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Add request ID to context
        requestID := generateRequestID()
        ctx := logger.WithRequestID(r.Context(), requestID)
        r = r.WithContext(ctx)

        // Create request logger
        log := logger.WithContext(ctx)

        log.InfoWithFields("HTTP request started", map[string]interface{}{
            "method": r.Method,
            "path": r.URL.Path,
            "remote_addr": r.RemoteAddr,
        })

        // Call next handler
        next.ServeHTTP(w, r)

        // Log completion
        duration := time.Since(start)
        log.InfoWithFields("HTTP request completed", map[string]interface{}{
            "duration_ms": duration.Milliseconds(),
        })
    })
}
```

### Example 2: Database Operations

```go
package database

import (
    "context"
    "time"

    "github.com/AINative-studio/ainative-code/internal/logger"
)

func ExecuteQuery(ctx context.Context, query string, args ...interface{}) error {
    log := logger.WithContext(ctx)

    log.DebugWithFields("Executing database query", map[string]interface{}{
        "query": query,
        "args": args,
    })

    start := time.Now()

    // Execute query
    result, err := db.ExecContext(ctx, query, args...)

    duration := time.Since(start)

    if err != nil {
        log.ErrorWithFields("Database query failed", map[string]interface{}{
            "query": query,
            "duration_ms": duration.Milliseconds(),
            "error": err.Error(),
        })
        return err
    }

    rowsAffected, _ := result.RowsAffected()

    log.InfoWithFields("Database query succeeded", map[string]interface{}{
        "rows_affected": rowsAffected,
        "duration_ms": duration.Milliseconds(),
    })

    return nil
}
```

### Example 3: Application Startup

```go
package main

import (
    "os"

    "github.com/AINative-studio/ainative-code/internal/logger"
)

func main() {
    // Initialize logger based on environment
    env := os.Getenv("ENVIRONMENT")

    var config *logger.Config
    if env == "production" {
        config = &logger.Config{
            Level:          logger.InfoLevel,
            Format:         logger.JSONFormat,
            Output:         "/var/log/ainative-code/app.log",
            EnableRotation: true,
            MaxSize:        100,
            MaxBackups:     10,
            MaxAge:         30,
            Compress:       true,
        }
    } else {
        config = &logger.Config{
            Level:  logger.DebugLevel,
            Format: logger.TextFormat,
            Output: "stdout",
            EnableCaller: true,
        }
    }

    log, err := logger.New(config)
    if err != nil {
        panic(err)
    }

    logger.SetGlobalLogger(log)

    logger.InfoWithFields("Application starting", map[string]interface{}{
        "version": "1.0.0",
        "environment": env,
        "pid": os.Getpid(),
    })

    // Application code...

    logger.Info("Application shutdown complete")
}
```

## Troubleshooting

### Log File Permissions

If you encounter permission errors when writing to log files:

```go
// Ensure the log directory exists and has proper permissions
dir := filepath.Dir(config.Output)
if err := os.MkdirAll(dir, 0755); err != nil {
    return fmt.Errorf("failed to create log directory: %w", err)
}
```

### Log Rotation Not Working

Ensure `EnableRotation` is set to `true` and the output is a file path (not stdout/stderr):

```go
config := &logger.Config{
    Output:         "/var/log/app.log",  // Must be a file path
    EnableRotation: true,                // Must be enabled
    MaxSize:        100,
}
```

### High Memory Usage

If you notice high memory usage, consider:

1. Reducing `MaxBackups` to keep fewer old log files
2. Enabling `Compress` to compress rotated files
3. Setting a lower `MaxSize` to rotate more frequently
4. Using JSON format instead of text format (more efficient)

### Performance Issues

If logging is causing performance issues:

1. Increase the log level to reduce log volume (INFO instead of DEBUG)
2. Disable caller information (`EnableCaller: false`)
3. Disable stack traces (`EnableStackTrace: false`)
4. Use the global logger instead of creating new loggers frequently

## Integration with Configuration Files

### YAML Configuration Example

```yaml
# config.yaml
logging:
  level: "info"
  format: "json"
  output: "/var/log/ainative-code/app.log"
  rotation:
    enabled: true
    max_size: 100
    max_backups: 5
    max_age: 30
    compress: true
  features:
    enable_caller: false
    enable_stack_trace: true
```

### Loading Configuration

```go
package config

import (
    "github.com/spf13/viper"
    "github.com/AINative-studio/ainative-code/internal/logger"
)

func LoadLoggerConfig() (*logger.Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    return &logger.Config{
        Level:            logger.LogLevel(viper.GetString("logging.level")),
        Format:           logger.OutputFormat(viper.GetString("logging.format")),
        Output:           viper.GetString("logging.output"),
        EnableRotation:   viper.GetBool("logging.rotation.enabled"),
        MaxSize:          viper.GetInt("logging.rotation.max_size"),
        MaxBackups:       viper.GetInt("logging.rotation.max_backups"),
        MaxAge:           viper.GetInt("logging.rotation.max_age"),
        Compress:         viper.GetBool("logging.rotation.compress"),
        EnableCaller:     viper.GetBool("logging.features.enable_caller"),
        EnableStackTrace: viper.GetBool("logging.features.enable_stack_trace"),
    }, nil
}
```

## License

Copyright (c) 2025 AINative Studio. All rights reserved.
