DELETE scope
FROM authorization_role_assignment_scopes scope
JOIN authorization_role_assignments assignment ON assignment.id = scope.assignment_id
WHERE assignment.source = 'dash_default';

DELETE FROM authorization_role_assignments WHERE source = 'dash_default';
DELETE FROM authorization_license_assignments WHERE source = 'dash_default';

ALTER TABLE authorization_role_assignments
  MODIFY COLUMN source enum('explicit','group','system') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'explicit';

ALTER TABLE authorization_license_assignments
  MODIFY COLUMN source enum('explicit','group','system') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'explicit';

SET @member_type_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_business_members'
      AND column_name = 'member_type'
);
SET @add_member_type_sql := IF(
    @member_type_column_exists,
    'SELECT 1',
    "ALTER TABLE authorization_business_members ADD COLUMN member_type enum('member','admin') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'member' AFTER user_id"
);
PREPARE add_member_type_statement FROM @add_member_type_sql;
EXECUTE add_member_type_statement;
DEALLOCATE PREPARE add_member_type_statement;

SET @dash_role_key_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_business_members'
      AND column_name = 'dash_role_key'
);
SET @backfill_member_type_sql := IF(
    @dash_role_key_column_exists,
    "UPDATE authorization_business_members SET member_type = 'admin' WHERE dash_role_key IN ('admin','owner')",
    'SELECT 1'
);
PREPARE backfill_member_type_statement FROM @backfill_member_type_sql;
EXECUTE backfill_member_type_statement;
DEALLOCATE PREPARE backfill_member_type_statement;

SET @legacy_user_permissions_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.tables
    WHERE table_schema = DATABASE()
      AND table_name = 'user_app_permissions'
);

-- The previous backfill ran while the old app was still writing this table.
-- Replay only missing normalized grants before removing that write path.
SET @backfill_members_sql := IF(
    @legacy_user_permissions_exists,
    "INSERT INTO authorization_business_members (business_id, user_id, member_type, status) SELECT DISTINCT legacy.business_id, legacy.user_id, IF(legacy.role IN ('owner','admin'), 'admin', 'member'), IF(legacy.is_active, 'active', 'revoked') FROM user_app_permissions legacy JOIN users account ON account.id = legacy.user_id JOIN businesses business ON business.id = legacy.business_id AND business.owner_id = legacy.owner_id WHERE legacy.business_id > 0 ON DUPLICATE KEY UPDATE member_type = IF(member_type = 'admin' OR VALUES(member_type) = 'admin', 'admin', 'member')",
    'SELECT 1'
);
PREPARE backfill_members_statement FROM @backfill_members_sql;
EXECUTE backfill_members_statement;
DEALLOCATE PREPARE backfill_members_statement;

SET @backfill_licenses_sql := IF(
    @legacy_user_permissions_exists,
    "INSERT INTO authorization_license_assignments (business_id, user_id, app_id, source, status) SELECT DISTINCT legacy.business_id, legacy.user_id, CONVERT(apps.app_id USING utf8mb4) COLLATE utf8mb4_unicode_ci, 'explicit', IF(legacy.is_active, 'active', 'revoked') FROM user_app_permissions legacy JOIN users account ON account.id = legacy.user_id JOIN businesses business ON business.id = legacy.business_id AND business.owner_id = legacy.owner_id JOIN JSON_TABLE(legacy.apps, '$[*]' COLUMNS(app_id varchar(32) PATH '$')) apps WHERE legacy.business_id > 0 AND BINARY apps.app_id IN (BINARY 'pos', BINARY 'erp', BINARY 'hr', BINARY 'crm', BINARY 'finance', BINARY 'flow', BINARY 'insights', BINARY 'menu', BINARY 'recruit') AND NOT EXISTS (SELECT 1 FROM authorization_license_assignments existing WHERE existing.business_id = legacy.business_id AND existing.user_id = legacy.user_id AND existing.group_id IS NULL AND existing.app_id = CONVERT(apps.app_id USING utf8mb4) COLLATE utf8mb4_unicode_ci)",
    'SELECT 1'
);
PREPARE backfill_licenses_statement FROM @backfill_licenses_sql;
EXECUTE backfill_licenses_statement;
DEALLOCATE PREPARE backfill_licenses_statement;

