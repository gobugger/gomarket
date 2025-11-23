-- name: CreateInvoice :one
INSERT INTO invoices (
	amount_pico, permanent
) VALUES(
	$1, $2
) RETURNING *;

-- name: GetInvoice :one
SELECT * FROM invoices 
WHERE id = $1;

-- name: GetInvoicesWithoutAddress :many
SELECT invoices.* 
FROM invoices
WHERE address = '';

-- name: SetInvoiceAddress :one
UPDATE invoices
SET address = $2
WHERE id = $1 AND address = '' AND LENGTH($2) = 95
RETURNING *;

-- name: GetPendingInvoices :many
SELECT invoices.* 
FROM invoices
WHERE status = 'pending'::invoice_status AND address != '';

-- name: UpdateInvoiceStatus :one
UPDATE invoices
SET status = $2
WHERE id = $1
RETURNING *;

-- name: UpdateInvoiceAmountUnlocked :one
UPDATE invoices
SET amount_unlocked_pico = $2
WHERE id = $1
RETURNING *;

-- name: DeleteInvoice :exec
DELETE FROM invoices
WHERE id = $1;

-- name: GetInvoiceForOrder :one
SELECT invoices.* 
FROM order_invoices
JOIN invoices ON invoices.id = order_invoices.invoice_id
WHERE order_invoices.order_id = $1;

-- name: GetUnpreparedInvoices :many
SELECT invoices.*
FROM invoices
WHERE status = 'pending'::invoice_status AND address = ''
FOR UPDATE;

-- name: GetInvoiceForWallet :one
SELECT invoices.*
FROM deposits
JOIN invoices ON invoices.id = deposits.invoice_id
WHERE deposits.wallet_id = $1;
