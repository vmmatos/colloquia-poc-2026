ALTER TABLE channels
  DROP CONSTRAINT IF EXISTS chk_name_len,
  DROP CONSTRAINT IF EXISTS chk_description_len;

ALTER TABLE channel_members
  DROP CONSTRAINT IF EXISTS chk_member_role;
