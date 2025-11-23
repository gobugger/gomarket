-- name: CreateDisputeOffer :one
INSERT INTO dispute_offers (
	refund_factor, order_id
) SELECT $1, $2
FROM orders
WHERE orders.id = $2 AND orders.status = 'disputed'
RETURNING *;

-- name: CreateDisputeOfferWithStatus :one
INSERT INTO dispute_offers (
	refund_factor, order_id, status
) SELECT $1, $2, $3
FROM orders
WHERE orders.id = $2 AND orders.status = 'disputed'
RETURNING *;


-- name: GetDisputeOffer :one
SELECT * FROM dispute_offers
WHERE id = $1;

-- name: GetDisputeOffersForOrder :many
SELECT * FROM dispute_offers
WHERE order_id = $1;

-- name: UpdateDisputeOfferStatus :one
UPDATE dispute_offers
SET status = $2
WHERE id = $1
RETURNING *;
