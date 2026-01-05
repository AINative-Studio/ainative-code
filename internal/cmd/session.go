package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/AINative-studio/ainative-code/internal/database"
	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/AINative-studio/ainative-code/internal/session"
)

var (
	sessionListAll    bool
	sessionLimit      int
	exportFormat      string
	exportOutput      string
	exportTemplate    string
	searchLimit       int
	searchDateFrom    string
	searchDateTo      string
	searchProvider    string
	searchOutputJSON  bool
)

// sessionCmd represents the session command
var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage chat sessions",
	Long: `Manage chat sessions including listing, viewing, and deleting sessions.

Sessions store conversation history and allow you to continue previous conversations.
Each session is identified by a unique ID and can be resumed using the chat command.

Examples:
  # List recent sessions
  ainative-code session list

  # List all sessions
  ainative-code session list --all

  # View a specific session
  ainative-code session show abc123

  # Delete a session
  ainative-code session delete abc123

  # Export a session
  ainative-code session export abc123 --output session.json`,
	Aliases: []string{"sessions", "sess"},
}

// sessionListCmd represents the session list command
var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List chat sessions",
	Long:  `List all chat sessions or recent sessions.`,
	Aliases: []string{"ls", "l"},
	RunE:  runSessionList,
}

// sessionShowCmd represents the session show command
var sessionShowCmd = &cobra.Command{
	Use:   "show [session-id]",
	Short: "Show session details",
	Long:  `Display detailed information about a specific session including messages.`,
	Aliases: []string{"view", "get"},
	Args:  cobra.ExactArgs(1),
	RunE:  runSessionShow,
}

// sessionDeleteCmd represents the session delete command
var sessionDeleteCmd = &cobra.Command{
	Use:   "delete [session-id]",
	Short: "Delete a session",
	Long:  `Delete a session and all associated messages.`,
	Aliases: []string{"rm", "remove"},
	Args:  cobra.ExactArgs(1),
	RunE:  runSessionDelete,
}

// sessionExportCmd represents the session export command
var sessionExportCmd = &cobra.Command{
	Use:   "export [session-id]",
	Short: "Export a session to various formats",
	Long: `Export a session to various formats including JSON, Markdown, and HTML.

The export command allows you to export conversation sessions to different formats
for backup, sharing, or documentation purposes.

Supported formats:
  - json: Complete session data with metadata (default)
  - markdown: Clean formatted markdown with code blocks
  - html: Styled HTML output with syntax highlighting

Examples:
  # Export to JSON (default)
  ainative-code session export abc123

  # Export to Markdown
  ainative-code session export abc123 --format markdown --output conversation.md

  # Export to HTML with custom output
  ainative-code session export abc123 --format html --output report.html

  # Export using custom template
  ainative-code session export abc123 --template custom.tmpl --output custom.md`,
	Args: cobra.ExactArgs(1),
	RunE: runSessionExport,
}

// sessionSearchCmd represents the session search command
var sessionSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search messages across all sessions",
	Long: `Search for messages across all conversation sessions using full-text search.

This command uses SQLite FTS5 (Full-Text Search) to quickly find messages matching your query.
Results include highlighted snippets and are ranked by relevance using BM25 algorithm.

Features:
  - Full-text search with relevance ranking
  - Context snippets with highlighted matches
  - Date range filtering
  - Provider/model filtering
  - Pagination support

Examples:
  # Search for messages about authentication
  ainative-code session search "authentication"

  # Search with a limit of 10 results
  ainative-code session search "golang" --limit 10

  # Search messages from a specific date range
  ainative-code session search "error" --date-from "2026-01-01" --date-to "2026-01-05"

  # Search only Claude messages
  ainative-code session search "explain" --provider claude

  # Search GPT-4 messages
  ainative-code session search "api" --provider gpt-4

  # Output results as JSON
  ainative-code session search "database" --json`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSessionSearch,
}

