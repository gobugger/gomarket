-- name: UpdateExchangeRates :exec
INSERT INTO exchange_rates (id, data)
VALUES (1, $1)
ON CONFLICT (id) DO UPDATE
SET data = EXCLUDED.data, 
	updated_at = NOW();

-- name: GetExchangeRates :one
SELECT * FROM exchange_rates WHERE id = 1;
