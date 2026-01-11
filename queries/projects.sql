-- name: ListProjects :many
SELECT * FROM projects ORDER BY "order" ASC, created_at DESC LIMIT $1 OFFSET $2;

-- name: ListProjectsWithSearch :many
SELECT * FROM projects 
WHERE ($1::text IS NULL OR $1::text = '' OR name ILIKE '%' || $1 || '%')
ORDER BY 
  CASE WHEN $2::text = 'order' AND $3::text = 'asc' THEN "order" ELSE NULL END ASC,
  CASE WHEN $2::text = 'order' AND $3::text = 'desc' THEN "order" ELSE NULL END DESC,
  CASE WHEN $2::text = 'name' AND $3::text = 'asc' THEN name ELSE NULL END ASC,
  CASE WHEN $2::text = 'name' AND $3::text = 'desc' THEN name ELSE NULL END DESC,
  CASE WHEN $2::text = 'created_at' AND $3::text = 'asc' THEN created_at ELSE NULL END ASC,
  CASE WHEN $2::text = 'created_at' AND $3::text = 'desc' THEN created_at ELSE NULL END DESC,
  CASE WHEN $2::text IS NULL OR $2::text = '' OR $2::text NOT IN ('order', 'name', 'created_at') THEN "order" ELSE NULL END ASC,
  created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountProjects :one
SELECT COUNT(*) FROM projects;

-- name: CountProjectsWithSearch :one
SELECT COUNT(*) FROM projects WHERE ($1::text IS NULL OR $1::text = '' OR name ILIKE '%' || $1 || '%');

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
