// Package tools provides an extensible tool execution framework for LLM assistants.
package tools

import (
	"fmt"
)

// ErrToolNotFound is returned when a requested tool is not registered.
type ErrToolNotFound struct {
	ToolName string
}

func (e *ErrToolNotFound) Error() string {
	return fmt.Sprintf("tool not found: %s", e.ToolName)
}

// ErrInvalidInput is returned when tool input validation fails.
type ErrInvalidInput struct {
	ToolName string
	Field    string
	Reason   string
}

func (e *ErrInvalidInput) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("invalid input for tool %s: field %s - %s", e.ToolName, e.Field, e.Reason)
	}
	return fmt.Sprintf("invalid input for tool %s: %s", e.ToolName, e.Reason)
}

// ErrExecutionFailed is returned when tool execution fails.
type ErrExecutionFailed struct {
	ToolName string
	Reason   string
	Cause    error
}

func (e *ErrExecutionFailed) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("execution failed for tool %s: %s - caused by: %v", e.ToolName, e.Reason, e.Cause)
	}
	return fmt.Sprintf("execution failed for tool %s: %s", e.ToolName, e.Reason)
}

func (e *ErrExecutionFailed) Unwrap() error {
	return e.Cause
}

// ErrTimeout is returned when tool execution exceeds the timeout.
type ErrTimeout struct {
	ToolName string
	Duration string
}

func (e *ErrTimeout) Error() string {
	return fmt.Sprintf("tool %s execution timed out after %s", e.ToolName, e.Duration)
}

// ErrPermissionDenied is returned when a tool operation is not permitted.
type ErrPermissionDenied struct {
	ToolName  string
	Operation string
	Resource  string
	Reason    string
}

func (e *ErrPermissionDenied) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("permission denied for tool %s: cannot %s resource %s - %s",
			e.ToolName, e.Operation, e.Resource, e.Reason)
	}
	return fmt.Sprintf("permission denied for tool %s: cannot %s - %s",
		e.ToolName, e.Operation, e.Reason)
}

// ErrToolConflict is returned when attempting to register a tool with a name that already exists.
type ErrToolConflict struct {
	ToolName string
}

func (e *ErrToolConflict) Error() string {
	return fmt.Sprintf("tool conflict: a tool with name %s is already registered", e.ToolName)
}

// ErrOutputTooLarge is returned when tool output exceeds the maximum allowed size.
type ErrOutputTooLarge struct {
	ToolName   string
	OutputSize int64
	MaxSize    int64
}

func (e *ErrOutputTooLarge) Error() string {
	return fmt.Sprintf("output too large for tool %s: %d bytes exceeds maximum of %d bytes",
		e.ToolName, e.OutputSize, e.MaxSize)
}
