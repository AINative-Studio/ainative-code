package design

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// SyncConfig represents the configuration for synchronization.
type SyncConfig struct {
	// ProjectID is the ID of the project to sync.
	ProjectID string

	// Direction is the direction of synchronization.
	Direction SyncDirection

	// WatchMode enables continuous watching for changes.
	WatchMode bool

	// WatchInterval is the interval for checking changes in watch mode.
	WatchInterval time.Duration

	// LocalPath is the local file path for tokens.
	LocalPath string

	// ConflictResolution is the strategy for resolving conflicts.
	ConflictResolution ConflictResolutionStrategy

	// DryRun performs a dry run without making actual changes.
	DryRun bool

	// Verbose enables verbose logging.
	Verbose bool
}

// Conflict represents a conflict between local and remote tokens.
type Conflict struct {
	// TokenName is the name of the conflicting token.
	TokenName string

	// LocalToken is the local version of the token.
	LocalToken *Token

	// RemoteToken is the remote version of the token.
	RemoteToken *Token

	// ConflictType is the type of conflict.
	ConflictType ConflictType

	// Resolution is the chosen resolution for this conflict.
	Resolution *ConflictResolution
}

// ConflictResolution represents the resolution of a conflict.
type ConflictResolution struct {
	// Strategy is the strategy used to resolve the conflict.
	Strategy ConflictResolutionStrategy

	// SelectedToken is the token that was selected as the resolution.
	SelectedToken *Token

	// Reason is the reason for the resolution.
	Reason string
}

// SyncResult represents the result of a synchronization operation.
type SyncResult struct {
	// Added is the number of tokens added.
	Added int

	// Updated is the number of tokens updated.
	Updated int

	// Deleted is the number of tokens deleted.
	Deleted int

	// Conflicts is the list of conflicts encountered.
	Conflicts []Conflict

	// Errors is the list of errors encountered.
	Errors []error

	// Duration is the time taken for the sync operation.
	Duration time.Duration

	// DryRun indicates if this was a dry run.
	DryRun bool
}

// Change represents a change detected in tokens.
type Change struct {
	// Type is the type of change.
	Type ChangeType

	// Token is the token that changed.
	Token *Token

	// OldToken is the previous version of the token (for updates).
	OldToken *Token

	// Timestamp is when the change was detected.
	Timestamp time.Time
}

// TokenWithMetadata extends Token with sync metadata.
type TokenWithMetadata struct {
	Token

	// LastSyncedAt is the timestamp when this token was last synced.
	LastSyncedAt time.Time `json:"last_synced_at,omitempty"`

	// Hash is the hash of the token for change detection.
	Hash string `json:"hash"`

	// Version is the version number for optimistic locking.
	Version int `json:"version"`
}

// ComputeHash computes a hash of the token for change detection.
func (t *Token) ComputeHash() (string, error) {
	// Create a normalized representation of the token
	data, err := json.Marshal(map[string]interface{}{
		"name":  t.Name,
		"type":  t.Type,
		"value": t.Value,
	})
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// Equals checks if two tokens are equal based on their content.
func (t *Token) Equals(other *Token) bool {
	if t == nil || other == nil {
		return t == other
	}

	return t.Name == other.Name &&
		t.Type == other.Type &&
		t.Value == other.Value
}

// SyncState represents the state of a synchronization session.
type SyncState struct {
	// ProjectID is the project being synced.
	ProjectID string

	// LastSyncedAt is the last successful sync timestamp.
	LastSyncedAt time.Time

	// LocalTokens is the snapshot of local tokens.
	LocalTokens map[string]*TokenWithMetadata

	// RemoteTokens is the snapshot of remote tokens.
	RemoteTokens map[string]*TokenWithMetadata
}

// WatchEvent represents a file system change event.
type WatchEvent struct {
	// Path is the path of the file that changed.
	Path string

	// Type is the type of change.
	Type ChangeType

	// Timestamp is when the change was detected.
	Timestamp time.Time
}
