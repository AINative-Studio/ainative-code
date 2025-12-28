-- Database schema for SQLC code generation
-- This file contains the DDL that SQLC will use to generate type-safe Go code

-- Enable foreign key constraints (important for SQLite)
PRAGMA foreign_keys = ON;

-- Metadata table for storing application-level key-value pairs
CREATE TABLE metadata (
    key TEXT PRIMARY KEY NOT NULL,
    value TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
) STRICT;

CREATE INDEX idx_metadata_updated_at ON metadata(updated_at);

-- Sessions table for managing conversation sessions
CREATE TABLE sessions (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'archived', 'deleted')),
    model TEXT,
    temperature REAL,
    max_tokens INTEGER,
    settings TEXT
) STRICT;

CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_updated_at ON sessions(updated_at DESC);
CREATE INDEX idx_sessions_created_at ON sessions(created_at DESC);
CREATE INDEX idx_sessions_name ON sessions(name);

-- Messages table for storing conversation messages
CREATE TABLE messages (
    id TEXT PRIMARY KEY NOT NULL,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system', 'tool')),
    content TEXT NOT NULL,
    timestamp TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    parent_id TEXT,
    tokens_used INTEGER,
    model TEXT,
    finish_reason TEXT,
    metadata TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES messages(id) ON DELETE SET NULL
) STRICT;

CREATE INDEX idx_messages_session_id ON messages(session_id);
CREATE INDEX idx_messages_timestamp ON messages(timestamp DESC);
CREATE INDEX idx_messages_role ON messages(role);
CREATE INDEX idx_messages_parent_id ON messages(parent_id);
CREATE INDEX idx_messages_session_timestamp ON messages(session_id, timestamp DESC);

-- Tool executions table for tracking tool usage
CREATE TABLE tool_executions (
    id TEXT PRIMARY KEY NOT NULL,
    message_id TEXT NOT NULL,
    tool_name TEXT NOT NULL,
    input TEXT NOT NULL,
    output TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'running', 'success', 'failed', 'timeout')),
    error TEXT,
    started_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TEXT,
    duration_ms INTEGER,
    retry_count INTEGER NOT NULL DEFAULT 0,
    metadata TEXT,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
) STRICT;

CREATE INDEX idx_tool_executions_message_id ON tool_executions(message_id);
CREATE INDEX idx_tool_executions_status ON tool_executions(status);
CREATE INDEX idx_tool_executions_tool_name ON tool_executions(tool_name);
CREATE INDEX idx_tool_executions_started_at ON tool_executions(started_at DESC);
CREATE INDEX idx_tool_executions_message_started ON tool_executions(message_id, started_at DESC);
