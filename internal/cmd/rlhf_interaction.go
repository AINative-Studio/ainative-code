package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/client/rlhf"
	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	interactionPrompt    string
	interactionResponse  string
	interactionScore     float64
	interactionModelID   string
	interactionSessionID string
	interactionMetadata  string
	interactionBatchFile string
	interactionAutoCapture bool
)

// rlhfInteractionCmd represents the rlhf interaction command
var rlhfInteractionCmd = &cobra.Command{
	Use:   "interaction",
	Short: "Submit interaction feedback",
	Long: `Submit feedback for an AI interaction with a score from 0.0 to 1.0.

This command allows you to provide feedback on AI interactions, including
the prompt, response, and a numerical score indicating quality.

The score should be between 0.0 (poor) and 1.0 (excellent).

Examples:
  # Submit interaction feedback
  ainative-code rlhf interaction \
    --prompt "What is 2+2?" \
    --response "2+2 equals 4" \
    --score 0.95

  # Include model and session information
  ainative-code rlhf interaction \
    --prompt "Explain quantum computing" \
    --response "Quantum computing uses..." \
    --score 0.85 \
    --model claude-3-5-sonnet-20241022 \
    --session session-123

  # Include metadata
  ainative-code rlhf interaction \
    --prompt "Write a poem" \
    --response "Roses are red..." \
    --score 0.75 \
    --metadata '{"task":"creative_writing","language":"en"}'

  # Submit batch feedback from file
  ainative-code rlhf interaction --batch interactions.json

  # Automatic capture from current session
  ainative-code rlhf interaction --auto-capture`,
	RunE: runRlhfInteraction,
}

func init() {
	// Interaction flags
	rlhfInteractionCmd.Flags().StringVarP(&interactionPrompt, "prompt", "p", "", "user prompt/question")
	rlhfInteractionCmd.Flags().StringVarP(&interactionResponse, "response", "r", "", "AI response")
	rlhfInteractionCmd.Flags().Float64VarP(&interactionScore, "score", "s", 0, "feedback score (0.0-1.0)")
	rlhfInteractionCmd.Flags().StringVar(&interactionModelID, "model", "", "model ID")
	rlhfInteractionCmd.Flags().StringVar(&interactionSessionID, "session", "", "session ID")
	rlhfInteractionCmd.Flags().StringVarP(&interactionMetadata, "metadata", "m", "", "metadata JSON object")
	rlhfInteractionCmd.Flags().StringVar(&interactionBatchFile, "batch", "", "batch file (JSON array of interactions)")
	rlhfInteractionCmd.Flags().BoolVar(&interactionAutoCapture, "auto-capture", false, "automatically capture from current session")
	rlhfInteractionCmd.Flags().BoolP("json", "j", false, "output as JSON")
}

func runRlhfInteraction(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Initialize RLHF client
	rlhfClient, err := initRlhfClient()
	if err != nil {
		return fmt.Errorf("failed to initialize RLHF client: %w", err)
	}

	// Handle batch submission
	if interactionBatchFile != "" {
		return runBatchInteraction(ctx, rlhfClient, jsonOutput)
	}

	// Handle auto-capture
	if interactionAutoCapture {
		return runAutoCapture(ctx, rlhfClient, jsonOutput)
	}

	// Validate required fields for single interaction
	if interactionPrompt == "" {
		return fmt.Errorf("--prompt is required")
	}
	if interactionResponse == "" {
		return fmt.Errorf("--response is required")
	}
	if interactionScore == 0 {
		return fmt.Errorf("--score is required (0.0-1.0)")
	}

	// Create feedback
	feedback := &rlhf.InteractionFeedback{
		Prompt:    interactionPrompt,
		Response:  interactionResponse,
		Score:     interactionScore,
		ModelID:   interactionModelID,
		SessionID: interactionSessionID,
		Timestamp: time.Now(),
	}

	// Parse metadata if provided
	if interactionMetadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(interactionMetadata), &metadata); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
		feedback.Metadata = metadata
	}

	logger.InfoEvent().
		Float64("score", feedback.Score).
		Str("model_id", feedback.ModelID).
		Msg("Submitting interaction feedback")

	// Submit feedback
	result, err := rlhfClient.SubmitInteractionFeedback(ctx, feedback)
	if err != nil {
		return fmt.Errorf("failed to submit feedback: %w", err)
	}

	// Output result
	if jsonOutput {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println("\n✓ Interaction feedback submitted successfully!")
		fmt.Printf("\nInteraction ID: %s\n", result.InteractionID)
		fmt.Printf("Status: %s\n", result.Status)
		if result.Message != "" {
			fmt.Printf("Message: %s\n", result.Message)
		}
		fmt.Printf("Created: %s\n", result.CreatedAt.Format(time.RFC3339))
	}

	return nil
}

