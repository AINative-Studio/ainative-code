package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/AINative-studio/ainative-code/internal/client"
	designclient "github.com/AINative-studio/ainative-code/internal/client/design"
	"github.com/AINative-studio/ainative-code/internal/design"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

var (
	syncProjectID     string
	syncWatch         bool
	syncDirection     string
	syncLocalPath     string
	syncConflict      string
	syncDryRun        bool
	syncVerbose       bool
	syncWatchInterval time.Duration
)

// designSyncCmd represents the design sync command
var designSyncCommand = &cobra.Command{
	Use:   "sync",
	Short: "Bidirectional design token synchronization",
	Long: `Synchronize design tokens between local files and AINative Design system.

This command provides bidirectional synchronization with conflict detection and resolution:
  - Pull: Download tokens from AINative Design to local file
  - Push: Upload tokens from local file to AINative Design
  - Bidirectional: Sync in both directions with conflict resolution

Watch mode enables continuous synchronization, automatically detecting and syncing changes.

Examples:
  # Pull tokens from AINative Design
  ainative-code design sync --project my-project --direction pull

  # Push tokens to AINative Design
  ainative-code design sync --project my-project --direction push

  # Bidirectional sync with conflict resolution
  ainative-code design sync --project my-project --direction bidirectional --conflict remote

  # Watch mode for continuous sync
  ainative-code design sync --project my-project --watch --local-path ./tokens.json

  # Dry run to preview changes
  ainative-code design sync --project my-project --dry-run`,
	RunE: runDesignSyncCommand,
}

func init() {
	designCmd.AddCommand(designSyncCommand)

	// Required flags
	designSyncCommand.Flags().StringVarP(&syncProjectID, "project", "p", "", "Project ID for synchronization (required)")
	designSyncCommand.MarkFlagRequired("project")

	// Optional flags
	designSyncCommand.Flags().BoolVarP(&syncWatch, "watch", "w", false, "Enable watch mode for continuous sync")
	designSyncCommand.Flags().StringVarP(&syncDirection, "direction", "d", "bidirectional", "Sync direction: pull, push, or bidirectional")
	designSyncCommand.Flags().StringVarP(&syncLocalPath, "local-path", "l", "./design-tokens.json", "Local file path for tokens")
	designSyncCommand.Flags().StringVarP(&syncConflict, "conflict", "c", "prompt", "Conflict resolution strategy: local, remote, newest, prompt, merge")
	designSyncCommand.Flags().BoolVar(&syncDryRun, "dry-run", false, "Perform dry run without making changes")
	designSyncCommand.Flags().BoolVarP(&syncVerbose, "verbose", "v", false, "Enable verbose logging")
	designSyncCommand.Flags().DurationVar(&syncWatchInterval, "watch-interval", 2*time.Second, "Debounce interval for watch mode")
}

func runDesignSyncCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate direction
	var direction design.SyncDirection
	switch syncDirection {
	case "pull":
		direction = design.SyncDirectionPull
	case "push":
		direction = design.SyncDirectionPush
	case "bidirectional", "both":
		direction = design.SyncDirectionBidirectional
	default:
		return fmt.Errorf("invalid sync direction: %s (must be pull, push, or bidirectional)", syncDirection)
	}

	// Validate conflict resolution strategy
	var conflictStrategy design.ConflictResolutionStrategy
	switch syncConflict {
	case "local":
		conflictStrategy = design.ConflictResolutionLocal
	case "remote":
		conflictStrategy = design.ConflictResolutionRemote
	case "newest":
		conflictStrategy = design.ConflictResolutionNewest
	case "prompt":
		conflictStrategy = design.ConflictResolutionPrompt
	case "merge":
		conflictStrategy = design.ConflictResolutionMerge
	default:
		return fmt.Errorf("invalid conflict resolution strategy: %s", syncConflict)
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(syncLocalPath)
	if err != nil {
		return fmt.Errorf("failed to resolve local path: %w", err)
	}

	logger.InfoEvent().
		Str("project_id", syncProjectID).
		Str("direction", syncDirection).
		Str("local_path", absPath).
		Str("conflict_strategy", syncConflict).
		Bool("watch", syncWatch).
		Bool("dry_run", syncDryRun).
		Msg("Starting design token sync")

	// Create API client
	apiClient := client.New()

	// Create design client
	designClient := designclient.New(
		designclient.WithAPIClient(apiClient),
		designclient.WithProjectID(syncProjectID),
	)

	// Create sync adapter
	adapter := designclient.NewSyncAdapter(designClient, syncProjectID)

	// Create sync configuration
	syncConfig := design.SyncConfig{
		ProjectID:          syncProjectID,
		Direction:          direction,
		WatchMode:          syncWatch,
		WatchInterval:      syncWatchInterval,
		LocalPath:          absPath,
		ConflictResolution: conflictStrategy,
		DryRun:             syncDryRun,
		Verbose:            syncVerbose,
	}

	// Create syncer
	syncer := design.NewSyncer(adapter, syncConfig)

	if syncWatch {
		return runWatchMode(ctx, syncer, absPath)
	}

	return runSingleSync(ctx, syncer)
}

// runSingleSync performs a one-time synchronization.
func runSingleSync(ctx context.Context, syncer *design.Syncer) error {
	fmt.Println("Starting synchronization...")

	result, err := syncer.Sync(ctx)
	if err != nil {
		return fmt.Errorf("synchronization failed: %w", err)
	}

	// Print results
	printSyncResult(result)

	// Print conflict summary if any
	if len(result.Conflicts) > 0 {
		design.PrintConflictSummary(result.Conflicts)
	}

	if result.DryRun {
		fmt.Println("\nDry run completed - no changes were made")
	} else {
		fmt.Println("\nSynchronization completed successfully")
	}

	return nil
}

// runWatchMode runs continuous synchronization in watch mode.
func runWatchMode(ctx context.Context, syncer *design.Syncer, localPath string) error {
	fmt.Println("Starting watch mode...")
	fmt.Printf("Watching: %s\n", localPath)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create cancellable context
	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create watcher configuration
	watchConfig := design.WatchConfig{
		Paths:            []string{filepath.Dir(localPath)},
		DebounceDuration: syncWatchInterval,
		SyncOnStart:      true,
		MaxRetries:       3,
		RetryDelay:       5 * time.Second,
	}

	// Create watcher
	watcher, err := design.NewWatcher(syncer, watchConfig)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	// Start watcher in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := watcher.Start(watchCtx); err != nil {
			errChan <- err
		}
	}()

	// Wait for interrupt signal or error
	select {
	case <-sigChan:
		fmt.Println("\nReceived interrupt signal, stopping watcher...")
		cancel()
		watcher.Stop()
	case err := <-errChan:
		return fmt.Errorf("watcher error: %w", err)
	}

	fmt.Println("Watch mode stopped")
	return nil
}

// printSyncResult prints a formatted summary of the sync result.
func printSyncResult(result *design.SyncResult) {
	fmt.Println("\n=== Sync Results ===")
	fmt.Printf("Duration: %v\n", result.Duration.Round(time.Millisecond))
	fmt.Printf("Added:    %d\n", result.Added)
	fmt.Printf("Updated:  %d\n", result.Updated)
	fmt.Printf("Deleted:  %d\n", result.Deleted)
	fmt.Printf("Conflicts: %d\n", len(result.Conflicts))

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors encountered:\n")
		for i, err := range result.Errors {
			fmt.Printf("  %d. %v\n", i+1, err)
		}
	}
}
