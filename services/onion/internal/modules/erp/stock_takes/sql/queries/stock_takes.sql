-- name: ListStockTakes :many
SELECT id, owner_id, business_id, warehouse_id, status, notes, started_at, completed_at, created_at
FROM erp_stock_takes
WHERE business_id = ?
ORDER BY created_at DESC;

-- name: ListStockTakesPaged :many
SELECT id, owner_id, business_id, warehouse_id, status, notes, started_at, completed_at, created_at
FROM erp_stock_takes
WHERE business_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: GetStockTakeItems :many
SELECT id, stock_take_id, product_id, expected_qty, counted_qty, variance, created_at
FROM erp_stock_take_items
WHERE stock_take_id = ?
ORDER BY id ASC;

-- name: GetStockTakeItemsByStockTakeIDs :many
SELECT id, stock_take_id, product_id, expected_qty, counted_qty, variance, created_at
FROM erp_stock_take_items
WHERE stock_take_id IN (sqlc.slice('stock_take_ids'))
ORDER BY id ASC;

-- name: InsertStockTake :execresult
INSERT INTO erp_stock_takes (owner_id, business_id, warehouse_id, status, notes)
VALUES (?, ?, ?, 'draft', ?);

-- name: GetWarehouseReference :one
-- cross-schema read from erp_warehouses (warehouses module) — validate the stock-take warehouse reference for the tenant.
SELECT id FROM erp_warehouses WHERE business_id = ? AND id = ?;

-- name: GetProductReference :one
-- cross-schema read from erp_products (products module) — validate the stock-take product reference for the tenant.
SELECT id FROM erp_products WHERE business_id = ? AND id = ?;

-- name: InsertStockTakeItem :execresult
INSERT INTO erp_stock_take_items (stock_take_id, product_id, expected_qty, counted_qty, variance)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateStockTakeStatus :exec
UPDATE erp_stock_takes
SET status = ?, notes = ?
WHERE business_id = ? AND id = ?;

-- name: UpdateStockTakeStatusInProgress :exec
UPDATE erp_stock_takes
SET status = ?, notes = ?, started_at = NOW()
WHERE business_id = ? AND id = ?;

-- name: UpdateStockTakeStatusCompleted :exec
UPDATE erp_stock_takes
SET status = ?, notes = ?, completed_at = NOW()
WHERE business_id = ? AND id = ?;

-- name: UpdateStockTakeItemCount :execrows
UPDATE erp_stock_take_items
SET counted_qty = ?, variance = ?
WHERE stock_take_id = ? AND product_id = ?;

-- name: GetStockTakeStatusForUpdate :one
SELECT warehouse_id, status
FROM erp_stock_takes
WHERE business_id = ? AND id = ?
FOR UPDATE;

-- name: InsertInventoryTransactionWithRef :execresult
-- cross-schema write to erp_inventory_transactions (inventory module) — record a stock-take adjustment in the inventory ledger with its stock-take reference.
INSERT INTO erp_inventory_transactions
(owner_id, business_id, product_id, type, quantity_change, quantity_before, quantity_after,
 reference_id, reference_type, created_by, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
