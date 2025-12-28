-- name: CreateMessage :exec
INSERT INTO messages (
    id, session_id, role, content, parent_id, tokens_used, model, finish_reason, metadata
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetMessage :one
SELECT id, session_id, role, content, timestamp, parent_id, tokens_used, model, finish_reason, metadata
FROM messages
WHERE id = ?;

-- name: ListMessagesBySession :many
SELECT id, session_id, role, content, timestamp, parent_id, tokens_used, model, finish_reason, metadata
FROM messages
WHERE session_id = ?
ORDER BY timestamp ASC;

-- name: ListMessagesBySessionPaginated :many
SELECT id, session_id, role, content, timestamp, parent_id, tokens_used, model, finish_reason, metadata
FROM messages
WHERE session_id = ?
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: ListMessagesByRole :many
SELECT id, session_id, role, content, timestamp, parent_id, tokens_used, model, finish_reason, metadata
FROM messages
WHERE session_id = ? AND role = ?
ORDER BY timestamp ASC;

-- name: ListMessagesByParent :many
SELECT id, session_id, role, content, timestamp, parent_id, tokens_used, model, finish_reason, metadata
FROM messages
WHERE parent_id = ?
ORDER BY timestamp ASC;

-- name: GetLatestMessage :one
SELECT id, session_id, role, content, timestamp, parent_id, tokens_used, model, finish_reason, metadata
FROM messages
WHERE session_id = ?
ORDER BY timestamp DESC
LIMIT 1;

-- name: GetMessageCount :one
SELECT COUNT(*) as count
FROM messages
WHERE session_id = ?;

-- name: GetMessageCountByRole :one
SELECT COUNT(*) as count
FROM messages
WHERE session_id = ? AND role = ?;

-- name: UpdateMessage :exec
UPDATE messages
SET content = ?,
    tokens_used = ?,
    finish_reason = ?,
    metadata = ?
WHERE id = ?;

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE id = ?;

-- name: DeleteMessagesBySession :exec
DELETE FROM messages
WHERE session_id = ?;

-- name: SearchMessages :many
SELECT id, session_id, role, content, timestamp, parent_id, tokens_used, model, finish_reason, metadata
FROM messages
WHERE session_id = ? AND content LIKE ?
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: GetMessagesByTimeRange :many
SELECT id, session_id, role, content, timestamp, parent_id, tokens_used, model, finish_reason, metadata
FROM messages
WHERE session_id = ?
  AND timestamp >= ?
  AND timestamp <= ?
ORDER BY timestamp ASC;

-- name: GetTotalTokensUsed :one
SELECT COALESCE(SUM(tokens_used), 0) as total_tokens
FROM messages
WHERE session_id = ?;

-- name: GetConversationThread :many
WITH RECURSIVE thread AS (
    -- Base case: start with the specified message
    SELECT messages.id, messages.session_id, messages.role, messages.content, messages.timestamp, messages.parent_id, messages.tokens_used, messages.model, messages.finish_reason, messages.metadata, 0 as depth
    FROM messages
    WHERE messages.id = ?

    UNION ALL

    -- Recursive case: get all ancestors
    SELECT m.id, m.session_id, m.role, m.content, m.timestamp, m.parent_id, m.tokens_used, m.model, m.finish_reason, m.metadata, t.depth + 1
    FROM messages m
    INNER JOIN thread t ON m.id = t.parent_id
)
SELECT thread.id, thread.session_id, thread.role, thread.content, thread.timestamp, thread.parent_id, thread.tokens_used, thread.model, thread.finish_reason, thread.metadata
FROM thread
ORDER BY thread.depth DESC, thread.timestamp ASC;
