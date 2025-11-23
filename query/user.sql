-- name: CreateUser :one
INSERT INTO users (
	username, password_hash, pgp_key
) VALUES (
	$1, $2, $3
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserWithName :one
SELECT * FROM users
WHERE username = $1;

-- name: UpdateUserPasswordHash :one
UPDATE users 
SET password_hash = sqlc.arg(new_password_hash) 
WHERE username = $1 AND password_hash = $2
RETURNING *;

-- name: UpdateUserPgpKey :one
UPDATE users 
SET pgp_key = $2 
WHERE id = $1
RETURNING *;

-- name: UpdateUserPrevLogin :one
UPDATE users 
SET prev_login = NOW() 
WHERE id = $1
RETURNING *;

-- name: UpdateUserSettings :exec
UPDATE users
SET locale = $2, currency = $3, twofa_enabled = $4, incognito_enabled = $5
WHERE id = $1
RETURNING *;

-- name: GetNumberOfUsers :one
SELECT COUNT(*) FROM users;
