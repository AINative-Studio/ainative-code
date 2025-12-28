-- name: CreateSession :exec
INSERT INTO sessions (
    id, name, status, model, temperature, max_tokens, settings
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
);

-- name: GetSession :one
SELECT id, name, created_at, updated_at, status, model, temperature, max_tokens, settings
FROM sessions
WHERE id = ? AND status != 'deleted';

-- name: GetSessionByID :one
SELECT id, name, created_at, updated_at, status, model, temperature, max_tokens, settings
FROM sessions
WHERE id = ?;

-- name: ListSessions :many
SELECT id, name, created_at, updated_at, status, model, temperature, max_tokens, settings
FROM sessions
WHERE status != 'deleted'
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;

-- name: ListAllSessions :many
SELECT id, name, created_at, updated_at, status, model, temperature, max_tokens, settings
FROM sessions
ORDER BY updated_at DESC;

-- name: ListSessionsByStatus :many
SELECT id, name, created_at, updated_at, status, model, temperature, max_tokens, settings
FROM sessions
WHERE status = ?
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateSession :exec
UPDATE sessions
SET name = ?,
    model = ?,
    temperature = ?,
    max_tokens = ?,
    settings = ?
WHERE id = ? AND status != 'deleted';

-- name: UpdateSessionStatus :exec
UPDATE sessions
SET status = ?
WHERE id = ?;

-- name: ArchiveSession :exec
UPDATE sessions
SET status = 'archived'
WHERE id = ?;

-- name: DeleteSession :exec
UPDATE sessions
SET status = 'deleted'
WHERE id = ?;

-- name: HardDeleteSession :exec
DELETE FROM sessions
WHERE id = ?;

-- name: CountSessions :one
SELECT COUNT(*) as count
FROM sessions
WHERE status != 'deleted';

-- name: CountSessionsByStatus :one
SELECT COUNT(*) as count
FROM sessions
WHERE status = ?;

-- name: SearchSessions :many
SELECT id, name, created_at, updated_at, status, model, temperature, max_tokens, settings
FROM sessions
WHERE status != 'deleted'
  AND (name LIKE ? OR id LIKE ?)
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;

-- name: GetSessionWithMessageCount :one
SELECT
    s.id,
    s.name,
    s.created_at,
    s.updated_at,
    s.status,
    s.model,
    s.temperature,
    s.max_tokens,
    s.settings,
    COUNT(m.id) as message_count
FROM sessions s
LEFT JOIN messages m ON s.id = m.session_id
WHERE s.id = ? AND s.status != 'deleted'
GROUP BY s.id;

-- name: TouchSession :exec
UPDATE sessions
SET updated_at = CURRENT_TIMESTAMP
WHERE id = ?;
