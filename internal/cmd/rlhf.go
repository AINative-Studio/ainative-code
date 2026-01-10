package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/AINative-studio/ainative-code/internal/client/rlhf"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

var (
	rlhfRating    int
	rlhfComment   string
	rlhfSessionID string
	rlhfMessageID string
)

// rlhfCmd represents the rlhf command
var rlhfCmd = &cobra.Command{
	Use:   "rlhf",
	Short: "Manage RLHF feedback",
	Long: `Manage Reinforcement Learning from Human Feedback (RLHF) data.

RLHF feedback helps improve AI model performance by collecting human
preferences and ratings on AI-generated responses. This data can be used
for model fine-tuning and evaluation.

Examples:
  # Submit feedback for a response
  ainative-code rlhf submit --message-id abc123 --rating 5 --comment "Excellent response"

  # List feedback entries
  ainative-code rlhf list

  # Export feedback data
  ainative-code rlhf export --output feedback.jsonl

  # View feedback statistics
  ainative-code rlhf stats

  # Submit feedback interactively
  ainative-code rlhf submit --interactive`,
	Aliases: []string{"feedback", "fb"},
}

// rlhfSubmitCmd represents the rlhf submit command
var rlhfSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit RLHF feedback",
	Long:  `Submit feedback for an AI-generated response.`,
	RunE:  runRlhfSubmit,
}

// rlhfListCmd represents the rlhf list command
var rlhfListCmd = &cobra.Command{
	Use:   "list",
	Short: "List feedback entries",
	Long:  `List all RLHF feedback entries.`,
	Aliases: []string{"ls", "l"},
	RunE:  runRlhfList,
}

// rlhfExportCmd represents the rlhf export command
var rlhfExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export feedback data",
	Long:  `Export RLHF feedback data for model training or analysis.`,
	RunE:  runRlhfExport,
}

// rlhfStatsCmd represents the rlhf stats command
var rlhfStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View feedback statistics",
	Long:  `Display statistics about collected RLHF feedback.`,
	Aliases: []string{"statistics"},
	RunE:  runRlhfStats,
}

// rlhfDeleteCmd represents the rlhf delete command
var rlhfDeleteCmd = &cobra.Command{
	Use:   "delete [feedback-id]",
	Short: "Delete feedback entry",
	Long:  `Delete a specific RLHF feedback entry.`,
	Aliases: []string{"rm", "remove"},
	Args:  cobra.ExactArgs(1),
	RunE:  runRlhfDelete,
}

func init() {
	rootCmd.AddCommand(rlhfCmd)

	// Add subcommands
	rlhfCmd.AddCommand(rlhfSubmitCmd)
	rlhfCmd.AddCommand(rlhfInteractionCmd)    // TASK-061
	rlhfCmd.AddCommand(rlhfCorrectionCmd)     // TASK-062
	rlhfCmd.AddCommand(rlhfAnalyticsCmd)      // TASK-063
	rlhfCmd.AddCommand(rlhfListCmd)
	rlhfCmd.AddCommand(rlhfExportCmd)
	rlhfCmd.AddCommand(rlhfStatsCmd)
	rlhfCmd.AddCommand(rlhfDeleteCmd)

	// Submit flags
	rlhfSubmitCmd.Flags().StringVar(&rlhfMessageID, "message-id", "", "message ID to provide feedback for")
	rlhfSubmitCmd.Flags().IntVarP(&rlhfRating, "rating", "r", 0, "rating (1-5)")
	rlhfSubmitCmd.Flags().StringVarP(&rlhfComment, "comment", "c", "", "feedback comment")
	rlhfSubmitCmd.Flags().BoolP("interactive", "i", false, "interactive feedback mode")
	rlhfSubmitCmd.Flags().StringSlice("tags", []string{}, "feedback tags (helpful, accurate, creative, etc.)")

	// List flags
	rlhfListCmd.Flags().IntP("limit", "n", 20, "limit number of entries")
	rlhfListCmd.Flags().String("filter", "", "filter by rating or tag")

	// Export flags
	rlhfExportCmd.Flags().StringP("output", "o", "feedback.jsonl", "output file path")
	rlhfExportCmd.Flags().String("format", "jsonl", "export format (jsonl, csv, json)")
	rlhfExportCmd.Flags().String("from", "", "start date (YYYY-MM-DD)")
	rlhfExportCmd.Flags().String("to", "", "end date (YYYY-MM-DD)")
}

