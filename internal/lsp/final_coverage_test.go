package lsp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClientExitCall tests calling Exit
func TestClientExitCall(t *testing.T) {
	config := DefaultConfig("go")
	client, err := NewClient(config)
	require.NoError(t, err)
	defer client.Close()

	// Without stdin set, this should fail
	ctx := context.Background()
	err = client.Exit(ctx)
	assert.Error(t, err) // Will fail because stdin is nil
}

// TestConfigMergeWithZeroValues tests merging with zero values
func TestConfigMergeWithZeroValues(t *testing.T) {
	base := DefaultConfig("go")
	original := base.InitTimeout

	override := &LanguageServerConfig{
		Language:       "", // Empty - should not override
		Command:        "", // Empty - should not override
		InitTimeout:    0,  // Zero - should not override
		RequestTimeout: 15 * time.Second, // Non-zero - should override
	}

	base.Merge(override)

	assert.Equal(t, original, base.InitTimeout) // Not overridden
	assert.Equal(t, 15*time.Second, base.RequestTimeout) // Overridden
}

// TestConfigMergeWithInitOptions tests merging initialization options
func TestConfigMergeWithInitOptions(t *testing.T) {
	base := DefaultConfig("go")

	newOptions := map[string]interface{}{"test": "value"}
	override := &LanguageServerConfig{
		InitializationOptions: newOptions,
	}

	base.Merge(override)
	assert.Equal(t, newOptions, base.InitializationOptions)
}

// TestConfigMergeHealthCheckInterval tests merging health check interval
func TestConfigMergeHealthCheckInterval(t *testing.T) {
	base := DefaultConfig("go")

	override := &LanguageServerConfig{
		HealthCheckInterval: 120 * time.Second,
	}

	base.Merge(override)
	assert.Equal(t, 120*time.Second, base.HealthCheckInterval)
}

// TestConfigCloneDeepCopy tests that clone is truly a deep copy
func TestConfigCloneDeepCopy(t *testing.T) {
	original := DefaultConfig("go")
	original.Args = []string{"serve", "--debug"}
	original.Env = map[string]string{"KEY1": "value1", "KEY2": "value2"}

	clone := original.Clone()

	// Modify clone
	clone.Args[0] = "different"
	clone.Env["KEY1"] = "modified"
	clone.Env["KEY3"] = "new"

	// Original should be unchanged
	assert.Equal(t, "serve", original.Args[0])
	assert.Equal(t, "value1", original.Env["KEY1"])
	assert.NotContains(t, original.Env, "KEY3")
}

// TestManagerStartHealthCheckZeroInterval tests health check with zero interval
func TestManagerStartHealthCheckZeroInterval(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := DefaultConfig("go")
	config.HealthCheckInterval = 0 // Zero interval - should not start health check

	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// Verify no health check was started
	manager.healthMu.RLock()
	_, exists := manager.healthChecks["go"]
	manager.healthMu.RUnlock()

	assert.False(t, exists)
}

// TestManagerHealthCheckAlreadyRunning tests starting health check twice
func TestManagerHealthCheckAlreadyRunning(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := DefaultConfig("go")
	config.HealthCheckInterval = 1 * time.Hour // Long interval

	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// Manually start health check
	manager.startHealthCheck("go")

	// Verify it's running
	manager.healthMu.RLock()
	_, exists := manager.healthChecks["go"]
	manager.healthMu.RUnlock()

	assert.True(t, exists)

	// Try to start again - should be no-op
	manager.startHealthCheck("go")

	// Still only one ticker
	manager.healthMu.RLock()
	count := len(manager.healthChecks)
	manager.healthMu.RUnlock()

	assert.Equal(t, 1, count)
}

// TestManagerStopNonexistentHealthCheck tests stopping non-existent health check
func TestManagerStopNonexistentHealthCheck(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// Should not panic
	manager.stopHealthCheck("nonexistent")
}

