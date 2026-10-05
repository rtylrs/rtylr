-- name: ListPurchaseOrders :many
SELECT id, owner_id, business_id, supplier_id, po_number, status, subtotal, tax_amount, total,
       expected_delivery, received_at, notes, created_by, created_at, updated_at, destination_location_id
FROM erp_purchase_orders
WHERE business_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListPurchaseOrdersByStatus :many
SELECT id, owner_id, business_id, supplier_id, po_number, status, subtotal, tax_amount, total,
       expected_delivery, received_at, notes, created_by, created_at, updated_at, destination_location_id
FROM erp_purchase_orders
WHERE business_id = ? AND status = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetPurchaseOrderByID :one
SELECT id, owner_id, business_id, supplier_id, po_number, status, subtotal, tax_amount, total,
       expected_delivery, received_at, notes, created_by, created_at, updated_at, destination_location_id
FROM erp_purchase_orders
WHERE id = ? AND business_id = ?;

-- name: GetPurchaseOrderItems :many
SELECT id, purchase_order_id, product_id, quantity_ordered, quantity_received, unit_cost, total_cost
FROM erp_purchase_order_items
WHERE purchase_order_id = ?;

-- name: InsertPurchaseOrder :execresult
INSERT INTO erp_purchase_orders
(owner_id, business_id, supplier_id, po_number, status, subtotal, tax_amount, total,
 expected_delivery, notes, created_by)
VALUES (?, ?, ?, ?, 'draft', ?, ?, ?, ?, ?, ?);

-- name: GetSupplierReference :one
-- cross-schema read from erp_suppliers (suppliers module) — validate the purchase order’s supplier reference for the tenant.
SELECT id FROM erp_suppliers WHERE business_id = ? AND id = ?;

-- name: GetProductReference :one
-- cross-schema read from erp_products (products module) — validate the purchase order item’s product reference for the tenant.
SELECT id FROM erp_products WHERE business_id = ? AND id = ?;

-- name: InsertPurchaseOrderItem :execresult
INSERT INTO erp_purchase_order_items
(purchase_order_id, product_id, quantity_ordered, quantity_received, unit_cost, total_cost)
VALUES (?, ?, ?, 0, ?, ?);

-- name: UpdatePurchaseOrderStatus :execresult
UPDATE erp_purchase_orders
SET status = ?
WHERE business_id = ? AND id = ? AND (status IS NULL OR status <> 'received');

-- name: UpdatePurchaseOrderStatusReceived :execresult
UPDATE erp_purchase_orders SET status = ?, received_at = NOW() WHERE business_id = ? AND id = ?;

-- name: UpdatePurchaseOrderStatusPartial :exec
UPDATE erp_purchase_orders SET status = ? WHERE business_id = ? AND id = ?;

-- name: GetPurchaseOrderStatus :one
SELECT status FROM erp_purchase_orders WHERE business_id = ? AND id = ?;

-- name: ApprovePurchaseOrder :execresult
UPDATE erp_purchase_orders
SET status = 'approved'
WHERE business_id = ? AND id = ? AND status IN ('draft', 'sent', 'partial');

-- name: GetPurchaseOrderForUpdate :one
SELECT status FROM erp_purchase_orders WHERE business_id = ? AND id = ? FOR UPDATE;

-- name: InsertInventoryTransactionWithRef :execresult
-- cross-schema write to erp_inventory_transactions (inventory module) — record a received purchase quantity in the inventory ledger with its purchase-order reference.
INSERT INTO erp_inventory_transactions
(owner_id, business_id, product_id, type, quantity_change, quantity_before, quantity_after,
 reference_id, reference_type, created_by, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertInventoryTransaction :execresult
-- cross-schema write to erp_inventory_transactions (inventory module) — record a received purchase quantity in the inventory ledger.
INSERT INTO erp_inventory_transactions
(owner_id, business_id, product_id, type, quantity_change, quantity_before, quantity_after, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: PurchaseOrderExists :one
SELECT EXISTS(SELECT 1 FROM erp_purchase_orders WHERE business_id = ? AND id = ?) AS present;
