SET @key_prefix_width := (
    SELECT COALESCE(MAX(character_maximum_length), 0)
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'hub_api_keys'
      AND column_name = 'key_prefix'
);
SET @widen_key_prefix_sql := IF(
    @key_prefix_width > 0 AND @key_prefix_width < 32,
    'ALTER TABLE hub_api_keys MODIFY COLUMN key_prefix varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL',
    'SELECT 1'
);
PREPARE widen_key_prefix FROM @widen_key_prefix_sql;
EXECUTE widen_key_prefix;
DEALLOCATE PREPARE widen_key_prefix;
