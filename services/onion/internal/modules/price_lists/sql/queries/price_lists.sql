-- name: ListPriceLists :many
SELECT id, owner_id, name, currency, is_default, created_at, updated_at
FROM erp_price_lists
WHERE owner_id = ?
ORDER BY name ASC;

-- name: CreatePriceList :execresult
INSERT INTO erp_price_lists (owner_id, name, currency, is_default)
VALUES (?, ?, ?, ?);
