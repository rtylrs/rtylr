-- name: GetDefaultWarehouseLocationID :one
SELECT id FROM stock_locations
 WHERE owner_id = ? AND kind = 'warehouse' AND source_table = 'native' AND source_id = 0
 LIMIT 1;

-- name: InsertDefaultWarehouseLocation :exec
INSERT INTO stock_locations (owner_id, business_id, kind, source_table, source_id, name, is_default, is_active)
 VALUES (?, ?, 'warehouse', 'native', 0, 'Default Warehouse', TRUE, TRUE)
 ON DUPLICATE KEY UPDATE id = id;

-- name: GetPOSLocationOwnerID :one
SELECT owner_id FROM pos_locations WHERE id = ?;

-- name: GetPOSBranchLocationID :one
SELECT id FROM stock_locations
 WHERE owner_id = ? AND source_table = 'pos_locations' AND source_id = ? LIMIT 1;

-- name: InsertPOSBranchLocation :exec
INSERT INTO stock_locations (owner_id, business_id, kind, source_table, source_id, name, is_default, is_active)
 VALUES (?, ?, 'pos_branch', 'pos_locations', ?, 'POS Branch', FALSE, TRUE)
 ON DUPLICATE KEY UPDATE source_id = VALUES(source_id);

-- name: GetWarehouseOwnerID :one
SELECT owner_id FROM erp_warehouses WHERE id = ?;

-- name: GetWarehouseLocationID :one
SELECT id FROM stock_locations
 WHERE owner_id = ? AND source_table = 'erp_warehouses' AND source_id = ? LIMIT 1;

-- name: InsertWarehouseLocation :exec
INSERT INTO stock_locations (owner_id, business_id, kind, source_table, source_id, name, is_default, is_active)
 SELECT ?, ?, 'warehouse', 'erp_warehouses', w.id, w.name, FALSE, w.is_active FROM erp_warehouses w WHERE w.id = ?
 ON DUPLICATE KEY UPDATE source_id = VALUES(source_id);

-- name: EnsureStockLevelRow :exec
INSERT INTO stock_levels (owner_id, business_id, location_id, product_id, on_hand)
 VALUES (?, ?, ?, ?, 0) ON DUPLICATE KEY UPDATE on_hand = on_hand;

-- name: LockStockLevelOnHand :one
SELECT on_hand FROM stock_levels
 WHERE owner_id = ? AND location_id = ? AND product_id = ? FOR UPDATE;

-- name: UpdateStockLevelOnHand :exec
UPDATE stock_levels SET on_hand = ?
 WHERE owner_id = ? AND location_id = ? AND product_id = ?;

-- name: InsertStockLedgerEntry :exec
INSERT INTO stock_ledger
 (owner_id, business_id, location_id, product_id, entry_type, quantity_change, quantity_before, quantity_after,
  reference_id, reference_type, transfer_group, channel, note, employee_id, created_by)
 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: SyncProductStockScalar :exec
UPDATE erp_products SET stock_quantity = stock_quantity + ?
 WHERE owner_id = ? AND id = ?;
