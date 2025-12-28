package session

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *database.DB {
	t.Helper()

	config := database.DefaultConfig(":memory:")
	db, err := database.Initialize(config)
	require.NoError(t, err, "failed to create test database")

	return db
}

// createTestSession creates a test session with default values
func createTestSession(t *testing.T, name string) *Session {
	t.Helper()

	model := "claude-3-5-sonnet-20241022"
	temp := 0.7
	maxTokens := int64(4096)

	return &Session{
		ID:          uuid.New().String(),
		Name:        name,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Status:      StatusActive,
		Model:       &model,
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		Settings: map[string]any{
			"theme": "dark",
			"auto_save": true,
		},
	}
}

// createTestMessage creates a test message with default values
func createTestMessage(t *testing.T, sessionID string, role MessageRole, content string) *Message {
	t.Helper()

	model := "claude-3-5-sonnet-20241022"
	tokens := int64(150)
	finishReason := "end_turn"

	return &Message{
		ID:           uuid.New().String(),
		SessionID:    sessionID,
		Role:         role,
		Content:      content,
		Timestamp:    time.Now().UTC(),
		TokensUsed:   &tokens,
		Model:        &model,
		FinishReason: &finishReason,
		Metadata: map[string]any{
			"source": "test",
		},
	}
}

func TestNewSQLiteManager(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	assert.NotNil(t, manager)
	assert.Equal(t, db, manager.db)
}

func TestCreateSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		session := createTestSession(t, "Test Session")
		err := manager.CreateSession(ctx, session)
		require.NoError(t, err)

		// Verify session was created
		retrieved, err := manager.GetSession(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.Equal(t, session.Name, retrieved.Name)
		assert.Equal(t, session.Status, retrieved.Status)
		assert.NotNil(t, retrieved.Model)
		assert.Equal(t, *session.Model, *retrieved.Model)
		assert.NotNil(t, retrieved.Temperature)
		assert.InDelta(t, *session.Temperature, *retrieved.Temperature, 0.001)
	})

	t.Run("NilSession", func(t *testing.T) {
		err := manager.CreateSession(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session is nil")
	})

	t.Run("EmptyName", func(t *testing.T) {
		session := createTestSession(t, "")
		err := manager.CreateSession(ctx, session)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptySessionName)
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		session := createTestSession(t, "Test")
		session.Status = SessionStatus("invalid")
		err := manager.CreateSession(ctx, session)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidStatus)
	})

	t.Run("DuplicateID", func(t *testing.T) {
		session := createTestSession(t, "Test Session")
		err := manager.CreateSession(ctx, session)
		require.NoError(t, err)

		// Try to create again with same ID
		err = manager.CreateSession(ctx, session)
		assert.Error(t, err)
	})
}

func TestGetSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		session := createTestSession(t, "Test Session")
		err := manager.CreateSession(ctx, session)
		require.NoError(t, err)

		retrieved, err := manager.GetSession(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.Equal(t, session.Name, retrieved.Name)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := manager.GetSession(ctx, "nonexistent-id")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("EmptyID", func(t *testing.T) {
		_, err := manager.GetSession(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidSessionID)
	})
}

func TestGetSessionSummary(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	// Add messages
	msg1 := createTestMessage(t, session.ID, RoleUser, "Hello")
	err = manager.AddMessage(ctx, msg1)
	require.NoError(t, err)

	msg2 := createTestMessage(t, session.ID, RoleAssistant, "Hi there")
	err = manager.AddMessage(ctx, msg2)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		summary, err := manager.GetSessionSummary(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, session.ID, summary.ID)
		assert.Equal(t, session.Name, summary.Name)
		assert.Equal(t, int64(2), summary.MessageCount)
		assert.Greater(t, summary.TotalTokens, int64(0))
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := manager.GetSessionSummary(ctx, "nonexistent-id")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})
}

