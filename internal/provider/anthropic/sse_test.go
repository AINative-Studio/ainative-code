package anthropic

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSEReader_ReadEvent(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedEvents []sseEvent
		expectEOF      bool
	}{
		{
			name: "single event with data",
			input: `event: message_start
data: {"type":"message_start"}

`,
			expectedEvents: []sseEvent{
				{
					eventType: "message_start",
					data:      `{"type":"message_start"}`,
				},
			},
		},
		{
			name: "multiple events",
			input: `event: message_start
data: {"type":"message_start"}

event: content_block_delta
data: {"delta":"text"}

event: message_stop
data: {}

`,
			expectedEvents: []sseEvent{
				{
					eventType: "message_start",
					data:      `{"type":"message_start"}`,
				},
				{
					eventType: "content_block_delta",
					data:      `{"delta":"text"}`,
				},
				{
					eventType: "message_stop",
					data:      `{}`,
				},
			},
		},
		{
			name: "multi-line data",
			input: `event: test
data: {"first":"line",
data: "second":"line"}

`,
			expectedEvents: []sseEvent{
				{
					eventType: "test",
					data:      "{\"first\":\"line\",\n\"second\":\"line\"}",
				},
			},
		},
		{
			name: "data only event (no event type)",
			input: `data: {"test":"data"}

`,
			expectedEvents: []sseEvent{
				{
					eventType: "",
					data:      `{"test":"data"}`,
				},
			},
		},
		{
			name: "event with comment (should be ignored)",
			input: `event: message_start
: this is a comment
data: {"type":"message_start"}

`,
			expectedEvents: []sseEvent{
				{
					eventType: "message_start",
					data:      `{"type":"message_start"}`,
				},
			},
		},
		{
			name: "empty lines between data",
			input: `event: test
data: line1

data: line2

`,
			expectedEvents: []sseEvent{
				{
					eventType: "test",
					data:      "line1",
				},
				{
					eventType: "",
					data:      "line2",
				},
			},
		},
		{
			name: "event with id field (should be ignored)",
			input: `event: message
id: 123
data: {"content":"test"}

`,
			expectedEvents: []sseEvent{
				{
					eventType: "message",
					data:      `{"content":"test"}`,
				},
			},
		},
		{
			name: "event with retry field (should be ignored)",
			input: `event: message
retry: 3000
data: {"content":"test"}

`,
			expectedEvents: []sseEvent{
				{
					eventType: "message",
					data:      `{"content":"test"}`,
				},
			},
		},
		{
			name:      "empty input",
			input:     "",
			expectEOF: true,
		},
		{
			name:      "only empty lines",
			input:     "\n\n\n",
			expectEOF: true,
		},
		{
			name: "data with spaces after colon",
			input: `event:    message_start
data:    {"type":"message_start"}

`,
			expectedEvents: []sseEvent{
				{
					eventType: "message_start",
					data:      `{"type":"message_start"}`,
				},
			},
		},
		{
			name: "real Anthropic SSE example",
			input: `event: message_start
data: {"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","content":[],"model":"claude-3-5-sonnet-20241022"}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"!"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":10}}

event: message_stop
data: {"type":"message_stop"}

`,
			expectedEvents: []sseEvent{
				{
					eventType: "message_start",
					data:      `{"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","content":[],"model":"claude-3-5-sonnet-20241022"}}`,
				},
				{
					eventType: "content_block_start",
					data:      `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
				},
				{
					eventType: "content_block_delta",
					data:      `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`,
				},
				{
					eventType: "content_block_delta",
					data:      `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"!"}}`,
				},
				{
					eventType: "content_block_stop",
					data:      `{"type":"content_block_stop","index":0}`,
				},
				{
					eventType: "message_delta",
					data:      `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":10}}`,
				},
				{
					eventType: "message_stop",
					data:      `{"type":"message_stop"}`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := newSSEReader(strings.NewReader(tt.input))

			for i, expectedEvent := range tt.expectedEvents {
				event, err := reader.readEvent()
				require.NoError(t, err, "Failed to read event %d", i)
				assert.Equal(t, expectedEvent.eventType, event.eventType, "Event %d type mismatch", i)
				assert.Equal(t, expectedEvent.data, event.data, "Event %d data mismatch", i)
			}

			// If we expect EOF, verify next read returns EOF
			if tt.expectEOF {
				event, err := reader.readEvent()
				assert.Equal(t, io.EOF, err)
				assert.Nil(t, event)
			}
		})
	}
}

