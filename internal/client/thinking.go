package client

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AINative-studio/ainative-code/internal/providers"
)

// ThinkingBlock represents a single thinking event block
type ThinkingBlock struct {
	Content   string        `json:"content"`
	Depth     int           `json:"depth"`
	Timestamp time.Time     `json:"timestamp"`
	Index     int           `json:"index"`
	Type      string        `json:"type"` // "start", "delta", "end"
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ThinkingState tracks the current state of thinking events
type ThinkingState struct {
	Blocks        []ThinkingBlock
	CurrentDepth  int
	TotalEvents   int
	IsThinking    bool
	StartTime     time.Time
	LastEventTime time.Time
}

// ThinkingEventParser handles parsing and processing of thinking events
type ThinkingEventParser struct {
	state *ThinkingState
}

// NewThinkingEventParser creates a new thinking event parser
func NewThinkingEventParser() *ThinkingEventParser {
	return &ThinkingEventParser{
		state: &ThinkingState{
			Blocks:       make([]ThinkingBlock, 0),
			CurrentDepth: 0,
			TotalEvents:  0,
			IsThinking:   false,
		},
	}
}

// ParseEvent parses a provider event and extracts thinking information
func (p *ThinkingEventParser) ParseEvent(event *providers.Event) (*ThinkingBlock, error) {
	if event == nil {
		return nil, fmt.Errorf("event cannot be nil")
	}

	// Only process thinking events
	if event.Type != providers.EventThinking {
		return nil, nil
	}

	// Parse the event data
	block, err := p.parseThinkingData(event)
	if err != nil {
		return nil, fmt.Errorf("failed to parse thinking event: %w", err)
	}

	// Update state
	p.updateState(block)

	return block, nil
}

// parseThinkingData extracts thinking block information from event data
func (p *ThinkingEventParser) parseThinkingData(event *providers.Event) (*ThinkingBlock, error) {
	block := &ThinkingBlock{
		Timestamp: event.Timestamp,
		Content:   event.Data,
		Metadata:  make(map[string]interface{}),
	}

	// Try to parse metadata if available
	if event.Metadata != nil {
		if depth, ok := event.Metadata["depth"].(float64); ok {
			block.Depth = int(depth)
		}
		if index, ok := event.Metadata["index"].(float64); ok {
			block.Index = int(index)
		}
		if blockType, ok := event.Metadata["type"].(string); ok {
			block.Type = blockType
		}

		// Copy additional metadata
		for k, v := range event.Metadata {
			if k != "depth" && k != "index" && k != "type" {
				block.Metadata[k] = v
			}
		}
	}

	// Set defaults if not provided
	if block.Type == "" {
		block.Type = "delta"
	}

	return block, nil
}

// updateState updates the parser's internal state based on the new block
func (p *ThinkingEventParser) updateState(block *ThinkingBlock) {
	if block == nil {
		return
	}

	p.state.TotalEvents++
	p.state.LastEventTime = block.Timestamp

	switch block.Type {
	case "start":
		p.state.IsThinking = true
		p.state.StartTime = block.Timestamp
		p.state.CurrentDepth = block.Depth
	case "end":
		p.state.IsThinking = false
	case "delta":
		if block.Depth > p.state.CurrentDepth {
			p.state.CurrentDepth = block.Depth
		}
	}

	p.state.Blocks = append(p.state.Blocks, *block)
}

// GetState returns the current thinking state
func (p *ThinkingEventParser) GetState() *ThinkingState {
	return p.state
}

// Reset resets the parser state
func (p *ThinkingEventParser) Reset() {
	p.state = &ThinkingState{
		Blocks:       make([]ThinkingBlock, 0),
		CurrentDepth: 0,
		TotalEvents:  0,
		IsThinking:   false,
	}
}

// FormatThinkingBlock formats a thinking block for display
func FormatThinkingBlock(block *ThinkingBlock, showDepth bool) string {
	if block == nil {
		return ""
	}

	var sb strings.Builder

	// Add depth indicator if requested
	if showDepth && block.Depth > 0 {
		indent := strings.Repeat("  ", block.Depth)
		sb.WriteString(indent)
	}

	// Add type prefix for non-delta events
	switch block.Type {
	case "start":
		sb.WriteString("[THINKING START] ")
	case "end":
		sb.WriteString("[THINKING END] ")
	}

	// Add content
	sb.WriteString(block.Content)

	return sb.String()
}

// MergeThinkingBlocks combines multiple thinking blocks into a single text
func MergeThinkingBlocks(blocks []ThinkingBlock) string {
	if len(blocks) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, block := range blocks {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(FormatThinkingBlock(&block, true))
	}

	return sb.String()
}

// ThinkingEventFromJSON parses a thinking event from JSON string
func ThinkingEventFromJSON(jsonData string) (*providers.Event, error) {
	var event providers.Event
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal thinking event: %w", err)
	}

	if event.Type != providers.EventThinking {
		return nil, fmt.Errorf("invalid event type: expected %s, got %s", providers.EventThinking, event.Type)
	}

	return &event, nil
}

// ValidateThinkingEvent validates that a thinking event has required fields
func ValidateThinkingEvent(event *providers.Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	if event.Type != providers.EventThinking {
		return fmt.Errorf("invalid event type: expected %s, got %s", providers.EventThinking, event.Type)
	}

	if event.Data == "" {
		return fmt.Errorf("thinking event must have non-empty data")
	}

	return nil
}

// GetThinkingDuration calculates the duration of a thinking session
func GetThinkingDuration(state *ThinkingState) time.Duration {
	if state == nil || state.StartTime.IsZero() {
		return 0
	}

	endTime := state.LastEventTime
	if endTime.IsZero() || endTime.Before(state.StartTime) {
		endTime = time.Now()
	}

	return endTime.Sub(state.StartTime)
}

// FilterThinkingBlocks filters thinking blocks by type
func FilterThinkingBlocks(blocks []ThinkingBlock, blockType string) []ThinkingBlock {
	filtered := make([]ThinkingBlock, 0)
	for _, block := range blocks {
		if block.Type == blockType {
			filtered = append(filtered, block)
		}
	}
	return filtered
}

// GetMaxDepth returns the maximum depth found in thinking blocks
func GetMaxDepth(blocks []ThinkingBlock) int {
	maxDepth := 0
	for _, block := range blocks {
		if block.Depth > maxDepth {
			maxDepth = block.Depth
		}
	}
	return maxDepth
}
