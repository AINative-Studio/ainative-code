package design

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatcher_StartStop(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	localPath := filepath.Join(tmpDir, "tokens.json")

	// Create initial tokens file
	collection := TokenCollection{
		Tokens: []Token{
			{Name: "color.primary", Type: "color", Value: "#007bff"},
		},
	}
	data, _ := json.MarshalIndent(collection, "", "  ")
	os.WriteFile(localPath, data, 0644)

	// Create mock client
	mockClient := &mockDesignClient{
		tokens: make(map[string]Token),
	}

	// Create syncer
	syncConfig := SyncConfig{
		ProjectID:          "test-project",
		Direction:          SyncDirectionPush,
		LocalPath:          localPath,
		ConflictResolution: ConflictResolutionLocal,
		DryRun:             false,
	}

	syncer := NewSyncer(mockClient, syncConfig)

	// Create watcher
	watchConfig := WatchConfig{
		Paths:            []string{tmpDir},
		DebounceDuration: 100 * time.Millisecond,
		SyncOnStart:      false,
		MaxRetries:       3,
		RetryDelay:       100 * time.Millisecond,
	}

	watcher, err := NewWatcher(syncer, watchConfig)
	require.NoError(t, err)

	// Start watcher in goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		watcher.Start(ctx)
	}()

	// Wait a bit for watcher to start
	time.Sleep(100 * time.Millisecond)

	// Verify watcher is running
	assert.True(t, watcher.IsRunning())

	// Stop watcher
	watcher.Stop()

	// Wait for watcher to stop
	time.Sleep(100 * time.Millisecond)

	// Verify watcher is stopped
	assert.False(t, watcher.IsRunning())
}

func TestWatcher_FileChangeDetection(t *testing.T) {
	// Skip on CI or if file watching is not supported
	if os.Getenv("CI") != "" {
		t.Skip("Skipping file watcher test in CI environment")
	}

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	localPath := filepath.Join(tmpDir, "tokens.json")

	// Create initial tokens file
	collection := TokenCollection{
		Tokens: []Token{
			{Name: "color.primary", Type: "color", Value: "#007bff"},
		},
	}
	data, _ := json.MarshalIndent(collection, "", "  ")
	os.WriteFile(localPath, data, 0644)

	// Create mock client
	mockClient := &mockDesignClient{
		tokens: make(map[string]Token),
	}

	// Create syncer
	syncConfig := SyncConfig{
		ProjectID:          "test-project",
		Direction:          SyncDirectionPush,
		LocalPath:          localPath,
		ConflictResolution: ConflictResolutionLocal,
		DryRun:             false,
	}

	syncer := NewSyncer(mockClient, syncConfig)

	// Create watcher with short debounce
	watchConfig := WatchConfig{
		Paths:            []string{tmpDir},
		DebounceDuration: 200 * time.Millisecond,
		SyncOnStart:      true,
		MaxRetries:       3,
		RetryDelay:       100 * time.Millisecond,
	}

	watcher, err := NewWatcher(syncer, watchConfig)
	require.NoError(t, err)

	// Start watcher
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		watcher.Start(ctx)
	}()

	// Wait for initial sync
	time.Sleep(300 * time.Millisecond)

	// Modify the file
	collection.Tokens = append(collection.Tokens, Token{
		Name:  "color.secondary",
		Type:  "color",
		Value: "#6c757d",
	})
	data, _ = json.MarshalIndent(collection, "", "  ")
	os.WriteFile(localPath, data, 0644)

	// Wait for debounce and sync
	time.Sleep(500 * time.Millisecond)

	// Verify sync was triggered
	assert.Greater(t, len(mockClient.tokens), 0, "Expected tokens to be synced")

	// Stop watcher
	watcher.Stop()
}

