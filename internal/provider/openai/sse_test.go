package openai

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSEReader_ReadEvent(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		validate    func(t *testing.T, event *streamEvent)
	}{
		{
			name: "simple data event",
			input: `data: {"test": "value"}

`,
			expectError: false,
			validate: func(t *testing.T, event *streamEvent) {
				assert.Equal(t, `{"test": "value"}`, event.data)
			},
		},
		{
			name: "event with type",
			input: `event: message
data: {"content": "hello"}

`,
			expectError: false,
			validate: func(t *testing.T, event *streamEvent) {
				assert.Equal(t, "message", event.eventType)
				assert.Equal(t, `{"content": "hello"}`, event.data)
			},
		},
		{
			name: "multiline data",
			input: `data: line 1
data: line 2
data: line 3

`,
			expectError: false,
			validate: func(t *testing.T, event *streamEvent) {
				assert.Equal(t, "line 1\nline 2\nline 3", event.data)
			},
		},
		{
			name: "event with comment",
			input: `: this is a comment
data: test

`,
			expectError: false,
			validate: func(t *testing.T, event *streamEvent) {
				assert.Equal(t, "test", event.data)
			},
		},
		{
			name: "done marker",
			input: `data: [DONE]

`,
			expectError: false,
			validate: func(t *testing.T, event *streamEvent) {
				assert.Equal(t, "[DONE]", event.data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := newSSEReader(strings.NewReader(tt.input))
			event, err := reader.readEvent()

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, event)
				if tt.validate != nil {
					tt.validate(t, event)
				}
			}
		})
	}
}

func TestSSEReader_MultipleEvents(t *testing.T) {
	input := `data: event 1

data: event 2

data: event 3

`

	reader := newSSEReader(strings.NewReader(input))

	// Read first event
	event1, err := reader.readEvent()
	require.NoError(t, err)
	assert.Equal(t, "event 1", event1.data)

	// Read second event
	event2, err := reader.readEvent()
	require.NoError(t, err)
	assert.Equal(t, "event 2", event2.data)

	// Read third event
	event3, err := reader.readEvent()
	require.NoError(t, err)
	assert.Equal(t, "event 3", event3.data)

	// EOF
	_, err = reader.readEvent()
	assert.Equal(t, io.EOF, err)
}

func TestSSEReader_EmptyStream(t *testing.T) {
	reader := newSSEReader(strings.NewReader(""))
	_, err := reader.readEvent()
	assert.Equal(t, io.EOF, err)
}

func TestSSEReader_OnlyComments(t *testing.T) {
	input := `: comment 1
: comment 2

`

	reader := newSSEReader(strings.NewReader(input))
	_, err := reader.readEvent()
	assert.Equal(t, io.EOF, err)
}

func TestSSEReader_RealWorldOpenAIStream(t *testing.T) {
	// Simulate actual OpenAI streaming response
	input := `data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":" there"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]

`

	reader := newSSEReader(strings.NewReader(input))
	var events []*streamEvent

	for {
		event, err := reader.readEvent()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		events = append(events, event)
	}

	assert.Len(t, events, 5)
	assert.Equal(t, "[DONE]", events[len(events)-1].data)
}
