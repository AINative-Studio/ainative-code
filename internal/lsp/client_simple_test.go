package lsp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewClient tests creating a new client
func TestNewClient(t *testing.T) {
	config := DefaultConfig("go")
	client, err := NewClient(config)

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "go", client.config.Language)
	assert.NotNil(t, client.responses)
}

// TestNewClientInvalidConfig tests creating a client with invalid config
func TestNewClientInvalidConfig(t *testing.T) {
	config := &LanguageServerConfig{
		Language: "", // Invalid - missing language
		Command:  "gopls",
	}

	_, err := NewClient(config)
	assert.Error(t, err)
}

// TestClientGetNextID tests ID generation
func TestClientGetNextID(t *testing.T) {
	config := DefaultConfig("go")
	client, err := NewClient(config)
	require.NoError(t, err)

	id1 := client.getNextID()
	assert.Equal(t, int64(1), id1)

	id2 := client.getNextID()
	assert.Equal(t, int64(2), id2)

	id3 := client.getNextID()
	assert.Equal(t, int64(3), id3)
}

// TestClientIsInitialized tests initialization state
func TestClientIsInitialized(t *testing.T) {
	config := DefaultConfig("go")
	client, err := NewClient(config)
	require.NoError(t, err)

	assert.False(t, client.isInitialized())

	client.initialized = true
	assert.True(t, client.isInitialized())
}

// TestClientIsShutdown tests shutdown state
func TestClientIsShutdown(t *testing.T) {
	config := DefaultConfig("go")
	client, err := NewClient(config)
	require.NoError(t, err)

	assert.False(t, client.IsShutdown())

	client.shutdown = true
	assert.True(t, client.IsShutdown())
}

// TestClientBuildCapabilities tests client capabilities
func TestClientBuildCapabilities(t *testing.T) {
	config := DefaultConfig("go")
	client, err := NewClient(config)
	require.NoError(t, err)

	caps := client.buildClientCapabilities()

	assert.True(t, caps.Workspace.ApplyEdit)
	assert.NotNil(t, caps.TextDocument.Completion)
	assert.NotNil(t, caps.TextDocument.Hover)
	assert.NotNil(t, caps.TextDocument.Definition)
	assert.NotNil(t, caps.TextDocument.References)
}

// TestClientClose tests closing a client
func TestClientClose(t *testing.T) {
	config := DefaultConfig("go")
	client, err := NewClient(config)
	require.NoError(t, err)

	// Close should not error even if client was never started
	err = client.Close()
	assert.NoError(t, err)
}

// TestClientOperationsBeforeInitialization tests that operations fail before initialization
func TestClientOperationsBeforeInitialization(t *testing.T) {
	config := &LanguageServerConfig{
		Language:       "go",
		Command:        "mock-lsp",
		InitTimeout:    5 * time.Second,
		RequestTimeout: 2 * time.Second,
		Env:            make(map[string]string),
	}

	client, err := NewClient(config)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	// These should fail because server is not initialized
	_, err = client.Completion(ctx, CompletionParams{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.Hover(ctx, HoverParams{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.Definition(ctx, DefinitionParams{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.References(ctx, ReferenceParams{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// TestClientMultipleInitialize tests that multiple initialize calls fail
func TestClientMultipleInitialize(t *testing.T) {
	// This test verifies the logic, actual initialization would require a real server
	config := &LanguageServerConfig{
		Language:       "go",
		Command:        "mock-lsp",
		InitTimeout:    5 * time.Second,
		RequestTimeout: 100 * time.Millisecond, // Short timeout for quick failure
		Env:            make(map[string]string),
	}

	client, err := NewClient(config)
	require.NoError(t, err)
	defer client.Close()

	// Manually set initialized to simulate successful initialization
	client.initialized = true

	ctx := context.Background()
	rootURI := "file:///workspace"

	// Second initialize should fail
	_, err = client.Initialize(ctx, &rootURI, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already initialized")
}
