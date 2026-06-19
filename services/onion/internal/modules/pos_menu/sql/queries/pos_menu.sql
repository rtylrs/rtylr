-- name: ListMenuItems :many
SELECT id, owner_id, business_id, category_id, name, sku, barcode,
       sale_price, tax_rate, is_active, created_at, updated_at
FROM erp_products
WHERE business_id = ? AND is_active = TRUE
ORDER BY name;

-- name: QuickCreateMenuItem :execresult
INSERT INTO erp_products (owner_id, business_id, name, sku, sale_price, is_active)
VALUES (?, ?, ?, ?, ?, TRUE);
