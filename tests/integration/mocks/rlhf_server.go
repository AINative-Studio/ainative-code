package mocks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// RLHFServer represents a mock RLHF feedback API server
type RLHFServer struct {
	Server           *httptest.Server
	Feedbacks        []FeedbackRecord
	mu               sync.RWMutex
	ShouldFailAuth   bool
	ShouldRateLimit  bool
	ShouldValidate   bool
	ResponseDelay    time.Duration
	SubmitCalled     bool
	CorrectionCalled bool
}

// FeedbackRecord represents an RLHF feedback submission
type FeedbackRecord struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "rating", "correction", "preference"
	SessionID   string                 `json:"session_id"`
	MessageID   string                 `json:"message_id"`
	Rating      int                    `json:"rating,omitempty"`
	Comment     string                 `json:"comment,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Correction  string                 `json:"correction,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	SubmittedAt time.Time              `json:"submitted_at"`
}

// NewRLHFServer creates a new mock RLHF server
func NewRLHFServer() *RLHFServer {
	rs := &RLHFServer{
		Feedbacks: make([]FeedbackRecord, 0),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/feedback/submit", rs.handleSubmitFeedback)
	mux.HandleFunc("/api/feedback/correction", rs.handleSubmitCorrection)
	mux.HandleFunc("/api/feedback/list", rs.handleListFeedback)

	rs.Server = httptest.NewServer(rs.authMiddleware(mux))
	return rs
}

// Close shuts down the mock server
func (rs *RLHFServer) Close() {
	rs.Server.Close()
}

// authMiddleware validates API authentication
func (rs *RLHFServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rs.ResponseDelay > 0 {
			time.Sleep(rs.ResponseDelay)
		}

		if rs.ShouldFailAuth {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "unauthorized",
				"message": "Invalid API credentials",
			})
			return
		}

		if rs.ShouldRateLimit {
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "rate_limit_exceeded",
				"message": "Too many feedback submissions",
			})
			return
		}

		// Validate API key
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "missing_api_key",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleSubmitFeedback handles feedback submission
func (rs *RLHFServer) handleSubmitFeedback(w http.ResponseWriter, r *http.Request) {
	rs.SubmitCalled = true
	rs.mu.Lock()
	defer rs.mu.Unlock()

	var req struct {
		SessionID string                 `json:"session_id"`
		MessageID string                 `json:"message_id"`
		Rating    int                    `json:"rating"`
		Comment   string                 `json:"comment"`
		Tags      []string               `json:"tags"`
		Metadata  map[string]interface{} `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": "Invalid JSON body",
		})
		return
	}

	// Validate required fields
	if rs.ShouldValidate {
		if req.MessageID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "validation_error",
				"message": "message_id is required",
			})
			return
		}

		if req.Rating < 1 || req.Rating > 5 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "validation_error",
				"message": "rating must be between 1 and 5",
			})
			return
		}
	}

	// Create feedback record
	feedback := FeedbackRecord{
		ID:          generateRandomString(16),
		Type:        "rating",
		SessionID:   req.SessionID,
		MessageID:   req.MessageID,
		Rating:      req.Rating,
		Comment:     req.Comment,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
		SubmittedAt: time.Now(),
	}

	rs.Feedbacks = append(rs.Feedbacks, feedback)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      feedback.ID,
		"message": "Feedback submitted successfully",
	})
}

// handleSubmitCorrection handles correction submission
func (rs *RLHFServer) handleSubmitCorrection(w http.ResponseWriter, r *http.Request) {
	rs.CorrectionCalled = true
	rs.mu.Lock()
	defer rs.mu.Unlock()

	var req struct {
		SessionID  string                 `json:"session_id"`
		MessageID  string                 `json:"message_id"`
		Correction string                 `json:"correction"`
		Comment    string                 `json:"comment"`
		Metadata   map[string]interface{} `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": "Invalid JSON body",
		})
		return
	}

	// Validate required fields
	if rs.ShouldValidate {
		if req.MessageID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "validation_error",
				"message": "message_id is required",
			})
			return
		}

		if req.Correction == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "validation_error",
				"message": "correction is required",
			})
			return
		}
	}

	// Create correction record
	feedback := FeedbackRecord{
		ID:          generateRandomString(16),
		Type:        "correction",
		SessionID:   req.SessionID,
		MessageID:   req.MessageID,
		Correction:  req.Correction,
		Comment:     req.Comment,
		Metadata:    req.Metadata,
		SubmittedAt: time.Now(),
	}

	rs.Feedbacks = append(rs.Feedbacks, feedback)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      feedback.ID,
		"message": "Correction submitted successfully",
	})
}

// handleListFeedback lists all feedback
func (rs *RLHFServer) handleListFeedback(w http.ResponseWriter, r *http.Request) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"feedbacks": rs.Feedbacks,
		"count":     len(rs.Feedbacks),
	})
}

// GetURL returns the base URL of the mock server
func (rs *RLHFServer) GetURL() string {
	return rs.Server.URL
}

// Reset clears all stored feedback
func (rs *RLHFServer) Reset() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.Feedbacks = make([]FeedbackRecord, 0)
}

// GetFeedbackCount returns the number of stored feedback records
func (rs *RLHFServer) GetFeedbackCount() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return len(rs.Feedbacks)
}

// GetFeedbackByType returns feedback filtered by type
func (rs *RLHFServer) GetFeedbackByType(feedbackType string) []FeedbackRecord {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	results := make([]FeedbackRecord, 0)
	for _, fb := range rs.Feedbacks {
		if fb.Type == feedbackType {
			results = append(results, fb)
		}
	}
	return results
}
