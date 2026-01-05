package rlhf

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

// Client represents a client for RLHF operations.
type Client struct {
	apiClient *client.Client
	baseURL   string
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithAPIClient sets the underlying HTTP API client.
func WithAPIClient(apiClient *client.Client) Option {
	return func(c *Client) {
		c.apiClient = apiClient
	}
}

// WithBaseURL sets the RLHF base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// New creates a new RLHF client with the specified options.
func New(opts ...Option) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SubmitInteractionFeedback submits feedback for a single AI interaction.
func (c *Client) SubmitInteractionFeedback(ctx context.Context, feedback *InteractionFeedback) (*InteractionFeedbackResponse, error) {
	logger.InfoEvent().
		Float64("score", feedback.Score).
		Str("model_id", feedback.ModelID).
		Str("session_id", feedback.SessionID).
		Msg("Submitting interaction feedback")

	// Validate feedback score
	if feedback.Score < 0.0 || feedback.Score > 1.0 {
		return nil, fmt.Errorf("feedback score must be between 0.0 and 1.0, got: %.2f", feedback.Score)
	}

	// Validate required fields
	if feedback.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	if feedback.Response == "" {
		return nil, fmt.Errorf("response is required")
	}

	// Set timestamp if not provided
	if feedback.Timestamp.IsZero() {
		feedback.Timestamp = time.Now()
	}

	path := "/api/interactions/feedback"
	respData, err := c.apiClient.Post(ctx, path, feedback)
	if err != nil {
		return nil, fmt.Errorf("failed to submit interaction feedback: %w", err)
	}

	var resp InteractionFeedbackResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.InfoEvent().
		Str("interaction_id", resp.InteractionID).
		Str("status", resp.Status).
		Msg("Interaction feedback submitted successfully")

	return &resp, nil
}

// SubmitBatchInteractionFeedback submits feedback for multiple interactions.
func (c *Client) SubmitBatchInteractionFeedback(ctx context.Context, batch *BatchInteractionFeedback) (*BatchInteractionFeedbackResponse, error) {
	logger.InfoEvent().
		Int("count", len(batch.Interactions)).
		Msg("Submitting batch interaction feedback")

	// Validate all interactions
	for i, feedback := range batch.Interactions {
		if feedback.Score < 0.0 || feedback.Score > 1.0 {
			return nil, fmt.Errorf("interaction %d: feedback score must be between 0.0 and 1.0, got: %.2f", i, feedback.Score)
		}
		if feedback.Prompt == "" {
			return nil, fmt.Errorf("interaction %d: prompt is required", i)
		}
		if feedback.Response == "" {
			return nil, fmt.Errorf("interaction %d: response is required", i)
		}
		if feedback.Timestamp.IsZero() {
			feedback.Timestamp = time.Now()
		}
	}

	path := "/api/interactions/feedback/batch"
	respData, err := c.apiClient.Post(ctx, path, batch)
	if err != nil {
		return nil, fmt.Errorf("failed to submit batch feedback: %w", err)
	}

	var resp BatchInteractionFeedbackResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.InfoEvent().
		Int("total_processed", resp.TotalProcessed).
		Int("success_count", resp.SuccessCount).
		Int("failure_count", resp.FailureCount).
		Msg("Batch feedback submitted")

	return &resp, nil
}

// SubmitCorrection submits a correction for an AI response.
func (c *Client) SubmitCorrection(ctx context.Context, correction *Correction) (*CorrectionResponse, error) {
	logger.InfoEvent().
		Str("interaction_id", correction.InteractionID).
		Str("reason", correction.Reason).
		Msg("Submitting correction")

	// Validate required fields
	if correction.InteractionID == "" {
		return nil, fmt.Errorf("interaction_id is required")
	}
	if correction.CorrectedResponse == "" {
		return nil, fmt.Errorf("corrected_response is required")
	}

	path := "/api/corrections"
	respData, err := c.apiClient.Post(ctx, path, correction)
	if err != nil {
		return nil, fmt.Errorf("failed to submit correction: %w", err)
	}

	var resp CorrectionResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.InfoEvent().
		Str("correction_id", resp.CorrectionID).
		Str("status", resp.Status).
		Msg("Correction submitted successfully")

	return &resp, nil
}

// GetAnalytics retrieves RLHF analytics for the specified parameters.
func (c *Client) GetAnalytics(ctx context.Context, req *AnalyticsRequest) (*Analytics, error) {
	logger.DebugEvent().
		Str("model_id", req.ModelID).
		Time("start_date", req.StartDate).
		Time("end_date", req.EndDate).
		Msg("Fetching RLHF analytics")

	// Validate date range
	if req.EndDate.Before(req.StartDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	// Build query parameters
	params := url.Values{}
	if req.ModelID != "" {
		params.Add("model_id", req.ModelID)
	}
	params.Add("start_date", req.StartDate.Format(time.RFC3339))
	params.Add("end_date", req.EndDate.Format(time.RFC3339))
	if req.Granularity != "" {
		params.Add("granularity", req.Granularity)
	}

	path := "/api/analytics?" + params.Encode()
	respData, err := c.apiClient.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	var analytics Analytics
	if err := json.Unmarshal(respData, &analytics); err != nil {
		return nil, fmt.Errorf("failed to parse analytics response: %w", err)
	}

	logger.InfoEvent().
		Float64("avg_score", analytics.AverageFeedbackScore).
		Int("total_interactions", analytics.TotalInteractions).
		Int("total_corrections", analytics.TotalCorrections).
		Msg("Analytics retrieved successfully")

	return &analytics, nil
}

// ExportAnalytics exports analytics data in the specified format.
func (c *Client) ExportAnalytics(ctx context.Context, req *AnalyticsExportRequest) ([]byte, error) {
	logger.InfoEvent().
		Str("format", string(req.Format)).
		Msg("Exporting analytics")

	// Validate format
	switch req.Format {
	case ExportFormatCSV, ExportFormatJSON, ExportFormatExcel:
		// Valid format
	default:
		return nil, fmt.Errorf("invalid export format: %s", req.Format)
	}

	path := "/api/analytics/export"
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to export analytics: %w", err)
	}

	logger.InfoEvent().
		Int("size_bytes", len(respData)).
		Msg("Analytics exported successfully")

	return respData, nil
}

// GetInteraction retrieves a specific interaction by ID.
func (c *Client) GetInteraction(ctx context.Context, interactionID string) (*InteractionFeedback, error) {
	logger.DebugEvent().
		Str("interaction_id", interactionID).
		Msg("Fetching interaction")

	if interactionID == "" {
		return nil, fmt.Errorf("interaction_id is required")
	}

	path := fmt.Sprintf("/api/interactions/%s", interactionID)
	respData, err := c.apiClient.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get interaction: %w", err)
	}

	var interaction InteractionFeedback
	if err := json.Unmarshal(respData, &interaction); err != nil {
		return nil, fmt.Errorf("failed to parse interaction response: %w", err)
	}

	return &interaction, nil
}

