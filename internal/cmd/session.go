package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/AINative-studio/ainative-code/internal/session"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	sessionListAll   bool
	sessionLimit     int
	exportFormat     string
	exportOutput     string
	exportTemplate   string
	searchLimit      int
	searchDateFrom   string
	searchDateTo     string
	searchProvider   string
	searchOutputJSON bool
	// Create command flags
	createTitle      string
	createTags       string
	createProvider   string
	createModel      string
	createMetadata   string
	createNoActivate bool
)

// sessionCmd represents the session command
var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage chat sessions",
	Long: `Manage chat sessions including creating, listing, viewing, and deleting sessions.

Sessions store conversation history and allow you to continue previous conversations.
Each session is identified by a unique ID and can be resumed using the chat command.

Examples:
  # Create a new session
  ainative-code session create --title "Bug Investigation"

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
	Use:     "list",
	Short:   "List chat sessions",
	Long:    `List all chat sessions or recent sessions.`,
	Aliases: []string{"ls", "l"},
	RunE:    runSessionList,
}

// sessionShowCmd represents the session show command
var sessionShowCmd = &cobra.Command{
	Use:     "show [session-id]",
	Short:   "Show session details",
	Long:    `Display detailed information about a specific session including messages.`,
	Aliases: []string{"view", "get"},
	Args:    cobra.ExactArgs(1),
	RunE:    runSessionShow,
}

// sessionDeleteCmd represents the session delete command
var sessionDeleteCmd = &cobra.Command{
	Use:     "delete [session-id]",
	Short:   "Delete a session",
	Long:    `Delete a session and all associated messages.`,
	Aliases: []string{"rm", "remove"},
	Args:    cobra.ExactArgs(1),
	RunE:    runSessionDelete,
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
	Args: cobra.ExactArgs(1),
	RunE: runSessionSearch,
}

// sessionCreateCmd represents the session create command
var sessionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new chat session",
	Long: `Create a new chat session with specified configuration.

The create command allows you to create a new session with custom settings
including title, tags, provider, model, and metadata. By default, the newly
created session will be activated for immediate use.

