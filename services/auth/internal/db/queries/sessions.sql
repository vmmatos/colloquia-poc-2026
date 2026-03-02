-- name: CreateSession :one
INSERT INTO sessions (
  id,
  user_id,
  refresh_token_hash,
  access_token_hash,
  expires_at,
  revoked
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: FindSessionByAccessTokenHash :one
SELECT * FROM sessions WHERE access_token_hash = $1 AND revoked = false;

-- name: FindSessionById :one
SELECT * FROM sessions WHERE id = $1;

-- name: RevokeSession :exec
UPDATE sessions SET revoked = true WHERE id = $1;