// ListInteractions lists interactions with optional filtering.
func (c *Client) ListInteractions(ctx context.Context, modelID, sessionID string, limit, offset int) ([]*InteractionFeedback, error) {
	logger.DebugEvent().
		Str("model_id", modelID).
		Str("session_id", sessionID).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Listing interactions")

	// Build query parameters
	params := url.Values{}
	if modelID != "" {
		params.Add("model_id", modelID)
	}
	if sessionID != "" {
		params.Add("session_id", sessionID)
	}
	if limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		params.Add("offset", fmt.Sprintf("%d", offset))
	}

	path := "/api/interactions"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respData, err := c.apiClient.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list interactions: %w", err)
	}

	var interactions []*InteractionFeedback
	if err := json.Unmarshal(respData, &interactions); err != nil {
		return nil, fmt.Errorf("failed to parse interactions response: %w", err)
	}

	logger.InfoEvent().
		Int("count", len(interactions)).
		Msg("Interactions listed successfully")

	return interactions, nil
}

// GetCorrection retrieves a specific correction by ID.
func (c *Client) GetCorrection(ctx context.Context, correctionID string) (*CorrectionResponse, error) {
	logger.DebugEvent().
		Str("correction_id", correctionID).
		Msg("Fetching correction")

	if correctionID == "" {
		return nil, fmt.Errorf("correction_id is required")
	}

	path := fmt.Sprintf("/api/corrections/%s", correctionID)
	respData, err := c.apiClient.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get correction: %w", err)
	}

	var correction CorrectionResponse
	if err := json.Unmarshal(respData, &correction); err != nil {
		return nil, fmt.Errorf("failed to parse correction response: %w", err)
	}

	return &correction, nil
}
