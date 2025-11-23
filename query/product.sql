-- name: CreateProduct :one
INSERT INTO products (
	title, description, category_id, inventory, ships_from, ships_to, vendor_id
) VALUES (
	$1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetProduct :one
SELECT * FROM products
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetProducts :many
SELECT * FROM products
WHERE deleted_at IS NULL;

-- name: GetProductsForVendor :many
SELECT * FROM products
WHERE vendor_id = $1 AND deleted_at IS NULL;

-- name: DeleteProduct :exec
UPDATE products
SET deleted_at = NOW()
WHERE id = $1;

-- name: UpdateProductInventory :exec
UPDATE products
SET inventory = $2
WHERE id = $1;

-- name: ReduceProductInventory :exec
UPDATE products
SET inventory = inventory - sqlc.arg(amount)::int
WHERE id = $1 AND inventory >= sqlc.arg(amount)::int AND sqlc.arg(amount)::int >= 0;
