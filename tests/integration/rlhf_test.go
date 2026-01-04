package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/tests/integration/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRLHFFeedbackSubmission_Rating tests feedback rating submission
func TestRLHFFeedbackSubmission_Rating(t *testing.T) {
	t.Run("should submit feedback rating successfully", func(t *testing.T) {
		// Given: Mock RLHF server
		server := mocks.NewRLHFServer()
		defer server.Close()

		// And: Feedback data
		feedbackData := map[string]interface{}{
			"session_id": "sess_123",
			"message_id": "msg_456",
			"rating":     5,
			"comment":    "Excellent response, very helpful!",
			"tags":       []string{"helpful", "accurate", "clear"},
			"metadata": map[string]interface{}{
				"response_time": 1.5,
				"tokens_used":   150,
			},
		}

		// When: Submitting feedback
		body, _ := json.Marshal(feedbackData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should succeed
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.True(t, server.SubmitCalled)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["success"].(bool))
		assert.NotEmpty(t, result["id"])
	})

	t.Run("should accept feedback with minimal data", func(t *testing.T) {
		// Given: Mock RLHF server
		server := mocks.NewRLHFServer()
		defer server.Close()

		// And: Minimal feedback data
		feedbackData := map[string]interface{}{
			"message_id": "msg_789",
			"rating":     3,
		}

		// When: Submitting minimal feedback
		body, _ := json.Marshal(feedbackData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should succeed
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("should validate rating range", func(t *testing.T) {
		// Given: Mock RLHF server with validation
		server := mocks.NewRLHFServer()
		defer server.Close()

		server.ShouldValidate = true

		// And: Invalid rating
		feedbackData := map[string]interface{}{
			"message_id": "msg_999",
			"rating":     10, // Invalid: out of 1-5 range
		}

		// When: Submitting invalid rating
		body, _ := json.Marshal(feedbackData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should be rejected
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should require message_id", func(t *testing.T) {
		// Given: Mock RLHF server with validation
		server := mocks.NewRLHFServer()
		defer server.Close()

		server.ShouldValidate = true

		// And: Feedback without message_id
		feedbackData := map[string]interface{}{
			"rating": 5,
		}

		// When: Submitting without message_id
		body, _ := json.Marshal(feedbackData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should be rejected
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestRLHFFeedbackSubmission_Correction tests correction submission
func TestRLHFFeedbackSubmission_Correction(t *testing.T) {
	t.Run("should submit correction successfully", func(t *testing.T) {
		// Given: Mock RLHF server
		server := mocks.NewRLHFServer()
		defer server.Close()

		// And: Correction data
		correctionData := map[string]interface{}{
			"session_id": "sess_123",
			"message_id": "msg_456",
			"correction": "The correct answer should be X instead of Y",
			"comment":    "The original response was factually incorrect",
			"metadata": map[string]interface{}{
				"error_type": "factual",
				"severity":   "high",
			},
		}

		// When: Submitting correction
		body, _ := json.Marshal(correctionData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/correction", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should succeed
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.True(t, server.CorrectionCalled)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["success"].(bool))
		assert.NotEmpty(t, result["id"])
	})

	t.Run("should validate correction is not empty", func(t *testing.T) {
		// Given: Mock RLHF server with validation
		server := mocks.NewRLHFServer()
		defer server.Close()

		server.ShouldValidate = true

		// And: Empty correction
		correctionData := map[string]interface{}{
			"message_id": "msg_789",
			"correction": "",
		}

		// When: Submitting empty correction
		body, _ := json.Marshal(correctionData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/correction", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should be rejected
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestRLHFFeedbackValidation tests feedback structure validation
func TestRLHFFeedbackValidation(t *testing.T) {
	t.Run("should validate complete feedback structure", func(t *testing.T) {
		// Given: Mock RLHF server
		server := mocks.NewRLHFServer()
		defer server.Close()

		// And: Well-structured feedback
		feedbackData := map[string]interface{}{
			"session_id": "sess_abc123",
			"message_id": "msg_def456",
			"rating":     4,
			"comment":    "Good response but could be more concise",
			"tags":       []string{"helpful", "verbose"},
			"metadata": map[string]interface{}{
				"user_role":     "developer",
				"context":       "code_review",
				"response_time": 2.3,
			},
		}

		// When: Submitting feedback
		body, _ := json.Marshal(feedbackData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should accept all fields
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("should handle multiple feedback submissions", func(t *testing.T) {
		// Given: Mock RLHF server
		server := mocks.NewRLHFServer()
		defer server.Close()

		// When: Submitting multiple feedbacks
		for i := 1; i <= 5; i++ {
			feedbackData := map[string]interface{}{
				"message_id": "msg_" + string(rune('0'+i)),
				"rating":     i,
			}

			body, _ := json.Marshal(feedbackData)
			req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/submit", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", "test-key")

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)

			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
			resp.Body.Close()
		}

		// Then: All should be stored
		assert.Equal(t, 5, server.GetFeedbackCount())
	})

	t.Run("should track different feedback types", func(t *testing.T) {
		// Given: Mock RLHF server
		server := mocks.NewRLHFServer()
		defer server.Close()

		// When: Submitting both ratings and corrections
		ratingData := map[string]interface{}{
			"message_id": "msg_rating",
			"rating":     5,
		}
		body, _ := json.Marshal(ratingData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")
		client := &http.Client{Timeout: 5 * time.Second}
		resp, _ := client.Do(req)
		resp.Body.Close()

		correctionData := map[string]interface{}{
			"message_id": "msg_correction",
			"correction": "This is the correct answer",
		}
		body, _ = json.Marshal(correctionData)
		req, _ = http.NewRequest("POST", server.GetURL()+"/api/feedback/correction", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")
		resp, _ = client.Do(req)
		resp.Body.Close()

		// Then: Should have both types
		ratings := server.GetFeedbackByType("rating")
		corrections := server.GetFeedbackByType("correction")

		assert.Equal(t, 1, len(ratings))
		assert.Equal(t, 1, len(corrections))
	})
}

// TestRLHFFeedbackErrorHandling tests error scenarios
func TestRLHFFeedbackErrorHandling(t *testing.T) {
	t.Run("should handle 401 unauthorized", func(t *testing.T) {
		// Given: Mock server with auth failure
		server := mocks.NewRLHFServer()
		defer server.Close()

		server.ShouldFailAuth = true

		// When: Making request without auth
		feedbackData := map[string]interface{}{
			"message_id": "msg_123",
			"rating":     5,
		}

		body, _ := json.Marshal(feedbackData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 401
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should handle 429 rate limit", func(t *testing.T) {
		// Given: Mock server with rate limiting
		server := mocks.NewRLHFServer()
		defer server.Close()

		server.ShouldRateLimit = true

		// When: Making request
		feedbackData := map[string]interface{}{
			"message_id": "msg_123",
			"rating":     5,
		}

		body, _ := json.Marshal(feedbackData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 429
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		assert.NotEmpty(t, resp.Header.Get("Retry-After"))
	})

	t.Run("should handle invalid JSON", func(t *testing.T) {
		// Given: Mock RLHF server
		server := mocks.NewRLHFServer()
		defer server.Close()

		// When: Sending invalid JSON
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/submit", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 400
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should require API key", func(t *testing.T) {
		// Given: Mock RLHF server
		server := mocks.NewRLHFServer()
		defer server.Close()

		// When: Making request without API key
		feedbackData := map[string]interface{}{
			"message_id": "msg_123",
			"rating":     5,
		}

		body, _ := json.Marshal(feedbackData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 401
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestRLHFFeedbackList tests listing feedback
func TestRLHFFeedbackList(t *testing.T) {
	t.Run("should list all feedback", func(t *testing.T) {
		// Given: Mock RLHF server with feedback
		server := mocks.NewRLHFServer()
		defer server.Close()

		// Add some feedback
		for i := 1; i <= 3; i++ {
			feedbackData := map[string]interface{}{
				"message_id": "msg_" + string(rune('0'+i)),
				"rating":     i,
			}

			body, _ := json.Marshal(feedbackData)
			req, _ := http.NewRequest("POST", server.GetURL()+"/api/feedback/submit", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", "test-key")

			client := &http.Client{Timeout: 5 * time.Second}
			resp, _ := client.Do(req)
			resp.Body.Close()
		}

		// When: Listing feedback
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/feedback/list", nil)
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return all feedback
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Contains(t, result, "feedbacks")
		assert.Equal(t, float64(3), result["count"].(float64))
	})
}