func TestListSessions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	// Create test sessions
	session1 := createTestSession(t, "Session 1")
	session1.Status = StatusActive
	err := manager.CreateSession(ctx, session1)
	require.NoError(t, err)

	session2 := createTestSession(t, "Session 2")
	session2.Status = StatusArchived
	err = manager.CreateSession(ctx, session2)
	require.NoError(t, err)

	session3 := createTestSession(t, "Session 3")
	session3.Status = StatusActive
	err = manager.CreateSession(ctx, session3)
	require.NoError(t, err)

	t.Run("AllSessions", func(t *testing.T) {
		sessions, err := manager.ListSessions(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(sessions), 2) // At least 2 active sessions
	})

	t.Run("FilterByStatus", func(t *testing.T) {
		sessions, err := manager.ListSessions(ctx, WithStatus(StatusArchived))
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(sessions), 1)
		for _, s := range sessions {
			assert.Equal(t, StatusArchived, s.Status)
		}
	})

	t.Run("WithLimit", func(t *testing.T) {
		sessions, err := manager.ListSessions(ctx, WithLimit(1))
		require.NoError(t, err)
		assert.LessOrEqual(t, len(sessions), 1)
	})

	t.Run("WithOffset", func(t *testing.T) {
		allSessions, err := manager.ListSessions(ctx)
		require.NoError(t, err)

		if len(allSessions) > 1 {
			sessions, err := manager.ListSessions(ctx, WithOffset(1))
			require.NoError(t, err)
			assert.LessOrEqual(t, len(sessions), len(allSessions)-1)
		}
	})
}

func TestUpdateSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Original Name")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		session.Name = "Updated Name"
		newTemp := 0.9
		session.Temperature = &newTemp

		err := manager.UpdateSession(ctx, session)
		require.NoError(t, err)

		// Verify update
		retrieved, err := manager.GetSession(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Name)
		assert.NotNil(t, retrieved.Temperature)
		assert.InDelta(t, 0.9, *retrieved.Temperature, 0.001)
	})

	t.Run("NilSession", func(t *testing.T) {
		err := manager.UpdateSession(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("EmptyName", func(t *testing.T) {
		session.Name = ""
		err := manager.UpdateSession(ctx, session)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptySessionName)
	})
}

func TestDeleteSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		err := manager.DeleteSession(ctx, session.ID)
		require.NoError(t, err)

		// Verify soft delete - session should be hidden from GetSession
		_, err = manager.GetSession(ctx, session.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("EmptyID", func(t *testing.T) {
		err := manager.DeleteSession(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidSessionID)
	})
}

func TestArchiveSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		err := manager.ArchiveSession(ctx, session.ID)
		require.NoError(t, err)

		// Verify archived
		retrieved, err := manager.GetSession(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, StatusArchived, retrieved.Status)
	})

	t.Run("EmptyID", func(t *testing.T) {
		err := manager.ArchiveSession(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidSessionID)
	})
}

func TestHardDeleteSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	// Add messages
	msg := createTestMessage(t, session.ID, RoleUser, "Test")
	err = manager.AddMessage(ctx, msg)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		err := manager.HardDeleteSession(ctx, session.ID)
		require.NoError(t, err)

		// Verify session is gone
		_, err = manager.GetSession(ctx, session.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSessionNotFound)

		// Verify messages are gone
		messages, err := manager.GetMessages(ctx, session.ID)
		require.NoError(t, err)
		assert.Empty(t, messages)
	})

	t.Run("EmptyID", func(t *testing.T) {
		err := manager.HardDeleteSession(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidSessionID)
	})
}

func TestAddMessage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		msg := createTestMessage(t, session.ID, RoleUser, "Hello")
		err := manager.AddMessage(ctx, msg)
		require.NoError(t, err)

		// Verify message was created
		retrieved, err := manager.GetMessage(ctx, msg.ID)
		require.NoError(t, err)
		assert.Equal(t, msg.ID, retrieved.ID)
		assert.Equal(t, msg.Content, retrieved.Content)
		assert.Equal(t, msg.Role, retrieved.Role)
	})

	t.Run("NilMessage", func(t *testing.T) {
		err := manager.AddMessage(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("EmptyContent", func(t *testing.T) {
		msg := createTestMessage(t, session.ID, RoleUser, "")
		err := manager.AddMessage(ctx, msg)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptyMessageContent)
	})

	t.Run("InvalidRole", func(t *testing.T) {
		msg := createTestMessage(t, session.ID, MessageRole("invalid"), "Test")
		err := manager.AddMessage(ctx, msg)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRole)
	})

	t.Run("CircularReference", func(t *testing.T) {
		msg := createTestMessage(t, session.ID, RoleUser, "Test")
		msg.ParentID = &msg.ID // Self-reference
		err := manager.AddMessage(ctx, msg)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrCircularReference)
	})
}

