package events

import (
	"context"
	"fmt"
	"sync"
)

const (
	// defaultBufferSize is the default size for the event buffer
	defaultBufferSize = 100
)

// BackpressurePolicy defines how the stream handles backpressure when the buffer is full
type BackpressurePolicy int

const (
	// BackpressureBlock blocks the sender until space is available
	BackpressureBlock BackpressurePolicy = iota

	// BackpressureDrop drops the event and returns an error
	BackpressureDrop
)

// EventStream manages a stream of events with buffering and backpressure handling
// It provides thread-safe operations for sending and receiving events
type EventStream struct {
	// events is the buffered channel for event delivery
	events chan *Event

	// bufferSize is the capacity of the events channel
	bufferSize int

	// closed indicates whether the stream has been closed
	closed bool

	// backpressurePolicy determines how to handle buffer overflow
	backpressurePolicy BackpressurePolicy

	// mu protects concurrent access to stream state
	mu sync.RWMutex
}

// NewEventStream creates a new event stream with the specified buffer size
// If bufferSize is 0, defaultBufferSize is used
func NewEventStream(bufferSize int) *EventStream {
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}

	return &EventStream{
		events:             make(chan *Event, bufferSize),
		bufferSize:         bufferSize,
		closed:             false,
		backpressurePolicy: BackpressureBlock,
		mu:                 sync.RWMutex{},
	}
}

// Send sends an event to the stream
// Returns an error if the stream is closed or if the event is invalid
// Behavior when buffer is full depends on the backpressure policy
func (s *EventStream) Send(event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Validate event
	if err := event.Validate(); err != nil {
		return ErrInvalidEvent(err.Error())
	}

	s.mu.RLock()
	closed := s.closed
	policy := s.backpressurePolicy
	s.mu.RUnlock()

	if closed {
		return ErrStreamClosed("stream")
	}

	// Handle backpressure based on policy
	if policy == BackpressureDrop {
		// Non-blocking send
		select {
		case s.events <- event:
			return nil
		default:
			return ErrStreamFull("stream", s.bufferSize)
		}
	}

	// BackpressureBlock - blocking send
	s.events <- event
	return nil
}

// SendWithContext sends an event to the stream with context support
// Returns an error if the context is cancelled, the stream is closed, or the event is invalid
func (s *EventStream) SendWithContext(ctx context.Context, event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Validate event
	if err := event.Validate(); err != nil {
		return ErrInvalidEvent(err.Error())
	}

	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()

	if closed {
		return ErrStreamClosed("stream")
	}

	// Send with context awareness
	select {
	case s.events <- event:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("failed to send event: %w", ctx.Err())
	}
}

// Receive returns the receive-only channel for consuming events
// The channel will be closed when the stream is closed
func (s *EventStream) Receive() <-chan *Event {
	return s.events
}

// Close gracefully shuts down the stream
// It closes the event channel, allowing consumers to drain remaining events
// Returns an error if the stream is already closed
func (s *EventStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStreamClosed("stream")
	}

	s.closed = true
	close(s.events)

	return nil
}

// IsClosed returns true if the stream has been closed
func (s *EventStream) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

// SetBackpressurePolicy sets the backpressure handling policy
// This should be called before sending events
func (s *EventStream) SetBackpressurePolicy(policy BackpressurePolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backpressurePolicy = policy
}

// BufferSize returns the capacity of the event buffer
func (s *EventStream) BufferSize() int {
	return s.bufferSize
}

// Len returns the current number of events in the buffer
func (s *EventStream) Len() int {
	return len(s.events)
}
