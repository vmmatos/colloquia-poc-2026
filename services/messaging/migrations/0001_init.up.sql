CREATE TABLE messages (
    id         UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID      NOT NULL,
    user_id    UUID      NOT NULL,
    content    TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_channel_created ON messages (channel_id, created_at DESC);
