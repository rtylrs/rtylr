-- support_customers.customer_id already exists in the base table declaration;
-- this only adds the index that a prior migration's guard never fired for
-- (it was keyed on the column's existence, so once the column landed by any
-- other means the index-adding branch could never run).
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'support_customers' AND INDEX_NAME = 'idx_support_customers_customer') = 0,
  'ALTER TABLE support_customers ADD INDEX idx_support_customers_customer (business_id, customer_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
