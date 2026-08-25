-- name: ListMenuItemsWithDetails :many
SELECT
    p.id,
    p.owner_id,
    p.business_id,
    p.category_id,
    COALESCE(c.name, 'Other') AS category_name,
    p.name,
    p.sku,
    p.barcode,
    p.sale_price,
    p.cost_price,
    p.tax_rate,
    p.description,
    COALESCE(img.url, '') AS image_url,
    COALESCE(pst.total_quantity, p.stock_quantity) AS stock_quantity,
    p.is_active,
    p.created_at,
    p.updated_at
FROM erp_products p
LEFT JOIN erp_categories c
  ON c.id = p.category_id
 AND c.business_id = p.business_id
LEFT JOIN product_stock_total pst
  ON pst.business_id = p.business_id
 AND pst.product_id = p.id
LEFT JOIN (
    SELECT business_id, product_id, url
    FROM (
        SELECT
            epi.business_id,
            epi.product_id,
            epi.url,
            ROW_NUMBER() OVER (
                PARTITION BY epi.business_id, epi.product_id
                ORDER BY epi.is_primary DESC, epi.position ASC, epi.id ASC
            ) AS image_rank
        FROM erp_product_images epi
        WHERE epi.business_id = ?
    ) ranked_product_images
    WHERE image_rank = 1
) img
  ON img.business_id = p.business_id
 AND img.product_id = p.id
WHERE p.business_id = ?
  AND p.is_active = TRUE
ORDER BY c.name, p.name
LIMIT 100;

-- name: GetMenuBusinessOwnerID :one
SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE;

-- name: ListMenuItems :many
SELECT id, owner_id, business_id, category_id, name, sku, barcode,
       sale_price, tax_rate, is_active, created_at, updated_at
FROM erp_products
WHERE business_id = ? AND is_active = TRUE
ORDER BY name
LIMIT 100;

-- name: QuickCreateMenuItem :execresult
INSERT INTO erp_products (owner_id, business_id, name, sku, sale_price, is_active)
VALUES (?, ?, ?, ?, ?, TRUE);

-- name: UpdateMenuItem :execresult
UPDATE erp_products
SET name = COALESCE(NULLIF(sqlc.arg(name), ''), name),
    sale_price = ?,
    is_active = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE business_id = ? AND id = ?;
