package benchmark

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/database"
	"github.com/google/uuid"
)

// BenchmarkDatabaseConnection measures database connection establishment time
func BenchmarkDatabaseConnection(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	dbPath := filepath.Join(helper.TempDir, "bench.db")
	cfg := database.DefaultConfig(dbPath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		db, err := database.Connect(cfg)
		if err != nil {
			b.Fatalf("Failed to connect: %v", err)
		}

		elapsed := time.Since(start)

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/connect")
		}

		db.Close()
	}
}

// BenchmarkDatabaseInitialization measures database initialization with migrations
func BenchmarkDatabaseInitialization(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dbPath := filepath.Join(helper.TempDir, fmt.Sprintf("init_%d.db", i))
		cfg := database.DefaultConfig(dbPath)

		start := time.Now()

		db, err := database.Initialize(cfg)
		if err != nil {
			b.Fatalf("Failed to initialize: %v", err)
		}

		elapsed := time.Since(start)

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/init")
		}

		db.Close()
	}
}

// BenchmarkSessionQueries measures session query performance
func BenchmarkSessionQueries(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	dbPath := filepath.Join(helper.TempDir, "sessions.db")
	cfg := database.DefaultConfig(dbPath)

	db, err := database.Initialize(cfg)
	if err != nil {
		b.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create test sessions
	sessionIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		sessionID := uuid.New().String()
		_, err := db.CreateSession(ctx, database.CreateSessionParams{
			ID:        sessionID,
			Title:     fmt.Sprintf("Test Session %d", i),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		})
		if err != nil {
			b.Fatalf("Failed to create session: %v", err)
		}
		sessionIDs[i] = sessionID
	}

	b.Run("GetSession", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sessionID := sessionIDs[i%len(sessionIDs)]

			start := time.Now()
			_, err := db.GetSession(ctx, sessionID)
			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Failed to get session: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/get")
			}
		}
	})

	b.Run("ListSessions", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()
			sessions, err := db.ListSessions(ctx)
			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Failed to list sessions: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/list")
				b.Logf("Listed %d sessions", len(sessions))
			}
		}
	})

	b.Run("SearchSessions", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()
			sessions, err := db.SearchSessions(ctx, "Test")
			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Failed to search sessions: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/search")
				b.Logf("Found %d sessions", len(sessions))
			}
		}
	})
}

// BenchmarkMessageInsertion measures message insertion performance
func BenchmarkMessageInsertion(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	dbPath := filepath.Join(helper.TempDir, "messages.db")
	cfg := database.DefaultConfig(dbPath)

	db, err := database.Initialize(cfg)
	if err != nil {
		b.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create a test session
	sessionID := uuid.New().String()
	_, err = db.CreateSession(ctx, database.CreateSessionParams{
		ID:        sessionID,
		Title:     "Benchmark Session",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	})
	if err != nil {
		b.Fatalf("Failed to create session: %v", err)
	}

	b.Run("SingleMessage", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()

			_, err := db.CreateMessage(ctx, database.CreateMessageParams{
				ID:        uuid.New().String(),
				SessionID: sessionID,
				Role:      "user",
				Content:   "Test message content",
				CreatedAt: time.Now().Unix(),
			})

			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Failed to create message: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/insert")
			}
		}
	})

	b.Run("BatchMessages", func(b *testing.B) {
		batchSize := 10

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()

			// Insert batch of messages
			for j := 0; j < batchSize; j++ {
				_, err := db.CreateMessage(ctx, database.CreateMessageParams{
					ID:        uuid.New().String(),
					SessionID: sessionID,
					Role:      "user",
					Content:   fmt.Sprintf("Batch message %d", j),
					CreatedAt: time.Now().Unix(),
				})
				if err != nil {
					b.Fatalf("Failed to create message: %v", err)
				}
			}

			elapsed := time.Since(start)

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/batch-10")
			}
		}
	})
}

