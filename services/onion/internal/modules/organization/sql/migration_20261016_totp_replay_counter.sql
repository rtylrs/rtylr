SET @mfa_counter_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'user_mfa_enrollments'
      AND column_name = 'last_totp_counter'
);
SET @mfa_add_counter_column_sql := IF(
    @mfa_counter_column_exists,
    'SELECT 1',
    'ALTER TABLE user_mfa_enrollments ADD COLUMN last_totp_counter bigint NULL DEFAULT NULL AFTER revoked_at'
);
PREPARE mfa_add_counter_column FROM @mfa_add_counter_column_sql;
EXECUTE mfa_add_counter_column;
DEALLOCATE PREPARE mfa_add_counter_column;
