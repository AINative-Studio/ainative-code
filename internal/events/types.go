package events

import (
	"encoding/json"
	"fmt"
	"time"
)

// EventType represents the type of streaming event
type EventType int

const (
	// EventTextDelta represents incremental text chunks from the LLM
	EventTextDelta EventType = iota

	// EventContentStart marks the beginning of a content block
	EventContentStart

	// EventContentEnd marks the end of a content block
	EventContentEnd

	// EventMessageStart marks the start of a message
	EventMessageStart

	// EventMessageStop marks the end of a message
	EventMessageStop

	// EventError represents an error event
	EventError

	// EventUsage represents token usage statistics
	EventUsage

	// EventThinking represents extended thinking events (e.g., Claude's thinking process)
	EventThinking
)

// String returns the string representation of EventType
func (e EventType) String() string {
	switch e {
	case EventTextDelta:
		return "TextDelta"
	case EventContentStart:
		return "ContentStart"
	case EventContentEnd:
		return "ContentEnd"
	case EventMessageStart:
		return "MessageStart"
	case EventMessageStop:
		return "MessageStop"
	case EventError:
		return "Error"
	case EventUsage:
		return "Usage"
	case EventThinking:
		return "Thinking"
	default:
		return "Unknown"
	}
}

// ParseEventType converts a string to EventType
func ParseEventType(s string) (EventType, error) {
	switch s {
	case "TextDelta":
		return EventTextDelta, nil
	case "ContentStart":
		return EventContentStart, nil
	case "ContentEnd":
		return EventContentEnd, nil
	case "MessageStart":
		return EventMessageStart, nil
	case "MessageStop":
		return EventMessageStop, nil
	case "Error":
		return EventError, nil
	case "Usage":
		return EventUsage, nil
	case "Thinking":
		return EventThinking, nil
	default:
		return 0, fmt.Errorf("unknown event type: %s", s)
	}
}

// Event represents a streaming event from an LLM provider
// It contains type information, event data, and timestamp
type Event struct {
	// Type specifies the kind of event (TextDelta, ContentStart, etc.)
	Type EventType

	// Data contains the event-specific payload as a flexible map
	// The structure depends on the event type
	Data map[string]interface{}

	// Timestamp records when the event was created
	Timestamp time.Time
}

// Validate checks if the event is valid
func (e *Event) Validate() error {
	// Check timestamp
	if e.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}

	// Check event type
	if e.Type.String() == "Unknown" {
		return fmt.Errorf("unknown event type: %d", e.Type)
	}

	// Check data
	if e.Data == nil {
		return fmt.Errorf("data cannot be nil")
	}

	return nil
}

// MarshalJSON implements json.Marshaler interface for Event
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias struct {
		Type      string                 `json:"type"`
		Data      map[string]interface{} `json:"data"`
		Timestamp time.Time              `json:"timestamp"`
	}

	return json.Marshal(Alias{
		Type:      e.Type.String(),
		Data:      e.Data,
		Timestamp: e.Timestamp,
	})
}

// UnmarshalJSON implements json.Unmarshaler interface for Event
func (e *Event) UnmarshalJSON(data []byte) error {
	type Alias struct {
		Type      string                 `json:"type"`
		Data      map[string]interface{} `json:"data"`
		Timestamp time.Time              `json:"timestamp"`
	}

	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Parse event type
	eventType, err := ParseEventType(alias.Type)
	if err != nil {
		return err
	}

	e.Type = eventType
	e.Data = alias.Data
	e.Timestamp = alias.Timestamp

	return nil
}

// NewEvent creates a new event with the given type and data
// It automatically sets the timestamp to the current time
func NewEvent(eventType EventType, data map[string]interface{}) (*Event, error) {
	event := &Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	if err := event.Validate(); err != nil {
		return nil, fmt.Errorf("invalid event: %w", err)
	}

	return event, nil
}

// TextDeltaEvent creates a text delta event with the given text
func TextDeltaEvent(text string) *Event {
	return &Event{
		Type: EventTextDelta,
		Data: map[string]interface{}{
			"text": text,
		},
		Timestamp: time.Now(),
	}
}

// ContentStartEvent creates a content start event with the given index
func ContentStartEvent(index int) *Event {
	return &Event{
		Type: EventContentStart,
		Data: map[string]interface{}{
			"index": index,
		},
		Timestamp: time.Now(),
	}
}

// ContentEndEvent creates a content end event with the given index
func ContentEndEvent(index int) *Event {
	return &Event{
		Type: EventContentEnd,
		Data: map[string]interface{}{
			"index": index,
		},
		Timestamp: time.Now(),
	}
}

// MessageStartEvent creates a message start event with the given message ID
func MessageStartEvent(messageID string) *Event {
	return &Event{
		Type: EventMessageStart,
		Data: map[string]interface{}{
			"message_id": messageID,
		},
		Timestamp: time.Now(),
	}
}

// MessageStopEvent creates a message stop event with the given message ID and reason
func MessageStopEvent(messageID, stopReason string) *Event {
	return &Event{
		Type: EventMessageStop,
		Data: map[string]interface{}{
			"message_id":  messageID,
			"stop_reason": stopReason,
		},
		Timestamp: time.Now(),
	}
}

// ErrorEvent creates an error event with the given error message
func ErrorEvent(errorMsg string) *Event {
	return &Event{
		Type: EventError,
		Data: map[string]interface{}{
			"error": errorMsg,
		},
		Timestamp: time.Now(),
	}
}

// UsageEvent creates a usage event with token usage statistics
func UsageEvent(promptTokens, completionTokens, totalTokens int) *Event {
	return &Event{
		Type: EventUsage,
		Data: map[string]interface{}{
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
			"total_tokens":      totalTokens,
		},
		Timestamp: time.Now(),
	}
}

// ThinkingEvent creates a thinking event with extended thinking text
func ThinkingEvent(thinking string) *Event {
	return &Event{
		Type: EventThinking,
		Data: map[string]interface{}{
			"thinking": thinking,
		},
		Timestamp: time.Now(),
	}
}