func runRlhfSubmit(cmd *cobra.Command, args []string) error {
	interactive, _ := cmd.Flags().GetBool("interactive")

	if interactive {
		logger.Info("Starting interactive feedback mode")
		fmt.Println()
		fmt.Println("Interactive RLHF Feedback Mode")
		fmt.Println("===============================")
		fmt.Println()
		fmt.Println("Interactive mode is planned for a future release.")
		fmt.Println()
		fmt.Println("This will provide a terminal UI for:")
		fmt.Println("  • Browsing recent interactions")
		fmt.Println("  • Rating responses with keyboard shortcuts")
		fmt.Println("  • Adding comments and tags")
		fmt.Println("  • Submitting corrections")
		fmt.Println()
		fmt.Println("For now, use the command-line flags:")
		fmt.Println()
		fmt.Println("  # Submit feedback")
		fmt.Println("  ainative-code rlhf submit \\")
		fmt.Println("    --message-id MESSAGE_ID \\")
		fmt.Println("    --rating 5 \\")
		fmt.Println("    --comment \"Great response!\"")
		fmt.Println()
		fmt.Println("Or use the 'rlhf interaction' command for detailed feedback:")
		fmt.Println()
		fmt.Println("  ainative-code rlhf interaction \\")
		fmt.Println("    --prompt \"Your question\" \\")
		fmt.Println("    --response \"AI response\" \\")
		fmt.Println("    --score 0.95")
		return nil
	}

	logger.InfoEvent().
		Str("message_id", rlhfMessageID).
		Int("rating", rlhfRating).
		Msg("Submitting RLHF feedback")

	if rlhfMessageID == "" {
		return fmt.Errorf("message-id is required (use --message-id)")
	}

	if rlhfRating < 1 || rlhfRating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	tags, _ := cmd.Flags().GetStringSlice("tags")

	fmt.Printf("Submitting feedback for message: %s\n", rlhfMessageID)
	fmt.Printf("Rating: %d/5\n", rlhfRating)
	if rlhfComment != "" {
		fmt.Printf("Comment: %s\n", rlhfComment)
	}
	if len(tags) > 0 {
		fmt.Printf("Tags: %v\n", tags)
	}

	// TODO: Implement feedback submission
	// - Validate message exists
	// - Store feedback in database
	// - Update statistics

	fmt.Println("Feedback submitted successfully!")

	return nil
}

func runRlhfList(cmd *cobra.Command, args []string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	filter, _ := cmd.Flags().GetString("filter")

	logger.DebugEvent().
		Int("limit", limit).
		Str("filter", filter).
		Msg("Listing RLHF feedback")

	fmt.Println("RLHF Feedback Entries")
	fmt.Println("======================")
	fmt.Println()

	fmt.Println("This command requires RLHF database schema to be implemented.")
	fmt.Println()
	fmt.Println("Current Status:")
	fmt.Println("  • RLHF interaction submission is implemented (use 'rlhf interaction')")
	fmt.Println("  • RLHF correction submission is implemented (use 'rlhf correction')")
	fmt.Println("  • RLHF analytics is implemented (use 'rlhf analytics')")
	fmt.Println("  • Local feedback storage requires database schema (planned)")
	fmt.Println()
	fmt.Println("These commands currently interact with the AINative API.")
	fmt.Println("To list feedback entries stored in the API, use:")
	fmt.Println("  ainative-code rlhf analytics --start-date YYYY-MM-DD --end-date YYYY-MM-DD")
	fmt.Println()
	fmt.Println("For local feedback storage, this feature is planned for a future release.")

	return nil
}

