-- finance_invoices.customer_id already exists via
-- zz_finance_restore_deleted_branch_columns.sql; this only adds the index a
-- prior migration's column-existence guard never let fire.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_invoices' AND INDEX_NAME = 'idx_finance_invoices_business_customer') = 0,
  'ALTER TABLE finance_invoices ADD INDEX idx_finance_invoices_business_customer (business_id, customer_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
