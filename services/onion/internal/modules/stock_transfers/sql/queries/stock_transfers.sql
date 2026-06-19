-- name: ListStockTransfers :many
SELECT id, owner_id, business_id, from_warehouse_id, to_warehouse_id, status, notes, created_at, updated_at
FROM erp_stock_transfers
WHERE business_id = ?
ORDER BY created_at DESC;

-- name: ListStockTransfersPaged :many
SELECT id, owner_id, business_id, from_warehouse_id, to_warehouse_id, status, notes, created_at, updated_at
FROM erp_stock_transfers
WHERE business_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetStockTransferForUpdate :one
SELECT from_warehouse_id, to_warehouse_id, status
FROM erp_stock_transfers
WHERE business_id = ? AND id = ?
FOR UPDATE;

-- name: GetTransferItems :many
SELECT id, transfer_id, product_id, quantity, created_at
FROM erp_stock_transfer_items
WHERE transfer_id = ?
ORDER BY id ASC;

-- name: InsertStockTransfer :execresult
INSERT INTO erp_stock_transfers (owner_id, business_id, from_warehouse_id, to_warehouse_id, status, notes)
VALUES (?, ?, ?, ?, 'draft', ?);

-- name: InsertStockTransferItem :execresult
INSERT INTO erp_stock_transfer_items (transfer_id, product_id, quantity)
VALUES (?, ?, ?);

-- name: UpdateStockTransferStatus :exec
UPDATE erp_stock_transfers
SET status = ?
WHERE business_id = ? AND id = ?;
