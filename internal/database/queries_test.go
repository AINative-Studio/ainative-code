package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMetadataOperations(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx := context.Background()

	// Test SetMetadata
	err := db.SetMetadata(ctx, SetMetadataParams{
		Key:   "test_key",
		Value: "test_value",
	})
	if err != nil {
		t.Fatalf("failed to set metadata: %v", err)
	}

	// Test GetMetadata
	metadata, err := db.GetMetadata(ctx, "test_key")
	if err != nil {
		t.Fatalf("failed to get metadata: %v", err)
	}
	if metadata.Value != "test_value" {
		t.Errorf("expected value 'test_value', got '%s'", metadata.Value)
	}

	// Test ListMetadata
	list, err := db.ListMetadata(ctx)
	if err != nil {
		t.Fatalf("failed to list metadata: %v", err)
	}
	if len(list) == 0 {
		t.Error("expected at least one metadata entry")
	}

	// Test MetadataExists
	exists, err := db.MetadataExists(ctx, "test_key")
	if err != nil {
		t.Fatalf("failed to check metadata existence: %v", err)
	}
	if !exists {
		t.Error("expected metadata to exist")
	}

	// Test DeleteMetadata
	err = db.DeleteMetadata(ctx, "test_key")
	if err != nil {
		t.Fatalf("failed to delete metadata: %v", err)
	}

	// Verify deletion
	exists, err = db.MetadataExists(ctx, "test_key")
	if err != nil {
		t.Fatalf("failed to check metadata existence after delete: %v", err)
	}
	if exists {
		t.Error("expected metadata to not exist after deletion")
	}
}

func TestSessionOperations(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx := context.Background()
	sessionID := uuid.New().String()

	// Test CreateSession
	err := db.CreateSession(ctx, CreateSessionParams{
		ID:          sessionID,
		Name:        "Test Session",
		Status:      "active",
		Model:       strPtr("gpt-4"),
		Temperature: float64Ptr(0.7),
		MaxTokens:   int64Ptr(2000),
		Settings:    strPtr(`{"key": "value"}`),
	})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Test GetSession
	session, err := db.GetSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	if session.Name != "Test Session" {
		t.Errorf("expected name 'Test Session', got '%s'", session.Name)
	}
	if session.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", session.Status)
	}

	// Test ListSessions
	sessions, err := db.ListSessions(ctx, ListSessionsParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}
	if len(sessions) == 0 {
		t.Error("expected at least one session")
	}

	// Test UpdateSession
	err = db.UpdateSession(ctx, UpdateSessionParams{
		Name:        "Updated Session",
		Model:       strPtr("gpt-4-turbo"),
		Temperature: float64Ptr(0.8),
		MaxTokens:   int64Ptr(3000),
		Settings:    strPtr(`{"updated": true}`),
		ID:          sessionID,
	})
	if err != nil {
		t.Fatalf("failed to update session: %v", err)
	}

	// Verify update
	session, err = db.GetSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to get updated session: %v", err)
	}
	if session.Name != "Updated Session" {
		t.Errorf("expected name 'Updated Session', got '%s'", session.Name)
	}

	// Test CountSessions
	count, err := db.CountSessions(ctx)
	if err != nil {
		t.Fatalf("failed to count sessions: %v", err)
	}
	if count == 0 {
		t.Error("expected at least one session")
	}

	// Test ArchiveSession
	err = db.ArchiveSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to archive session: %v", err)
	}

	// Verify archive - GetSession still returns archived (only filters 'deleted')
	session, err = db.GetSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to get archived session: %v", err)
	}
	if session.Status != "archived" {
		t.Errorf("expected status 'archived', got '%s'", session.Status)
	}

	// Test DeleteSession (soft delete)
	err = db.UpdateSessionStatus(ctx, UpdateSessionStatusParams{
		Status: "active",
		ID:     sessionID,
	})
	if err != nil {
		t.Fatalf("failed to reactivate session: %v", err)
	}

	err = db.DeleteSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to delete session: %v", err)
	}

	// Verify soft delete
	session, err = db.GetSession(ctx, sessionID)
	if err == nil {
		t.Error("expected error getting deleted session")
	}
}

