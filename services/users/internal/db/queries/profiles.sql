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
