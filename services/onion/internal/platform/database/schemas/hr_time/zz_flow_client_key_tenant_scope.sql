-- Scope Flow work-day idempotency keys to the tenant as well as the owner.
-- The owner value is shared by businesses that belong to one account, so an
-- owner-only key can reject an unrelated business's offline retry.
SET @sql := IF((SELECT COUNT(*)
                FROM INFORMATION_SCHEMA.STATISTICS
                WHERE TABLE_SCHEMA = DATABASE()
                  AND TABLE_NAME = 'hr_time_entries'
                  AND INDEX_NAME = 'uk_time_entries_client_key') > 0,
               'ALTER TABLE hr_time_entries DROP INDEX uk_time_entries_client_key',
               'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF((SELECT COUNT(*)
                FROM INFORMATION_SCHEMA.STATISTICS
                WHERE TABLE_SCHEMA = DATABASE()
                  AND TABLE_NAME = 'hr_time_entries'
                  AND INDEX_NAME = 'uk_time_entries_client_key_tenant') = 0,
               'ALTER TABLE hr_time_entries ADD UNIQUE KEY uk_time_entries_client_key_tenant (owner_id, business_id, client_key)',
               'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
