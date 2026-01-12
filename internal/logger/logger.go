// Package logger provides a structured logging system with configurable log levels,
// output formats, rotation, and context-aware logging capabilities.
package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	// DebugLevel logs are typically voluminous and are usually disabled in production
	DebugLevel LogLevel = "debug"
	// InfoLevel is the default logging priority
	InfoLevel LogLevel = "info"
	// WarnLevel logs are more important than Info, but don't need individual human review
	WarnLevel LogLevel = "warn"
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs
	ErrorLevel LogLevel = "error"
)

// OutputFormat specifies the format for log output
type OutputFormat string

const (
	// JSONFormat outputs logs in JSON format
	JSONFormat OutputFormat = "json"
	// TextFormat outputs logs in human-readable console format
	TextFormat OutputFormat = "text"
)

// Config holds the configuration for the logger
type Config struct {
	// Level sets the minimum log level that will be output
	Level LogLevel

	// Format specifies the output format (json or text)
	Format OutputFormat

	// Output specifies where logs should be written (file path or "stdout"/"stderr")
	Output string

	// EnableRotation enables log file rotation
	EnableRotation bool

	// MaxSize is the maximum size in megabytes of the log file before it gets rotated
	// Only applies when EnableRotation is true
	MaxSize int

	// MaxBackups is the maximum number of old log files to retain
	// Only applies when EnableRotation is true
	MaxBackups int

	// MaxAge is the maximum number of days to retain old log files
	// Only applies when EnableRotation is true
	MaxAge int

	// Compress determines if the rotated log files should be compressed using gzip
	// Only applies when EnableRotation is true
	Compress bool

	// EnableCaller adds the file and line number where the log was called
	EnableCaller bool

	// EnableStackTrace adds stack traces for error level logs
	EnableStackTrace bool
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:            InfoLevel,
		Format:           TextFormat,
		Output:           "stderr",
		EnableRotation:   false,
		MaxSize:          100, // 100 MB
		MaxBackups:       3,
		MaxAge:           28, // 28 days
		Compress:         true,
		EnableCaller:     false,
		EnableStackTrace: false,
	}
}

// Logger wraps zerolog.Logger with additional context
type Logger struct {
	logger zerolog.Logger
	config *Config
}

// LoggerInterface defines the interface for logging operations
// This allows for test mocks and alternative logger implementations
type LoggerInterface interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	sessionIDKey contextKey = "session_id"
	userIDKey    contextKey = "user_id"
)

// New creates a new Logger instance with the provided configuration
func New(config *Config) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Set global log level
	level, err := parseLogLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	zerolog.SetGlobalLevel(level)

	// Configure output writer
	var writer io.Writer
	switch config.Output {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		// File output
		if config.EnableRotation {
			writer = &lumberjack.Logger{
				Filename:   config.Output,
				MaxSize:    config.MaxSize,
				MaxBackups: config.MaxBackups,
				MaxAge:     config.MaxAge,
				Compress:   config.Compress,
			}
		} else {
			// Ensure directory exists
			dir := filepath.Dir(config.Output)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create log directory: %w", err)
			}

			file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return nil, fmt.Errorf("failed to open log file: %w", err)
			}
			writer = file
		}
	}

	// Configure format
	if config.Format == TextFormat {
		writer = zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: time.RFC3339,
			NoColor:    config.Output != "stdout" && config.Output != "stderr",
		}
	}

	// Create logger
	zlog := zerolog.New(writer).With().Timestamp().Logger()

	// Enable caller if configured
	if config.EnableCaller {
		zlog = zlog.With().Caller().Logger()
	}

	return &Logger{
		logger: zlog,
		config: config,
	}, nil
}

// parseLogLevel converts LogLevel to zerolog.Level
func parseLogLevel(level LogLevel) (zerolog.Level, error) {
	switch level {
	case DebugLevel:
		return zerolog.DebugLevel, nil
	case InfoLevel:
		return zerolog.InfoLevel, nil
	case WarnLevel:
		return zerolog.WarnLevel, nil
	case ErrorLevel:
		return zerolog.ErrorLevel, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

// WithContext returns a logger with context values extracted and added as fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.logger

	// Extract and add request ID
	if requestID, ok := ctx.Value(requestIDKey).(string); ok && requestID != "" {
		logger = logger.With().Str("request_id", requestID).Logger()
	}

	// Extract and add session ID
	if sessionID, ok := ctx.Value(sessionIDKey).(string); ok && sessionID != "" {
		logger = logger.With().Str("session_id", sessionID).Logger()
	}

	// Extract and add user ID
	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		logger = logger.With().Str("user_id", userID).Logger()
	}

	return &Logger{
		logger: logger,
		config: l.config,
	}
}

// With returns a new logger with additional fields
func (l *Logger) With() zerolog.Context {
	return l.logger.With()
}

// Debug logs a debug level message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug level message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// DebugWithFields logs a debug level message with structured fields
func (l *Logger) DebugWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Info logs an info level message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info level message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// InfoWithFields logs an info level message with structured fields
func (l *Logger) InfoWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Warn logs a warning level message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning level message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// WarnWithFields logs a warning level message with structured fields
func (l *Logger) WarnWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Warn()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Error logs an error level message
func (l *Logger) Error(msg string) {
	event := l.logger.Error()
	if l.config.EnableStackTrace {
		event = event.Stack()
	}
	event.Msg(msg)
}

// Errorf logs a formatted error level message
func (l *Logger) Errorf(format string, args ...interface{}) {
	event := l.logger.Error()
	if l.config.EnableStackTrace {
		event = event.Stack()
	}
	event.Msgf(format, args...)
}

// ErrorWithFields logs an error level message with structured fields
func (l *Logger) ErrorWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Error()
	if l.config.EnableStackTrace {
		event = event.Stack()
	}
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// ErrorWithErr logs an error level message with an error
func (l *Logger) ErrorWithErr(msg string, err error) {
	event := l.logger.Error().Err(err)
	if l.config.EnableStackTrace {
		event = event.Stack()
	}
	event.Msg(msg)
}

// Fatal logs a fatal level message and exits the program
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal level message and exits the program
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// GetZerologLogger returns the underlying zerolog.Logger for advanced usage
func (l *Logger) GetZerologLogger() zerolog.Logger {
	return l.logger
}

// Context helpers for adding IDs to context

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithSessionID adds a session ID to the context
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDKey).(string)
	return requestID, ok
}

// GetSessionID retrieves the session ID from the context
func GetSessionID(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDKey).(string)
	return sessionID, ok
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
