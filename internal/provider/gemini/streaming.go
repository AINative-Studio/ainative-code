package gemini

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/provider"
)

// streamResponse handles streaming responses from the Gemini API
// Gemini uses Server-Sent Events (SSE) for streaming
func (g *GeminiProvider) streamResponse(ctx context.Context, body io.ReadCloser, eventChan chan<- provider.Event, model string) {
	defer close(eventChan)
	defer body.Close()

	reader := bufio.NewReader(body)
	var currentText string

	// Send start event
	eventChan <- provider.Event{
		Type: provider.EventTypeContentStart,
	}

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

		// Read line
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				eventChan <- provider.Event{
					Type:  provider.EventTypeError,
					Error: provider.NewProviderError("gemini", model, fmt.Errorf("stream read error: %w", err)),
				}
			} else {
				// EOF reached, send completion event
				eventChan <- provider.Event{
					Type:    provider.EventTypeContentEnd,
					Content: currentText,
					Done:    true,
				}
			}
			return
		}

		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// SSE format: "data: {...}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Extract JSON data
		data := strings.TrimPrefix(line, "data: ")

		// Parse the chunk
		var chunk streamResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			// Skip unparseable chunks
			continue
		}

		// Check for prompt feedback (safety blocks)
		if chunk.PromptFeedback != nil && chunk.PromptFeedback.BlockReason != "" {
			eventChan <- provider.Event{
				Type:  provider.EventTypeError,
				Error: provider.NewProviderError("gemini", model,
					fmt.Errorf("prompt blocked: %s", chunk.PromptFeedback.BlockReason)),
			}
			return
		}

		// Process candidates
		if len(chunk.Candidates) > 0 {
			candidate := chunk.Candidates[0]

			// Extract text delta first (before checking finish reason)
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					currentText += part.Text
					eventChan <- provider.Event{
						Type:    provider.EventTypeContentDelta,
						Content: part.Text,
					}
				}
			}

			// Then check for finish reason
			if candidate.FinishReason != "" {
				// Check if blocked by safety
				if candidate.FinishReason == "SAFETY" {
					eventChan <- provider.Event{
						Type:  provider.EventTypeError,
						Error: provider.NewProviderError("gemini", model,
							fmt.Errorf("response blocked due to safety settings")),
					}
					return
				}

				// Normal completion
				eventChan <- provider.Event{
					Type:    provider.EventTypeContentEnd,
					Content: currentText,
					Done:    true,
				}
				return
			}
		}
	}
}
