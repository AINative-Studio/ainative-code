package events

import (
	"context"
	"fmt"
	"time"
)

// Example_basicEventStream demonstrates basic event stream usage
func Example_basicEventStream() {
	// Create a new event stream with buffer size of 10
	stream := NewEventStream(10)
	defer stream.Close()

	// Send events to the stream
	stream.Send(MessageStartEvent("msg_001"))
	stream.Send(ContentStartEvent(0))
	stream.Send(TextDeltaEvent("Hello "))
	stream.Send(TextDeltaEvent("World!"))
	stream.Send(ContentEndEvent(0))
	stream.Send(UsageEvent(10, 2, 12))
	stream.Send(MessageStopEvent("msg_001", "end_turn"))

	// Receive and process events
	for event := range stream.Receive() {
		fmt.Printf("Event Type: %s\n", event.Type)

		switch event.Type {
		case EventTextDelta:
			fmt.Printf("  Text: %s\n", event.Data["text"])
		case EventUsage:
			fmt.Printf("  Tokens: %v\n", event.Data["total_tokens"])
		}
	}
}

// Example_contextAwareStreaming demonstrates context-aware event streaming
func Example_contextAwareStreaming() {
	stream := NewEventStream(10)
	defer stream.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send events with context awareness
	go func() {
		for i := 0; i < 100; i++ {
			event := TextDeltaEvent(fmt.Sprintf("chunk %d", i))

			err := stream.SendWithContext(ctx, event)
			if err != nil {
				fmt.Printf("Send error: %v\n", err)
				return
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Receive events until context is cancelled
	for {
		select {
		case event, ok := <-stream.Receive():
			if !ok {
				fmt.Println("Stream closed")
				return
			}
			fmt.Printf("Received: %s\n", event.Data["text"])
		case <-ctx.Done():
			fmt.Println("Context cancelled")
			return
		}
	}
}

// Example_backpressureHandling demonstrates backpressure handling
func Example_backpressureHandling() {
	// Create stream with small buffer
	stream := NewEventStream(5)
	defer stream.Close()

	// Set backpressure policy to drop events when buffer is full
	stream.SetBackpressurePolicy(BackpressureDrop)

	// Try to send more events than buffer can hold
	for i := 0; i < 10; i++ {
		event := TextDeltaEvent(fmt.Sprintf("event %d", i))
		err := stream.Send(event)

		if err != nil {
			if IsStreamFull(err) {
				fmt.Printf("Buffer full, event %d dropped\n", i)
			} else {
				fmt.Printf("Error: %v\n", err)
			}
		} else {
			fmt.Printf("Event %d sent successfully\n", i)
		}
	}
}

// Example_streamManager demonstrates managing multiple concurrent streams
func Example_streamManager() {
	// Create a stream manager with default buffer size of 50
	manager := NewStreamManager(50)
	defer manager.CloseAll()

	// Create multiple streams for different sessions
	stream1, _ := manager.CreateStream("session-user-123")
	stream2, _ := manager.CreateStream("session-user-456")
	stream3, _ := manager.CreateStream("session-user-789")

	// Send events to different streams
	stream1.Send(TextDeltaEvent("Hello from user 123"))
	stream2.Send(TextDeltaEvent("Hello from user 456"))
	stream3.Send(TextDeltaEvent("Hello from user 789"))

	// List all active streams
	streams := manager.ListStreams()
	fmt.Printf("Active streams: %d\n", len(streams))

	// Get a specific stream
	if stream, err := manager.GetStream("session-user-123"); err == nil {
		fmt.Printf("Stream found, buffer size: %d\n", stream.BufferSize())
	}

	// Close a specific stream
	manager.CloseStream("session-user-456")

	// Cleanup inactive streams (older than 1 hour)
	cleaned := manager.CleanupInactive(1 * time.Hour)
	fmt.Printf("Cleaned up %d inactive streams\n", cleaned)
}

// Example_errorHandling demonstrates proper error handling
func Example_errorHandling() {
	manager := NewStreamManager(10)
	defer manager.CloseAll()

	// Try to get a non-existent stream
	_, err := manager.GetStream("nonexistent")
	if IsStreamNotFound(err) {
		fmt.Println("Stream not found, creating new one...")
		manager.CreateStream("nonexistent")
	}

	// Try to send to a closed stream
	stream, _ := manager.CreateStream("test-stream")
	stream.Close()

	err = stream.Send(TextDeltaEvent("test"))
	if IsStreamClosed(err) {
		fmt.Println("Cannot send to closed stream")
	}

	// Handle invalid events
	invalidEvent := &Event{
		Type: EventTextDelta,
		Data: nil, // Invalid: data cannot be nil
	}

	err = stream.Send(invalidEvent)
	if IsInvalidEvent(err) {
		fmt.Printf("Invalid event: %v\n", err)
	}
}

// Example_realTimeLLMStreaming demonstrates real-time LLM response streaming
func Example_realTimeLLMStreaming() {
	manager := NewStreamManager(100)
	defer manager.CloseAll()

	sessionID := "llm-session-001"

	// Create stream for LLM session
	stream, _ := manager.CreateStream(sessionID)

	// Simulate LLM provider sending streaming events
	go func() {
		// Start of message
		stream.Send(MessageStartEvent("msg_abc123"))

		// Content block starts
		stream.Send(ContentStartEvent(0))

		// Stream thinking process (for models like Claude with extended thinking)
		stream.Send(ThinkingEvent("Analyzing the user's question..."))

		// Stream text deltas
		words := []string{"The", "answer", "is", "42"}
		for _, word := range words {
			stream.Send(TextDeltaEvent(word + " "))
			time.Sleep(100 * time.Millisecond)
		}

		// Content block ends
		stream.Send(ContentEndEvent(0))

		// Usage statistics
		stream.Send(UsageEvent(15, 4, 19))

		// Message stop
		stream.Send(MessageStopEvent("msg_abc123", "end_turn"))

		// Close the stream
		stream.Close()
	}()

	// Client receives and processes events
	var fullResponse string

	for event := range stream.Receive() {
		switch event.Type {
		case EventMessageStart:
			fmt.Printf("Message started: %s\n", event.Data["message_id"])

		case EventThinking:
			fmt.Printf("Thinking: %s\n", event.Data["thinking"])

		case EventTextDelta:
			text := event.Data["text"].(string)
			fullResponse += text
			fmt.Printf("Delta: %s", text)

		case EventUsage:
			fmt.Printf("\nToken usage - Total: %v\n", event.Data["total_tokens"])

		case EventMessageStop:
			fmt.Printf("Message stopped: %s\n", event.Data["stop_reason"])

		case EventError:
			fmt.Printf("Error: %s\n", event.Data["error"])
		}
	}

	fmt.Printf("\nFull response: %s\n", fullResponse)
}

// Example_getOrCreate demonstrates the GetOrCreate pattern
func Example_getOrCreate() {
	manager := NewStreamManager(10)
	defer manager.CloseAll()

	sessionID := "user-session-xyz"

	// First call creates the stream
	stream1, created1, _ := manager.GetOrCreate(sessionID)
	fmt.Printf("First call - Created: %v\n", created1) // true

	// Second call returns existing stream
	stream2, created2, _ := manager.GetOrCreate(sessionID)
	fmt.Printf("Second call - Created: %v\n", created2) // false

	// Both references point to same stream
	fmt.Printf("Same stream: %v\n", stream1 == stream2) // true
}

// Example_jsonSerialization demonstrates event JSON serialization
func Example_jsonSerialization() {
	// Create an event
	event := UsageEvent(100, 50, 150)

	// Marshal to JSON
	jsonData, err := event.MarshalJSON()
	if err != nil {
		fmt.Printf("Marshal error: %v\n", err)
		return
	}

	fmt.Printf("JSON: %s\n", string(jsonData))

	// Unmarshal from JSON
	var deserializedEvent Event
	err = deserializedEvent.UnmarshalJSON(jsonData)
	if err != nil {
		fmt.Printf("Unmarshal error: %v\n", err)
		return
	}

	fmt.Printf("Type: %s\n", deserializedEvent.Type)
	fmt.Printf("Tokens: %v\n", deserializedEvent.Data["total_tokens"])
}
