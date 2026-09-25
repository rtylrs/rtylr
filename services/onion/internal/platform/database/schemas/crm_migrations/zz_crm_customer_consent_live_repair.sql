-- Repair consent for existing CRM installations where the original migration
-- was skipped or was incorrectly marked applied by the migration runner.
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'crm_customers' AND COLUMN_NAME = 'consent') = 0, 'ALTER TABLE crm_customers ADD COLUMN consent JSON NULL', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
