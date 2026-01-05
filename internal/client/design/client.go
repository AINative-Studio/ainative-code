package design

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/design"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

// Client represents a client for AINative Design API operations.
type Client struct {
	apiClient *client.Client
	projectID string
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithAPIClient sets the underlying HTTP API client.
func WithAPIClient(apiClient *client.Client) Option {
	return func(c *Client) {
		c.apiClient = apiClient
	}
}

// WithProjectID sets the project ID for design operations.
func WithProjectID(projectID string) Option {
	return func(c *Client) {
		c.projectID = projectID
	}
}

// New creates a new Design client with the specified options.
func New(opts ...Option) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// UploadTokensRequest represents a request to upload design tokens.
type UploadTokensRequest struct {
	ProjectID          string                                   `json:"project_id"`
	Tokens             []*design.Token                          `json:"tokens"`
	ConflictResolution design.ConflictResolutionStrategyUpload `json:"conflict_resolution"`
}

// UploadTokensResponse represents the response from uploading tokens.
type UploadTokensResponse struct {
	Success       bool   `json:"success"`
	UploadedCount int    `json:"uploaded_count"`
	SkippedCount  int    `json:"skipped_count"`
	UpdatedCount  int    `json:"updated_count"`
	Message       string `json:"message,omitempty"`
}

// TokenQueryRequest represents a request to query tokens.
type TokenQueryRequest struct {
	ProjectID string   `json:"project_id"`
	Types     []string `json:"types,omitempty"`
	Category  string   `json:"category,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
}

// TokenQueryResponse represents the response from querying tokens.
type TokenQueryResponse struct {
	Tokens []*design.Token `json:"tokens"`
	Total  int             `json:"total"`
}

// DeleteTokenRequest represents a request to delete a token.
type DeleteTokenRequest struct {
	ProjectID string `json:"project_id"`
	TokenName string `json:"token_name"`
}

// DeleteTokenResponse represents the response from deleting a token.
type DeleteTokenResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ProgressCallback is called during upload to report progress.
type ProgressCallback func(uploaded, total int)

// UploadTokens uploads design tokens to the AINative Design system.
func (c *Client) UploadTokens(ctx context.Context, tokens []*design.Token, resolution design.ConflictResolutionStrategyUpload, callback ProgressCallback) (*UploadTokensResponse, error) {
	logger.InfoEvent().
		Int("token_count", len(tokens)).
		Str("conflict_resolution", string(resolution)).
		Msg("Uploading design tokens")

	if c.projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens to upload")
	}

	// For large token sets, upload in batches
	batchSize := 100
	totalUploaded := 0
	totalSkipped := 0
	totalUpdated := 0

	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}

		batch := tokens[i:end]

		req := &UploadTokensRequest{
			ProjectID:          c.projectID,
			Tokens:             batch,
			ConflictResolution: resolution,
		}

		path := "/api/v1/design/tokens/upload"
		respData, err := c.apiClient.Post(ctx, path, req)
		if err != nil {
			return nil, fmt.Errorf("failed to upload token batch: %w", err)
		}

		var resp UploadTokensResponse
		if err := json.Unmarshal(respData, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		totalUploaded += resp.UploadedCount
		totalSkipped += resp.SkippedCount
		totalUpdated += resp.UpdatedCount

		// Report progress
		if callback != nil {
			callback(end, len(tokens))
		}

		logger.DebugEvent().
			Int("batch_uploaded", resp.UploadedCount).
			Int("batch_skipped", resp.SkippedCount).
			Int("batch_updated", resp.UpdatedCount).
			Int("batch_number", (i/batchSize)+1).
			Msg("Batch uploaded")
	}

	logger.InfoEvent().
		Int("total_uploaded", totalUploaded).
		Int("total_skipped", totalSkipped).
		Int("total_updated", totalUpdated).
		Msg("Token upload completed")

	return &UploadTokensResponse{
		Success:       true,
		UploadedCount: totalUploaded,
		SkippedCount:  totalSkipped,
		UpdatedCount:  totalUpdated,
		Message:       fmt.Sprintf("Successfully processed %d tokens", len(tokens)),
	}, nil
}

// GetTokens retrieves design tokens from the AINative Design system.
func (c *Client) GetTokens(ctx context.Context, types []string, category string, limit, offset int) ([]*design.Token, int, error) {
	logger.DebugEvent().
		Strs("types", types).
		Str("category", category).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Retrieving design tokens")

	if c.projectID == "" {
		return nil, 0, fmt.Errorf("project ID is required")
	}

	if limit == 0 {
		limit = 100
	}

	req := &TokenQueryRequest{
		ProjectID: c.projectID,
		Types:     types,
		Category:  category,
		Limit:     limit,
		Offset:    offset,
	}

	path := "/api/v1/design/tokens/query"
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query tokens: %w", err)
	}

	var resp TokenQueryResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, 0, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Int("count", len(resp.Tokens)).
		Int("total", resp.Total).
		Msg("Tokens retrieved successfully")

	return resp.Tokens, resp.Total, nil
}

// DeleteToken deletes a design token from the AINative Design system.
func (c *Client) DeleteToken(ctx context.Context, tokenName string) error {
	logger.InfoEvent().
		Str("token_name", tokenName).
		Msg("Deleting design token")

	if c.projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	if tokenName == "" {
		return fmt.Errorf("token name is required")
	}

	req := &DeleteTokenRequest{
		ProjectID: c.projectID,
		TokenName: tokenName,
	}

	path := "/api/v1/design/tokens/delete"
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	var resp DeleteTokenResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("delete operation failed: %s", resp.Message)
	}

	logger.InfoEvent().
		Str("token_name", tokenName).
		Msg("Token deleted successfully")

	return nil
}

// ValidateTokens validates a batch of design tokens.
func (c *Client) ValidateTokens(tokens []*design.Token) *design.ValidationResult {
	validator := design.NewValidator()
	return validator.ValidateBatch(tokens)
}
