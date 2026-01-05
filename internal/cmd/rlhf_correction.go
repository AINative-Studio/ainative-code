package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client/rlhf"
	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	correctionInteractionID string
	correctionResponse      string
	correctionReason        string
	correctionNotes         string
	correctionTags          []string
	correctionShowDiff      bool
)

// rlhfCorrectionCmd represents the rlhf correction command
var rlhfCorrectionCmd = &cobra.Command{
	Use:   "correction",
	Short: "Submit correction for AI response",
	Long: `Submit a correction for an incorrect or suboptimal AI response.

This command allows you to provide a corrected version of an AI response
along with the reason for the correction. The system will generate a diff
showing the changes between the original and corrected responses.

Corrections help improve the AI model by identifying and documenting
incorrect or suboptimal responses.

Examples:
  # Submit a correction with reason
  ainative-code rlhf correction \
    --interaction-id interaction-123 \
    --corrected-response "The corrected answer is Paris" \
    --reason "Inaccurate information"

  # Include notes and tags
  ainative-code rlhf correction \
    --interaction-id interaction-123 \
    --corrected-response "Improved response text" \
    --reason "Poor formatting" \
    --notes "The original response was hard to read" \
    --tags accuracy,formatting

  # Show diff visualization
  ainative-code rlhf correction \
    --interaction-id interaction-123 \
    --corrected-response "Better response" \
    --reason "Clarity" \
    --show-diff`,
	RunE: runRlhfCorrection,
}

func init() {
	// Correction flags
	rlhfCorrectionCmd.Flags().StringVarP(&correctionInteractionID, "interaction-id", "i", "", "interaction ID to correct (required)")
	rlhfCorrectionCmd.Flags().StringVarP(&correctionResponse, "corrected-response", "r", "", "corrected response text (required)")
	rlhfCorrectionCmd.Flags().StringVar(&correctionReason, "reason", "", "reason for correction")
	rlhfCorrectionCmd.Flags().StringVarP(&correctionNotes, "notes", "n", "", "additional notes")
	rlhfCorrectionCmd.Flags().StringSliceVarP(&correctionTags, "tags", "t", []string{}, "correction tags (accuracy,clarity,etc)")
	rlhfCorrectionCmd.Flags().BoolVar(&correctionShowDiff, "show-diff", true, "show diff visualization")
	rlhfCorrectionCmd.Flags().BoolP("json", "j", false, "output as JSON")

	rlhfCorrectionCmd.MarkFlagRequired("interaction-id")
	rlhfCorrectionCmd.MarkFlagRequired("corrected-response")
}

func runRlhfCorrection(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Initialize RLHF client
	rlhfClient, err := initRlhfClient()
	if err != nil {
		return fmt.Errorf("failed to initialize RLHF client: %w", err)
	}

	// First, get the original interaction to show context
	var originalInteraction *rlhf.InteractionFeedback
	if correctionShowDiff && !jsonOutput {
		originalInteraction, err = rlhfClient.GetInteraction(ctx, correctionInteractionID)
		if err != nil {
			logger.WarnEvent().
				Err(err).
				Msg("Could not fetch original interaction for diff")
		}
	}

	// Create correction
	correction := &rlhf.Correction{
		InteractionID:     correctionInteractionID,
		CorrectedResponse: correctionResponse,
		Reason:            correctionReason,
		Notes:             correctionNotes,
		Tags:              correctionTags,
	}

	logger.InfoEvent().
		Str("interaction_id", correction.InteractionID).
		Str("reason", correction.Reason).
		Int("tags_count", len(correction.Tags)).
		Msg("Submitting correction")

	// Submit correction
	result, err := rlhfClient.SubmitCorrection(ctx, correction)
	if err != nil {
		return fmt.Errorf("failed to submit correction: %w", err)
	}

	// Output result
	if jsonOutput {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println("\n✓ Correction submitted successfully!")
		fmt.Printf("\nCorrection ID: %s\n", result.CorrectionID)
		fmt.Printf("Interaction ID: %s\n", result.InteractionID)
		fmt.Printf("Status: %s\n", result.Status)
		if result.Message != "" {
			fmt.Printf("Message: %s\n", result.Message)
		}
		fmt.Printf("Created: %s\n", result.CreatedAt.Format(time.RFC3339))

		// Show diff if available and requested
		if correctionShowDiff && result.Diff != nil {
			fmt.Println("\n" + strings.Repeat("=", 80))
			displayDiff(originalInteraction, result.Diff)
			fmt.Println(strings.Repeat("=", 80))

			if result.Diff.SimilarityScore > 0 {
				fmt.Printf("\nSimilarity Score: %.2f%%\n", result.Diff.SimilarityScore*100)
			}
		}
	}

	return nil
}

// displayDiff shows a color-coded diff between original and corrected responses
func displayDiff(original *rlhf.InteractionFeedback, diff *rlhf.DiffResult) {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Println(bold("\nDiff Visualization:"))
	fmt.Println()

	// Show original prompt if available
	if original != nil && original.Prompt != "" {
		fmt.Printf("%s\n", yellow("Original Prompt:"))
		fmt.Printf("%s\n\n", original.Prompt)
	}

	// Show original response
	fmt.Printf("%s\n", red("- Original Response:"))
	originalLines := strings.Split(diff.Original, "\n")
	for _, line := range originalLines {
		fmt.Printf("%s %s\n", red("-"), line)
	}

	fmt.Println()

	// Show corrected response
	fmt.Printf("%s\n", green("+ Corrected Response:"))
	correctedLines := strings.Split(diff.Corrected, "\n")
	for _, line := range correctedLines {
		fmt.Printf("%s %s\n", green("+"), line)
	}

	// Show individual changes if available
	if len(diff.Changes) > 0 {
		fmt.Println()
		fmt.Printf("%s\n", bold("Changes:"))
		for i, change := range diff.Changes {
			var typeColor func(...interface{}) string
			switch change.Type {
			case "add":
				typeColor = green
			case "remove":
				typeColor = red
			case "modify":
				typeColor = yellow
			default:
				typeColor = func(a ...interface{}) string { return fmt.Sprint(a...) }
			}

			fmt.Printf("  %d. [%s] Line %d: %s\n",
				i+1,
				typeColor(change.Type),
				change.Line,
				change.Content,
			)
		}
	}
}

// Helper function to calculate simple similarity score
func calculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	if s1 == "" || s2 == "" {
		return 0.0
	}

	// Simple character-based similarity
	longer := s1
	shorter := s2
	if len(s2) > len(s1) {
		longer = s2
		shorter = s1
	}

	longerLength := len(longer)
	if longerLength == 0 {
		return 1.0
	}

	// Count matching characters
	matches := 0
	for i, c := range shorter {
		if i < len(longer) && longer[i] == byte(c) {
			matches++
		}
	}

	return float64(matches) / float64(longerLength)
}

// displayCorrectionSummary shows a summary of corrections
func displayCorrectionSummary(corrections []*rlhf.CorrectionResponse) {
	if len(corrections) == 0 {
		fmt.Println("No corrections found.")
		return
	}

	fmt.Printf("\n%-20s | %-36s | %-15s | %s\n",
		"Correction ID", "Interaction ID", "Status", "Created")
	fmt.Println(strings.Repeat("-", 100))

	for _, corr := range corrections {
		fmt.Printf("%-20s | %-36s | %-15s | %s\n",
			truncate(corr.CorrectionID, 20),
			truncate(corr.InteractionID, 36),
			corr.Status,
			corr.CreatedAt.Format("2006-01-02 15:04"),
		)
	}

	fmt.Printf("\nTotal: %d corrections\n", len(corrections))
}
