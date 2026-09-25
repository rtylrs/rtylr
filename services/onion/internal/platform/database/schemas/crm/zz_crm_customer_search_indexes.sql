-- Adds indexed CRM customer lookup/search support.

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'crm_customers'
     AND INDEX_NAME = 'idx_crm_customers_business_created') = 0,
  'ALTER TABLE crm_customers ADD INDEX idx_crm_customers_business_created (business_id, created_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'crm_customers'
     AND INDEX_NAME = 'idx_crm_customers_business_tier') = 0,
  'ALTER TABLE crm_customers ADD INDEX idx_crm_customers_business_tier (business_id, loyalty_tier)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'crm_customers'
     AND INDEX_NAME = 'idx_crm_customers_search') = 0,
  'ALTER TABLE crm_customers ADD FULLTEXT INDEX idx_crm_customers_search (first_name, last_name, email, phone)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
