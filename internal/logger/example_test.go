package logger_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AINative-studio/ainative-code/internal/logger"
)

// Example_basicUsage demonstrates basic logging
func Example_basicUsage() {
	// Create a logger with text format for readable output
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}

	log, err := logger.New(config)
	if err != nil {
		panic(err)
	}

	log.Info("Application started")
	log.Infof("Server listening on port %d", 8080)
	log.Warn("This is a warning")
}

// Example_structuredLogging demonstrates structured logging with fields
func Example_structuredLogging() {
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: "stdout",
	}

	log, err := logger.New(config)
	if err != nil {
		panic(err)
	}

	log.InfoWithFields("User action", map[string]interface{}{
		"user_id": "user123",
		"action":  "login",
		"ip":      "192.168.1.1",
	})
}

// Example_contextAwareLogging demonstrates context-aware logging
func Example_contextAwareLogging() {
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: "stdout",
	}

	log, err := logger.New(config)
	if err != nil {
		panic(err)
	}

	// Create context with IDs
	ctx := context.Background()
	ctx = logger.WithRequestID(ctx, "req-abc123")
	ctx = logger.WithSessionID(ctx, "sess-xyz789")

	// Create context-aware logger
	ctxLog := log.WithContext(ctx)

	// All logs will include request_id and session_id
	ctxLog.Info("Processing request")
	ctxLog.Info("Request completed")
}

// Example_fileLogging demonstrates logging to a file
func Example_fileLogging() {
	tmpDir := os.TempDir()
	logFile := filepath.Join(tmpDir, "app.log")

	config := &logger.Config{
		Level:  logger.DebugLevel,
		Format: logger.JSONFormat,
		Output: logFile,
	}

	log, err := logger.New(config)
	if err != nil {
		panic(err)
	}

	log.Debug("Debug message")
	log.Info("Info message")
	log.Warn("Warning message")

	fmt.Printf("Logs written to: %s\n", logFile)
}

// Example_logRotation demonstrates log rotation configuration
func Example_logRotation() {
	tmpDir := os.TempDir()
	logFile := filepath.Join(tmpDir, "rotating.log")

	config := &logger.Config{
		Level:          logger.InfoLevel,
		Format:         logger.JSONFormat,
		Output:         logFile,
		EnableRotation: true,
		MaxSize:        100, // 100 MB
		MaxBackups:     5,   // Keep 5 old files
		MaxAge:         30,  // 30 days
		Compress:       true,
	}

	log, err := logger.New(config)
	if err != nil {
		panic(err)
	}

	log.Info("This log will be rotated based on the configuration")

	fmt.Println("Log rotation configured successfully")
	// Output: Log rotation configured successfully
}

// Example_globalLogger demonstrates using the global logger
func Example_globalLogger() {
	// The global logger is automatically initialized with default config
	logger.Info("Using global logger")
	logger.Infof("Formatted message: %s", "example")

	// You can also set a custom global logger
	config := &logger.Config{
		Level:  logger.DebugLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}

	customLog, err := logger.New(config)
	if err != nil {
		panic(err)
	}

	logger.SetGlobalLogger(customLog)
	logger.Debug("This debug message will now be logged")
}

// Example_errorLogging demonstrates error logging
func Example_errorLogging() {
	config := &logger.Config{
		Level:            logger.ErrorLevel,
		Format:           logger.JSONFormat,
		Output:           "stdout",
		EnableStackTrace: true,
	}

	log, err := logger.New(config)
	if err != nil {
		panic(err)
	}

	// Log error with message
	log.Error("Something went wrong")

	// Log error with formatted message
	log.Errorf("Failed to process item %d", 42)

	// Log error with error object
	someError := fmt.Errorf("database connection failed")
	log.ErrorWithErr("Database error", someError)

	// Log error with structured fields
	log.ErrorWithFields("Processing failed", map[string]interface{}{
		"item_id": 123,
		"reason":  "timeout",
	})
}

// Example_differentLevels demonstrates all log levels
func Example_differentLevels() {
	config := &logger.Config{
		Level:  logger.DebugLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}

	log, err := logger.New(config)
	if err != nil {
		panic(err)
	}

	log.Debug("This is a debug message")
	log.Info("This is an info message")
	log.Warn("This is a warning message")
	log.Error("This is an error message")
}

// Example_httpServerLogging demonstrates logging in an HTTP server context
func Example_httpServerLogging() {
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: "stdout",
	}

	log, err := logger.New(config)
	if err != nil {
		panic(err)
	}

	// Simulate HTTP request handling
	requestID := "req-123456"
	userID := "user-789"

	ctx := context.Background()
	ctx = logger.WithRequestID(ctx, requestID)
	ctx = logger.WithUserID(ctx, userID)

	reqLog := log.WithContext(ctx)

	reqLog.InfoWithFields("HTTP request started", map[string]interface{}{
		"method": "GET",
		"path":   "/api/users",
	})

	reqLog.InfoWithFields("HTTP request completed", map[string]interface{}{
		"status":      200,
		"duration_ms": 45,
	})
}

// Example_environmentBasedConfig demonstrates configuring logger based on environment
func Example_environmentBasedConfig() {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	var config *logger.Config

	switch env {
	case "production":
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
	case "development":
		config = &logger.Config{
			Level:        logger.DebugLevel,
			Format:       logger.TextFormat,
			Output:       "stdout",
			EnableCaller: true,
		}
	default:
		config = logger.DefaultConfig()
	}

	log, err := logger.New(config)
	if err != nil {
		panic(err)
	}

	log.InfoWithFields("Logger configured", map[string]interface{}{
		"environment": env,
	})
}
