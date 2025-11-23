-- name: CreateTicket :one
INSERT INTO tickets (
	subject, message, author_id
) VALUES(
	$1, $2, $3
) RETURNING *;

-- name: GetTicket :one
SELECT * FROM tickets
WHERE id = $1;

-- name: GetOpenTickets :many
SELECT * FROM tickets
WHERE is_open = true;

-- name: GetTicketsForAuthor :many
SELECT * FROM tickets
WHERE author_id = $1;

-- name: CloseTicket :one
UPDATE tickets
SET is_open = false
WHERE id = $1
RETURNING *;
