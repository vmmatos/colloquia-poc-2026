ALTER TABLE user_profiles
  ADD CONSTRAINT chk_email_len    CHECK (length(email)    <= 254),
  ADD CONSTRAINT chk_name_len     CHECK (length(name)     <= 100),
  ADD CONSTRAINT chk_avatar_len   CHECK (length(avatar)   <= 2048),
  ADD CONSTRAINT chk_bio_len      CHECK (length(bio)      <= 500),
  ADD CONSTRAINT chk_timezone_len CHECK (length(timezone) <= 64),
  ADD CONSTRAINT chk_status_len   CHECK (length(status)   <= 50),
  ADD CONSTRAINT chk_language_len CHECK (length(language) <= 10);