func TestMessageOperations(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx := context.Background()
	sessionID := uuid.New().String()
	messageID := uuid.New().String()

	// Create a session first
	err := db.CreateSession(ctx, CreateSessionParams{
		ID:     sessionID,
		Name:   "Test Session",
		Status: "active",
	})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Test CreateMessage
	err = db.CreateMessage(ctx, CreateMessageParams{
		ID:           messageID,
		SessionID:    sessionID,
		Role:         "user",
		Content:      "Hello, world!",
		ParentID:     nil,
		TokensUsed:   int64Ptr(10),
		Model:        strPtr("gpt-4"),
		FinishReason: strPtr("stop"),
		Metadata:     strPtr(`{"key": "value"}`),
	})
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}

	// Test GetMessage
	message, err := db.GetMessage(ctx, messageID)
	if err != nil {
		t.Fatalf("failed to get message: %v", err)
	}
	if message.Content != "Hello, world!" {
		t.Errorf("expected content 'Hello, world!', got '%s'", message.Content)
	}
	if message.Role != "user" {
		t.Errorf("expected role 'user', got '%s'", message.Role)
	}

	// Test ListMessagesBySession
	messages, err := db.ListMessagesBySession(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to list messages: %v", err)
	}
	if len(messages) == 0 {
		t.Error("expected at least one message")
	}

	// Test GetMessageCount
	count, err := db.GetMessageCount(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to count messages: %v", err)
	}
	if count == 0 {
		t.Error("expected at least one message")
	}

	// Test UpdateMessage
	err = db.UpdateMessage(ctx, UpdateMessageParams{
		Content:      "Updated content",
		TokensUsed:   int64Ptr(15),
		FinishReason: strPtr("length"),
		Metadata:     strPtr(`{"updated": true}`),
		ID:           messageID,
	})
	if err != nil {
		t.Fatalf("failed to update message: %v", err)
	}

	// Verify update
	message, err = db.GetMessage(ctx, messageID)
	if err != nil {
		t.Fatalf("failed to get updated message: %v", err)
	}
	if message.Content != "Updated content" {
		t.Errorf("expected content 'Updated content', got '%s'", message.Content)
	}

	// Test GetTotalTokensUsed
	totalTokens, err := db.GetTotalTokensUsed(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to get total tokens: %v", err)
	}
	if totalTokens == 0 {
		t.Error("expected tokens used to be greater than 0")
	}

	// Test DeleteMessage
	err = db.DeleteMessage(ctx, messageID)
	if err != nil {
		t.Fatalf("failed to delete message: %v", err)
	}

	// Verify deletion
	_, err = db.GetMessage(ctx, messageID)
	if err == nil {
		t.Error("expected error getting deleted message")
	}
}

