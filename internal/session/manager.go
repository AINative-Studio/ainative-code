package session

import (
	"context"
	"io"
)

// Manager defines the interface for session management operations
type Manager interface {
	// Session operations
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, id string) (*Session, error)
	GetSessionSummary(ctx context.Context, id string) (*SessionSummary, error)
	ListSessions(ctx context.Context, opts ...ListOption) ([]*Session, error)
	UpdateSession(ctx context.Context, session *Session) error
	DeleteSession(ctx context.Context, id string) error
	ArchiveSession(ctx context.Context, id string) error
	HardDeleteSession(ctx context.Context, id string) error

	// Message operations
	AddMessage(ctx context.Context, message *Message) error
	GetMessage(ctx context.Context, id string) (*Message, error)
	GetMessages(ctx context.Context, sessionID string) ([]*Message, error)
	GetMessagesPaginated(ctx context.Context, sessionID string, limit, offset int64) ([]*Message, error)
	GetConversationThread(ctx context.Context, messageID string) ([]*Message, error)
	UpdateMessage(ctx context.Context, message *Message) error
	DeleteMessage(ctx context.Context, id string) error

	// Search operations
	SearchSessions(ctx context.Context, query string, opts ...SearchOption) ([]*Session, error)
	SearchMessages(ctx context.Context, sessionID string, query string, opts ...SearchOption) ([]*Message, error)

	// Statistics operations
	GetSessionMessageCount(ctx context.Context, sessionID string) (int64, error)
	GetTotalTokensUsed(ctx context.Context, sessionID string) (int64, error)

	// Export/Import operations
	ExportSession(ctx context.Context, sessionID string, format ExportFormat, w io.Writer) error
	ImportSession(ctx context.Context, r io.Reader) (*Session, error)

	// Utility operations
	TouchSession(ctx context.Context, id string) error
	Close() error
}
