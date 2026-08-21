-- Hot-path tenant indexes for stock and variant reads. Keep these in the
-- traversed ERP schema directory; schemas/_perf is not applied by migrate.sh.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'stock_ledger' AND INDEX_NAME = 'idx_stock_ledger_business_product_created') = 0,
  'ALTER TABLE stock_ledger ADD INDEX idx_stock_ledger_business_product_created (business_id, product_id, created_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_product_variants' AND INDEX_NAME = 'idx_erp_variants_business_product') = 0,
  'ALTER TABLE erp_product_variants ADD INDEX idx_erp_variants_business_product (business_id, product_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
