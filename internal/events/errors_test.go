package events

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrStreamClosed(t *testing.T) {
	streamID := "stream-123"
	err := ErrStreamClosed(streamID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream is closed")
	assert.Contains(t, err.Error(), streamID)
	assert.True(t, errors.Is(err, ErrClosed))
}

func TestErrStreamNotFound(t *testing.T) {
	streamID := "stream-456"
	err := ErrStreamNotFound(streamID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream not found")
	assert.Contains(t, err.Error(), streamID)
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestErrStreamFull(t *testing.T) {
	streamID := "stream-789"
	bufferSize := 100
	err := ErrStreamFull(streamID, bufferSize)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream buffer is full")
	assert.Contains(t, err.Error(), streamID)
	assert.Contains(t, err.Error(), "100")
	assert.True(t, errors.Is(err, ErrFull))
}

func TestErrInvalidEvent(t *testing.T) {
	reason := "missing required field"
	err := ErrInvalidEvent(reason)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event")
	assert.Contains(t, err.Error(), reason)
	assert.True(t, errors.Is(err, ErrInvalid))
}

func TestStreamError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *StreamError
		expected string
	}{
		{
			name: "stream closed error",
			err: &StreamError{
				StreamID: "stream-1",
				Op:       "send",
				Err:      ErrClosed,
			},
			expected: "stream stream-1: send: stream is closed",
		},
		{
			name: "stream not found error",
			err: &StreamError{
				StreamID: "stream-2",
				Op:       "get",
				Err:      ErrNotFound,
			},
			expected: "stream stream-2: get: stream not found",
		},
		{
			name: "stream full error",
			err: &StreamError{
				StreamID: "stream-3",
				Op:       "send",
				Err:      ErrFull,
			},
			expected: "stream stream-3: send: stream buffer is full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestStreamError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	streamErr := &StreamError{
		StreamID: "stream-1",
		Op:       "test",
		Err:      innerErr,
	}

	assert.Equal(t, innerErr, streamErr.Unwrap())
	assert.True(t, errors.Is(streamErr, innerErr))
}

func TestIsStreamClosed(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrClosed directly",
			err:      ErrClosed,
			expected: true,
		},
		{
			name:     "wrapped ErrClosed",
			err:      ErrStreamClosed("stream-1"),
			expected: true,
		},
		{
			name:     "StreamError with ErrClosed",
			err: &StreamError{
				StreamID: "stream-1",
				Op:       "send",
				Err:      ErrClosed,
			},
			expected: true,
		},
		{
			name:     "different error",
			err:      errors.New("other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsStreamClosed(tt.err))
		})
	}
}

func TestIsStreamNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrNotFound directly",
			err:      ErrNotFound,
			expected: true,
		},
		{
			name:     "wrapped ErrNotFound",
			err:      ErrStreamNotFound("stream-1"),
			expected: true,
		},
		{
			name:     "StreamError with ErrNotFound",
			err: &StreamError{
				StreamID: "stream-1",
				Op:       "get",
				Err:      ErrNotFound,
			},
			expected: true,
		},
		{
			name:     "different error",
			err:      errors.New("other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsStreamNotFound(tt.err))
		})
	}
}

func TestIsStreamFull(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrFull directly",
			err:      ErrFull,
			expected: true,
		},
		{
			name:     "wrapped ErrFull",
			err:      ErrStreamFull("stream-1", 100),
			expected: true,
		},
		{
			name:     "StreamError with ErrFull",
			err: &StreamError{
				StreamID: "stream-1",
				Op:       "send",
				Err:      ErrFull,
			},
			expected: true,
		},
		{
			name:     "different error",
			err:      errors.New("other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsStreamFull(tt.err))
		})
	}
}

func TestIsInvalidEvent(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrInvalid directly",
			err:      ErrInvalid,
			expected: true,
		},
		{
			name:     "wrapped ErrInvalid",
			err:      ErrInvalidEvent("missing field"),
			expected: true,
		},
		{
			name:     "different error",
			err:      errors.New("other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsInvalidEvent(tt.err))
		})
	}
}
