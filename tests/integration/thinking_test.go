package integration

import (
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/config"
	"github.com/AINative-studio/ainative-code/internal/providers"
	"github.com/AINative-studio/ainative-code/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestThinkingEndToEndFlow tests the complete flow from API event to UI rendering
func TestThinkingEndToEndFlow(t *testing.T) {
	t.Run("parses and renders single thinking event", func(t *testing.T) {
		// Create parser
		parser := client.NewThinkingEventParser()

		// Create mock thinking event
		event := &providers.Event{
			Type:      providers.EventThinking,
			Data:      "Analyzing the problem step by step",
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"type":  "delta",
				"depth": float64(1),
			},
		}

		// Parse event
		block, err := parser.ParseEvent(event)
		require.NoError(t, err)
		require.NotNil(t, block)

		// Create TUI state
		tuiState := tui.NewThinkingState()
		tuiBlock := tuiState.AddThinkingBlock(block.Content, block.Depth)

		// Verify state
		assert.Equal(t, 1, len(tuiState.Blocks))
		assert.Equal(t, block.Content, tuiBlock.Content)
		assert.Equal(t, block.Depth, tuiBlock.Depth)

		// Render
		config := tui.DefaultThinkingConfig()
		rendered := tui.RenderThinkingBlock(tuiBlock, config)

		assert.NotEmpty(t, rendered)
		assert.Contains(t, rendered, "Thinking")
	})

	t.Run("processes sequence of thinking events", func(t *testing.T) {
		parser := client.NewThinkingEventParser()
		tuiState := tui.NewThinkingState()

		// Create sequence of events
		events := []*providers.Event{
			{
				Type:      providers.EventThinking,
				Data:      "Starting analysis",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"type":  "start",
					"depth": float64(0),
				},
			},
			{
				Type:      providers.EventThinking,
				Data:      "First consideration",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"type":  "delta",
					"depth": float64(1),
				},
			},
			{
				Type:      providers.EventThinking,
				Data:      "Deeper analysis",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"type":  "delta",
					"depth": float64(2),
				},
			},
			{
				Type:      providers.EventThinking,
				Data:      "Conclusion reached",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"type": "end",
				},
			},
		}

		// Process each event
		for _, evt := range events {
			block, err := parser.ParseEvent(evt)
			require.NoError(t, err)
			require.NotNil(t, block)

			tuiState.AddThinkingBlock(block.Content, block.Depth)
		}

		// Verify all events processed
		assert.Equal(t, 4, len(tuiState.Blocks))

		// Verify state
		state := parser.GetState()
		assert.False(t, state.IsThinking) // Should be ended
		assert.Equal(t, 4, state.TotalEvents)
		assert.Equal(t, 2, state.CurrentDepth) // Max depth reached

		// Render all
		config := tui.DefaultThinkingConfig()
		rendered := tui.RenderAllThinkingBlocks(tuiState, config)
		assert.NotEmpty(t, rendered)
	})
}

