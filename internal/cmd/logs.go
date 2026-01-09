package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

var (
	logsTail   int
	logsFollow bool
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View application logs",
	Long: `View application logs from the project-relative log file.

By default, shows the last 1000 lines of logs. Use --tail to limit the number
of lines, or --follow to stream logs in real-time (like 'tail -f').

Examples:
  # Show last 1000 lines
  ainative-code logs

  # Show last 500 lines
  ainative-code logs --tail 500

  # Follow logs in real-time
  ainative-code logs --follow

  # Combine tail and follow
  ainative-code logs --tail 100 --follow

Logs are stored in: ./.ainative-code/logs/ainative-code.log`,
	RunE: runLogs,
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().IntVarP(&logsTail, "tail", "n", 1000, "number of lines to show")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "follow log output (like tail -f)")
}

func runLogs(cmd *cobra.Command, args []string) error {
	// Get log file path
	logPath := getProjectLogPath()

	// Check if log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return fmt.Errorf("log file not found: %s\n\nLogs may not have been written yet. Try running a command first.", logPath)
	}

	// Open log file
	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	if logsFollow {
		// Follow mode: stream logs in real-time
		return followLogs(file, logsTail)
	}

	// Tail mode: show last N lines
	return tailLogs(file, logsTail)
}

// tailLogs shows the last N lines of the log file
func tailLogs(file *os.File, n int) error {
	// Read all lines
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	// Get last N lines
	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}

	// Print lines
	for _, line := range lines[start:] {
		fmt.Println(line)
	}

	return nil
}

// followLogs streams log file in real-time (like tail -f)
func followLogs(file *os.File, initialLines int) error {
	// First, show the last N lines
	if err := tailLogs(file, initialLines); err != nil {
		return err
	}

	// Then follow new content
	fmt.Println("\n--- Following logs (Ctrl+C to stop) ---")

	// Seek to end of file
	if _, err := file.Seek(0, 2); err != nil {
		return fmt.Errorf("failed to seek to end of file: %w", err)
	}

	scanner := bufio.NewScanner(file)
	for {
		// Try to read new lines
		if scanner.Scan() {
			fmt.Println(scanner.Text())
		} else {
			// No new content, wait a bit
			time.Sleep(100 * time.Millisecond)

			// Check for errors
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("failed to read log file: %w", err)
			}
		}
	}
}

// getProjectLogPath returns the path to the project log file
func getProjectLogPath() string {
	// Use project-relative path (like Crush)
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to home directory
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".ainative-code", "logs", "ainative-code.log")
	}

	return filepath.Join(cwd, ".ainative-code", "logs", "ainative-code.log")
}

// InitProjectLogs initializes logging to project-relative log file
func InitProjectLogs() error {
	logPath := getProjectLogPath()
	logDir := filepath.Dir(logPath)

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Initialize logger with file output
	_, err := logger.New(&logger.Config{
		Level:          logger.InfoLevel,
		Format:         logger.JSONFormat,
		Output:         logPath,
		EnableRotation: true,
		MaxSize:        100, // 100 MB
		MaxBackups:     3,
		MaxAge:         7, // days
		Compress:       true,
	})

	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	return nil
}
