package cmd

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/AINative-studio/ainative-code/internal/database"
	"github.com/AINative-studio/ainative-code/internal/session"
)

// TestSessionDeleteCommand tests the session delete command initialization
func TestSessionDeleteCommand(t *testing.T) {
	if sessionDeleteCmd == nil {
		t.Fatal("sessionDeleteCmd should not be nil")
	}

	if sessionDeleteCmd.Use != "delete [session-id]" {
		t.Errorf("expected Use 'delete [session-id]', got %s", sessionDeleteCmd.Use)
	}

	if sessionDeleteCmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if len(sessionDeleteCmd.Aliases) == 0 {
		t.Error("expected aliases to be set")
	}

	// Verify aliases include 'rm' and 'remove'
	hasRM := false
	hasRemove := false
	for _, alias := range sessionDeleteCmd.Aliases {
		if alias == "rm" {
			hasRM = true
		}
		if alias == "remove" {
			hasRemove = true
		}
	}

	if !hasRM {
		t.Error("expected 'rm' alias to be present")
	}

	if !hasRemove {
		t.Error("expected 'remove' alias to be present")
	}
}

// TestSessionDeletion tests the actual session deletion with database
func TestSessionDeletion(t *testing.T) {
	// Create a temporary database for testing
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize test database
	config := database.DefaultConfig(dbPath)
	db, err := database.Initialize(config)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create session manager
	mgr := session.NewSQLiteManager(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a test session
	testSession := &session.Session{
		ID:        uuid.New().String(),
		Name:      "Test Session for Deletion",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    session.StatusActive,
	}

	err = mgr.CreateSession(ctx, testSession)
	if err != nil {
		t.Fatalf("failed to create test session: %v", err)
	}

	// Add some messages to the session
	testMessage := &session.Message{
		ID:        uuid.New().String(),
		SessionID: testSession.ID,
		Role:      session.RoleUser,
		Content:   "Test message",
		Timestamp: time.Now(),
	}

	err = mgr.AddMessage(ctx, testMessage)
	if err != nil {
		t.Fatalf("failed to add test message: %v", err)
	}

	// Verify session exists
	retrievedSession, err := mgr.GetSession(ctx, testSession.ID)
	if err != nil {
		t.Fatalf("failed to retrieve session before deletion: %v", err)
	}

	if retrievedSession.ID != testSession.ID {
		t.Errorf("session ID mismatch before deletion: expected %s, got %s",
			testSession.ID, retrievedSession.ID)
	}

	// Verify message exists
	messages, err := mgr.GetMessages(ctx, testSession.ID)
	if err != nil {
		t.Fatalf("failed to retrieve messages before deletion: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("expected 1 message before deletion, got %d", len(messages))
	}

	// Test hard delete (permanent deletion)
	err = mgr.HardDeleteSession(ctx, testSession.ID)
	if err != nil {
		t.Fatalf("failed to delete session: %v", err)
	}

	// Verify session no longer exists
	_, err = mgr.GetSession(ctx, testSession.ID)
	if err == nil {
		t.Error("expected error when retrieving deleted session, but got nil")
	}

	// Verify messages were also deleted
	messages, err = mgr.GetMessages(ctx, testSession.ID)
	if err != nil {
		t.Fatalf("failed to retrieve messages after deletion: %v", err)
	}

	if len(messages) != 0 {
		t.Errorf("expected 0 messages after deletion, got %d", len(messages))
	}
}

// TestSessionDeletionNonExistent tests deletion of non-existent session
func TestSessionDeletionNonExistent(t *testing.T) {
	// Create a temporary database for testing
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize test database
	config := database.DefaultConfig(dbPath)
	db, err := database.Initialize(config)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create session manager
	mgr := session.NewSQLiteManager(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to delete a non-existent session
	nonExistentID := uuid.New().String()
	err = mgr.HardDeleteSession(ctx, nonExistentID)

	// This should succeed (no-op) since the session doesn't exist
	// The implementation may handle this differently, so we just verify it doesn't panic
	if err != nil {
		t.Logf("deletion of non-existent session returned error: %v (this may be expected)", err)
	}
}

// TestSessionDeletionEmptyID tests deletion with empty session ID
func TestSessionDeletionEmptyID(t *testing.T) {
	// Create a temporary database for testing
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize test database
	config := database.DefaultConfig(dbPath)
	db, err := database.Initialize(config)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create session manager
	mgr := session.NewSQLiteManager(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to delete with empty ID
	err = mgr.HardDeleteSession(ctx, "")

	// This should return an error
	if err == nil {
		t.Error("expected error when deleting with empty ID, but got nil")
	}
}

// TestGetSessionMessageCount tests message count retrieval
func TestGetSessionMessageCount(t *testing.T) {
	// Create a temporary database for testing
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize test database
	config := database.DefaultConfig(dbPath)
	db, err := database.Initialize(config)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create session manager
	mgr := session.NewSQLiteManager(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a test session
	testSession := &session.Session{
		ID:        uuid.New().String(),
		Name:      "Test Session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    session.StatusActive,
	}

	err = mgr.CreateSession(ctx, testSession)
	if err != nil {
		t.Fatalf("failed to create test session: %v", err)
	}

	// Initially, message count should be 0
	count, err := mgr.GetSessionMessageCount(ctx, testSession.ID)
	if err != nil {
		t.Fatalf("failed to get message count: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 messages initially, got %d", count)
	}

	// Add some messages
	for i := 0; i < 5; i++ {
		msg := &session.Message{
			ID:        uuid.New().String(),
			SessionID: testSession.ID,
			Role:      session.RoleUser,
			Content:   "Test message",
			Timestamp: time.Now(),
		}
		err = mgr.AddMessage(ctx, msg)
		if err != nil {
			t.Fatalf("failed to add message: %v", err)
		}
	}

	// Check message count again
	count, err = mgr.GetSessionMessageCount(ctx, testSession.ID)
	if err != nil {
		t.Fatalf("failed to get message count: %v", err)
	}

	if count != 5 {
		t.Errorf("expected 5 messages, got %d", count)
	}
}
