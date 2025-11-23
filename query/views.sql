-- name: GetViewOrder :one
SELECT sqlc.embed(orders), 
	sqlc.embed(dm), 
	sqlc.embed(customer),
	sqlc.embed(vendor)
FROM orders
JOIN delivery_methods AS dm ON dm.id = orders.delivery_method_id
JOIN users AS customer ON customer.id = orders.customer_id
JOIN users AS vendor ON vendor.id = orders.vendor_id
WHERE orders.id = $1;

-- name: GetViewOrdersForVendor :many
SELECT sqlc.embed(orders), 
	sqlc.embed(dm),
	sqlc.embed(customer),
	sqlc.embed(vendor)
FROM orders
JOIN delivery_methods AS dm ON dm.id = orders.delivery_method_id
JOIN users AS customer ON customer.id = orders.customer_id
JOIN users AS vendor ON vendor.id = orders.vendor_id
WHERE orders.vendor_id = $1;

-- name: GetViewOrdersForCustomer :many
SELECT sqlc.embed(orders), 
	sqlc.embed(dm),
	sqlc.embed(customer),
	sqlc.embed(vendor)
FROM orders
JOIN delivery_methods AS dm ON dm.id = orders.delivery_method_id
JOIN users AS customer ON customer.id = orders.customer_id
JOIN users AS vendor ON vendor.id = orders.vendor_id
WHERE orders.customer_id = $1;

-- name: GetViewCartForCustomer :many
SELECT price_tiers.id AS price_tier_id,
	sqlc.embed(products),
	sqlc.embed(vendors),
	price_tiers.price_cent,
	price_tiers.quantity,
	COUNT(cart_items.price_id)
FROM cart_items
JOIN price_tiers ON price_tiers.id = cart_items.price_id
JOIN products ON products.id = price_tiers.product_id
JOIN users AS vendors ON vendors.id = products.vendor_id
WHERE cart_items.customer_id = $1
GROUP BY price_tiers.id, products.id, vendors.id, price_tiers.price_cent, price_tiers.quantity;

-- name: GetViewCartForCustomerByVendor :many
SELECT price_tiers.id AS price_tier_id,
	sqlc.embed(products),
	sqlc.embed(vendors),
	price_tiers.price_cent,
	price_tiers.quantity,
	COUNT(cart_items.price_id)
FROM cart_items
JOIN price_tiers ON price_tiers.id = cart_items.price_id
JOIN products ON products.id = price_tiers.product_id AND products.vendor_id = $2
JOIN users AS vendors ON vendors.id = products.vendor_id
WHERE cart_items.customer_id = $1
GROUP BY price_tiers.id, products.id, vendors.id, price_tiers.price_cent, price_tiers.quantity;
