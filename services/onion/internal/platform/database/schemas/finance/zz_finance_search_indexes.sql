-- Adds indexed Finance list/search support for invoices and expenses.

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_invoices'
     AND INDEX_NAME = 'idx_finance_invoices_business_created') = 0,
  'ALTER TABLE finance_invoices ADD INDEX idx_finance_invoices_business_created (business_id, created_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_invoices'
     AND INDEX_NAME = 'idx_finance_invoices_business_status_created') = 0,
  'ALTER TABLE finance_invoices ADD INDEX idx_finance_invoices_business_status_created (business_id, status, created_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_invoices'
     AND INDEX_NAME = 'idx_finance_invoices_search') = 0,
  'ALTER TABLE finance_invoices ADD FULLTEXT INDEX idx_finance_invoices_search (invoice_number, client_name, client_email)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_expenses'
     AND INDEX_NAME = 'idx_finance_expenses_business_date') = 0,
  'ALTER TABLE finance_expenses ADD INDEX idx_finance_expenses_business_date (business_id, expense_date)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_expenses'
     AND INDEX_NAME = 'idx_finance_expenses_business_category_date') = 0,
  'ALTER TABLE finance_expenses ADD INDEX idx_finance_expenses_business_category_date (business_id, category_id, expense_date)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_expenses'
     AND INDEX_NAME = 'idx_finance_expenses_search') = 0,
  'ALTER TABLE finance_expenses ADD FULLTEXT INDEX idx_finance_expenses_search (description, vendor)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
