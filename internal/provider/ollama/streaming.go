package ollama

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/AINative-studio/ainative-code/internal/provider"
)

// streamReader wraps a response body for reading JSON-delimited streaming chunks
type streamReader struct {
	body    io.ReadCloser
	scanner *bufio.Scanner
}

// newStreamReader creates a new stream reader
func newStreamReader(body io.ReadCloser) *streamReader {
	scanner := bufio.NewScanner(body)
	// Set a large buffer size for potentially large chunks
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	return &streamReader{
		body:    body,
		scanner: scanner,
	}
}

// readChunk reads the next chunk from the stream
func (s *streamReader) readChunk() (*ollamaStreamResponse, error) {
	for s.scanner.Scan() {
		line := s.scanner.Bytes()
		if len(line) == 0 {
			continue // Skip empty lines
		}

		return parseStreamChunk(line)
	}

	if err := s.scanner.Err(); err != nil {
		return nil, err
	}

	return nil, io.EOF
}

// parseStreamChunk parses a single JSON chunk from the stream
func parseStreamChunk(data []byte) (*ollamaStreamResponse, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty chunk")
	}

	var resp ollamaStreamResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse chunk: %w", err)
	}

	return &resp, nil
}

// handleStreamResponse processes streaming responses from Ollama
func handleStreamResponse(ctx context.Context, body io.ReadCloser, eventChan chan<- provider.Event, model string) {
	defer close(eventChan)
	defer body.Close()

	reader := newStreamReader(body)
	var fullContent string

	// Send start event
	eventChan <- provider.Event{
		Type: provider.EventTypeContentStart,
	}

	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			eventChan <- provider.Event{
				Type:  provider.EventTypeError,
				Error: fmt.Errorf("stream cancelled: %w", ctx.Err()),
			}
			return
		default:
		}

		// Read next chunk
		chunk, err := reader.readChunk()
		if err == io.EOF {
			// Stream ended normally
			eventChan <- provider.Event{
				Type:    provider.EventTypeContentEnd,
				Content: fullContent,
				Done:    true,
			}
			return
		}
		if err != nil {
			eventChan <- provider.Event{
				Type:  provider.EventTypeError,
				Error: provider.NewProviderError("ollama", model, err),
			}
			return
		}

		// Check for error in response
		if chunk.Error != "" {
			eventChan <- provider.Event{
				Type:  provider.EventTypeError,
				Error: parseOllamaError(0, []byte(fmt.Sprintf(`{"error":"%s"}`, chunk.Error)), model),
			}
			return
		}

		// Handle content delta
		if chunk.Message.Content != "" {
			fullContent += chunk.Message.Content
			eventChan <- provider.Event{
				Type:    provider.EventTypeContentDelta,
				Content: chunk.Message.Content,
			}
		}

		// Check if done
		if chunk.Done {
			eventChan <- provider.Event{
				Type:    provider.EventTypeContentEnd,
				Content: fullContent,
				Done:    true,
			}
			return
		}
	}
}

// convertUsageStats converts Ollama stats to provider Usage
func convertUsageStats(resp *ollamaStreamResponse) provider.Usage {
	return provider.Usage{
		PromptTokens:     resp.PromptEvalCount,
		CompletionTokens: resp.EvalCount,
		TotalTokens:      resp.PromptEvalCount + resp.EvalCount,
	}
}

// aggregateStreamContent collects all content from a stream
// This is a helper for non-streaming use cases
func aggregateStreamContent(eventChan <-chan provider.Event) (string, provider.Usage, error) {
	var content string
	var usage provider.Usage

	for event := range eventChan {
		switch event.Type {
		case provider.EventTypeContentDelta:
			content += event.Content
		case provider.EventTypeContentEnd:
			if event.Content != "" {
				content = event.Content
			}
			return content, usage, nil
		case provider.EventTypeError:
			return content, usage, event.Error
		}
	}

	return content, usage, nil
}
