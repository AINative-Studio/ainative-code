# Events API Reference

**Import Path**: `github.com/AINative-studio/ainative-code/internal/client` (future), `github.com/AINative-studio/ainative-code/internal/provider`

The events system provides streaming capabilities for real-time responses from LLM providers.

## Table of Contents

- [Provider Events](#provider-events)
- [Event Types](#event-types)
- [Streaming](#streaming)
- [Event Handlers](#event-handlers)
- [Usage Examples](#usage-examples)

## Provider Events

### Event

```go
type Event struct {
    Type    EventType
    Content string
    Error   error
    Done    bool
}
```

Represents a streaming event from a provider.

**Fields**:
- `Type` - The type of event (ContentStart, ContentDelta, ContentEnd, Error)
- `Content` - Incremental content (for ContentDelta events)
- `Error` - Error information (for Error events)
- `Done` - Indicates if this is the final event

### EventType

```go
type EventType int

const (
    EventTypeContentDelta EventType = iota // Incremental content
    EventTypeContentStart                   // Stream started
    EventTypeContentEnd                     // Stream completed
    EventTypeError                          // Error occurred
)
```

**String Method**:

```go
func (e EventType) String() string
```

Returns the string representation of EventType.

**Example**:

```go
switch event.Type {
case provider.EventTypeContentStart:
    fmt.Println("Stream started")
case provider.EventTypeContentDelta:
    fmt.Print(event.Content)  // Print incrementally
case provider.EventTypeContentEnd:
    fmt.Println("\nStream completed")
case provider.EventTypeError:
    log.Printf("Error: %v", event.Error)
}
```

## Event Types

### ContentStart

Sent when a stream begins.

**Properties**:
- `Type`: EventTypeContentStart
- `Content`: Empty string
- `Error`: nil
- `Done`: false

### ContentDelta

Sent for each incremental chunk of content.

**Properties**:
- `Type`: EventTypeContentDelta
- `Content`: The incremental text chunk
- `Error`: nil
- `Done`: false

### ContentEnd

Sent when a stream completes successfully.

**Properties**:
- `Type`: EventTypeContentEnd
- `Content`: Empty string
- `Error`: nil
- `Done`: true

### Error

Sent when an error occurs during streaming.

**Properties**:
- `Type`: EventTypeError
- `Content`: Empty string
- `Error`: The error that occurred
- `Done`: true

## Streaming

### Stream Method

```go
func (p Provider) Stream(ctx context.Context, messages []Message, opts ...StreamOption) (<-chan Event, error)
```

Initiates a streaming request and returns a channel for receiving events.

**Parameters**:
- `ctx` - Context for cancellation
- `messages` - Conversation messages
- `opts` - Stream options (model, temperature, etc.)

**Returns**:
- `<-chan Event` - Read-only event channel
- `error` - Error if stream cannot be initiated

**Example**:

```go
ctx := context.Background()

messages := []provider.Message{
    {Role: "user", Content: "Explain quantum computing"},
}

eventChan, err := provider.Stream(ctx, messages,
    provider.StreamWithModel("claude-3-sonnet-20240229"),
    provider.StreamWithMaxTokens(2048),
)
if err != nil {
    log.Fatalf("Failed to start stream: %v", err)
}

// Process events
for event := range eventChan {
    switch event.Type {
    case provider.EventTypeContentDelta:
        fmt.Print(event.Content)
    case provider.EventTypeError:
        log.Printf("Error: %v", event.Error)
    }

    if event.Done {
        break
    }
}
```

## Event Handlers

### Basic Event Handler

```go
type EventHandler func(event provider.Event)

func processEvents(eventChan <-chan provider.Event, handler EventHandler) {
    for event := range eventChan {
        handler(event)
        if event.Done {
            break
        }
    }
}

// Usage
processEvents(eventChan, func(event provider.Event) {
    if event.Type == provider.EventTypeContentDelta {
        fmt.Print(event.Content)
    }
})
```

### Buffered Event Handler

```go
type BufferedEventHandler struct {
    buffer   strings.Builder
    callback func(string)
}

func NewBufferedEventHandler(callback func(string)) *BufferedEventHandler {
    return &BufferedEventHandler{
        callback: callback,
    }
}

func (h *BufferedEventHandler) Handle(event provider.Event) {
    switch event.Type {
    case provider.EventTypeContentStart:
        h.buffer.Reset()

    case provider.EventTypeContentDelta:
        h.buffer.WriteString(event.Content)

    case provider.EventTypeContentEnd:
        if h.callback != nil {
            h.callback(h.buffer.String())
        }

    case provider.EventTypeError:
        log.Printf("Stream error: %v", event.Error)
    }
}

// Usage
handler := NewBufferedEventHandler(func(fullContent string) {
    fmt.Printf("Full response: %s\n", fullContent)
})

for event := range eventChan {
    handler.Handle(event)
    if event.Done {
        break
    }
}
```

## Usage Examples

### Basic Streaming

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AINative-studio/ainative-code/internal/provider"
)

func main() {
    ctx := context.Background()

    // Get provider
    p, _ := provider.GetRegistry().Get("anthropic")

    // Start stream
    messages := []provider.Message{
        {Role: "user", Content: "Write a haiku about coding"},
    }

    eventChan, err := p.Stream(ctx, messages)
    if err != nil {
        log.Fatal(err)
    }

    // Process events
    fmt.Print("Response: ")
    for event := range eventChan {
        if event.Type == provider.EventTypeContentDelta {
            fmt.Print(event.Content)
        }
        if event.Done {
            fmt.Println()
            break
        }
    }
}
```

### Streaming with Progress Indicator

```go
import "time"

func streamWithProgress(ctx context.Context, p provider.Provider, messages []provider.Message) (string, error) {
    eventChan, err := p.Stream(ctx, messages)
    if err != nil {
        return "", err
    }

    var content strings.Builder
    lastUpdate := time.Now()

    for event := range eventChan {
        switch event.Type {
        case provider.EventTypeContentDelta:
            content.WriteString(event.Content)

            // Update progress every 100ms
            if time.Since(lastUpdate) > 100*time.Millisecond {
                fmt.Printf("\rReceived %d characters...", content.Len())
                lastUpdate = time.Now()
            }

        case provider.EventTypeContentEnd:
            fmt.Printf("\rCompleted: %d characters\n", content.Len())

        case provider.EventTypeError:
            return "", event.Error
        }

        if event.Done {
            break
        }
    }

    return content.String(), nil
}
```

### Streaming to File

```go
import "os"

func streamToFile(ctx context.Context, p provider.Provider, messages []provider.Message, filepath string) error {
    file, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer file.Close()

    eventChan, err := p.Stream(ctx, messages)
    if err != nil {
        return err
    }

    for event := range eventChan {
        switch event.Type {
        case provider.EventTypeContentDelta:
            if _, err := file.WriteString(event.Content); err != nil {
                return err
            }

        case provider.EventTypeError:
            return event.Error
        }

        if event.Done {
            break
        }
    }

    return nil
}
```

### Concurrent Streaming

```go
func streamMultiple(ctx context.Context, p provider.Provider, queries []string) []string {
    var wg sync.WaitGroup
    results := make([]string, len(queries))

    for i, query := range queries {
        wg.Add(1)
        go func(index int, q string) {
            defer wg.Done()

            messages := []provider.Message{
                {Role: "user", Content: q},
            }

            eventChan, err := p.Stream(ctx, messages)
            if err != nil {
                log.Printf("Stream %d failed: %v", index, err)
                return
            }

            var content strings.Builder
            for event := range eventChan {
                if event.Type == provider.EventTypeContentDelta {
                    content.WriteString(event.Content)
                }
                if event.Done {
                    break
                }
            }

            results[index] = content.String()
        }(i, query)
    }

    wg.Wait()
    return results
}
```

### Stream with Timeout

```go
func streamWithTimeout(p provider.Provider, messages []provider.Message, timeout time.Duration) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    eventChan, err := p.Stream(ctx, messages)
    if err != nil {
        return "", err
    }

    var content strings.Builder

    for {
        select {
        case event, ok := <-eventChan:
            if !ok {
                return content.String(), nil
            }

            switch event.Type {
            case provider.EventTypeContentDelta:
                content.WriteString(event.Content)
            case provider.EventTypeError:
                return "", event.Error
            }

            if event.Done {
                return content.String(), nil
            }

        case <-ctx.Done():
            return "", fmt.Errorf("stream timeout after %v", timeout)
        }
    }
}
```

## Best Practices

### 1. Always Handle Context Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

eventChan, _ := provider.Stream(ctx, messages)

// Cancel on interrupt
go func() {
    <-interruptChan
    cancel()
}()
```

### 2. Check for Done Flag

```go
for event := range eventChan {
    // Process event...

    if event.Done {
        break  // Always break on Done
    }
}
```

### 3. Handle All Event Types

```go
for event := range eventChan {
    switch event.Type {
    case provider.EventTypeContentStart:
        // Initialize
    case provider.EventTypeContentDelta:
        // Process content
    case provider.EventTypeContentEnd:
        // Finalize
    case provider.EventTypeError:
        // Handle error
    }

    if event.Done {
        break
    }
}
```

### 4. Use Buffering for Large Responses

```go
var buffer strings.Builder
buffer.Grow(4096)  // Pre-allocate

for event := range eventChan {
    if event.Type == provider.EventTypeContentDelta {
        buffer.WriteString(event.Content)
    }
    if event.Done {
        break
    }
}
```

### 5. Implement Error Recovery

```go
maxRetries := 3
for attempt := 0; attempt < maxRetries; attempt++ {
    eventChan, err := provider.Stream(ctx, messages)
    if err != nil {
        time.Sleep(time.Duration(1<<uint(attempt)) * time.Second)
        continue
    }

    err = processStream(eventChan)
    if err == nil {
        break  // Success
    }
}
```

## Related Documentation

- [Providers](providers.md) - Provider streaming
- [Core Packages](core-packages.md) - Client integration
- [Errors](errors.md) - Error handling
