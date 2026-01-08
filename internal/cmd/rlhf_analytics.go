package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client/rlhf"
	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	analyticsModelID    string
	analyticsStartDate  string
	analyticsEndDate    string
	analyticsGranularity string
	analyticsExport     string
	analyticsExportFormat string
)

// rlhfAnalyticsCmd represents the rlhf analytics command
var rlhfAnalyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "View RLHF feedback analytics",
	Long: `View analytics and statistics for RLHF feedback data.

Display metrics including:
  - Average feedback score
  - Total interactions
  - Correction rate
  - Feedback distribution
  - Top correction reasons
  - Trending data over time

Analytics can be filtered by model, date range, and granularity.
Results can be exported to CSV or JSON formats.

Examples:
  # View analytics for the last 7 days
  ainative-code rlhf analytics \
    --start-date 2026-01-01 \
    --end-date 2026-01-07

  # Filter by specific model
  ainative-code rlhf analytics \
    --model claude-3-5-sonnet-20241022 \
    --start-date 2026-01-01 \
    --end-date 2026-01-07

  # View with daily granularity
  ainative-code rlhf analytics \
    --start-date 2026-01-01 \
    --end-date 2026-01-31 \
    --granularity day

  # Export to CSV
  ainative-code rlhf analytics \
    --start-date 2026-01-01 \
    --end-date 2026-01-07 \
    --export analytics.csv \
    --export-format csv

  # Export to JSON
  ainative-code rlhf analytics \
    --start-date 2026-01-01 \
    --end-date 2026-01-07 \
    --export analytics.json \
    --export-format json`,
	RunE: runRlhfAnalytics,
}

func init() {
	// Analytics flags
	rlhfAnalyticsCmd.Flags().StringVarP(&analyticsModelID, "model", "m", "", "filter by model ID")
	rlhfAnalyticsCmd.Flags().StringVar(&analyticsStartDate, "start-date", "", "start date (YYYY-MM-DD) (required)")
	rlhfAnalyticsCmd.Flags().StringVar(&analyticsEndDate, "end-date", "", "end date (YYYY-MM-DD) (required)")
	rlhfAnalyticsCmd.Flags().StringVarP(&analyticsGranularity, "granularity", "g", "day", "granularity (hour, day, week, month)")
	rlhfAnalyticsCmd.Flags().StringVarP(&analyticsExport, "export", "e", "", "export to file")
	rlhfAnalyticsCmd.Flags().StringVarP(&analyticsExportFormat, "export-format", "f", "json", "export format (csv, json)")
	rlhfAnalyticsCmd.Flags().BoolP("json", "j", false, "output as JSON")

	rlhfAnalyticsCmd.MarkFlagRequired("start-date")
	rlhfAnalyticsCmd.MarkFlagRequired("end-date")
}

func runRlhfAnalytics(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Parse dates
	startDate, err := time.Parse("2006-01-02", analyticsStartDate)
	if err != nil {
		return fmt.Errorf("invalid start-date format (use YYYY-MM-DD): %w", err)
	}

	endDate, err := time.Parse("2006-01-02", analyticsEndDate)
	if err != nil {
		return fmt.Errorf("invalid end-date format (use YYYY-MM-DD): %w", err)
	}

	// Initialize RLHF client
	rlhfClient, err := initRlhfClient()
	if err != nil {
		return fmt.Errorf("failed to initialize RLHF client: %w", err)
	}

	// Create analytics request
	req := &rlhf.AnalyticsRequest{
		ModelID:     analyticsModelID,
		StartDate:   startDate,
		EndDate:     endDate,
		Granularity: analyticsGranularity,
	}

	logger.InfoEvent().
		Str("model_id", req.ModelID).
		Time("start_date", req.StartDate).
		Time("end_date", req.EndDate).
		Str("granularity", req.Granularity).
		Msg("Fetching RLHF analytics")

	// Get analytics
	analytics, err := rlhfClient.GetAnalytics(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get analytics: %w", err)
	}

	// Handle export if requested
	if analyticsExport != "" {
		return exportAnalytics(analytics, analyticsExport, analyticsExportFormat)
	}

	// Output analytics
	if jsonOutput {
		output, _ := json.MarshalIndent(analytics, "", "  ")
		fmt.Println(string(output))
	} else {
		displayAnalytics(analytics)
	}

	return nil
}

