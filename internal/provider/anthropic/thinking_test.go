package anthropic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseThinkingBlockStart(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		expectError bool
		expectIndex int
	}{
		{
			name:        "valid thinking block start",
			data:        `{"type":"thinking_block","index":0}`,
			expectError: false,
			expectIndex: 0,
		},
		{
			name:        "thinking block start with different index",
			data:        `{"type":"thinking_block","index":2}`,
			expectError: false,
			expectIndex: 2,
		},
		{
			name:        "invalid JSON",
			data:        `{invalid json}`,
			expectError: true,
		},
		{
			name:        "empty data",
			data:        ``,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := parseThinkingBlockStart(tt.data)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, block)
			} else {
				require.NoError(t, err)
				require.NotNil(t, block)
				assert.Equal(t, tt.expectIndex, block.Index)
				assert.Equal(t, "thinking", block.Type)
				assert.Empty(t, block.Content)
				assert.Greater(t, block.Timestamp, int64(0))
			}
		})
	}
}

func TestParseThinkingBlockDelta(t *testing.T) {
	tests := []struct {
		name            string
		data            string
		expectError     bool
		expectContent   string
		expectIndex     int
		expectNilResult bool
	}{
		{
			name:          "valid thinking delta",
			data:          `{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"This is my thought"}}`,
			expectError:   false,
			expectContent: "This is my thought",
			expectIndex:   0,
		},
		{
			name:            "non-thinking delta type",
			data:            `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Regular text"}}`,
			expectError:     false,
			expectNilResult: true,
		},
		{
			name:          "thinking delta with empty content",
			data:          `{"type":"content_block_delta","index":1,"delta":{"type":"thinking_delta","thinking":""}}`,
			expectError:   false,
			expectContent: "",
			expectIndex:   1,
		},
		{
			name:        "invalid JSON",
			data:        `{invalid}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := parseThinkingBlockDelta(tt.data)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				if tt.expectNilResult {
					assert.Nil(t, block)
				} else {
					require.NotNil(t, block)
					assert.Equal(t, tt.expectContent, block.Content)
					assert.Equal(t, tt.expectIndex, block.Index)
					assert.Equal(t, "thinking", block.Type)
					assert.Greater(t, block.Timestamp, int64(0))
				}
			}
		})
	}
}

func TestParseThinkingBlockStop(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		expectError bool
		expectIndex int
	}{
		{
			name:        "valid thinking block stop",
			data:        `{"type":"content_block_stop","index":0}`,
			expectError: false,
			expectIndex: 0,
		},
		{
			name:        "thinking block stop with different index",
			data:        `{"type":"content_block_stop","index":5}`,
			expectError: false,
			expectIndex: 5,
		},
		{
			name:        "invalid JSON",
			data:        `not json`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := parseThinkingBlockStop(tt.data)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, block)
			} else {
				require.NoError(t, err)
				require.NotNil(t, block)
				assert.Equal(t, tt.expectIndex, block.Index)
				assert.Equal(t, "thinking_stop", block.Type)
				assert.Empty(t, block.Content)
				assert.Greater(t, block.Timestamp, int64(0))
			}
		})
	}
}

func TestIsThinkingEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		expected  bool
	}{
		{
			name:      "thinking_block_start is thinking event",
			eventType: "thinking_block_start",
			expected:  true,
		},
		{
			name:      "thinking_block_delta is thinking event",
			eventType: "thinking_block_delta",
			expected:  true,
		},
		{
			name:      "thinking_block_stop is thinking event",
			eventType: "thinking_block_stop",
			expected:  true,
		},
		{
			name:      "message_start is not thinking event",
			eventType: "message_start",
			expected:  false,
		},
		{
			name:      "content_block_delta is not thinking event",
			eventType: "content_block_delta",
			expected:  false,
		},
		{
			name:      "message_stop is not thinking event",
			eventType: "message_stop",
			expected:  false,
		},
		{
			name:      "empty string is not thinking event",
			eventType: "",
			expected:  false,
		},
		{
			name:      "random string is not thinking event",
			eventType: "random_event",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isThinkingEvent(tt.eventType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestThinkingBlockTimestamps(t *testing.T) {
	// Test that timestamps are generated correctly
	data := `{"type":"thinking_block","index":0}`

	block1, err := parseThinkingBlockStart(data)
	require.NoError(t, err)
	require.NotNil(t, block1)

	// Parse again after a small delay
	block2, err := parseThinkingBlockStart(data)
	require.NoError(t, err)
	require.NotNil(t, block2)

	// Timestamps should be close but potentially different
	assert.Greater(t, block1.Timestamp, int64(0))
	assert.Greater(t, block2.Timestamp, int64(0))

	// Both should be recent (within last few seconds)
	// This is a simple sanity check
	assert.InDelta(t, block1.Timestamp, block2.Timestamp, 5)
}

func TestThinkingBlockMultipleDeltas(t *testing.T) {
	// Simulate multiple thinking deltas for the same block
	deltas := []string{
		`{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"First "}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"thought "}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"fragment"}}`,
	}

	var content string
	for _, delta := range deltas {
		block, err := parseThinkingBlockDelta(delta)
		require.NoError(t, err)
		require.NotNil(t, block)
		assert.Equal(t, 0, block.Index)
		content += block.Content
	}

	assert.Equal(t, "First thought fragment", content)
}

func TestThinkingBlockSpecialCharacters(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "newlines in thinking",
			content: "First line\nSecond line\nThird line",
		},
		{
			name:    "unicode characters",
			content: "思考内容 🤔 Думаю",
		},
		{
			name:    "quotes and escapes",
			content: `He said "hello" and she replied \"hi\"`,
		},
		{
			name:    "JSON-like content",
			content: `{"key": "value", "nested": {"data": 123}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create properly escaped JSON
			data := `{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"` +
				escapeJSONString(tt.content) + `"}}`

			block, err := parseThinkingBlockDelta(data)
			require.NoError(t, err)
			require.NotNil(t, block)
			assert.Equal(t, tt.content, block.Content)
		})
	}
}

// Helper function to escape strings for JSON
func escapeJSONString(s string) string {
	result := ""
	for _, c := range s {
		switch c {
		case '"':
			result += `\"`
		case '\\':
			result += `\\`
		case '\n':
			result += `\n`
		case '\r':
			result += `\r`
		case '\t':
			result += `\t`
		default:
			result += string(c)
		}
	}
	return result
}
