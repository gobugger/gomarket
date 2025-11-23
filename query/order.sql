-- name: CreateOrder :one
INSERT INTO orders (
	details, total_price_pico, customer_id, delivery_method_id, vendor_id
)
SELECT
	$1, $2, $3, $4, dm.vendor_id
FROM delivery_methods AS dm
WHERE dm.id = $4
RETURNING *;

-- name: GetOrder :one
SELECT * FROM orders
WHERE id = $1;

-- name: GetOrdersWithStatus :many
SELECT * FROM orders
WHERE status = $1;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2
WHERE id = $1 AND status = ANY(sqlc.arg(valid_statuses)::order_status[])
RETURNING *;

-- name: GetOrdersForCustomer :many
SELECT * FROM orders
WHERE customer_id = $1;

-- name: GetOrdersForVendor :many
SELECT orders.* FROM orders
WHERE orders.vendor_id = $1;

-- name: GetCustomerForOrder :one
SELECT users.*
FROM orders
JOIN users ON users.id = orders.customer_id
WHERE orders.id = $1;

-- name: GetVendorForOrder :one
SELECT users.*
FROM orders 
JOIN users ON users.id = orders.vendor_id	
WHERE orders.id = $1;

-- name: ExtendOrder :one
UPDATE orders
SET num_extends = num_extends + 1
WHERE id = $1 AND num_extends < 2
RETURNING num_extends;

-- name: GetOrdersWithStatuses :many
SELECT *
FROM orders
WHERE status = ANY(sqlc.arg(statuses)::order_status[]);

-- name: DeleteOrder :exec
DELETE FROM orders WHERE id = $1;

-- name: CreateOrderInvoice :one
INSERT INTO order_invoices (
	order_id, invoice_id
)
VALUES (
	$1, $2
)
RETURNING *;

-- name: GetOrdersByStatusXInvoiceStatus :many
SELECT orders.*
FROM order_invoices
JOIN orders ON orders.id = order_invoices.order_id AND orders.status = sqlc.arg(order_status)::order_status
JOIN invoices ON invoices.id = order_invoices.invoice_id AND invoices.status = sqlc.arg(invoice_status)::invoice_status;

-- name: CreateOrderItems :copyfrom
INSERT INTO order_items (
	order_id, price_id, count
) VALUES (
	$1, $2, $3
);

-- name: GetOrderItems :many
SELECT order_items.*, price_tiers.product_id, price_tiers.quantity, price_tiers.price_cent
FROM order_items
JOIN price_tiers ON price_tiers.id = order_items.price_id
WHERE order_id = $1;