func TestGetMessage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	msg := createTestMessage(t, session.ID, RoleUser, "Test")
	err = manager.AddMessage(ctx, msg)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		retrieved, err := manager.GetMessage(ctx, msg.ID)
		require.NoError(t, err)
		assert.Equal(t, msg.ID, retrieved.ID)
		assert.Equal(t, msg.Content, retrieved.Content)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := manager.GetMessage(ctx, "nonexistent-id")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrMessageNotFound)
	})

	t.Run("EmptyID", func(t *testing.T) {
		_, err := manager.GetMessage(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidMessageID)
	})
}

func TestGetMessages(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	// Add messages
	msg1 := createTestMessage(t, session.ID, RoleUser, "First")
	err = manager.AddMessage(ctx, msg1)
	require.NoError(t, err)

	msg2 := createTestMessage(t, session.ID, RoleAssistant, "Second")
	err = manager.AddMessage(ctx, msg2)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		messages, err := manager.GetMessages(ctx, session.ID)
		require.NoError(t, err)
		assert.Len(t, messages, 2)
	})

	t.Run("EmptySession", func(t *testing.T) {
		emptySession := createTestSession(t, "Empty")
		err := manager.CreateSession(ctx, emptySession)
		require.NoError(t, err)

		messages, err := manager.GetMessages(ctx, emptySession.ID)
		require.NoError(t, err)
		assert.Empty(t, messages)
	})
}

func TestGetMessagesPaginated(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	// Add 5 messages
	for i := 0; i < 5; i++ {
		msg := createTestMessage(t, session.ID, RoleUser, "Message")
		err = manager.AddMessage(ctx, msg)
		require.NoError(t, err)
	}

	t.Run("FirstPage", func(t *testing.T) {
		messages, err := manager.GetMessagesPaginated(ctx, session.ID, 2, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(messages), 2)
	})

	t.Run("SecondPage", func(t *testing.T) {
		messages, err := manager.GetMessagesPaginated(ctx, session.ID, 2, 2)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(messages), 2)
	})
}

func TestGetConversationThread(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	// Create message thread
	msg1 := createTestMessage(t, session.ID, RoleUser, "First")
	err = manager.AddMessage(ctx, msg1)
	require.NoError(t, err)

	msg2 := createTestMessage(t, session.ID, RoleAssistant, "Second")
	msg2.ParentID = &msg1.ID
	err = manager.AddMessage(ctx, msg2)
	require.NoError(t, err)

	msg3 := createTestMessage(t, session.ID, RoleUser, "Third")
	msg3.ParentID = &msg2.ID
	err = manager.AddMessage(ctx, msg3)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		thread, err := manager.GetConversationThread(ctx, msg3.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(thread), 1)
	})

	t.Run("EmptyID", func(t *testing.T) {
		_, err := manager.GetConversationThread(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidMessageID)
	})
}

func TestUpdateMessage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	msg := createTestMessage(t, session.ID, RoleUser, "Original")
	err = manager.AddMessage(ctx, msg)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		msg.Content = "Updated"
		err := manager.UpdateMessage(ctx, msg)
		require.NoError(t, err)

		// Verify update
		retrieved, err := manager.GetMessage(ctx, msg.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", retrieved.Content)
	})

	t.Run("NilMessage", func(t *testing.T) {
		err := manager.UpdateMessage(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("EmptyContent", func(t *testing.T) {
		msg.Content = ""
		err := manager.UpdateMessage(ctx, msg)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptyMessageContent)
	})
}

func TestDeleteMessage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	msg := createTestMessage(t, session.ID, RoleUser, "Test")
	err = manager.AddMessage(ctx, msg)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		err := manager.DeleteMessage(ctx, msg.ID)
		require.NoError(t, err)

		// Verify deleted
		_, err = manager.GetMessage(ctx, msg.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrMessageNotFound)
	})

	t.Run("EmptyID", func(t *testing.T) {
		err := manager.DeleteMessage(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidMessageID)
	})
}

