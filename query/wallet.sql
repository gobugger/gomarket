-- name: CreateWallet :one
INSERT INTO wallets (
	user_id
) VALUES(
	$1
) RETURNING *;

-- name: GetWallet :one
SELECT * FROM wallets
WHERE id = $1;

-- name: GetWallets :many
SELECT * FROM wallets;

-- name: GetWalletForUser :one
SELECT wallets.* 
FROM wallets
JOIN users ON users.id = wallets.user_id
WHERE users.id = $1;

-- name: AddWalletBalance :one
UPDATE wallets
SET balance_pico = balance_pico + sqlc.arg(amount)
WHERE id = $1 AND sqlc.arg(amount)::bigint >= 0
RETURNING *;

-- name: ReduceWalletBalance :one
UPDATE wallets
SET balance_pico = balance_pico - sqlc.arg(amount)::numeric
WHERE id = $1 AND balance_pico >= sqlc.arg(amount)::numeric AND sqlc.arg(amount)::numeric >= 0
RETURNING *;
