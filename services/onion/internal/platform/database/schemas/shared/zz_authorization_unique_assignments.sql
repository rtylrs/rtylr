-- Keep one pending or active assignment per target while retaining unlimited
-- revoked history for a role that is later granted again.
SET @legacy_index_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_role_assignments'
      AND index_name = 'uk_auth_active_direct_role'
);
SET @drop_legacy_index_sql := IF(
    @legacy_index_exists,
    'ALTER TABLE authorization_role_assignments DROP INDEX uk_auth_active_direct_role',
    'SELECT 1'
);
PREPARE drop_legacy_index_statement FROM @drop_legacy_index_sql;
EXECUTE drop_legacy_index_statement;
DEALLOCATE PREPARE drop_legacy_index_statement;

SET @active_user_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_role_assignments'
      AND column_name = 'active_user_id'
);
SET @add_active_user_column_sql := IF(
    @active_user_column_exists,
    'SELECT 1',
    "ALTER TABLE authorization_role_assignments ADD COLUMN active_user_id bigint GENERATED ALWAYS AS (CASE WHEN status IN ('pending','active') AND user_id IS NOT NULL THEN user_id ELSE NULL END) VIRTUAL AFTER status"
);
PREPARE add_active_user_column_statement FROM @add_active_user_column_sql;
EXECUTE add_active_user_column_statement;
DEALLOCATE PREPARE add_active_user_column_statement;

SET @active_group_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_role_assignments'
      AND column_name = 'active_group_id'
);
SET @add_active_group_column_sql := IF(
    @active_group_column_exists,
    'SELECT 1',
    "ALTER TABLE authorization_role_assignments ADD COLUMN active_group_id bigint GENERATED ALWAYS AS (CASE WHEN status IN ('pending','active') AND group_id IS NOT NULL THEN group_id ELSE NULL END) VIRTUAL AFTER active_user_id"
);
PREPARE add_active_group_column_statement FROM @add_active_group_column_sql;
EXECUTE add_active_group_column_statement;
DEALLOCATE PREPARE add_active_group_column_statement;

SET @active_user_index_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_role_assignments'
      AND index_name = 'uk_auth_active_user_role'
);
SET @add_active_user_index_sql := IF(
    @active_user_index_exists,
    'SELECT 1',
    'CREATE UNIQUE INDEX uk_auth_active_user_role ON authorization_role_assignments (business_id, active_user_id, app_id, role_id, status)'
);
PREPARE add_active_user_index_statement FROM @add_active_user_index_sql;
EXECUTE add_active_user_index_statement;
DEALLOCATE PREPARE add_active_user_index_statement;

SET @active_group_index_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_role_assignments'
      AND index_name = 'uk_auth_active_group_role'
);
SET @add_active_group_index_sql := IF(
    @active_group_index_exists,
    'SELECT 1',
    'CREATE UNIQUE INDEX uk_auth_active_group_role ON authorization_role_assignments (business_id, active_group_id, app_id, role_id, status)'
);
PREPARE add_active_group_index_statement FROM @add_active_group_index_sql;
EXECUTE add_active_group_index_statement;
DEALLOCATE PREPARE add_active_group_index_statement;