// TestThinkingWithConfiguration tests thinking with different configurations
func TestThinkingWithConfiguration(t *testing.T) {
	t.Run("respects max depth configuration", func(t *testing.T) {
		cfg := &config.Config{
			LLM: config.LLMConfig{
				DefaultProvider: "anthropic",
				Anthropic: &config.AnthropicConfig{
					APIKey: "test-key",
					Model:  "claude-3-5-sonnet-20241022",
					ExtendedThinking: &config.ExtendedThinkingConfig{
						Enabled:    true,
						AutoExpand: false,
						MaxDepth:   5,
					},
				},
			},
		}

		maxDepth := config.GetMaxThinkingDepth(cfg)
		assert.Equal(t, 5, maxDepth)

		// Create events up to max depth
		parser := client.NewThinkingEventParser()
		tuiState := tui.NewThinkingState()

		for i := 0; i <= maxDepth; i++ {
			event := &providers.Event{
				Type:      providers.EventThinking,
				Data:      "Deep thought",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"depth": float64(i),
				},
			}

			block, err := parser.ParseEvent(event)
			require.NoError(t, err)

			tuiState.AddThinkingBlock(block.Content, block.Depth)
		}

		assert.Equal(t, maxDepth+1, len(tuiState.Blocks))
	})

	t.Run("respects auto expand configuration", func(t *testing.T) {
		cfg := &config.Config{
			LLM: config.LLMConfig{
				Anthropic: &config.AnthropicConfig{
					ExtendedThinking: &config.ExtendedThinkingConfig{
						Enabled:    true,
						AutoExpand: true,
						MaxDepth:   10,
					},
				},
			},
		}

		autoExpand := config.ShouldAutoExpandThinking(cfg)
		assert.True(t, autoExpand)

		tuiState := tui.NewThinkingState()
		tuiState.AddThinkingBlock("Test", 0)

		// With auto expand, blocks should not be collapsed
		for _, block := range tuiState.Blocks {
			assert.False(t, block.Collapsed)
		}
	})

	t.Run("handles disabled thinking", func(t *testing.T) {
		cfg := &config.Config{
			LLM: config.LLMConfig{
				Anthropic: &config.AnthropicConfig{
					ExtendedThinking: &config.ExtendedThinkingConfig{
						Enabled:    false,
						AutoExpand: false,
						MaxDepth:   10,
					},
				},
			},
		}

		enabled := config.IsExtendedThinkingEnabled(cfg)
		assert.False(t, enabled)

		// Even with events, shouldn't show if disabled
		tuiState := tui.NewThinkingState()
		tuiState.ShowThinking = false
		tuiState.AddThinkingBlock("Test", 0)

		tuiConfig := tui.DefaultThinkingConfig()
		rendered := tui.RenderAllThinkingBlocks(tuiState, tuiConfig)
		assert.Empty(t, rendered)
	})
}

// TestThinkingInteractivity tests interactive features
func TestThinkingInteractivity(t *testing.T) {
	t.Run("toggles collapse state", func(t *testing.T) {
		tuiState := tui.NewThinkingState()
		block := tuiState.AddThinkingBlock("Content", 0)

		// Initially expanded
		assert.False(t, block.Collapsed)

		// Toggle to collapsed
		tuiState.ToggleBlock(block.ID)
		assert.True(t, block.Collapsed)

		// Toggle back
		tuiState.ToggleBlock(block.ID)
		assert.False(t, block.Collapsed)
	})

	t.Run("toggles display visibility", func(t *testing.T) {
		tuiState := tui.NewThinkingState()
		tuiState.AddThinkingBlock("Test", 0)

		// Initially visible
		assert.True(t, tuiState.ShowThinking)

		// Toggle to hidden
		tuiState.ToggleDisplay()
		assert.False(t, tuiState.ShowThinking)

		// Toggle back
		tuiState.ToggleDisplay()
		assert.True(t, tuiState.ShowThinking)
	})

	t.Run("collapses all blocks", func(t *testing.T) {
		tuiState := tui.NewThinkingState()
		tuiState.AddThinkingBlock("Block 1", 0)
		tuiState.AddThinkingBlock("Block 2", 1)
		tuiState.AddThinkingBlock("Block 3", 0)

		tuiState.CollapseAll()

		for _, block := range tuiState.Blocks {
			assert.True(t, block.Collapsed)
		}
	})

	t.Run("expands all blocks", func(t *testing.T) {
		tuiState := tui.NewThinkingState()
		block1 := tuiState.AddThinkingBlock("Block 1", 0)
		block2 := tuiState.AddThinkingBlock("Block 2", 1)

		// Manually collapse
		block1.Collapsed = true
		block2.Collapsed = true

		tuiState.ExpandAll()

		for _, block := range tuiState.Blocks {
			assert.False(t, block.Collapsed)
		}
	})
}

