-- name: CreateOrderChatMessage :one
INSERT INTO order_chat_messages (
	message, author_id, order_id
) VALUES (
	$1, $2, $3
) RETURNING *;

-- name: GetOrderChatMessagesForOrder :many
SELECT * FROM order_chat_messages
WHERE order_id = $1;

-- name: GetOrderChatMessagesForOrderJoinAuthor :many
SELECT sqlc.embed(chat), sqlc.embed(users)
FROM order_chat_messages AS chat
JOIN users ON users.id = chat.author_id
WHERE chat.order_id = $1
ORDER BY chat.created_at ASC;
