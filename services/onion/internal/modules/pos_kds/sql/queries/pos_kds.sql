-- name: ListActiveKDSOrders :many
-- cross-schema read from pos_orders (POS orders module) — list active orders for the kitchen display.
SELECT id, order_number, type, status, table_id, customer_name, kitchen_notes, created_at, owner_id, business_id
FROM pos_orders
WHERE business_id = ? AND status IN ('confirmed','preparing','ready')
ORDER BY created_at ASC;

-- name: GetKDSOrderItems :many
-- cross-schema read from pos_order_items (POS orders module) — load the kitchen items belonging to an order.
SELECT id, order_id, name, quantity, notes, status
FROM pos_order_items WHERE order_id = ? ORDER BY id;

-- name: GetKDSOrderItemsForOrders :many
-- cross-schema read from pos_order_items (POS orders module) — load kitchen items for a batch of orders.
SELECT id, order_id, name, quantity, notes, status
FROM pos_order_items
WHERE order_id IN (sqlc.slice('order_ids'))
ORDER BY id;

-- name: UpdateKDSItemStatus :exec
-- cross-schema write to pos_order_items (POS orders module) — apply a kitchen status transition to an item.
UPDATE pos_order_items SET status = ? WHERE order_id = ? AND id = ?;

-- name: UpdateKDSItemStatusByItemID :exec
-- cross-schema read from pos_orders and write to pos_order_items (POS orders module) — authorize and apply a kitchen status transition by item ID.
UPDATE pos_order_items oi
JOIN pos_orders o ON o.id = oi.order_id
SET oi.status = ?
WHERE oi.id = ? AND o.business_id = ?;
