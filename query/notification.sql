-- name: CreateNotification :one
INSERT INTO notifications (
	title, content, user_id
) VALUES (
	$1, $2, $3
) RETURNING *;

-- name: GetCountOfUnseenNotificationsForUser :one
SELECT COUNT(*)
FROM notifications
WHERE user_id = $1 AND is_seen = false;

-- name: GetNotificationsForUser :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = $1;

-- name: DeleteAllNotificationsForUser :exec
DELETE FROM notifications
WHERE user_id = $1;
