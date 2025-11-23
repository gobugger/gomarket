-- name: CreateReview :one
INSERT INTO reviews (
    grade, comment, order_id
) 
VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetReviewsForVendor :many
SELECT sqlc.embed(reviews), sqlc.embed(users)
FROM orders
JOIN reviews ON reviews.order_id = orders.id
JOIN users ON users.id = orders.customer_id
WHERE orders.vendor_id = $1;

-- name: DeleteReview :exec
DELETE FROM reviews
WHERE id = $1;

-- name: CreateProductReview :one
INSERT INTO product_reviews (
    grade, comment, order_item_id
) 
VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetReviewsForProduct :many
SELECT sqlc.embed(product_reviews),
	price_tiers.quantity * order_items.count AS total_quantity,
	price_tiers.price_cent * order_items.count AS total_price_cent,
	users.username AS author_name,
	reviews.created_at AS created_at
FROM product_reviews
JOIN order_items ON order_items.id = product_reviews.order_item_id
JOIN price_tiers ON price_tiers.id = order_items.price_id
JOIN orders ON orders.id = order_items.order_id
JOIN reviews ON reviews.order_id = orders.id
JOIN users ON users.id = orders.customer_id
WHERE price_tiers.product_id = $1;
