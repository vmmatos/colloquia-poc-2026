ALTER TABLE messages
  ADD CONSTRAINT chk_content_len CHECK (length(content) <= 4000);