func init() {
	rootCmd.AddCommand(sessionCmd)

	// Add subcommands
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionShowCmd)
	sessionCmd.AddCommand(sessionDeleteCmd)
	sessionCmd.AddCommand(sessionExportCmd)
	sessionCmd.AddCommand(sessionSearchCmd)

	// Session list flags
	sessionListCmd.Flags().BoolVarP(&sessionListAll, "all", "a", false, "list all sessions")
	sessionListCmd.Flags().IntVarP(&sessionLimit, "limit", "n", 10, "limit number of sessions to display")

	// Session export flags
	sessionExportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "export format: json, markdown, html")
	sessionExportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "output file path (default: session-<id>.<format>)")
	sessionExportCmd.Flags().StringVarP(&exportTemplate, "template", "t", "", "custom template file path (optional)")

	// Session search flags
	sessionSearchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 50, "maximum number of results to return")
	sessionSearchCmd.Flags().StringVar(&searchDateFrom, "date-from", "", "filter messages from this date (YYYY-MM-DD)")
	sessionSearchCmd.Flags().StringVar(&searchDateTo, "date-to", "", "filter messages until this date (YYYY-MM-DD)")
	sessionSearchCmd.Flags().StringVarP(&searchProvider, "provider", "p", "", "filter by provider/model (e.g., 'claude', 'gpt')")
	sessionSearchCmd.Flags().BoolVar(&searchOutputJSON, "json", false, "output results as JSON")
}

func runSessionList(cmd *cobra.Command, args []string) error {
	logger.DebugEvent().
		Bool("all", sessionListAll).
		Int("limit", sessionLimit).
		Msg("Listing sessions")

	fmt.Println("Session List - Coming soon!")
	fmt.Printf("All: %v, Limit: %d\n", sessionListAll, sessionLimit)

	// TODO: Implement session listing from ZeroDB
	return nil
}

func runSessionShow(cmd *cobra.Command, args []string) error {
	sessionID := args[0]
	logger.DebugEvent().Str("session_id", sessionID).Msg("Showing session")

	fmt.Printf("Session Details for: %s - Coming soon!\n", sessionID)

	// TODO: Implement session detail retrieval from ZeroDB
	return nil
}

func runSessionDelete(cmd *cobra.Command, args []string) error {
	sessionID := args[0]
	logger.DebugEvent().Str("session_id", sessionID).Msg("Deleting session")

	fmt.Printf("Deleting session: %s - Coming soon!\n", sessionID)

	// TODO: Implement session deletion from ZeroDB
	return nil
}

func runSessionExport(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	// Validate and normalize format
	format := strings.ToLower(exportFormat)
	var exportFormatEnum session.ExportFormat

	switch format {
	case "json":
		exportFormatEnum = session.ExportFormatJSON
	case "markdown", "md":
		exportFormatEnum = session.ExportFormatMarkdown
	case "html", "htm":
		exportFormatEnum = session.ExportFormatHTML
	default:
		return fmt.Errorf("invalid format: %s (supported: json, markdown, html)", format)
	}

	// Determine output file path
	output := exportOutput
	if output == "" {
		ext := format
		if format == "md" {
			ext = "markdown"
		} else if format == "htm" {
			ext = "html"
		}
		output = fmt.Sprintf("session-%s.%s", sessionID, ext)
	}

	logger.InfoEvent().
		Str("session_id", sessionID).
		Str("format", format).
		Str("output", output).
		Str("template", exportTemplate).
		Msg("Exporting session")

	// Initialize database connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.Open()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create session manager
	mgr := session.NewSQLiteManager(db)

	// Get session
	sess, err := mgr.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Get messages
	messages, err := mgr.GetMessages(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get messages: %w", err)
	}

	fmt.Printf("Exporting session '%s' (%d messages) to %s...\n", sess.Name, len(messages), output)

	// Create exporter
	exporter := session.NewExporter(&session.ExporterOptions{
		IncludeMetadata: true,
		PrettyPrint:     true,
	})

	// Create output directory if needed
	outputDir := filepath.Dir(output)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Create output file
	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Export based on format or custom template
	var exportErr error

	if exportTemplate != "" {
		// Use custom template
		logger.DebugEvent().
			Str("template", exportTemplate).
			Msg("Using custom template")
		exportErr = exporter.ExportWithTemplate(file, exportTemplate, sess, messages)
	} else {
		// Use built-in format
		switch exportFormatEnum {
		case session.ExportFormatJSON:
			exportErr = exporter.ExportToJSON(file, sess, messages)
		case session.ExportFormatMarkdown:
			exportErr = exporter.ExportToMarkdown(file, sess, messages)
		case session.ExportFormatHTML:
			exportErr = exporter.ExportToHTML(file, sess, messages)
		default:
			exportErr = fmt.Errorf("unsupported format: %s", format)
		}
	}

	if exportErr != nil {
		return fmt.Errorf("export failed: %w", exportErr)
	}

	// Get file info for size
	fileInfo, err := file.Stat()
	if err == nil {
		fmt.Printf("\nExport completed successfully!\n")
		fmt.Printf("  File: %s\n", output)
		fmt.Printf("  Size: %d bytes\n", fileInfo.Size())
		fmt.Printf("  Format: %s\n", format)
		fmt.Printf("  Messages: %d\n", len(messages))

		// Calculate total tokens if available
		var totalTokens int64
		for _, msg := range messages {
			if msg.TokensUsed != nil {
				totalTokens += *msg.TokensUsed
			}
		}
		if totalTokens > 0 {
			fmt.Printf("  Total Tokens: %d\n", totalTokens)
		}
	}

	logger.InfoEvent().
		Str("session_id", sessionID).
		Str("output", output).
		Str("format", format).
		Int("messages", len(messages)).
		Msg("Session exported successfully")

	return nil
}

func runSessionSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	logger.InfoEvent().
		Str("query", query).
		Int("limit", searchLimit).
		Str("date_from", searchDateFrom).
		Str("date_to", searchDateTo).
		Str("provider", searchProvider).
		Bool("json_output", searchOutputJSON).
		Msg("Searching sessions")

	// Initialize database connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.Open()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create session manager
	mgr := session.NewSQLiteManager(db)

	// Build search options
	opts := &session.SearchOptions{
		Query:  query,
		Limit:  int64(searchLimit),
		Offset: 0,
	}

	// Parse date filters if provided
	if searchDateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", searchDateFrom)
		if err != nil {
			return fmt.Errorf("invalid date-from format (use YYYY-MM-DD): %w", err)
		}
		opts.DateFrom = &dateFrom
	}

	if searchDateTo != "" {
		dateTo, err := time.Parse("2006-01-02", searchDateTo)
		if err != nil {
			return fmt.Errorf("invalid date-to format (use YYYY-MM-DD): %w", err)
		}
		// Set to end of day
		dateTo = dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		opts.DateTo = &dateTo
	}

	if searchProvider != "" {
		opts.Provider = searchProvider
	}

	// Execute search
	results, err := mgr.SearchAllMessages(ctx, opts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Output results
	if searchOutputJSON {
		return outputSearchResultsJSON(results)
	}

	return outputSearchResultsTable(results)
}

func outputSearchResultsJSON(results *session.SearchResultSet) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func outputSearchResultsTable(results *session.SearchResultSet) error {
	fmt.Printf("\nSearch Results for: %q\n", results.Query)
	fmt.Printf("Found %d matches (showing %d)\n\n", results.TotalCount, len(results.Results))

	if len(results.Results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	// Use color codes for highlighting
	const (
		colorReset  = "\033[0m"
		colorCyan   = "\033[36m"
		colorYellow = "\033[33m"
		colorGreen  = "\033[32m"
		colorGray   = "\033[90m"
		colorBold   = "\033[1m"
	)

	for i, result := range results.Results {
		fmt.Printf("%s%d.%s %s%s%s\n",
			colorBold, i+1, colorReset,
			colorCyan, result.SessionName, colorReset)

		fmt.Printf("   %sSession:%s %s | %sRole:%s %s | %sScore:%s %.2f\n",
			colorGray, colorReset, result.Message.SessionID[:8]+"...",
			colorGray, colorReset, result.Message.Role,
			colorGray, colorReset, result.RelevanceScore)

		if result.Message.Model != nil {
			fmt.Printf("   %sModel:%s %s\n",
				colorGray, colorReset, *result.Message.Model)
		}

		fmt.Printf("   %sTime:%s %s\n",
			colorGray, colorReset, result.Message.Timestamp.Format("2006-01-02 15:04:05"))

		// Display snippet with highlighted terms
		snippet := result.Snippet
		// Convert HTML marks to terminal colors
		snippet = strings.ReplaceAll(snippet, "<mark>", colorYellow+colorBold)
		snippet = strings.ReplaceAll(snippet, "</mark>", colorReset)

		fmt.Printf("\n   %s%s%s\n\n",
			colorGreen, snippet, colorReset)

		// Separator between results
		if i < len(results.Results)-1 {
			fmt.Println("   " + strings.Repeat("-", 70))
			fmt.Println()
		}
	}

	// Pagination info
	if results.TotalCount > int64(len(results.Results)) {
		remaining := results.TotalCount - int64(len(results.Results))
		fmt.Printf("\n%s%d more results available. Use --limit to see more.%s\n",
			colorGray, remaining, colorReset)
	}

	return nil
}
