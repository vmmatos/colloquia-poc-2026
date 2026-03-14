CREATE TABLE channels (
    id           UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT      NOT NULL,
    description  TEXT      NOT NULL DEFAULT '',
    is_private   BOOLEAN   NOT NULL DEFAULT FALSE,
    created_by   UUID      NOT NULL,
    archived     BOOLEAN   NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE channel_members (
    channel_id  UUID      NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id     UUID      NOT NULL,
    role        TEXT      NOT NULL DEFAULT 'member',
    joined_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (channel_id, user_id)
);

CREATE INDEX idx_channel_members_user    ON channel_members(user_id);
CREATE INDEX idx_channel_members_channel ON channel_members(channel_id);
CREATE INDEX idx_channels_is_private     ON channels(is_private);
CREATE INDEX idx_channels_created_by     ON channels(created_by);
