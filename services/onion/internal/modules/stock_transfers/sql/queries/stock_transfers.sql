-- name: ListStockTransfers :many
SELECT id, owner_id, business_id, from_warehouse_id, to_warehouse_id, status, notes, created_at, updated_at
FROM erp_stock_transfers
WHERE business_id = ?
ORDER BY created_at DESC
LIMIT 200;

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

-- name: GetWarehouseReference :one
SELECT id FROM erp_warehouses WHERE business_id = ? AND id = ?;

-- name: GetProductReference :one
SELECT id FROM erp_products WHERE business_id = ? AND id = ?;

-- name: InsertStockTransferItem :execresult
INSERT INTO erp_stock_transfer_items (transfer_id, product_id, quantity)
VALUES (?, ?, ?);

-- name: UpdateStockTransferStatus :execresult
UPDATE erp_stock_transfers
SET status = ?
WHERE business_id = ? AND id = ?
  AND (status IS NULL OR status <> 'received');

-- name: GetStockTransferByID :one
SELECT id, owner_id, business_id, from_warehouse_id, to_warehouse_id, status, notes, created_at, updated_at
FROM erp_stock_transfers
WHERE business_id = ? AND id = ?
ORDER BY id ASC
LIMIT 1;

-- name: ListTransferItemsByTransferIDs :many
SELECT id, transfer_id, product_id, quantity, created_at
FROM erp_stock_transfer_items
WHERE transfer_id IN (sqlc.slice('transfer_ids'))
ORDER BY id ASC
LIMIT 200;

