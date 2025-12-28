package logger

import (
	"context"
	"sync"

	"github.com/rs/zerolog"
)

var (
	// globalLogger is the global logger instance
	globalLogger *Logger
	// mu protects the global logger
	mu sync.RWMutex
)

// init initializes the global logger with default configuration
func init() {
	var err error
	globalLogger, err = New(DefaultConfig())
	if err != nil {
		panic("failed to initialize global logger: " + err.Error())
	}
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	mu.RLock()
	defer mu.RUnlock()
	return globalLogger
}

// Global logging functions that use the global logger instance

// Debug logs a debug level message using the global logger
func Debug(msg string) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Debug(msg)
}

// Debugf logs a formatted debug level message using the global logger
func Debugf(format string, args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Debugf(format, args...)
}

// DebugWithFields logs a debug level message with structured fields using the global logger
func DebugWithFields(msg string, fields map[string]interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.DebugWithFields(msg, fields)
}

// Info logs an info level message using the global logger
func Info(msg string) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Info(msg)
}

// Infof logs a formatted info level message using the global logger
func Infof(format string, args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Infof(format, args...)
}

// InfoWithFields logs an info level message with structured fields using the global logger
func InfoWithFields(msg string, fields map[string]interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.InfoWithFields(msg, fields)
}

// Warn logs a warning level message using the global logger
func Warn(msg string) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Warn(msg)
}

// Warnf logs a formatted warning level message using the global logger
func Warnf(format string, args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Warnf(format, args...)
}

// WarnWithFields logs a warning level message with structured fields using the global logger
func WarnWithFields(msg string, fields map[string]interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.WarnWithFields(msg, fields)
}

// Error logs an error level message using the global logger
func Error(msg string) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Error(msg)
}

// Errorf logs a formatted error level message using the global logger
func Errorf(format string, args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Errorf(format, args...)
}

// ErrorWithFields logs an error level message with structured fields using the global logger
func ErrorWithFields(msg string, fields map[string]interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.ErrorWithFields(msg, fields)
}

// ErrorWithErr logs an error level message with an error using the global logger
func ErrorWithErr(msg string, err error) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.ErrorWithErr(msg, err)
}

// Fatal logs a fatal level message and exits the program using the global logger
func Fatal(msg string) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Fatal(msg)
}

// Fatalf logs a formatted fatal level message and exits the program using the global logger
func Fatalf(format string, args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Fatalf(format, args...)
}

// WithContext returns a logger with context values extracted and added as fields
func WithContext(ctx context.Context) *Logger {
	mu.RLock()
	defer mu.RUnlock()
	return globalLogger.WithContext(ctx)
}

// Init initializes the global logger with default configuration
func Init() {
	// Already initialized in init(), this is a no-op for compatibility
}

// SetLevel sets the log level for the global logger
func SetLevel(level string) error {
	config := DefaultConfig()
	config.Level = LogLevel(level)

	newLogger, err := New(config)
	if err != nil {
		return err
	}

	SetGlobalLogger(newLogger)
	return nil
}

// Event helpers that return zerolog.Event for chaining

// DebugEvent returns a debug level event for chaining
func DebugEvent() *zerolog.Event {
	mu.RLock()
	defer mu.RUnlock()
	return globalLogger.logger.Debug()
}

// InfoEvent returns an info level event for chaining
func InfoEvent() *zerolog.Event {
	mu.RLock()
	defer mu.RUnlock()
	return globalLogger.logger.Info()
}

// WarnEvent returns a warn level event for chaining
func WarnEvent() *zerolog.Event {
	mu.RLock()
	defer mu.RUnlock()
	return globalLogger.logger.Warn()
}

// ErrorEvent returns an error level event for chaining
func ErrorEvent() *zerolog.Event {
	mu.RLock()
	defer mu.RUnlock()
	event := globalLogger.logger.Error()
	if globalLogger.config.EnableStackTrace {
		event = event.Stack()
	}
	return event
}
