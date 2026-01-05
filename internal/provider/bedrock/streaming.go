package bedrock

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/AINative-studio/ainative-code/internal/provider"
)

// streamEvent represents a streaming event from Bedrock
type streamEvent struct {
	EventType    string
	Text         string
	ErrorMessage string
}

// streamReader reads streaming events from Bedrock
type streamReader struct {
	reader io.ReadCloser
	scanner *bufio.Scanner
}

// newStreamReader creates a new stream reader
func newStreamReader(reader io.ReadCloser) *streamReader {
	return &streamReader{
		reader:  reader,
		scanner: bufio.NewScanner(reader),
	}
}

// readEvent reads the next event from the stream
func (sr *streamReader) readEvent() ([]byte, error) {
	if sr.scanner.Scan() {
		return sr.scanner.Bytes(), nil
	}

	if err := sr.scanner.Err(); err != nil {
		return nil, err
	}

	return nil, io.EOF
}

// parseStreamingEvents parses streaming events from Bedrock and sends them to the event channel
func parseStreamingEvents(ctx context.Context, body io.ReadCloser, eventChan chan<- provider.Event, model string) {
	defer close(eventChan)
	defer body.Close()

	reader := newStreamReader(body)
	var currentText string

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			eventChan <- provider.Event{
				Type:  provider.EventTypeError,
				Error: ctx.Err(),
			}
			return
		default:
		}

		// Read next event
		eventData, err := reader.readEvent()
		if err != nil {
			if err != io.EOF {
				eventChan <- provider.Event{
					Type:  provider.EventTypeError,
					Error: provider.NewProviderError("bedrock", model, err),
				}
			}
			return
		}

		// Parse event
		event, err := parseStreamEvent(eventData)
		if err != nil {
			// Skip invalid events
			continue
		}

		if event == nil {
			// Skip unsupported events
			continue
		}

		// Handle event
		switch event.EventType {
		case "messageStart", "contentBlockStart":
			eventChan <- provider.Event{
				Type: provider.EventTypeContentStart,
			}

		case "contentBlockDelta":
			currentText += event.Text
			eventChan <- provider.Event{
				Type:    provider.EventTypeContentDelta,
				Content: event.Text,
			}

		case "messageStop", "contentBlockStop":
			if event.EventType == "messageStop" {
				eventChan <- provider.Event{
					Type:    provider.EventTypeContentEnd,
					Content: currentText,
					Done:    true,
				}
				return
			}

		case "error":
			eventChan <- provider.Event{
				Type:  provider.EventTypeError,
				Error: provider.NewProviderError("bedrock", model, fmt.Errorf("stream error: %s", event.ErrorMessage)),
			}
			return

		case "metadata":
			// Skip metadata events
			continue
		}
	}
}

// parseStreamEvent parses a single streaming event
func parseStreamEvent(data []byte) (*streamEvent, error) {
	// Bedrock sends JSON events, one per line
	var eventData map[string]interface{}
	if err := json.Unmarshal(data, &eventData); err != nil {
		return nil, err
	}

	// Determine event type based on keys
	event := &streamEvent{}

	if _, ok := eventData["messageStart"]; ok {
		event.EventType = "messageStart"
		return event, nil
	}

	if _, ok := eventData["contentBlockStart"]; ok {
		event.EventType = "contentBlockStart"
		return event, nil
	}

	if delta, ok := eventData["contentBlockDelta"]; ok {
		event.EventType = "contentBlockDelta"
		if deltaMap, ok := delta.(map[string]interface{}); ok {
			if deltaContent, ok := deltaMap["delta"].(map[string]interface{}); ok {
				if text, ok := deltaContent["text"].(string); ok {
					event.Text = text
				}
			}
		}
		return event, nil
	}

	if _, ok := eventData["contentBlockStop"]; ok {
		event.EventType = "contentBlockStop"
		return event, nil
	}

	if _, ok := eventData["messageStop"]; ok {
		event.EventType = "messageStop"
		return event, nil
	}

	if errData, ok := eventData["error"]; ok {
		event.EventType = "error"
		if errMap, ok := errData.(map[string]interface{}); ok {
			if msg, ok := errMap["message"].(string); ok {
				event.ErrorMessage = msg
			}
		}
		return event, nil
	}

	if _, ok := eventData["metadata"]; ok {
		event.EventType = "metadata"
		return event, nil
	}

	// Unknown event type
	return nil, nil
}

// handleStreamingChunk processes a streaming chunk and sends appropriate events
func handleStreamingChunk(event *streamEvent, eventChan chan<- provider.Event, model string) {
	switch event.EventType {
	case "messageStart", "contentBlockStart":
		eventChan <- provider.Event{
			Type: provider.EventTypeContentStart,
		}

	case "contentBlockDelta":
		eventChan <- provider.Event{
			Type:    provider.EventTypeContentDelta,
			Content: event.Text,
		}

	case "messageStop":
		eventChan <- provider.Event{
			Type: provider.EventTypeContentEnd,
			Done: true,
		}

	case "error":
		eventChan <- provider.Event{
			Type:  provider.EventTypeError,
			Error: provider.NewProviderError("bedrock", model, fmt.Errorf("%s", event.ErrorMessage)),
		}
	}
}
