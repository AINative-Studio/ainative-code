package events

import (
	"errors"
	"fmt"
)

// Sentinel errors for stream operations
var (
	// ErrClosed indicates the stream has been closed
	ErrClosed = errors.New("stream is closed")

	// ErrNotFound indicates the stream was not found
	ErrNotFound = errors.New("stream not found")

	// ErrFull indicates the stream buffer is full (backpressure)
	ErrFull = errors.New("stream buffer is full")

	// ErrInvalid indicates an invalid event
	ErrInvalid = errors.New("invalid event")
)

// StreamError represents an error that occurred during stream operations
// It wraps the underlying error and provides context about the stream and operation
type StreamError struct {
	// StreamID identifies the stream where the error occurred
	StreamID string

	// Op is the operation that failed (e.g., "send", "receive", "close")
	Op string

	// Err is the underlying error
	Err error
}

// Error implements the error interface
func (e *StreamError) Error() string {
	return fmt.Sprintf("stream %s: %s: %v", e.StreamID, e.Op, e.Err)
}

// Unwrap returns the underlying error for error chain support
func (e *StreamError) Unwrap() error {
	return e.Err
}

// ErrStreamClosed creates a new stream closed error
func ErrStreamClosed(streamID string) error {
	return &StreamError{
		StreamID: streamID,
		Op:       "access",
		Err:      ErrClosed,
	}
}

// ErrStreamNotFound creates a new stream not found error
func ErrStreamNotFound(streamID string) error {
	return &StreamError{
		StreamID: streamID,
		Op:       "lookup",
		Err:      ErrNotFound,
	}
}

// ErrStreamFull creates a new stream full error (backpressure)
func ErrStreamFull(streamID string, bufferSize int) error {
	return &StreamError{
		StreamID: streamID,
		Op:       "send",
		Err:      fmt.Errorf("%w: buffer size %d", ErrFull, bufferSize),
	}
}

// ErrInvalidEvent creates a new invalid event error
func ErrInvalidEvent(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalid, reason)
}

// IsStreamClosed checks if an error is a stream closed error
func IsStreamClosed(err error) bool {
	return errors.Is(err, ErrClosed)
}

// IsStreamNotFound checks if an error is a stream not found error
func IsStreamNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsStreamFull checks if an error is a stream full error
func IsStreamFull(err error) bool {
	return errors.Is(err, ErrFull)
}

// IsInvalidEvent checks if an error is an invalid event error
func IsInvalidEvent(err error) bool {
	return errors.Is(err, ErrInvalid)
}
