-- name: CreateChannel :one
INSERT INTO channels (name, description, is_private, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetChannelByID :one
SELECT * FROM channels
WHERE id = $1;

-- name: ArchiveChannel :exec
UPDATE channels
SET archived = TRUE, updated_at = NOW()
WHERE id = $1;

-- name: AddChannelMember :one
INSERT INTO channel_members (channel_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: RemoveChannelMember :exec
DELETE FROM channel_members
WHERE channel_id = $1 AND user_id = $2;

-- name: GetChannelMember :one
SELECT * FROM channel_members
WHERE channel_id = $1 AND user_id = $2;

-- name: ListUserChannels :many
SELECT
    c.*,
    (SELECT COUNT(*) FROM channel_members m WHERE m.channel_id = c.id)::int AS member_count
FROM channels c
JOIN channel_members cm ON c.id = cm.channel_id
WHERE cm.user_id = $1
ORDER BY c.created_at DESC;

-- name: ListChannelMembers :many
SELECT * FROM channel_members
WHERE channel_id = $1
ORDER BY joined_at ASC;

-- name: CountChannelMembers :one
SELECT COUNT(*)::int FROM channel_members
WHERE channel_id = $1;
