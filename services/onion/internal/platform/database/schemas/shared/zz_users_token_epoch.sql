-- Per-user token epoch for server-side JWT revocation (#762). Access tokens
-- embed the epoch at issuance; the validator compares it against this column
-- and rejects mismatches. Bumping the column (whenever a password is set or
-- changed, or console-admin status flips) instantly invalidates every token
-- issued before it. Existing rows start at 0, matching tokens that predate
-- the claim, so no re-login is required until the first bump.

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'token_epoch') = 0,
  'ALTER TABLE users ADD COLUMN token_epoch BIGINT NOT NULL DEFAULT 0',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
