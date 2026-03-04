-- name: CreateUser :one
INSERT INTO users (
  id,
  email,
  password_hash
) VALUES ($1, $2, $3)
RETURNING *;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: FindUserById :one
SELECT * FROM users WHERE id = $1;

-- name: IncrementFailedLoginAttempts :one
UPDATE users
SET failed_login_attempts = failed_login_attempts + 1,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: LockUser :exec
UPDATE users
SET locked_until = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: ResetFailedLoginAttempts :exec
UPDATE users
SET failed_login_attempts = 0,
    locked_until = NULL,
    updated_at = NOW()
WHERE id = $1;
