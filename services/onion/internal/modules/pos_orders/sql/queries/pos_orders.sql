-- name: ListLocationsByOwner :many
SELECT id, owner_id, business_id, name, address, phone, is_active, created_at
FROM pos_locations WHERE business_id = ? ORDER BY name ASC LIMIT 500;

-- name: CreateLocation :execresult
INSERT INTO pos_locations (owner_id, business_id, name, address, phone) VALUES (?, ?, ?, ?, ?);

-- name: ListTablesByOwner :many
SELECT id, owner_id, business_id, location_id, name, capacity, status,
       current_session_order_id, qr_code_url, created_at, updated_at
FROM pos_tables
WHERE business_id = sqlc.arg(business_id)
  AND (sqlc.arg(location_id) = 0 OR location_id = sqlc.arg(location_id))
  AND (sqlc.arg(status) = '' OR status = sqlc.arg(status))
ORDER BY name ASC, id ASC;

-- name: CreatePosTable :execresult
INSERT INTO pos_tables (owner_id, business_id, location_id, name, capacity, qr_code_url)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdatePosTable :exec
UPDATE pos_tables SET location_id = ?, name = ?, capacity = ?, qr_code_url = ?
WHERE business_id = ? AND id = ?;

-- name: UpdateTableStatus :exec
UPDATE pos_tables SET status = ? WHERE business_id = ? AND id = ?;

-- name: DeletePosTable :exec
DELETE FROM pos_tables WHERE business_id = ? AND id = ?;

-- name: SetTableOccupied :execresult
UPDATE pos_tables SET status = 'occupied', current_session_order_id = ? WHERE business_id = ? AND id = ?;

-- name: FreeTable :exec
UPDATE pos_tables SET status = 'available', current_session_order_id = NULL WHERE id = ? AND business_id = ?;

-- name: GetCurrentOpenSession :one
SELECT id, owner_id, business_id, location_id, staff_id, status,
       opening_cash, closing_cash, expected_cash, cash_difference,
       total_sales, total_orders, total_cash_sales, total_card_sales,
       total_online_sales, total_refunds, opened_at, closed_at, notes,
       terminal_id
FROM pos_sessions
WHERE business_id = ? AND staff_id = ? AND status = 'open'
LIMIT 1;

-- name: OpenSession :execresult
INSERT INTO pos_sessions (owner_id, business_id, location_id, staff_id, status, opening_cash)
VALUES (?, ?, ?, ?, 'open', ?);

-- name: CloseSession :exec
UPDATE pos_sessions
SET status = 'closed', closing_cash = ?, expected_cash = ?, cash_difference = ?,
    closed_at = NOW(), notes = ?
WHERE business_id = ? AND id = ? AND status = 'open';

-- name: GetSessionTotals :one
SELECT
    COALESCE(SUM(total), 0) AS total_sales,
    COALESCE(SUM(CASE WHEN payment_method = 'cash' THEN total ELSE 0 END), 0) AS cash_sales,
    COALESCE(SUM(CASE WHEN payment_method = 'card' THEN total ELSE 0 END), 0) AS card_sales,
    COALESCE(SUM(CASE WHEN payment_method = 'online' THEN total ELSE 0 END), 0) AS online_sales,
    COUNT(*) AS total_orders
FROM pos_orders
WHERE session_id = ? AND payment_status = 'paid';

-- name: UpdateSessionTotals :exec
UPDATE pos_sessions
SET total_sales = ?, total_cash_sales = ?, total_card_sales = ?,
    total_online_sales = ?, total_orders = ?
WHERE id = ?;

-- name: GetSessionByID :one
SELECT id, owner_id, business_id, location_id, staff_id, status,
       opening_cash, closing_cash, expected_cash, cash_difference,
       total_sales, total_orders, total_cash_sales, total_card_sales,
       total_online_sales, total_refunds, opened_at, closed_at, notes,
       terminal_id
FROM pos_sessions
WHERE business_id = ? AND id = ?;

-- name: UpdateSessionRefundTotals :exec
UPDATE pos_sessions SET total_refunds = total_refunds + ? WHERE id = ?;

-- name: GetOrder :one
SELECT id, owner_id, business_id, session_id, table_id, location_id,
       order_number, type, status, customer_id, customer_name,
       staff_id, subtotal, discount_amount, tax_amount, total,
       payment_method, payment_status, amount_paid, change_due,
       notes, kitchen_notes, created_at, updated_at, paid_at,
       terminal_id, branch_id, tip_amount, void_reason, voided_by, voided_at, exchange_order_id
