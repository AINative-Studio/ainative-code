package session

import (
	"encoding/json"
	"time"
)

// SessionStatus represents the status of a session
type SessionStatus string

const (
	// StatusActive represents an active session
	StatusActive SessionStatus = "active"

	// StatusArchived represents an archived session
	StatusArchived SessionStatus = "archived"

	// StatusDeleted represents a deleted session (soft delete)
	StatusDeleted SessionStatus = "deleted"
)

// IsValid checks if a session status is valid
func (s SessionStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusArchived, StatusDeleted:
		return true
	default:
		return false
	}
}

// MessageRole represents the role of a message
type MessageRole string

const (
	// RoleUser represents a user message
	RoleUser MessageRole = "user"

	// RoleAssistant represents an assistant message
	RoleAssistant MessageRole = "assistant"

	// RoleSystem represents a system message
	RoleSystem MessageRole = "system"

	// RoleTool represents a tool message
	RoleTool MessageRole = "tool"
)

// IsValid checks if a message role is valid
func (r MessageRole) IsValid() bool {
	switch r {
	case RoleUser, RoleAssistant, RoleSystem, RoleTool:
		return true
	default:
		return false
	}
}

// Session represents a conversation session
type Session struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Status      SessionStatus  `json:"status"`
	Model       *string        `json:"model,omitempty"`
	Temperature *float64       `json:"temperature,omitempty"`
	MaxTokens   *int64         `json:"max_tokens,omitempty"`
	Settings    map[string]any `json:"settings,omitempty"`
}

// Message represents a conversation message
type Message struct {
	ID           string         `json:"id"`
	SessionID    string         `json:"session_id"`
	Role         MessageRole    `json:"role"`
	Content      string         `json:"content"`
	Timestamp    time.Time      `json:"timestamp"`
	ParentID     *string        `json:"parent_id,omitempty"`
	TokensUsed   *int64         `json:"tokens_used,omitempty"`
	Model        *string        `json:"model,omitempty"`
	FinishReason *string        `json:"finish_reason,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// SessionSummary represents a session with summary information
type SessionSummary struct {
	Session
	MessageCount int64 `json:"message_count"`
	TotalTokens  int64 `json:"total_tokens,omitempty"`
}

// ExportFormat represents the format for session export
type ExportFormat string

const (
	// ExportFormatJSON exports session as JSON
	ExportFormatJSON ExportFormat = "json"

	// ExportFormatMarkdown exports session as Markdown
	ExportFormatMarkdown ExportFormat = "markdown"

	// ExportFormatText exports session as plain text
	ExportFormatText ExportFormat = "text"
)

// IsValid checks if an export format is valid
func (f ExportFormat) IsValid() bool {
	switch f {
	case ExportFormatJSON, ExportFormatMarkdown, ExportFormatText:
		return true
	default:
		return false
	}
}

// SessionExport represents an exported session
type SessionExport struct {
	Session  Session   `json:"session"`
	Messages []Message `json:"messages"`
}

// MarshalSettings marshals settings map to JSON string
func MarshalSettings(settings map[string]any) (string, error) {
	if settings == nil || len(settings) == 0 {
		return "", nil
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalSettings unmarshals JSON string to settings map
func UnmarshalSettings(data string) (map[string]any, error) {
	if data == "" {
		return nil, nil
	}
	var settings map[string]any
	if err := json.Unmarshal([]byte(data), &settings); err != nil {
		return nil, err
	}
	return settings, nil
}

// MarshalMetadata marshals metadata map to JSON string
func MarshalMetadata(metadata map[string]any) (string, error) {
	if metadata == nil || len(metadata) == 0 {
		return "", nil
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalMetadata unmarshals JSON string to metadata map
func UnmarshalMetadata(data string) (map[string]any, error) {
	if data == "" {
		return nil, nil
	}
	var metadata map[string]any
	if err := json.Unmarshal([]byte(data), &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}
