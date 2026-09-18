-- Role and Access Control V2 integrity hardening.
-- This migration fails closed on ambiguous existing rows before adding
-- constraints. It is intentionally separate from the backfill so operators
-- can review the reported counts before applying it.

SET @duplicate_direct_roles := (
    SELECT COUNT(*)
    FROM (
        SELECT business_id, user_id, app_id, role_id, COUNT(*) AS row_count
        FROM authorization_role_assignments
        WHERE status = 'active' AND user_id IS NOT NULL
        GROUP BY business_id, user_id, app_id, role_id
        HAVING COUNT(*) > 1
    ) duplicate_rows
);

SET @duplicate_direct_licenses := (
    SELECT COUNT(*)
    FROM (
        SELECT business_id, user_id, app_id, COUNT(*) AS row_count
        FROM authorization_license_assignments
        WHERE status = 'active' AND user_id IS NOT NULL
        GROUP BY business_id, user_id, app_id
        HAVING COUNT(*) > 1
    ) duplicate_rows
);

SET @invalid_assignment_targets := (
    SELECT COUNT(*)
    FROM authorization_role_assignments
    WHERE (user_id IS NULL AND group_id IS NULL)
       OR (user_id IS NOT NULL AND group_id IS NOT NULL)
);

SET @invalid_license_targets := (
    SELECT COUNT(*)
    FROM authorization_license_assignments
    WHERE (user_id IS NULL AND group_id IS NULL)
       OR (user_id IS NOT NULL AND group_id IS NOT NULL)
);

SET @orphan_member_businesses := (
    SELECT COUNT(*)
    FROM authorization_business_members m
    LEFT JOIN businesses b ON b.id = m.business_id
    WHERE b.id IS NULL
);

SET @orphan_member_users := (
    SELECT COUNT(*)
    FROM authorization_business_members m
    LEFT JOIN users u ON u.id = m.user_id
    WHERE u.id IS NULL
);

SET @orphan_role_assignment_businesses := (
    SELECT COUNT(*)
    FROM authorization_role_assignments a
    LEFT JOIN businesses b ON b.id = a.business_id
    WHERE b.id IS NULL
);

SET @orphan_role_assignment_users := (
    SELECT COUNT(*)
    FROM authorization_role_assignments a
    LEFT JOIN users u ON u.id = a.user_id
    WHERE a.user_id IS NOT NULL AND u.id IS NULL
);

SET @orphan_role_assignment_groups := (
    SELECT COUNT(*)
    FROM authorization_role_assignments a
    LEFT JOIN user_groups g ON g.id = a.group_id
    WHERE a.group_id IS NOT NULL AND g.id IS NULL
);

SET @orphan_license_assignment_businesses := (
    SELECT COUNT(*)
    FROM authorization_license_assignments a
    LEFT JOIN businesses b ON b.id = a.business_id
    WHERE b.id IS NULL
);

SET @orphan_license_assignment_users := (
    SELECT COUNT(*)
    FROM authorization_license_assignments a
    LEFT JOIN users u ON u.id = a.user_id
    WHERE a.user_id IS NOT NULL AND u.id IS NULL
);

SET @orphan_license_assignment_groups := (
    SELECT COUNT(*)
    FROM authorization_license_assignments a
    LEFT JOIN user_groups g ON g.id = a.group_id
    WHERE a.group_id IS NOT NULL AND g.id IS NULL
);

SET @invalid_scope_rows := (
    SELECT COUNT(*)
    FROM authorization_role_assignment_scopes
    WHERE assignment_id IS NULL OR scope_type = '' OR scope_id = ''
);

