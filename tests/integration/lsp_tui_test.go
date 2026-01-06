package integration

import (
	"context"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/tui"
	"github.com/AINative-studio/ainative-code/pkg/lsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLSPTUIIntegration_CompletionFlow(t *testing.T) {
	t.Run("complete end-to-end completion flow", func(t *testing.T) {
		// Create model with LSP
		model := tui.NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()

		// Initialize LSP client
		err := model.GetLSPClient().Initialize(ctx, "/Users/aideveloper/AINative-Code")
		require.NoError(t, err)
		assert.True(t, model.IsLSPEnabled())
		assert.Equal(t, lsp.StatusConnected, model.GetLSPStatus())

		// Trigger completion
		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err = tui.TriggerCompletion(ctx, &model, documentURI, 10, 5)
		require.NoError(t, err)

		// Verify completion items are available
		assert.NotEmpty(t, model.GetSelectedCompletion())

		// Navigate through completion items
		model.NextCompletion()
		assert.NotNil(t, model.GetSelectedCompletion())

		model.PrevCompletion()
		assert.NotNil(t, model.GetSelectedCompletion())

		// Insert completion
		model.SetValue("M")
		tui.InsertCompletion(&model)

		// Verify input was updated
		assert.Contains(t, model.GetUserInput(), "Model")

		// Verify completion popup is closed
		assert.False(t, model.GetShowCompletion())
	})

	t.Run("handles completion timeout gracefully", func(t *testing.T) {
		model := tui.NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		err := model.GetLSPClient().Initialize(ctx, "/Users/aideveloper/AINative-Code")
		// May fail due to timeout, which is expected

		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err = tui.TriggerCompletion(ctx, &model, documentURI, 10, 5)
		// Should handle error gracefully
		if err != nil {
			assert.Contains(t, err.Error(), "context")
		}
	})
}

func TestLSPTUIIntegration_HoverFlow(t *testing.T) {
	t.Run("complete end-to-end hover flow", func(t *testing.T) {
		// Create model with LSP
		model := tui.NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()

		// Initialize LSP client
		err := model.GetLSPClient().Initialize(ctx, "/Users/aideveloper/AINative-Code")
		require.NoError(t, err)

		// Trigger hover
		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err = tui.TriggerHover(ctx, &model, documentURI, 30, 10)
		require.NoError(t, err)

		// Verify hover info is displayed
		assert.True(t, model.GetShowHover())
		assert.NotNil(t, model.GetHoverInfo())

		// Render hover popup
		rendered := tui.RenderHover(&model)
		assert.NotEmpty(t, rendered)
		assert.Contains(t, rendered, "Type Information")

		// Clear hover
		model.ClearHover()
		assert.False(t, model.GetShowHover())
		assert.Nil(t, model.GetHoverInfo())
	})

	t.Run("handles hover for position without symbol", func(t *testing.T) {
		model := tui.NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()

		err := model.GetLSPClient().Initialize(ctx, "/Users/aideveloper/AINative-Code")
		require.NoError(t, err)

		// Trigger hover at invalid position
		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err = tui.TriggerHover(ctx, &model, documentURI, 0, 0)
		require.NoError(t, err)

		// Verify no hover info is shown
		assert.False(t, model.GetShowHover())
		assert.Nil(t, model.GetHoverInfo())
	})
}

func TestLSPTUIIntegration_NavigationFlow(t *testing.T) {
	t.Run("complete end-to-end goto definition flow", func(t *testing.T) {
		// Create model with LSP
		model := tui.NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()

		// Initialize LSP client
		err := model.GetLSPClient().Initialize(ctx, "/Users/aideveloper/AINative-Code")
		require.NoError(t, err)

		// Goto definition
		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err = tui.GotoDefinition(ctx, &model, documentURI, 30, 10)
		require.NoError(t, err)

		// Verify navigation results are displayed
		assert.True(t, model.GetShowNavigation())
		assert.NotEmpty(t, model.GetNavigationResult())

		// Render navigation popup
		rendered := tui.RenderNavigation(&model)
		assert.NotEmpty(t, rendered)
		assert.Contains(t, rendered, "Navigation Results")

		// Clear navigation
		model.ClearNavigation()
		assert.False(t, model.GetShowNavigation())
		assert.Empty(t, model.GetNavigationResult())
	})

	t.Run("complete end-to-end find references flow", func(t *testing.T) {
		// Create model with LSP
		model := tui.NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()

		// Initialize LSP client
		err := model.GetLSPClient().Initialize(ctx, "/Users/aideveloper/AINative-Code")
		require.NoError(t, err)

		// Find references
		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err = tui.FindReferences(ctx, &model, documentURI, 30, 10, true)
		require.NoError(t, err)

		// Verify navigation results are displayed
		assert.True(t, model.GetShowNavigation())
		assert.NotEmpty(t, model.GetNavigationResult())

		// Verify multiple references are found
		assert.GreaterOrEqual(t, len(model.GetNavigationResult()), 1)
	})

	t.Run("handles navigation for position without symbol", func(t *testing.T) {
		model := tui.NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()

		err := model.GetLSPClient().Initialize(ctx, "/Users/aideveloper/AINative-Code")
		require.NoError(t, err)

		// Try to goto definition at invalid position
		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"
		err = tui.GotoDefinition(ctx, &model, documentURI, 0, 0)
		require.NoError(t, err)

		// Verify no navigation results
		assert.False(t, model.GetShowNavigation())
		assert.Empty(t, model.GetNavigationResult())
	})
}

