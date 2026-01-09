-- name: ListProjectImagesByProjectID :many
SELECT * FROM project_images WHERE project_id = $1 ORDER BY "order" ASC;

-- name: GetProjectImageByID :one
SELECT * FROM project_images WHERE id = $1;

-- name: CreateProjectImage :one
INSERT INTO project_images (name, url, project_id, "order", blur_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING *;

-- name: UpdateProjectImage :exec
UPDATE project_images
SET name = $2,
    url = $3,
    "order" = $4,
    blur_hash = $5,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteProjectImage :exec
DELETE FROM project_images WHERE id = $1;

-- name: DeleteProjectImagesByProjectID :many
SELECT * FROM project_images WHERE project_id = $1;

-- name: ListProjectImageIDsByProjectID :many
SELECT id FROM project_images WHERE project_id = $1;
