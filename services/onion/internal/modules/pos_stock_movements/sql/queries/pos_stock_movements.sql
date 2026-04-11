-- name: ListStockMovements :many
SELECT id, business_id, branch_id, product_id, type, quantity, reference_id, reference_type,
       note, employee_id, created_at
FROM pos_stock_movements
WHERE business_id = ? AND branch_id = ? AND product_id = ?
ORDER BY created_at DESC;

-- name: CreateStockMovement :execresult
INSERT INTO pos_stock_movements (business_id, branch_id, product_id, type, quantity, reference_id, reference_type, note, employee_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListBranchInventory :many
SELECT id, business_id, branch_id, product_id, quantity, reorder_point, reorder_quantity, updated_at
FROM pos_branch_inventory
WHERE business_id = ? AND branch_id = ?
ORDER BY product_id;

-- name: GetBranchInventoryItem :one
SELECT id, business_id, branch_id, product_id, quantity, reorder_point, reorder_quantity, updated_at
FROM pos_branch_inventory
WHERE business_id = ? AND branch_id = ? AND product_id = ?;

-- name: UpsertBranchInventory :execresult
INSERT INTO pos_branch_inventory (business_id, branch_id, product_id, quantity)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE quantity = quantity + VALUES(quantity);

-- name: SetBranchInventoryQuantity :execresult
INSERT INTO pos_branch_inventory (business_id, branch_id, product_id, quantity)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE quantity = VALUES(quantity);

-- name: ListLowStockItems :many
SELECT id, business_id, branch_id, product_id, quantity, reorder_point, reorder_quantity, updated_at
FROM pos_branch_inventory
WHERE business_id = ? AND branch_id = ? AND reorder_point IS NOT NULL AND quantity <= reorder_point
ORDER BY quantity ASC;
