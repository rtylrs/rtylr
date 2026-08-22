-- Bills carry the supplier name they were entered under (crm.sales-style
-- snapshot): finance.bill.v1 now requires supplier_name and treats
-- supplier_id as optional, so a one-off bill records who it is owed to even
-- when no ERP supplier exists. Empty for legacy rows; new writes always set it.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_bills' AND COLUMN_NAME = 'supplier_name') = 0,
  'ALTER TABLE finance_bills ADD COLUMN supplier_name VARCHAR(255) NOT NULL DEFAULT '''' AFTER supplier_id',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
