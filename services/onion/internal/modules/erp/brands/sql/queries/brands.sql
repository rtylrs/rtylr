-- name: ListBrands :many
WITH brand_page AS (
    SELECT b.id, b.owner_id, b.business_id, b.name, b.logo_url, b.description, b.created_at, b.updated_at
    FROM erp_brands b
    WHERE b.business_id = ?
), product_counts AS (
    -- cross-schema read from erp_products (products module) — include the number of products assigned to each brand.
    SELECT p.brand_id, COUNT(*) AS product_count
    FROM erp_products p FORCE INDEX (idx_erp_products_business_brand)
    JOIN brand_page b ON b.id = p.brand_id AND b.business_id = p.business_id
    GROUP BY p.brand_id
)
SELECT b.id, b.owner_id, b.business_id, b.name, b.logo_url, b.description,
       COALESCE(pc.product_count, 0) AS product_count,
       b.created_at, b.updated_at
FROM brand_page b
LEFT JOIN product_counts pc ON pc.brand_id = b.id
ORDER BY b.name ASC, b.id ASC;

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
WITH brand_page AS (
    SELECT b.id, b.owner_id, b.business_id, b.name, b.logo_url, b.description, b.created_at, b.updated_at
    FROM erp_brands b
    WHERE b.business_id = ?
    ORDER BY b.name ASC, b.id ASC
    LIMIT ? OFFSET ?
), product_counts AS (
    -- cross-schema read from erp_products (products module) — include the number of products assigned to each brand.
    SELECT p.brand_id, COUNT(*) AS product_count
    FROM erp_products p FORCE INDEX (idx_erp_products_business_brand)
    JOIN brand_page b ON b.id = p.brand_id AND b.business_id = p.business_id
    GROUP BY p.brand_id
)
SELECT b.id, b.owner_id, b.business_id, b.name, b.logo_url, b.description,
       COALESCE(pc.product_count, 0) AS product_count,
       b.created_at, b.updated_at
FROM brand_page b
LEFT JOIN product_counts pc ON pc.brand_id = b.id
ORDER BY b.name ASC, b.id ASC;

-- name: CountBrands :one
SELECT COUNT(*) FROM erp_brands WHERE business_id = ?;
