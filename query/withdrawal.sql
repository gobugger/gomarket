-- name: CreateWithdrawal :one
INSERT INTO withdrawals (
	amount_pico, destination_address, status
) VALUES(
	$1, $2, $3
) RETURNING *;

-- name: GetWithdrawalsWithStatus :many
SELECT * FROM withdrawals
WHERE status = $1;

-- name: UpdateWithdrawalStatus :one
UPDATE withdrawals
SET status = $2
WHERE id = $1
RETURNING *;

-- name: DeleteWithdrawal :exec
DELETE FROM withdrawals
WHERE id = $1;
