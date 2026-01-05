package rlhf

import "time"

// InteractionFeedback represents feedback for an AI interaction.
type InteractionFeedback struct {
	// Prompt is the user's input/question
	Prompt string `json:"prompt"`

	// Response is the AI's response
	Response string `json:"response"`

	// Score is the feedback score (0.0 to 1.0)
	Score float64 `json:"score"`

	// Metadata contains additional context
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// SessionID links the interaction to a chat session
	SessionID string `json:"session_id,omitempty"`

	// ModelID identifies the model used
	ModelID string `json:"model_id,omitempty"`

	// Timestamp of the interaction
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// InteractionFeedbackResponse represents the API response for feedback submission.
type InteractionFeedbackResponse struct {
	// InteractionID is the unique identifier for this interaction
	InteractionID string `json:"interaction_id"`

	// Status indicates success/failure
	Status string `json:"status"`

	// Message provides additional information
	Message string `json:"message,omitempty"`

	// CreatedAt is when the feedback was recorded
	CreatedAt time.Time `json:"created_at"`
}

// BatchInteractionFeedback represents multiple feedback submissions.
type BatchInteractionFeedback struct {
	Interactions []*InteractionFeedback `json:"interactions"`
}

// BatchInteractionFeedbackResponse represents the response for batch submission.
type BatchInteractionFeedbackResponse struct {
	// Successful submissions
	Successful []string `json:"successful"`

	// Failed submissions with reasons
	Failed map[string]string `json:"failed,omitempty"`

	// TotalProcessed count
	TotalProcessed int `json:"total_processed"`

	// SuccessCount count
	SuccessCount int `json:"success_count"`

	// FailureCount count
	FailureCount int `json:"failure_count"`
}

// Correction represents a correction to an AI response.
type Correction struct {
	// InteractionID is the ID of the interaction being corrected
	InteractionID string `json:"interaction_id"`

	// CorrectedResponse is the improved response
	CorrectedResponse string `json:"corrected_response"`

	// Reason explains why the correction was needed
	Reason string `json:"reason,omitempty"`

	// Notes provides additional context
	Notes string `json:"notes,omitempty"`

	// Tags for categorization
	Tags []string `json:"tags,omitempty"`
}

// CorrectionResponse represents the API response for correction submission.
type CorrectionResponse struct {
	// CorrectionID is the unique identifier
	CorrectionID string `json:"correction_id"`

	// InteractionID links to the original interaction
	InteractionID string `json:"interaction_id"`

	// Status indicates success/failure
	Status string `json:"status"`

	// Message provides additional information
	Message string `json:"message,omitempty"`

	// Diff shows the changes made
	Diff *DiffResult `json:"diff,omitempty"`

	// CreatedAt is when the correction was recorded
	CreatedAt time.Time `json:"created_at"`
}

// DiffResult represents the diff between original and corrected responses.
type DiffResult struct {
	// Original response
	Original string `json:"original"`

	// Corrected response
	Corrected string `json:"corrected"`

	// Changes is a list of change operations
	Changes []*Change `json:"changes"`

	// SimilarityScore (0.0 to 1.0)
	SimilarityScore float64 `json:"similarity_score"`
}

// Change represents a single change in a diff.
type Change struct {
	// Type: "add", "remove", "modify"
	Type string `json:"type"`

	// Line number
	Line int `json:"line"`

	// Content of the change
	Content string `json:"content"`
}

// Analytics represents RLHF analytics data.
type Analytics struct {
	// Model ID
	ModelID string `json:"model_id"`

	// Time range
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`

	// Metrics
	AverageFeedbackScore float64 `json:"average_feedback_score"`
	TotalInteractions    int     `json:"total_interactions"`
	TotalCorrections     int     `json:"total_corrections"`
	CorrectionRate       float64 `json:"correction_rate"` // Percentage

	// Distribution
	ScoreDistribution map[string]int `json:"score_distribution"` // Buckets: "0.0-0.2", "0.2-0.4", etc.

	// Top issues
	TopCorrectionReasons []ReasonCount `json:"top_correction_reasons,omitempty"`

	// Trending data
	TrendingData []TrendPoint `json:"trending_data,omitempty"`
}

// ReasonCount represents a correction reason and its frequency.
type ReasonCount struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

// TrendPoint represents a data point in a trend.
type TrendPoint struct {
	Date  time.Time `json:"date"`
	Score float64   `json:"score"`
	Count int       `json:"count"`
}

// AnalyticsRequest represents a request for analytics data.
type AnalyticsRequest struct {
	// ModelID to filter by (optional)
	ModelID string `json:"model_id,omitempty"`

	// Date range
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`

	// Granularity for trending data: "hour", "day", "week", "month"
	Granularity string `json:"granularity,omitempty"`
}

// ExportFormat specifies the export format.
type ExportFormat string

const (
	// ExportFormatCSV exports to CSV
	ExportFormatCSV ExportFormat = "csv"

	// ExportFormatJSON exports to JSON
	ExportFormatJSON ExportFormat = "json"

	// ExportFormatExcel exports to Excel
	ExportFormatExcel ExportFormat = "excel"
)

// AnalyticsExportRequest represents a request to export analytics.
type AnalyticsExportRequest struct {
	Analytics *AnalyticsRequest `json:"analytics"`
	Format    ExportFormat      `json:"format"`
}
