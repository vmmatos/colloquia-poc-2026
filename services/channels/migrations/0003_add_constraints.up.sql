ALTER TABLE channels
  ADD CONSTRAINT chk_name_len        CHECK (length(name)        <= 80),
  ADD CONSTRAINT chk_description_len CHECK (length(description) <= 500);

ALTER TABLE channel_members
  ADD CONSTRAINT chk_member_role CHECK (role IN ('member', 'admin', 'owner'));
