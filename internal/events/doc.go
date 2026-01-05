// Package events provides a real-time event streaming system for processing LLM responses.
//
// The events package implements a robust, thread-safe event streaming infrastructure
// designed for real-time LLM response processing. It supports buffered event streams,
// backpressure handling, concurrent stream management, and context-aware operations.
//
// # Core Components
//
// The package consists of four main components:
//
// 1. Event Types (types.go) - Defines event structures and event types
// 2. EventStream (stream.go) - Manages individual event streams with buffering
// 3. StreamManager (manager.go) - Manages multiple concurrent streams
// 4. Error Types (errors.go) - Provides comprehensive error handling
//
// # Event Types
//
// The package supports eight event types for LLM streaming:
//
//   - EventTextDelta: Incremental text chunks from the LLM
//   - EventContentStart: Beginning of a content block
//   - EventContentEnd: End of a content block
//   - EventMessageStart: Start of a message
//   - EventMessageStop: End of a message
//   - EventError: Error events during streaming
//   - EventUsage: Token usage statistics
//   - EventThinking: Extended thinking events (e.g., Claude's reasoning)
//
// # Basic Usage
//
// Create and use a simple event stream:
//
//	stream := events.NewEventStream(100)
//	defer stream.Close()
//
//	// Send events
//	stream.Send(events.TextDeltaEvent("Hello "))
//	stream.Send(events.TextDeltaEvent("World!"))
//
//	// Receive events
//	for event := range stream.Receive() {
//	    fmt.Print(event.Data["text"])
//	}
//
// # Stream Management
//
// Manage multiple concurrent streams:
//
//	manager := events.NewStreamManager(100)
//	defer manager.CloseAll()
//
//	// Create streams for different sessions
//	stream1, _ := manager.CreateStream("session-1")
//	stream2, _ := manager.CreateStream("session-2")
//
//	// Get or create pattern
//	stream, created, _ := manager.GetOrCreate("session-3")
//
// # Context Awareness
//
// Support for context cancellation and timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	err := stream.SendWithContext(ctx, event)
//	if err != nil {
//	    if errors.Is(err, context.DeadlineExceeded) {
//	        // Handle timeout
//	    }
//	}
//
// # Backpressure Handling
//
// Two backpressure policies are supported:
//
//	// Block until space is available (default)
//	stream.SetBackpressurePolicy(events.BackpressureBlock)
//
//	// Drop events when buffer is full
//	stream.SetBackpressurePolicy(events.BackpressureDrop)
//
// # Error Handling
//
// Comprehensive error types with sentinel errors:
//
//	err := manager.GetStream("nonexistent")
//	if events.IsStreamNotFound(err) {
//	    // Handle stream not found
//	}
//
//	err = stream.Send(event)
//	if events.IsStreamClosed(err) {
//	    // Handle closed stream
//	}
//
//	if events.IsStreamFull(err) {
//	    // Handle backpressure
//	}
//
// # Thread Safety
//
// All operations are thread-safe. The package uses sync.RWMutex for state
// protection and Go channels for event delivery, ensuring safe concurrent access.
//
// # Performance
//
// Buffer sizes affect performance and memory usage:
//
//   - Small buffers (10-50): Lower memory, higher backpressure risk
//   - Medium buffers (100-500): Balanced for most use cases
//   - Large buffers (1000+): High throughput, higher memory usage
//
// # JSON Serialization
//
// Events support JSON marshaling and unmarshaling:
//
//	event := events.UsageEvent(100, 50, 150)
//	jsonData, _ := json.Marshal(event)
//
//	var newEvent events.Event
//	json.Unmarshal(jsonData, &newEvent)
//
// # Integration
//
// The events package is designed to integrate with:
//
//   - internal/provider: LLM provider streaming implementations
//   - internal/client: Client-side event consumption
//   - internal/tui: Terminal UI real-time updates
//
// For more examples and detailed documentation, see the README.md and examples.go files.
package events
