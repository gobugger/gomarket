-- name: CreateDeposit :one
INSERT INTO deposits (
	wallet_id, invoice_id
) VALUES(
	$1, $2
) RETURNING *;

-- name: GetOutdatedDeposits :many
SELECT deposits.*, invoices.*
FROM deposits
JOIN invoices ON invoices.id = deposits.invoice_id
WHERE deposits.amount_deposited_pico < invoices.amount_unlocked_pico
FOR UPDATE;

-- name: UpdateAmountDeposited :one
UPDATE deposits
SET amount_deposited_pico = $2
WHERE id = $1
RETURNING *;

-- name: GetDepositForWallet :one
SELECT sqlc.embed(wallets), sqlc.embed(invoices)
FROM deposits
JOIN wallets ON wallets.id = deposits.wallet_id
JOIN invoices ON invoices.id = deposits.invoice_id
WHERE deposits.wallet_id = $1;

-- name: GetDepositForUser :one
SELECT sqlc.embed(wallets), sqlc.embed(invoices)
FROM deposits
JOIN wallets ON wallets.id = deposits.wallet_id
JOIN invoices ON invoices.id = deposits.invoice_id
WHERE wallets.user_id = $1;
