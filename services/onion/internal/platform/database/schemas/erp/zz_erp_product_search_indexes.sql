-- Adds indexed product lookup/search support for ERP/POS catalog screens.

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_products'
     AND INDEX_NAME = 'idx_erp_products_business_name') = 0,
  'ALTER TABLE erp_products ADD INDEX idx_erp_products_business_name (business_id, name)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_products'
     AND INDEX_NAME = 'idx_erp_products_business_category_name') = 0,
  'ALTER TABLE erp_products ADD INDEX idx_erp_products_business_category_name (business_id, category_id, name)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_products'
     AND INDEX_NAME = 'idx_erp_products_business_active') = 0,
  'ALTER TABLE erp_products ADD INDEX idx_erp_products_business_active (business_id, is_active, track_inventory)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_products'
     AND INDEX_NAME = 'idx_erp_products_search') = 0,
  'ALTER TABLE erp_products ADD FULLTEXT INDEX idx_erp_products_search (name, sku, barcode)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
