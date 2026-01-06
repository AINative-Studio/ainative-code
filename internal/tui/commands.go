package tui

import (
	"context"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/AINative-studio/ainative-code/internal/provider/anthropic"
)

// ProcessUserInput creates a command that processes user input and streams the response
func ProcessUserInput(input string, apiKey string) tea.Cmd {
	return func() tea.Msg {
		// Create Anthropic provider
		config := anthropic.Config{
			APIKey: apiKey,
		}

		client, err := anthropic.NewAnthropicProvider(config)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to create provider: %w", err)}
		}
		defer client.Close()

		// Create messages
		messages := []provider.Message{
			{
				Role:    "user",
				Content: input,
			},
		}

		// Create streaming request
		ctx := context.Background()
		streamChan, err := client.Stream(
			ctx,
			messages,
			provider.StreamWithModel("claude-3-5-sonnet-20241022"),
			provider.StreamWithMaxTokens(1024),
		)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to start stream: %w", err)}
		}

		// Process stream
		for event := range streamChan {
			if event.Error != nil {
				return errMsg{err: fmt.Errorf("streaming error: %w", event.Error)}
			}

			if event.Content != "" {
				// Send chunk to UI
				return streamChunkMsg{content: event.Content}
			}
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
