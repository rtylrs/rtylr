-- Backfill the normalized model from the legacy permission row without
-- replacing any normalized writes that may have happened during rollout.

-- A legacy row without business_id is safe to backfill only when its owner
-- has exactly one business. The preflight reports the count and stops the
-- migration instead of guessing which tenant owns an ambiguous row.
SELECT COUNT(*) AS ambiguous_legacy_business_mappings
FROM user_app_permissions uap
WHERE (uap.business_id IS NULL OR uap.business_id = 0)
  AND (
      SELECT COUNT(*)
      FROM businesses b
      WHERE b.owner_id = uap.owner_id
  ) <> 1;

SET @ambiguous_legacy_business_mappings := (
    SELECT COUNT(*)
    FROM user_app_permissions uap
    WHERE (uap.business_id IS NULL OR uap.business_id = 0)
      AND (
          SELECT COUNT(*)
          FROM businesses b
          WHERE b.owner_id = uap.owner_id
      ) <> 1
);
SET @sql := IF(
    @ambiguous_legacy_business_mappings > 0,
    'SELECT role_access_v2_backfill_abort FROM user_app_permissions',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- A nonzero business_id is safe only when it belongs to the legacy row owner.
-- Missing businesses are also rejected here so the backfill cannot create
-- normalized rows for an invalid tenant before the integrity migration runs.
SELECT COUNT(*) AS invalid_legacy_business_mappings
FROM user_app_permissions uap
LEFT JOIN businesses b ON b.id = uap.business_id
WHERE uap.business_id IS NOT NULL
  AND uap.business_id <> 0
  AND (b.id IS NULL OR b.owner_id <> uap.owner_id);

SET @invalid_legacy_business_mappings := (
    SELECT COUNT(*)
    FROM user_app_permissions uap
    LEFT JOIN businesses b ON b.id = uap.business_id
    WHERE uap.business_id IS NOT NULL
      AND uap.business_id <> 0
      AND (b.id IS NULL OR b.owner_id <> uap.owner_id)
);
SET @sql := IF(
    @ambiguous_legacy_business_mappings > 0
        OR @invalid_legacy_business_mappings > 0,
    'SELECT role_access_v2_backfill_abort FROM user_app_permissions',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

START TRANSACTION;

INSERT INTO authorization_business_members
    (business_id, user_id, dash_role_id, dash_role_key, status)
SELECT legacy.business_id,
       legacy.user_id,
       role.id,
       legacy.dash_role_key,
       CASE WHEN legacy.is_active THEN 'active' ELSE 'revoked' END
FROM (
    SELECT COALESCE(NULLIF(uap.business_id, 0), (
               SELECT MIN(b.id)
               FROM businesses b
               WHERE b.owner_id = uap.owner_id
               GROUP BY b.owner_id
               HAVING COUNT(*) = 1
           )) AS business_id,
           uap.user_id,
           uap.is_active,
           CASE uap.role
               WHEN 'owner' THEN 'owner'
               WHEN 'admin' THEN 'admin'
               WHEN 'operator' THEN 'manager'
               WHEN 'billing' THEN 'manager'
               WHEN 'menu_manager' THEN 'manager'
               WHEN 'store_manager' THEN 'manager'
               WHEN 'viewer' THEN 'viewer'
               ELSE 'member'
           END AS dash_role_key
    FROM user_app_permissions uap
) legacy
JOIN authorization_roles role
  ON role.business_id = 0
 AND role.app_id = 'dash'
 AND role.role_key = CONCAT('dash.', legacy.dash_role_key)
WHERE legacy.business_id IS NOT NULL
  AND legacy.business_id > 0
ON DUPLICATE KEY UPDATE user_id = VALUES(user_id);
SELECT ROW_COUNT() AS authorization_business_members_affected;

INSERT INTO authorization_license_assignments
    (business_id, user_id, app_id, source, status)
SELECT DISTINCT legacy.business_id,
       legacy.user_id,
       apps.app_id,
       'explicit',
       CASE WHEN legacy.is_active THEN 'active' ELSE 'revoked' END
FROM (
    SELECT COALESCE(NULLIF(uap.business_id, 0), (
               SELECT MIN(b.id)
               FROM businesses b
               WHERE b.owner_id = uap.owner_id
               GROUP BY b.owner_id
               HAVING COUNT(*) = 1
           )) AS business_id,
           uap.user_id,
           uap.apps,
           uap.is_active
    FROM user_app_permissions uap
) legacy
JOIN (
    SELECT 'pos' AS app_id
    UNION ALL SELECT 'erp'
    UNION ALL SELECT 'hr'
    UNION ALL SELECT 'crm'
    UNION ALL SELECT 'finance'
    UNION ALL SELECT 'flow'
    UNION ALL SELECT 'insights'
    UNION ALL SELECT 'menu'
    UNION ALL SELECT 'recruit'
) apps
  ON JSON_VALID(legacy.apps)
 AND JSON_CONTAINS(legacy.apps, JSON_QUOTE(apps.app_id), '$')
WHERE legacy.business_id IS NOT NULL
  AND legacy.business_id > 0
  AND apps.app_id IN ('pos','erp','hr','crm','finance','flow','insights','menu','recruit')
  AND NOT EXISTS (
      SELECT 1
      FROM authorization_license_assignments existing
      WHERE existing.business_id = legacy.business_id
        AND existing.user_id = legacy.user_id
        AND existing.group_id IS NULL
        AND existing.app_id = apps.app_id
  );
SELECT ROW_COUNT() AS authorization_license_assignments_inserted;

INSERT INTO authorization_role_assignments
    (business_id, user_id, app_id, role_id, source, status, is_primary)
SELECT DISTINCT legacy.business_id,
       legacy.user_id,
       apps.app_id,
       role.id,
       'explicit',
       CASE WHEN legacy.is_active THEN 'active' ELSE 'revoked' END,
       TRUE
FROM (
    SELECT COALESCE(NULLIF(uap.business_id, 0), (
               SELECT MIN(b.id)
               FROM businesses b
               WHERE b.owner_id = uap.owner_id
               GROUP BY b.owner_id
               HAVING COUNT(*) = 1
           )) AS business_id,
           uap.user_id,
           uap.apps,
           uap.role,
           uap.is_active
    FROM user_app_permissions uap
) legacy
JOIN (
    SELECT 'pos' AS app_id
    UNION ALL SELECT 'erp'
    UNION ALL SELECT 'hr'
    UNION ALL SELECT 'crm'
    UNION ALL SELECT 'finance'
    UNION ALL SELECT 'flow'
    UNION ALL SELECT 'insights'
    UNION ALL SELECT 'menu'
    UNION ALL SELECT 'recruit'
) apps
  ON JSON_VALID(legacy.apps)
 AND JSON_CONTAINS(legacy.apps, JSON_QUOTE(apps.app_id), '$')
JOIN authorization_roles role
  ON role.business_id = 0
 AND role.app_id = apps.app_id
 AND role.role_key = CONCAT(
     apps.app_id,
     '.',
     CASE
         WHEN legacy.role IN ('owner','admin') THEN 'admin'
         WHEN legacy.role IN ('operator','store_manager','menu_manager')
              AND apps.app_id = 'pos' THEN 'location_manager'
         WHEN legacy.role IN ('operator','store_manager','menu_manager')
              AND apps.app_id = 'flow' THEN 'manager'
         WHEN legacy.role IN ('operator','store_manager','menu_manager')
              AND apps.app_id IN ('erp','hr','crm','finance','recruit') THEN 'manager'
         WHEN legacy.role IN ('operator','store_manager','menu_manager')
              AND apps.app_id = 'menu' THEN 'manager'
         -- The legacy read-only role stays read-only in every app; without
         -- this arm it falls through to the per-app working roles below and
         -- is escalated (pos.cashier / flow.contributor / hr.employee).
         WHEN legacy.role = 'viewer' THEN 'viewer'
         WHEN apps.app_id = 'pos' THEN 'cashier'
         WHEN apps.app_id = 'flow' THEN 'contributor'
         WHEN apps.app_id = 'hr' THEN 'employee'
         ELSE 'viewer'
     END
 )
WHERE legacy.business_id IS NOT NULL
  AND legacy.business_id > 0
  AND apps.app_id IN ('pos','erp','hr','crm','finance','flow','insights','menu','recruit')
  AND NOT EXISTS (
      SELECT 1
      FROM authorization_role_assignments existing
      WHERE existing.business_id = legacy.business_id
        AND existing.user_id = legacy.user_id
        AND existing.group_id IS NULL
        AND existing.app_id = apps.app_id
      AND existing.role_id = role.id
  );
SELECT ROW_COUNT() AS authorization_role_assignments_inserted;

COMMIT;
