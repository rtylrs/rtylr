-- Business-leading order history index. `id` makes equal creation timestamps
-- stable for the paginated storefront response. Roll back only after checking
-- live plans: DROP INDEX idx_store_orders_business_created ON store_orders.
-- Existing owner-leading indexes remain unchanged.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'store_orders' AND INDEX_NAME = 'idx_store_orders_business_created') = 0,
  'ALTER TABLE store_orders ADD INDEX idx_store_orders_business_created (business_id, created_at, id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
