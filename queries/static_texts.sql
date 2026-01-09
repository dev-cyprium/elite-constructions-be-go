-- name: ListStaticTexts :many
SELECT * FROM static_texts ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CountStaticTexts :one
SELECT COUNT(*) FROM static_texts;

-- name: ListAllStaticTexts :many
SELECT * FROM static_texts ORDER BY created_at DESC;

-- name: GetStaticTextByID :one
SELECT * FROM static_texts WHERE id = $1;

-- name: UpdateStaticTextContent :exec
UPDATE static_texts
SET content = $2,
    updated_at = NOW()
WHERE id = $1;
