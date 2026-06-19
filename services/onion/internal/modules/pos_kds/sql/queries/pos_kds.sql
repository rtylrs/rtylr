-- name: ListActiveKDSOrders :many
SELECT id, order_number, type, status, table_id, customer_name, kitchen_notes, created_at, owner_id, business_id
FROM pos_orders
WHERE business_id = ? AND status IN ('confirmed','preparing','ready')
ORDER BY created_at ASC;

-- name: GetKDSOrderItems :many
SELECT id, order_id, name, quantity, notes, status
FROM pos_order_items WHERE order_id = ? ORDER BY id;

-- name: UpdateKDSItemStatus :exec
UPDATE pos_order_items SET status = ? WHERE order_id = ? AND id = ?;

-- name: UpdateKDSItemStatusByItemID :exec
UPDATE pos_order_items oi
JOIN pos_orders o ON o.id = oi.order_id
SET oi.status = ?
WHERE oi.id = ? AND o.business_id = ?;
