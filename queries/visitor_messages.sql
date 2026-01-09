-- name: ListVisitorMessages :many
SELECT * FROM visitor_messages ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CountVisitorMessages :one
SELECT COUNT(*) FROM visitor_messages;

-- name: CreateVisitorMessage :one
INSERT INTO visitor_messages (email, address, description, seen, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING *;

-- name: DeleteVisitorMessage :exec
DELETE FROM visitor_messages WHERE id = $1;