// displayAnalytics shows analytics in a human-readable format
func displayAnalytics(analytics *rlhf.Analytics) {
	bold := color.New(color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Println()
	fmt.Println(bold("RLHF Analytics Report"))
	fmt.Println(strings.Repeat("=", 80))

	// Overview section
	fmt.Println()
	fmt.Println(bold("Overview:"))
	fmt.Printf("  Model: %s\n", analytics.ModelID)
	fmt.Printf("  Period: %s to %s\n",
		analytics.StartDate.Format("2006-01-02"),
		analytics.EndDate.Format("2006-01-02"))
	fmt.Println()

	// Key metrics
	fmt.Println(bold("Key Metrics:"))

	scoreColor := green
	if analytics.AverageFeedbackScore < 0.7 {
		scoreColor = yellow
	}
	if analytics.AverageFeedbackScore < 0.5 {
		scoreColor = red
	}

	fmt.Printf("  Average Feedback Score: %s\n", scoreColor(fmt.Sprintf("%.2f / 1.00", analytics.AverageFeedbackScore)))
	fmt.Printf("  Total Interactions: %s\n", green(analytics.TotalInteractions))
	fmt.Printf("  Total Corrections: %d\n", analytics.TotalCorrections)

	corrRateColor := green
	if analytics.CorrectionRate > 10.0 {
		corrRateColor = yellow
	}
	if analytics.CorrectionRate > 20.0 {
		corrRateColor = red
	}
	fmt.Printf("  Correction Rate: %s\n", corrRateColor(fmt.Sprintf("%.1f%%", analytics.CorrectionRate)))
	fmt.Println()

	// Score distribution
	if len(analytics.ScoreDistribution) > 0 {
		fmt.Println(bold("Score Distribution:"))
		displayScoreDistribution(analytics.ScoreDistribution, analytics.TotalInteractions)
		fmt.Println()
	}

	// Top correction reasons
	if len(analytics.TopCorrectionReasons) > 0 {
		fmt.Println(bold("Top Correction Reasons:"))
		for i, reason := range analytics.TopCorrectionReasons {
			percentage := 0.0
			if analytics.TotalCorrections > 0 {
				percentage = float64(reason.Count) / float64(analytics.TotalCorrections) * 100
			}
			fmt.Printf("  %d. %s (%d corrections, %.1f%%)\n",
				i+1, reason.Reason, reason.Count, percentage)
		}
		fmt.Println()
	}

	// Trending data
	if len(analytics.TrendingData) > 0 {
		fmt.Println(bold("Trending Data:"))
		displayTrendChart(analytics.TrendingData)
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 80))
}

// displayScoreDistribution shows a bar chart of score distribution
func displayScoreDistribution(distribution map[string]int, total int) {
	// Sort buckets
	buckets := []string{"0.0-0.2", "0.2-0.4", "0.4-0.6", "0.6-0.8", "0.8-1.0"}

	maxCount := 0
	for _, count := range distribution {
		if count > maxCount {
			maxCount = count
		}
	}

	for _, bucket := range buckets {
		count := distribution[bucket]
		percentage := 0.0
		if total > 0 {
			percentage = float64(count) / float64(total) * 100
		}

		// Create bar
		barLength := 40
		filled := 0
		if maxCount > 0 {
			filled = int(float64(count) / float64(maxCount) * float64(barLength))
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barLength-filled)

		// Color based on score range
		var barColor func(...interface{}) string
		if strings.HasPrefix(bucket, "0.8") || strings.HasPrefix(bucket, "0.6") {
			barColor = color.New(color.FgGreen).SprintFunc()
		} else if strings.HasPrefix(bucket, "0.4") {
			barColor = color.New(color.FgYellow).SprintFunc()
		} else {
			barColor = color.New(color.FgRed).SprintFunc()
		}

		fmt.Printf("  %s │%s│ %4d (%.1f%%)\n",
			bucket,
			barColor(bar),
			count,
			percentage,
		)
	}
}