// TestThinkingPerformance tests performance with large datasets
func TestThinkingPerformance(t *testing.T) {
	t.Run("handles many thinking events", func(t *testing.T) {
		parser := client.NewThinkingEventParser()
		tuiState := tui.NewThinkingState()

		// Create 100 events
		eventCount := 100
		for i := 0; i < eventCount; i++ {
			event := &providers.Event{
				Type:      providers.EventThinking,
				Data:      "Thinking step",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"depth": float64(i % 10),
				},
			}

			block, err := parser.ParseEvent(event)
			require.NoError(t, err)
			tuiState.AddThinkingBlock(block.Content, block.Depth)
		}

		assert.Equal(t, eventCount, len(tuiState.Blocks))

		// Should still be able to render
		config := tui.DefaultThinkingConfig()
		rendered := tui.RenderAllThinkingBlocks(tuiState, config)
		assert.NotEmpty(t, rendered)
	})

	t.Run("handles large thinking content", func(t *testing.T) {
		// Create large content block
		largeContent := ""
		for i := 0; i < 1000; i++ {
			largeContent += "This is a line of thinking content. "
		}

		parser := client.NewThinkingEventParser()
		event := &providers.Event{
			Type:      providers.EventThinking,
			Data:      largeContent,
			Timestamp: time.Now(),
		}

		block, err := parser.ParseEvent(event)
		require.NoError(t, err)
		assert.Equal(t, largeContent, block.Content)

		// Should render without error
		tuiState := tui.NewThinkingState()
		tuiBlock := tuiState.AddThinkingBlock(block.Content, 0)

		config := tui.DefaultThinkingConfig()
		rendered := tui.RenderThinkingBlock(tuiBlock, config)
		assert.NotEmpty(t, rendered)
	})
}

// TestThinkingErrorHandling tests error scenarios
func TestThinkingErrorHandling(t *testing.T) {
	t.Run("handles malformed events", func(t *testing.T) {
		parser := client.NewThinkingEventParser()

		// Nil event
		block, err := parser.ParseEvent(nil)
		assert.Error(t, err)
		assert.Nil(t, block)

		// Wrong event type
		wrongEvent := &providers.Event{
			Type:      providers.EventTextDelta,
			Data:      "Not thinking",
			Timestamp: time.Now(),
		}

		block, err = parser.ParseEvent(wrongEvent)
		assert.NoError(t, err) // Should not error, just return nil
		assert.Nil(t, block)
	})

	t.Run("handles missing metadata", func(t *testing.T) {
		parser := client.NewThinkingEventParser()

		event := &providers.Event{
			Type:      providers.EventThinking,
			Data:      "Content",
			Timestamp: time.Now(),
			Metadata:  nil, // No metadata
		}

		block, err := parser.ParseEvent(event)
		require.NoError(t, err)
		require.NotNil(t, block)

		// Should use defaults
		assert.Equal(t, "delta", block.Type)
		assert.Equal(t, 0, block.Depth)
	})

	t.Run("handles empty content", func(t *testing.T) {
		tuiState := tui.NewThinkingState()
		block := tuiState.AddThinkingBlock("", 0)

		assert.NotNil(t, block)
		assert.Empty(t, block.Content)

		// Should still render
		config := tui.DefaultThinkingConfig()
		rendered := tui.RenderThinkingBlock(block, config)
		assert.NotEmpty(t, rendered)
	})
}

