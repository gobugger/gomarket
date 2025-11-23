-- name: CreateTransaction :one
INSERT INTO transactions (
	hash
) VALUES (
	$1
) RETURNING *;

-- name: GetTransactions :many
SELECT * FROM transactions;

-- name: GetTransactionsBefore :many
SELECT * FROM transactions
WHERE created_at < sqlc.arg(t);

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = $1;