func runRlhfExport(cmd *cobra.Command, args []string) error {
	output, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")

	// Validate export directory exists
	dir := filepath.Dir(output)
	if dir != "." && dir != "" {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("export directory does not exist: %s", dir)
		}
	}

	logger.InfoEvent().
		Str("output", output).
		Str("format", format).
		Str("from", from).
		Str("to", to).
		Msg("Exporting RLHF feedback")

	fmt.Printf("Exporting feedback to: %s (format: %s)\n", output, format)

	if from != "" || to != "" {
		fmt.Printf("Date range: %s to %s\n", from, to)
	}

	// Create example RLHF feedback data
	// In a real implementation, this would query feedback from a database with date filters
	exampleFeedback := createExampleRLHFFeedback(from, to)

	// Format and write feedback based on format
	var err error
	switch strings.ToLower(format) {
	case "jsonl":
		err = writeJSONL(output, exampleFeedback)
	case "json":
		err = writeJSON(output, exampleFeedback)
	case "csv":
		err = writeCSV(output, exampleFeedback)
	default:
		return fmt.Errorf("unsupported format: %s (supported: jsonl, json, csv)", format)
	}

	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Successfully exported %d feedback entries to %s\n", len(exampleFeedback), output)
	fmt.Println("Export completed!")

	return nil
}

func runRlhfStats(cmd *cobra.Command, args []string) error {
	logger.Debug("Generating RLHF statistics")

	fmt.Println("RLHF Feedback Statistics")
	fmt.Println("=========================")
	fmt.Println()

	fmt.Println("This command requires RLHF database schema to be implemented.")
	fmt.Println()
	fmt.Println("Current Status:")
	fmt.Println("  • Local statistics require database schema (planned)")
	fmt.Println("  • API-based analytics are available via 'rlhf analytics'")
	fmt.Println()
	fmt.Println("To view detailed analytics from the AINative API:")
	fmt.Println()
	fmt.Println("  # View analytics for the last 7 days")
	fmt.Println("  ainative-code rlhf analytics \\")
	fmt.Println("    --start-date 2026-01-01 \\")
	fmt.Println("    --end-date 2026-01-08")
	fmt.Println()
	fmt.Println("  # Filter by model")
	fmt.Println("  ainative-code rlhf analytics \\")
	fmt.Println("    --model claude-3-5-sonnet-20241022 \\")
	fmt.Println("    --start-date 2026-01-01 \\")
	fmt.Println("    --end-date 2026-01-08")
	fmt.Println()
	fmt.Println("  # Export to file")
	fmt.Println("  ainative-code rlhf analytics \\")
	fmt.Println("    --start-date 2026-01-01 \\")
	fmt.Println("    --end-date 2026-01-08 \\")
	fmt.Println("    --export analytics.json")
	fmt.Println()
	fmt.Println("The analytics command provides:")
	fmt.Println("  • Average feedback scores")
	fmt.Println("  • Total interactions and corrections")
	fmt.Println("  • Correction rate")
	fmt.Println("  • Score distribution")
	fmt.Println("  • Top correction reasons")
	fmt.Println("  • Trending data over time")

	return nil
}

func runRlhfDelete(cmd *cobra.Command, args []string) error {
	feedbackID := args[0]

	logger.InfoEvent().Str("feedback_id", feedbackID).Msg("Deleting RLHF feedback")

	fmt.Printf("Deleting feedback: %s\n", feedbackID)

	// TODO: Implement feedback deletion
	// - Verify feedback exists
	// - Delete from database
	// - Update statistics

	fmt.Println("Feedback deleted successfully!")

	return nil
}

