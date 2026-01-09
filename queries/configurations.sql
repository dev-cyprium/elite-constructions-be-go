-- name: ListAllConfigurations :many
SELECT * FROM configurations ORDER BY created_at DESC;

-- name: GetConfigurationByKey :one
SELECT * FROM configurations WHERE key = $1;

-- name: UpdateConfigurationValue :exec
UPDATE configurations
SET value = $2,
    updated_at = NOW()
WHERE key = $1;
