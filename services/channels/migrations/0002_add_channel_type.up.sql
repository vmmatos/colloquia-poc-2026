ALTER TABLE channels ALTER COLUMN name DROP NOT NULL;
ALTER TABLE channels ADD COLUMN type TEXT NOT NULL DEFAULT 'channel'
  CHECK (type IN ('dm', 'group', 'channel'));
ALTER TABLE channels ADD COLUMN dm_key TEXT;
CREATE UNIQUE INDEX idx_channels_dm_key ON channels (dm_key) WHERE dm_key IS NOT NULL;
