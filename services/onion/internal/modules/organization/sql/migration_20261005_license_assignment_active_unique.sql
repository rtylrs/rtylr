SET @license_index_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_license_assignments'
      AND index_name = 'uk_auth_active_direct_license'
);
SET @drop_license_index_sql := IF(
    @license_index_exists,
    'ALTER TABLE authorization_license_assignments DROP INDEX uk_auth_active_direct_license',
    'SELECT 1'
);
PREPARE drop_license_index_statement FROM @drop_license_index_sql;
EXECUTE drop_license_index_statement;
DEALLOCATE PREPARE drop_license_index_statement;

SET @license_active_user_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_license_assignments'
      AND column_name = 'active_user_id'
);
SET @add_license_active_user_column_sql := IF(
    @license_active_user_column_exists,
    'SELECT 1',
    "ALTER TABLE authorization_license_assignments ADD COLUMN active_user_id bigint GENERATED ALWAYS AS (CASE WHEN status IN ('pending','active') AND user_id IS NOT NULL THEN user_id ELSE NULL END) VIRTUAL AFTER status"
);
PREPARE add_license_active_user_column_statement FROM @add_license_active_user_column_sql;
EXECUTE add_license_active_user_column_statement;
DEALLOCATE PREPARE add_license_active_user_column_statement;

SET @license_active_user_index_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_license_assignments'
      AND index_name = 'uk_auth_active_direct_license'
);
SET @add_license_active_user_index_sql := IF(
    @license_active_user_index_exists,
    'SELECT 1',
    'CREATE UNIQUE INDEX uk_auth_active_direct_license ON authorization_license_assignments (business_id, active_user_id, app_id, status)'
);
PREPARE add_license_active_user_index_statement FROM @add_license_active_user_index_sql;
EXECUTE add_license_active_user_index_statement;
DEALLOCATE PREPARE add_license_active_user_index_statement;
