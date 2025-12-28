// Package errors provides a comprehensive error handling framework for AINative Code.
// It includes custom error types, error wrapping/unwrapping, stack traces, and recovery strategies.
package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// ErrorCode represents a unique error code for categorization
type ErrorCode string

const (
	// Configuration error codes
	ErrCodeConfigInvalid     ErrorCode = "CONFIG_INVALID"
	ErrCodeConfigMissing     ErrorCode = "CONFIG_MISSING"
	ErrCodeConfigParse       ErrorCode = "CONFIG_PARSE"
	ErrCodeConfigValidation  ErrorCode = "CONFIG_VALIDATION"

	// Authentication error codes
	ErrCodeAuthFailed        ErrorCode = "AUTH_FAILED"
	ErrCodeAuthInvalidToken  ErrorCode = "AUTH_INVALID_TOKEN"
	ErrCodeAuthExpiredToken  ErrorCode = "AUTH_EXPIRED_TOKEN"
	ErrCodeAuthPermissionDenied ErrorCode = "AUTH_PERMISSION_DENIED"
	ErrCodeAuthInvalidCredentials ErrorCode = "AUTH_INVALID_CREDENTIALS"

	// Provider error codes
	ErrCodeProviderUnavailable ErrorCode = "PROVIDER_UNAVAILABLE"
	ErrCodeProviderTimeout     ErrorCode = "PROVIDER_TIMEOUT"
	ErrCodeProviderRateLimit   ErrorCode = "PROVIDER_RATE_LIMIT"
	ErrCodeProviderInvalidResponse ErrorCode = "PROVIDER_INVALID_RESPONSE"
	ErrCodeProviderNotFound    ErrorCode = "PROVIDER_NOT_FOUND"

	// Tool execution error codes
	ErrCodeToolNotFound        ErrorCode = "TOOL_NOT_FOUND"
	ErrCodeToolExecutionFailed ErrorCode = "TOOL_EXECUTION_FAILED"
	ErrCodeToolTimeout         ErrorCode = "TOOL_TIMEOUT"
	ErrCodeToolInvalidInput    ErrorCode = "TOOL_INVALID_INPUT"
	ErrCodeToolPermissionDenied ErrorCode = "TOOL_PERMISSION_DENIED"

	// Database error codes
	ErrCodeDBConnection        ErrorCode = "DB_CONNECTION_FAILED"
	ErrCodeDBQuery             ErrorCode = "DB_QUERY_FAILED"
	ErrCodeDBNotFound          ErrorCode = "DB_NOT_FOUND"
	ErrCodeDBDuplicate         ErrorCode = "DB_DUPLICATE"
	ErrCodeDBConstraint        ErrorCode = "DB_CONSTRAINT_VIOLATION"
	ErrCodeDBTransaction       ErrorCode = "DB_TRANSACTION_FAILED"

	// Security error codes
	ErrCodeSecurityViolation   ErrorCode = "SECURITY_VIOLATION"
	ErrCodeSecurityInvalidKey  ErrorCode = "SECURITY_INVALID_KEY"
	ErrCodeSecurityEncryption  ErrorCode = "SECURITY_ENCRYPTION_FAILED"
	ErrCodeSecurityDecryption  ErrorCode = "SECURITY_DECRYPTION_FAILED"
)

// Severity represents the severity level of an error
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// StackFrame represents a single frame in the stack trace
type StackFrame struct {
	File     string
	Line     int
	Function string
}

// BaseError is the foundation for all custom errors in the system
type BaseError struct {
	code      ErrorCode
	message   string
	userMsg   string
	severity  Severity
	cause     error
	stack     []StackFrame
	metadata  map[string]interface{}
	retryable bool
}

// Error implements the error interface
func (e *BaseError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.code, e.message, e.cause)
	}
	return fmt.Sprintf("[%s] %s", e.code, e.message)
}

// Unwrap returns the underlying error
func (e *BaseError) Unwrap() error {
	return e.cause
}

// Code returns the error code
func (e *BaseError) Code() ErrorCode {
	return e.code
}

// UserMessage returns a user-friendly error message
func (e *BaseError) UserMessage() string {
	if e.userMsg != "" {
		return e.userMsg
	}
	return e.message
}

// Severity returns the error severity
func (e *BaseError) Severity() Severity {
	return e.severity
}

// Stack returns the stack trace
func (e *BaseError) Stack() []StackFrame {
	return e.stack
}

// IsRetryable returns whether the error is retryable
func (e *BaseError) IsRetryable() bool {
	return e.retryable
}

// Metadata returns error metadata
func (e *BaseError) Metadata() map[string]interface{} {
	return e.metadata
}

// WithMetadata adds metadata to the error
func (e *BaseError) WithMetadata(key string, value interface{}) *BaseError {
	if e.metadata == nil {
		e.metadata = make(map[string]interface{})
	}
	e.metadata[key] = value
	return e
}

// StackTrace returns a formatted stack trace string
func (e *BaseError) StackTrace() string {
	if len(e.stack) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Stack trace:\n")
	for _, frame := range e.stack {
		sb.WriteString(fmt.Sprintf("  %s:%d %s\n", frame.File, frame.Line, frame.Function))
	}
	return sb.String()
}

// captureStack captures the current stack trace
func captureStack(skip int) []StackFrame {
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(skip, pcs[:])

	frames := make([]StackFrame, 0, n)
	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pcs[i])
		if fn == nil {
			continue
		}
		file, line := fn.FileLine(pcs[i])
		frames = append(frames, StackFrame{
			File:     file,
			Line:     line,
			Function: fn.Name(),
		})
	}
	return frames
}

// newError creates a new BaseError with stack trace
func newError(code ErrorCode, message string, severity Severity, retryable bool) *BaseError {
	return &BaseError{
		code:      code,
		message:   message,
		severity:  severity,
		retryable: retryable,
		stack:     captureStack(3),
		metadata:  make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string) error {
	if err == nil {
		return nil
	}

	return &BaseError{
		code:     code,
		message:  message,
		cause:    err,
		severity: SeverityMedium,
		stack:    captureStack(2),
		metadata: make(map[string]interface{}),
	}
}

// Wrapf wraps an error with a formatted message
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return Wrap(err, code, fmt.Sprintf(format, args...))
}

// Is reports whether any error in err's chain matches target
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// GetCode extracts the error code from an error
func GetCode(err error) ErrorCode {
	// Try interface first for embedded types
	type coder interface {
		Code() ErrorCode
	}
	if c, ok := err.(coder); ok {
		return c.Code()
	}

	// Fall back to errors.As for wrapped types
	var baseErr *BaseError
	if errors.As(err, &baseErr) {
		return baseErr.Code()
	}
	return ""
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	// Try interface first for embedded types
	type retryable interface {
		IsRetryable() bool
	}
	if r, ok := err.(retryable); ok {
		return r.IsRetryable()
	}

	// Fall back to errors.As for wrapped types
	var baseErr *BaseError
	if errors.As(err, &baseErr) {
		return baseErr.IsRetryable()
	}
	return false
}

// GetSeverity extracts the severity from an error
func GetSeverity(err error) Severity {
	// Try interface first for embedded types
	type severityProvider interface {
		Severity() Severity
	}
	if s, ok := err.(severityProvider); ok {
		return s.Severity()
	}

	// Fall back to errors.As for wrapped types
	var baseErr *BaseError
	if errors.As(err, &baseErr) {
		return baseErr.Severity()
	}
	return SeverityLow
}
