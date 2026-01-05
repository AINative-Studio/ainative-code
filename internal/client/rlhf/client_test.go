package rlhf_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/client/rlhf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer() (*httptest.Server, *rlhf.Client) {
	mux := http.NewServeMux()

	// Mock interaction feedback endpoint
	mux.HandleFunc("/api/interactions/feedback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var feedback rlhf.InteractionFeedback
		if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate score
		if feedback.Score < 0.0 || feedback.Score > 1.0 {
			http.Error(w, "invalid score", http.StatusBadRequest)
			return
		}

		resp := rlhf.InteractionFeedbackResponse{
			InteractionID: "interaction-123",
			Status:        "success",
			Message:       "Feedback recorded successfully",
			CreatedAt:     time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Mock batch feedback endpoint
	mux.HandleFunc("/api/interactions/feedback/batch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var batch rlhf.BatchInteractionFeedback
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := rlhf.BatchInteractionFeedbackResponse{
			Successful: []string{"interaction-123", "interaction-124"},
			Failed: map[string]string{
				"interaction-125": "invalid score",
			},
			TotalProcessed: len(batch.Interactions),
			SuccessCount:   len(batch.Interactions) - 1,
			FailureCount:   1,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Mock correction endpoint
	mux.HandleFunc("/api/corrections", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var correction rlhf.Correction
		if err := json.NewDecoder(r.Body).Decode(&correction); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := rlhf.CorrectionResponse{
			CorrectionID:  "correction-123",
			InteractionID: correction.InteractionID,
			Status:        "success",
			Message:       "Correction recorded successfully",
			Diff: &rlhf.DiffResult{
				Original:        "Original response",
				Corrected:       correction.CorrectedResponse,
				SimilarityScore: 0.85,
				Changes: []*rlhf.Change{
					{Type: "modify", Line: 1, Content: "Updated text"},
				},
			},
			CreatedAt: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Mock analytics endpoint
	mux.HandleFunc("/api/analytics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp := rlhf.Analytics{
			ModelID:              "claude-3-5-sonnet-20241022",
			StartDate:            time.Now().AddDate(0, 0, -7),
			EndDate:              time.Now(),
			AverageFeedbackScore: 0.85,
			TotalInteractions:    100,
			TotalCorrections:     10,
			CorrectionRate:       10.0,
			ScoreDistribution: map[string]int{
				"0.0-0.2": 5,
				"0.2-0.4": 10,
				"0.4-0.6": 15,
				"0.6-0.8": 30,
				"0.8-1.0": 40,
			},
			TopCorrectionReasons: []rlhf.ReasonCount{
				{Reason: "Inaccurate information", Count: 5},
				{Reason: "Poor formatting", Count: 3},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Mock get interaction endpoint
	mux.HandleFunc("/api/interactions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp := rlhf.InteractionFeedback{
			Prompt:    "Test prompt",
			Response:  "Test response",
			Score:     0.9,
			ModelID:   "claude-3-5-sonnet-20241022",
			SessionID: "session-123",
			Timestamp: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(mux)

	// Create API client
	apiClient := client.New(
		client.WithBaseURL(server.URL),
	)

	// Create RLHF client
	rlhfClient := rlhf.New(
		rlhf.WithAPIClient(apiClient),
		rlhf.WithBaseURL(server.URL),
	)

	return server, rlhfClient
}

func TestSubmitInteractionFeedback(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()

	ctx := context.Background()

	t.Run("successful submission", func(t *testing.T) {
		feedback := &rlhf.InteractionFeedback{
			Prompt:    "What is 2+2?",
			Response:  "2+2 equals 4",
			Score:     0.95,
			ModelID:   "claude-3-5-sonnet-20241022",
			SessionID: "session-123",
		}

		resp, err := client.SubmitInteractionFeedback(ctx, feedback)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "interaction-123", resp.InteractionID)
		assert.Equal(t, "success", resp.Status)
	})

	t.Run("invalid score - too high", func(t *testing.T) {
		feedback := &rlhf.InteractionFeedback{
			Prompt:   "Test",
			Response: "Test response",
			Score:    1.5,
		}

		_, err := client.SubmitInteractionFeedback(ctx, feedback)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "feedback score must be between 0.0 and 1.0")
	})

	t.Run("invalid score - too low", func(t *testing.T) {
		feedback := &rlhf.InteractionFeedback{
			Prompt:   "Test",
			Response: "Test response",
			Score:    -0.5,
		}

		_, err := client.SubmitInteractionFeedback(ctx, feedback)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "feedback score must be between 0.0 and 1.0")
	})

	t.Run("missing prompt", func(t *testing.T) {
		feedback := &rlhf.InteractionFeedback{
			Response: "Test response",
			Score:    0.8,
		}

		_, err := client.SubmitInteractionFeedback(ctx, feedback)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "prompt is required")
	})

	t.Run("missing response", func(t *testing.T) {
		feedback := &rlhf.InteractionFeedback{
			Prompt: "Test prompt",
			Score:  0.8,
		}

		_, err := client.SubmitInteractionFeedback(ctx, feedback)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "response is required")
	})
}

func TestSubmitBatchInteractionFeedback(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()

	ctx := context.Background()

	t.Run("successful batch submission", func(t *testing.T) {
		batch := &rlhf.BatchInteractionFeedback{
			Interactions: []*rlhf.InteractionFeedback{
				{
					Prompt:   "Question 1",
					Response: "Answer 1",
					Score:    0.9,
				},
				{
					Prompt:   "Question 2",
					Response: "Answer 2",
					Score:    0.8,
				},
				{
					Prompt:   "Question 3",
					Response: "Answer 3",
					Score:    1.0,
				},
			},
		}

		resp, err := client.SubmitBatchInteractionFeedback(ctx, batch)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 3, resp.TotalProcessed)
		assert.Equal(t, 2, resp.SuccessCount)
		assert.Equal(t, 1, resp.FailureCount)
	})

	t.Run("batch with invalid interaction", func(t *testing.T) {
		batch := &rlhf.BatchInteractionFeedback{
			Interactions: []*rlhf.InteractionFeedback{
				{
					Prompt:   "Question 1",
					Response: "Answer 1",
					Score:    1.5, // Invalid score
				},
			},
		}

		_, err := client.SubmitBatchInteractionFeedback(ctx, batch)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "feedback score must be between 0.0 and 1.0")
	})
}

func TestSubmitCorrection(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()

	ctx := context.Background()

	t.Run("successful correction", func(t *testing.T) {
		correction := &rlhf.Correction{
			InteractionID:     "interaction-123",
			CorrectedResponse: "This is the corrected response",
			Reason:            "Inaccurate information",
			Notes:             "The original response contained outdated data",
			Tags:              []string{"accuracy", "factual"},
		}

		resp, err := client.SubmitCorrection(ctx, correction)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "correction-123", resp.CorrectionID)
		assert.Equal(t, "interaction-123", resp.InteractionID)
		assert.Equal(t, "success", resp.Status)
		assert.NotNil(t, resp.Diff)
		assert.Equal(t, 0.85, resp.Diff.SimilarityScore)
	})

	t.Run("missing interaction ID", func(t *testing.T) {
		correction := &rlhf.Correction{
			CorrectedResponse: "Corrected text",
		}

		_, err := client.SubmitCorrection(ctx, correction)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "interaction_id is required")
	})

	t.Run("missing corrected response", func(t *testing.T) {
		correction := &rlhf.Correction{
			InteractionID: "interaction-123",
		}

		_, err := client.SubmitCorrection(ctx, correction)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "corrected_response is required")
	})
}

