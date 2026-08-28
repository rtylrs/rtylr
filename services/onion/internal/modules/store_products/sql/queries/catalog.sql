-- cross-schema read from erp_products + erp_product_images — catalog enrichment for storefront list
-- name: ListStoreProductsWithCatalog :many
WITH store_product_page AS (
    SELECT
        id,
        owner_id,
        business_id,
        product_id,
        title,
        slug,
        description,
        images,
        price,
        compare_at_price,
        is_published,
        sort_order,
        created_at,
        updated_at
    FROM store_products store_product
    WHERE store_product.business_id = ?
    ORDER BY store_product.sort_order ASC, store_product.id ASC
    LIMIT 200
),
catalog_product_page AS (
    SELECT DISTINCT page.business_id, page.product_id
    FROM store_product_page page
    WHERE page.product_id IS NOT NULL
),
catalog_images AS (
    SELECT images.business_id, images.product_id, JSON_ARRAYAGG(images.url) AS images
    FROM erp_product_images images
    INNER JOIN catalog_product_page page
      ON page.business_id = images.business_id AND page.product_id = images.product_id
    GROUP BY images.business_id, images.product_id
)
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
FROM store_product_page sp
LEFT JOIN erp_products ep
  ON ep.business_id = sp.business_id AND ep.id = sp.product_id
LEFT JOIN catalog_images epi
  ON epi.business_id = sp.business_id AND epi.product_id = sp.product_id
ORDER BY sp.sort_order ASC, sp.id ASC;

-- cross-schema read from erp_products — catalog product lookup for storefront sync
-- name: GetCatalogProductByID :one
SELECT id
FROM erp_products
WHERE business_id = ? AND id = ?
ORDER BY id ASC
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
    owner_id,
    business_id,
    name,
    description,
    sale_price,
    compare_at_price,
    is_active,
    track_inventory
) VALUES (?, ?, ?, ?, ?, ?, TRUE, TRUE);

-- cross-schema read from store_products — linked product id for update flow
-- name: GetStoreProductLinkedID :one
SELECT product_id
FROM store_products
WHERE business_id = ? AND id = ?
ORDER BY id ASC
LIMIT 1;

-- cross-schema write to erp_product_images — clear images before re-sync
-- name: DeleteCatalogProductImages :exec
DELETE FROM erp_product_images
WHERE business_id = ? AND product_id = ?;

-- cross-schema write to erp_product_images — insert synced catalog image
-- name: InsertCatalogProductImage :exec
INSERT INTO erp_product_images (
    owner_id,
    business_id,
    product_id,
    url,
    alt,
    position,
    is_primary
) VALUES (?, ?, ?, ?, ?, ?, ?);
