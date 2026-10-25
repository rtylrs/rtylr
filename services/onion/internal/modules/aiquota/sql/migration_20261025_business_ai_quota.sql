-- The AI allowance follows the business's plan, so its usage and allocation
-- mode are kept per business. Existing usage moves to the owner's primary
-- business; the allocation mode is copied to each of the owner's active
-- businesses. owner_id stays, nullable, until a later release drops it, and the
-- old owner-keyed unique index stays with it so the previous build keeps
-- counting correctly while this one rolls out. Fresh databases skip all of it.

DROP TEMPORARY TABLE IF EXISTS quota_owner_primary_business;
CREATE TEMPORARY TABLE quota_owner_primary_business AS
SELECT b.owner_id,
       COALESCE(
           MIN(CASE WHEN COALESCE(b.is_active, TRUE) = TRUE AND b.id = u.business_id THEN b.id END),
           MIN(CASE WHEN COALESCE(b.is_active, TRUE) = TRUE THEN b.id END),
           MIN(b.id)
       ) AS business_id
FROM businesses b
LEFT JOIN users u ON u.id = b.owner_id
GROUP BY b.owner_id;

-- ai_usage_monthly ------------------------------------------------------------

SET @legacy_usage := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'ai_usage_monthly' AND column_name = 'owner_id'
);

SET @missing := (
    SELECT COUNT(*) = 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'ai_usage_monthly' AND column_name = 'business_id'
);
SET @stmt := IF(@missing, 'ALTER TABLE ai_usage_monthly ADD COLUMN business_id bigint DEFAULT NULL AFTER id', 'SELECT 1');
PREPARE quota_stmt FROM @stmt;
EXECUTE quota_stmt;
DEALLOCATE PREPARE quota_stmt;

SET @stmt := IF(@legacy_usage, 'ALTER TABLE ai_usage_monthly MODIFY owner_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE quota_stmt FROM @stmt;
EXECUTE quota_stmt;
DEALLOCATE PREPARE quota_stmt;

SET @stmt := IF(@legacy_usage,
    'UPDATE ai_usage_monthly usage_row
     JOIN quota_owner_primary_business p ON p.owner_id = usage_row.owner_id
     SET usage_row.business_id = p.business_id
     WHERE usage_row.business_id IS NULL',
    'SELECT 1');
PREPARE quota_stmt FROM @stmt;
EXECUTE quota_stmt;
DEALLOCATE PREPARE quota_stmt;

SET @missing := (
    SELECT COUNT(*) = 0 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = 'ai_usage_monthly' AND index_name = 'uniq_ai_usage_business_user_month'
);
SET @stmt := IF(@missing,
    'ALTER TABLE ai_usage_monthly ADD UNIQUE KEY uniq_ai_usage_business_user_month (business_id, user_id, billing_month)',
    'SELECT 1');
PREPARE quota_stmt FROM @stmt;
EXECUTE quota_stmt;
DEALLOCATE PREPARE quota_stmt;

-- ai_quota_settings -----------------------------------------------------------

SET @legacy_settings := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'ai_quota_settings' AND column_name = 'owner_id'
);

SET @missing := (
    SELECT COUNT(*) = 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'ai_quota_settings' AND column_name = 'business_id'
);
SET @stmt := IF(@missing, 'ALTER TABLE ai_quota_settings ADD COLUMN business_id bigint DEFAULT NULL FIRST', 'SELECT 1');
PREPARE quota_stmt FROM @stmt;
EXECUTE quota_stmt;
DEALLOCATE PREPARE quota_stmt;

SET @stmt := IF(@legacy_settings,
    'UPDATE ai_quota_settings setting
     JOIN quota_owner_primary_business p ON p.owner_id = setting.owner_id
     SET setting.business_id = p.business_id
     WHERE setting.business_id IS NULL',
    'SELECT 1');
PREPARE quota_stmt FROM @stmt;
EXECUTE quota_stmt;
DEALLOCATE PREPARE quota_stmt;

SET @owner_primary_key := (
    SELECT COUNT(*) > 0 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = 'ai_quota_settings'
      AND index_name = 'PRIMARY' AND column_name = 'owner_id'
);
SET @stmt := IF(@owner_primary_key, 'ALTER TABLE ai_quota_settings DROP PRIMARY KEY', 'SELECT 1');
PREPARE quota_stmt FROM @stmt;
EXECUTE quota_stmt;
DEALLOCATE PREPARE quota_stmt;

SET @stmt := IF(@legacy_settings, 'ALTER TABLE ai_quota_settings MODIFY owner_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE quota_stmt FROM @stmt;
EXECUTE quota_stmt;
DEALLOCATE PREPARE quota_stmt;

SET @stmt := IF(@legacy_settings,
    'INSERT INTO ai_quota_settings (business_id, owner_id, mode)
     SELECT b.id, setting.owner_id, setting.mode
     FROM ai_quota_settings setting
     JOIN quota_owner_primary_business p ON p.owner_id = setting.owner_id AND p.business_id = setting.business_id
     JOIN businesses b ON b.owner_id = setting.owner_id AND COALESCE(b.is_active, TRUE) = TRUE
     WHERE b.id <> setting.business_id
       AND NOT EXISTS (SELECT 1 FROM ai_quota_settings taken WHERE taken.business_id = b.id)',
    'SELECT 1');
PREPARE quota_stmt FROM @stmt;
EXECUTE quota_stmt;
DEALLOCATE PREPARE quota_stmt;

-- A mode whose owner has no business cannot be reached any more.
DELETE FROM ai_quota_settings WHERE business_id IS NULL;

SET @nullable := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'ai_quota_settings'
      AND column_name = 'business_id' AND is_nullable = 'YES'
);
SET @stmt := IF(@nullable, 'ALTER TABLE ai_quota_settings MODIFY business_id bigint NOT NULL', 'SELECT 1');
PREPARE quota_stmt FROM @stmt;
EXECUTE quota_stmt;
DEALLOCATE PREPARE quota_stmt;

SET @missing := (
    SELECT COUNT(*) = 0 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = 'ai_quota_settings' AND index_name = 'PRIMARY'
);
SET @stmt := IF(@missing, 'ALTER TABLE ai_quota_settings ADD PRIMARY KEY (business_id)', 'SELECT 1');
PREPARE quota_stmt FROM @stmt;
EXECUTE quota_stmt;
DEALLOCATE PREPARE quota_stmt;

DROP TEMPORARY TABLE IF EXISTS quota_owner_primary_business;
