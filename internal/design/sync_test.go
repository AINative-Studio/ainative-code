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

// mockDesignClient implements DesignClient for testing
type mockDesignClient struct {
	tokens       map[string]Token
	getTokensErr error
	uploadErr    error
	deleteErr    error
}

func (m *mockDesignClient) GetTokens(ctx context.Context, projectID string) ([]Token, error) {
	if m.getTokensErr != nil {
		return nil, m.getTokensErr
	}

	tokens := make([]Token, 0, len(m.tokens))
	for _, token := range m.tokens {
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (m *mockDesignClient) UploadTokens(ctx context.Context, projectID string, tokens []Token) error {
	if m.uploadErr != nil {
		return m.uploadErr
	}

	for _, token := range tokens {
		m.tokens[token.Name] = token
	}

	return nil
}

func (m *mockDesignClient) DeleteToken(ctx context.Context, projectID string, tokenName string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}

	delete(m.tokens, tokenName)
	return nil
}

func TestSyncer_PullTokens(t *testing.T) {
	tests := []struct {
		name           string
		remoteTokens   map[string]Token
		localTokens    map[string]Token
		expectedAdded  int
		expectedUpdate int
		expectedDelete int
	}{
		{
			name: "pull new tokens from remote",
			remoteTokens: map[string]Token{
				"color.primary":   {Name: "color.primary", Type: "color", Value: "#007bff"},
				"spacing.small":   {Name: "spacing.small", Type: "spacing", Value: "8px"},
			},
			localTokens:    map[string]Token{},
			expectedAdded:  2,
			expectedUpdate: 0,
			expectedDelete: 0,
		},
		{
			name: "pull updated tokens from remote",
			remoteTokens: map[string]Token{
				"color.primary": {Name: "color.primary", Type: "color", Value: "#0056b3"},
			},
			localTokens: map[string]Token{
				"color.primary": {Name: "color.primary", Type: "color", Value: "#007bff"},
			},
			expectedAdded:  0,
			expectedUpdate: 1,
			expectedDelete: 0,
		},
		{
			name:         "pull with local token deleted remotely",
			remoteTokens: map[string]Token{},
			localTokens: map[string]Token{
				"color.primary": {Name: "color.primary", Type: "color", Value: "#007bff"},
			},
			expectedAdded:  0,
			expectedUpdate: 0,
			expectedDelete: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test files
			tmpDir, err := os.MkdirTemp("", "sync-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			localPath := filepath.Join(tmpDir, "tokens.json")

			// Write local tokens if any
			if len(tt.localTokens) > 0 {
				localTokenSlice := make([]Token, 0, len(tt.localTokens))
				for _, token := range tt.localTokens {
					localTokenSlice = append(localTokenSlice, token)
				}
				collection := TokenCollection{Tokens: localTokenSlice}
				data, _ := json.MarshalIndent(collection, "", "  ")
				os.WriteFile(localPath, data, 0644)
			}

			// Create mock client with remote tokens
			mockClient := &mockDesignClient{
				tokens: tt.remoteTokens,
			}

			// Create syncer
			config := SyncConfig{
				ProjectID:          "test-project",
				Direction:          SyncDirectionPull,
				LocalPath:          localPath,
				ConflictResolution: ConflictResolutionRemote,
				DryRun:             false,
			}

			syncer := NewSyncer(mockClient, config)

			// Execute sync
			ctx := context.Background()
			result, err := syncer.Sync(ctx)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedAdded, result.Added)
			assert.Equal(t, tt.expectedUpdate, result.Updated)
			assert.Equal(t, tt.expectedDelete, result.Deleted)
		})
	}
}