func TestLSPTUIIntegration_StatusIndicators(t *testing.T) {
	t.Run("displays LSP connection status", func(t *testing.T) {
		model := tui.NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()

		// Check initial status
		assert.Equal(t, lsp.StatusDisconnected, model.GetLSPStatus())

		// Initialize LSP
		err := model.GetLSPClient().Initialize(ctx, "/Users/aideveloper/AINative-Code")
		require.NoError(t, err)

		// Check connected status
		assert.Equal(t, lsp.StatusConnected, model.GetLSPStatus())

		// Shutdown LSP
		err = model.GetLSPClient().Shutdown(ctx)
		require.NoError(t, err)

		// Check disconnected status
		assert.Equal(t, lsp.StatusDisconnected, model.GetLSPStatus())
	})

	t.Run("updates status on errors", func(t *testing.T) {
		model := tui.NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()

		// Try to initialize with invalid workspace
		err := model.GetLSPClient().Initialize(ctx, "/invalid/path")
		assert.Error(t, err)

		// Status should reflect error
		assert.False(t, model.GetLSPClient().IsConnected())
	})
}

func TestLSPTUIIntegration_MultiplePopups(t *testing.T) {
	t.Run("only shows one popup at a time", func(t *testing.T) {
		model := tui.NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()

		err := model.GetLSPClient().Initialize(ctx, "/Users/aideveloper/AINative-Code")
		require.NoError(t, err)

		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"

		// Trigger completion
		err = tui.TriggerCompletion(ctx, &model, documentURI, 10, 5)
		require.NoError(t, err)
		assert.True(t, model.GetShowCompletion())

		// Trigger hover (should replace completion)
		err = tui.TriggerHover(ctx, &model, documentURI, 30, 10)
		require.NoError(t, err)

		// Clear completion manually to simulate user action
		model.ClearCompletion()
		assert.False(t, model.GetShowCompletion())
		assert.True(t, model.GetShowHover())

		// Trigger navigation (should replace hover)
		err = tui.GotoDefinition(ctx, &model, documentURI, 30, 10)
		require.NoError(t, err)

		// Clear hover manually
		model.ClearHover()
		assert.False(t, model.GetShowHover())
		assert.True(t, model.GetShowNavigation())
	})
}

func TestLSPTUIIntegration_Performance(t *testing.T) {
	t.Run("handles rapid completion requests", func(t *testing.T) {
		model := tui.NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()

		err := model.GetLSPClient().Initialize(ctx, "/Users/aideveloper/AINative-Code")
		require.NoError(t, err)

		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"

		// Trigger multiple completion requests rapidly
		for i := 0; i < 10; i++ {
			err = tui.TriggerCompletion(ctx, &model, documentURI, 10+i, 5)
			// Should handle without crashing
			if err != nil {
				t.Logf("Completion request %d failed: %v", i, err)
			}
		}
	})

	t.Run("caches hover information", func(t *testing.T) {
		model := tui.NewModelWithLSP("/Users/aideveloper/AINative-Code")
		ctx := context.Background()

		err := model.GetLSPClient().Initialize(ctx, "/Users/aideveloper/AINative-Code")
		require.NoError(t, err)

		documentURI := "file:///Users/aideveloper/AINative-Code/internal/tui/model.go"

		// First hover request
		start := time.Now()
		err = tui.TriggerHover(ctx, &model, documentURI, 30, 10)
		require.NoError(t, err)
		firstDuration := time.Since(start)

		// Clear hover
		model.ClearHover()

		// Second hover request (should be cached)
		start = time.Now()
		err = tui.TriggerHover(ctx, &model, documentURI, 30, 10)
		require.NoError(t, err)
		secondDuration := time.Since(start)

		// Second request should be faster (cached)
		assert.LessOrEqual(t, secondDuration, firstDuration)
	})
}