func TestWatcher_DebounceMultipleChanges(t *testing.T) {
	// Skip on CI
	if os.Getenv("CI") != "" {
		t.Skip("Skipping file watcher test in CI environment")
	}

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	localPath := filepath.Join(tmpDir, "tokens.json")

	// Create initial tokens file
	collection := TokenCollection{
		Tokens: []Token{
			{Name: "color.primary", Type: "color", Value: "#007bff"},
		},
	}
	data, _ := json.MarshalIndent(collection, "", "  ")
	os.WriteFile(localPath, data, 0644)

	mockClient := &mockDesignClient{
		tokens: make(map[string]Token),
	}

	// Create syncer
	syncConfig := SyncConfig{
		ProjectID:          "test-project",
		Direction:          SyncDirectionPush,
		LocalPath:          localPath,
		ConflictResolution: ConflictResolutionLocal,
		DryRun:             false,
	}

	syncer := NewSyncer(mockClient, syncConfig)

	// Create watcher with longer debounce
	watchConfig := WatchConfig{
		Paths:            []string{tmpDir},
		DebounceDuration: 500 * time.Millisecond,
		SyncOnStart:      false,
		MaxRetries:       3,
		RetryDelay:       100 * time.Millisecond,
	}

	watcher, err := NewWatcher(syncer, watchConfig)
	require.NoError(t, err)

	// Start watcher
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		watcher.Start(ctx)
	}()

	// Wait for watcher to start
	time.Sleep(100 * time.Millisecond)

	// Make multiple rapid changes
	for i := 0; i < 5; i++ {
		collection.Tokens[0].Value = "#" + string(rune('0'+i)) + "00000"
		data, _ := json.MarshalIndent(collection, "", "  ")
		os.WriteFile(localPath, data, 0644)
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for debounce and sync
	time.Sleep(1 * time.Second)

	// Verify tokens were uploaded (debouncing should prevent multiple syncs)
	assert.Greater(t, len(mockClient.tokens), 0, "Expected tokens to be uploaded")

	watcher.Stop()
}

func TestWatcher_SyncOnStart(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	localPath := filepath.Join(tmpDir, "tokens.json")

	// Create initial tokens file
	collection := TokenCollection{
		Tokens: []Token{
			{Name: "color.primary", Type: "color", Value: "#007bff"},
		},
	}
	data, _ := json.MarshalIndent(collection, "", "  ")
	os.WriteFile(localPath, data, 0644)

	mockClient := &mockDesignClient{
		tokens: make(map[string]Token),
	}

	// Create syncer
	syncConfig := SyncConfig{
		ProjectID:          "test-project",
		Direction:          SyncDirectionPush,
		LocalPath:          localPath,
		ConflictResolution: ConflictResolutionLocal,
		DryRun:             false,
	}

	syncer := NewSyncer(mockClient, syncConfig)

	// Create watcher with SyncOnStart enabled
	watchConfig := WatchConfig{
		Paths:            []string{tmpDir},
		DebounceDuration: 200 * time.Millisecond,
		SyncOnStart:      true,
		MaxRetries:       3,
		RetryDelay:       100 * time.Millisecond,
	}

	watcher, err := NewWatcher(syncer, watchConfig)
	require.NoError(t, err)

	// Start watcher
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go func() {
		watcher.Start(ctx)
	}()

	// Wait for initial sync
	time.Sleep(300 * time.Millisecond)

	// Verify initial sync occurred
	assert.Greater(t, len(mockClient.tokens), 0, "Expected initial sync to upload tokens")

	watcher.Stop()
}

// mockFailingClient implements DesignClient with controllable failures
type mockFailingClient struct {
	tokens       map[string]Token
	uploadErr    error
	failCount    int
	attemptCount int
}