func TestSearchSessions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	// Create test sessions with distinct names
	session1 := createTestSession(t, "Project Alpha Discussion")
	err := manager.CreateSession(ctx, session1)
	require.NoError(t, err)

	session2 := createTestSession(t, "Beta Testing Notes")
	err = manager.CreateSession(ctx, session2)
	require.NoError(t, err)

	t.Run("SearchByName", func(t *testing.T) {
		sessions, err := manager.SearchSessions(ctx, "Alpha")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(sessions), 1)
		found := false
		for _, s := range sessions {
			if s.ID == session1.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected to find session1 in search results")
	})

	t.Run("SearchByID", func(t *testing.T) {
		// Search by partial ID
		idPrefix := session2.ID[:8]
		sessions, err := manager.SearchSessions(ctx, idPrefix)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(sessions), 1)
	})

	t.Run("NoResults", func(t *testing.T) {
		sessions, err := manager.SearchSessions(ctx, "NonExistentTerm12345")
		require.NoError(t, err)
		assert.Empty(t, sessions)
	})
}

func TestSearchMessages(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	msg1 := createTestMessage(t, session.ID, RoleUser, "Hello world")
	err = manager.AddMessage(ctx, msg1)
	require.NoError(t, err)

	msg2 := createTestMessage(t, session.ID, RoleAssistant, "Goodbye moon")
	err = manager.AddMessage(ctx, msg2)
	require.NoError(t, err)

	t.Run("SearchContent", func(t *testing.T) {
		messages, err := manager.SearchMessages(ctx, session.ID, "world")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(messages), 1)
		found := false
		for _, m := range messages {
			if m.ID == msg1.ID {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("NoResults", func(t *testing.T) {
		messages, err := manager.SearchMessages(ctx, session.ID, "NonExistent")
		require.NoError(t, err)
		assert.Empty(t, messages)
	})
}

func TestGetSessionMessageCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	// Add messages
	for i := 0; i < 3; i++ {
		msg := createTestMessage(t, session.ID, RoleUser, "Message")
		err = manager.AddMessage(ctx, msg)
		require.NoError(t, err)
	}

	t.Run("Success", func(t *testing.T) {
		count, err := manager.GetSessionMessageCount(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})

	t.Run("EmptySession", func(t *testing.T) {
		emptySession := createTestSession(t, "Empty")
		err := manager.CreateSession(ctx, emptySession)
		require.NoError(t, err)

		count, err := manager.GetSessionMessageCount(ctx, emptySession.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestGetTotalTokensUsed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	tokens1 := int64(100)
	tokens2 := int64(200)

	msg1 := createTestMessage(t, session.ID, RoleUser, "First")
	msg1.TokensUsed = &tokens1
	err = manager.AddMessage(ctx, msg1)
	require.NoError(t, err)

	msg2 := createTestMessage(t, session.ID, RoleAssistant, "Second")
	msg2.TokensUsed = &tokens2
	err = manager.AddMessage(ctx, msg2)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		total, err := manager.GetTotalTokensUsed(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(300), total)
	})

	t.Run("NoMessages", func(t *testing.T) {
		emptySession := createTestSession(t, "Empty")
		err := manager.CreateSession(ctx, emptySession)
		require.NoError(t, err)

		total, err := manager.GetTotalTokensUsed(ctx, emptySession.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
	})
}

func TestExportSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Export Test")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	msg := createTestMessage(t, session.ID, RoleUser, "Test message")
	err = manager.AddMessage(ctx, msg)
	require.NoError(t, err)

	t.Run("JSON", func(t *testing.T) {
		var buf strings.Builder
		err := manager.ExportSession(ctx, session.ID, ExportFormatJSON, &buf)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, session.ID)
		assert.Contains(t, output, "Export Test")
		assert.Contains(t, output, "Test message")
	})

	t.Run("Markdown", func(t *testing.T) {
		var buf strings.Builder
		err := manager.ExportSession(ctx, session.ID, ExportFormatMarkdown, &buf)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "# Export Test")
		assert.Contains(t, output, session.ID)
		assert.Contains(t, output, "**user**:")
	})

	t.Run("Text", func(t *testing.T) {
		var buf strings.Builder
		err := manager.ExportSession(ctx, session.ID, ExportFormatText, &buf)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Session: Export Test")
		assert.Contains(t, output, "[user]:")
	})

	t.Run("InvalidFormat", func(t *testing.T) {
		var buf strings.Builder
		err := manager.ExportSession(ctx, session.ID, ExportFormat("invalid"), &buf)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidExportFormat)
	})

	t.Run("EmptySessionID", func(t *testing.T) {
		var buf strings.Builder
		err := manager.ExportSession(ctx, "", ExportFormatJSON, &buf)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidSessionID)
	})
}

func TestImportSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	// Create and export a session
	originalSession := createTestSession(t, "Import Test")
	err := manager.CreateSession(ctx, originalSession)
	require.NoError(t, err)

	msg := createTestMessage(t, originalSession.ID, RoleUser, "Test message")
	err = manager.AddMessage(ctx, msg)
	require.NoError(t, err)

	var exportBuf strings.Builder
	err = manager.ExportSession(ctx, originalSession.ID, ExportFormatJSON, &exportBuf)
	require.NoError(t, err)

	// Hard delete the original
	err = manager.HardDeleteSession(ctx, originalSession.ID)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		importReader := strings.NewReader(exportBuf.String())
		imported, err := manager.ImportSession(ctx, importReader)
		require.NoError(t, err)
		assert.NotNil(t, imported)
		assert.Equal(t, originalSession.ID, imported.ID)
		assert.Equal(t, originalSession.Name, imported.Name)

		// Verify messages were imported
		messages, err := manager.GetMessages(ctx, imported.ID)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		invalidReader := strings.NewReader("{invalid json}")
		_, err := manager.ImportSession(ctx, invalidReader)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidImportData)
	})
}

func TestTouchSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewSQLiteManager(db)
	ctx := context.Background()

	session := createTestSession(t, "Test Session")
	err := manager.CreateSession(ctx, session)
	require.NoError(t, err)

	// Get original updated_at
	original, err := manager.GetSession(ctx, session.ID)
	require.NoError(t, err)
	originalUpdated := original.UpdatedAt

	// Wait >1 second for SQLite's second-level timestamp precision to change
	time.Sleep(1100 * time.Millisecond)

	t.Run("Success", func(t *testing.T) {
		err := manager.TouchSession(ctx, session.ID)
		require.NoError(t, err)

		// Verify updated_at changed
		touched, err := manager.GetSession(ctx, session.ID)
		require.NoError(t, err)
		assert.True(t, touched.UpdatedAt.After(originalUpdated))
	})

	t.Run("EmptyID", func(t *testing.T) {
		err := manager.TouchSession(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidSessionID)
	})
}

func TestClose(t *testing.T) {
	db := setupTestDB(t)
	manager := NewSQLiteManager(db)

	err := manager.Close()
	assert.NoError(t, err)
}

func TestTypeConversions(t *testing.T) {
	t.Run("ParseTimestamp", func(t *testing.T) {
		// SQLite format
		ts, err := parseTimestamp("2024-01-15 10:30:45")
		require.NoError(t, err)
		assert.Equal(t, 2024, ts.Year())
		assert.Equal(t, time.January, ts.Month())
		assert.Equal(t, 15, ts.Day())

		// RFC3339 format (fallback)
		ts, err = parseTimestamp("2024-01-15T10:30:45Z")
		require.NoError(t, err)
		assert.Equal(t, 2024, ts.Year())

		// Invalid format
		_, err = parseTimestamp("invalid")
		assert.Error(t, err)

		// Empty string
		_, err = parseTimestamp("")
		assert.Error(t, err)
	})

	t.Run("FormatTimestamp", func(t *testing.T) {
		ts := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
		formatted := formatTimestamp(ts)
		assert.Equal(t, "2024-01-15 10:30:45", formatted)
	})

	t.Run("SettingsMarshalUnmarshal", func(t *testing.T) {
		settings := map[string]any{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}

		marshaled, err := MarshalSettings(settings)
		require.NoError(t, err)
		assert.NotEmpty(t, marshaled)

		unmarshaled, err := UnmarshalSettings(marshaled)
		require.NoError(t, err)
		assert.Equal(t, "value1", unmarshaled["key1"])
		assert.Equal(t, float64(42), unmarshaled["key2"]) // JSON numbers become float64
		assert.Equal(t, true, unmarshaled["key3"])
	})

	t.Run("MetadataMarshalUnmarshal", func(t *testing.T) {
		metadata := map[string]any{
			"source": "test",
			"count": 10,
		}

		marshaled, err := MarshalMetadata(metadata)
		require.NoError(t, err)
		assert.NotEmpty(t, marshaled)

		unmarshaled, err := UnmarshalMetadata(marshaled)
		require.NoError(t, err)
		assert.Equal(t, "test", unmarshaled["source"])
		assert.Equal(t, float64(10), unmarshaled["count"])
	})
}