// TestThinkingValidation tests validation of thinking events
func TestThinkingValidation(t *testing.T) {
	t.Run("validates correct thinking event", func(t *testing.T) {
		event := &providers.Event{
			Type:      providers.EventThinking,
			Data:      "Valid thinking",
			Timestamp: time.Now(),
		}

		err := client.ValidateThinkingEvent(event)
		assert.NoError(t, err)
	})

	t.Run("rejects invalid event type", func(t *testing.T) {
		event := &providers.Event{
			Type:      providers.EventTextDelta,
			Data:      "Wrong type",
			Timestamp: time.Now(),
		}

		err := client.ValidateThinkingEvent(event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid event type")
	})

	t.Run("rejects empty data", func(t *testing.T) {
		event := &providers.Event{
			Type:      providers.EventThinking,
			Data:      "",
			Timestamp: time.Now(),
		}

		err := client.ValidateThinkingEvent(event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non-empty data")
	})
}

// TestThinkingStateManagement tests state management across operations
func TestThinkingStateManagement(t *testing.T) {
	t.Run("maintains state across multiple operations", func(t *testing.T) {
		parser := client.NewThinkingEventParser()

		// Add start event
		startEvent := &providers.Event{
			Type:      providers.EventThinking,
			Data:      "Start",
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"type": "start",
			},
		}

		_, err := parser.ParseEvent(startEvent)
		require.NoError(t, err)

		state := parser.GetState()
		assert.True(t, state.IsThinking)
		assert.Equal(t, 1, state.TotalEvents)

		// Add delta events
		for i := 0; i < 5; i++ {
			deltaEvent := &providers.Event{
				Type:      providers.EventThinking,
				Data:      "Delta",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"type":  "delta",
					"depth": float64(i),
				},
			}

			_, err := parser.ParseEvent(deltaEvent)
			require.NoError(t, err)
		}

		state = parser.GetState()
		assert.True(t, state.IsThinking)
		assert.Equal(t, 6, state.TotalEvents)
		assert.Equal(t, 4, state.CurrentDepth)

		// Add end event
		endEvent := &providers.Event{
			Type:      providers.EventThinking,
			Data:      "End",
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"type": "end",
			},
		}

		_, err = parser.ParseEvent(endEvent)
		require.NoError(t, err)

		state = parser.GetState()
		assert.False(t, state.IsThinking)
		assert.Equal(t, 7, state.TotalEvents)

		// Reset and verify
		parser.Reset()
		state = parser.GetState()
		assert.Equal(t, 0, state.TotalEvents)
		assert.False(t, state.IsThinking)
		assert.Empty(t, state.Blocks)
	})
}

// TestThinkingDurationCalculation tests timing calculations
func TestThinkingDurationCalculation(t *testing.T) {
	t.Run("calculates thinking duration", func(t *testing.T) {
		startTime := time.Now().Add(-5 * time.Second)
		endTime := time.Now()

		state := &client.ThinkingState{
			StartTime:     startTime,
			LastEventTime: endTime,
		}

		duration := client.GetThinkingDuration(state)

		// Should be approximately 5 seconds (with some tolerance)
		assert.InDelta(t, 5.0, duration.Seconds(), 0.1)
	})
}

// TestThinkingBlockFiltering tests filtering operations
func TestThinkingBlockFiltering(t *testing.T) {
	t.Run("filters blocks by type", func(t *testing.T) {
		blocks := []client.ThinkingBlock{
			{Type: "start", Content: "Start 1"},
			{Type: "delta", Content: "Delta 1"},
			{Type: "delta", Content: "Delta 2"},
			{Type: "end", Content: "End 1"},
			{Type: "delta", Content: "Delta 3"},
		}

		startBlocks := client.FilterThinkingBlocks(blocks, "start")
		assert.Equal(t, 1, len(startBlocks))

		deltaBlocks := client.FilterThinkingBlocks(blocks, "delta")
		assert.Equal(t, 3, len(deltaBlocks))

		endBlocks := client.FilterThinkingBlocks(blocks, "end")
		assert.Equal(t, 1, len(endBlocks))
	})

	t.Run("gets maximum depth", func(t *testing.T) {
		blocks := []client.ThinkingBlock{
			{Depth: 1},
			{Depth: 5},
			{Depth: 3},
			{Depth: 2},
		}

		maxDepth := client.GetMaxDepth(blocks)
		assert.Equal(t, 5, maxDepth)
	})
}

// Benchmark tests

func BenchmarkThinkingEndToEnd(b *testing.B) {
	parser := client.NewThinkingEventParser()
	tuiState := tui.NewThinkingState()
	config := tui.DefaultThinkingConfig()

	event := &providers.Event{
		Type:      providers.EventThinking,
		Data:      "Benchmark thinking content",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"depth": float64(2),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block, _ := parser.ParseEvent(event)
		tuiBlock := tuiState.AddThinkingBlock(block.Content, block.Depth)
		_ = tui.RenderThinkingBlock(tuiBlock, config)
	}
}

func BenchmarkLargeThinkingSequence(b *testing.B) {
	parser := client.NewThinkingEventParser()

	events := make([]*providers.Event, 100)
	for i := 0; i < 100; i++ {
		events[i] = &providers.Event{
			Type:      providers.EventThinking,
			Data:      "Thinking step",
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"depth": float64(i % 10),
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Reset()
		for _, evt := range events {
			_, _ = parser.ParseEvent(evt)
		}
	}
}