func TestSSEReader_MultipleReads(t *testing.T) {
	input := `event: first
data: data1

event: second
data: data2

event: third
data: data3

`

	reader := newSSEReader(strings.NewReader(input))

	// Read first event
	event1, err := reader.readEvent()
	require.NoError(t, err)
	assert.Equal(t, "first", event1.eventType)
	assert.Equal(t, "data1", event1.data)

	// Read second event
	event2, err := reader.readEvent()
	require.NoError(t, err)
	assert.Equal(t, "second", event2.eventType)
	assert.Equal(t, "data2", event2.data)

	// Read third event
	event3, err := reader.readEvent()
	require.NoError(t, err)
	assert.Equal(t, "third", event3.eventType)
	assert.Equal(t, "data3", event3.data)

	// Read should return EOF
	event4, err := reader.readEvent()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, event4)
}

func TestSSEReader_IncompleteEvent(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectEvent bool
		expectEOF   bool
	}{
		{
			name: "event type without data",
			input: `event: test

`,
			expectEvent: true,
		},
		{
			name: "data without event type",
			input: `data: test

`,
			expectEvent: true,
		},
		{
			name: "event without final newline",
			input: `event: test
data: test`,
			expectEvent: true,
			expectEOF:   false,
		},
		{
			name: "partial event at EOF",
			input: `event: complete
data: complete_data

event: partial`,
			expectEvent: true, // Should get the complete event
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := newSSEReader(strings.NewReader(tt.input))

			event, err := reader.readEvent()

			if tt.expectEvent {
				assert.NoError(t, err)
				assert.NotNil(t, event)
			} else if tt.expectEOF {
				assert.Equal(t, io.EOF, err)
			}
		})
	}
}

func TestSSEReader_EdgeCases(t *testing.T) {
	t.Run("very long data line", func(t *testing.T) {
		longData := strings.Repeat("a", 10000)
		input := "event: test\ndata: " + longData + "\n\n"

		reader := newSSEReader(strings.NewReader(input))
		event, err := reader.readEvent()

		require.NoError(t, err)
		assert.Equal(t, "test", event.eventType)
		assert.Equal(t, longData, event.data)
	})

	t.Run("many events", func(t *testing.T) {
		var input strings.Builder
		numEvents := 1000

		for i := 0; i < numEvents; i++ {
			input.WriteString("event: test\n")
			input.WriteString("data: data\n\n")
		}

		reader := newSSEReader(strings.NewReader(input.String()))

		for i := 0; i < numEvents; i++ {
			event, err := reader.readEvent()
			require.NoError(t, err)
			assert.Equal(t, "test", event.eventType)
			assert.Equal(t, "data", event.data)
		}

		// Should reach EOF
		event, err := reader.readEvent()
		assert.Equal(t, io.EOF, err)
		assert.Nil(t, event)
	})

	t.Run("mixed line endings", func(t *testing.T) {
		// This tests with only \n endings (Go's standard)
		input := "event: test\ndata: test\n\n"

		reader := newSSEReader(strings.NewReader(input))
		event, err := reader.readEvent()

		require.NoError(t, err)
		assert.Equal(t, "test", event.eventType)
		assert.Equal(t, "test", event.data)
	})

	t.Run("unicode data", func(t *testing.T) {
		unicodeData := "Hello 世界 🌍"
		input := "event: test\ndata: " + unicodeData + "\n\n"

		reader := newSSEReader(strings.NewReader(input))
		event, err := reader.readEvent()

		require.NoError(t, err)
		assert.Equal(t, "test", event.eventType)
		assert.Equal(t, unicodeData, event.data)
	})

	t.Run("field with no value", func(t *testing.T) {
		input := "event:\ndata:\n\n"

		reader := newSSEReader(strings.NewReader(input))
		event, err := reader.readEvent()

		require.NoError(t, err)
		assert.Equal(t, "", event.eventType)
		assert.Equal(t, "", event.data)
	})

	t.Run("field with only spaces", func(t *testing.T) {
		input := "event:   \ndata:   \n\n"

		reader := newSSEReader(strings.NewReader(input))
		event, err := reader.readEvent()

		require.NoError(t, err)
		assert.Equal(t, "", event.eventType) // TrimSpace removes all spaces
		assert.Equal(t, "", event.data)      // TrimSpace removes all spaces
	})
}

func TestNewSSEReader(t *testing.T) {
	input := "test input"
	reader := newSSEReader(strings.NewReader(input))

	assert.NotNil(t, reader)
	assert.NotNil(t, reader.scanner)
}
