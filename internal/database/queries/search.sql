-- FTS5 Search Queries
-- These queries are executed using raw SQL due to SQLC limitations with FTS5 virtual tables
-- The search.go implementation will handle these queries manually

-- Search messages with full-text search (basic query)
-- Query: SELECT m.id, m.session_id, m.role, m.content, m.timestamp, m.parent_id, m.tokens_used, m.model, m.finish_reason, m.metadata, s.name as session_name, s.status as session_status, snippet(fts, 3, '<mark>', '</mark>', '...', 32) as snippet, bm25(fts) as relevance_score FROM messages_fts fts JOIN messages m ON fts.message_id = m.id JOIN sessions s ON m.session_id = s.id WHERE fts MATCH ? ORDER BY relevance_score LIMIT ? OFFSET ?

-- Search messages with date range filter
-- Query: SELECT m.id, m.session_id, m.role, m.content, m.timestamp, m.parent_id, m.tokens_used, m.model, m.finish_reason, m.metadata, s.name as session_name, s.status as session_status, snippet(fts, 3, '<mark>', '</mark>', '...', 32) as snippet, bm25(fts) as relevance_score FROM messages_fts fts JOIN messages m ON fts.message_id = m.id JOIN sessions s ON m.session_id = s.id WHERE fts MATCH ? AND m.timestamp >= ? AND m.timestamp <= ? ORDER BY relevance_score LIMIT ? OFFSET ?

-- Search messages with provider filter
-- Query: SELECT m.id, m.session_id, m.role, m.content, m.timestamp, m.parent_id, m.tokens_used, m.model, m.finish_reason, m.metadata, s.name as session_name, s.status as session_status, snippet(fts, 3, '<mark>', '</mark>', '...', 32) as snippet, bm25(fts) as relevance_score FROM messages_fts fts JOIN messages m ON fts.message_id = m.id JOIN sessions s ON m.session_id = s.id WHERE fts MATCH ? AND m.model LIKE ? ORDER BY relevance_score LIMIT ? OFFSET ?

-- Search messages with all filters
-- Query: SELECT m.id, m.session_id, m.role, m.content, m.timestamp, m.parent_id, m.tokens_used, m.model, m.finish_reason, m.metadata, s.name as session_name, s.status as session_status, snippet(fts, 3, '<mark>', '</mark>', '...', 32) as snippet, bm25(fts) as relevance_score FROM messages_fts fts JOIN messages m ON fts.message_id = m.id JOIN sessions s ON m.session_id = s.id WHERE fts MATCH ? AND m.timestamp >= ? AND m.timestamp <= ? AND m.model LIKE ? ORDER BY relevance_score LIMIT ? OFFSET ?

-- Count search results
-- Query: SELECT COUNT(*) FROM messages_fts fts JOIN messages m ON fts.message_id = m.id WHERE fts MATCH ?

-- Count search results with date range
-- Query: SELECT COUNT(*) FROM messages_fts fts JOIN messages m ON fts.message_id = m.id WHERE fts MATCH ? AND m.timestamp >= ? AND m.timestamp <= ?

-- Count search results with provider
-- Query: SELECT COUNT(*) FROM messages_fts fts JOIN messages m ON fts.message_id = m.id WHERE fts MATCH ? AND m.model LIKE ?

-- Count search results with all filters
-- Query: SELECT COUNT(*) FROM messages_fts fts JOIN messages m ON fts.message_id = m.id WHERE fts MATCH ? AND m.timestamp >= ? AND m.timestamp <= ? AND m.model LIKE ?

-- Rebuild FTS index
-- Query: INSERT INTO messages_fts(messages_fts) VALUES('rebuild')

-- Optimize FTS index
-- Query: INSERT INTO messages_fts(messages_fts) VALUES('optimize')
