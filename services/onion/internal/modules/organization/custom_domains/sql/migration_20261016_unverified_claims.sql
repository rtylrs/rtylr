SET @cd_verified_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'custom_domains'
      AND column_name = 'verified_domain'
);
SET @cd_add_verified_column_sql := IF(
    @cd_verified_column_exists,
    'SELECT 1',
    "ALTER TABLE custom_domains ADD COLUMN verified_domain varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci GENERATED ALWAYS AS (CASE WHEN status = 'verified' THEN domain ELSE NULL END) VIRTUAL AFTER updated_at"
);
PREPARE cd_add_verified_column FROM @cd_add_verified_column_sql;
EXECUTE cd_add_verified_column;
DEALLOCATE PREPARE cd_add_verified_column;

SET @cd_verified_index_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'custom_domains'
      AND index_name = 'uk_custom_domains_verified_domain'
);
SET @cd_add_verified_index_sql := IF(
    @cd_verified_index_exists,
    'SELECT 1',
    'ALTER TABLE custom_domains ADD UNIQUE KEY uk_custom_domains_verified_domain (verified_domain)'
);
PREPARE cd_add_verified_index FROM @cd_add_verified_index_sql;
EXECUTE cd_add_verified_index;
DEALLOCATE PREPARE cd_add_verified_index;

SET @cd_domain_index_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'custom_domains'
      AND index_name = 'idx_custom_domains_domain'
);
SET @cd_add_domain_index_sql := IF(
    @cd_domain_index_exists,
    'SELECT 1',
    'ALTER TABLE custom_domains ADD KEY idx_custom_domains_domain (domain)'
);
PREPARE cd_add_domain_index FROM @cd_add_domain_index_sql;
EXECUTE cd_add_domain_index;
DEALLOCATE PREPARE cd_add_domain_index;

SET @cd_old_unique_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'custom_domains'
      AND index_name = 'uk_custom_domains_domain'
);
SET @cd_drop_old_unique_sql := IF(
    @cd_old_unique_exists,
    'ALTER TABLE custom_domains DROP INDEX uk_custom_domains_domain',
    'SELECT 1'
);
PREPARE cd_drop_old_unique FROM @cd_drop_old_unique_sql;
EXECUTE cd_drop_old_unique;
DEALLOCATE PREPARE cd_drop_old_unique;
