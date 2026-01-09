-- name: ListProjects :many
SELECT * FROM projects ORDER BY "order" ASC, created_at DESC LIMIT $1 OFFSET $2;

-- name: CountProjects :one
SELECT COUNT(*) FROM projects;

-- name: GetProjectByID :one
SELECT * FROM projects WHERE id = $1;

-- name: ListPublicProjects :many
SELECT * FROM projects ORDER BY "order" ASC, created_at DESC;

-- name: ListHighlightedProjects :many
SELECT * FROM projects WHERE highlighted = true ORDER BY "order" ASC, created_at DESC;

-- name: ListPublicProjectsPaginated :many
SELECT * FROM projects ORDER BY "order" ASC, created_at DESC LIMIT $1 OFFSET $2;

-- name: CreateProject :one
INSERT INTO projects (status, name, category, client, "order", highlighted, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING *;

-- name: UpdateProject :exec
UPDATE projects
SET status = $2,
    name = $3,
    category = $4,
    client = $5,
    "order" = $6,
    highlighted = $7,
    updated_at = NOW()
WHERE id = $1;

-- name: ToggleProjectHighlight :exec
UPDATE projects
SET highlighted = NOT highlighted,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = $1;
