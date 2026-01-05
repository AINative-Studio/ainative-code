package design

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AINative-studio/ainative-code/internal/logger"
)

// Syncer handles bidirectional design token synchronization.
type Syncer struct {
	client   DesignClient
	resolver *ConflictResolver
	config   SyncConfig
}

// DesignClient defines the interface for interacting with the design API.
type DesignClient interface {
	// GetTokens retrieves all design tokens for a project.
	GetTokens(ctx context.Context, projectID string) ([]Token, error)

	// UploadTokens uploads design tokens to the remote project.
	UploadTokens(ctx context.Context, projectID string, tokens []Token) error

	// DeleteToken deletes a design token from the remote project.
	DeleteToken(ctx context.Context, projectID string, tokenName string) error
}

// NewSyncer creates a new design token syncer.
func NewSyncer(client DesignClient, config SyncConfig) *Syncer {
	return &Syncer{
		client: client,
		resolver: NewConflictResolver(ConflictResolverConfig{
			Strategy: config.ConflictResolution,
		}),
		config: config,
	}
}

// Sync performs a synchronization operation based on the configured direction.
func (s *Syncer) Sync(ctx context.Context) (*SyncResult, error) {
	startTime := time.Now()

	result := &SyncResult{
		Conflicts: make([]Conflict, 0),
		Errors:    make([]error, 0),
		DryRun:    s.config.DryRun,
	}

	logger.InfoEvent().
		Str("project_id", s.config.ProjectID).
		Str("direction", string(s.config.Direction)).
		Bool("dry_run", s.config.DryRun).
		Msg("Starting design token sync")

	// Load local tokens
	localTokens, err := s.loadLocalTokens()
	if err != nil {
		return nil, fmt.Errorf("failed to load local tokens: %w", err)
	}

	// Load remote tokens
	remoteTokens, err := s.loadRemoteTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load remote tokens: %w", err)
	}

	// Perform sync based on direction
	switch s.config.Direction {
	case SyncDirectionPull:
		result, err = s.pullTokens(ctx, localTokens, remoteTokens)
	case SyncDirectionPush:
		result, err = s.pushTokens(ctx, localTokens, remoteTokens)
	case SyncDirectionBidirectional:
		result, err = s.bidirectionalSync(ctx, localTokens, remoteTokens)
	default:
		return nil, fmt.Errorf("unsupported sync direction: %s", s.config.Direction)
	}

	if err != nil {
		return nil, err
	}

	result.Duration = time.Since(startTime)

	logger.InfoEvent().
		Int("added", result.Added).
		Int("updated", result.Updated).
		Int("deleted", result.Deleted).
		Int("conflicts", len(result.Conflicts)).
		Dur("duration", result.Duration).
		Msg("Design token sync completed")

	return result, nil
}

// pullTokens pulls tokens from remote to local.
func (s *Syncer) pullTokens(ctx context.Context, localTokens, remoteTokens map[string]*TokenWithMetadata) (*SyncResult, error) {
	result := &SyncResult{
		Conflicts: make([]Conflict, 0),
		Errors:    make([]error, 0),
		DryRun:    s.config.DryRun,
	}

	tokensToWrite := make(map[string]*Token)

	// Process all remote tokens
	for name, remoteToken := range remoteTokens {
		localToken, existsLocally := localTokens[name]

		if !existsLocally {
			// New token from remote
			tokensToWrite[name] = &remoteToken.Token
			result.Added++
			logger.DebugEvent().Str("token", name).Msg("Adding new token from remote")
		} else if !localToken.Equals(&remoteToken.Token) {
			// Token exists but has different value
			conflict := s.detectConflict(localToken, remoteToken)
			if conflict != nil {
				// Resolve conflict (in pull mode, prefer remote)
				resolution := s.resolver.Resolve(*conflict)
				if resolution.SelectedToken != nil {
					tokensToWrite[name] = resolution.SelectedToken
					result.Updated++
					conflict.Resolution = &resolution
					result.Conflicts = append(result.Conflicts, *conflict)
				}
			} else {
				tokensToWrite[name] = &remoteToken.Token
				result.Updated++
			}
			logger.DebugEvent().Str("token", name).Msg("Updating token from remote")
		}
	}

	// Check for local tokens that were deleted remotely
	for name := range localTokens {
		if _, existsRemotely := remoteTokens[name]; !existsRemotely {
			result.Deleted++
			logger.DebugEvent().Str("token", name).Msg("Deleting local token (deleted remotely)")
		}
	}

	// Write tokens to local file if not dry run
	if !s.config.DryRun && len(tokensToWrite) > 0 {
		if err := s.saveLocalTokens(tokensToWrite); err != nil {
			return nil, fmt.Errorf("failed to save local tokens: %w", err)
		}
	}

	return result, nil
}

