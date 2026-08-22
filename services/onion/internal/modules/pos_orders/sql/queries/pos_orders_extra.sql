-- name: GetOrderItemsForOrders :many
SELECT id, order_id, COALESCE(menu_item_id,0) AS menu_item_id, name, COALESCE(sku,'') AS sku,
       quantity, unit_price, COALESCE(discount_pct,0) AS discount_pct, COALESCE(discount_amount,0) AS discount_amount,
       COALESCE(tax_rate,0) AS tax_rate, total, COALESCE(notes,'') AS notes, COALESCE(status,'pending') AS status,
       created_at
FROM pos_order_items
WHERE order_id IN (sqlc.slice('order_ids'))
ORDER BY created_at ASC, id ASC;

-- name: GetOrderItemStatus :one
SELECT status FROM pos_order_items WHERE id = ? AND order_id = ?;

-- name: FreeTableByID :exec
UPDATE pos_tables SET status = 'available' WHERE id = ? AND business_id = ?;

-- name: OccupyTableByID :exec
UPDATE pos_tables SET status = 'occupied' WHERE id = ? AND business_id = ?;

-- name: MoveOrderToTable :exec
UPDATE pos_orders SET table_id = ? WHERE id = ? AND business_id = ?;

-- cross-schema read from stock_core — idempotency guard on the shared stock ledger
-- name: HasSaleLedgerEntry :one
SELECT 1 FROM stock_ledger
WHERE business_id = ? AND reference_type = 'pos_order' AND reference_id = ? AND entry_type = 'sale'
LIMIT 1;

-- name: GetOrderLocationID :one
SELECT COALESCE(location_id, 0) FROM pos_orders WHERE business_id = ? AND id = ?;

-- cross-schema read from erp — join POS line items against ERP products to find inventory-tracked SKUs
-- name: ListInventoryTrackedOrderLines :many
SELECT oi.menu_item_id, oi.quantity
FROM pos_order_items oi
JOIN erp_products p ON p.id = oi.menu_item_id AND p.business_id = ?
WHERE oi.order_id = ?
  AND oi.menu_item_id IS NOT NULL
  AND p.track_inventory = TRUE;

-- cross-schema read from erp — public menu item lookup
-- name: GetPublicMenuItem :one
SELECT id, name, sku, sale_price, tax_rate
FROM erp_products
WHERE business_id = ? AND id = ? AND is_active = TRUE;

-- name: CreatePublicOrder :execresult
INSERT INTO pos_orders
(owner_id, business_id, order_number, type, status, customer_id, customer_name, customer_phone,
 subtotal, tax_amount, total, payment_status, notes, created_at, updated_at)
VALUES (?, ?, ?, 'online', 'pending', ?, ?, ?, ?, ?, ?, 'unpaid', ?, NOW(), NOW());

-- name: CreatePublicOrderItem :exec
INSERT INTO pos_order_items
(order_id, menu_item_id, name, quantity, unit_price, total, notes, status, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', NOW());