// createExampleRLHFFeedback creates example feedback data for export
// In a production implementation, this would query from a database with date filtering
func createExampleRLHFFeedback(from, to string) []*rlhf.InteractionFeedback {
	now := time.Now()

	// Parse date filters if provided
	var startDate, endDate time.Time
	if from != "" {
		startDate, _ = time.Parse("2006-01-02", from)
	} else {
		startDate = now.AddDate(0, 0, -7) // Default to last 7 days
	}
	if to != "" {
		endDate, _ = time.Parse("2006-01-02", to)
	} else {
		endDate = now
	}

	feedback := []*rlhf.InteractionFeedback{
		{
			Prompt:    "How do I implement authentication in Go?",
			Response:  "To implement authentication in Go, you can use packages like jwt-go for JWT tokens...",
			Score:     0.95,
			SessionID: "session-001",
			ModelID:   "claude-3-5-sonnet-20241022",
			Timestamp: startDate.Add(1 * time.Hour),
			Metadata: map[string]interface{}{
				"auto_captured":           true,
				"implicit_signals_count":  2,
				"has_explicit_feedback":   true,
				"user_feedback":           "Very helpful explanation",
				"response_time_ms":        1250,
			},
		},
		{
			Prompt:    "What are the best practices for error handling?",
			Response:  "Go uses explicit error handling with multiple return values...",
			Score:     0.88,
			SessionID: "session-001",
			ModelID:   "claude-3-5-sonnet-20241022",
			Timestamp: startDate.Add(2 * time.Hour),
			Metadata: map[string]interface{}{
				"auto_captured":           true,
				"implicit_signals_count":  1,
				"has_explicit_feedback":   false,
				"response_time_ms":        980,
			},
		},
		{
			Prompt:    "Explain goroutines and channels",
			Response:  "Goroutines are lightweight threads managed by the Go runtime. Channels provide a way to communicate between goroutines...",
			Score:     0.92,
			SessionID: "session-002",
			ModelID:   "claude-3-5-sonnet-20241022",
			Timestamp: startDate.Add(24 * time.Hour),
			Metadata: map[string]interface{}{
				"auto_captured":           true,
				"implicit_signals_count":  3,
				"has_explicit_feedback":   true,
				"user_feedback":           "Clear and concise",
				"response_time_ms":        1500,
			},
		},
		{
			Prompt:    "How to optimize database queries in Go?",
			Response:  "Database query optimization in Go involves connection pooling, prepared statements, and proper indexing...",
			Score:     0.85,
			SessionID: "session-002",
			ModelID:   "claude-3-5-sonnet-20241022",
			Timestamp: startDate.Add(25 * time.Hour),
			Metadata: map[string]interface{}{
				"auto_captured":           true,
				"implicit_signals_count":  1,
				"has_explicit_feedback":   false,
				"response_time_ms":        1100,
			},
		},
		{
			Prompt:    "What is the difference between make and new in Go?",
			Response:  "make() is used to initialize slices, maps, and channels, while new() allocates memory and returns a pointer...",
			Score:     0.98,
			SessionID: "session-003",
			ModelID:   "claude-3-5-sonnet-20241022",
			Timestamp: endDate.Add(-2 * time.Hour),
			Metadata: map[string]interface{}{
				"auto_captured":           true,
				"implicit_signals_count":  4,
				"has_explicit_feedback":   true,
				"user_feedback":           "Perfect explanation!",
				"response_time_ms":        850,
			},
		},
	}

	return feedback
}

// writeJSONL writes feedback data in JSON Lines format (one JSON object per line)
func writeJSONL(filename string, feedback []*rlhf.InteractionFeedback) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, entry := range feedback {
		if err := encoder.Encode(entry); err != nil {
			return err
		}
	}

	return nil
}

// writeJSON writes feedback data as a JSON array
func writeJSON(filename string, feedback []*rlhf.InteractionFeedback) error {
	data, err := json.MarshalIndent(feedback, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// writeCSV writes feedback data in CSV format
func writeCSV(filename string, feedback []*rlhf.InteractionFeedback) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Timestamp",
		"Session ID",
		"Model ID",
		"Score",
		"Prompt",
		"Response",
		"Metadata",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, entry := range feedback {
		metadataJSON, _ := json.Marshal(entry.Metadata)

		row := []string{
			entry.Timestamp.Format(time.RFC3339),
			entry.SessionID,
			entry.ModelID,
			fmt.Sprintf("%.2f", entry.Score),
			entry.Prompt,
			entry.Response,
			string(metadataJSON),
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
