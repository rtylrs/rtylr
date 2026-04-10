-- name: ListBrands :many
SELECT b.id, b.owner_id, b.name, b.logo_url, b.description,
       (SELECT COUNT(*) FROM erp_products p WHERE p.owner_id = b.owner_id AND p.brand_id = b.id) AS product_count,
       b.created_at, b.updated_at
FROM erp_brands b
WHERE b.owner_id = ?
ORDER BY b.name ASC;

-- name: CreateBrand :execresult
INSERT INTO erp_brands (owner_id, name, logo_url, description)
VALUES (?, ?, ?, ?);

-- name: UpdateBrand :exec
UPDATE erp_brands
SET name = ?, logo_url = ?, description = ?
WHERE owner_id = ? AND id = ?;

-- name: DeleteBrand :exec
DELETE FROM erp_brands WHERE owner_id = ? AND id = ?;
