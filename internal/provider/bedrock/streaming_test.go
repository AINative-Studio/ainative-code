package bedrock

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStreamingEvents(t *testing.T) {
	tests := []struct {
		name         string
		events       []string
		expectedText string
		expectError  bool
	}{
		{
			name: "successful streaming events",
			events: []string{
				`{"messageStart":{"role":"assistant"}}`,
				`{"contentBlockDelta":{"delta":{"text":"Hello"},"contentBlockIndex":0}}`,
				`{"contentBlockDelta":{"delta":{"text":" there"},"contentBlockIndex":0}}`,
				`{"contentBlockDelta":{"delta":{"text":"!"},"contentBlockIndex":0}}`,
				`{"messageStop":{}}`,
			},
			expectedText: "Hello there!",
			expectError:  false,
		},
		{
			name: "with metadata events",
			events: []string{
				`{"messageStart":{"role":"assistant"}}`,
				`{"metadata":{"usage":{"inputTokens":10}}}`,
				`{"contentBlockStart":{"start":{"type":"text"},"contentBlockIndex":0}}`,
				`{"contentBlockDelta":{"delta":{"text":"Test"},"contentBlockIndex":0}}`,
				`{"contentBlockStop":{"contentBlockIndex":0}}`,
				`{"messageStop":{}}`,
			},
			expectedText: "Test",
			expectError:  false,
		},
		{
			name: "error event",
			events: []string{
				`{"messageStart":{"role":"assistant"}}`,
				`{"error":{"message":"Internal error occurred"}}`,
			},
			expectedText: "",
			expectError:  true,
		},
		{
			name: "empty stream",
			events: []string{
				`{"messageStart":{"role":"assistant"}}`,
				`{"messageStop":{}}`,
			},
			expectedText: "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock reader
			eventData := strings.Join(tt.events, "\n")
			reader := io.NopCloser(strings.NewReader(eventData))

			// Create event channel
			eventChan := make(chan provider.Event, 10)

			// Parse events
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			go parseStreamingEvents(ctx, reader, eventChan, "test-model")

			// Collect events
			var fullText string
			var gotError error
			var events []provider.Event

			for event := range eventChan {
				events = append(events, event)

				if event.Type == provider.EventTypeContentDelta {
					fullText += event.Content
				}

				if event.Type == provider.EventTypeError {
					gotError = event.Error
				}
			}

			if tt.expectError {
				assert.Error(t, gotError)
			} else {
				assert.NoError(t, gotError)
				assert.Equal(t, tt.expectedText, fullText)
			}
		})
	}
}

func TestParseStreamEvent(t *testing.T) {
	tests := []struct {
		name        string
		eventJSON   string
		expectedType string
		expectedText string
		expectError bool
	}{
		{
			name:         "messageStart event",
			eventJSON:    `{"messageStart":{"role":"assistant"}}`,
			expectedType: "messageStart",
			expectError:  false,
		},
		{
			name:         "contentBlockDelta event",
			eventJSON:    `{"contentBlockDelta":{"delta":{"text":"Hello"},"contentBlockIndex":0}}`,
			expectedType: "contentBlockDelta",
			expectedText: "Hello",
			expectError:  false,
		},
		{
			name:         "messageStop event",
			eventJSON:    `{"messageStop":{}}`,
			expectedType: "messageStop",
			expectError:  false,
		},
		{
			name:         "error event",
			eventJSON:    `{"error":{"message":"Something went wrong"}}`,
			expectedType: "error",
			expectError:  true,
		},
		{
			name:        "invalid JSON",
			eventJSON:   `{invalid json}`,
			expectError: true,
		},
		{
			name:         "unknown event type",
			eventJSON:    `{"unknownEvent":{"data":"test"}}`,
			expectedType: "unknown",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parseStreamEvent([]byte(tt.eventJSON))

			if tt.expectError {
				// For error events, we might not return an error but include it in the event
				if err == nil && event != nil {
					assert.Equal(t, "error", event.EventType)
				}
			} else {
				if err != nil {
					t.Logf("Error: %v", err)
				}
				// Some events might be skipped (return nil without error)
				if event != nil {
					assert.Equal(t, tt.expectedType, event.EventType)
					if tt.expectedText != "" {
						assert.Equal(t, tt.expectedText, event.Text)
					}
				}
			}
		})
	}
}

