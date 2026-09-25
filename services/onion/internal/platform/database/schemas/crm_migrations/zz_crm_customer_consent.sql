-- Preserve explicit CRM communication consent for existing installations.
-- MySQL 8.4 does not support ALTER TABLE ... ADD COLUMN IF NOT EXISTS;
-- inspect INFORMATION_SCHEMA so this migration remains idempotent.
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'crm_customers' AND COLUMN_NAME = 'consent') = 0, 'ALTER TABLE crm_customers ADD COLUMN consent JSON NULL', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