// displayTrendChart shows an ASCII chart of trending data
func displayTrendChart(trends []rlhf.TrendPoint) {
	if len(trends) == 0 {
		return
	}

	// Find min/max scores for scaling
	minScore := 1.0
	maxScore := 0.0
	for _, point := range trends {
		if point.Score < minScore {
			minScore = point.Score
		}
		if point.Score > maxScore {
			maxScore = point.Score
		}
	}

	// Display chart
	chartHeight := 10
	for i := chartHeight; i >= 0; i-- {
		scoreAtLevel := minScore + (maxScore-minScore)*float64(i)/float64(chartHeight)
		fmt.Printf("  %.2f │", scoreAtLevel)

		for _, point := range trends {
			if point.Score >= scoreAtLevel {
				fmt.Print("█")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}

	// X-axis
	fmt.Print("       └")
	fmt.Print(strings.Repeat("─", len(trends)))
	fmt.Println()

	// Dates
	fmt.Print("        ")
	if len(trends) > 0 {
		fmt.Printf("%s", trends[0].Date.Format("01/02"))
		if len(trends) > 1 {
			fmt.Print(strings.Repeat(" ", len(trends)-11))
			fmt.Printf("%s", trends[len(trends)-1].Date.Format("01/02"))
		}
	}
	fmt.Println()
}

// exportAnalytics exports analytics to a file
func exportAnalytics(analytics *rlhf.Analytics, filename, format string) error {
	logger.InfoEvent().
		Str("filename", filename).
		Str("format", format).
		Msg("Exporting analytics")

	switch format {
	case "json":
		return exportAnalyticsJSON(analytics, filename)
	case "csv":
		return exportAnalyticsCSV(analytics, filename)
	default:
		return fmt.Errorf("unsupported export format: %s (use 'json' or 'csv')", format)
	}
}

// exportAnalyticsJSON exports analytics to JSON file
func exportAnalyticsJSON(analytics *rlhf.Analytics, filename string) error {
	data, err := json.MarshalIndent(analytics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analytics: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✓ Analytics exported to %s (JSON format)\n", filename)
	return nil
}

// exportAnalyticsCSV exports analytics to CSV file
func exportAnalyticsCSV(analytics *rlhf.Analytics, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Model ID",
		"Start Date",
		"End Date",
		"Average Score",
		"Total Interactions",
		"Total Corrections",
		"Correction Rate",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data
	row := []string{
		analytics.ModelID,
		analytics.StartDate.Format("2006-01-02"),
		analytics.EndDate.Format("2006-01-02"),
		fmt.Sprintf("%.2f", analytics.AverageFeedbackScore),
		fmt.Sprintf("%d", analytics.TotalInteractions),
		fmt.Sprintf("%d", analytics.TotalCorrections),
		fmt.Sprintf("%.2f%%", analytics.CorrectionRate),
	}
	if err := writer.Write(row); err != nil {
		return fmt.Errorf("failed to write row: %w", err)
	}

	// Write score distribution
	if len(analytics.ScoreDistribution) > 0 {
		writer.Write([]string{}) // Empty line
		writer.Write([]string{"Score Range", "Count"})

		buckets := []string{"0.0-0.2", "0.2-0.4", "0.4-0.6", "0.6-0.8", "0.8-1.0"}
		for _, bucket := range buckets {
			count := analytics.ScoreDistribution[bucket]
			writer.Write([]string{bucket, fmt.Sprintf("%d", count)})
		}
	}

	// Write top correction reasons
	if len(analytics.TopCorrectionReasons) > 0 {
		writer.Write([]string{}) // Empty line
		writer.Write([]string{"Correction Reason", "Count"})

		for _, reason := range analytics.TopCorrectionReasons {
			writer.Write([]string{reason.Reason, fmt.Sprintf("%d", reason.Count)})
		}
	}

	// Write trending data
	if len(analytics.TrendingData) > 0 {
		writer.Write([]string{}) // Empty line
		writer.Write([]string{"Date", "Score", "Count"})

		for _, point := range analytics.TrendingData {
			writer.Write([]string{
				point.Date.Format("2006-01-02"),
				fmt.Sprintf("%.2f", point.Score),
				fmt.Sprintf("%d", point.Count),
			})
		}
	}

	fmt.Printf("✓ Analytics exported to %s (CSV format)\n", filename)
	return nil
}