// pushTokens pushes tokens from local to remote.
func (s *Syncer) pushTokens(ctx context.Context, localTokens, remoteTokens map[string]*TokenWithMetadata) (*SyncResult, error) {
	result := &SyncResult{
		Conflicts: make([]Conflict, 0),
		Errors:    make([]error, 0),
		DryRun:    s.config.DryRun,
	}

	tokensToUpload := make([]Token, 0)
	tokensToDelete := make([]string, 0)

	// Process all local tokens
	for name, localToken := range localTokens {
		remoteToken, existsRemotely := remoteTokens[name]

		if !existsRemotely {
			// New token to upload
			tokensToUpload = append(tokensToUpload, localToken.Token)
			result.Added++
			logger.DebugEvent().Str("token", name).Msg("Uploading new token to remote")
		} else if !localToken.Equals(&remoteToken.Token) {
			// Token exists but has different value
			conflict := s.detectConflict(localToken, remoteToken)
			if conflict != nil {
				// Resolve conflict (in push mode, prefer local)
				resolution := s.resolver.Resolve(*conflict)
				if resolution.SelectedToken != nil {
					tokensToUpload = append(tokensToUpload, *resolution.SelectedToken)
					result.Updated++
					conflict.Resolution = &resolution
					result.Conflicts = append(result.Conflicts, *conflict)
				}
			} else {
				tokensToUpload = append(tokensToUpload, localToken.Token)
				result.Updated++
			}
			logger.DebugEvent().Str("token", name).Msg("Updating remote token")
		}
	}

	// Check for remote tokens that were deleted locally
	for name := range remoteTokens {
		if _, existsLocally := localTokens[name]; !existsLocally {
			tokensToDelete = append(tokensToDelete, name)
			result.Deleted++
			logger.DebugEvent().Str("token", name).Msg("Deleting remote token (deleted locally)")
		}
	}

	// Upload and delete tokens if not dry run
	if !s.config.DryRun {
		if len(tokensToUpload) > 0 {
			if err := s.client.UploadTokens(ctx, s.config.ProjectID, tokensToUpload); err != nil {
				return nil, fmt.Errorf("failed to upload tokens: %w", err)
			}
		}

		for _, tokenName := range tokensToDelete {
			if err := s.client.DeleteToken(ctx, s.config.ProjectID, tokenName); err != nil {
				logger.WarnEvent().Err(err).Str("token", tokenName).Msg("Failed to delete remote token")
				result.Errors = append(result.Errors, err)
			}
		}
	}

	return result, nil
}

// bidirectionalSync performs a bidirectional synchronization.
func (s *Syncer) bidirectionalSync(ctx context.Context, localTokens, remoteTokens map[string]*TokenWithMetadata) (*SyncResult, error) {
	result := &SyncResult{
		Conflicts: make([]Conflict, 0),
		Errors:    make([]error, 0),
		DryRun:    s.config.DryRun,
	}

	tokensToWrite := make(map[string]*Token)
	tokensToUpload := make([]Token, 0)
	tokensToDelete := make([]string, 0)

	// Create a union of all token names
	allTokenNames := make(map[string]bool)
	for name := range localTokens {
		allTokenNames[name] = true
	}
	for name := range remoteTokens {
		allTokenNames[name] = true
	}

	// Process each token
	for name := range allTokenNames {
		localToken, existsLocally := localTokens[name]
		remoteToken, existsRemotely := remoteTokens[name]

		if existsLocally && !existsRemotely {
			// Token only exists locally - upload it
			tokensToUpload = append(tokensToUpload, localToken.Token)
			result.Added++
			logger.DebugEvent().Str("token", name).Msg("Uploading new local token")
		} else if !existsLocally && existsRemotely {
			// Token only exists remotely - download it
			tokensToWrite[name] = &remoteToken.Token
			result.Added++
			logger.DebugEvent().Str("token", name).Msg("Downloading new remote token")
		} else if existsLocally && existsRemotely {
			// Token exists in both - check for conflicts
			if !localToken.Equals(&remoteToken.Token) {
				conflict := s.detectConflict(localToken, remoteToken)
				if conflict != nil {
					// Resolve conflict
					resolution := s.resolver.Resolve(*conflict)
					if resolution.SelectedToken != nil {
						// Update both local and remote with resolved token
						tokensToWrite[name] = resolution.SelectedToken
						tokensToUpload = append(tokensToUpload, *resolution.SelectedToken)
						result.Updated++
						conflict.Resolution = &resolution
						result.Conflicts = append(result.Conflicts, *conflict)
					}
				}
			}
		}
	}

	// Apply changes if not dry run
	if !s.config.DryRun {
		// Write local changes
		if len(tokensToWrite) > 0 {
			if err := s.saveLocalTokens(tokensToWrite); err != nil {
				return nil, fmt.Errorf("failed to save local tokens: %w", err)
			}
		}

		// Upload remote changes
		if len(tokensToUpload) > 0 {
			if err := s.client.UploadTokens(ctx, s.config.ProjectID, tokensToUpload); err != nil {
				return nil, fmt.Errorf("failed to upload tokens: %w", err)
			}
		}

		// Delete remote tokens
		for _, tokenName := range tokensToDelete {
			if err := s.client.DeleteToken(ctx, s.config.ProjectID, tokenName); err != nil {
				logger.WarnEvent().Err(err).Str("token", tokenName).Msg("Failed to delete remote token")
				result.Errors = append(result.Errors, err)
			}
		}
	}

	return result, nil
}