Examples:
  # Create a session with a title
  ainative-code session create --title "Bug Investigation"

  # Create a session with tags
  ainative-code session create --title "API Development" --tags "golang,api,rest"

  # Create a session with specific provider and model
  ainative-code session create --title "Code Review" --provider anthropic --model claude-3-5-sonnet-20241022

  # Create a session with custom metadata
  ainative-code session create --title "Project Planning" --metadata '{"project":"myapp","priority":"high"}'

  # Create a session without activating it
  ainative-code session create --title "Draft Session" --no-activate`,
	RunE: runSessionCreate,
}

func init() {
	rootCmd.AddCommand(sessionCmd)

	// Add subcommands
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionShowCmd)
	sessionCmd.AddCommand(sessionDeleteCmd)
	sessionCmd.AddCommand(sessionExportCmd)
	sessionCmd.AddCommand(sessionSearchCmd)
	sessionCmd.AddCommand(sessionCreateCmd)

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

	// Session create flags
	sessionCreateCmd.Flags().StringVarP(&createTitle, "title", "t", "", "session title (required)")
	sessionCreateCmd.MarkFlagRequired("title")
	sessionCreateCmd.Flags().StringVar(&createTags, "tags", "", "comma-separated list of tags")
	sessionCreateCmd.Flags().StringVarP(&createProvider, "provider", "p", "", "AI provider name (e.g., anthropic, openai)")
	sessionCreateCmd.Flags().StringVarP(&createModel, "model", "m", "", "model name (e.g., claude-3-5-sonnet-20241022)")
	sessionCreateCmd.Flags().StringVar(&createMetadata, "metadata", "", "JSON metadata string")
	sessionCreateCmd.Flags().BoolVar(&createNoActivate, "no-activate", false, "do not activate the session after creation")
}

func runSessionList(cmd *cobra.Command, args []string) error {
	logger.DebugEvent().
		Bool("all", sessionListAll).
		Int("limit", sessionLimit).
		Msg("Listing sessions")

	// Initialize database connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := getDatabase()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create session manager
	mgr := session.NewSQLiteManager(db)

	// Build list options
	var opts []session.ListOption
	if !sessionListAll {
		opts = append(opts, session.WithLimit(int64(sessionLimit)))
	}

	sessions, err := mgr.ListSessions(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found.")
		fmt.Println("\nCreate a new session with:")
		fmt.Println("  ainative-code session create --title \"My Session\"")
		return nil
	}

	// Display sessions in a table format
	fmt.Printf("\nFound %d session(s):\n\n", len(sessions))

	// Use color codes for better readability
	const (
		colorReset  = "\033[0m"
		colorCyan   = "\033[36m"
		colorYellow = "\033[33m"
		colorGreen  = "\033[32m"
		colorGray   = "\033[90m"
		colorBold   = "\033[1m"
	)

	for i, sess := range sessions {
		fmt.Printf("%s%d.%s %s%s%s\n",
			colorBold, i+1, colorReset,
			colorCyan, sess.Name, colorReset)

		fmt.Printf("   %sID:%s %s\n",
			colorGray, colorReset, sess.ID)

		if sess.Model != nil && *sess.Model != "" {
			fmt.Printf("   %sModel:%s %s\n",
				colorGray, colorReset, *sess.Model)
		}

		fmt.Printf("   %sCreated:%s %s | %sStatus:%s %s\n",
			colorGray, colorReset, sess.CreatedAt.Format("2006-01-02 15:04"),
			colorGray, colorReset, sess.Status)

		if i < len(sessions)-1 {
			fmt.Println()
		}
	}

	if !sessionListAll && len(sessions) == sessionLimit {
		fmt.Printf("\n%sShowing %d sessions. Use --all to see all sessions.%s\n",
			colorGray, sessionLimit, colorReset)
	}

	return nil
}

func runSessionShow(cmd *cobra.Command, args []string) error {
	sessionID := args[0]
	logger.DebugEvent().Str("session_id", sessionID).Msg("Showing session")

	// Initialize database connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := getDatabase()
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

	// Display session details
	displaySessionDetails(sess, messages)

	return nil
}

func displaySessionDetails(sess *session.Session, messages []*session.Message) {
	// Color codes for better readability
	const (
		colorReset  = "\033[0m"
		colorCyan   = "\033[36m"
		colorYellow = "\033[33m"
		colorGreen  = "\033[32m"
		colorGray   = "\033[90m"
		colorBold   = "\033[1m"
		colorBlue   = "\033[34m"
	)

	fmt.Printf("\n%s=== Session Details ===%s\n\n", colorBold, colorReset)

	// Session metadata
	fmt.Printf("%sSession ID:%s %s\n", colorGray, colorReset, sess.ID)
	fmt.Printf("%sTitle:%s %s%s%s\n", colorGray, colorReset, colorCyan, sess.Name, colorReset)
	fmt.Printf("%sStatus:%s %s\n", colorGray, colorReset, sess.Status)
	fmt.Printf("%sCreated:%s %s\n", colorGray, colorReset, sess.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("%sUpdated:%s %s\n", colorGray, colorReset, sess.UpdatedAt.Format("2006-01-02 15:04:05"))

	if sess.Model != nil && *sess.Model != "" {
		fmt.Printf("%sModel:%s %s\n", colorGray, colorReset, *sess.Model)
	}

	if sess.Temperature != nil {
		fmt.Printf("%sTemperature:%s %.2f\n", colorGray, colorReset, *sess.Temperature)
	}

	if sess.MaxTokens != nil {
		fmt.Printf("%sMax Tokens:%s %d\n", colorGray, colorReset, *sess.MaxTokens)
	}

	// Statistics
	fmt.Printf("\n%s=== Statistics ===%s\n\n", colorBold, colorReset)
	fmt.Printf("%sTotal Messages:%s %d\n", colorGray, colorReset, len(messages))

	// Count messages by role
	roleCounts := make(map[string]int)
	var totalTokens int64
	for _, msg := range messages {
		roleCounts[string(msg.Role)]++
		if msg.TokensUsed != nil {
			totalTokens += *msg.TokensUsed
		}
	}

	if userCount := roleCounts["user"]; userCount > 0 {
		fmt.Printf("%sUser Messages:%s %d\n", colorGray, colorReset, userCount)
	}
	if assistantCount := roleCounts["assistant"]; assistantCount > 0 {
		fmt.Printf("%sAssistant Messages:%s %d\n", colorGray, colorReset, assistantCount)
	}
	if systemCount := roleCounts["system"]; systemCount > 0 {
		fmt.Printf("%sSystem Messages:%s %d\n", colorGray, colorReset, systemCount)
	}
	if toolCount := roleCounts["tool"]; toolCount > 0 {
		fmt.Printf("%sTool Messages:%s %d\n", colorGray, colorReset, toolCount)
	}

	if totalTokens > 0 {
		fmt.Printf("%sTotal Tokens:%s %d\n", colorGray, colorReset, totalTokens)
	}

	// Messages
	if len(messages) > 0 {
		fmt.Printf("\n%s=== Messages ===%s\n\n", colorBold, colorReset)

		for i, msg := range messages {
			// Role header with color
			roleColor := colorGreen
			if msg.Role == "user" {
				roleColor = colorBlue
			} else if msg.Role == "system" {
				roleColor = colorYellow
			}

			fmt.Printf("%s%d. [%s%s%s] %s%s\n",
				colorBold, i+1,
				roleColor, strings.ToUpper(string(msg.Role)), colorReset,
				colorGray, msg.Timestamp.Format("2006-01-02 15:04:05"))

			// Message metadata
			if msg.Model != nil && *msg.Model != "" {
				fmt.Printf("   %sModel:%s %s\n", colorGray, colorReset, *msg.Model)
			}
			if msg.TokensUsed != nil {
				fmt.Printf("   %sTokens:%s %d\n", colorGray, colorReset, *msg.TokensUsed)
			}
			if msg.FinishReason != nil && *msg.FinishReason != "" {
				fmt.Printf("   %sFinish Reason:%s %s\n", colorGray, colorReset, *msg.FinishReason)
			}

			// Message content
			fmt.Printf("\n")
			content := msg.Content
			// Truncate very long messages
			maxContentLength := 1000
			if len(content) > maxContentLength {
				content = content[:maxContentLength] + fmt.Sprintf("\n\n%s... (truncated, %d more characters)%s",
					colorGray, len(msg.Content)-maxContentLength, colorReset)
			}

			// Indent content
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				fmt.Printf("   %s\n", line)
			}

			if i < len(messages)-1 {
				fmt.Printf("\n%s%s%s\n\n", colorGray, strings.Repeat("-", 80), colorReset)
			}
		}
	} else {
		fmt.Printf("\n%sNo messages in this session yet.%s\n", colorGray, colorReset)
	}

	fmt.Println()
}

func runSessionDelete(cmd *cobra.Command, args []string) error {
	sessionID := args[0]
	logger.DebugEvent().Str("session_id", sessionID).Msg("Deleting session")

	// Initialize database connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := getDatabase()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create session manager
	mgr := session.NewSQLiteManager(db)

	// Verify session exists before attempting deletion
	sess, err := mgr.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Display session info and request confirmation
	fmt.Printf("\nYou are about to delete the following session:\n")
	fmt.Printf("  ID: %s\n", sess.ID)
	fmt.Printf("  Title: %s\n", sess.Name)
	fmt.Printf("  Status: %s\n", sess.Status)
	fmt.Printf("  Created: %s\n", sess.CreatedAt.Format("2006-01-02 15:04"))

	// Get message count to show user what will be deleted
	messageCount, err := mgr.GetSessionMessageCount(ctx, sessionID)
	if err == nil && messageCount > 0 {
		fmt.Printf("  Messages: %d\n", messageCount)
	}

	fmt.Printf("\nThis will permanently delete the session and all its messages.\n")
	fmt.Printf("Are you sure you want to continue? (y/N): ")

	// Read user confirmation
	var response string
	fmt.Scanln(&response)

	// Check for positive confirmation
	if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
		fmt.Println("\nDeletion cancelled.")
		logger.InfoEvent().
			Str("session_id", sessionID).
			Msg("Session deletion cancelled by user")
		return nil
	}

	// Perform hard delete (permanent deletion)
	if err := mgr.HardDeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	fmt.Printf("\nSession '%s' deleted successfully.\n", sess.Name)

	logger.InfoEvent().
		Str("session_id", sessionID).
		Str("title", sess.Name).
		Int64("messages_deleted", messageCount).
		Msg("Session deleted successfully")

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

	db, err := getDatabase()
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
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return fmt.Errorf("search query cannot be empty. Usage: ainative-code session search <query>")
	}

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

	db, err := getDatabase()
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

func runSessionCreate(cmd *cobra.Command, args []string) error {
	// Validate required title
	title := strings.TrimSpace(createTitle)
	if title == "" {
		return fmt.Errorf("session title cannot be empty")
	}

	logger.InfoEvent().
		Str("title", title).
		Str("tags", createTags).
		Str("provider", createProvider).
		Str("model", createModel).
		Bool("no_activate", createNoActivate).
		Msg("Creating new session")

	// Parse tags if provided
	var tags []string
	if createTags != "" {
		tags = strings.Split(createTags, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
		// Remove empty tags
		var validTags []string
		for _, tag := range tags {
			if tag != "" {
				validTags = append(validTags, tag)
			}
		}
		tags = validTags
	}

	// Parse metadata if provided
	var metadata map[string]interface{}
	if createMetadata != "" {
		if err := json.Unmarshal([]byte(createMetadata), &metadata); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
	}

	// Add tags to metadata if provided
	if len(tags) > 0 {
		if metadata == nil {
			metadata = make(map[string]interface{})
		}
		metadata["tags"] = tags
	}

	// Validate provider if specified
	if createProvider != "" {
		provider := strings.ToLower(strings.TrimSpace(createProvider))
		validProviders := []string{"anthropic", "openai", "azure", "bedrock", "gemini", "ollama", "meta"}
		isValid := false
		for _, vp := range validProviders {
			if provider == vp {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid provider: %s (valid options: %s)",
				createProvider, strings.Join(validProviders, ", "))
		}
	}

	// Initialize database connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := getDatabase()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create session manager
	mgr := session.NewSQLiteManager(db)

	// Generate unique session ID
	sessionID := uuid.New().String()

	// Create session object
	sess := &session.Session{
		ID:        sessionID,
		Name:      title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    session.StatusActive,
		Settings:  metadata,
	}

	// Add provider to model if specified
	if createProvider != "" {
		provider := strings.ToLower(strings.TrimSpace(createProvider))
		sess.Model = &provider
	}

	// Override with specific model if provided
	if createModel != "" {
		model := strings.TrimSpace(createModel)
		sess.Model = &model
	}

	// Create the session in database
	if err := mgr.CreateSession(ctx, sess); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Output success message
	fmt.Printf("\nSession created successfully!\n")
	fmt.Printf("  ID: %s\n", sessionID)
	fmt.Printf("  Title: %s\n", title)

	if len(tags) > 0 {
		fmt.Printf("  Tags: %s\n", strings.Join(tags, ", "))
	}

	if sess.Model != nil {
		fmt.Printf("  Model: %s\n", *sess.Model)
	}

	fmt.Printf("  Status: %s\n", sess.Status)
	fmt.Printf("  Created: %s\n", sess.CreatedAt.Format(time.RFC3339))

	// Activate session unless --no-activate flag is set
	if !createNoActivate {
		// Store the active session ID in a config file or environment
		// For now, we'll just output a message
		fmt.Printf("\nSession activated. Use this ID to continue the conversation:\n")
		fmt.Printf("  ainative-code chat --session-id %s\n", sessionID)

		// TODO: Implement actual session activation by storing session ID
		// in configuration file or environment variable
	}

	logger.InfoEvent().
		Str("session_id", sessionID).
		Str("title", title).
		Bool("activated", !createNoActivate).
		Msg("Session created successfully")

	return nil
}