func TestSyncer_PushTokens(t *testing.T) {
	tests := []struct {
		name           string
		localTokens    map[string]Token
		remoteTokens   map[string]Token
		expectedAdded  int
		expectedUpdate int
		expectedDelete int
	}{
		{
			name: "push new tokens to remote",
			localTokens: map[string]Token{
				"color.primary": {Name: "color.primary", Type: "color", Value: "#007bff"},
				"spacing.small": {Name: "spacing.small", Type: "spacing", Value: "8px"},
			},
			remoteTokens:   map[string]Token{},
			expectedAdded:  2,
			expectedUpdate: 0,
			expectedDelete: 0,
		},
		{
			name: "push updated tokens to remote",
			localTokens: map[string]Token{
				"color.primary": {Name: "color.primary", Type: "color", Value: "#0056b3"},
			},
			remoteTokens: map[string]Token{
				"color.primary": {Name: "color.primary", Type: "color", Value: "#007bff"},
			},
			expectedAdded:  0,
			expectedUpdate: 1,
			expectedDelete: 0,
		},
		{
			name:        "push with remote token deleted locally",
			localTokens: map[string]Token{},
			remoteTokens: map[string]Token{
				"color.primary": {Name: "color.primary", Type: "color", Value: "#007bff"},
			},
			expectedAdded:  0,
			expectedUpdate: 0,
			expectedDelete: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test files
			tmpDir, err := os.MkdirTemp("", "sync-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			localPath := filepath.Join(tmpDir, "tokens.json")

			// Write local tokens
			if len(tt.localTokens) > 0 {
				localTokenSlice := make([]Token, 0, len(tt.localTokens))
				for _, token := range tt.localTokens {
					localTokenSlice = append(localTokenSlice, token)
				}
				collection := TokenCollection{Tokens: localTokenSlice}
				data, _ := json.MarshalIndent(collection, "", "  ")
				os.WriteFile(localPath, data, 0644)
			}

			// Create mock client with remote tokens
			mockClient := &mockDesignClient{
				tokens: tt.remoteTokens,
			}

			// Create syncer
			config := SyncConfig{
				ProjectID:          "test-project",
				Direction:          SyncDirectionPush,
				LocalPath:          localPath,
				ConflictResolution: ConflictResolutionLocal,
				DryRun:             false,
			}

			syncer := NewSyncer(mockClient, config)

			// Execute sync
			ctx := context.Background()
			result, err := syncer.Sync(ctx)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedAdded, result.Added)
			assert.Equal(t, tt.expectedUpdate, result.Updated)
			assert.Equal(t, tt.expectedDelete, result.Deleted)
		})
	}
}

func TestSyncer_BidirectionalSync(t *testing.T) {
	tests := []struct {
		name             string
		localTokens      map[string]Token
		remoteTokens     map[string]Token
		expectedAdded    int
		expectedUpdated  int
		expectedConflict int
	}{
		{
			name: "bidirectional with new tokens on both sides",
			localTokens: map[string]Token{
				"color.primary": {Name: "color.primary", Type: "color", Value: "#007bff"},
			},
			remoteTokens: map[string]Token{
				"spacing.small": {Name: "spacing.small", Type: "spacing", Value: "8px"},
			},
			expectedAdded:    2,
			expectedUpdated:  0,
			expectedConflict: 0,
		},
		{
			name: "bidirectional with conflicting updates",
			localTokens: map[string]Token{
				"color.primary": {Name: "color.primary", Type: "color", Value: "#0056b3"},
			},
			remoteTokens: map[string]Token{
				"color.primary": {Name: "color.primary", Type: "color", Value: "#007bff"},
			},
			expectedAdded:    0,
			expectedUpdated:  1,
			expectedConflict: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "sync-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			localPath := filepath.Join(tmpDir, "tokens.json")

			// Write local tokens
			if len(tt.localTokens) > 0 {
				localTokenSlice := make([]Token, 0, len(tt.localTokens))
				for _, token := range tt.localTokens {
					localTokenSlice = append(localTokenSlice, token)
				}
				collection := TokenCollection{Tokens: localTokenSlice}
				data, _ := json.MarshalIndent(collection, "", "  ")
				os.WriteFile(localPath, data, 0644)
			}

			// Create mock client
			mockClient := &mockDesignClient{
				tokens: tt.remoteTokens,
			}

			// Create syncer
			config := SyncConfig{
				ProjectID:          "test-project",
				Direction:          SyncDirectionBidirectional,
				LocalPath:          localPath,
				ConflictResolution: ConflictResolutionRemote,
				DryRun:             false,
			}

			syncer := NewSyncer(mockClient, config)

			// Execute sync
			ctx := context.Background()
			result, err := syncer.Sync(ctx)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedAdded, result.Added)
			assert.Equal(t, tt.expectedConflict, len(result.Conflicts))
		})
	}
}

func TestSyncer_DryRun(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	localPath := filepath.Join(tmpDir, "tokens.json")

	// Create local tokens
	localTokens := map[string]Token{
		"color.primary": {Name: "color.primary", Type: "color", Value: "#007bff"},
	}

	localTokenSlice := make([]Token, 0, len(localTokens))
	for _, token := range localTokens {
		localTokenSlice = append(localTokenSlice, token)
	}
	collection := TokenCollection{Tokens: localTokenSlice}
	data, _ := json.MarshalIndent(collection, "", "  ")
	os.WriteFile(localPath, data, 0644)

	// Create mock client with different remote tokens
	mockClient := &mockDesignClient{
		tokens: map[string]Token{
			"color.primary": {Name: "color.primary", Type: "color", Value: "#0056b3"},
		},
	}

	// Create syncer with dry run
	config := SyncConfig{
		ProjectID:          "test-project",
		Direction:          SyncDirectionPull,
		LocalPath:          localPath,
		ConflictResolution: ConflictResolutionRemote,
		DryRun:             true,
	}

	syncer := NewSyncer(mockClient, config)

	// Execute sync
	ctx := context.Background()
	result, err := syncer.Sync(ctx)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.DryRun)

	// Verify local file was not modified
	fileData, err := os.ReadFile(localPath)
	require.NoError(t, err)

	var savedCollection TokenCollection
	json.Unmarshal(fileData, &savedCollection)

	// Local file should still have original value
	assert.Equal(t, "#007bff", savedCollection.Tokens[0].Value)
}

