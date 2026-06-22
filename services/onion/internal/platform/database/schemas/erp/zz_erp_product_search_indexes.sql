-- Adds indexed product lookup/search support for ERP/POS catalog screens.
-- The migration runner uses --force, so duplicate-index errors are safe when
-- this has already been applied.

ALTER TABLE erp_products ADD INDEX idx_erp_products_business_name (business_id, name);
ALTER TABLE erp_products ADD INDEX idx_erp_products_business_category_name (business_id, category_id, name);
ALTER TABLE erp_products ADD INDEX idx_erp_products_business_active (business_id, is_active, track_inventory);
ALTER TABLE erp_products ADD FULLTEXT INDEX idx_erp_products_search (name, sku, barcode);
