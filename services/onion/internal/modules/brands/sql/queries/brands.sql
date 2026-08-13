-- name: ListBrands :many
SELECT b.id, b.owner_id, b.business_id, b.name, b.logo_url, b.description,
       COALESCE(products.product_count, 0) AS product_count,
       b.created_at, b.updated_at
FROM erp_brands b
LEFT JOIN (
    SELECT brand_id, COUNT(*) AS product_count
    FROM erp_products p
    WHERE p.business_id = ?
    GROUP BY brand_id
) AS products ON products.brand_id = b.id
WHERE b.business_id = ?
ORDER BY b.name ASC;

-- name: CreateBrand :execresult
INSERT INTO erp_brands (owner_id, business_id, name, logo_url, description)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateBrand :exec
UPDATE erp_brands
SET name = ?, logo_url = ?, description = ?
WHERE business_id = ? AND id = ?;

-- name: DeleteBrand :exec
DELETE FROM erp_brands WHERE business_id = ? AND id = ?;

-- name: ListBrandsPaginated :many
SELECT b.id, b.owner_id, b.business_id, b.name, b.logo_url, b.description,
       COALESCE(products.product_count, 0) AS product_count,
       b.created_at, b.updated_at
FROM erp_brands b
LEFT JOIN (
    SELECT brand_id, COUNT(*) AS product_count
    FROM erp_products p
    WHERE p.business_id = ?
    GROUP BY brand_id
) AS products ON products.brand_id = b.id
WHERE b.business_id = ?
ORDER BY b.name ASC
LIMIT ? OFFSET ?;
