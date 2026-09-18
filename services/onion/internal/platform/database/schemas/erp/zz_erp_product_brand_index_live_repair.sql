-- Repair the tenant/brand lookup index forced by the brands queries.
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_products' AND INDEX_NAME = 'idx_erp_products_business_brand') = 0, 'ALTER TABLE erp_products ADD INDEX idx_erp_products_business_brand (business_id, brand_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
