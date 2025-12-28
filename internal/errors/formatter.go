package errors

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DebugMode controls whether stack traces are included in error output
var DebugMode = false

// EnableDebugMode enables debug mode with stack traces
func EnableDebugMode() {
	DebugMode = true
}

// DisableDebugMode disables debug mode
func DisableDebugMode() {
	DebugMode = false
}

// IsDebugMode returns whether debug mode is enabled
func IsDebugMode() bool {
	return DebugMode
}

// Format formats an error for display
func Format(err error) string {
	if err == nil {
		return ""
	}

	var baseErr *BaseError
	if !As(err, &baseErr) {
		return err.Error()
	}

	var sb strings.Builder

	// Add error code and message
	sb.WriteString(fmt.Sprintf("[%s] %s", baseErr.Code(), baseErr.message))

	// Add stack trace in debug mode
	if DebugMode && len(baseErr.Stack()) > 0 {
		sb.WriteString("\n\n")
		sb.WriteString(baseErr.StackTrace())
	}

	// Add cause if present
	if baseErr.cause != nil {
		sb.WriteString("\n\nCaused by: ")
		sb.WriteString(baseErr.cause.Error())
	}

	// Add metadata in debug mode
	if DebugMode && len(baseErr.Metadata()) > 0 {
		sb.WriteString("\n\nMetadata:\n")
		for key, value := range baseErr.Metadata() {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	return sb.String()
}

// FormatUser formats an error for end-user display (friendly message)
func FormatUser(err error) string {
	if err == nil {
		return ""
	}

	// Try to extract user message from different error types
	type userMessenger interface {
		UserMessage() string
	}

	if um, ok := err.(userMessenger); ok {
		return um.UserMessage()
	}

	var baseErr *BaseError
	if As(err, &baseErr) {
		return baseErr.UserMessage()
	}

	return "An unexpected error occurred. Please try again."
}

// ErrorResponse represents a JSON error response
type ErrorResponse struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	UserMsg   string                 `json:"user_message,omitempty"`
	Severity  string                 `json:"severity"`
	Retryable bool                   `json:"retryable"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Stack     []StackFrame           `json:"stack,omitempty"`
}

// ToJSON converts an error to a JSON representation
func ToJSON(err error) ([]byte, error) {
	if err == nil {
		return nil, nil
	}

	var baseErr *BaseError
	if !As(err, &baseErr) {
		// Fallback for non-BaseError errors
		response := ErrorResponse{
			Code:      "UNKNOWN_ERROR",
			Message:   err.Error(),
			Severity:  string(SeverityMedium),
			Retryable: false,
		}
		return json.Marshal(response)
	}

	response := ErrorResponse{
		Code:      string(baseErr.Code()),
		Message:   baseErr.message,
		UserMsg:   baseErr.UserMessage(),
		Severity:  string(baseErr.Severity()),
		Retryable: baseErr.IsRetryable(),
		Metadata:  baseErr.Metadata(),
	}

	// Include stack trace only in debug mode
	if DebugMode {
		response.Stack = baseErr.Stack()
	}

	return json.Marshal(response)
}

// FromJSON creates an error from a JSON representation
func FromJSON(data []byte) error {
	var response ErrorResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	baseErr := &BaseError{
		code:      ErrorCode(response.Code),
		message:   response.Message,
		userMsg:   response.UserMsg,
		severity:  Severity(response.Severity),
		retryable: response.Retryable,
		metadata:  response.Metadata,
		stack:     response.Stack,
	}

	return baseErr
}

// ErrorSummary provides a summary of an error for logging
type ErrorSummary struct {
	Code      string
	Message   string
	Severity  string
	Retryable bool
	HasCause  bool
	StackSize int
}

// Summarize creates a summary of an error
func Summarize(err error) ErrorSummary {
	if err == nil {
		return ErrorSummary{}
	}

	var baseErr *BaseError
	if !As(err, &baseErr) {
		return ErrorSummary{
			Code:     "UNKNOWN",
			Message:  err.Error(),
			Severity: string(SeverityMedium),
		}
	}

	return ErrorSummary{
		Code:      string(baseErr.Code()),
		Message:   baseErr.message,
		Severity:  string(baseErr.Severity()),
		Retryable: baseErr.IsRetryable(),
		HasCause:  baseErr.cause != nil,
		StackSize: len(baseErr.Stack()),
	}
}

// Chain returns all errors in the error chain
func Chain(err error) []error {
	if err == nil {
		return nil
	}

	var chain []error
	for err != nil {
		chain = append(chain, err)
		err = Unwrap(err)
	}
	return chain
}

// Unwrap returns the underlying error
func Unwrap(err error) error {
	type unwrapper interface {
		Unwrap() error
	}

	u, ok := err.(unwrapper)
	if !ok {
		return nil
	}
	return u.Unwrap()
}

// RootCause returns the root cause of an error
func RootCause(err error) error {
	for {
		unwrapped := Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

// FormatChain formats the entire error chain
func FormatChain(err error) string {
	if err == nil {
		return ""
	}

	chain := Chain(err)
	var sb strings.Builder

	for i, e := range chain {
		if i > 0 {
			sb.WriteString("\n\nWrapped by:\n")
		}
		sb.WriteString(Format(e))
	}

	return sb.String()
}
