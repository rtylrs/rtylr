-- name: ListPurchaseOrders :many
SELECT id, owner_id, supplier_id, po_number, status, subtotal, tax_amount, total,
       expected_delivery, received_at, notes, created_by, created_at, updated_at
FROM erp_purchase_orders
WHERE owner_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListPurchaseOrdersByStatus :many
SELECT id, owner_id, supplier_id, po_number, status, subtotal, tax_amount, total,
       expected_delivery, received_at, notes, created_by, created_at, updated_at
FROM erp_purchase_orders
WHERE owner_id = ? AND status = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetPurchaseOrderByID :one
SELECT id, owner_id, supplier_id, po_number, status, subtotal, tax_amount, total,
       expected_delivery, received_at, notes, created_by, created_at, updated_at
FROM erp_purchase_orders
WHERE id = ? AND owner_id = ?;

-- name: GetPurchaseOrderItems :many
SELECT id, purchase_order_id, product_id, quantity_ordered, quantity_received, unit_cost, total_cost
FROM erp_purchase_order_items
WHERE purchase_order_id = ?;

-- name: InsertPurchaseOrder :execresult
INSERT INTO erp_purchase_orders
(owner_id, supplier_id, po_number, status, subtotal, tax_amount, total,
 expected_delivery, notes, created_by)
VALUES (?, ?, ?, 'draft', ?, ?, ?, ?, ?, ?);

-- name: InsertPurchaseOrderItem :execresult
INSERT INTO erp_purchase_order_items
(purchase_order_id, product_id, quantity_ordered, quantity_received, unit_cost, total_cost)
VALUES (?, ?, ?, 0, ?, ?);

-- name: UpdatePurchaseOrderStatus :exec
UPDATE erp_purchase_orders SET status = ? WHERE owner_id = ? AND id = ?;

-- name: UpdatePurchaseOrderStatusReceived :exec
UPDATE erp_purchase_orders SET status = ?, received_at = NOW() WHERE owner_id = ? AND id = ?;

-- name: GetPurchaseOrderStatus :one
SELECT status FROM erp_purchase_orders WHERE owner_id = ? AND id = ?;

-- name: ApprovePurchaseOrder :exec
UPDATE erp_purchase_orders SET status = 'approved' WHERE owner_id = ? AND id = ?;

-- name: GetPurchaseOrderForUpdate :one
SELECT status FROM erp_purchase_orders WHERE owner_id = ? AND id = ? FOR UPDATE;

-- name: MarkPurchaseOrderItemsReceived :exec
UPDATE erp_purchase_order_items SET quantity_received = quantity_ordered
WHERE purchase_order_id = ?;

-- name: InsertInventoryTransactionWithRef :execresult
INSERT INTO erp_inventory_transactions
(owner_id, product_id, type, quantity_change, quantity_before, quantity_after,
 reference_id, reference_type, created_by, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertInventoryTransaction :execresult
INSERT INTO erp_inventory_transactions
(owner_id, product_id, type, quantity_change, quantity_before, quantity_after, notes)
VALUES (?, ?, ?, ?, ?, ?, ?);
