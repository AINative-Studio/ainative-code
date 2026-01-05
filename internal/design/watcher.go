package design

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

// Watcher monitors file system changes and triggers synchronization.
type Watcher struct {
	syncer     *Syncer
	fsWatcher  *fsnotify.Watcher
	config     WatchConfig
	mu         sync.Mutex
	stopChan   chan struct{}
	syncQueue  chan WatchEvent
	isRunning  bool
}

// WatchConfig configures the file watcher.
type WatchConfig struct {
	// Paths to watch for changes.
	Paths []string

	// Debounce duration to prevent rapid re-syncs.
	DebounceDuration time.Duration

	// SyncOnStart triggers an initial sync when watch mode starts.
	SyncOnStart bool

	// MaxRetries for sync operations.
	MaxRetries int

	// RetryDelay between sync retries.
	RetryDelay time.Duration
}

// NewWatcher creates a new file system watcher.
func NewWatcher(syncer *Syncer, config WatchConfig) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	if config.DebounceDuration == 0 {
		config.DebounceDuration = 2 * time.Second
	}

	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	if config.RetryDelay == 0 {
		config.RetryDelay = 5 * time.Second
	}

	return &Watcher{
		syncer:    syncer,
		fsWatcher: fsWatcher,
		config:    config,
		stopChan:  make(chan struct{}),
		syncQueue: make(chan WatchEvent, 100),
	}, nil
}

// Start begins watching for file changes.
func (w *Watcher) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.isRunning {
		w.mu.Unlock()
		return fmt.Errorf("watcher is already running")
	}
	w.isRunning = true
	w.mu.Unlock()

	// Add paths to watch
	for _, path := range w.config.Paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			logger.WarnEvent().Err(err).Str("path", path).Msg("Failed to resolve absolute path")
			continue
		}

		if err := w.fsWatcher.Add(absPath); err != nil {
			logger.WarnEvent().Err(err).Str("path", absPath).Msg("Failed to add path to watcher")
			continue
		}

		logger.InfoEvent().Str("path", absPath).Msg("Watching path for changes")
	}

	// Perform initial sync if configured
	if w.config.SyncOnStart {
		logger.InfoEvent().Msg("Performing initial sync")
		if _, err := w.performSync(ctx); err != nil {
			logger.WarnEvent().Err(err).Msg("Initial sync failed")
		}
	}

	// Start event processing goroutines
	var wg sync.WaitGroup

	// Goroutine 1: Watch file system events
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.watchFileSystem(ctx)
	}()

	// Goroutine 2: Process sync queue with debouncing
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.processSyncQueue(ctx)
	}()

	logger.InfoEvent().Msg("File watcher started")

	// Wait for context cancellation or stop signal
	select {
	case <-ctx.Done():
		logger.InfoEvent().Msg("Context cancelled, stopping watcher")
	case <-w.stopChan:
		logger.InfoEvent().Msg("Stop signal received")
	}

	// Cleanup
	w.mu.Lock()
	w.isRunning = false
	w.mu.Unlock()

	close(w.syncQueue)
	w.fsWatcher.Close()

	// Wait for goroutines to finish
	wg.Wait()

	logger.InfoEvent().Msg("File watcher stopped")
	return nil
}

// Stop stops the file watcher.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isRunning {
		return
	}

	close(w.stopChan)
}

// watchFileSystem monitors file system events.
func (w *Watcher) watchFileSystem(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}

			w.handleFileSystemEvent(event)

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}

			logger.WarnEvent().Err(err).Msg("File watcher error")
		}
	}
}

// handleFileSystemEvent processes a file system event.
func (w *Watcher) handleFileSystemEvent(event fsnotify.Event) {
	logger.DebugEvent().
		Str("path", event.Name).
		Str("op", event.Op.String()).
		Msg("File system event")

	// Determine change type
	var changeType ChangeType
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		changeType = ChangeTypeCreate
	case event.Op&fsnotify.Write == fsnotify.Write:
		changeType = ChangeTypeUpdate
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		changeType = ChangeTypeDelete
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		changeType = ChangeTypeDelete
	default:
		// Ignore other operations
		return
	}

	// Queue sync event
	select {
	case w.syncQueue <- WatchEvent{
		Path:      event.Name,
		Type:      changeType,
		Timestamp: time.Now(),
	}:
		logger.DebugEvent().Str("path", event.Name).Msg("Queued sync event")
	default:
		logger.WarnEvent().Msg("Sync queue full, dropping event")
	}
}

// processSyncQueue processes queued sync events with debouncing.
func (w *Watcher) processSyncQueue(ctx context.Context) {
	var lastSync time.Time
	timer := time.NewTimer(0)
	<-timer.C // Drain the timer

	pendingSync := false

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case event, ok := <-w.syncQueue:
			if !ok {
				return
			}

			logger.DebugEvent().
				Str("path", event.Path).
				Str("type", string(event.Type)).
				Msg("Processing sync event")

			// Mark that we have a pending sync
			pendingSync = true

			// Reset debounce timer
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(w.config.DebounceDuration)

		case <-timer.C:
			if !pendingSync {
				continue
			}

			// Check if enough time has passed since last sync
			if time.Since(lastSync) < w.config.DebounceDuration {
				// Reset timer to wait a bit longer
				timer.Reset(w.config.DebounceDuration - time.Since(lastSync))
				continue
			}

			// Perform sync
			logger.InfoEvent().Msg("Debounce period elapsed, triggering sync")
			result, err := w.performSyncWithRetry(ctx)
			if err != nil {
				logger.ErrorEvent().Err(err).Msg("Sync failed after retries")
			} else {
				logger.InfoEvent().
					Int("added", result.Added).
					Int("updated", result.Updated).
					Int("deleted", result.Deleted).
					Msg("Sync completed successfully")
			}

			lastSync = time.Now()
			pendingSync = false
		}
	}
}

// performSyncWithRetry performs sync with retry logic.
func (w *Watcher) performSyncWithRetry(ctx context.Context) (*SyncResult, error) {
	var lastErr error

	for attempt := 1; attempt <= w.config.MaxRetries; attempt++ {
		result, err := w.performSync(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err
		logger.WarnEvent().
			Err(err).
			Int("attempt", attempt).
			Int("max_retries", w.config.MaxRetries).
			Msg("Sync attempt failed")

		if attempt < w.config.MaxRetries {
			logger.InfoEvent().
				Dur("delay", w.config.RetryDelay).
				Msg("Waiting before retry")

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-w.stopChan:
				return nil, fmt.Errorf("watcher stopped")
			case <-time.After(w.config.RetryDelay):
				// Continue to next attempt
			}
		}
	}

	return nil, fmt.Errorf("sync failed after %d attempts: %w", w.config.MaxRetries, lastErr)
}

// performSync executes a single sync operation.
func (w *Watcher) performSync(ctx context.Context) (*SyncResult, error) {
	logger.DebugEvent().Msg("Performing sync operation")

	result, err := w.syncer.Sync(ctx)
	if err != nil {
		return nil, err
	}

	// Print conflict summary if any conflicts were encountered
	if len(result.Conflicts) > 0 {
		PrintConflictSummary(result.Conflicts)
	}

	return result, nil
}

// IsRunning returns whether the watcher is currently running.
func (w *Watcher) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.isRunning
}