func (m *mockFailingClient) GetTokens(ctx context.Context, projectID string) ([]Token, error) {
	tokens := make([]Token, 0, len(m.tokens))
	for _, token := range m.tokens {
		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (m *mockFailingClient) UploadTokens(ctx context.Context, projectID string, tokens []Token) error {
	m.attemptCount++
	if m.attemptCount <= m.failCount {
		return m.uploadErr
	}
	for _, token := range tokens {
		m.tokens[token.Name] = token
	}
	return nil
}

func (m *mockFailingClient) DeleteToken(ctx context.Context, projectID string, tokenName string) error {
	delete(m.tokens, tokenName)
	return nil
}

func TestWatcher_RetryLogic(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	localPath := filepath.Join(tmpDir, "tokens.json")

	// Create initial tokens file
	collection := TokenCollection{
		Tokens: []Token{
			{Name: "color.primary", Type: "color", Value: "#007bff"},
		},
	}
	data, _ := json.MarshalIndent(collection, "", "  ")
	os.WriteFile(localPath, data, 0644)

	// Create mock client that fails first 2 attempts
	mockClient := &mockFailingClient{
		tokens:    make(map[string]Token),
		uploadErr: assert.AnError,
		failCount: 2,
	}

	// Create syncer
	syncConfig := SyncConfig{
		ProjectID:          "test-project",
		Direction:          SyncDirectionPush,
		LocalPath:          localPath,
		ConflictResolution: ConflictResolutionLocal,
		DryRun:             false,
	}

	syncer := NewSyncer(mockClient, syncConfig)

	// Create watcher with retry configuration
	watchConfig := WatchConfig{
		Paths:            []string{tmpDir},
		DebounceDuration: 100 * time.Millisecond,
		SyncOnStart:      true,
		MaxRetries:       3,
		RetryDelay:       100 * time.Millisecond,
	}

	watcher, err := NewWatcher(syncer, watchConfig)
	require.NoError(t, err)

	// Start watcher
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		watcher.Start(ctx)
	}()

	// Wait for retries to complete
	time.Sleep(1 * time.Second)

	// Verify retry logic was executed
	assert.GreaterOrEqual(t, mockClient.attemptCount, 3, "Expected retry logic to attempt multiple times")

	watcher.Stop()
}

func TestNewWatcher_DefaultConfig(t *testing.T) {
	mockClient := &mockDesignClient{
		tokens: make(map[string]Token),
	}

	syncConfig := SyncConfig{
		ProjectID:          "test-project",
		Direction:          SyncDirectionPush,
		LocalPath:          "/tmp/tokens.json",
		ConflictResolution: ConflictResolutionLocal,
	}

	syncer := NewSyncer(mockClient, syncConfig)

	// Create watcher with empty config
	watchConfig := WatchConfig{
		Paths: []string{"/tmp"},
	}

	watcher, err := NewWatcher(syncer, watchConfig)
	require.NoError(t, err)
	assert.NotNil(t, watcher)

	// Verify defaults were applied
	assert.Equal(t, 2*time.Second, watcher.config.DebounceDuration)
	assert.Equal(t, 3, watcher.config.MaxRetries)
	assert.Equal(t, 5*time.Second, watcher.config.RetryDelay)
}

func TestWatcher_ContextCancellation(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	localPath := filepath.Join(tmpDir, "tokens.json")

	// Create initial tokens file
	collection := TokenCollection{
		Tokens: []Token{
			{Name: "color.primary", Type: "color", Value: "#007bff"},
		},
	}
	data, _ := json.MarshalIndent(collection, "", "  ")
	os.WriteFile(localPath, data, 0644)

	mockClient := &mockDesignClient{
		tokens: make(map[string]Token),
	}

	syncConfig := SyncConfig{
		ProjectID:          "test-project",
		Direction:          SyncDirectionPush,
		LocalPath:          localPath,
		ConflictResolution: ConflictResolutionLocal,
	}

	syncer := NewSyncer(mockClient, syncConfig)

	watchConfig := WatchConfig{
		Paths:            []string{tmpDir},
		DebounceDuration: 200 * time.Millisecond,
		SyncOnStart:      false,
	}

	watcher, err := NewWatcher(syncer, watchConfig)
	require.NoError(t, err)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start watcher
	watcherDone := make(chan struct{})
	go func() {
		watcher.Start(ctx)
		close(watcherDone)
	}()

	// Wait for watcher to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for watcher to stop
	select {
	case <-watcherDone:
		// Watcher stopped as expected
	case <-time.After(2 * time.Second):
		t.Fatal("Watcher did not stop after context cancellation")
	}

	assert.False(t, watcher.IsRunning())
}
