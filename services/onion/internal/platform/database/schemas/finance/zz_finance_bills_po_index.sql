-- finance_bills.purchase_order_id already exists via
-- zz_finance_restore_deleted_branch_columns.sql; this only adds the index a
-- prior migration's column-existence guard never let fire.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_bills' AND INDEX_NAME = 'idx_finance_bills_business_po') = 0,
  'ALTER TABLE finance_bills ADD INDEX idx_finance_bills_business_po (business_id, purchase_order_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
