-- name: ListStockTakes :many
SELECT id, owner_id, warehouse_id, status, notes, started_at, completed_at, created_at
FROM erp_stock_takes
WHERE owner_id = ?
ORDER BY created_at DESC;

-- name: GetStockTakeItems :many
SELECT id, stock_take_id, product_id, expected_qty, counted_qty, variance, created_at
FROM erp_stock_take_items
WHERE stock_take_id = ?
ORDER BY id ASC;

-- name: InsertStockTake :execresult
INSERT INTO erp_stock_takes (owner_id, warehouse_id, status, notes)
VALUES (?, ?, 'draft', ?);

-- name: InsertStockTakeItem :execresult
INSERT INTO erp_stock_take_items (stock_take_id, product_id, expected_qty, counted_qty, variance)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateStockTakeStatus :exec
UPDATE erp_stock_takes
SET status = ?, notes = ?
WHERE owner_id = ? AND id = ?;

-- name: UpdateStockTakeStatusInProgress :exec
UPDATE erp_stock_takes
SET status = ?, notes = ?, started_at = NOW()
WHERE owner_id = ? AND id = ?;

-- name: UpdateStockTakeStatusCompleted :exec
UPDATE erp_stock_takes
SET status = ?, notes = ?, completed_at = NOW()
WHERE owner_id = ? AND id = ?;

-- name: DeleteStockTakeItems :exec
DELETE FROM erp_stock_take_items WHERE stock_take_id = ?;
