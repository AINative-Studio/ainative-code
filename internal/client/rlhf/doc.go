// Package rlhf provides a client for AINative RLHF (Reinforcement Learning from Human Feedback) operations.
//
// The RLHF client enables interaction feedback collection, correction submission,
// and analytics viewing for improving AI model performance through human feedback.
//
// Features:
//   - Interaction feedback submission with scores (0.0-1.0)
//   - Correction submission with diff visualization
//   - Automatic interaction capture from chat sessions
//   - Batch feedback submission
//   - Analytics viewing and export
//
// Example usage:
//
//	import (
//	    "context"
//	    "github.com/AINative-studio/ainative-code/internal/client"
//	    "github.com/AINative-studio/ainative-code/internal/client/rlhf"
//	)
//
//	// Create HTTP client
//	apiClient := client.New(
//	    client.WithBaseURL("https://api.ainative.studio"),
//	    client.WithServicePath("rlhf", "/v1/rlhf"),
//	)
//
//	// Create RLHF client
//	rlhfClient := rlhf.New(
//	    rlhf.WithAPIClient(apiClient),
//	    rlhf.WithBaseURL("https://api.ainative.studio/v1/rlhf"),
//	)
//
//	// Submit interaction feedback
//	feedback := &rlhf.InteractionFeedback{
//	    Prompt: "What is the capital of France?",
//	    Response: "Paris is the capital of France.",
//	    Score: 0.95,
//	    Metadata: map[string]interface{}{
//	        "model": "claude-3-5-sonnet-20241022",
//	        "session_id": "session-123",
//	    },
//	}
//	result, err := rlhfClient.SubmitInteractionFeedback(context.Background(), feedback)
package rlhf
