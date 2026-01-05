package lsp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("creates client with default configuration", func(t *testing.T) {
		client := NewClient()
		require.NotNil(t, client)
		assert.NotNil(t, client.config)
		assert.Equal(t, 500*time.Millisecond, client.config.CompletionDebounce)
		assert.Equal(t, 5*time.Second, client.config.RequestTimeout)
		assert.True(t, client.config.EnableCache)
	})

	t.Run("creates client with custom configuration", func(t *testing.T) {
		config := Config{
			CompletionDebounce: 300 * time.Millisecond,
			RequestTimeout:     3 * time.Second,
			EnableCache:        false,
		}
		client := NewClientWithConfig(config)
		require.NotNil(t, client)
		assert.Equal(t, config.CompletionDebounce, client.config.CompletionDebounce)
		assert.Equal(t, config.RequestTimeout, client.config.RequestTimeout)
		assert.False(t, client.config.EnableCache)
	})
}

func TestClient_Initialize(t *testing.T) {
	t.Run("initializes successfully with valid config", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		err := client.Initialize(ctx, "/Users/aideveloper/AINative-Code")
		assert.NoError(t, err)
		assert.True(t, client.IsConnected())
	})

	t.Run("fails with invalid workspace path", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		err := client.Initialize(ctx, "/nonexistent/path/that/does/not/exist")
		assert.Error(t, err)
		assert.False(t, client.IsConnected())
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		client := NewClient()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := client.Initialize(ctx, "/Users/aideveloper/AINative-Code")
		assert.Error(t, err)
		assert.False(t, client.IsConnected())
	})
}

func TestClient_Completion(t *testing.T) {
	t.Run("returns completion items for valid position", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		_ = client.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		params := CompletionParams{
			TextDocument: TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: Position{
				Line:      10,
				Character: 5,
			},
		}

		items, err := client.GetCompletion(ctx, params)
		assert.NoError(t, err)
		assert.NotNil(t, items)
	})

	t.Run("handles timeout gracefully", func(t *testing.T) {
		client := NewClient()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		params := CompletionParams{
			TextDocument: TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: Position{
				Line:      10,
				Character: 5,
			},
		}

		_, err := client.GetCompletion(ctx, params)
		assert.Error(t, err)
	})

	t.Run("uses cache for repeated requests", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		_ = client.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		params := CompletionParams{
			TextDocument: TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: Position{
				Line:      10,
				Character: 5,
			},
		}

		// First request
		items1, err1 := client.GetCompletion(ctx, params)
		assert.NoError(t, err1)

		// Second request (should use cache)
		items2, err2 := client.GetCompletion(ctx, params)
		assert.NoError(t, err2)
		assert.Equal(t, items1, items2)
	})
}

func TestClient_Hover(t *testing.T) {
	t.Run("returns hover information for valid symbol", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		_ = client.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		params := HoverParams{
			TextDocument: TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: Position{
				Line:      30,
				Character: 10,
			},
		}

		hover, err := client.GetHover(ctx, params)
		assert.NoError(t, err)
		assert.NotNil(t, hover)
	})

	t.Run("returns nil for position without symbol", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		_ = client.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		params := HoverParams{
			TextDocument: TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: Position{
				Line:      0,
				Character: 0,
			},
		}

		hover, err := client.GetHover(ctx, params)
		assert.NoError(t, err)
		assert.Nil(t, hover)
	})
}

func TestClient_Definition(t *testing.T) {
	t.Run("returns definition location for valid symbol", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		_ = client.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		params := DefinitionParams{
			TextDocument: TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: Position{
				Line:      30,
				Character: 10,
			},
		}

		locations, err := client.GetDefinition(ctx, params)
		assert.NoError(t, err)
		assert.NotNil(t, locations)
	})

	t.Run("returns empty array for position without definition", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		_ = client.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		params := DefinitionParams{
			TextDocument: TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: Position{
				Line:      0,
				Character: 0,
			},
		}

		locations, err := client.GetDefinition(ctx, params)
		assert.NoError(t, err)
		assert.Empty(t, locations)
	})
}

func TestClient_References(t *testing.T) {
	t.Run("returns references for valid symbol", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		_ = client.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		params := ReferencesParams{
			TextDocument: TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: Position{
				Line:      30,
				Character: 10,
			},
			Context: ReferencesContext{
				IncludeDeclaration: true,
			},
		}

		locations, err := client.GetReferences(ctx, params)
		assert.NoError(t, err)
		assert.NotNil(t, locations)
	})

	t.Run("returns empty array for position without references", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		_ = client.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		params := ReferencesParams{
			TextDocument: TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: Position{
				Line:      0,
				Character: 0,
			},
			Context: ReferencesContext{
				IncludeDeclaration: false,
			},
		}

		locations, err := client.GetReferences(ctx, params)
		assert.NoError(t, err)
		assert.Empty(t, locations)
	})
}

func TestClient_Shutdown(t *testing.T) {
	t.Run("shuts down successfully", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		_ = client.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		err := client.Shutdown(ctx)
		assert.NoError(t, err)
		assert.False(t, client.IsConnected())
	})

	t.Run("handles multiple shutdown calls", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		_ = client.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		err1 := client.Shutdown(ctx)
		assert.NoError(t, err1)

		err2 := client.Shutdown(ctx)
		assert.NoError(t, err2) // Should not error on second call
	})
}

func TestClient_CancelRequest(t *testing.T) {
	t.Run("cancels pending request", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		_ = client.Initialize(ctx, "/Users/aideveloper/AINative-Code")

		// Start a request
		ctx, cancel := context.WithCancel(ctx)
		params := CompletionParams{
			TextDocument: TextDocumentIdentifier{
				URI: "file:///Users/aideveloper/AINative-Code/internal/tui/model.go",
			},
			Position: Position{
				Line:      10,
				Character: 5,
			},
		}

		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		_, err := client.GetCompletion(ctx, params)
		assert.Error(t, err) // Should error due to cancellation
	})
}
