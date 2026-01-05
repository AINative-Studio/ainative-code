package design

import "fmt"

// TokenType represents the category of a design token.
type TokenType string

const (
	// TokenTypeColor represents color tokens (hex, rgb, hsl)
	TokenTypeColor TokenType = "color"
	// TokenTypeTypography represents typography-related tokens
	TokenTypeTypography TokenType = "typography"
	// TokenTypeSpacing represents spacing tokens (margin, padding)
	TokenTypeSpacing TokenType = "spacing"
	// TokenTypeShadow represents shadow tokens
	TokenTypeShadow TokenType = "shadow"
	// TokenTypeBorderRadius represents border-radius tokens
	TokenTypeBorderRadius TokenType = "border-radius"
)

// Token represents a single design token extracted from CSS/SCSS/LESS.
type Token struct {
	Name        string            `json:"name" yaml:"name"`
	Type        TokenType         `json:"type" yaml:"type"`
	Value       string            `json:"value" yaml:"value"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Category    string            `json:"category,omitempty" yaml:"category,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// TokenCollection represents a collection of design tokens organized by type.
type TokenCollection struct {
	Tokens   []Token           `json:"tokens" yaml:"tokens"`
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ColorFormat represents the format of a color value.
type ColorFormat string

const (
	ColorFormatHex  ColorFormat = "hex"
	ColorFormatRGB  ColorFormat = "rgb"
	ColorFormatRGBA ColorFormat = "rgba"
	ColorFormatHSL  ColorFormat = "hsl"
	ColorFormatHSLA ColorFormat = "hsla"
)

// ParseError represents an error that occurred during parsing.
type ParseError struct {
	Line    int
	Column  int
	Message string
	Source  string
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("parse error at line %d, column %d: %s", e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

// ExtractionResult represents the result of token extraction.
type ExtractionResult struct {
	Tokens   []Token
	Warnings []string
	Errors   []error
}

// OutputFormat represents the format for token output.
type OutputFormat string

const (
	OutputFormatJSON     OutputFormat = "json"
	OutputFormatYAML     OutputFormat = "yaml"
	OutputFormatTailwind OutputFormat = "tailwind"
)

// SyncDirection represents the direction of synchronization.
type SyncDirection string

const (
	// SyncDirectionPull pulls tokens from remote to local.
	SyncDirectionPull SyncDirection = "pull"

	// SyncDirectionPush pushes tokens from local to remote.
	SyncDirectionPush SyncDirection = "push"

	// SyncDirectionBidirectional performs bidirectional sync.
	SyncDirectionBidirectional SyncDirection = "bidirectional"
)

// ConflictResolutionStrategy represents how to handle conflicts.
type ConflictResolutionStrategy string

const (
	// ConflictResolutionLocal prefers local changes over remote.
	ConflictResolutionLocal ConflictResolutionStrategy = "local"

	// ConflictResolutionRemote prefers remote changes over remote.
	ConflictResolutionRemote ConflictResolutionStrategy = "remote"

	// ConflictResolutionNewest prefers the newest changes based on timestamp.
	ConflictResolutionNewest ConflictResolutionStrategy = "newest"

	// ConflictResolutionPrompt prompts the user to resolve conflicts.
	ConflictResolutionPrompt ConflictResolutionStrategy = "prompt"

	// ConflictResolutionMerge attempts to merge changes.
	ConflictResolutionMerge ConflictResolutionStrategy = "merge"
)

// ConflictType represents the type of conflict.
type ConflictType string

const (
	// ConflictTypeBothModified indicates both local and remote were modified.
	ConflictTypeBothModified ConflictType = "both_modified"

	// ConflictTypeLocalDeleted indicates local was deleted but remote was modified.
	ConflictTypeLocalDeleted ConflictType = "local_deleted"

	// ConflictTypeRemoteDeleted indicates remote was deleted but local was modified.
	ConflictTypeRemoteDeleted ConflictType = "remote_deleted"

	// ConflictTypeTypeChange indicates the token type changed.
	ConflictTypeTypeChange ConflictType = "type_change"
)

// ChangeType represents the type of change detected.
type ChangeType string

const (
	// ChangeTypeCreate indicates a token was created.
	ChangeTypeCreate ChangeType = "create"

	// ChangeTypeUpdate indicates a token was updated.
	ChangeTypeUpdate ChangeType = "update"

	// ChangeTypeDelete indicates a token was deleted.
	ChangeTypeDelete ChangeType = "delete"
)
