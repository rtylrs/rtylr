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

-- name: OrderExists :one
SELECT COUNT(*) FROM pos_orders WHERE business_id = ? AND id = ?;

-- name: OrderItemExists :one
SELECT COUNT(*) FROM pos_order_items WHERE order_id = ? AND id = ?;

-- name: FreeTableByID :exec
UPDATE pos_tables SET status = 'available' WHERE id = ? AND business_id = ?;

-- name: OccupyTableByID :execresult
UPDATE pos_tables SET status = 'occupied' WHERE id = ? AND business_id = ?;

-- name: MoveOrderToTable :exec
UPDATE pos_orders SET table_id = ? WHERE id = ? AND business_id = ?;

-- name: GetOrderLocationID :one
SELECT COALESCE(location_id, 0) FROM pos_orders WHERE business_id = ? AND id = ?;

-- name: ListPaidOrderProductLines :many
SELECT oi.menu_item_id, oi.quantity
FROM pos_order_items oi
JOIN pos_orders o ON o.id = oi.order_id
WHERE o.business_id = ? AND o.id = ? AND oi.menu_item_id IS NOT NULL
ORDER BY oi.id;

-- name: GetPOSComboForOrder :one
SELECT name, price
FROM pos_combos
WHERE business_id = ? AND id = ? AND is_active = TRUE;

-- name: GetPOSModifierTotal :one
SELECT COALESCE(SUM(price_delta), 0) AS price_delta, COUNT(*) AS modifier_count
FROM pos_modifiers
WHERE business_id = ? AND id IN (sqlc.slice('modifier_ids'));

-- name: GetPOSUserCapabilities :one
SELECT can_apply_discount, max_discount_pct
FROM pos_users
WHERE business_id = ?
  AND user_id = ?
  AND is_active = TRUE
  AND (location_id = ? OR ? = 0)
ORDER BY CASE WHEN location_id = ? THEN 0 ELSE 1 END, id
LIMIT 1;

-- name: CreatePublicOrder :execresult
INSERT INTO pos_orders
(owner_id, business_id, order_number, type, status, customer_id, customer_name, customer_phone,
 subtotal, tax_amount, total, payment_status, notes, created_at, updated_at)
VALUES (?, ?, ?, 'online', 'pending', ?, ?, ?, ?, ?, ?, 'unpaid', ?, NOW(), NOW());

-- name: CreatePublicOrderItem :exec
INSERT INTO pos_order_items
(order_id, menu_item_id, name, quantity, unit_price, total, notes, status, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', NOW());

-- name: EnsureOrderSequence :exec
INSERT INTO pos_order_sequences (business_id, prefix, last_sequence) VALUES (?, ?, 0)
ON DUPLICATE KEY UPDATE last_sequence = last_sequence;

-- name: LockOrderSequence :one
SELECT last_sequence FROM pos_order_sequences WHERE business_id = ? AND prefix = ? FOR UPDATE;

-- name: AdvanceOrderSequence :exec
UPDATE pos_order_sequences SET last_sequence = ? WHERE business_id = ? AND prefix = ?;
