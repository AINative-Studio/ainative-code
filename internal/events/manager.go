package events

import (
	"fmt"
	"sync"
	"time"
)

// StreamManager manages multiple concurrent event streams
// It provides thread-safe operations for creating, retrieving, and closing streams
type StreamManager struct {
	// streams maps stream IDs to EventStream instances
	streams map[string]*managedStream

	// defaultBufferSize is the buffer size for newly created streams
	defaultBufferSize int

	// mu protects concurrent access to the streams map
	mu sync.RWMutex
}

// managedStream wraps an EventStream with additional metadata
type managedStream struct {
	stream       *EventStream
	lastActivity time.Time
}

// NewStreamManager creates a new stream manager with the specified default buffer size
func NewStreamManager(bufferSize int) *StreamManager {
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize // Use the package-level default
	}

	return &StreamManager{
		streams:           make(map[string]*managedStream),
		defaultBufferSize: bufferSize,
		mu:                sync.RWMutex{},
	}
}

// CreateStream creates a new event stream with the given ID
// Returns an error if the stream ID is empty or if a stream with the same ID already exists
func (m *StreamManager) CreateStream(streamID string) (*EventStream, error) {
	if streamID == "" {
		return nil, fmt.Errorf("stream ID cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if stream already exists
	if _, exists := m.streams[streamID]; exists {
		return nil, fmt.Errorf("stream %s already exists", streamID)
	}

	// Create new stream
	stream := NewEventStream(m.defaultBufferSize)
	m.streams[streamID] = &managedStream{
		stream:       stream,
		lastActivity: time.Now(),
	}

	return stream, nil
}

// GetStream retrieves an existing stream by ID
// Returns an error if the stream ID is empty or if the stream does not exist
func (m *StreamManager) GetStream(streamID string) (*EventStream, error) {
	if streamID == "" {
		return nil, fmt.Errorf("stream ID cannot be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	managed, exists := m.streams[streamID]
	if !exists {
		return nil, ErrStreamNotFound(streamID)
	}

	// Update last activity
	managed.lastActivity = time.Now()

	return managed.stream, nil
}

// GetOrCreate retrieves an existing stream or creates a new one if it doesn't exist
// Returns the stream, a boolean indicating if it was created, and any error
func (m *StreamManager) GetOrCreate(streamID string) (*EventStream, bool, error) {
	if streamID == "" {
		return nil, false, fmt.Errorf("stream ID cannot be empty")
	}

	// Try to get existing stream first (read lock)
	m.mu.RLock()
	managed, exists := m.streams[streamID]
	m.mu.RUnlock()

	if exists {
		managed.lastActivity = time.Now()
		return managed.stream, false, nil
	}

	// Create new stream (write lock)
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check in case another goroutine created it
	managed, exists = m.streams[streamID]
	if exists {
		managed.lastActivity = time.Now()
		return managed.stream, false, nil
	}

	// Create new stream
	stream := NewEventStream(m.defaultBufferSize)
	m.streams[streamID] = &managedStream{
		stream:       stream,
		lastActivity: time.Now(),
	}

	return stream, true, nil
}

// CloseStream closes and removes a stream by ID
// Returns an error if the stream ID is empty or if the stream does not exist
func (m *StreamManager) CloseStream(streamID string) error {
	if streamID == "" {
		return fmt.Errorf("stream ID cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	managed, exists := m.streams[streamID]
	if !exists {
		return ErrStreamNotFound(streamID)
	}

	// Close the stream
	if err := managed.stream.Close(); err != nil && !IsStreamClosed(err) {
		return fmt.Errorf("failed to close stream %s: %w", streamID, err)
	}

	// Remove from map
	delete(m.streams, streamID)

	return nil
}

// ListStreams returns a list of all active stream IDs
func (m *StreamManager) ListStreams() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	streamIDs := make([]string, 0, len(m.streams))
	for id := range m.streams {
		streamIDs = append(streamIDs, id)
	}

	return streamIDs
}

// StreamCount returns the number of active streams
func (m *StreamManager) StreamCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.streams)
}

// CloseAll closes all active streams and clears the manager
func (m *StreamManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, managed := range m.streams {
		// Best effort close - ignore errors
		managed.stream.Close()
		delete(m.streams, id)
	}
}

// CleanupInactive removes streams that have been inactive for longer than the threshold
// Returns the number of streams cleaned up
func (m *StreamManager) CleanupInactive(threshold time.Duration) int {
	if threshold <= 0 {
		return 0
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	cleaned := 0

	for id, managed := range m.streams {
		if now.Sub(managed.lastActivity) > threshold {
			// Close and remove inactive stream
			managed.stream.Close()
			delete(m.streams, id)
			cleaned++
		}
	}

	return cleaned
}

// StreamInfo contains information about a stream
type StreamInfo struct {
	ID           string
	BufferSize   int
	CurrentLoad  int
	IsClosed     bool
	LastActivity time.Time
}

// GetStreamInfo returns detailed information about a stream
func (m *StreamManager) GetStreamInfo(streamID string) (*StreamInfo, error) {
	if streamID == "" {
		return nil, fmt.Errorf("stream ID cannot be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	managed, exists := m.streams[streamID]
	if !exists {
		return nil, ErrStreamNotFound(streamID)
	}

	return &StreamInfo{
		ID:           streamID,
		BufferSize:   managed.stream.BufferSize(),
		CurrentLoad:  managed.stream.Len(),
		IsClosed:     managed.stream.IsClosed(),
		LastActivity: managed.lastActivity,
	}, nil
}

// ListStreamInfo returns detailed information about all streams
func (m *StreamManager) ListStreamInfo() []*StreamInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info := make([]*StreamInfo, 0, len(m.streams))
	for id, managed := range m.streams {
		info = append(info, &StreamInfo{
			ID:           id,
			BufferSize:   managed.stream.BufferSize(),
			CurrentLoad:  managed.stream.Len(),
			IsClosed:     managed.stream.IsClosed(),
			LastActivity: managed.lastActivity,
		})
	}

	return info
}
