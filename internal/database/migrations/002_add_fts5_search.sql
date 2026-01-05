-- Migration: 002_add_fts5_search
-- Description: Add FTS5 full-text search support for messages
-- Author: AINative-Code Team
-- Date: 2026-01-05

-- +migrate Up
-- Create FTS5 virtual table for full-text search on messages
CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
    message_id UNINDEXED,
    session_id UNINDEXED,
    role UNINDEXED,
    content,
    timestamp UNINDEXED,
    model UNINDEXED,
    tokenize = 'porter unicode61'
);

-- Populate FTS5 table with existing messages
INSERT INTO messages_fts (message_id, session_id, role, content, timestamp, model)
SELECT id, session_id, role, content, timestamp, model
FROM messages;

-- Create triggers to keep FTS5 index synchronized with messages table
CREATE TRIGGER IF NOT EXISTS messages_fts_insert
AFTER INSERT ON messages
BEGIN
    INSERT INTO messages_fts (message_id, session_id, role, content, timestamp, model)
    VALUES (NEW.id, NEW.session_id, NEW.role, NEW.content, NEW.timestamp, NEW.model);
END;

CREATE TRIGGER IF NOT EXISTS messages_fts_update
AFTER UPDATE ON messages
BEGIN
    UPDATE messages_fts
    SET content = NEW.content,
        role = NEW.role,
        model = NEW.model,
        timestamp = NEW.timestamp
    WHERE message_id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS messages_fts_delete
AFTER DELETE ON messages
BEGIN
    DELETE FROM messages_fts WHERE message_id = OLD.id;
END;

-- Update metadata to track FTS5 support
INSERT OR REPLACE INTO metadata (key, value) VALUES ('fts5_enabled', 'true');
INSERT OR REPLACE INTO metadata (key, value) VALUES ('fts5_version', '1');

-- +migrate Down
-- Drop triggers
DROP TRIGGER IF EXISTS messages_fts_delete;
DROP TRIGGER IF EXISTS messages_fts_update;
DROP TRIGGER IF EXISTS messages_fts_insert;

-- Drop FTS5 virtual table
DROP TABLE IF EXISTS messages_fts;

-- Remove metadata entries
DELETE FROM metadata WHERE key IN ('fts5_enabled', 'fts5_version');
