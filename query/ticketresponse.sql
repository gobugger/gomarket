-- name: CreateTicketResponse :one
INSERT INTO ticket_responses (
	message, ticket_id, author_name
) VALUES(
	$1, $2, $3
) RETURNING *;

-- name: GetTicketResponsesForTicket :many
SELECT * FROM ticket_responses
WHERE ticket_id = $1
ORDER BY created_at ASC;
