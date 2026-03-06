CREATE TABLE user_profiles (
    id         UUID PRIMARY KEY,
    email      TEXT UNIQUE NOT NULL,
    name       TEXT NOT NULL DEFAULT '',
    avatar     TEXT NOT NULL DEFAULT '',
    bio        TEXT NOT NULL DEFAULT '',
    timezone   TEXT NOT NULL DEFAULT 'UTC',
    status     TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_profiles_email  ON user_profiles(email);
CREATE INDEX idx_user_profiles_status ON user_profiles(status);