// BenchmarkMessageRetrieval measures message retrieval performance
func BenchmarkMessageRetrieval(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	dbPath := filepath.Join(helper.TempDir, "messages.db")
	cfg := database.DefaultConfig(dbPath)

	db, err := database.Initialize(cfg)
	if err != nil {
		b.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create test session with messages
	sessionID := uuid.New().String()
	_, err = db.CreateSession(ctx, database.CreateSessionParams{
		ID:        sessionID,
		Title:     "Benchmark Session",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	})
	if err != nil {
		b.Fatalf("Failed to create session: %v", err)
	}

	// Insert 100 messages
	for i := 0; i < 100; i++ {
		_, err := db.CreateMessage(ctx, database.CreateMessageParams{
			ID:        uuid.New().String(),
			SessionID: sessionID,
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
			CreatedAt: time.Now().Unix(),
		})
		if err != nil {
			b.Fatalf("Failed to create message: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		messages, err := db.GetSessionMessages(ctx, sessionID)
		elapsed := time.Since(start)

		if err != nil {
			b.Fatalf("Failed to get messages: %v", err)
		}

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/get-100-messages")
			b.Logf("Retrieved %d messages", len(messages))
		}
	}
}

// BenchmarkTransactionPerformance measures transaction overhead
func BenchmarkTransactionPerformance(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	dbPath := filepath.Join(helper.TempDir, "tx.db")
	cfg := database.DefaultConfig(dbPath)

	db, err := database.Initialize(cfg)
	if err != nil {
		b.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	b.Run("WithoutTransaction", func(b *testing.B) {
		sessionID := uuid.New().String()
		_, _ = db.CreateSession(ctx, database.CreateSessionParams{
			ID:        sessionID,
			Title:     "Test",
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()

			_, err := db.CreateMessage(ctx, database.CreateMessageParams{
				ID:        uuid.New().String(),
				SessionID: sessionID,
				Role:      "user",
				Content:   "Test",
				CreatedAt: time.Now().Unix(),
			})

			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Failed: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/no-tx")
			}
		}
	})

	b.Run("WithTransaction", func(b *testing.B) {
		sessionID := uuid.New().String()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()

			err := db.WithTx(ctx, func(q *database.Queries) error {
				_, err := q.CreateMessage(ctx, database.CreateMessageParams{
					ID:        uuid.New().String(),
					SessionID: sessionID,
					Role:      "user",
					Content:   "Test",
					CreatedAt: time.Now().Unix(),
				})
				return err
			})

			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Failed: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/with-tx")
			}
		}
	})
}

// BenchmarkDatabaseExportPerformance measures export operation performance
func BenchmarkDatabaseExportPerformance(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	dbPath := filepath.Join(helper.TempDir, "export.db")
	cfg := database.DefaultConfig(dbPath)

	db, err := database.Initialize(cfg)
	if err != nil {
		b.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create test data
	sessionID := uuid.New().String()
	_, err = db.CreateSession(ctx, database.CreateSessionParams{
		ID:        sessionID,
		Title:     "Export Test",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	})
	if err != nil {
		b.Fatalf("Failed to create session: %v", err)
	}

	// Insert messages
	for i := 0; i < 50; i++ {
		_, err := db.CreateMessage(ctx, database.CreateMessageParams{
			ID:        uuid.New().String(),
			SessionID: sessionID,
			Role:      "user",
			Content:   fmt.Sprintf("Export message %d", i),
			CreatedAt: time.Now().Unix(),
		})
		if err != nil {
			b.Fatalf("Failed to create message: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Export session (get session + messages)
		session, err := db.GetSession(ctx, sessionID)
		if err != nil {
			b.Fatalf("Failed to get session: %v", err)
		}

		messages, err := db.GetSessionMessages(ctx, sessionID)
		if err != nil {
			b.Fatalf("Failed to get messages: %v", err)
		}

		elapsed := time.Since(start)

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/export")
			b.Logf("Exported session with %d messages", len(messages))
			_ = session
		}
	}
}

// BenchmarkDatabaseConcurrency measures concurrent database access
func BenchmarkDatabaseConcurrency(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	dbPath := filepath.Join(helper.TempDir, "concurrent.db")
	cfg := database.DefaultConfig(dbPath)
	cfg.MaxOpenConns = 20

	db, err := database.Initialize(cfg)
	if err != nil {
		b.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create test session
	sessionID := uuid.New().String()
	_, err = db.CreateSession(ctx, database.CreateSessionParams{
		ID:        sessionID,
		Title:     "Concurrent Test",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	})
	if err != nil {
		b.Fatalf("Failed to create session: %v", err)
	}

	concurrency := []int{1, 5, 10}

	for _, n := range concurrency {
		b.Run(fmt.Sprintf("Workers_%d", n), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				done := make(chan bool, n)

				for j := 0; j < n; j++ {
					go func() {
						_, err := db.CreateMessage(ctx, database.CreateMessageParams{
							ID:        uuid.New().String(),
							SessionID: sessionID,
							Role:      "user",
							Content:   "Concurrent message",
							CreatedAt: time.Now().Unix(),
						})
						done <- err == nil
					}()
				}

				// Wait for all workers
				for j := 0; j < n; j++ {
					<-done
				}
			}
		})
	}
}
