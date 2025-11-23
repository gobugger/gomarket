-- name: CreatePriceTier :one
INSERT INTO price_tiers (
	quantity, price_cent, product_id
) VALUES ($1, $2, $3) 
RETURNING *;

-- name: GetPriceTier :one
SELECT * FROM price_tiers
WHERE id = $1;

-- name: GetPriceTiers :many
SELECT * 
FROM price_tiers
WHERE product_id = $1 AND price_tiers.deleted_at IS NULL;

-- name: GetAllPriceTiers :many
SELECT *
FROM price_tiers
WHERE deleted_at IS NULL;

-- name: GetVendorForPriceTier :one
SELECT users.*
FROM price_tiers 
JOIN products ON products.id = price_tiers.product_id
JOIN users ON users.id = products.vendor_id	
WHERE price_tiers.id = $1;


-- name: GetPriceTiersForCart :many
SELECT price_tiers.*,
	COUNT(cart_items.price_id)
FROM cart_items
JOIN price_tiers ON price_tiers.id = cart_items.price_id
JOIN products ON products.id = price_tiers.product_id AND products.vendor_id = $2
WHERE cart_items.customer_id = $1
GROUP BY price_tiers.id, cart_items.customer_id;
