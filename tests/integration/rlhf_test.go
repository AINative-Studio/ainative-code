// +build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/client/rlhf"
	"github.com/stretchr/testify/suite"
)

// RLHFIntegrationTestSuite tests RLHF feedback functionality.
type RLHFIntegrationTestSuite struct {
	suite.Suite
	rlhfClient *rlhf.Client
	mockServer *httptest.Server
	cleanup    func()
}

// SetupTest runs before each test in the suite.
func (s *RLHFIntegrationTestSuite) SetupTest() {
	// Create mock RLHF server
	mux := http.NewServeMux()

	// Interaction feedback endpoint
	mux.HandleFunc("/api/interactions/feedback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var feedback rlhf.InteractionFeedback
		if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
			return
		}

		// Validate feedback
		if feedback.Score < 0.0 || feedback.Score > 1.0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_score"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&rlhf.InteractionFeedbackResponse{
			InteractionID: "interaction_12345",
			Status:        "success",
			Message:       "Feedback recorded successfully",
			CreatedAt:     time.Now(),
		})
	})

	// Batch feedback endpoint
	mux.HandleFunc("/api/interactions/feedback/batch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var batch rlhf.BatchInteractionFeedback
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		successful := []string{}
		failed := map[string]string{}

		for i, interaction := range batch.Interactions {
			if interaction.Score >= 0.0 && interaction.Score <= 1.0 && interaction.Prompt != "" && interaction.Response != "" {
				successful = append(successful, string(rune('A'+i)))
			} else {
				failed[string(rune('A'+i))] = "invalid data"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&rlhf.BatchInteractionFeedbackResponse{
			Successful:     successful,
			Failed:         failed,
			TotalProcessed: len(batch.Interactions),
			SuccessCount:   len(successful),
			FailureCount:   len(failed),
		})
	})

	// Corrections endpoint
	mux.HandleFunc("/api/corrections", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var correction rlhf.Correction
		if err := json.NewDecoder(r.Body).Decode(&correction); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&rlhf.CorrectionResponse{
			CorrectionID:  "correction_67890",
			InteractionID: correction.InteractionID,
			Status:        "success",
			Message:       "Correction recorded successfully",
			Diff: &rlhf.DiffResult{
				Original:        "Original response text",
				Corrected:       correction.CorrectedResponse,
				SimilarityScore: 0.85,
				Changes: []*rlhf.Change{
					{Type: "modify", Line: 1, Content: "Changed text"},
				},
			},
			CreatedAt: time.Now(),
		})
	})

	// Analytics endpoint
	mux.HandleFunc("/api/analytics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&rlhf.Analytics{
			ModelID:              "claude-3-opus-20240229",
			StartDate:            time.Now().AddDate(0, 0, -30),
			EndDate:              time.Now(),
			AverageFeedbackScore: 0.87,
			TotalInteractions:    1500,
			TotalCorrections:     45,
			CorrectionRate:       3.0,
			ScoreDistribution: map[string]int{
				"0.0-0.2": 10,
				"0.2-0.4": 25,
				"0.4-0.6": 150,
				"0.6-0.8": 500,
				"0.8-1.0": 815,
			},
			TopCorrectionReasons: []rlhf.ReasonCount{
				{Reason: "Inaccurate information", Count: 15},
				{Reason: "Incomplete response", Count: 12},
				{Reason: "Tone adjustment", Count: 8},
			},
		})
	})

	// Export endpoint
	mux.HandleFunc("/api/analytics/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var exportReq rlhf.AnalyticsExportRequest
		if err := json.NewDecoder(r.Body).Decode(&exportReq); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Return mock export data
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"exported": true, "format": "` + string(exportReq.Format) + `"}`))
	})

	// Get interaction endpoint
	mux.HandleFunc("/api/interactions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&rlhf.InteractionFeedback{
			Prompt:    "Test prompt",
			Response:  "Test response",
			Score:     0.9,
			SessionID: "session_123",
			ModelID:   "claude-3-opus-20240229",
			Timestamp: time.Now(),
		})
	})

	// List interactions endpoint
	mux.HandleFunc("/api/interactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		interactions := []*rlhf.InteractionFeedback{
			{
				Prompt:    "Test prompt 1",
				Response:  "Test response 1",
				Score:     0.9,
				SessionID: "session_123",
				ModelID:   "claude-3-opus-20240229",
				Timestamp: time.Now(),
			},
			{
				Prompt:    "Test prompt 2",
				Response:  "Test response 2",
				Score:     0.85,
				SessionID: "session_124",
				ModelID:   "claude-3-opus-20240229",
				Timestamp: time.Now(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(interactions)
	})

	s.mockServer = httptest.NewServer(mux)

	// Create HTTP API client
	apiClient := client.New(
		client.WithBaseURL(s.mockServer.URL),
	)

	// Create RLHF client
	s.rlhfClient = rlhf.New(
		rlhf.WithAPIClient(apiClient),
		rlhf.WithBaseURL(s.mockServer.URL),
	)

	s.cleanup = func() {
		s.mockServer.Close()
	}
}

// TearDownTest runs after each test in the suite.
func (s *RLHFIntegrationTestSuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// TestInteractionFeedbackSubmission tests submitting feedback for an interaction.
func (s *RLHFIntegrationTestSuite) TestInteractionFeedbackSubmission() {
	// Given: Interaction feedback
	ctx := context.Background()

	feedback := &rlhf.InteractionFeedback{
		Prompt:    "What is the capital of France?",
		Response:  "The capital of France is Paris.",
		Score:     0.95,
		SessionID: "session_abc123",
		ModelID:   "claude-3-opus-20240229",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"user_id":  "user_456",
			"language": "en",
		},
	}

	// When: Submitting the feedback
	response, err := s.rlhfClient.SubmitInteractionFeedback(ctx, feedback)

	// Then: Should submit successfully
	s.Require().NoError(err, "Failed to submit interaction feedback")
	s.NotNil(response, "Response should not be nil")
	s.NotEmpty(response.InteractionID, "Interaction ID should be set")
	s.Equal("success", response.Status, "Status should be success")
	s.False(response.CreatedAt.IsZero(), "CreatedAt should be set")
}

// TestFeedbackScoreValidation tests validation of feedback scores.
func (s *RLHFIntegrationTestSuite) TestFeedbackScoreValidation() {
	// Given: Feedback with invalid score
	ctx := context.Background()

	// When: Submitting feedback with score > 1.0
	invalidFeedback := &rlhf.InteractionFeedback{
		Prompt:   "Test prompt",
		Response: "Test response",
		Score:    1.5, // Invalid: > 1.0
	}

	_, err := s.rlhfClient.SubmitInteractionFeedback(ctx, invalidFeedback)

	// Then: Should return validation error
	s.Error(err, "Should error on score > 1.0")
	s.Contains(err.Error(), "score", "Error should mention score")

	// When: Submitting feedback with score < 0.0
	invalidFeedback2 := &rlhf.InteractionFeedback{
		Prompt:   "Test prompt",
		Response: "Test response",
		Score:    -0.5, // Invalid: < 0.0
	}

	_, err = s.rlhfClient.SubmitInteractionFeedback(ctx, invalidFeedback2)

	// Then: Should return validation error
	s.Error(err, "Should error on score < 0.0")
}

// TestBatchFeedbackSubmission tests submitting multiple feedback entries at once.
func (s *RLHFIntegrationTestSuite) TestBatchFeedbackSubmission() {
	// Given: Multiple interaction feedbacks
	ctx := context.Background()

	batch := &rlhf.BatchInteractionFeedback{
		Interactions: []*rlhf.InteractionFeedback{
			{
				Prompt:    "Question 1",
				Response:  "Answer 1",
				Score:     0.9,
				SessionID: "session_1",
				ModelID:   "claude-3-opus-20240229",
				Timestamp: time.Now(),
			},
			{
				Prompt:    "Question 2",
				Response:  "Answer 2",
				Score:     0.85,
				SessionID: "session_1",
				ModelID:   "claude-3-opus-20240229",
				Timestamp: time.Now(),
			},
			{
				Prompt:    "Question 3",
				Response:  "Answer 3",
				Score:     0.92,
				SessionID: "session_1",
				ModelID:   "claude-3-opus-20240229",
				Timestamp: time.Now(),
			},
		},
	}

	// When: Submitting batch feedback
	response, err := s.rlhfClient.SubmitBatchInteractionFeedback(ctx, batch)

	// Then: Should process all feedback
	s.Require().NoError(err, "Failed to submit batch feedback")
	s.NotNil(response, "Response should not be nil")
	s.Equal(3, response.TotalProcessed, "Should process 3 interactions")
	s.Equal(3, response.SuccessCount, "All 3 should succeed")
	s.Equal(0, response.FailureCount, "No failures expected")
}

// TestCorrectionSubmission tests submitting corrections for AI responses.
func (s *RLHFIntegrationTestSuite) TestCorrectionSubmission() {
	// Given: A correction
	ctx := context.Background()

	correction := &rlhf.Correction{
		InteractionID:     "interaction_12345",
		CorrectedResponse: "This is the corrected and improved response.",
		Reason:            "Original response was incomplete",
		Notes:             "Added more detail and context",
		Tags:              []string{"incomplete", "accuracy"},
	}

	// When: Submitting the correction
	response, err := s.rlhfClient.SubmitCorrection(ctx, correction)

	// Then: Should submit successfully
	s.Require().NoError(err, "Failed to submit correction")
	s.NotNil(response, "Response should not be nil")
	s.NotEmpty(response.CorrectionID, "Correction ID should be set")
	s.Equal("interaction_12345", response.InteractionID, "Interaction ID should match")
	s.Equal("success", response.Status, "Status should be success")
	s.NotNil(response.Diff, "Diff should be present")
	s.False(response.CreatedAt.IsZero(), "CreatedAt should be set")
}

// TestCorrectionDiffGeneration tests that diff is generated for corrections.
func (s *RLHFIntegrationTestSuite) TestCorrectionDiffGeneration() {
	// Given: A correction
	ctx := context.Background()

	correction := &rlhf.Correction{
		InteractionID:     "interaction_67890",
		CorrectedResponse: "Corrected text with improvements",
		Reason:            "Accuracy improvement",
	}

	// When: Submitting the correction
	response, err := s.rlhfClient.SubmitCorrection(ctx, correction)

	// Then: Diff should be generated
	s.Require().NoError(err, "Failed to submit correction")
	s.NotNil(response.Diff, "Diff should be generated")
	s.NotEmpty(response.Diff.Original, "Original text should be present")
	s.NotEmpty(response.Diff.Corrected, "Corrected text should be present")
	s.GreaterOrEqual(response.Diff.SimilarityScore, 0.0, "Similarity score should be >= 0")
	s.LessOrEqual(response.Diff.SimilarityScore, 1.0, "Similarity score should be <= 1")
	s.NotEmpty(response.Diff.Changes, "Changes should be present")
}

// TestAnalyticsRetrieval tests retrieving RLHF analytics.
func (s *RLHFIntegrationTestSuite) TestAnalyticsRetrieval() {
	// Given: Analytics request
	ctx := context.Background()

	req := &rlhf.AnalyticsRequest{
		ModelID:     "claude-3-opus-20240229",
		StartDate:   time.Now().AddDate(0, 0, -30),
		EndDate:     time.Now(),
		Granularity: "day",
	}

	// When: Retrieving analytics
	analytics, err := s.rlhfClient.GetAnalytics(ctx, req)

	// Then: Should retrieve successfully
	s.Require().NoError(err, "Failed to retrieve analytics")
	s.NotNil(analytics, "Analytics should not be nil")
	s.Equal("claude-3-opus-20240229", analytics.ModelID, "Model ID should match")
	s.GreaterOrEqual(analytics.AverageFeedbackScore, 0.0, "Average score should be >= 0")
	s.LessOrEqual(analytics.AverageFeedbackScore, 1.0, "Average score should be <= 1")
	s.GreaterOrEqual(analytics.TotalInteractions, 0, "Total interactions should be >= 0")
	s.GreaterOrEqual(analytics.TotalCorrections, 0, "Total corrections should be >= 0")
}

// TestAnalyticsScoreDistribution tests score distribution in analytics.
func (s *RLHFIntegrationTestSuite) TestAnalyticsScoreDistribution() {
	// Given: Analytics request
	ctx := context.Background()

	req := &rlhf.AnalyticsRequest{
		StartDate: time.Now().AddDate(0, 0, -7),
		EndDate:   time.Now(),
	}

	// When: Retrieving analytics
	analytics, err := s.rlhfClient.GetAnalytics(ctx, req)

	// Then: Should include score distribution
	s.Require().NoError(err, "Failed to retrieve analytics")
	s.NotNil(analytics.ScoreDistribution, "Score distribution should be present")
	s.NotEmpty(analytics.ScoreDistribution, "Score distribution should not be empty")

	// Verify buckets
	expectedBuckets := []string{"0.0-0.2", "0.2-0.4", "0.4-0.6", "0.6-0.8", "0.8-1.0"}
	for _, bucket := range expectedBuckets {
		s.Contains(analytics.ScoreDistribution, bucket, "Should contain bucket: "+bucket)
	}
}

// TestTopCorrectionReasons tests retrieving top correction reasons.
func (s *RLHFIntegrationTestSuite) TestTopCorrectionReasons() {
	// Given: Analytics request
	ctx := context.Background()

	req := &rlhf.AnalyticsRequest{
		StartDate: time.Now().AddDate(0, 0, -30),
		EndDate:   time.Now(),
	}

	// When: Retrieving analytics
	analytics, err := s.rlhfClient.GetAnalytics(ctx, req)

	// Then: Should include top correction reasons
	s.Require().NoError(err, "Failed to retrieve analytics")
	s.NotNil(analytics.TopCorrectionReasons, "Top correction reasons should be present")
	s.NotEmpty(analytics.TopCorrectionReasons, "Top correction reasons should not be empty")

	// Verify structure
	for _, reason := range analytics.TopCorrectionReasons {
		s.NotEmpty(reason.Reason, "Reason should not be empty")
		s.GreaterOrEqual(reason.Count, 0, "Count should be >= 0")
	}
}

// TestAnalyticsExport tests exporting analytics in various formats.
func (s *RLHFIntegrationTestSuite) TestAnalyticsExport() {
	// Given: Export request
	ctx := context.Background()

	testCases := []struct {
		name   string
		format rlhf.ExportFormat
	}{
		{"CSV Export", rlhf.ExportFormatCSV},
		{"JSON Export", rlhf.ExportFormatJSON},
		{"Excel Export", rlhf.ExportFormatExcel},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := &rlhf.AnalyticsExportRequest{
				Analytics: &rlhf.AnalyticsRequest{
					StartDate: time.Now().AddDate(0, 0, -7),
					EndDate:   time.Now(),
				},
				Format: tc.format,
			}

			// When: Exporting analytics
			data, err := s.rlhfClient.ExportAnalytics(ctx, req)

			// Then: Should export successfully
			s.Require().NoError(err, "Failed to export analytics")
			s.NotNil(data, "Export data should not be nil")
			s.NotEmpty(data, "Export data should not be empty")
		})
	}
}

// TestInteractionRetrieval tests retrieving a specific interaction.
func (s *RLHFIntegrationTestSuite) TestInteractionRetrieval() {
	// Given: An interaction ID
	ctx := context.Background()
	interactionID := "interaction_12345"

	// When: Retrieving the interaction
	interaction, err := s.rlhfClient.GetInteraction(ctx, interactionID)

	// Then: Should retrieve successfully
	s.Require().NoError(err, "Failed to retrieve interaction")
	s.NotNil(interaction, "Interaction should not be nil")
	s.NotEmpty(interaction.Prompt, "Prompt should not be empty")
	s.NotEmpty(interaction.Response, "Response should not be empty")
	s.GreaterOrEqual(interaction.Score, 0.0, "Score should be >= 0")
	s.LessOrEqual(interaction.Score, 1.0, "Score should be <= 1")
}

// TestInteractionListing tests listing interactions with filters.
func (s *RLHFIntegrationTestSuite) TestInteractionListing() {
	// Given: Filter parameters
	ctx := context.Background()
	modelID := "claude-3-opus-20240229"
	sessionID := "session_123"
	limit := 10
	offset := 0

	// When: Listing interactions
	interactions, err := s.rlhfClient.ListInteractions(ctx, modelID, sessionID, limit, offset)

	// Then: Should list successfully
	s.Require().NoError(err, "Failed to list interactions")
	s.NotNil(interactions, "Interactions should not be nil")
	s.NotEmpty(interactions, "Interactions list should not be empty")

	// Verify all interactions match filters
	for _, interaction := range interactions {
		s.NotEmpty(interaction.Prompt, "Prompt should not be empty")
		s.NotEmpty(interaction.Response, "Response should not be empty")
	}
}

// TestRequiredFieldValidation tests validation of required fields.
func (s *RLHFIntegrationTestSuite) TestRequiredFieldValidation() {
	// Given: Feedback missing required fields
	ctx := context.Background()

	// When: Submitting feedback without prompt
	feedbackNoPrompt := &rlhf.InteractionFeedback{
		Response: "Response without prompt",
		Score:    0.8,
	}

	_, err := s.rlhfClient.SubmitInteractionFeedback(ctx, feedbackNoPrompt)

	// Then: Should return validation error
	s.Error(err, "Should error when prompt is missing")

	// When: Submitting feedback without response
	feedbackNoResponse := &rlhf.InteractionFeedback{
		Prompt: "Prompt without response",
		Score:  0.8,
	}

	_, err = s.rlhfClient.SubmitInteractionFeedback(ctx, feedbackNoResponse)

	// Then: Should return validation error
	s.Error(err, "Should error when response is missing")
}

// TestCorrectionRequiredFieldValidation tests correction field validation.
func (s *RLHFIntegrationTestSuite) TestCorrectionRequiredFieldValidation() {
	// Given: Correction with missing fields
	ctx := context.Background()

	// When: Submitting correction without interaction ID
	correctionNoID := &rlhf.Correction{
		CorrectedResponse: "Corrected response",
	}

	_, err := s.rlhfClient.SubmitCorrection(ctx, correctionNoID)

	// Then: Should return validation error
	s.Error(err, "Should error when interaction_id is missing")

	// When: Submitting correction without corrected response
	correctionNoResponse := &rlhf.Correction{
		InteractionID: "interaction_123",
	}

	_, err = s.rlhfClient.SubmitCorrection(ctx, correctionNoResponse)

	// Then: Should return validation error
	s.Error(err, "Should error when corrected_response is missing")
}

// TestAnalyticsDateRangeValidation tests validation of date ranges.
func (s *RLHFIntegrationTestSuite) TestAnalyticsDateRangeValidation() {
	// Given: Analytics request with invalid date range
	ctx := context.Background()

	req := &rlhf.AnalyticsRequest{
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 0, -7), // End before start
	}

	// When: Retrieving analytics
	_, err := s.rlhfClient.GetAnalytics(ctx, req)

	// Then: Should return validation error
	s.Error(err, "Should error when end_date is before start_date")
}

// TestTimestampAutoPopulation tests that timestamps are auto-populated.
func (s *RLHFIntegrationTestSuite) TestTimestampAutoPopulation() {
	// Given: Feedback without timestamp
	ctx := context.Background()

	feedback := &rlhf.InteractionFeedback{
		Prompt:   "Test prompt",
		Response: "Test response",
		Score:    0.9,
		// Timestamp not set
	}

	// When: Submitting feedback
	response, err := s.rlhfClient.SubmitInteractionFeedback(ctx, feedback)

	// Then: Timestamp should be auto-populated
	s.Require().NoError(err, "Failed to submit feedback")
	s.False(feedback.Timestamp.IsZero(), "Timestamp should be auto-populated by client")
	s.False(response.CreatedAt.IsZero(), "CreatedAt should be set by server")
}

// TestMetadataHandling tests handling of custom metadata.
func (s *RLHFIntegrationTestSuite) TestMetadataHandling() {
	// Given: Feedback with custom metadata
	ctx := context.Background()

	feedback := &rlhf.InteractionFeedback{
		Prompt:   "Test with metadata",
		Response: "Response with metadata",
		Score:    0.88,
		Metadata: map[string]interface{}{
			"user_id":       "user_789",
			"feature_flag":  "new_ui",
			"response_time": 1.5,
			"tokens_used":   150,
		},
	}

	// When: Submitting feedback
	response, err := s.rlhfClient.SubmitInteractionFeedback(ctx, feedback)

	// Then: Should handle metadata
	s.Require().NoError(err, "Failed to submit feedback with metadata")
	s.NotNil(response, "Response should not be nil")
}

// TestConcurrentFeedbackSubmission tests concurrent feedback submissions.
func (s *RLHFIntegrationTestSuite) TestConcurrentFeedbackSubmission() {
	// Given: Multiple concurrent feedback submissions
	ctx := context.Background()
	concurrentOps := 5
	done := make(chan bool, concurrentOps)
	errors := make(chan error, concurrentOps)

	// When: Submitting feedback concurrently
	for i := 0; i < concurrentOps; i++ {
		go func(index int) {
			feedback := &rlhf.InteractionFeedback{
				Prompt:   "Concurrent test prompt",
				Response: "Concurrent test response",
				Score:    0.9,
			}

			_, err := s.rlhfClient.SubmitInteractionFeedback(ctx, feedback)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < concurrentOps; i++ {
		<-done
	}
	close(errors)

	// Then: All operations should succeed
	s.Empty(errors, "No errors should occur during concurrent submissions")
}

// TestRLHFIntegrationTestSuite runs the test suite.
func TestRLHFIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(RLHFIntegrationTestSuite))
}
