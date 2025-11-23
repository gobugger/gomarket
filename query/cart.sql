-- name: CreateCartItem :one
INSERT INTO cart_items (
	customer_id, price_id
) VALUES (
	$1, $2
) RETURNING *;

-- name: ClearCart :exec
DELETE FROM cart_items
WHERE customer_id = $1;

-- name: GetCartItems :many
SELECT * FROM cart_items
WHERE customer_id = $1;

-- name: GetCartItemsByVendor :many
SELECT * 
FROM cart_items
JOIN price_tiers ON price_tiers.id = cart_items.price_id
JOIN products ON products.id = price_tiers.product_id AND products.vendor_id = $2
WHERE cart_items.customer_id = $1;

-- name: GetCartPriceTiersForCustomer :many
SELECT price_tiers.* 
FROM price_tiers
JOIN products ON products.id = price_tiers.product_id AND products.vendor_id = $2
WHERE price_tiers.id IN (SELECT price_id FROM cart_items WHERE customer_id = $1)
FOR UPDATE;

-- name: SetCartItemInvalid :exec
UPDATE cart_items
SET invalid = true
WHERE id = $1;

-- name: GetNumCartItemsForCustomer :one
SELECT COUNT(*)
FROM cart_items
WHERE cart_items.customer_id = $1;

-- name: RemoveCartItem :exec
DELETE FROM cart_items
WHERE id = (
	SELECT id
	FROM cart_items
	WHERE cart_items.customer_id = $1 AND cart_items.price_id = $2
	LIMIT 1
);

-- name: RemoveCartItems :exec
DELETE FROM cart_items
WHERE cart_items.customer_id = $1 AND $2 = (
	SELECT products.vendor_id
	FROM price_tiers
	JOIN products ON products.id = price_tiers.product_id
	WHERE price_tiers.id = cart_items.price_id
);