func runBatchInteraction(ctx context.Context, rlhfClient *rlhf.Client, jsonOutput bool) error {
	// Read batch file
	data, err := os.ReadFile(interactionBatchFile)
	if err != nil {
		return fmt.Errorf("failed to read batch file: %w", err)
	}

	var batch rlhf.BatchInteractionFeedback
	if err := json.Unmarshal(data, &batch); err != nil {
		return fmt.Errorf("failed to parse batch file: %w", err)
	}

	logger.InfoEvent().
		Int("count", len(batch.Interactions)).
		Msg("Submitting batch interaction feedback")

	// Submit batch
	result, err := rlhfClient.SubmitBatchInteractionFeedback(ctx, &batch)
	if err != nil {
		return fmt.Errorf("failed to submit batch feedback: %w", err)
	}

	// Output result
	if jsonOutput {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println("\n✓ Batch feedback submitted successfully!")
		fmt.Printf("\nTotal Processed: %d\n", result.TotalProcessed)
		fmt.Printf("Successful: %d\n", result.SuccessCount)
		fmt.Printf("Failed: %d\n", result.FailureCount)

		if len(result.Successful) > 0 {
			fmt.Println("\nSuccessful Submissions:")
			for _, id := range result.Successful {
				fmt.Printf("  - %s\n", id)
			}
		}

		if len(result.Failed) > 0 {
			fmt.Println("\nFailed Submissions:")
			for id, reason := range result.Failed {
				fmt.Printf("  - %s: %s\n", id, reason)
			}
		}
	}

	return nil
}

func runAutoCapture(ctx context.Context, rlhfClient *rlhf.Client, jsonOutput bool) error {
	// TODO: Implement auto-capture from current chat session
	// This would integrate with session management to automatically
	// capture interactions from the active session

	fmt.Println("Auto-capture feature coming soon!")
	fmt.Println("This will automatically capture interactions from your current chat session.")

	return nil
}

func initRlhfClient() (*rlhf.Client, error) {
	// Get configuration
	baseURL := viper.GetString("rlhf.base_url")
	if baseURL == "" {
		baseURL = viper.GetString("ainative.base_url")
		if baseURL == "" {
			baseURL = "https://api.ainative.studio"
		}
	}

	// Create API client
	apiClient := client.New(
		client.WithBaseURL(baseURL),
		// Note: WithServicePath may not be available in current client version
		// Using base URL with path included instead
	)

	// Create RLHF client
	rlhfClient := rlhf.New(
		rlhf.WithAPIClient(apiClient),
		rlhf.WithBaseURL(baseURL+"/v1/rlhf"),
	)

	return rlhfClient, nil
}

// Helper function to display interactions in table format
func displayInteractionsTable(interactions []*rlhf.InteractionFeedback) {
	// Print header
	fmt.Printf("%-50s | %-6s | %-20s | %-15s | %s\n",
		"Prompt", "Score", "Model", "Session", "Timestamp")
	fmt.Println(strings.Repeat("-", 120))

	// Print rows
	for _, interaction := range interactions {
		prompt := interaction.Prompt
		if len(prompt) > 50 {
			prompt = prompt[:47] + "..."
		}

		fmt.Printf("%-50s | %6.2f | %-20s | %-15s | %s\n",
			prompt,
			interaction.Score,
			truncate(interaction.ModelID, 20),
			truncate(interaction.SessionID, 15),
			interaction.Timestamp.Format("2006-01-02 15:04"),
		)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatJSON(data interface{}) string {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", data)
	}
	return string(output)
}

// Additional helper for pretty-printing metadata
func formatMetadata(metadata map[string]interface{}) string {
	if metadata == nil || len(metadata) == 0 {
		return "-"
	}

	parts := make([]string, 0, len(metadata))
	for k, v := range metadata {
		parts = append(parts, fmt.Sprintf("%s:%v", k, v))
	}
	return strings.Join(parts, ", ")
}
