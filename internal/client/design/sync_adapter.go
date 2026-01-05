package design

import (
	"context"
	"fmt"

	"github.com/AINative-studio/ainative-code/internal/design"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

// SyncAdapter adapts the Design API client to the sync.DesignClient interface.
// This adapter bridges the gap between the HTTP client API and the sync engine's expectations.
type SyncAdapter struct {
	client    *Client
	projectID string
}

// NewSyncAdapter creates a new sync adapter wrapping the design client.
func NewSyncAdapter(client *Client, projectID string) *SyncAdapter {
	return &SyncAdapter{
		client:    client,
		projectID: projectID,
	}
}

// GetTokens retrieves all design tokens for the project.
// This method implements the sync.DesignClient interface.
func (a *SyncAdapter) GetTokens(ctx context.Context, projectID string) ([]design.Token, error) {
	logger.DebugEvent().
		Str("project_id", projectID).
		Msg("Fetching tokens via sync adapter")

	// Fetch all tokens in batches
	var allTokens []*design.Token
	offset := 0
	limit := 100

	for {
		tokens, total, err := a.client.GetTokens(ctx, nil, "", limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tokens: %w", err)
		}

		allTokens = append(allTokens, tokens...)

		// Check if we've fetched all tokens
		if len(allTokens) >= total {
			break
		}

		offset += limit
	}

	logger.DebugEvent().
		Int("count", len(allTokens)).
		Msg("Fetched all tokens")

	// Convert from pointer slice to value slice
	result := make([]design.Token, len(allTokens))
	for i, token := range allTokens {
		result[i] = *token
	}

	return result, nil
}

// UploadTokens uploads design tokens to the remote project.
// This method implements the sync.DesignClient interface.
func (a *SyncAdapter) UploadTokens(ctx context.Context, projectID string, tokens []design.Token) error {
	logger.InfoEvent().
		Str("project_id", projectID).
		Int("count", len(tokens)).
		Msg("Uploading tokens via sync adapter")

	if len(tokens) == 0 {
		return nil
	}

	// Convert from value slice to pointer slice
	tokenPtrs := make([]*design.Token, len(tokens))
	for i := range tokens {
		tokenPtrs[i] = &tokens[i]
	}

	// Use overwrite strategy for sync operations
	_, err := a.client.UploadTokens(ctx, tokenPtrs, design.ConflictOverwrite, nil)
	if err != nil {
		return fmt.Errorf("failed to upload tokens: %w", err)
	}

	logger.InfoEvent().
		Int("count", len(tokens)).
		Msg("Successfully uploaded tokens")

	return nil
}

// DeleteToken deletes a design token from the remote project.
// This method implements the sync.DesignClient interface.
func (a *SyncAdapter) DeleteToken(ctx context.Context, projectID string, tokenName string) error {
	logger.InfoEvent().
		Str("project_id", projectID).
		Str("token_name", tokenName).
		Msg("Deleting token via sync adapter")

	if err := a.client.DeleteToken(ctx, tokenName); err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	logger.InfoEvent().
		Str("token_name", tokenName).
		Msg("Successfully deleted token")

	return nil
}

// Ensure SyncAdapter implements the sync.DesignClient interface
var _ design.DesignClient = (*SyncAdapter)(nil)
