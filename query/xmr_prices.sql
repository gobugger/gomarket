-- name: UpdateXMRPrices :exec
INSERT INTO xmr_prices (id, data)
VALUES (1, $1)
ON CONFLICT (id) DO UPDATE
SET data = EXCLUDED.data, 
	updated_at = NOW();

-- name: GetXMRPrices :one
SELECT * FROM xmr_prices WHERE id = 1;
