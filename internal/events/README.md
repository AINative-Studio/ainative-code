# Events Package

Real-time event streaming system for processing LLM responses with buffering, backpressure handling, and concurrent stream management.

## Overview

The events package provides a robust, thread-safe event streaming system designed for real-time LLM response processing. It supports:

- **Buffered Event Streams**: Configurable buffer sizes with automatic backpressure handling
- **Multiple Event Types**: Text deltas, content blocks, messages, usage stats, thinking, and errors
- **Stream Management**: Concurrent management of multiple event streams
- **Context Awareness**: Support for context cancellation and timeouts
- **Error Handling**: Comprehensive error types with sentinel errors
- **JSON Serialization**: Full support for event marshaling/unmarshaling

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      StreamManager                           │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  streams: map[string]*managedStream                   │  │
│  │  - session-1 -> EventStream (buffer: 100)            │  │
│  │  - session-2 -> EventStream (buffer: 100)            │  │
│  │  - session-3 -> EventStream (buffer: 100)            │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                               │
│  Operations:                                                  │
│  - CreateStream(id) -> EventStream                           │
│  - GetStream(id) -> EventStream                              │
│  - GetOrCreate(id) -> EventStream, created                   │
│  - CloseStream(id)                                           │
│  - CleanupInactive(threshold)                                │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ manages
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      EventStream                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  events: chan *Event (buffered)                       │  │
│  │  backpressurePolicy: Block | Drop                     │  │
│  │  closed: bool                                         │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                               │
│  Operations:                                                  │
│  - Send(event) -> error                                      │
│  - SendWithContext(ctx, event) -> error                      │
│  - Receive() -> <-chan Event                                 │
│  - Close() -> error                                          │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ transports
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                         Event                                │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Type: EventType                                      │  │
│  │  Data: map[string]interface{}                         │  │
│  │  Timestamp: time.Time                                 │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                               │
│  Event Types:                                                │
│  - EventTextDelta      - Incremental text chunks            │
│  - EventContentStart   - Content block start                │
│  - EventContentEnd     - Content block end                  │
│  - EventMessageStart   - Message start                      │
│  - EventMessageStop    - Message stop                       │
│  - EventError          - Error events                       │
│  - EventUsage          - Token usage stats                  │
│  - EventThinking       - Extended thinking                  │
└─────────────────────────────────────────────────────────────┘
```

## Event Flow Diagram

```
LLM Provider                EventStream              Client Application
     │                           │                            │
     │──── MessageStart ────────>│                            │
     │                           │────── Receive() ──────────>│
     │                           │                            │
     │──── ContentStart ────────>│                            │
     │                           │────── Receive() ──────────>│
     │                           │                            │
     │──── TextDelta("The") ────>│                            │
     │                           │────── Receive() ──────────>│
     │                           │                            │
     │──── TextDelta("answer")──>│                            │
     │                           │────── Receive() ──────────>│
     │                           │                            │
     │──── TextDelta("is") ─────>│                            │
     │                           │────── Receive() ──────────>│
     │                           │                            │
     │──── TextDelta("42") ─────>│                            │
     │                           │────── Receive() ──────────>│
     │                           │                            │
     │──── ContentEnd ──────────>│                            │
     │                           │────── Receive() ──────────>│
     │                           │                            │
     │──── Usage(10,4,14) ──────>│                            │
     │                           │────── Receive() ──────────>│
     │                           │                            │
     │──── MessageStop ─────────>│                            │
     │                           │────── Receive() ──────────>│
     │                           │                            │
     │                           │<──── Close() ───────────────│
     │                           │                            │
     │                           │ (channel closed)           │
     │                           │                            │
```

## Usage Examples

### Basic Event Streaming

```go
// Create a new event stream
stream := events.NewEventStream(100)
defer stream.Close()

// Send events
stream.Send(events.MessageStartEvent("msg_123"))
stream.Send(events.TextDeltaEvent("Hello "))
stream.Send(events.TextDeltaEvent("World!"))
stream.Send(events.MessageStopEvent("msg_123", "end_turn"))

// Receive events
for event := range stream.Receive() {
    switch event.Type {
    case events.EventTextDelta:
        fmt.Print(event.Data["text"])
    case events.EventMessageStop:
        fmt.Println("\nDone!")
    }
}
```

### Context-Aware Streaming

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

stream := events.NewEventStream(10)
defer stream.Close()

// Send with context
err := stream.SendWithContext(ctx, events.TextDeltaEvent("test"))
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Timeout!")
    }
}
```

### Managing Multiple Streams

```go
manager := events.NewStreamManager(100)
defer manager.CloseAll()

// Create streams for different sessions
stream1, _ := manager.CreateStream("session-user-123")
stream2, _ := manager.CreateStream("session-user-456")

// Send events to specific streams
stream1.Send(events.TextDeltaEvent("Hello user 123"))
stream2.Send(events.TextDeltaEvent("Hello user 456"))

// Get or create pattern
stream, created, _ := manager.GetOrCreate("session-user-789")
if created {
    fmt.Println("New stream created")
}
```

### Backpressure Handling

