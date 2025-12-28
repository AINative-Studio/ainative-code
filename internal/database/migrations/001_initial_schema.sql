-- Migration: 001_initial_schema
-- Description: Initial database schema for session management and conversation persistence
-- Author: AINative-Code Team
-- Date: 2025-12-27

-- +migrate Up
-- Enable foreign key constraints (important for SQLite)
PRAGMA foreign_keys = ON;

-- Metadata table for storing application-level key-value pairs
CREATE TABLE IF NOT EXISTS metadata (
    key TEXT PRIMARY KEY NOT NULL,
    value TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
) STRICT;

-- Create index for faster metadata lookups
CREATE INDEX IF NOT EXISTS idx_metadata_updated_at ON metadata(updated_at);

-- Sessions table for managing conversation sessions
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- Optional fields for enhanced session management
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'archived', 'deleted')),
    model TEXT,
    temperature REAL,
    max_tokens INTEGER,
    -- Metadata JSON for flexible extension
    settings TEXT -- JSON blob for additional settings
) STRICT;

-- Create indexes for faster session queries
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_name ON sessions(name);

-- Messages table for storing conversation messages
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY NOT NULL,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system', 'tool')),
    content TEXT NOT NULL,
    timestamp TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- Optional fields for enhanced message tracking
    parent_id TEXT, -- For threaded conversations
    tokens_used INTEGER,
    model TEXT,
    finish_reason TEXT,
    -- Metadata JSON for flexible extension
    metadata TEXT, -- JSON blob for additional metadata
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES messages(id) ON DELETE SET NULL
) STRICT;

-- Create indexes for faster message queries
CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_messages_role ON messages(role);
CREATE INDEX IF NOT EXISTS idx_messages_parent_id ON messages(parent_id);
CREATE INDEX IF NOT EXISTS idx_messages_session_timestamp ON messages(session_id, timestamp DESC);

-- Tool executions table for tracking tool usage
CREATE TABLE IF NOT EXISTS tool_executions (
    id TEXT PRIMARY KEY NOT NULL,
    message_id TEXT NOT NULL,
    tool_name TEXT NOT NULL,
    input TEXT NOT NULL, -- JSON blob of input parameters
    output TEXT, -- JSON blob of output results
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'running', 'success', 'failed', 'timeout')),
    error TEXT, -- Error message if status is 'failed'
    started_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TEXT,
    duration_ms INTEGER, -- Duration in milliseconds
    retry_count INTEGER NOT NULL DEFAULT 0,
    -- Metadata JSON for flexible extension
    metadata TEXT, -- JSON blob for additional metadata
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
) STRICT;

-- Create indexes for faster tool execution queries
CREATE INDEX IF NOT EXISTS idx_tool_executions_message_id ON tool_executions(message_id);
CREATE INDEX IF NOT EXISTS idx_tool_executions_status ON tool_executions(status);
CREATE INDEX IF NOT EXISTS idx_tool_executions_tool_name ON tool_executions(tool_name);
CREATE INDEX IF NOT EXISTS idx_tool_executions_started_at ON tool_executions(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_tool_executions_message_started ON tool_executions(message_id, started_at DESC);

-- Create triggers to automatically update updated_at timestamps
CREATE TRIGGER IF NOT EXISTS update_metadata_timestamp
AFTER UPDATE ON metadata
FOR EACH ROW
BEGIN
    UPDATE metadata SET updated_at = CURRENT_TIMESTAMP WHERE key = NEW.key;
END;

CREATE TRIGGER IF NOT EXISTS update_sessions_timestamp
AFTER UPDATE ON sessions
FOR EACH ROW
BEGIN
    UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Insert initial metadata
INSERT OR IGNORE INTO metadata (key, value) VALUES ('schema_version', '1');
INSERT OR IGNORE INTO metadata (key, value) VALUES ('created_at', datetime('now'));

-- +migrate Down
-- Drop triggers
DROP TRIGGER IF EXISTS update_sessions_timestamp;
DROP TRIGGER IF EXISTS update_metadata_timestamp;

-- Drop indexes
DROP INDEX IF EXISTS idx_tool_executions_message_started;
DROP INDEX IF EXISTS idx_tool_executions_started_at;
DROP INDEX IF EXISTS idx_tool_executions_tool_name;
DROP INDEX IF EXISTS idx_tool_executions_status;
DROP INDEX IF EXISTS idx_tool_executions_message_id;
DROP INDEX IF EXISTS idx_messages_session_timestamp;
DROP INDEX IF EXISTS idx_messages_parent_id;
DROP INDEX IF EXISTS idx_messages_role;
DROP INDEX IF EXISTS idx_messages_timestamp;
DROP INDEX IF EXISTS idx_messages_session_id;
DROP INDEX IF EXISTS idx_sessions_name;
DROP INDEX IF EXISTS idx_sessions_created_at;
DROP INDEX IF EXISTS idx_sessions_updated_at;
DROP INDEX IF EXISTS idx_sessions_status;
DROP INDEX IF EXISTS idx_metadata_updated_at;

-- Drop tables (in reverse order of creation due to foreign keys)
DROP TABLE IF EXISTS tool_executions;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS metadata;