SET @invalid_scope_ownership := (
    SELECT COUNT(*)
    FROM authorization_role_assignment_scopes ras
    JOIN authorization_role_assignments a ON a.id = ras.assignment_id
    WHERE ras.scope_type NOT IN ('pos_location', 'flow_project')
       OR (ras.scope_type = 'pos_location' AND NOT EXISTS (
           SELECT 1 FROM pos_locations l
           WHERE l.id = CAST(ras.scope_id AS UNSIGNED)
             AND l.business_id = a.business_id
       ))
       OR (ras.scope_type = 'flow_project' AND NOT EXISTS (
           SELECT 1 FROM flow_projects p
           WHERE p.id = CAST(ras.scope_id AS UNSIGNED)
             AND p.business_id = a.business_id
       ))
);

SET @missing_role_permission_roles := (
    SELECT COUNT(*)
    FROM authorization_role_permissions rp
    LEFT JOIN authorization_roles r ON r.id = rp.role_id
    WHERE r.id IS NULL
);

SET @missing_role_permission_keys := (
    SELECT COUNT(*)
    FROM authorization_role_permissions rp
    LEFT JOIN authorization_permission_catalog p ON p.permission_key = rp.permission_key
    WHERE p.permission_key IS NULL
);

SET @missing_assignment_roles := (
    SELECT COUNT(*)
    FROM authorization_role_assignments ra
    LEFT JOIN authorization_roles r ON r.id = ra.role_id
    WHERE r.id IS NULL
);

SET @missing_scope_assignments := (
    SELECT COUNT(*)
    FROM authorization_role_assignment_scopes ras
    LEFT JOIN authorization_role_assignments ra ON ra.id = ras.assignment_id
    WHERE ra.id IS NULL
);

-- sqlc parses migration files as semicolon-delimited statements, so the
-- preflight abort uses a conditional invalid statement instead of a stored
-- procedure. MySQL prepares the invalid branch only when a violation exists.
SET @integrity_violation_count := @duplicate_direct_roles
    + @duplicate_direct_licenses
    + @invalid_assignment_targets
    + @invalid_license_targets
    + @orphan_member_businesses
    + @orphan_member_users
    + @orphan_role_assignment_businesses
    + @orphan_role_assignment_users
    + @orphan_role_assignment_groups
    + @orphan_license_assignment_businesses
    + @orphan_license_assignment_users
    + @orphan_license_assignment_groups
    + @invalid_scope_rows
    + @invalid_scope_ownership
    + @missing_role_permission_roles
    + @missing_role_permission_keys
    + @missing_assignment_roles
    + @missing_scope_assignments;
