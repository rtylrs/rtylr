-- Scope Flow work-day idempotency keys to the tenant as well as the owner.
-- The owner value is shared by businesses that belong to one account, so an
-- owner-only key can reject an unrelated business's offline retry.
--
-- Lives in flow/ (not hr_time/) and must run after v8_time_tracking.sql: on a
-- fresh database only that file adds hr_time_entries.client_key, and directory
-- order applies flow/ after hr_time/. Both branches are guarded, so databases
-- where this already ran under its old hr_time/ path are unaffected.
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
