-- name: CreateToolExecution :exec
INSERT INTO tool_executions (
    id, message_id, tool_name, input, output, status, error, started_at, completed_at, duration_ms, retry_count, metadata
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetToolExecution :one
SELECT id, message_id, tool_name, input, output, status, error, started_at, completed_at, duration_ms, retry_count, metadata
FROM tool_executions
WHERE id = ?;

-- name: ListToolExecutionsByMessage :many
SELECT id, message_id, tool_name, input, output, status, error, started_at, completed_at, duration_ms, retry_count, metadata
FROM tool_executions
WHERE message_id = ?
ORDER BY started_at ASC;

-- name: ListToolExecutionsByStatus :many
SELECT id, message_id, tool_name, input, output, status, error, started_at, completed_at, duration_ms, retry_count, metadata
FROM tool_executions
WHERE status = ?
ORDER BY started_at DESC
LIMIT ? OFFSET ?;

-- name: ListToolExecutionsByName :many
SELECT id, message_id, tool_name, input, output, status, error, started_at, completed_at, duration_ms, retry_count, metadata
FROM tool_executions
WHERE tool_name = ?
ORDER BY started_at DESC
LIMIT ? OFFSET ?;

-- name: ListToolExecutionsBySession :many
SELECT te.id, te.message_id, te.tool_name, te.input, te.output, te.status, te.error, te.started_at, te.completed_at, te.duration_ms, te.retry_count, te.metadata
FROM tool_executions te
INNER JOIN messages m ON te.message_id = m.id
WHERE m.session_id = ?
ORDER BY te.started_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateToolExecutionStatus :exec
UPDATE tool_executions
SET status = ?,
    completed_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateToolExecutionOutput :exec
UPDATE tool_executions
SET output = ?,
    status = ?,
    completed_at = CURRENT_TIMESTAMP,
    duration_ms = ?
WHERE id = ?;

-- name: UpdateToolExecutionError :exec
UPDATE tool_executions
SET error = ?,
    status = 'failed',
    completed_at = CURRENT_TIMESTAMP,
    duration_ms = ?
WHERE id = ?;

-- name: IncrementRetryCount :exec
UPDATE tool_executions
SET retry_count = retry_count + 1
WHERE id = ?;

-- name: DeleteToolExecution :exec
DELETE FROM tool_executions
WHERE id = ?;

-- name: DeleteToolExecutionsByMessage :exec
DELETE FROM tool_executions
WHERE message_id = ?;

-- name: GetToolExecutionCount :one
SELECT COUNT(*) as count
FROM tool_executions
WHERE message_id = ?;

-- name: GetToolExecutionCountByStatus :one
SELECT COUNT(*) as count
FROM tool_executions
WHERE status = ?;

-- name: GetToolExecutionStats :one
SELECT
    COUNT(*) as total_executions,
    COUNT(CASE WHEN status = 'success' THEN 1 END) as successful_executions,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_executions,
    COUNT(CASE WHEN status = 'timeout' THEN 1 END) as timeout_executions,
    AVG(CASE WHEN duration_ms IS NOT NULL THEN duration_ms END) as avg_duration_ms,
    MAX(duration_ms) as max_duration_ms,
    MIN(duration_ms) as min_duration_ms
FROM tool_executions
WHERE tool_name = ?;

-- name: GetRecentToolExecutions :many
SELECT id, message_id, tool_name, input, output, status, error, started_at, completed_at, duration_ms, retry_count, metadata
FROM tool_executions
ORDER BY started_at DESC
LIMIT ?;

-- name: GetFailedToolExecutions :many
SELECT id, message_id, tool_name, input, output, status, error, started_at, completed_at, duration_ms, retry_count, metadata
FROM tool_executions
WHERE status = 'failed'
ORDER BY started_at DESC
LIMIT ? OFFSET ?;

-- name: GetPendingToolExecutions :many
SELECT id, message_id, tool_name, input, output, status, error, started_at, completed_at, duration_ms, retry_count, metadata
FROM tool_executions
WHERE status IN ('pending', 'running')
ORDER BY started_at ASC;

-- name: GetToolExecutionsByTimeRange :many
SELECT id, message_id, tool_name, input, output, status, error, started_at, completed_at, duration_ms, retry_count, metadata
FROM tool_executions
WHERE started_at >= ? AND started_at <= ?
ORDER BY started_at DESC;

-- name: GetToolUsageStats :many
SELECT
    tool_name,
    COUNT(*) as total_uses,
    COUNT(CASE WHEN status = 'success' THEN 1 END) as successful_uses,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_uses,
    AVG(CASE WHEN duration_ms IS NOT NULL THEN duration_ms END) as avg_duration_ms
FROM tool_executions
GROUP BY tool_name
ORDER BY total_uses DESC;
