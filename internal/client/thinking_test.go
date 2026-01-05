package client

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewThinkingEventParser tests the parser constructor
func TestNewThinkingEventParser(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "creates parser with initialized state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewThinkingEventParser()

			require.NotNil(t, parser)
			require.NotNil(t, parser.state)
			assert.Equal(t, 0, parser.state.TotalEvents)
			assert.Equal(t, 0, parser.state.CurrentDepth)
			assert.False(t, parser.state.IsThinking)
			assert.NotNil(t, parser.state.Blocks)
			assert.Empty(t, parser.state.Blocks)
		})
	}
}

// TestParseEvent tests parsing thinking events
func TestParseEvent(t *testing.T) {
	tests := []struct {
		name        string
		event       *providers.Event
		expectBlock bool
		expectError bool
		errorMsg    string
	}{
		{
			name: "parses valid thinking event with basic data",
			event: &providers.Event{
				Type:      providers.EventThinking,
				Data:      "Analyzing the problem...",
				Timestamp: time.Now(),
			},
			expectBlock: true,
			expectError: false,
		},
		{
			name: "parses thinking event with metadata",
			event: &providers.Event{
				Type:      providers.EventThinking,
				Data:      "Deep thought process",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"depth": float64(2),
					"index": float64(5),
					"type":  "delta",
				},
			},
			expectBlock: true,
			expectError: false,
		},
		{
			name: "returns nil for non-thinking event",
			event: &providers.Event{
				Type:      providers.EventTextDelta,
				Data:      "Regular text",
				Timestamp: time.Now(),
			},
			expectBlock: false,
			expectError: false,
		},
		{
			name:        "returns error for nil event",
			event:       nil,
			expectBlock: false,
			expectError: true,
			errorMsg:    "event cannot be nil",
		},
		{
			name: "parses thinking start event",
			event: &providers.Event{
				Type:      providers.EventThinking,
				Data:      "Starting to think...",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"type": "start",
				},
			},
			expectBlock: true,
			expectError: false,
		},
		{
			name: "parses thinking end event",
			event: &providers.Event{
				Type:      providers.EventThinking,
				Data:      "Finished thinking",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"type": "end",
				},
			},
			expectBlock: true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewThinkingEventParser()

			block, err := parser.ParseEvent(tt.event)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)

			if tt.expectBlock {
				require.NotNil(t, block)
				assert.Equal(t, tt.event.Data, block.Content)
				assert.Equal(t, tt.event.Timestamp, block.Timestamp)

				// Verify metadata parsing
				if tt.event.Metadata != nil {
					if depth, ok := tt.event.Metadata["depth"].(float64); ok {
						assert.Equal(t, int(depth), block.Depth)
					}
					if index, ok := tt.event.Metadata["index"].(float64); ok {
						assert.Equal(t, int(index), block.Index)
					}
					if blockType, ok := tt.event.Metadata["type"].(string); ok {
						assert.Equal(t, blockType, block.Type)
					}
				}
			} else {
				assert.Nil(t, block)
			}
		})
	}
}

// TestParseEventStateUpdates tests that parsing updates state correctly
func TestParseEventStateUpdates(t *testing.T) {
	parser := NewThinkingEventParser()

	// Start event
	startEvent := &providers.Event{
		Type:      providers.EventThinking,
		Data:      "Start",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"type":  "start",
			"depth": float64(0),
		},
	}

	block, err := parser.ParseEvent(startEvent)
	require.NoError(t, err)
	require.NotNil(t, block)

	state := parser.GetState()
	assert.True(t, state.IsThinking)
	assert.Equal(t, 1, state.TotalEvents)
	assert.Equal(t, 1, len(state.Blocks))

	// Delta event
	deltaEvent := &providers.Event{
		Type:      providers.EventThinking,
		Data:      "Thinking...",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"type":  "delta",
			"depth": float64(1),
		},
	}

	block, err = parser.ParseEvent(deltaEvent)
	require.NoError(t, err)
	require.NotNil(t, block)

	state = parser.GetState()
	assert.True(t, state.IsThinking)
	assert.Equal(t, 2, state.TotalEvents)
	assert.Equal(t, 1, state.CurrentDepth)
	assert.Equal(t, 2, len(state.Blocks))

	// End event
	endEvent := &providers.Event{
		Type:      providers.EventThinking,
		Data:      "End",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"type": "end",
		},
	}

	block, err = parser.ParseEvent(endEvent)
	require.NoError(t, err)
	require.NotNil(t, block)

	state = parser.GetState()
	assert.False(t, state.IsThinking)
	assert.Equal(t, 3, state.TotalEvents)
	assert.Equal(t, 3, len(state.Blocks))
}

