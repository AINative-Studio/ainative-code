package session

import (
	"errors"
	"fmt"
)

// Common session errors
var (
	// ErrSessionNotFound is returned when a session is not found
	ErrSessionNotFound = errors.New("session not found")

	// ErrMessageNotFound is returned when a message is not found
	ErrMessageNotFound = errors.New("message not found")

	// ErrInvalidSessionID is returned when a session ID is invalid
	ErrInvalidSessionID = errors.New("invalid session ID")

	// ErrInvalidMessageID is returned when a message ID is invalid
	ErrInvalidMessageID = errors.New("invalid message ID")

	// ErrInvalidStatus is returned when a session status is invalid
	ErrInvalidStatus = errors.New("invalid session status")

	// ErrInvalidRole is returned when a message role is invalid
	ErrInvalidRole = errors.New("invalid message role")

	// ErrEmptySessionName is returned when session name is empty
	ErrEmptySessionName = errors.New("session name cannot be empty")

	// ErrEmptyMessageContent is returned when message content is empty
	ErrEmptyMessageContent = errors.New("message content cannot be empty")

	// ErrSessionDeleted is returned when operating on a deleted session
	ErrSessionDeleted = errors.New("session is deleted")

	// ErrInvalidExportFormat is returned when export format is invalid
	ErrInvalidExportFormat = errors.New("invalid export format")

	// ErrInvalidImportData is returned when import data is invalid
	ErrInvalidImportData = errors.New("invalid import data")

	// ErrCircularReference is returned when a message references itself as parent
	ErrCircularReference = errors.New("circular reference detected in message thread")

	// ErrEmptySearchQuery is returned when search query is empty
	ErrEmptySearchQuery = errors.New("search query cannot be empty")

	// ErrSearchLimitExceeded is returned when search limit exceeds maximum
	ErrSearchLimitExceeded = errors.New("search limit exceeds maximum allowed (1000)")

	// ErrInvalidDateRange is returned when date range is invalid
	ErrInvalidDateRange = errors.New("invalid date range: date_from must be before date_to")
)

// SessionError wraps errors with additional context
type SessionError struct {
	Op      string // Operation that failed
	Err     error  // Underlying error
	Context string // Additional context
}

// Error implements the error interface
func (e *SessionError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("session: %s: %s: %v", e.Op, e.Context, e.Err)
	}
	return fmt.Sprintf("session: %s: %v", e.Op, e.Err)
}

// Unwrap implements error unwrapping
func (e *SessionError) Unwrap() error {
	return e.Err
}

// NewSessionError creates a new SessionError
func NewSessionError(op string, err error, context string) *SessionError {
	return &SessionError{
		Op:      op,
		Err:     err,
		Context: context,
	}
}
