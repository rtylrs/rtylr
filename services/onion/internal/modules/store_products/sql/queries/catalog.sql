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
    SELECT business_id, product_id, JSON_ARRAYAGG(url) AS images
    FROM erp_product_images
    GROUP BY business_id, product_id
) epi
  ON epi.business_id = sp.business_id AND epi.product_id = sp.product_id
WHERE sp.business_id = ?
ORDER BY sp.sort_order ASC;