```go
stream := events.NewEventStream(5) // Small buffer
defer stream.Close()

// Set policy to drop events when buffer is full
stream.SetBackpressurePolicy(events.BackpressureDrop)

err := stream.Send(event)
if events.IsStreamFull(err) {
    fmt.Println("Buffer full, event dropped")
}

// Alternative: Block until space is available (default)
stream.SetBackpressurePolicy(events.BackpressureBlock)
```

### Error Handling

```go
// Check for specific errors
err := manager.GetStream("nonexistent")
if events.IsStreamNotFound(err) {
    fmt.Println("Stream not found")
}

err = stream.Send(event)
if events.IsStreamClosed(err) {
    fmt.Println("Stream is closed")
}

if events.IsStreamFull(err) {
    fmt.Println("Buffer is full")
}

if events.IsInvalidEvent(err) {
    fmt.Println("Invalid event")
}
```

### JSON Serialization

```go
// Create event
event := events.UsageEvent(100, 50, 150)

// Marshal to JSON
jsonData, _ := json.Marshal(event)
// Output: {"type":"Usage","data":{"prompt_tokens":100,...},"timestamp":"..."}

// Unmarshal from JSON
var newEvent events.Event
json.Unmarshal(jsonData, &newEvent)
```

## Event Types

### EventTextDelta
Incremental text chunks from the LLM.

```go
event := events.TextDeltaEvent("Hello ")
// Data: {"text": "Hello "}
```

### EventContentStart / EventContentEnd
Marks the beginning/end of a content block.

```go
startEvent := events.ContentStartEvent(0)
// Data: {"index": 0}

endEvent := events.ContentEndEvent(0)
// Data: {"index": 0}
```

### EventMessageStart / EventMessageStop
Marks the start/end of a message.

```go
startEvent := events.MessageStartEvent("msg_123")
// Data: {"message_id": "msg_123"}

stopEvent := events.MessageStopEvent("msg_123", "end_turn")
// Data: {"message_id": "msg_123", "stop_reason": "end_turn"}
```

### EventError
Error events during streaming.

```go
event := events.ErrorEvent("Connection timeout")
// Data: {"error": "Connection timeout"}
```

### EventUsage
Token usage statistics.

```go
event := events.UsageEvent(100, 50, 150)
// Data: {
//   "prompt_tokens": 100,
//   "completion_tokens": 50,
//   "total_tokens": 150
// }
```

### EventThinking
Extended thinking events (e.g., Claude's reasoning process).

```go
event := events.ThinkingEvent("Analyzing the problem...")
// Data: {"thinking": "Analyzing the problem..."}
```

## Error Types

### Sentinel Errors

- `ErrClosed` - Stream is closed
- `ErrNotFound` - Stream not found
- `ErrFull` - Stream buffer is full (backpressure)
- `ErrInvalid` - Invalid event

### Helper Functions

```go
ErrStreamClosed(streamID) error
ErrStreamNotFound(streamID) error
ErrStreamFull(streamID, bufferSize) error
ErrInvalidEvent(reason) error
```

### Error Checking

```go
IsStreamClosed(err) bool
IsStreamNotFound(err) bool
IsStreamFull(err) bool
IsInvalidEvent(err) bool
```

## Best Practices

### 1. Always Close Streams

```go
stream := events.NewEventStream(100)
defer stream.Close() // Always defer close
```

### 2. Use Context for Long-Running Operations

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := stream.SendWithContext(ctx, event)
```

### 3. Handle Backpressure Appropriately

```go
// For real-time UI updates: use Drop policy
stream.SetBackpressurePolicy(events.BackpressureDrop)

// For critical events: use Block policy (default)
stream.SetBackpressurePolicy(events.BackpressureBlock)
```

### 4. Clean Up Inactive Streams

```go
manager := events.NewStreamManager(100)

// Periodically cleanup
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for range ticker.C {
        cleaned := manager.CleanupInactive(24 * time.Hour)
        log.Printf("Cleaned up %d inactive streams", cleaned)
    }
}()
```

### 5. Use GetOrCreate for Idempotent Operations

```go
stream, created, err := manager.GetOrCreate(sessionID)
if err != nil {
    return err
}

if created {
    log.Printf("Created new stream for session: %s", sessionID)
}
```

## Thread Safety

All operations are thread-safe:

- **EventStream**: Uses `sync.RWMutex` for state protection
- **StreamManager**: Uses `sync.RWMutex` for map protection
- **Channels**: Go channels provide inherent thread safety

## Performance Considerations

### Buffer Sizing

- **Small buffers (10-50)**: Lower memory, higher backpressure risk
- **Medium buffers (100-500)**: Balanced for most use cases
- **Large buffers (1000+)**: High throughput, higher memory usage

### Backpressure Policies

- **Block**: Guarantees delivery but may block senders
- **Drop**: Never blocks but may lose events under load

### Cleanup

Regular cleanup of inactive streams prevents memory leaks:

```go
// Cleanup streams inactive for > 1 hour
manager.CleanupInactive(1 * time.Hour)
```

## Testing

Run tests with coverage:

```bash
go test -v -race -cover ./internal/events/
```

Current coverage: **84.3%**

## Integration

The events package integrates with:

- `internal/provider` - LLM provider streaming
- `internal/client` - Client-side event consumption
- `internal/tui` - Terminal UI updates

## License

Part of the AINative-Code project.
