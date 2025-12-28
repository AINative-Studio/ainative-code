package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/AINative-studio/ainative-code/internal/database"
)

// SQLiteManager implements the Manager interface using SQLite database
type SQLiteManager struct {
	db *database.DB
}

// NewSQLiteManager creates a new SQLiteManager instance
func NewSQLiteManager(db *database.DB) *SQLiteManager {
	return &SQLiteManager{
		db: db,
	}
}

// parseTimestamp converts SQLite TEXT timestamp to time.Time
func parseTimestamp(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}

	// SQLite CURRENT_TIMESTAMP format: "2006-01-02 15:04:05"
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		// Try RFC3339 format as fallback
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse timestamp %q: %w", s, err)
		}
	}
	return t, nil
}

// formatTimestamp converts time.Time to SQLite TEXT timestamp format
func formatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// convertDBSessionToDomain converts database.Session to domain Session
func convertDBSessionToDomain(dbSession database.Session) (*Session, error) {
	createdAt, err := parseTimestamp(dbSession.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	updatedAt, err := parseTimestamp(dbSession.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	var settings map[string]any
	if dbSession.Settings != nil && *dbSession.Settings != "" {
		settings, err = UnmarshalSettings(*dbSession.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
	}

	return &Session{
		ID:          dbSession.ID,
		Name:        dbSession.Name,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Status:      SessionStatus(dbSession.Status),
		Model:       dbSession.Model,
		Temperature: dbSession.Temperature,
		MaxTokens:   dbSession.MaxTokens,
		Settings:    settings,
	}, nil
}

// convertDomainSessionToCreateParams converts domain Session to database.CreateSessionParams
func convertDomainSessionToCreateParams(s *Session) (database.CreateSessionParams, error) {
	var settingsStr *string
	if s.Settings != nil && len(s.Settings) > 0 {
		str, err := MarshalSettings(s.Settings)
		if err != nil {
			return database.CreateSessionParams{}, fmt.Errorf("failed to marshal settings: %w", err)
		}
		settingsStr = &str
	}

	return database.CreateSessionParams{
		ID:          s.ID,
		Name:        s.Name,
		Status:      string(s.Status),
		Model:       s.Model,
		Temperature: s.Temperature,
		MaxTokens:   s.MaxTokens,
		Settings:    settingsStr,
	}, nil
}

// convertDomainSessionToUpdateParams converts domain Session to database.UpdateSessionParams
func convertDomainSessionToUpdateParams(s *Session) (database.UpdateSessionParams, error) {
	var settingsStr *string
	if s.Settings != nil && len(s.Settings) > 0 {
		str, err := MarshalSettings(s.Settings)
		if err != nil {
			return database.UpdateSessionParams{}, fmt.Errorf("failed to marshal settings: %w", err)
		}
		settingsStr = &str
	}

	return database.UpdateSessionParams{
		Name:        s.Name,
		Model:       s.Model,
		Temperature: s.Temperature,
		MaxTokens:   s.MaxTokens,
		Settings:    settingsStr,
		ID:          s.ID,
	}, nil
}

// convertDBMessageToDomain converts database.Message to domain Message
func convertDBMessageToDomain(dbMsg database.Message) (*Message, error) {
	timestamp, err := parseTimestamp(dbMsg.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	var metadata map[string]any
	if dbMsg.Metadata != nil && *dbMsg.Metadata != "" {
		metadata, err = UnmarshalMetadata(*dbMsg.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &Message{
		ID:           dbMsg.ID,
		SessionID:    dbMsg.SessionID,
		Role:         MessageRole(dbMsg.Role),
		Content:      dbMsg.Content,
		Timestamp:    timestamp,
		ParentID:     dbMsg.ParentID,
		TokensUsed:   dbMsg.TokensUsed,
		Model:        dbMsg.Model,
		FinishReason: dbMsg.FinishReason,
		Metadata:     metadata,
	}, nil
}

// convertDomainMessageToCreateParams converts domain Message to database.CreateMessageParams
func convertDomainMessageToCreateParams(m *Message) (database.CreateMessageParams, error) {
	var metadataStr *string
	if m.Metadata != nil && len(m.Metadata) > 0 {
		str, err := MarshalMetadata(m.Metadata)
		if err != nil {
			return database.CreateMessageParams{}, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataStr = &str
	}

	return database.CreateMessageParams{
		ID:           m.ID,
		SessionID:    m.SessionID,
		Role:         string(m.Role),
		Content:      m.Content,
		ParentID:     m.ParentID,
		TokensUsed:   m.TokensUsed,
		Model:        m.Model,
		FinishReason: m.FinishReason,
		Metadata:     metadataStr,
	}, nil
}

// convertDomainMessageToUpdateParams converts domain Message to database.UpdateMessageParams
func convertDomainMessageToUpdateParams(m *Message) (database.UpdateMessageParams, error) {
	var metadataStr *string
	if m.Metadata != nil && len(m.Metadata) > 0 {
		str, err := MarshalMetadata(m.Metadata)
		if err != nil {
			return database.UpdateMessageParams{}, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataStr = &str
	}

	return database.UpdateMessageParams{
		Content:      m.Content,
		TokensUsed:   m.TokensUsed,
		FinishReason: m.FinishReason,
		Metadata:     metadataStr,
		ID:           m.ID,
	}, nil
}

// CreateSession creates a new session
func (m *SQLiteManager) CreateSession(ctx context.Context, session *Session) error {
	if session == nil {
		return NewSessionError("CreateSession", ErrInvalidSessionID, "session is nil")
	}

	if session.Name == "" {
		return NewSessionError("CreateSession", ErrEmptySessionName, "")
	}

	if !session.Status.IsValid() {
		return NewSessionError("CreateSession", ErrInvalidStatus, string(session.Status))
	}

	params, err := convertDomainSessionToCreateParams(session)
	if err != nil {
		return NewSessionError("CreateSession", err, "failed to convert session")
	}

	if err := m.db.CreateSession(ctx, params); err != nil {
		return NewSessionError("CreateSession", err, "database error")
	}

	return nil
}

// GetSession retrieves a session by ID
func (m *SQLiteManager) GetSession(ctx context.Context, id string) (*Session, error) {
	if id == "" {
		return nil, NewSessionError("GetSession", ErrInvalidSessionID, "empty ID")
	}

	dbSession, err := m.db.GetSession(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewSessionError("GetSession", ErrSessionNotFound, id)
		}
		return nil, NewSessionError("GetSession", err, "database error")
	}

	session, err := convertDBSessionToDomain(dbSession)
	if err != nil {
		return nil, NewSessionError("GetSession", err, "conversion error")
	}

	return session, nil
}

// GetSessionSummary retrieves a session with message count and total tokens
func (m *SQLiteManager) GetSessionSummary(ctx context.Context, id string) (*SessionSummary, error) {
	if id == "" {
		return nil, NewSessionError("GetSessionSummary", ErrInvalidSessionID, "empty ID")
	}

	row, err := m.db.GetSessionWithMessageCount(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewSessionError("GetSessionSummary", ErrSessionNotFound, id)
		}
		return nil, NewSessionError("GetSessionSummary", err, "database error")
	}

	// Convert row to Session
	createdAt, err := parseTimestamp(row.CreatedAt)
	if err != nil {
		return nil, NewSessionError("GetSessionSummary", err, "failed to parse created_at")
	}

	updatedAt, err := parseTimestamp(row.UpdatedAt)
	if err != nil {
		return nil, NewSessionError("GetSessionSummary", err, "failed to parse updated_at")
	}

	var settings map[string]any
	if row.Settings != nil && *row.Settings != "" {
		settings, err = UnmarshalSettings(*row.Settings)
		if err != nil {
			return nil, NewSessionError("GetSessionSummary", err, "failed to unmarshal settings")
		}
	}

	// Get total tokens
	totalTokensRaw, err := m.db.GetTotalTokensUsed(ctx, id)
	if err != nil {
		return nil, NewSessionError("GetSessionSummary", err, "failed to get total tokens")
	}

	var totalTokens int64
	if totalTokensRaw != nil {
		// Handle type assertion from interface{}
		switch v := totalTokensRaw.(type) {
		case int64:
			totalTokens = v
		case int:
			totalTokens = int64(v)
		case float64:
			totalTokens = int64(v)
		}
	}

	return &SessionSummary{
		Session: Session{
			ID:          row.ID,
			Name:        row.Name,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			Status:      SessionStatus(row.Status),
			Model:       row.Model,
			Temperature: row.Temperature,
			MaxTokens:   row.MaxTokens,
			Settings:    settings,
		},
		MessageCount: row.MessageCount,
		TotalTokens:  totalTokens,
	}, nil
}

// ListSessions lists sessions with optional filters
func (m *SQLiteManager) ListSessions(ctx context.Context, opts ...ListOption) ([]*Session, error) {
	options := ApplyListOptions(opts...)

	var dbSessions []database.Session
	var err error

	if options.Status != "" {
		// Filter by status
		params := database.ListSessionsByStatusParams{
			Status: string(options.Status),
			Limit:  options.Limit,
			Offset: options.Offset,
		}
		dbSessions, err = m.db.ListSessionsByStatus(ctx, params)
	} else {
		// No status filter
		params := database.ListSessionsParams{
			Limit:  options.Limit,
			Offset: options.Offset,
		}
		dbSessions, err = m.db.ListSessions(ctx, params)
	}

	if err != nil {
		return nil, NewSessionError("ListSessions", err, "database error")
	}

	sessions := make([]*Session, 0, len(dbSessions))
	for _, dbSession := range dbSessions {
		session, err := convertDBSessionToDomain(dbSession)
		if err != nil {
			return nil, NewSessionError("ListSessions", err, "conversion error")
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// UpdateSession updates an existing session
func (m *SQLiteManager) UpdateSession(ctx context.Context, session *Session) error {
	if session == nil {
		return NewSessionError("UpdateSession", ErrInvalidSessionID, "session is nil")
	}

	if session.ID == "" {
		return NewSessionError("UpdateSession", ErrInvalidSessionID, "empty ID")
	}

	if session.Name == "" {
		return NewSessionError("UpdateSession", ErrEmptySessionName, "")
	}

	params, err := convertDomainSessionToUpdateParams(session)
	if err != nil {
		return NewSessionError("UpdateSession", err, "failed to convert session")
	}

	if err := m.db.UpdateSession(ctx, params); err != nil {
		return NewSessionError("UpdateSession", err, "database error")
	}

	return nil
}

// DeleteSession soft-deletes a session by setting status to 'deleted'
func (m *SQLiteManager) DeleteSession(ctx context.Context, id string) error {
	if id == "" {
		return NewSessionError("DeleteSession", ErrInvalidSessionID, "empty ID")
	}

	if err := m.db.DeleteSession(ctx, id); err != nil {
		return NewSessionError("DeleteSession", err, "database error")
	}

	return nil
}

// ArchiveSession archives a session by setting status to 'archived'
func (m *SQLiteManager) ArchiveSession(ctx context.Context, id string) error {
	if id == "" {
		return NewSessionError("ArchiveSession", ErrInvalidSessionID, "empty ID")
	}

	if err := m.db.ArchiveSession(ctx, id); err != nil {
		return NewSessionError("ArchiveSession", err, "database error")
	}

	return nil
}

// HardDeleteSession permanently deletes a session and all its messages
func (m *SQLiteManager) HardDeleteSession(ctx context.Context, id string) error {
	if id == "" {
		return NewSessionError("HardDeleteSession", ErrInvalidSessionID, "empty ID")
	}

	// Use transaction to ensure atomicity
	err := m.db.WithTx(ctx, func(q *database.Queries) error {
		// Delete all messages first (due to foreign key constraint)
		if err := q.DeleteMessagesBySession(ctx, id); err != nil {
			return fmt.Errorf("failed to delete messages: %w", err)
		}

		// Then delete the session
		if err := q.HardDeleteSession(ctx, id); err != nil {
			return fmt.Errorf("failed to delete session: %w", err)
		}

		return nil
	})

	if err != nil {
		return NewSessionError("HardDeleteSession", err, "transaction error")
	}

	return nil
}

// AddMessage adds a new message to a session
func (m *SQLiteManager) AddMessage(ctx context.Context, message *Message) error {
	if message == nil {
		return NewSessionError("AddMessage", ErrInvalidMessageID, "message is nil")
	}

	if message.Content == "" {
		return NewSessionError("AddMessage", ErrEmptyMessageContent, "")
	}

	if !message.Role.IsValid() {
		return NewSessionError("AddMessage", ErrInvalidRole, string(message.Role))
	}

	// Validate parent ID to prevent circular references
	if message.ParentID != nil && *message.ParentID == message.ID {
		return NewSessionError("AddMessage", ErrCircularReference, message.ID)
	}

	params, err := convertDomainMessageToCreateParams(message)
	if err != nil {
		return NewSessionError("AddMessage", err, "failed to convert message")
	}

	if err := m.db.CreateMessage(ctx, params); err != nil {
		return NewSessionError("AddMessage", err, "database error")
	}

	return nil
}

// GetMessage retrieves a message by ID
func (m *SQLiteManager) GetMessage(ctx context.Context, id string) (*Message, error) {
	if id == "" {
		return nil, NewSessionError("GetMessage", ErrInvalidMessageID, "empty ID")
	}

	dbMsg, err := m.db.GetMessage(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewSessionError("GetMessage", ErrMessageNotFound, id)
		}
		return nil, NewSessionError("GetMessage", err, "database error")
	}

	message, err := convertDBMessageToDomain(dbMsg)
	if err != nil {
		return nil, NewSessionError("GetMessage", err, "conversion error")
	}

	return message, nil
}

// GetMessages retrieves all messages for a session
func (m *SQLiteManager) GetMessages(ctx context.Context, sessionID string) ([]*Message, error) {
	if sessionID == "" {
		return nil, NewSessionError("GetMessages", ErrInvalidSessionID, "empty session ID")
	}

	dbMessages, err := m.db.ListMessagesBySession(ctx, sessionID)
	if err != nil {
		return nil, NewSessionError("GetMessages", err, "database error")
	}

	messages := make([]*Message, 0, len(dbMessages))
	for _, dbMsg := range dbMessages {
		message, err := convertDBMessageToDomain(dbMsg)
		if err != nil {
			return nil, NewSessionError("GetMessages", err, "conversion error")
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// GetMessagesPaginated retrieves messages for a session with pagination
func (m *SQLiteManager) GetMessagesPaginated(ctx context.Context, sessionID string, limit, offset int64) ([]*Message, error) {
	if sessionID == "" {
		return nil, NewSessionError("GetMessagesPaginated", ErrInvalidSessionID, "empty session ID")
	}

	params := database.ListMessagesBySessionPaginatedParams{
		SessionID: sessionID,
		Limit:     limit,
		Offset:    offset,
	}

	dbMessages, err := m.db.ListMessagesBySessionPaginated(ctx, params)
	if err != nil {
		return nil, NewSessionError("GetMessagesPaginated", err, "database error")
	}

	messages := make([]*Message, 0, len(dbMessages))
	for _, dbMsg := range dbMessages {
		message, err := convertDBMessageToDomain(dbMsg)
		if err != nil {
			return nil, NewSessionError("GetMessagesPaginated", err, "conversion error")
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// GetConversationThread retrieves all messages in a conversation thread
func (m *SQLiteManager) GetConversationThread(ctx context.Context, messageID string) ([]*Message, error) {
	if messageID == "" {
		return nil, NewSessionError("GetConversationThread", ErrInvalidMessageID, "empty message ID")
	}

	rows, err := m.db.GetConversationThread(ctx, messageID)
	if err != nil {
		return nil, NewSessionError("GetConversationThread", err, "database error")
	}

	messages := make([]*Message, 0, len(rows))
	for _, row := range rows {
		timestamp, err := parseTimestamp(row.Timestamp)
		if err != nil {
			return nil, NewSessionError("GetConversationThread", err, "failed to parse timestamp")
		}

		var metadata map[string]any
		if row.Metadata != nil && *row.Metadata != "" {
			metadata, err = UnmarshalMetadata(*row.Metadata)
			if err != nil {
				return nil, NewSessionError("GetConversationThread", err, "failed to unmarshal metadata")
			}
		}

		message := &Message{
			ID:           row.ID,
			SessionID:    row.SessionID,
			Role:         MessageRole(row.Role),
			Content:      row.Content,
			Timestamp:    timestamp,
			ParentID:     row.ParentID,
			TokensUsed:   row.TokensUsed,
			Model:        row.Model,
			FinishReason: row.FinishReason,
			Metadata:     metadata,
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// UpdateMessage updates an existing message
func (m *SQLiteManager) UpdateMessage(ctx context.Context, message *Message) error {
	if message == nil {
		return NewSessionError("UpdateMessage", ErrInvalidMessageID, "message is nil")
	}

	if message.ID == "" {
		return NewSessionError("UpdateMessage", ErrInvalidMessageID, "empty ID")
	}

	if message.Content == "" {
		return NewSessionError("UpdateMessage", ErrEmptyMessageContent, "")
	}

	params, err := convertDomainMessageToUpdateParams(message)
	if err != nil {
		return NewSessionError("UpdateMessage", err, "failed to convert message")
	}

	if err := m.db.UpdateMessage(ctx, params); err != nil {
		return NewSessionError("UpdateMessage", err, "database error")
	}

	return nil
}

// DeleteMessage deletes a message
func (m *SQLiteManager) DeleteMessage(ctx context.Context, id string) error {
	if id == "" {
		return NewSessionError("DeleteMessage", ErrInvalidMessageID, "empty ID")
	}

	if err := m.db.DeleteMessage(ctx, id); err != nil {
		return NewSessionError("DeleteMessage", err, "database error")
	}

	return nil
}

// SearchSessions searches for sessions by name or ID
func (m *SQLiteManager) SearchSessions(ctx context.Context, query string, opts ...SearchOption) ([]*Session, error) {
	options := ApplySearchOptions(opts...)

	// Add LIKE wildcards
	likePattern := "%" + query + "%"

	params := database.SearchSessionsParams{
		Name:   likePattern,
		ID:     likePattern,
		Limit:  options.Limit,
		Offset: options.Offset,
	}

	dbSessions, err := m.db.SearchSessions(ctx, params)
	if err != nil {
		return nil, NewSessionError("SearchSessions", err, "database error")
	}

	sessions := make([]*Session, 0, len(dbSessions))
	for _, dbSession := range dbSessions {
		session, err := convertDBSessionToDomain(dbSession)
		if err != nil {
			return nil, NewSessionError("SearchSessions", err, "conversion error")
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// SearchMessages searches for messages within a session by content
func (m *SQLiteManager) SearchMessages(ctx context.Context, sessionID string, query string, opts ...SearchOption) ([]*Message, error) {
	if sessionID == "" {
		return nil, NewSessionError("SearchMessages", ErrInvalidSessionID, "empty session ID")
	}

	options := ApplySearchOptions(opts...)

	// Add LIKE wildcards
	likePattern := "%" + query + "%"

	params := database.SearchMessagesParams{
		SessionID: sessionID,
		Content:   likePattern,
		Limit:     options.Limit,
		Offset:    options.Offset,
	}

	dbMessages, err := m.db.SearchMessages(ctx, params)
	if err != nil {
		return nil, NewSessionError("SearchMessages", err, "database error")
	}

	messages := make([]*Message, 0, len(dbMessages))
	for _, dbMsg := range dbMessages {
		message, err := convertDBMessageToDomain(dbMsg)
		if err != nil {
			return nil, NewSessionError("SearchMessages", err, "conversion error")
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// GetSessionMessageCount returns the number of messages in a session
func (m *SQLiteManager) GetSessionMessageCount(ctx context.Context, sessionID string) (int64, error) {
	if sessionID == "" {
		return 0, NewSessionError("GetSessionMessageCount", ErrInvalidSessionID, "empty session ID")
	}

	count, err := m.db.GetMessageCount(ctx, sessionID)
	if err != nil {
		return 0, NewSessionError("GetSessionMessageCount", err, "database error")
	}

	return count, nil
}

// GetTotalTokensUsed returns the total tokens used in a session
func (m *SQLiteManager) GetTotalTokensUsed(ctx context.Context, sessionID string) (int64, error) {
	if sessionID == "" {
		return 0, NewSessionError("GetTotalTokensUsed", ErrInvalidSessionID, "empty session ID")
	}

	totalTokensRaw, err := m.db.GetTotalTokensUsed(ctx, sessionID)
	if err != nil {
		return 0, NewSessionError("GetTotalTokensUsed", err, "database error")
	}

	var totalTokens int64
	if totalTokensRaw != nil {
		// Handle type assertion from interface{}
		switch v := totalTokensRaw.(type) {
		case int64:
			totalTokens = v
		case int:
			totalTokens = int64(v)
		case float64:
			totalTokens = int64(v)
		}
	}

	return totalTokens, nil
}

// ExportSession exports a session to the specified format
func (m *SQLiteManager) ExportSession(ctx context.Context, sessionID string, format ExportFormat, w io.Writer) error {
	if sessionID == "" {
		return NewSessionError("ExportSession", ErrInvalidSessionID, "empty session ID")
	}

	if !format.IsValid() {
		return NewSessionError("ExportSession", ErrInvalidExportFormat, string(format))
	}

	// Get session
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return NewSessionError("ExportSession", err, "failed to get session")
	}

	// Get all messages
	messages, err := m.GetMessages(ctx, sessionID)
	if err != nil {
		return NewSessionError("ExportSession", err, "failed to get messages")
	}

	// Export based on format
	switch format {
	case ExportFormatJSON:
		export := SessionExport{
			Session:  *session,
			Messages: make([]Message, len(messages)),
		}
		for i, msg := range messages {
			export.Messages[i] = *msg
		}

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(export); err != nil {
			return NewSessionError("ExportSession", err, "failed to encode JSON")
		}

	case ExportFormatMarkdown:
		// Write session header
		fmt.Fprintf(w, "# %s\n\n", session.Name)
		fmt.Fprintf(w, "**ID:** %s\n", session.ID)
		fmt.Fprintf(w, "**Status:** %s\n", session.Status)
		fmt.Fprintf(w, "**Created:** %s\n", session.CreatedAt.Format(time.RFC3339))
		fmt.Fprintf(w, "**Updated:** %s\n\n", session.UpdatedAt.Format(time.RFC3339))

		if session.Model != nil {
			fmt.Fprintf(w, "**Model:** %s\n", *session.Model)
		}

		fmt.Fprintf(w, "\n---\n\n")

		// Write messages
		for _, msg := range messages {
			fmt.Fprintf(w, "**%s**: ", msg.Role)
			fmt.Fprintf(w, "%s\n\n", msg.Content)
			fmt.Fprintf(w, "*%s*\n\n", msg.Timestamp.Format(time.RFC3339))

			if msg.TokensUsed != nil {
				fmt.Fprintf(w, "*Tokens: %d*\n\n", *msg.TokensUsed)
			}

			fmt.Fprintf(w, "---\n\n")
		}

	case ExportFormatText:
		// Write session header
		fmt.Fprintf(w, "Session: %s\n", session.Name)
		fmt.Fprintf(w, "ID: %s\n", session.ID)
		fmt.Fprintf(w, "Status: %s\n", session.Status)
		fmt.Fprintf(w, "Created: %s\n", session.CreatedAt.Format(time.RFC3339))
		fmt.Fprintf(w, "Updated: %s\n\n", session.UpdatedAt.Format(time.RFC3339))

		fmt.Fprintf(w, "========================================\n\n")

		// Write messages
		for _, msg := range messages {
			fmt.Fprintf(w, "[%s]: ", msg.Role)
			fmt.Fprintf(w, "%s\n\n", msg.Content)

			if msg.TokensUsed != nil {
				fmt.Fprintf(w, "(Tokens: %d)\n\n", *msg.TokensUsed)
			}

			fmt.Fprintf(w, "----------------------------------------\n\n")
		}

	default:
		return NewSessionError("ExportSession", ErrInvalidExportFormat, string(format))
	}

	return nil
}

// ImportSession imports a session from JSON format
func (m *SQLiteManager) ImportSession(ctx context.Context, r io.Reader) (*Session, error) {
	var export SessionExport
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&export); err != nil {
		return nil, NewSessionError("ImportSession", ErrInvalidImportData, fmt.Sprintf("failed to decode JSON: %v", err))
	}

	// Validate session
	if export.Session.ID == "" {
		return nil, NewSessionError("ImportSession", ErrInvalidImportData, "session ID is empty")
	}

	if export.Session.Name == "" {
		return nil, NewSessionError("ImportSession", ErrInvalidImportData, "session name is empty")
	}

	// Use transaction to ensure atomicity
	err := m.db.WithTx(ctx, func(q *database.Queries) error {
		// Create session
		params, err := convertDomainSessionToCreateParams(&export.Session)
		if err != nil {
			return fmt.Errorf("failed to convert session: %w", err)
		}

		if err := q.CreateSession(ctx, params); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		// Create all messages
		for _, msg := range export.Messages {
			msgCopy := msg // Create copy to avoid pointer issues
			msgParams, err := convertDomainMessageToCreateParams(&msgCopy)
			if err != nil {
				return fmt.Errorf("failed to convert message %s: %w", msg.ID, err)
			}

			if err := q.CreateMessage(ctx, msgParams); err != nil {
				return fmt.Errorf("failed to create message %s: %w", msg.ID, err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, NewSessionError("ImportSession", err, "transaction error")
	}

	return &export.Session, nil
}

// TouchSession updates the session's updated_at timestamp
func (m *SQLiteManager) TouchSession(ctx context.Context, id string) error {
	if id == "" {
		return NewSessionError("TouchSession", ErrInvalidSessionID, "empty ID")
	}

	if err := m.db.TouchSession(ctx, id); err != nil {
		return NewSessionError("TouchSession", err, "database error")
	}

	return nil
}

// Close closes the database connection
func (m *SQLiteManager) Close() error {
	// The database.DB doesn't expose a Close method
	// This is handled at a higher level by the caller
	return nil
}