// TestParseThinkingData tests the internal data parsing function
func TestParseThinkingData(t *testing.T) {
	tests := []struct {
		name          string
		event         *providers.Event
		expectedDepth int
		expectedIndex int
		expectedType  string
	}{
		{
			name: "parses all metadata fields",
			event: &providers.Event{
				Type:      providers.EventThinking,
				Data:      "Test",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"depth": float64(3),
					"index": float64(10),
					"type":  "start",
					"extra": "metadata",
				},
			},
			expectedDepth: 3,
			expectedIndex: 10,
			expectedType:  "start",
		},
		{
			name: "uses default type when not provided",
			event: &providers.Event{
				Type:      providers.EventThinking,
				Data:      "Test",
				Timestamp: time.Now(),
			},
			expectedDepth: 0,
			expectedIndex: 0,
			expectedType:  "delta",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewThinkingEventParser()

			block, err := parser.parseThinkingData(tt.event)
			require.NoError(t, err)
			require.NotNil(t, block)

			assert.Equal(t, tt.expectedDepth, block.Depth)
			assert.Equal(t, tt.expectedIndex, block.Index)
			assert.Equal(t, tt.expectedType, block.Type)
			assert.Equal(t, tt.event.Data, block.Content)
			assert.Equal(t, tt.event.Timestamp, block.Timestamp)
		})
	}
}

// TestUpdateState tests state update logic
func TestUpdateState(t *testing.T) {
	tests := []struct {
		name           string
		block          *ThinkingBlock
		expectedEvents int
		expectedThinking bool
		expectedDepth  int
	}{
		{
			name: "updates state with start block",
			block: &ThinkingBlock{
				Type:      "start",
				Depth:     0,
				Timestamp: time.Now(),
			},
			expectedEvents:   1,
			expectedThinking: true,
			expectedDepth:    0,
		},
		{
			name: "updates state with delta block",
			block: &ThinkingBlock{
				Type:      "delta",
				Depth:     2,
				Timestamp: time.Now(),
			},
			expectedEvents:   1,
			expectedThinking: false,
			expectedDepth:    2,
		},
		{
			name: "updates state with end block",
			block: &ThinkingBlock{
				Type:      "end",
				Depth:     0,
				Timestamp: time.Now(),
			},
			expectedEvents:   1,
			expectedThinking: false,
			expectedDepth:    0,
		},
		{
			name:             "handles nil block",
			block:            nil,
			expectedEvents:   0,
			expectedThinking: false,
			expectedDepth:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewThinkingEventParser()

			parser.updateState(tt.block)

			state := parser.GetState()
			assert.Equal(t, tt.expectedEvents, state.TotalEvents)
			assert.Equal(t, tt.expectedThinking, state.IsThinking)
			assert.Equal(t, tt.expectedDepth, state.CurrentDepth)

			if tt.block != nil {
				assert.Equal(t, 1, len(state.Blocks))
			} else {
				assert.Equal(t, 0, len(state.Blocks))
			}
		})
	}
}

// TestGetState tests state retrieval
func TestGetState(t *testing.T) {
	parser := NewThinkingEventParser()

	state1 := parser.GetState()
	require.NotNil(t, state1)

	// Add a block
	parser.updateState(&ThinkingBlock{
		Type:      "delta",
		Content:   "test",
		Timestamp: time.Now(),
	})

	state2 := parser.GetState()
	require.NotNil(t, state2)

	// Should be the same state object
	assert.Equal(t, state1, state2)
	assert.Equal(t, 1, state2.TotalEvents)
}

// TestReset tests parser reset functionality
func TestReset(t *testing.T) {
	parser := NewThinkingEventParser()

	// Add some events
	event := &providers.Event{
		Type:      providers.EventThinking,
		Data:      "Test",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"depth": float64(2),
		},
	}

	_, err := parser.ParseEvent(event)
	require.NoError(t, err)

	state := parser.GetState()
	assert.Equal(t, 1, state.TotalEvents)
	assert.Equal(t, 1, len(state.Blocks))

	// Reset
	parser.Reset()

	state = parser.GetState()
	assert.Equal(t, 0, state.TotalEvents)
	assert.Equal(t, 0, state.CurrentDepth)
	assert.False(t, state.IsThinking)
	assert.Empty(t, state.Blocks)
}

