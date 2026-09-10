-- Hot-path tenant indexes for stock and variant reads. Keep these in the
-- traversed ERP schema directory; schemas/_perf is not applied by migrate.sh.
--
-- Each ALTER is guarded on both the table and the index: stock_ledger is
-- created by erp/migrations/002_unified_stock.sql, which this runner does not
-- traverse, so a fresh database legitimately has no such table yet and an
-- unguarded ALTER would abort the whole migration run.
SET @sql := IF(
  (SELECT COUNT(*)
   FROM INFORMATION_SCHEMA.TABLES
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'stock_ledger') = 0,
  'SELECT 1',
  IF(
    (SELECT COUNT(*)
     FROM INFORMATION_SCHEMA.STATISTICS
     WHERE TABLE_SCHEMA = DATABASE()
       AND TABLE_NAME = 'stock_ledger'
       AND INDEX_NAME = 'idx_stock_ledger_business_product_created') = 0,
  'ALTER TABLE stock_ledger ADD INDEX idx_stock_ledger_business_product_created (business_id, product_id, created_at)',
  'SELECT 1')
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*)
   FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE()
     AND TABLE_NAME = 'erp_product_variants'
     AND INDEX_NAME = 'idx_erp_variants_business_product') = 0,
  'ALTER TABLE erp_product_variants ADD INDEX idx_erp_variants_business_product (business_id, product_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
