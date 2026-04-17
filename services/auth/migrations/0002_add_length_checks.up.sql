ALTER TABLE users
  ADD CONSTRAINT chk_email_len    CHECK (length(email)         <= 254),
  ADD CONSTRAINT chk_password_len CHECK (length(password_hash) <= 512);
