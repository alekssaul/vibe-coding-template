-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys
WHERE key_hash = ? LIMIT 1;

-- name: ListAPIKeys :many
SELECT * FROM api_keys
ORDER BY id DESC;

-- name: CreateAPIKey :one
INSERT INTO api_keys (name, key_hash, permission)
VALUES (?, ?, ?)
RETURNING *;

-- name: DeleteAPIKey :exec
DELETE FROM api_keys
WHERE id = ?;

-- name: CountAPIKeys :one
SELECT COUNT(*) FROM api_keys;
