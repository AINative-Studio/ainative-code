// Package fixtures provides test data fixtures for integration tests.
package fixtures

import (
	"time"

	"github.com/AINative-studio/ainative-code/internal/session"
	"github.com/google/uuid"
)

// NewTestSession creates a test session with default values.
func NewTestSession() *session.Session {
	id := uuid.New().String()
	model := "claude-3-5-sonnet-20241022"
	temp := 0.7
	maxTokens := int64(4096)

	return &session.Session{
		ID:          id,
		Name:        "Test Session",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Status:      session.StatusActive,
		Model:       &model,
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		Settings:    map[string]any{},
	}
}

// NewTestSessionWithName creates a test session with a specific name.
func NewTestSessionWithName(name string) *session.Session {
	s := NewTestSession()
	s.Name = name
	return s
}

// NewTestMessage creates a test message with default values.
func NewTestMessage(sessionID string, role session.MessageRole, content string) *session.Message {
	id := uuid.New().String()
	tokens := int64(100)

	return &session.Message{
		ID:           id,
		SessionID:    sessionID,
		Role:         role,
		Content:      content,
		Timestamp:    time.Now(),
		TokensUsed:   &tokens,
		Model:        stringPtr("claude-3-5-sonnet-20241022"),
		FinishReason: stringPtr("end_turn"),
		Metadata:     map[string]any{},
	}
}

// NewTestMessageThread creates a thread of related messages.
func NewTestMessageThread(sessionID string, count int) []*session.Message {
	messages := make([]*session.Message, 0, count)
	var parentID *string

	for i := 0; i < count; i++ {
		role := session.RoleUser
		if i%2 == 1 {
			role = session.RoleAssistant
		}

		msg := NewTestMessage(sessionID, role, "Test message content")
		msg.ParentID = parentID
		messages = append(messages, msg)

		// Set parent for next message
		id := msg.ID
		parentID = &id
	}

	return messages
}

func stringPtr(s string) *string {
	return &s
}
