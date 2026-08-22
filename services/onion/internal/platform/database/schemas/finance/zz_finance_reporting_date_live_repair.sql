-- Repair the reporting date used by invoice analytics on existing databases.
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_invoices' AND COLUMN_NAME = 'reporting_date') = 0, 'ALTER TABLE finance_invoices ADD COLUMN reporting_date DATE GENERATED ALWAYS AS (COALESCE(paid_date, issue_date)) STORED', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_invoices' AND INDEX_NAME = 'idx_finance_invoices_business_status_reporting_date') = 0, 'ALTER TABLE finance_invoices ADD INDEX idx_finance_invoices_business_status_reporting_date (business_id, status, reporting_date)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