func TestSyncer_ConflictDetection(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	localPath := filepath.Join(tmpDir, "tokens.json")

	// Create conflicting local and remote tokens
	localTokens := map[string]Token{
		"color.primary": {Name: "color.primary", Type: "color", Value: "#007bff"},
	}

	localTokenSlice := make([]Token, 0, len(localTokens))
	for _, token := range localTokens {
		localTokenSlice = append(localTokenSlice, token)
	}
	collection := TokenCollection{Tokens: localTokenSlice}
	data, _ := json.MarshalIndent(collection, "", "  ")
	os.WriteFile(localPath, data, 0644)

	mockClient := &mockDesignClient{
		tokens: map[string]Token{
			"color.primary": {Name: "color.primary", Type: "color", Value: "#ff0000"},
		},
	}

	// Create syncer
	config := SyncConfig{
		ProjectID:          "test-project",
		Direction:          SyncDirectionBidirectional,
		LocalPath:          localPath,
		ConflictResolution: ConflictResolutionRemote,
		DryRun:             false,
	}

	syncer := NewSyncer(mockClient, config)

	// Execute sync
	ctx := context.Background()
	result, err := syncer.Sync(ctx)

	// Assert
	require.NoError(t, err)
	assert.Greater(t, len(result.Conflicts), 0, "Expected conflicts to be detected")

	for _, conflict := range result.Conflicts {
		assert.NotNil(t, conflict.LocalToken)
		assert.NotNil(t, conflict.RemoteToken)
		assert.NotNil(t, conflict.Resolution)
	}
}

func TestToken_ComputeHash(t *testing.T) {
	token1 := Token{
		Name:  "color.primary",
		Type:  "color",
		Value: "#007bff",
	}

	token2 := Token{
		Name:  "color.primary",
		Type:  "color",
		Value: "#007bff",
	}

	token3 := Token{
		Name:  "color.primary",
		Type:  "color",
		Value: "#ff0000",
	}

	// Same tokens should have same hash
	hash1, err := token1.ComputeHash()
	require.NoError(t, err)

	hash2, err := token2.ComputeHash()
	require.NoError(t, err)

	assert.Equal(t, hash1, hash2)

	// Different tokens should have different hash
	hash3, err := token3.ComputeHash()
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash3)
}

func TestToken_Equals(t *testing.T) {
	token1 := &Token{
		Name:  "color.primary",
		Type:  "color",
		Value: "#007bff",
	}

	token2 := &Token{
		Name:  "color.primary",
		Type:  "color",
		Value: "#007bff",
	}

	token3 := &Token{
		Name:  "color.secondary",
		Type:  "color",
		Value: "#007bff",
	}

	assert.True(t, token1.Equals(token2))
	assert.False(t, token1.Equals(token3))
	assert.False(t, token1.Equals(nil))
}

func BenchmarkSyncer_PullTokens(b *testing.B) {
	// Create temporary directory
	tmpDir, _ := os.MkdirTemp("", "sync-bench-*")
	defer os.RemoveAll(tmpDir)

	localPath := filepath.Join(tmpDir, "tokens.json")

	// Create mock client with 1000 tokens
	tokens := make(map[string]Token, 1000)
	for i := 0; i < 1000; i++ {
		name := "token-" + string(rune(i))
		tokens[name] = Token{
			Name:  name,
			Type:  "color",
			Value: "#000000",
		}
	}

	mockClient := &mockDesignClient{
		tokens: tokens,
	}

	config := SyncConfig{
		ProjectID:          "test-project",
		Direction:          SyncDirectionPull,
		LocalPath:          localPath,
		ConflictResolution: ConflictResolutionRemote,
		DryRun:             false,
	}

	syncer := NewSyncer(mockClient, config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		syncer.Sync(ctx)
	}
}

func TestSyncResult_Duration(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	localPath := filepath.Join(tmpDir, "tokens.json")

	mockClient := &mockDesignClient{
		tokens: map[string]Token{
			"color.primary": {Name: "color.primary", Type: "color", Value: "#007bff"},
		},
	}

	config := SyncConfig{
		ProjectID:          "test-project",
		Direction:          SyncDirectionPull,
		LocalPath:          localPath,
		ConflictResolution: ConflictResolutionRemote,
		DryRun:             false,
	}

	syncer := NewSyncer(mockClient, config)

	start := time.Now()
	ctx := context.Background()
	result, err := syncer.Sync(ctx)

	require.NoError(t, err)
	assert.Greater(t, result.Duration, time.Duration(0))
	assert.Less(t, result.Duration, time.Since(start)+time.Millisecond)
}