// TestFormatThinkingBlock tests block formatting
func TestFormatThinkingBlock(t *testing.T) {
	tests := []struct {
		name      string
		block     *ThinkingBlock
		showDepth bool
		expected  string
	}{
		{
			name: "formats delta block without depth",
			block: &ThinkingBlock{
				Type:    "delta",
				Content: "Thinking...",
				Depth:   0,
			},
			showDepth: false,
			expected:  "Thinking...",
		},
		{
			name: "formats delta block with depth",
			block: &ThinkingBlock{
				Type:    "delta",
				Content: "Deep thought",
				Depth:   2,
			},
			showDepth: true,
			expected:  "    Deep thought",
		},
		{
			name: "formats start block",
			block: &ThinkingBlock{
				Type:    "start",
				Content: "Beginning",
				Depth:   0,
			},
			showDepth: false,
			expected:  "[THINKING START] Beginning",
		},
		{
			name: "formats end block",
			block: &ThinkingBlock{
				Type:    "end",
				Content: "Done",
				Depth:   0,
			},
			showDepth: false,
			expected:  "[THINKING END] Done",
		},
		{
			name:      "handles nil block",
			block:     nil,
			showDepth: false,
			expected:  "",
		},
		{
			name: "formats start block with depth",
			block: &ThinkingBlock{
				Type:    "start",
				Content: "Nested start",
				Depth:   1,
			},
			showDepth: true,
			expected:  "  [THINKING START] Nested start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatThinkingBlock(tt.block, tt.showDepth)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMergeThinkingBlocks tests merging multiple blocks
func TestMergeThinkingBlocks(t *testing.T) {
	tests := []struct {
		name     string
		blocks   []ThinkingBlock
		expected string
	}{
		{
			name:     "returns empty for no blocks",
			blocks:   []ThinkingBlock{},
			expected: "",
		},
		{
			name: "merges single block",
			blocks: []ThinkingBlock{
				{Type: "delta", Content: "First", Depth: 0},
			},
			expected: "First",
		},
		{
			name: "merges multiple blocks with depth",
			blocks: []ThinkingBlock{
				{Type: "start", Content: "Begin", Depth: 0},
				{Type: "delta", Content: "Think", Depth: 1},
				{Type: "end", Content: "Done", Depth: 0},
			},
			expected: "[THINKING START] Begin\n  Think\n[THINKING END] Done",
		},
		{
			name: "preserves depth indentation",
			blocks: []ThinkingBlock{
				{Type: "delta", Content: "Level 0", Depth: 0},
				{Type: "delta", Content: "Level 1", Depth: 1},
				{Type: "delta", Content: "Level 2", Depth: 2},
			},
			expected: "Level 0\n  Level 1\n    Level 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeThinkingBlocks(tt.blocks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestThinkingEventFromJSON tests JSON parsing
func TestThinkingEventFromJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectError bool
		errorMsg    string
	}{
		{
			name: "parses valid thinking event JSON",
			jsonData: `{
				"type": "thinking",
				"data": "Test thinking",
				"timestamp": "2024-01-01T00:00:00Z"
			}`,
			expectError: false,
		},
		{
			name: "parses thinking event with metadata",
			jsonData: `{
				"type": "thinking",
				"data": "Test",
				"timestamp": "2024-01-01T00:00:00Z",
				"metadata": {
					"depth": 1,
					"index": 5
				}
			}`,
			expectError: false,
		},
		{
			name:        "returns error for invalid JSON",
			jsonData:    `{"invalid json`,
			expectError: true,
			errorMsg:    "failed to unmarshal",
		},
		{
			name: "returns error for non-thinking event type",
			jsonData: `{
				"type": "text_delta",
				"data": "Not thinking",
				"timestamp": "2024-01-01T00:00:00Z"
			}`,
			expectError: true,
			errorMsg:    "invalid event type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ThinkingEventFromJSON(tt.jsonData)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, event)
			assert.Equal(t, providers.EventThinking, event.Type)
		})
	}
}

// TestValidateThinkingEvent tests event validation
func TestValidateThinkingEvent(t *testing.T) {
	tests := []struct {
		name        string
		event       *providers.Event
		expectError bool
		errorMsg    string
	}{
		{
			name: "validates correct thinking event",
			event: &providers.Event{
				Type:      providers.EventThinking,
				Data:      "Valid data",
				Timestamp: time.Now(),
			},
			expectError: false,
		},
		{
			name:        "returns error for nil event",
			event:       nil,
			expectError: true,
			errorMsg:    "event cannot be nil",
		},
		{
			name: "returns error for wrong event type",
			event: &providers.Event{
				Type:      providers.EventTextDelta,
				Data:      "Wrong type",
				Timestamp: time.Now(),
			},
			expectError: true,
			errorMsg:    "invalid event type",
		},
		{
			name: "returns error for empty data",
			event: &providers.Event{
				Type:      providers.EventThinking,
				Data:      "",
				Timestamp: time.Now(),
			},
			expectError: true,
			errorMsg:    "must have non-empty data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateThinkingEvent(tt.event)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestGetThinkingDuration tests duration calculation
func TestGetThinkingDuration(t *testing.T) {
	tests := []struct {
		name     string
		state    *ThinkingState
		expected time.Duration
	}{
		{
			name:     "returns zero for nil state",
			state:    nil,
			expected: 0,
		},
		{
			name: "returns zero for zero start time",
			state: &ThinkingState{
				StartTime:     time.Time{},
				LastEventTime: time.Now(),
			},
			expected: 0,
		},
		{
			name: "calculates duration correctly",
			state: &ThinkingState{
				StartTime:     time.Now().Add(-5 * time.Second),
				LastEventTime: time.Now(),
			},
			expected: 5 * time.Second,
		},
		{
			name: "handles missing end time",
			state: &ThinkingState{
				StartTime:     time.Now().Add(-3 * time.Second),
				LastEventTime: time.Time{},
			},
			expected: 3 * time.Second, // Approximate, uses current time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := GetThinkingDuration(tt.state)

			if tt.name == "handles missing end time" {
				// For this case, duration should be approximately the expected value
				assert.GreaterOrEqual(t, duration, tt.expected-100*time.Millisecond)
			} else {
				// Allow 100ms tolerance for timing
				assert.InDelta(t, tt.expected.Seconds(), duration.Seconds(), 0.1)
			}
		})
	}
}

// TestFilterThinkingBlocks tests block filtering
func TestFilterThinkingBlocks(t *testing.T) {
	blocks := []ThinkingBlock{
		{Type: "start", Content: "Start 1"},
		{Type: "delta", Content: "Delta 1"},
		{Type: "delta", Content: "Delta 2"},
		{Type: "end", Content: "End 1"},
		{Type: "start", Content: "Start 2"},
		{Type: "delta", Content: "Delta 3"},
	}

	tests := []struct {
		name          string
		blockType     string
		expectedCount int
	}{
		{
			name:          "filters start blocks",
			blockType:     "start",
			expectedCount: 2,
		},
		{
			name:          "filters delta blocks",
			blockType:     "delta",
			expectedCount: 3,
		},
		{
			name:          "filters end blocks",
			blockType:     "end",
			expectedCount: 1,
		},
		{
			name:          "returns empty for non-existent type",
			blockType:     "invalid",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterThinkingBlocks(blocks, tt.blockType)
			assert.Equal(t, tt.expectedCount, len(filtered))

			// Verify all filtered blocks have correct type
			for _, block := range filtered {
				assert.Equal(t, tt.blockType, block.Type)
			}
		})
	}
}

// TestGetMaxDepth tests maximum depth calculation
func TestGetMaxDepth(t *testing.T) {
	tests := []struct {
		name     string
		blocks   []ThinkingBlock
		expected int
	}{
		{
			name:     "returns zero for empty blocks",
			blocks:   []ThinkingBlock{},
			expected: 0,
		},
		{
			name: "returns max depth from single block",
			blocks: []ThinkingBlock{
				{Depth: 3},
			},
			expected: 3,
		},
		{
			name: "returns max depth from multiple blocks",
			blocks: []ThinkingBlock{
				{Depth: 1},
				{Depth: 5},
				{Depth: 3},
				{Depth: 2},
			},
			expected: 5,
		},
		{
			name: "handles all zero depth",
			blocks: []ThinkingBlock{
				{Depth: 0},
				{Depth: 0},
				{Depth: 0},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxDepth := GetMaxDepth(tt.blocks)
			assert.Equal(t, tt.expected, maxDepth)
		})
	}
}

// TestMultipleEventsSequence tests processing a sequence of events
func TestMultipleEventsSequence(t *testing.T) {
	parser := NewThinkingEventParser()

	events := []*providers.Event{
		{
			Type:      providers.EventThinking,
			Data:      "Starting analysis",
			Timestamp: time.Now(),
			Metadata:  map[string]interface{}{"type": "start"},
		},
		{
			Type:      providers.EventThinking,
			Data:      "Analyzing step 1",
			Timestamp: time.Now(),
			Metadata:  map[string]interface{}{"type": "delta", "depth": float64(1)},
		},
		{
			Type:      providers.EventThinking,
			Data:      "Deep dive",
			Timestamp: time.Now(),
			Metadata:  map[string]interface{}{"type": "delta", "depth": float64(2)},
		},
		{
			Type:      providers.EventThinking,
			Data:      "Analyzing step 2",
			Timestamp: time.Now(),
			Metadata:  map[string]interface{}{"type": "delta", "depth": float64(1)},
		},
		{
			Type:      providers.EventThinking,
			Data:      "Complete",
			Timestamp: time.Now(),
			Metadata:  map[string]interface{}{"type": "end"},
		},
	}

	// Process all events
	for _, event := range events {
		_, err := parser.ParseEvent(event)
		require.NoError(t, err)
	}

	state := parser.GetState()

	// Verify final state
	assert.Equal(t, 5, state.TotalEvents)
	assert.Equal(t, 5, len(state.Blocks))
	assert.False(t, state.IsThinking)
	assert.Equal(t, 2, state.CurrentDepth) // Max depth reached

	// Verify block sequence
	assert.Equal(t, "start", state.Blocks[0].Type)
	assert.Equal(t, "delta", state.Blocks[1].Type)
	assert.Equal(t, "delta", state.Blocks[2].Type)
	assert.Equal(t, "delta", state.Blocks[3].Type)
	assert.Equal(t, "end", state.Blocks[4].Type)
}

// TestThinkingBlockJSONMarshaling tests JSON marshaling/unmarshaling
func TestThinkingBlockJSONMarshaling(t *testing.T) {
	original := ThinkingBlock{
		Content:   "Test content",
		Depth:     2,
		Timestamp: time.Now().Round(time.Second),
		Index:     5,
		Type:      "delta",
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	// Marshal
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal
	var unmarshaled ThinkingBlock
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original.Content, unmarshaled.Content)
	assert.Equal(t, original.Depth, unmarshaled.Depth)
	assert.Equal(t, original.Index, unmarshaled.Index)
	assert.Equal(t, original.Type, unmarshaled.Type)
	assert.True(t, original.Timestamp.Equal(unmarshaled.Timestamp))
}

// TestConcurrentParsing tests parser with concurrent access (edge case)
func TestConcurrentParsing(t *testing.T) {
	parser := NewThinkingEventParser()

	event := &providers.Event{
		Type:      providers.EventThinking,
		Data:      "Concurrent test",
		Timestamp: time.Now(),
	}

	// Note: This test documents that the parser is NOT thread-safe
	// In production, callers should synchronize access
	_, err := parser.ParseEvent(event)
	require.NoError(t, err)

	state := parser.GetState()
	assert.Equal(t, 1, state.TotalEvents)
}

// Benchmark tests

func BenchmarkParseEvent(b *testing.B) {
	parser := NewThinkingEventParser()
	event := &providers.Event{
		Type:      providers.EventThinking,
		Data:      "Benchmark thinking",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"depth": float64(2),
			"index": float64(10),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseEvent(event)
	}
}

func BenchmarkFormatThinkingBlock(b *testing.B) {
	block := &ThinkingBlock{
		Type:    "delta",
		Content: "Benchmark content",
		Depth:   3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatThinkingBlock(block, true)
	}
}

func BenchmarkMergeThinkingBlocks(b *testing.B) {
	blocks := make([]ThinkingBlock, 100)
	for i := 0; i < 100; i++ {
		blocks[i] = ThinkingBlock{
			Type:    "delta",
			Content: "Block content",
			Depth:   i % 5,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MergeThinkingBlocks(blocks)
	}
}
