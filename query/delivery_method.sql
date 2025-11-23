-- name: CreateDeliveryMethod :one
INSERT INTO delivery_methods (
	description, price_cent, vendor_id
) VALUES (
	$1, $2, $3
) RETURNING *;

-- name: DeleteDeliveryMethod :exec
UPDATE delivery_methods 
SET deleted_at = NOW()
WHERE id = $1;

-- name: GetDeliveryMethod :one
SELECT * 
FROM delivery_methods
WHERE id = $1;

-- name: GetDeliveryMethodForOrder :one
SELECT delivery_methods.* 
FROM delivery_methods
JOIN orders ON orders.delivery_method_id = delivery_methods.id
WHERE orders.id = $1;

-- name: GetDeliveryMethodsForProduct :many
SELECT delivery_methods.* 
FROM delivery_methods
JOIN products ON products.vendor_id = delivery_methods.vendor_id
WHERE products.id = $1 AND delivery_methods.deleted_at IS NULL;

-- name: GetDeliveryMethodsForVendor :many
SELECT * 
FROM delivery_methods
WHERE delivery_methods.vendor_id = $1 AND delivery_methods.deleted_at IS NULL;

-- name: GetAllDeliveryMethods :many
SELECT * 
FROM delivery_methods
WHERE delivery_methods.deleted_at IS NULL;