SET @backfill_roles_sql := IF(
    @legacy_user_permissions_exists,
    "INSERT INTO authorization_role_assignments (business_id, user_id, app_id, role_id, source, status, is_primary) SELECT DISTINCT legacy.business_id, legacy.user_id, profile.app_id, profile.id, 'explicit', IF(legacy.is_active, 'active', 'revoked'), TRUE FROM user_app_permissions legacy JOIN users account ON account.id = legacy.user_id JOIN businesses business ON business.id = legacy.business_id AND business.owner_id = legacy.owner_id JOIN JSON_TABLE(legacy.apps, '$[*]' COLUMNS(app_id varchar(32) PATH '$')) apps JOIN authorization_roles profile ON profile.business_id = 0 AND profile.app_id = CONVERT(apps.app_id USING utf8mb4) COLLATE utf8mb4_unicode_ci AND profile.role_key = CONVERT(CONCAT(apps.app_id, '.', CASE WHEN BINARY legacy.role IN (BINARY 'owner', BINARY 'admin') THEN 'admin' WHEN BINARY legacy.role IN (BINARY 'operator', BINARY 'store_manager', BINARY 'menu_manager') AND BINARY apps.app_id = BINARY 'pos' THEN 'location_manager' WHEN BINARY legacy.role IN (BINARY 'operator', BINARY 'store_manager', BINARY 'menu_manager') AND BINARY apps.app_id = BINARY 'flow' THEN 'manager' WHEN BINARY legacy.role IN (BINARY 'operator', BINARY 'store_manager', BINARY 'menu_manager') AND BINARY apps.app_id IN (BINARY 'erp', BINARY 'hr', BINARY 'crm', BINARY 'finance', BINARY 'recruit') THEN 'manager' WHEN BINARY legacy.role IN (BINARY 'operator', BINARY 'store_manager', BINARY 'menu_manager') AND BINARY apps.app_id = BINARY 'menu' THEN 'manager' WHEN BINARY legacy.role = BINARY 'viewer' THEN 'viewer' WHEN BINARY apps.app_id = BINARY 'pos' THEN 'cashier' WHEN BINARY apps.app_id = BINARY 'flow' THEN 'contributor' WHEN BINARY apps.app_id = BINARY 'hr' THEN 'employee' ELSE 'viewer' END) USING utf8mb4) COLLATE utf8mb4_unicode_ci WHERE legacy.business_id > 0 AND BINARY apps.app_id IN (BINARY 'pos', BINARY 'erp', BINARY 'hr', BINARY 'crm', BINARY 'finance', BINARY 'flow', BINARY 'insights', BINARY 'menu', BINARY 'recruit') AND NOT EXISTS (SELECT 1 FROM authorization_role_assignments existing WHERE existing.business_id = legacy.business_id AND existing.user_id = legacy.user_id AND existing.group_id IS NULL AND existing.app_id = profile.app_id AND existing.role_id = profile.id)",
    'SELECT 1'
);
PREPARE backfill_roles_statement FROM @backfill_roles_sql;
EXECUTE backfill_roles_statement;
DEALLOCATE PREPARE backfill_roles_statement;

DELETE FROM authorization_role_permissions WHERE permission_key = 'flow:settings:manage';
DELETE FROM authorization_permission_catalog WHERE permission_key = 'flow:settings:manage';

DELETE scope
FROM authorization_role_assignment_scopes scope
JOIN authorization_role_assignments assignment ON assignment.id = scope.assignment_id
JOIN authorization_roles profile ON profile.id = assignment.role_id
WHERE profile.business_id = 0 AND profile.app_id = 'dash';

DELETE assignment
FROM authorization_role_assignments assignment
JOIN authorization_roles profile ON profile.id = assignment.role_id
WHERE profile.business_id = 0 AND profile.app_id = 'dash';