// detectConflict detects if there's a conflict between local and remote tokens.
func (s *Syncer) detectConflict(localToken, remoteToken *TokenWithMetadata) *Conflict {
	// If tokens are equal, no conflict
	if localToken.Equals(&remoteToken.Token) {
		return nil
	}

	// Determine conflict type
	conflictType := ConflictTypeBothModified
	if localToken.Type != remoteToken.Type {
		conflictType = ConflictTypeTypeChange
	}

	return &Conflict{
		TokenName:    localToken.Name,
		LocalToken:   &localToken.Token,
		RemoteToken:  &remoteToken.Token,
		ConflictType: conflictType,
	}
}

// loadLocalTokens loads design tokens from the local file.
func (s *Syncer) loadLocalTokens() (map[string]*TokenWithMetadata, error) {
	if s.config.LocalPath == "" {
		return make(map[string]*TokenWithMetadata), nil
	}

	// Check if file exists
	if _, err := os.Stat(s.config.LocalPath); os.IsNotExist(err) {
		logger.DebugEvent().Str("path", s.config.LocalPath).Msg("Local token file does not exist, starting fresh")
		return make(map[string]*TokenWithMetadata), nil
	}

	// Read file
	data, err := os.ReadFile(s.config.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read local token file: %w", err)
	}

	// Parse JSON
	var collection TokenCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("failed to parse local token file: %w", err)
	}

	// Convert to map with metadata
	tokens := make(map[string]*TokenWithMetadata)
	for i := range collection.Tokens {
		token := &collection.Tokens[i]
		hash, _ := token.ComputeHash()
		tokens[token.Name] = &TokenWithMetadata{
			Token:        *token,
			LastSyncedAt: time.Now(),
			Hash:         hash,
			Version:      1,
		}
	}

	logger.DebugEvent().Int("count", len(tokens)).Msg("Loaded local tokens")
	return tokens, nil
}

// loadRemoteTokens loads design tokens from the remote API.
func (s *Syncer) loadRemoteTokens(ctx context.Context) (map[string]*TokenWithMetadata, error) {
	tokens, err := s.client.GetTokens(ctx, s.config.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote tokens: %w", err)
	}

	// Convert to map with metadata
	tokenMap := make(map[string]*TokenWithMetadata)
	for i := range tokens {
		token := &tokens[i]
		hash, _ := token.ComputeHash()
		tokenMap[token.Name] = &TokenWithMetadata{
			Token:        *token,
			LastSyncedAt: time.Now(),
			Hash:         hash,
			Version:      1,
		}
	}

	logger.DebugEvent().Int("count", len(tokenMap)).Msg("Loaded remote tokens")
	return tokenMap, nil
}

// saveLocalTokens saves design tokens to the local file.
func (s *Syncer) saveLocalTokens(tokens map[string]*Token) error {
	// Ensure directory exists
	dir := filepath.Dir(s.config.LocalPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Convert map to slice
	tokenSlice := make([]Token, 0, len(tokens))
	for _, token := range tokens {
		tokenSlice = append(tokenSlice, *token)
	}

	// Create collection
	collection := TokenCollection{
		Tokens: tokenSlice,
		Metadata: map[string]string{
			"synced_at": time.Now().Format(time.RFC3339),
			"project":   s.config.ProjectID,
		},
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	// Write to file
	if err := os.WriteFile(s.config.LocalPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	logger.DebugEvent().Str("path", s.config.LocalPath).Int("count", len(tokens)).Msg("Saved local tokens")
	return nil
}
