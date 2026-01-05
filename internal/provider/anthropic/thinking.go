package anthropic

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/AINative-studio/ainative-code/internal/provider"
)

// thinkingBlockStart represents the start of a thinking block in the stream
type thinkingBlockStart struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// thinkingBlockDelta represents incremental thinking content
type thinkingBlockDelta struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta struct {
		Type     string `json:"type"`
		Thinking string `json:"thinking"`
	} `json:"delta"`
}

// thinkingBlockStop represents the end of a thinking block
type thinkingBlockStop struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// parseThinkingBlockStart parses a thinking_block_start event
func parseThinkingBlockStart(data string) (*provider.ThinkingBlock, error) {
	var start thinkingBlockStart
	if err := json.Unmarshal([]byte(data), &start); err != nil {
		return nil, fmt.Errorf("failed to parse thinking_block_start: %w", err)
	}

	return &provider.ThinkingBlock{
		Content:   "",
		Index:     start.Index,
		Timestamp: time.Now().Unix(),
		Type:      "thinking",
	}, nil
}

// parseThinkingBlockDelta parses a thinking_block_delta event
func parseThinkingBlockDelta(data string) (*provider.ThinkingBlock, error) {
	var delta thinkingBlockDelta
	if err := json.Unmarshal([]byte(data), &delta); err != nil {
		return nil, fmt.Errorf("failed to parse thinking_block_delta: %w", err)
	}

	// Only process thinking_delta type
	if delta.Delta.Type != "thinking_delta" {
		return nil, nil
	}

	return &provider.ThinkingBlock{
		Content:   delta.Delta.Thinking,
		Index:     delta.Index,
		Timestamp: time.Now().Unix(),
		Type:      "thinking",
	}, nil
}

// parseThinkingBlockStop parses a thinking_block_stop event
func parseThinkingBlockStop(data string) (*provider.ThinkingBlock, error) {
	var stop thinkingBlockStop
	if err := json.Unmarshal([]byte(data), &stop); err != nil {
		return nil, fmt.Errorf("failed to parse thinking_block_stop: %w", err)
	}

	return &provider.ThinkingBlock{
		Content:   "", // No content on stop event
		Index:     stop.Index,
		Timestamp: time.Now().Unix(),
		Type:      "thinking_stop",
	}, nil
}

// isThinkingEvent checks if an SSE event type is a thinking-related event
func isThinkingEvent(eventType string) bool {
	switch eventType {
	case "thinking_block_start", "thinking_block_delta", "thinking_block_stop":
		return true
	default:
		return false
	}
}
