-- name: UpdateSettings :exec
UPDATE settings
SET data = $1, updated_at = NOW()
WHERE id = 1;

-- name: GetSettings :one
SELECT data FROM settings WHERE id = 1;