func TestToolExecutionOperations(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx := context.Background()
	sessionID := uuid.New().String()
	messageID := uuid.New().String()
	toolExecID := uuid.New().String()

	// Create session and message first
	err := db.CreateSession(ctx, CreateSessionParams{
		ID:     sessionID,
		Name:   "Test Session",
		Status: "active",
	})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	err = db.CreateMessage(ctx, CreateMessageParams{
		ID:        messageID,
		SessionID: sessionID,
		Role:      "assistant",
		Content:   "Using tool",
	})
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}

	// Test CreateToolExecution
	now := time.Now().Format(time.RFC3339)
	completedAt := time.Now().Add(100 * time.Millisecond).Format(time.RFC3339)
	err = db.CreateToolExecution(ctx, CreateToolExecutionParams{
		ID:          toolExecID,
		MessageID:   messageID,
		ToolName:    "calculator",
		Input:       `{"operation": "add", "a": 1, "b": 2}`,
		Output:      strPtr(`{"result": 3}`),
		Status:      "success",
		Error:       nil,
		StartedAt:   now,
		CompletedAt: strPtr(completedAt),
		DurationMs:  int64Ptr(100),
		RetryCount:  0,
		Metadata:    strPtr(`{}`),
	})
	if err != nil {
		t.Fatalf("failed to create tool execution: %v", err)
	}

	// Test GetToolExecution
	toolExec, err := db.GetToolExecution(ctx, toolExecID)
	if err != nil {
		t.Fatalf("failed to get tool execution: %v", err)
	}
	if toolExec.ToolName != "calculator" {
		t.Errorf("expected tool name 'calculator', got '%s'", toolExec.ToolName)
	}
	if toolExec.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", toolExec.Status)
	}

	// Test ListToolExecutionsByMessage
	executions, err := db.ListToolExecutionsByMessage(ctx, messageID)
	if err != nil {
		t.Fatalf("failed to list tool executions: %v", err)
	}
	if len(executions) == 0 {
		t.Error("expected at least one tool execution")
	}

	// Test GetToolExecutionCount
	count, err := db.GetToolExecutionCount(ctx, messageID)
	if err != nil {
		t.Fatalf("failed to count tool executions: %v", err)
	}
	if count == 0 {
		t.Error("expected at least one tool execution")
	}

	// Test UpdateToolExecutionStatus
	err = db.UpdateToolExecutionStatus(ctx, UpdateToolExecutionStatusParams{
		Status: "failed",
		ID:     toolExecID,
	})
	if err != nil {
		t.Fatalf("failed to update tool execution status: %v", err)
	}

	// Verify update
	toolExec, err = db.GetToolExecution(ctx, toolExecID)
	if err != nil {
		t.Fatalf("failed to get updated tool execution: %v", err)
	}
	if toolExec.Status != "failed" {
		t.Errorf("expected status 'failed', got '%s'", toolExec.Status)
	}

	// Test GetToolExecutionStats
	stats, err := db.GetToolExecutionStats(ctx, "calculator")
	if err != nil {
		t.Fatalf("failed to get tool execution stats: %v", err)
	}
	if stats.TotalExecutions == 0 {
		t.Error("expected at least one execution in stats")
	}

	// Test DeleteToolExecution
	err = db.DeleteToolExecution(ctx, toolExecID)
	if err != nil {
		t.Fatalf("failed to delete tool execution: %v", err)
	}

	// Verify deletion
	_, err = db.GetToolExecution(ctx, toolExecID)
	if err == nil {
		t.Error("expected error getting deleted tool execution")
	}
}

func TestCascadeDelete(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx := context.Background()
	sessionID := uuid.New().String()
	messageID := uuid.New().String()
	toolExecID := uuid.New().String()

	// Create session, message, and tool execution
	err := db.CreateSession(ctx, CreateSessionParams{
		ID:     sessionID,
		Name:   "Test Session",
		Status: "active",
	})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	err = db.CreateMessage(ctx, CreateMessageParams{
		ID:        messageID,
		SessionID: sessionID,
		Role:      "user",
		Content:   "Test message",
	})
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}

	err = db.CreateToolExecution(ctx, CreateToolExecutionParams{
		ID:        toolExecID,
		MessageID: messageID,
		ToolName:  "test_tool",
		Input:     `{}`,
		Status:    "pending",
		StartedAt: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("failed to create tool execution: %v", err)
	}

	// Hard delete the session (should cascade)
	err = db.HardDeleteSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to hard delete session: %v", err)
	}

	// Verify message was cascaded
	_, err = db.GetMessage(ctx, messageID)
	if err == nil {
		t.Error("expected message to be deleted via cascade")
	}

	// Verify tool execution was cascaded
	_, err = db.GetToolExecution(ctx, toolExecID)
	if err == nil {
		t.Error("expected tool execution to be deleted via cascade")
	}
}

// Helper functions for creating pointer types (SQLC uses pointers for nullable fields)
func strPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// setupTestDatabase creates an initialized test database
func setupTestDatabase(t *testing.T) *DB {
	config := DefaultConfig(":memory:")
	db, err := Initialize(config)
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}
	return db
}
