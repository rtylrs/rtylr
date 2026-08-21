-- POS PINs are bcrypt hashes. The legacy plaintext column is no longer read
-- or written and must not remain available to application queries.
SET @sql := IF(
    EXISTS (
        SELECT 1
        FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE()
          AND TABLE_NAME = 'pos_users'
          AND COLUMN_NAME = 'pin'
    ),
    'ALTER TABLE pos_users DROP COLUMN pin',
    'SELECT 1'
);
PREPARE drop_legacy_pos_pin FROM @sql;
EXECUTE drop_legacy_pos_pin;
DEALLOCATE PREPARE drop_legacy_pos_pin;
