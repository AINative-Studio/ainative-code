package integration

import (
	"context"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/session"
	"github.com/google/uuid"
)

// CreateTestSession is a helper function to create a test session
func CreateTestSession(t *testing.T, mgr session.Manager, name string) *session.Session {
	t.Helper()

	sess := &session.Session{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    session.StatusActive,
	}

	ctx := context.Background()
	err := mgr.CreateSession(ctx, sess)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	return sess
}

// CreateTestMessage is a helper function to create a test message
func CreateTestMessage(t *testing.T, mgr session.Manager, sessionID string, role session.MessageRole, content string) *session.Message {
	t.Helper()

	msg := &session.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	err := mgr.AddMessage(ctx, msg)
	if err != nil {
		t.Fatalf("Failed to create test message: %v", err)
	}

	return msg
}

// CreateTestConversation creates a session with a sample conversation
func CreateTestConversation(t *testing.T, mgr session.Manager) (*session.Session, []*session.Message) {
	t.Helper()

	sess := CreateTestSession(t, mgr, "Test Conversation")

	messages := []*session.Message{
		CreateTestMessage(t, mgr, sess.ID, session.RoleUser, "Hello, how are you?"),
		CreateTestMessage(t, mgr, sess.ID, session.RoleAssistant, "I'm doing well, thank you!"),
		CreateTestMessage(t, mgr, sess.ID, session.RoleUser, "Can you help me with a task?"),
		CreateTestMessage(t, mgr, sess.ID, session.RoleAssistant, "Of course! What do you need help with?"),
	}

	return sess, messages
}

// AssertSessionExists asserts that a session exists in the database
func AssertSessionExists(t *testing.T, mgr session.Manager, sessionID string) {
	t.Helper()

	ctx := context.Background()
	_, err := mgr.GetSession(ctx, sessionID)
	if err != nil {
		t.Errorf("Expected session %s to exist, but got error: %v", sessionID, err)
	}
}

// AssertSessionNotExists asserts that a session does not exist in the database
func AssertSessionNotExists(t *testing.T, mgr session.Manager, sessionID string) {
	t.Helper()

	ctx := context.Background()
	_, err := mgr.GetSession(ctx, sessionID)
	if err == nil {
		t.Errorf("Expected session %s to not exist, but it was found", sessionID)
	}
}

// AssertMessageCount asserts the number of messages in a session
func AssertMessageCount(t *testing.T, mgr session.Manager, sessionID string, expected int64) {
	t.Helper()

	ctx := context.Background()
	count, err := mgr.GetSessionMessageCount(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get message count: %v", err)
	}

	if count != expected {
		t.Errorf("Expected %d messages in session %s, but found %d", expected, sessionID, count)
	}
}
