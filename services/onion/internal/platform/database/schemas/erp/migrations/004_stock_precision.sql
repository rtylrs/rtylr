-- Widen the legacy scalar to match stock_levels.on_hand DECIMAL(14,3) so the
-- incremental scalar sync can't overflow. Widening only; safe + lossless.
ALTER TABLE erp_products MODIFY stock_quantity DECIMAL(14,3) NOT NULL DEFAULT 0;
ALTER TABLE erp_product_variants MODIFY stock_quantity DECIMAL(14,3) NOT NULL DEFAULT 0;
