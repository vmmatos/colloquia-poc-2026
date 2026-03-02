DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

DROP INDEX IF EXISTS idx_sessions_access_hash;
DROP INDEX IF EXISTS idx_sessions_refresh_hash;
DROP INDEX IF EXISTS idx_sessions_user_id;