FROM pos_orders
WHERE business_id = ? AND id = ?;

-- name: CreateOrder :execresult
INSERT INTO pos_orders
(owner_id, business_id, session_id, table_id, location_id, order_number, type, status,
 customer_id, customer_name, staff_id, subtotal, tax_amount, total, payment_method, notes, kitchen_notes)
VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: CreateOrderItem :execresult
INSERT INTO pos_order_items
(order_id, menu_item_id, name, sku, quantity, unit_price,
 discount_pct, discount_amount, tax_rate, total, notes, modifiers, combo_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateOrderStatus :execresult
UPDATE pos_orders SET status = ? WHERE business_id = ? AND id = ?;

-- name: UpdateOrderItemStatus :execresult
UPDATE pos_order_items SET status = ? WHERE order_id = ? AND id = ?;

-- name: GetOrderItems :many
SELECT id, order_id, menu_item_id, name, sku,
       quantity, unit_price, discount_pct, discount_amount, tax_rate, total,
       notes, status, created_at, COALESCE(modifiers, JSON_ARRAY()) AS modifiers, combo_id, variant_name
FROM pos_order_items
WHERE order_id = ?;

-- name: MarkOrderPaid :execresult
UPDATE pos_orders
SET status = 'paid', payment_status = 'paid', payment_method = ?,
    amount_paid = ?, change_due = ?, paid_at = ?
WHERE business_id = ? AND id = ? AND payment_status <> 'paid';

-- name: GetOrderForPayment :one
SELECT total, table_id, type, payment_status FROM pos_orders WHERE business_id = ? AND id = ? FOR UPDATE;

-- The refundable cap is GREATEST(total, amount_paid - change_due): the net cash
-- this order actually left with the merchant. amount_paid alone would include
-- change already handed back, and total alone would refuse a tip that was
-- genuinely retained. The locked read carries all three so the cap is decided
-- on the row the caller is about to write.
-- name: GetOrderForRefund :one
SELECT total, amount_paid, change_due, payment_status FROM pos_orders WHERE business_id = ? AND id = ? FOR UPDATE;

-- name: SumRefundsForOrder :one
SELECT COALESCE(SUM(amount), 0) AS refunded FROM pos_refunds WHERE business_id = ? AND order_id = ?;

-- name: CreateRefund :execresult
INSERT INTO pos_refunds (owner_id, business_id, order_id, amount, reason, refunded_by)
VALUES (?, ?, ?, ?, ?, ?);

-- Retires the sale: only reached once the cumulative refunds have met the
-- refundable cap the caller computed from the locked row. A refund below the
-- cap must NOT call this; it leaves the order at payment_status 'paid' so the
-- goods the customer took stay in GetSessionTotals and the dashboard.
--
-- Note the enum's 'partial' value is deliberately not used for a part-returned
-- order. Every revenue query keys on payment_status = 'paid'
-- (GetSessionTotals below, plus pos_day_summary, pos_receipts and insights),
-- so writing 'partial' would drop the sale from those reports exactly as
-- 'refunded' does. The partial refund is recorded in pos_refunds and rolled
-- into the session's total_refunds instead.
-- name: MarkOrderRefunded :exec
UPDATE pos_orders SET status = 'refunded', payment_status = 'refunded'
WHERE business_id = ? AND id = ?;

-- name: GetDashboardSalesToday :one
SELECT
    COALESCE(SUM(total), 0) AS total_sales,
    COUNT(*) AS order_count,
    COALESCE(AVG(total), 0) AS avg_order
FROM pos_orders
WHERE business_id = ?
  AND payment_status = 'paid'
  AND paid_at >= CURDATE()
  AND paid_at < DATE_ADD(CURDATE(), INTERVAL 1 DAY);

-- name: GetTableTurnover :one
SELECT COUNT(*) AS table_count,
       COALESCE(SUM(CASE WHEN status != 'available' THEN 1 ELSE 0 END), 0) AS turned_tables
FROM pos_tables WHERE business_id = ?;

-- name: GetOrderSessionID :one
SELECT session_id FROM pos_orders WHERE id = ?;

-- name: CountTodaysOrders :one
SELECT COUNT(*) AS cnt FROM pos_orders
WHERE business_id = ?
  AND created_at >= CURDATE()
  AND created_at < DATE_ADD(CURDATE(), INTERVAL 1 DAY);

-- name: MoveOrderItems :exec
UPDATE pos_order_items SET order_id = ? WHERE id = ? AND order_id = ?;

-- name: GetOrderItemsTotal :one
SELECT COALESCE(SUM(total), 0) AS subtotal
FROM pos_order_items
WHERE order_id = ?;

-- name: UpdateOrderTotals :exec
UPDATE pos_orders
SET subtotal = sqlc.arg(subtotal), total = sqlc.arg(total)
WHERE id = ?;

-- name: CancelOrder :exec
UPDATE pos_orders SET status = 'cancelled' WHERE id = ? AND business_id = ?;

-- name: GetOrderSubtotal :one
SELECT o.id,
       COALESCE(o.discount_amount, 0) AS existing_discount,
       COALESCE(SUM((i.unit_price * i.quantity) - COALESCE(i.discount_amount, 0)), 0) AS subtotal,
       COALESCE(SUM(((i.unit_price * i.quantity) - COALESCE(i.discount_amount, 0)) * COALESCE(i.tax_rate, 0) / 100), 0) AS tax_amount
FROM pos_orders o
LEFT JOIN pos_order_items i ON i.order_id = o.id AND i.status != 'cancelled'
WHERE o.business_id = ? AND o.id = ?
GROUP BY o.id;

-- name: GetOrderForDiscount :one
SELECT payment_status
FROM pos_orders
WHERE business_id = ? AND id = ?
FOR UPDATE;

-- name: ApplyDiscount :exec
UPDATE pos_orders SET discount_amount = ?, total = ?, notes = CONCAT(COALESCE(notes,''), ?)
WHERE id = ? AND business_id = ?;

-- name: VoidOrder :execresult
UPDATE pos_orders SET status = 'voided', notes = CONCAT(COALESCE(notes,''), ?)
WHERE id = ? AND business_id = ?;

-- name: HoldOrder :execresult
UPDATE pos_orders SET status = 'held', notes = CONCAT(COALESCE(notes,''), ?)
WHERE id = ? AND business_id = ?;

-- name: SetOrderPaddleTransactionID :execresult
UPDATE pos_orders SET paddle_transaction_id = ?
WHERE id = ? AND business_id = ?;

-- name: MarkOrderPaidByPaddleTransaction :execresult
UPDATE pos_orders SET status = 'paid', payment_status = 'paid', paid_at = NOW()
WHERE paddle_transaction_id = ? AND payment_status <> 'paid';

-- name: GetOrderIDByPaddleTransaction :one
SELECT id, owner_id FROM pos_orders WHERE paddle_transaction_id = ? LIMIT 1;

-- name: RecordPaddleEventProcessed :execresult
INSERT INTO pos_paddle_events_processed (transaction_id, event_type, owner_id)
VALUES (?, ?, ?);

-- name: IsPaddleEventProcessed :one
SELECT COUNT(*) FROM pos_paddle_events_processed WHERE transaction_id = ?;

-- name: VoidOrderItem :execresult
UPDATE pos_order_items SET status = 'cancelled' WHERE id = ? AND order_id = ? AND status IN ('pending','preparing');

-- name: GetActiveOrderItemsTotal :one
SELECT COALESCE(SUM((unit_price * quantity) - COALESCE(discount_amount, 0)), 0) AS subtotal,
       COALESCE(SUM(((unit_price * quantity) - COALESCE(discount_amount, 0)) * COALESCE(tax_rate, 0) / 100), 0) AS tax_amount
FROM pos_order_items
WHERE order_id = ?
  AND status != 'cancelled';

-- name: UpdateOrderTotal :exec
UPDATE pos_orders
SET subtotal = sqlc.arg(subtotal),
    tax_amount = sqlc.arg(tax_amount),
    total = GREATEST(sqlc.arg(subtotal) + sqlc.arg(tax_amount) - discount_amount, 0)
WHERE id = ? AND business_id = ?;

-- name: IsOnlineOrderingEnabled :one
-- cross-schema read from business_addons (addons module) — gate online ordering on the business add-on entitlement.
SELECT COUNT(*) FROM business_addons WHERE business_id = ? AND addon_id = 'online_ordering' AND is_enabled = TRUE;

-- name: GetPublicOrderNumberStatus :one
SELECT order_number, status FROM pos_orders WHERE business_id = ? AND id = ? LIMIT 1;
