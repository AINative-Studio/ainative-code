package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventType_String(t *testing.T) {
	tests := []struct {
		name     string
		eventType EventType
		expected string
	}{
		{
			name:     "EventTextDelta",
			eventType: EventTextDelta,
			expected: "TextDelta",
		},
		{
			name:     "EventContentStart",
			eventType: EventContentStart,
			expected: "ContentStart",
		},
		{
			name:     "EventContentEnd",
			eventType: EventContentEnd,
			expected: "ContentEnd",
		},
		{
			name:     "EventMessageStart",
			eventType: EventMessageStart,
			expected: "MessageStart",
		},
		{
			name:     "EventMessageStop",
			eventType: EventMessageStop,
			expected: "MessageStop",
		},
		{
			name:     "EventError",
			eventType: EventError,
			expected: "Error",
		},
		{
			name:     "EventUsage",
			eventType: EventUsage,
			expected: "Usage",
		},
		{
			name:     "EventThinking",
			eventType: EventThinking,
			expected: "Thinking",
		},
		{
			name:     "Unknown event type",
			eventType: EventType(999),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.eventType.String())
		})
	}
}

func TestParseEventType(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  EventType
		wantErr   bool
	}{
		{
			name:     "TextDelta",
			input:    "TextDelta",
			expected: EventTextDelta,
			wantErr:  false,
		},
		{
			name:     "ContentStart",
			input:    "ContentStart",
			expected: EventContentStart,
			wantErr:  false,
		},
		{
			name:     "ContentEnd",
			input:    "ContentEnd",
			expected: EventContentEnd,
			wantErr:  false,
		},
		{
			name:     "MessageStart",
			input:    "MessageStart",
			expected: EventMessageStart,
			wantErr:  false,
		},
		{
			name:     "MessageStop",
			input:    "MessageStop",
			expected: EventMessageStop,
			wantErr:  false,
		},
		{
			name:     "Error",
			input:    "Error",
			expected: EventError,
			wantErr:  false,
		},
		{
			name:     "Usage",
			input:    "Usage",
			expected: EventUsage,
			wantErr:  false,
		},
		{
			name:     "Thinking",
			input:    "Thinking",
			expected: EventThinking,
			wantErr:  false,
		},
		{
			name:     "Unknown type",
			input:    "InvalidType",
			expected: EventType(0),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseEventType(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "unknown event type")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid text delta event",
			event: Event{
				Type:      EventTextDelta,
				Data:      map[string]interface{}{"text": "hello"},
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid content start event",
			event: Event{
				Type:      EventContentStart,
				Data:      map[string]interface{}{"index": 0},
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid usage event",
			event: Event{
				Type: EventUsage,
				Data: map[string]interface{}{
					"prompt_tokens":     100,
					"completion_tokens": 50,
					"total_tokens":      150,
				},
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid event - zero timestamp",
			event: Event{
				Type: EventTextDelta,
				Data: map[string]interface{}{"text": "hello"},
			},
			wantErr: true,
			errMsg:  "timestamp cannot be zero",
		},
		{
			name: "invalid event - unknown type",
			event: Event{
				Type:      EventType(999),
				Data:      map[string]interface{}{"text": "hello"},
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "unknown event type",
		},
		{
			name: "invalid event - nil data",
			event: Event{
				Type:      EventTextDelta,
				Data:      nil,
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "data cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEvent_MarshalJSON(t *testing.T) {
	now := time.Now()
	event := Event{
		Type: EventTextDelta,
		Data: map[string]interface{}{
			"text": "hello world",
			"index": 0,
		},
		Timestamp: now,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify JSON structure
	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "TextDelta", decoded["type"])
	assert.NotNil(t, decoded["data"])
	assert.NotNil(t, decoded["timestamp"])
}

func TestEvent_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"type": "TextDelta",
		"data": {
			"text": "hello world",
			"index": 0
		},
		"timestamp": "2024-01-01T12:00:00Z"
	}`

	var event Event
	err := json.Unmarshal([]byte(jsonData), &event)
	require.NoError(t, err)

	assert.Equal(t, EventTextDelta, event.Type)
	assert.NotNil(t, event.Data)
	assert.Equal(t, "hello world", event.Data["text"])
	assert.Equal(t, float64(0), event.Data["index"]) // JSON numbers are float64
	assert.False(t, event.Timestamp.IsZero())
}

func TestEvent_UnmarshalJSON_InvalidType(t *testing.T) {
	jsonData := `{
		"type": "InvalidType",
		"data": {"text": "hello"},
		"timestamp": "2024-01-01T12:00:00Z"
	}`

	var event Event
	err := json.Unmarshal([]byte(jsonData), &event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown event type")
}

func TestEvent_UnmarshalJSON_MalformedJSON(t *testing.T) {
	jsonData := `{
		"type": "TextDelta",
		"data": "not an object"
	}`

	var event Event
	err := json.Unmarshal([]byte(jsonData), &event)
	require.Error(t, err)
}

func TestNewEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		data      map[string]interface{}
		wantErr   bool
	}{
		{
			name:      "valid event",
			eventType: EventTextDelta,
			data:      map[string]interface{}{"text": "hello"},
			wantErr:   false,
		},
		{
			name:      "valid event with complex data",
			eventType: EventUsage,
			data: map[string]interface{}{
				"prompt_tokens":     100,
				"completion_tokens": 50,
			},
			wantErr: false,
		},
		{
			name:      "invalid event - nil data",
			eventType: EventTextDelta,
			data:      nil,
			wantErr:   true,
		},
		{
			name:      "invalid event - unknown type",
			eventType: EventType(999),
			data:      map[string]interface{}{"text": "hello"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := NewEvent(tt.eventType, tt.data)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, event)
			} else {
				require.NoError(t, err)
				require.NotNil(t, event)
				assert.Equal(t, tt.eventType, event.Type)
				assert.Equal(t, tt.data, event.Data)
				assert.False(t, event.Timestamp.IsZero())
			}
		})
	}
}

func TestTextDeltaEvent(t *testing.T) {
	text := "hello world"
	event := TextDeltaEvent(text)

	require.NotNil(t, event)
	assert.Equal(t, EventTextDelta, event.Type)
	assert.Equal(t, text, event.Data["text"])
	assert.False(t, event.Timestamp.IsZero())
	assert.NoError(t, event.Validate())
}

func TestContentStartEvent(t *testing.T) {
	index := 0
	event := ContentStartEvent(index)

	require.NotNil(t, event)
	assert.Equal(t, EventContentStart, event.Type)
	assert.Equal(t, index, event.Data["index"])
	assert.False(t, event.Timestamp.IsZero())
	assert.NoError(t, event.Validate())
}

func TestContentEndEvent(t *testing.T) {
	index := 0
	event := ContentEndEvent(index)

	require.NotNil(t, event)
	assert.Equal(t, EventContentEnd, event.Type)
	assert.Equal(t, index, event.Data["index"])
	assert.False(t, event.Timestamp.IsZero())
	assert.NoError(t, event.Validate())
}

func TestMessageStartEvent(t *testing.T) {
	messageID := "msg_123"
	event := MessageStartEvent(messageID)

	require.NotNil(t, event)
	assert.Equal(t, EventMessageStart, event.Type)
	assert.Equal(t, messageID, event.Data["message_id"])
	assert.False(t, event.Timestamp.IsZero())
	assert.NoError(t, event.Validate())
}

func TestMessageStopEvent(t *testing.T) {
	messageID := "msg_123"
	reason := "end_turn"
	event := MessageStopEvent(messageID, reason)

	require.NotNil(t, event)
	assert.Equal(t, EventMessageStop, event.Type)
	assert.Equal(t, messageID, event.Data["message_id"])
	assert.Equal(t, reason, event.Data["stop_reason"])
	assert.False(t, event.Timestamp.IsZero())
	assert.NoError(t, event.Validate())
}

func TestErrorEvent(t *testing.T) {
	errMsg := "something went wrong"
	event := ErrorEvent(errMsg)

	require.NotNil(t, event)
	assert.Equal(t, EventError, event.Type)
	assert.Equal(t, errMsg, event.Data["error"])
	assert.False(t, event.Timestamp.IsZero())
	assert.NoError(t, event.Validate())
}

func TestUsageEvent(t *testing.T) {
	promptTokens := 100
	completionTokens := 50
	totalTokens := 150
	event := UsageEvent(promptTokens, completionTokens, totalTokens)

	require.NotNil(t, event)
	assert.Equal(t, EventUsage, event.Type)
	assert.Equal(t, promptTokens, event.Data["prompt_tokens"])
	assert.Equal(t, completionTokens, event.Data["completion_tokens"])
	assert.Equal(t, totalTokens, event.Data["total_tokens"])
	assert.False(t, event.Timestamp.IsZero())
	assert.NoError(t, event.Validate())
}

func TestThinkingEvent(t *testing.T) {
	thinkingText := "analyzing the problem..."
	event := ThinkingEvent(thinkingText)

	require.NotNil(t, event)
	assert.Equal(t, EventThinking, event.Type)
	assert.Equal(t, thinkingText, event.Data["thinking"])
	assert.False(t, event.Timestamp.IsZero())
	assert.NoError(t, event.Validate())
}