SET @sql := IF(
    @integrity_violation_count > 0,
    'SELECT role_access_v2_integrity_abort FROM authorization_roles',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
     WHERE TABLE_SCHEMA = DATABASE()
       AND TABLE_NAME = 'authorization_role_assignments'
       AND INDEX_NAME = 'uk_auth_active_direct_role'),
    'SELECT 1',
    'ALTER TABLE authorization_role_assignments ADD UNIQUE KEY uk_auth_active_direct_role (business_id, user_id, app_id, role_id, status)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND CONSTRAINT_NAME = 'fk_auth_member_business'),
    'SELECT 1',
    'ALTER TABLE authorization_business_members ADD CONSTRAINT fk_auth_member_business FOREIGN KEY (business_id) REFERENCES businesses (id)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND CONSTRAINT_NAME = 'fk_auth_member_user'),
    'SELECT 1',
    'ALTER TABLE authorization_business_members ADD CONSTRAINT fk_auth_member_user FOREIGN KEY (user_id) REFERENCES users (id)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND CONSTRAINT_NAME = 'fk_auth_license_business'),
    'SELECT 1',
    'ALTER TABLE authorization_license_assignments ADD CONSTRAINT fk_auth_license_business FOREIGN KEY (business_id) REFERENCES businesses (id)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND CONSTRAINT_NAME = 'fk_auth_license_user'),
    'SELECT 1',
    'ALTER TABLE authorization_license_assignments ADD CONSTRAINT fk_auth_license_user FOREIGN KEY (user_id) REFERENCES users (id)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND CONSTRAINT_NAME = 'fk_auth_license_group'),
    'SELECT 1',
    'ALTER TABLE authorization_license_assignments ADD CONSTRAINT fk_auth_license_group FOREIGN KEY (group_id) REFERENCES user_groups (id)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND CONSTRAINT_NAME = 'fk_auth_role_assignment_business'),
    'SELECT 1',
    'ALTER TABLE authorization_role_assignments ADD CONSTRAINT fk_auth_role_assignment_business FOREIGN KEY (business_id) REFERENCES businesses (id)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND CONSTRAINT_NAME = 'fk_auth_role_assignment_user'),
    'SELECT 1',
    'ALTER TABLE authorization_role_assignments ADD CONSTRAINT fk_auth_role_assignment_user FOREIGN KEY (user_id) REFERENCES users (id)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND CONSTRAINT_NAME = 'fk_auth_role_assignment_group'),
    'SELECT 1',
    'ALTER TABLE authorization_role_assignments ADD CONSTRAINT fk_auth_role_assignment_group FOREIGN KEY (group_id) REFERENCES user_groups (id)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND TABLE_NAME = 'authorization_role_assignment_scopes'
       AND CONSTRAINT_NAME = 'chk_auth_assignment_scope_type'),
    'SELECT 1',
    'ALTER TABLE authorization_role_assignment_scopes ADD CONSTRAINT chk_auth_assignment_scope_type CHECK (scope_type IN (''pos_location'', ''flow_project'') AND scope_id REGEXP ''^[1-9][0-9]*$'')'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND CONSTRAINT_NAME = 'fk_auth_role_permissions_role'),
    'SELECT 1',
    'ALTER TABLE authorization_role_permissions ADD CONSTRAINT fk_auth_role_permissions_role FOREIGN KEY (role_id) REFERENCES authorization_roles (id)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND CONSTRAINT_NAME = 'fk_auth_role_permissions_permission'),
    'SELECT 1',
    'ALTER TABLE authorization_role_permissions ADD CONSTRAINT fk_auth_role_permissions_permission FOREIGN KEY (permission_key) REFERENCES authorization_permission_catalog (permission_key)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND CONSTRAINT_NAME = 'fk_auth_role_assignments_role'),
    'SELECT 1',
    'ALTER TABLE authorization_role_assignments ADD CONSTRAINT fk_auth_role_assignments_role FOREIGN KEY (role_id) REFERENCES authorization_roles (id)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND CONSTRAINT_NAME = 'fk_auth_assignment_scopes_assignment'),
    'SELECT 1',
    'ALTER TABLE authorization_role_assignment_scopes ADD CONSTRAINT fk_auth_assignment_scopes_assignment FOREIGN KEY (assignment_id) REFERENCES authorization_role_assignments (id)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
     WHERE TABLE_SCHEMA = DATABASE()
       AND TABLE_NAME = 'authorization_license_assignments'
       AND INDEX_NAME = 'uk_auth_active_direct_license'),
    'SELECT 1',
    'ALTER TABLE authorization_license_assignments ADD UNIQUE KEY uk_auth_active_direct_license (business_id, user_id, app_id, status)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND TABLE_NAME = 'authorization_role_assignments'
       AND CONSTRAINT_NAME = 'chk_auth_role_assignment_one_target'),
    'SELECT 1',
    'ALTER TABLE authorization_role_assignments ADD CONSTRAINT chk_auth_role_assignment_one_target CHECK ((user_id IS NOT NULL AND group_id IS NULL) OR (user_id IS NULL AND group_id IS NOT NULL))'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
       AND TABLE_NAME = 'authorization_license_assignments'
       AND CONSTRAINT_NAME = 'chk_auth_license_assignment_one_target'),
    'SELECT 1',
    'ALTER TABLE authorization_license_assignments ADD CONSTRAINT chk_auth_license_assignment_one_target CHECK ((user_id IS NOT NULL AND group_id IS NULL) OR (user_id IS NULL AND group_id IS NOT NULL))'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
