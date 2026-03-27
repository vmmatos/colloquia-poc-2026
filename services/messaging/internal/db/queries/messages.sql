-- name: InsertMessage :one
INSERT INTO messages (channel_id, user_id, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListMessagesFirst :many
SELECT m.id, m.channel_id, m.user_id, m.content, m.created_at
FROM messages m
WHERE m.channel_id = $1
ORDER BY m.created_at DESC
LIMIT $2;

-- name: ListMessagesFromCursor :many
SELECT m.id, m.channel_id, m.user_id, m.content, m.created_at
FROM messages m
WHERE m.channel_id = $1
  AND m.created_at < (SELECT m2.created_at FROM messages m2 WHERE m2.id = $2)
ORDER BY m.created_at DESC
LIMIT $3;