DELETE permissions
FROM authorization_role_permissions permissions
JOIN authorization_roles profile ON profile.id = permissions.role_id
WHERE profile.business_id = 0 AND profile.app_id = 'dash';

DELETE FROM authorization_roles WHERE business_id = 0 AND app_id = 'dash';
DROP TABLE IF EXISTS authorization_dash_app_defaults;
DROP TABLE IF EXISTS hub_roles;
DROP TABLE IF EXISTS user_app_permissions;
DROP TABLE IF EXISTS user_business_access;
DROP TABLE IF EXISTS user_group_resource_permissions;
DROP TABLE IF EXISTS user_group_app_permissions;

ALTER TABLE user_groups MODIFY COLUMN business_id bigint NOT NULL;
ALTER TABLE user_group_members MODIFY COLUMN business_id bigint NOT NULL;

SET @legacy_group_name_index_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'user_groups'
      AND index_name = 'uk_group_owner_name'
);
SET @drop_legacy_group_name_index_sql := IF(
    @legacy_group_name_index_exists,
    'ALTER TABLE user_groups DROP INDEX uk_group_owner_name',
    'SELECT 1'
);
PREPARE drop_legacy_group_name_index_statement FROM @drop_legacy_group_name_index_sql;
EXECUTE drop_legacy_group_name_index_statement;
DEALLOCATE PREPARE drop_legacy_group_name_index_statement;

SET @business_group_name_index_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'user_groups'
      AND index_name = 'uk_group_business_name'
);
SET @add_business_group_name_index_sql := IF(
    @business_group_name_index_exists,
    'SELECT 1',
    'ALTER TABLE user_groups ADD UNIQUE INDEX uk_group_business_name (business_id, name)'
);
PREPARE add_business_group_name_index_statement FROM @add_business_group_name_index_sql;
EXECUTE add_business_group_name_index_statement;
DEALLOCATE PREPARE add_business_group_name_index_statement;

SET @legacy_group_member_index_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'user_group_members'
      AND index_name = 'idx_ugm_user'
);
SET @drop_legacy_group_member_index_sql := IF(
    @legacy_group_member_index_exists,
    'ALTER TABLE user_group_members DROP INDEX idx_ugm_user',
    'SELECT 1'
);
PREPARE drop_legacy_group_member_index_statement FROM @drop_legacy_group_member_index_sql;
EXECUTE drop_legacy_group_member_index_statement;
DEALLOCATE PREPARE drop_legacy_group_member_index_statement;

SET @business_group_member_index_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'user_group_members'
      AND index_name = 'idx_ugm_business_user'
);
SET @add_business_group_member_index_sql := IF(
    @business_group_member_index_exists,
    'SELECT 1',
    'ALTER TABLE user_group_members ADD INDEX idx_ugm_business_user (business_id, user_id)'
);
PREPARE add_business_group_member_index_statement FROM @add_business_group_member_index_sql;
EXECUTE add_business_group_member_index_statement;
DEALLOCATE PREPARE add_business_group_member_index_statement;

SET @dash_role_id_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_business_members'
      AND column_name = 'dash_role_id'
);
SET @drop_dash_role_id_sql := IF(
    @dash_role_id_column_exists,
    'ALTER TABLE authorization_business_members DROP COLUMN dash_role_id',
    'SELECT 1'
);
PREPARE drop_dash_role_id_statement FROM @drop_dash_role_id_sql;
EXECUTE drop_dash_role_id_statement;
DEALLOCATE PREPARE drop_dash_role_id_statement;

SET @dash_role_key_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_business_members'
      AND column_name = 'dash_role_key'
);
SET @drop_dash_role_key_sql := IF(
    @dash_role_key_column_exists,
    'ALTER TABLE authorization_business_members DROP COLUMN dash_role_key',
    'SELECT 1'
);
PREPARE drop_dash_role_key_statement FROM @drop_dash_role_key_sql;
EXECUTE drop_dash_role_key_statement;
DEALLOCATE PREPARE drop_dash_role_key_statement;
