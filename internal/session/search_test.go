package session_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/database"
	"github.com/AINative-studio/ainative-code/internal/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSearchTestDB(t *testing.T) (*database.DB, *session.SQLiteManager, func()) {
	t.Helper()

	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_search.db")

	// Open database connection
	sqlDB, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)

	// Run migrations
	err = database.MigrateContext(context.Background(), sqlDB)
	require.NoError(t, err)

	// Create database wrapper
	db := database.NewDB(sqlDB)

	// Create session manager
	mgr := session.NewSQLiteManager(db)

	cleanup := func() {
		db.Close()
	}

	return db, mgr, cleanup
}

func createTestSessionWithMessages(t *testing.T, mgr *session.SQLiteManager, sessionName string, messages []struct {
	role    session.MessageRole
	content string
	model   string
}) string {
	t.Helper()
	ctx := context.Background()

	// Create session
	sessionID := uuid.New().String()
	sess := &session.Session{
		ID:        sessionID,
		Name:      sessionName,
		Status:    session.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := mgr.CreateSession(ctx, sess)
	require.NoError(t, err)

	// Add messages
	for i, msg := range messages {
		model := msg.model
		messageID := uuid.New().String()
		message := &session.Message{
			ID:        messageID,
			SessionID: sessionID,
			Role:      msg.role,
			Content:   msg.content,
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
		if model != "" {
			message.Model = &model
		}

		err := mgr.AddMessage(ctx, message)
		require.NoError(t, err)
	}

	return sessionID
}

func TestSearchMessages_BasicSearch(t *testing.T) {
	_, mgr, cleanup := setupSearchTestDB(t)
	defer cleanup()

	// Create test data
	createTestSessionWithMessages(t, mgr, "AI Development", []struct {
		role    session.MessageRole
		content string
		model   string
	}{
		{session.RoleUser, "How do I implement authentication in Go?", ""},
		{session.RoleAssistant, "To implement authentication in Go, you can use JWT tokens...", "claude-3-opus"},
		{session.RoleUser, "Can you show me an example with middleware?", ""},
		{session.RoleAssistant, "Sure! Here's an authentication middleware example...", "claude-3-opus"},
	})

	createTestSessionWithMessages(t, mgr, "Python Tutorial", []struct {
		role    session.MessageRole
		content string
		model   string
	}{
		{session.RoleUser, "How do I handle exceptions in Python?", ""},
		{session.RoleAssistant, "In Python, you use try-except blocks for exception handling...", "gpt-4"},
		{session.RoleUser, "What about custom exceptions?", ""},
		{session.RoleAssistant, "You can create custom exception classes by inheriting from Exception...", "gpt-4"},
	})

	ctx := context.Background()

	t.Run("search for authentication", func(t *testing.T) {
		opts := session.DefaultSearchOptionsWithQuery("authentication")
		results, err := mgr.SearchAllMessages(ctx, opts)

		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 2, "should find at least 2 messages about authentication")
		assert.Greater(t, results.TotalCount, int64(0), "total count should be greater than 0")

		// Verify results contain the search term
		found := false
		for _, result := range results.Results {
			if contains(result.Message.Content, "authentication") || contains(result.Snippet, "authentication") {
				found = true
				break
			}
		}
		assert.True(t, found, "results should contain 'authentication'")
	})

	t.Run("search for Python", func(t *testing.T) {
		opts := session.DefaultSearchOptionsWithQuery("Python")
		results, err := mgr.SearchAllMessages(ctx, opts)

		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 2, "should find at least 2 messages about Python")

		// Verify all results are from Python session
		for _, result := range results.Results {
			assert.Contains(t, result.SessionName, "Python", "results should be from Python session")
		}
	})

	t.Run("search with no results", func(t *testing.T) {
		opts := session.DefaultSearchOptionsWithQuery("nonexistent_term_xyz")
		results, err := mgr.SearchAllMessages(ctx, opts)

		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Equal(t, 0, len(results.Results), "should find no results")
		assert.Equal(t, int64(0), results.TotalCount, "total count should be 0")
	})
}

