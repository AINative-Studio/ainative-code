package ollama

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStreamChunk(t *testing.T) {
	tests := []struct {
		name          string
		chunk         string
		expectError   bool
		expectedDone  bool
		expectedMsg   string
		expectedError string
	}{
		{
			name: "content chunk",
			chunk: `{"model":"llama2","created_at":"2024-01-01T00:00:00Z","message":{"role":"assistant","content":"Hello"},"done":false}`,
			expectError: false,
			expectedDone: false,
			expectedMsg: "Hello",
		},
		{
			name: "final chunk with stats",
			chunk: `{"model":"llama2","created_at":"2024-01-01T00:00:00Z","message":{"role":"assistant","content":""},"done":true,"total_duration":1000000,"prompt_eval_count":10,"eval_count":20}`,
			expectError: false,
			expectedDone: true,
		},
		{
			name: "error chunk",
			chunk: `{"error":"model not found"}`,
			expectError: false,
			expectedError: "model not found",
		},
		{
			name: "invalid json",
			chunk: `{invalid json`,
			expectError: true,
		},
		{
			name: "empty chunk",
			chunk: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := parseStreamChunk([]byte(tt.chunk))

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDone, resp.Done)
				if tt.expectedMsg != "" {
					assert.Equal(t, tt.expectedMsg, resp.Message.Content)
				}
				if tt.expectedError != "" {
					assert.Equal(t, tt.expectedError, resp.Error)
				}
			}
		})
	}
}

func TestStreamReader(t *testing.T) {
	t.Run("read multiple chunks", func(t *testing.T) {
		chunks := []string{
			`{"model":"llama2","message":{"role":"assistant","content":"Hello"},"done":false}`,
			`{"model":"llama2","message":{"role":"assistant","content":" world"},"done":false}`,
			`{"model":"llama2","message":{"role":"assistant","content":"!"},"done":true}`,
		}

		body := strings.Join(chunks, "\n")
		reader := newStreamReader(io.NopCloser(strings.NewReader(body)))

		var responses []*ollamaStreamResponse
		for {
			resp, err := reader.readChunk()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			responses = append(responses, resp)
		}

		assert.Len(t, responses, 3)
		assert.Equal(t, "Hello", responses[0].Message.Content)
		assert.Equal(t, " world", responses[1].Message.Content)
		assert.Equal(t, "!", responses[2].Message.Content)
		assert.True(t, responses[2].Done)
	})

	t.Run("read with empty lines", func(t *testing.T) {
		body := `{"model":"llama2","message":{"content":"Hello"},"done":false}

{"model":"llama2","message":{"content":"World"},"done":true}`

		reader := newStreamReader(io.NopCloser(strings.NewReader(body)))

		resp1, err := reader.readChunk()
		require.NoError(t, err)
		assert.Equal(t, "Hello", resp1.Message.Content)

		resp2, err := reader.readChunk()
		require.NoError(t, err)
		assert.Equal(t, "World", resp2.Message.Content)
		assert.True(t, resp2.Done)
	})

	t.Run("read with error chunk", func(t *testing.T) {
		body := `{"error":"something went wrong"}`
		reader := newStreamReader(io.NopCloser(strings.NewReader(body)))

		resp, err := reader.readChunk()
		require.NoError(t, err)
		assert.Equal(t, "something went wrong", resp.Error)
	})
}

