-- name: CreateBan :one
INSERT INTO bans (
	user_id
) VALUES(
	$1
) RETURNING *;

-- name: GetBanForUser :one
SELECT * FROM bans
WHERE user_id = $1;

-- name: DeleteBan :exec
DELETE FROM bans WHERE id = $1;
