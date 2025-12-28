package errors

import (
	"fmt"
	"time"
)

// ToolExecutionError represents errors during tool execution
type ToolExecutionError struct {
	*BaseError
	ToolName   string
	ToolPath   string
	Parameters map[string]interface{}
	ExitCode   int
	Output     string
}

// NewToolExecutionError creates a new tool execution error
func NewToolExecutionError(code ErrorCode, message string, toolName string) *ToolExecutionError {
	baseErr := newError(code, message, SeverityMedium, false)
	return &ToolExecutionError{
		BaseError:  baseErr,
		ToolName:   toolName,
		Parameters: make(map[string]interface{}),
	}
}

// NewToolNotFoundError creates an error for tools that cannot be found
func NewToolNotFoundError(toolName string) *ToolExecutionError {
	msg := fmt.Sprintf("Tool '%s' not found", toolName)
	userMsg := fmt.Sprintf("Tool error: '%s' is not available or not installed. Please verify the tool is properly configured.", toolName)

	err := NewToolExecutionError(ErrCodeToolNotFound, msg, toolName)
	err.userMsg = userMsg
	err.severity = SeverityHigh
	return err
}

// NewToolExecutionFailedError creates an error for failed tool executions
func NewToolExecutionFailedError(toolName string, exitCode int, output string, cause error) *ToolExecutionError {
	msg := fmt.Sprintf("Tool '%s' execution failed with exit code %d", toolName, exitCode)
	userMsg := fmt.Sprintf("Tool execution failed: '%s' encountered an error. Please check the tool configuration and try again.", toolName)

	baseErr := newError(ErrCodeToolExecutionFailed, msg, SeverityMedium, false)
	baseErr.cause = cause
	baseErr.userMsg = userMsg

	return &ToolExecutionError{
		BaseError:  baseErr,
		ToolName:   toolName,
		ExitCode:   exitCode,
		Output:     output,
		Parameters: make(map[string]interface{}),
	}
}

// NewToolTimeoutError creates an error for tool execution timeouts
func NewToolTimeoutError(toolName string, timeout time.Duration) *ToolExecutionError {
	msg := fmt.Sprintf("Tool '%s' execution timed out after %v", toolName, timeout)
	userMsg := fmt.Sprintf("Tool timeout: '%s' took too long to execute. The operation has been cancelled.", toolName)

	err := NewToolExecutionError(ErrCodeToolTimeout, msg, toolName)
	err.userMsg = userMsg
	err.retryable = true
	return err
}

// NewToolInvalidInputError creates an error for invalid tool input
func NewToolInvalidInputError(toolName, paramName, reason string) *ToolExecutionError {
	msg := fmt.Sprintf("Invalid input for tool '%s', parameter '%s': %s", toolName, paramName, reason)
	userMsg := fmt.Sprintf("Invalid input: The parameter '%s' for tool '%s' is not valid. %s", paramName, toolName, reason)

	err := NewToolExecutionError(ErrCodeToolInvalidInput, msg, toolName)
	err.userMsg = userMsg
	err.severity = SeverityLow
	return err
}

// NewToolPermissionDeniedError creates an error for permission-related tool failures
func NewToolPermissionDeniedError(toolName, resource string) *ToolExecutionError {
	msg := fmt.Sprintf("Permission denied: tool '%s' cannot access resource '%s'", toolName, resource)
	userMsg := fmt.Sprintf("Permission denied: '%s' does not have the required permissions to access the requested resource.", toolName)

	err := NewToolExecutionError(ErrCodeToolPermissionDenied, msg, toolName)
	err.userMsg = userMsg
	err.severity = SeverityHigh
	return err
}

// WithPath sets the tool path
func (e *ToolExecutionError) WithPath(path string) *ToolExecutionError {
	e.ToolPath = path
	return e
}

// WithParameter adds a parameter to the error context
func (e *ToolExecutionError) WithParameter(name string, value interface{}) *ToolExecutionError {
	e.Parameters[name] = value
	return e
}

// WithExitCode sets the exit code
func (e *ToolExecutionError) WithExitCode(exitCode int) *ToolExecutionError {
	e.ExitCode = exitCode
	return e
}

// WithOutput sets the tool output
func (e *ToolExecutionError) WithOutput(output string) *ToolExecutionError {
	e.Output = output
	return e
}

// GetOutput returns the tool output, truncated if necessary
func (e *ToolExecutionError) GetOutput(maxLength int) string {
	if len(e.Output) <= maxLength {
		return e.Output
	}
	return e.Output[:maxLength] + "... (truncated)"
}
