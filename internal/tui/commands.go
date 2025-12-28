package tui

import (
	"context"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ProcessUserInput creates a command that processes user input and streams the response
func ProcessUserInput(input string, apiKey string) tea.Cmd {
	return func() tea.Msg {
		// Create Anthropic client
		client := anthropic.NewClient(
			option.WithAPIKey(apiKey),
		)

		// Create streaming request
		ctx := context.Background()
		stream := client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
			Model: anthropic.F(anthropic.ModelClaude_3_5_Sonnet_20241022),
			Messages: anthropic.F([]anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(input)),
			}),
			MaxTokens: anthropic.Int(1024),
		})

		// Process stream
		for stream.Next() {
			event := stream.Current()

			switch delta := event.Delta.(type) {
			case anthropic.ContentBlockDeltaEventDelta:
				if textDelta, ok := delta.AsUnion().(anthropic.TextDelta); ok {
					// Send chunk to UI
					return streamChunkMsg{content: textDelta.Text}
				}
			}
		}

		// Check for streaming errors
		if err := stream.Err(); err != nil {
			return errMsg{err: fmt.Errorf("streaming error: %w", err)}
		}

		// Signal completion
		return streamDoneMsg{}
	}
}

// StreamResponse creates a command that streams API responses
// This is a more generic streaming handler that can be composed
func StreamResponse(reader io.Reader) tea.Cmd {
	return func() tea.Msg {
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				// Send chunk to UI
				return streamChunkMsg{content: string(buf[:n])}
			}
			if err == io.EOF {
				// End of stream
				return streamDoneMsg{}
			}
			if err != nil {
				// Stream error
				return errMsg{err: fmt.Errorf("stream read error: %w", err)}
			}
		}
	}
}

// HandleAPIError creates a command that handles API errors
func HandleAPIError(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg{err: err}
	}
}

// WaitForStream creates a command that waits for streaming to complete
// This can be used for synchronization or timeout handling
func WaitForStream(done <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		<-done
		return streamDoneMsg{}
	}
}

// BatchStreamChunks creates a command that batches multiple stream chunks
// This can improve performance for high-frequency streaming
func BatchStreamChunks(chunks []string) tea.Cmd {
	return func() tea.Msg {
		combined := ""
		for _, chunk := range chunks {
			combined += chunk
		}
		return streamChunkMsg{content: combined}
	}
}
