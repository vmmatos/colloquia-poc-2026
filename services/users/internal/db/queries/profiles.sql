-- name: CreateUserProfile :one
INSERT INTO user_profiles (id, email) VALUES ($1, $2) RETURNING *;

-- name: GetUserProfile :one
SELECT * FROM user_profiles WHERE id = $1;

-- name: BatchGetUserProfiles :many
SELECT * FROM user_profiles WHERE id = ANY($1::uuid[]);

-- name: UpdateUserProfile :one
UPDATE user_profiles
SET name       = $2,
    avatar     = $3,
    bio        = $4,
    timezone   = $5,
    status     = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListUsers :many
SELECT * FROM user_profiles ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: SearchUsers :many
SELECT * FROM user_profiles
WHERE name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%'
ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: TouchLastSeen :exec
UPDATE user_profiles SET last_seen_at = NOW() WHERE id = $1;
