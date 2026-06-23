-- Business-scoped lookup indexes for pre-launch ERP/POS catalog load checks.
-- MySQL has no portable ADD INDEX IF NOT EXISTS, so each ALTER is guarded
-- through INFORMATION_SCHEMA and prepared SQL.

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_categories' AND INDEX_NAME = 'idx_erp_categories_business_name') = 0,
  'ALTER TABLE erp_categories ADD INDEX idx_erp_categories_business_name (business_id, name)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_suppliers' AND INDEX_NAME = 'idx_erp_suppliers_business_active_name') = 0,
  'ALTER TABLE erp_suppliers ADD INDEX idx_erp_suppliers_business_active_name (business_id, is_active, name)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_products' AND INDEX_NAME = 'idx_erp_products_business_sku') = 0,
  'ALTER TABLE erp_products ADD INDEX idx_erp_products_business_sku (business_id, sku)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_products' AND INDEX_NAME = 'idx_erp_products_business_barcode') = 0,
  'ALTER TABLE erp_products ADD INDEX idx_erp_products_business_barcode (business_id, barcode)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_products' AND INDEX_NAME = 'idx_erp_products_business_inventory_name') = 0,
  'ALTER TABLE erp_products ADD INDEX idx_erp_products_business_inventory_name (business_id, is_active, track_inventory, name)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_products' AND INDEX_NAME = 'idx_erp_products_business_inventory_stock') = 0,
  'ALTER TABLE erp_products ADD INDEX idx_erp_products_business_inventory_stock (business_id, is_active, track_inventory, stock_quantity)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_inventory_transactions' AND INDEX_NAME = 'idx_erp_inventory_tx_business_product_created') = 0,
  'ALTER TABLE erp_inventory_transactions ADD INDEX idx_erp_inventory_tx_business_product_created (business_id, product_id, created_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_inventory_transactions' AND INDEX_NAME = 'idx_erp_inventory_tx_business_reference') = 0,
  'ALTER TABLE erp_inventory_transactions ADD INDEX idx_erp_inventory_tx_business_reference (business_id, reference_type, reference_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_purchase_orders' AND INDEX_NAME = 'idx_erp_purchase_orders_business_created') = 0,
  'ALTER TABLE erp_purchase_orders ADD INDEX idx_erp_purchase_orders_business_created (business_id, created_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_purchase_orders' AND INDEX_NAME = 'idx_erp_purchase_orders_business_status_created') = 0,
  'ALTER TABLE erp_purchase_orders ADD INDEX idx_erp_purchase_orders_business_status_created (business_id, status, created_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_purchase_order_items' AND INDEX_NAME = 'idx_erp_po_items_order') = 0,
  'ALTER TABLE erp_purchase_order_items ADD INDEX idx_erp_po_items_order (purchase_order_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_purchase_order_items' AND INDEX_NAME = 'idx_erp_po_items_product') = 0,
  'ALTER TABLE erp_purchase_order_items ADD INDEX idx_erp_po_items_product (product_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_product_images' AND INDEX_NAME = 'idx_product_images_business_product_primary') = 0,
  'ALTER TABLE erp_product_images ADD INDEX idx_product_images_business_product_primary (business_id, product_id, is_primary, position, id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_product_attributes' AND INDEX_NAME = 'idx_product_attributes_business_product_position') = 0,
  'ALTER TABLE erp_product_attributes ADD INDEX idx_product_attributes_business_product_position (business_id, product_id, position, id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
