-- cross-schema read from erp_products + erp_product_images — catalog enrichment for storefront list
-- name: ListStoreProductsWithCatalog :many
SELECT
    sp.id,
    sp.owner_id,
    sp.business_id,
    sp.product_id,
    COALESCE(ep.name, sp.title) AS title,
    sp.slug,
    COALESCE(ep.description, sp.description) AS description,
    COALESCE(epi.images, sp.images, JSON_ARRAY()) AS images,
    COALESCE(ep.sale_price, sp.price) AS price,
    COALESCE(ep.compare_at_price, sp.compare_at_price) AS compare_at_price,
    sp.is_published,
    sp.sort_order,
    sp.created_at,
    sp.updated_at
FROM store_products sp
LEFT JOIN erp_products ep
  ON ep.business_id = sp.business_id AND ep.id = sp.product_id
LEFT JOIN (
    SELECT images.business_id, images.product_id, JSON_ARRAYAGG(images.url) AS images
    FROM erp_product_images images
    WHERE images.business_id = ?
    GROUP BY images.business_id, images.product_id
) epi
  ON epi.business_id = sp.business_id AND epi.product_id = sp.product_id
WHERE sp.business_id = ?
ORDER BY sp.sort_order ASC, sp.id ASC;

-- cross-schema read from erp_products — catalog product lookup for storefront sync
-- name: GetCatalogProductByID :one
SELECT id
FROM erp_products
WHERE business_id = ? AND id = ?
LIMIT 1;

-- cross-schema write to erp_products — update linked catalog product from storefront
-- name: UpdateCatalogProduct :exec
UPDATE erp_products
SET
    name = ?,
    description = ?,
    sale_price = ?,
    compare_at_price = ?,
    is_active = TRUE
WHERE business_id = ? AND id = ?;

-- cross-schema write to erp_products — create catalog product from storefront
-- name: InsertCatalogProduct :execresult
INSERT INTO erp_products (
    business_id,
    name,
    description,
    sale_price,
    compare_at_price,
    is_active,
    track_inventory
) VALUES (?, ?, ?, ?, ?, TRUE, TRUE);

-- cross-schema read from store_products — linked product id for update flow
-- name: GetStoreProductLinkedID :one
SELECT product_id
FROM store_products
WHERE business_id = ? AND id = ?
LIMIT 1;

-- cross-schema write to erp_product_images — clear images before re-sync
-- name: DeleteCatalogProductImages :exec
DELETE FROM erp_product_images
WHERE business_id = ? AND product_id = ?;

-- cross-schema write to erp_product_images — insert synced catalog image
-- name: InsertCatalogProductImage :exec
INSERT INTO erp_product_images (
    business_id,
    product_id,
    url,
    alt,
    position,
    is_primary
) VALUES (?, ?, ?, ?, ?, ?);