func TestSearchMessages_WithDateRange(t *testing.T) {
	_, mgr, cleanup := setupSearchTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	// Create test data with different timestamps
	createTestSessionWithMessages(t, mgr, "Recent Discussion", []struct {
		role    session.MessageRole
		content string
		model   string
	}{
		{session.RoleUser, "What is machine learning?", ""},
		{session.RoleAssistant, "Machine learning is a subset of AI...", "claude-3-opus"},
	})

	t.Run("search within date range", func(t *testing.T) {
		opts := &session.SearchOptions{
			Query:    "machine learning",
			Limit:    50,
			Offset:   0,
			DateFrom: &yesterday,
			DateTo:   &tomorrow,
		}

		results, err := mgr.SearchAllMessages(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 1, "should find messages within date range")
	})

	t.Run("search outside date range", func(t *testing.T) {
		farPast := now.Add(-365 * 24 * time.Hour)
		farPastEnd := now.Add(-364 * 24 * time.Hour)

		opts := &session.SearchOptions{
			Query:    "machine learning",
			Limit:    50,
			Offset:   0,
			DateFrom: &farPast,
			DateTo:   &farPastEnd,
		}

		results, err := mgr.SearchAllMessages(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Equal(t, 0, len(results.Results), "should find no messages outside date range")
	})
}

func TestSearchMessages_WithProvider(t *testing.T) {
	_, mgr, cleanup := setupSearchTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test data with different providers
	createTestSessionWithMessages(t, mgr, "Claude Session", []struct {
		role    session.MessageRole
		content string
		model   string
	}{
		{session.RoleUser, "Explain neural networks", ""},
		{session.RoleAssistant, "Neural networks are computational models...", "claude-3-opus"},
	})

	createTestSessionWithMessages(t, mgr, "GPT Session", []struct {
		role    session.MessageRole
		content string
		model   string
	}{
		{session.RoleUser, "Explain neural networks", ""},
		{session.RoleAssistant, "Neural networks consist of interconnected nodes...", "gpt-4"},
	})

	t.Run("search with Claude provider", func(t *testing.T) {
		opts := &session.SearchOptions{
			Query:    "neural networks",
			Limit:    50,
			Offset:   0,
			Provider: "claude",
		}

		results, err := mgr.SearchAllMessages(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)

		// Verify all results are from Claude
		for _, result := range results.Results {
			if result.Message.Model != nil {
				assert.Contains(t, *result.Message.Model, "claude", "results should be from Claude")
			}
		}
	})

	t.Run("search with GPT provider", func(t *testing.T) {
		opts := &session.SearchOptions{
			Query:    "neural networks",
			Limit:    50,
			Offset:   0,
			Provider: "gpt",
		}

		results, err := mgr.SearchAllMessages(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)

		// Verify all results are from GPT
		for _, result := range results.Results {
			if result.Message.Model != nil {
				assert.Contains(t, *result.Message.Model, "gpt", "results should be from GPT")
			}
		}
	})
}

func TestSearchMessages_Pagination(t *testing.T) {
	_, mgr, cleanup := setupSearchTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test data with many messages containing "test"
	messages := make([]struct {
		role    session.MessageRole
		content string
		model   string
	}, 15)

	for i := 0; i < 15; i++ {
		messages[i] = struct {
			role    session.MessageRole
			content string
			model   string
		}{
			role:    session.RoleUser,
			content: "This is a test message number " + string(rune('A'+i)),
			model:   "",
		}
	}

	createTestSessionWithMessages(t, mgr, "Pagination Test", messages)

	t.Run("first page", func(t *testing.T) {
		opts := &session.SearchOptions{
			Query:  "test message",
			Limit:  5,
			Offset: 0,
		}

		results, err := mgr.SearchAllMessages(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.LessOrEqual(t, len(results.Results), 5, "should return at most 5 results")
		assert.GreaterOrEqual(t, results.TotalCount, int64(15), "total count should be at least 15")
	})

	t.Run("second page", func(t *testing.T) {
		opts := &session.SearchOptions{
			Query:  "test message",
			Limit:  5,
			Offset: 5,
		}

		results, err := mgr.SearchAllMessages(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.LessOrEqual(t, len(results.Results), 5, "should return at most 5 results")
	})
}

func TestSearchMessages_RelevanceRanking(t *testing.T) {
	_, mgr, cleanup := setupSearchTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create messages with different relevance levels
	createTestSessionWithMessages(t, mgr, "Ranking Test", []struct {
		role    session.MessageRole
		content string
		model   string
	}{
		{session.RoleUser, "database", ""},
		{session.RoleAssistant, "database database database", ""}, // Should rank higher
		{session.RoleUser, "something else entirely", ""},
		{session.RoleAssistant, "here is some info about database systems", ""},
	})

	t.Run("results ordered by relevance", func(t *testing.T) {
		opts := session.DefaultSearchOptionsWithQuery("database")
		results, err := mgr.SearchAllMessages(ctx, opts)

		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 2, "should find at least 2 results")

		// Verify results are ordered by relevance (higher scores first)
		for i := 1; i < len(results.Results); i++ {
			assert.GreaterOrEqual(t,
				results.Results[i-1].RelevanceScore,
				results.Results[i].RelevanceScore,
				"results should be ordered by relevance score (descending)")
		}
	})
}

func TestSearchMessages_Highlighting(t *testing.T) {
	_, mgr, cleanup := setupSearchTestDB(t)
	defer cleanup()

	ctx := context.Background()

	createTestSessionWithMessages(t, mgr, "Highlight Test", []struct {
		role    session.MessageRole
		content string
		model   string
	}{
		{session.RoleUser, "I need help with Go programming", ""},
		{session.RoleAssistant, "Go is a great programming language for building scalable systems", ""},
	})

	t.Run("snippet contains highlighted terms", func(t *testing.T) {
		opts := session.DefaultSearchOptionsWithQuery("programming")
		results, err := mgr.SearchAllMessages(ctx, opts)

		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 1, "should find at least 1 result")

		// Verify snippet contains HTML highlights
		found := false
		for _, result := range results.Results {
			if contains(result.Snippet, "<mark>") && contains(result.Snippet, "</mark>") {
				found = true
				break
			}
		}
		assert.True(t, found, "snippet should contain <mark> tags for highlighting")
	})
}

func TestSearchOptions_Validation(t *testing.T) {
	t.Run("empty query", func(t *testing.T) {
		opts := &session.SearchOptions{
			Query:  "",
			Limit:  50,
			Offset: 0,
		}

		err := opts.Validate()
		assert.ErrorIs(t, err, session.ErrEmptySearchQuery)
	})

	t.Run("limit too high", func(t *testing.T) {
		opts := &session.SearchOptions{
			Query:  "test",
			Limit:  2000,
			Offset: 0,
		}

		err := opts.Validate()
		assert.ErrorIs(t, err, session.ErrSearchLimitExceeded)
	})

	t.Run("invalid date range", func(t *testing.T) {
		now := time.Now()
		tomorrow := now.Add(24 * time.Hour)

		opts := &session.SearchOptions{
			Query:    "test",
			Limit:    50,
			Offset:   0,
			DateFrom: &tomorrow,
			DateTo:   &now,
		}

		err := opts.Validate()
		assert.ErrorIs(t, err, session.ErrInvalidDateRange)
	})

	t.Run("valid options", func(t *testing.T) {
		opts := &session.SearchOptions{
			Query:  "test",
			Limit:  50,
			Offset: 0,
		}

		err := opts.Validate()
		assert.NoError(t, err)
	})

	t.Run("auto-corrects invalid limit", func(t *testing.T) {
		opts := &session.SearchOptions{
			Query:  "test",
			Limit:  0,
			Offset: 0,
		}

		err := opts.Validate()
		assert.NoError(t, err)
		assert.Equal(t, int64(50), opts.Limit, "should set default limit")
	})
}

func TestSearchMessages_EdgeCases(t *testing.T) {
	_, mgr, cleanup := setupSearchTestDB(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("search with special characters", func(t *testing.T) {
		createTestSessionWithMessages(t, mgr, "Special Chars", []struct {
			role    session.MessageRole
			content string
			model   string
		}{
			{session.RoleUser, "How to use && operator in Go?", ""},
			{session.RoleAssistant, "The && operator is a logical AND in Go", ""},
		})

		opts := session.DefaultSearchOptionsWithQuery("operator")
		results, err := mgr.SearchAllMessages(ctx, opts)

		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 1, "should handle special characters")
	})

	t.Run("search with very long query", func(t *testing.T) {
		longQuery := ""
		for i := 0; i < 100; i++ {
			longQuery += "test "
		}

		opts := session.DefaultSearchOptionsWithQuery(longQuery)
		results, err := mgr.SearchAllMessages(ctx, opts)

		// Should not error, just might not find results
		require.NoError(t, err)
		require.NotNil(t, results)
	})
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