// TestClientBuildCapabilitiesComplete tests all capability fields
func TestClientBuildCapabilitiesComplete(t *testing.T) {
	config := DefaultConfig("go")
	client, err := NewClient(config)
	require.NoError(t, err)
	defer client.Close()

	caps := client.buildClientCapabilities()

	// Verify workspace capabilities
	assert.True(t, caps.Workspace.ApplyEdit)
	assert.NotNil(t, caps.Workspace.WorkspaceEdit)
	assert.True(t, caps.Workspace.WorkspaceEdit.DocumentChanges)
	assert.NotNil(t, caps.Workspace.DidChangeConfiguration)
	assert.True(t, caps.Workspace.DidChangeConfiguration.DynamicRegistration)

	// Verify text document capabilities
	assert.NotNil(t, caps.TextDocument.Synchronization)
	assert.True(t, caps.TextDocument.Synchronization.DynamicRegistration)
	assert.True(t, caps.TextDocument.Synchronization.WillSave)
	assert.True(t, caps.TextDocument.Synchronization.WillSaveWaitUntil)
	assert.True(t, caps.TextDocument.Synchronization.DidSave)

	// Verify completion capabilities
	assert.NotNil(t, caps.TextDocument.Completion)
	assert.True(t, caps.TextDocument.Completion.DynamicRegistration)
	assert.NotNil(t, caps.TextDocument.Completion.CompletionItem)
	assert.True(t, caps.TextDocument.Completion.CompletionItem.SnippetSupport)
	assert.True(t, caps.TextDocument.Completion.CompletionItem.CommitCharactersSupport)
	assert.Contains(t, caps.TextDocument.Completion.CompletionItem.DocumentationFormat, MarkupKindMarkdown)

	// Verify hover capabilities
	assert.NotNil(t, caps.TextDocument.Hover)
	assert.True(t, caps.TextDocument.Hover.DynamicRegistration)
	assert.Contains(t, caps.TextDocument.Hover.ContentFormat, MarkupKindMarkdown)

	// Verify definition capabilities
	assert.NotNil(t, caps.TextDocument.Definition)
	assert.True(t, caps.TextDocument.Definition.DynamicRegistration)
	assert.True(t, caps.TextDocument.Definition.LinkSupport)

	// Verify references capabilities
	assert.NotNil(t, caps.TextDocument.References)
	assert.True(t, caps.TextDocument.References.DynamicRegistration)
}

// TestCompletionItemKindConstants tests that constants are defined
func TestCompletionItemKindConstants(t *testing.T) {
	assert.Equal(t, 1, CompletionItemKindText)
	assert.Equal(t, 2, CompletionItemKindMethod)
	assert.Equal(t, 3, CompletionItemKindFunction)
	assert.Equal(t, 25, CompletionItemKindTypeParameter)
}

// TestCompletionTriggerKindConstants tests trigger kind constants
func TestCompletionTriggerKindConstants(t *testing.T) {
	assert.Equal(t, 1, CompletionTriggerKindInvoked)
	assert.Equal(t, 2, CompletionTriggerKindTriggerCharacter)
	assert.Equal(t, 3, CompletionTriggerKindTriggerForIncompleteCompletions)
}

// TestErrorCodeConstants tests error code constants
func TestErrorCodeConstants(t *testing.T) {
	assert.Equal(t, -32700, ParseError)
	assert.Equal(t, -32600, InvalidRequest)
	assert.Equal(t, -32601, MethodNotFound)
	assert.Equal(t, -32602, InvalidParams)
	assert.Equal(t, -32603, InternalError)
	assert.Equal(t, -32002, ServerNotInitialized)
}

// TestManagerWithAutoRestartDisabled tests manager with auto-restart disabled
func TestManagerWithAutoRestartDisabled(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := DefaultConfig("go")
	config.AutoRestart = false

	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// Should not restart when auto-restart is disabled
	assert.False(t, manager.shouldRestart("go"))
}

// TestManagerIncrementAndReset tests restart counter
func TestManagerIncrementAndReset(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := DefaultConfig("go")
	config.AutoRestart = true
	config.MaxRestarts = 5

	err := manager.RegisterLanguage(config)
	require.NoError(t, err)

	// Initial count
	assert.Equal(t, 0, manager.GetRestartCount("go"))

	// Increment
	manager.incrementRestartCount("go")
	manager.incrementRestartCount("go")
	manager.incrementRestartCount("go")
	assert.Equal(t, 3, manager.GetRestartCount("go"))

	// Reset
	manager.resetRestartCount("go")
	assert.Equal(t, 0, manager.GetRestartCount("go"))
}
