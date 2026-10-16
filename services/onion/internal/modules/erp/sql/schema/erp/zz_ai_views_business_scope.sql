-- Business-scoped projection of this module's tables for the AI analyst (aiquery).
-- Published here so the owner controls the columns aiquery may read.
-- owner_id is shared by every business of an account, so scope by business_id.
-- Kept as its own zz_ file: migrate.sh records applied files by name.

CREATE OR REPLACE VIEW ai_products AS
SELECT id, owner_id, business_id, name, sku, category_id, brand_id, supplier_id,
       cost_price, sale_price, compare_at_price, currency,
       stock_quantity, low_stock_threshold, reorder_quantity,
       is_active, track_inventory, lead_time_days, created_at, updated_at
FROM erp_products;