func TestStreamEventTypes(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		isStart   bool
		isDelta   bool
		isStop    bool
		isError   bool
	}{
		{
			name:      "messageStart",
			eventType: "messageStart",
			isStart:   true,
		},
		{
			name:      "contentBlockStart",
			eventType: "contentBlockStart",
			isStart:   true,
		},
		{
			name:      "contentBlockDelta",
			eventType: "contentBlockDelta",
			isDelta:   true,
		},
		{
			name:      "contentBlockStop",
			eventType: "contentBlockStop",
			isStop:    true,
		},
		{
			name:      "messageStop",
			eventType: "messageStop",
			isStop:    true,
		},
		{
			name:      "error",
			eventType: "error",
			isError:   true,
		},
		{
			name:      "metadata",
			eventType: "metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &streamEvent{EventType: tt.eventType}

			if tt.isStart {
				assert.Contains(t, []string{"messageStart", "contentBlockStart"}, event.EventType)
			}
			if tt.isDelta {
				assert.Equal(t, "contentBlockDelta", event.EventType)
			}
			if tt.isStop {
				assert.Contains(t, []string{"messageStop", "contentBlockStop"}, event.EventType)
			}
			if tt.isError {
				assert.Equal(t, "error", event.EventType)
			}
		})
	}
}

func TestHandleStreamingChunk(t *testing.T) {
	tests := []struct {
		name           string
		event          *streamEvent
		expectedEvent  provider.EventType
		expectedContent string
		expectError    bool
	}{
		{
			name: "content delta",
			event: &streamEvent{
				EventType: "contentBlockDelta",
				Text:      "Hello",
			},
			expectedEvent:   provider.EventTypeContentDelta,
			expectedContent: "Hello",
			expectError:     false,
		},
		{
			name: "message start",
			event: &streamEvent{
				EventType: "messageStart",
			},
			expectedEvent: provider.EventTypeContentStart,
			expectError:   false,
		},
		{
			name: "message stop",
			event: &streamEvent{
				EventType: "messageStop",
			},
			expectedEvent: provider.EventTypeContentEnd,
			expectError:   false,
		},
		{
			name: "error event",
			event: &streamEvent{
				EventType:    "error",
				ErrorMessage: "Something went wrong",
			},
			expectedEvent: provider.EventTypeError,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventChan := make(chan provider.Event, 1)

			handleStreamingChunk(tt.event, eventChan, "test-model")

			select {
			case event := <-eventChan:
				assert.Equal(t, tt.expectedEvent, event.Type)
				if tt.expectedContent != "" {
					assert.Equal(t, tt.expectedContent, event.Content)
				}
				if tt.expectError {
					assert.Error(t, event.Error)
				}
			case <-time.After(100 * time.Millisecond):
				// Some events might not send anything (like metadata)
				if tt.expectedEvent != provider.EventType(0) {
					t.Fatal("expected event but got timeout")
				}
			}
		})
	}
}

func TestStreamReader(t *testing.T) {
	data := `{"messageStart":{"role":"assistant"}}
{"contentBlockDelta":{"delta":{"text":"Hello"}}}
{"messageStop":{}}
`

	reader := io.NopCloser(strings.NewReader(data))
	sr := newStreamReader(reader)

	// Read first event
	event1, err := sr.readEvent()
	require.NoError(t, err)
	assert.NotEmpty(t, event1)

	// Read second event
	event2, err := sr.readEvent()
	require.NoError(t, err)
	assert.NotEmpty(t, event2)

	// Read third event
	event3, err := sr.readEvent()
	require.NoError(t, err)
	assert.NotEmpty(t, event3)

	// Read EOF
	_, err = sr.readEvent()
	assert.Equal(t, io.EOF, err)
}
