package anthropic

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// sseReader reads Server-Sent Events from an io.Reader
type sseReader struct {
	scanner *bufio.Scanner
}

// newSSEReader creates a new SSE reader
func newSSEReader(r io.Reader) *sseReader {
	return &sseReader{
		scanner: bufio.NewScanner(r),
	}
}

// readEvent reads the next SSE event from the stream
func (r *sseReader) readEvent() (*sseEvent, error) {
	var event sseEvent
	var dataLines []string

	for r.scanner.Scan() {
		line := r.scanner.Text()

		// Empty line indicates end of event
		if line == "" {
			if event.eventType != "" || len(dataLines) > 0 {
				event.data = strings.Join(dataLines, "\n")
				return &event, nil
			}
			continue
		}

		// Parse SSE field
		if strings.HasPrefix(line, "event:") {
			event.eventType = strings.TrimSpace(line[6:])
		} else if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(line[5:]))
		}
		// Ignore other fields (id, retry, comments)
	}

	// Check for scanner errors
	if err := r.scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	// If we get here, we've reached EOF
	if event.eventType != "" || len(dataLines) > 0 {
		event.data = strings.Join(dataLines, "\n")
		return &event, nil
	}

	return nil, io.EOF
}