func TestGetAnalytics(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()

	ctx := context.Background()

	t.Run("successful analytics retrieval", func(t *testing.T) {
		req := &rlhf.AnalyticsRequest{
			ModelID:     "claude-3-5-sonnet-20241022",
			StartDate:   time.Now().AddDate(0, 0, -7),
			EndDate:     time.Now(),
			Granularity: "day",
		}

		analytics, err := client.GetAnalytics(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, analytics)
		assert.Equal(t, "claude-3-5-sonnet-20241022", analytics.ModelID)
		assert.Equal(t, 0.85, analytics.AverageFeedbackScore)
		assert.Equal(t, 100, analytics.TotalInteractions)
		assert.Equal(t, 10, analytics.TotalCorrections)
		assert.Equal(t, 10.0, analytics.CorrectionRate)
		assert.Len(t, analytics.ScoreDistribution, 5)
		assert.Len(t, analytics.TopCorrectionReasons, 2)
	})

	t.Run("invalid date range", func(t *testing.T) {
		req := &rlhf.AnalyticsRequest{
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(0, 0, -7), // End before start
		}

		_, err := client.GetAnalytics(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "end_date must be after start_date")
	})
}

func TestGetInteraction(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()

	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		interaction, err := client.GetInteraction(ctx, "interaction-123")
		require.NoError(t, err)
		assert.NotNil(t, interaction)
		assert.Equal(t, "Test prompt", interaction.Prompt)
		assert.Equal(t, "Test response", interaction.Response)
		assert.Equal(t, 0.9, interaction.Score)
	})

	t.Run("missing interaction ID", func(t *testing.T) {
		_, err := client.GetInteraction(ctx, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "interaction_id is required")
	})
}

func TestGetCorrection(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()

	ctx := context.Background()

	t.Run("missing correction ID", func(t *testing.T) {
		_, err := client.GetCorrection(ctx, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "correction_id is required")
	})
}
