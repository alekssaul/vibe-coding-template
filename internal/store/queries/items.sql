-- name: GetItem :one
SELECT * FROM items
WHERE id = ? LIMIT 1;

-- name: ListItems :many
SELECT * FROM items
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchItems :many
SELECT * FROM items
WHERE name LIKE '%' || ? || '%' OR description LIKE '%' || ? || '%'
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: CountItems :one
SELECT COUNT(*) FROM items;

-- name: CountItemsSearch :one
SELECT COUNT(*) FROM items
WHERE name LIKE '%' || ? || '%' OR description LIKE '%' || ? || '%';

-- name: CreateItem :one
INSERT INTO items (name, description)
VALUES (?, ?)
RETURNING *;

-- name: UpdateItem :one
UPDATE items
SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteItem :exec
DELETE FROM items
WHERE id = ?;
