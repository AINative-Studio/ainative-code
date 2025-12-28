-- name: GetMetadata :one
SELECT key, value, created_at, updated_at
FROM metadata
WHERE key = ?;

-- name: ListMetadata :many
SELECT key, value, created_at, updated_at
FROM metadata
ORDER BY key;

-- name: SetMetadata :exec
INSERT OR REPLACE INTO metadata (key, value)
VALUES (?, ?);

-- name: DeleteMetadata :exec
DELETE FROM metadata
WHERE key = ?;

-- name: MetadataExists :one
SELECT COUNT(*) > 0 as key_exists
FROM metadata
WHERE key = ?;
