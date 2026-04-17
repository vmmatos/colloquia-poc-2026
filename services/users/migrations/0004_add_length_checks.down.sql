ALTER TABLE user_profiles
  DROP CONSTRAINT IF EXISTS chk_email_len,
  DROP CONSTRAINT IF EXISTS chk_name_len,
  DROP CONSTRAINT IF EXISTS chk_avatar_len,
  DROP CONSTRAINT IF EXISTS chk_bio_len,
  DROP CONSTRAINT IF EXISTS chk_timezone_len,
  DROP CONSTRAINT IF EXISTS chk_status_len,
  DROP CONSTRAINT IF EXISTS chk_language_len;