func TestHandleStreamResponse(t *testing.T) {
	t.Run("successful stream", func(t *testing.T) {
		chunks := []string{
			`{"model":"llama2","message":{"role":"assistant","content":"Hello"},"done":false}`,
			`{"model":"llama2","message":{"role":"assistant","content":" there"},"done":false}`,
			`{"model":"llama2","message":{"role":"assistant","content":"!"},"done":true,"eval_count":3}`,
		}

		body := strings.Join(chunks, "\n")
		bodyReader := io.NopCloser(strings.NewReader(body))

		ctx := context.Background()
		eventChan := make(chan provider.Event, 10)

		go handleStreamResponse(ctx, bodyReader, eventChan, "llama2")

		var events []provider.Event
		for event := range eventChan {
			events = append(events, event)
		}

		// Should receive: start, 3 deltas, end
		assert.GreaterOrEqual(t, len(events), 4)

		// Check for start event
		assert.Equal(t, provider.EventTypeContentStart, events[0].Type)

		// Check content events
		var content string
		for i := 1; i < len(events)-1; i++ {
			if events[i].Type == provider.EventTypeContentDelta {
				content += events[i].Content
			}
		}
		assert.Equal(t, "Hello there!", content)

		// Check for end event
		lastEvent := events[len(events)-1]
		assert.Equal(t, provider.EventTypeContentEnd, lastEvent.Type)
		assert.True(t, lastEvent.Done)
	})

	t.Run("stream with error", func(t *testing.T) {
		body := `{"error":"model not found"}`
		bodyReader := io.NopCloser(strings.NewReader(body))

		ctx := context.Background()
		eventChan := make(chan provider.Event, 10)

		go handleStreamResponse(ctx, bodyReader, eventChan, "unknown")

		var events []provider.Event
		for event := range eventChan {
			events = append(events, event)
		}

		// Should receive start and error
		assert.GreaterOrEqual(t, len(events), 1)

		// Find error event
		var foundError bool
		for _, event := range events {
			if event.Type == provider.EventTypeError {
				foundError = true
				assert.Error(t, event.Error)
				break
			}
		}
		assert.True(t, foundError)
	})

	t.Run("stream with cancelled context", func(t *testing.T) {
		// Create a slow reader that simulates streaming
		slowReader := &slowReader{
			chunks: []string{
				`{"model":"llama2","message":{"content":"Hello"},"done":false}`,
				`{"model":"llama2","message":{"content":"World"},"done":true}`,
			},
			delay: 100 * time.Millisecond,
		}

		ctx, cancel := context.WithCancel(context.Background())
		eventChan := make(chan provider.Event, 10)

		go handleStreamResponse(ctx, io.NopCloser(slowReader), eventChan, "llama2")

		// Cancel after short delay
		time.Sleep(10 * time.Millisecond)
		cancel()

		// Collect events
		var events []provider.Event
		for event := range eventChan {
			events = append(events, event)
			if event.Type == provider.EventTypeError && event.Error != nil {
				break
			}
		}

		// Should have received an error due to cancellation
		var foundCancelError bool
		for _, event := range events {
			if event.Type == provider.EventTypeError {
				foundCancelError = true
				break
			}
		}
		assert.True(t, foundCancelError)
	})
}

// slowReader simulates a slow streaming response
type slowReader struct {
	chunks []string
	pos    int
	delay  time.Duration
	buf    *bufio.Reader
}

func (r *slowReader) Read(p []byte) (n int, err error) {
	if r.buf == nil {
		if r.pos >= len(r.chunks) {
			return 0, io.EOF
		}
		time.Sleep(r.delay)
		chunk := r.chunks[r.pos] + "\n"
		r.pos++
		r.buf = bufio.NewReader(bytes.NewReader([]byte(chunk)))
	}

	n, err = r.buf.Read(p)
	if err == io.EOF && r.pos < len(r.chunks) {
		r.buf = nil
		return n, nil
	}
	return n, err
}

func (r *slowReader) Close() error {
	return nil
}

func TestConvertUsageStats(t *testing.T) {
	resp := &ollamaStreamResponse{
		PromptEvalCount: 10,
		EvalCount:       20,
	}

	usage := convertUsageStats(resp)

	assert.Equal(t, 10, usage.PromptTokens)
	assert.Equal(t, 20, usage.CompletionTokens)
	assert.Equal(t, 30, usage.TotalTokens)
}

func TestNewStreamReader(t *testing.T) {
	body := `{"model":"llama2","message":{"content":"test"},"done":true}`
	reader := newStreamReader(io.NopCloser(strings.NewReader(body)))

	assert.NotNil(t, reader)
	assert.NotNil(t, reader.scanner)
	assert.NotNil(t, reader.body)
}
