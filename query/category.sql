-- name: CreateCategory :one
INSERT INTO categories (
	parent_id, name
) VALUES (
	$1, $2
) RETURNING *;

-- name: GetCategories :many
SELECT * FROM categories;

-- name: GetCategory :one
SELECT *
FROM categories
WHERE id = $1;

-- name: DeleteCategory :exec
UPDATE categories
SET deleted_at = NOW()
WHERE id = $1;
