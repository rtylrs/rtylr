-- name: ListStoreOrders :many
SELECT id, owner_id, customer_email, customer_name, customer_phone, status, subtotal, tax, shipping_cost, total, shipping_address, notes, created_at, updated_at
FROM store_orders WHERE owner_id = ? ORDER BY created_at DESC;

-- name: GetStoreOrder :one
SELECT id, owner_id, customer_email, customer_name, customer_phone, status, subtotal, tax, shipping_cost, total, shipping_address, notes, created_at, updated_at
FROM store_orders WHERE owner_id = ? AND id = ?;

-- name: ListStoreOrderItems :many
SELECT id, order_id, product_id, name, quantity, price, total, created_at
FROM store_order_items WHERE order_id = ? ORDER BY id ASC;

-- name: CreateStoreOrder :execresult
INSERT INTO store_orders (owner_id, customer_email, customer_name, customer_phone, status, subtotal, tax, shipping_cost, total, shipping_address, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: CreateStoreOrderItem :execresult
INSERT INTO store_order_items (order_id, product_id, name, quantity, price, total)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateStoreOrder :exec
UPDATE store_orders SET customer_email = ?, customer_name = ?, customer_phone = ?, status = ?, subtotal = ?, tax = ?, shipping_cost = ?, total = ?, shipping_address = ?, notes = ?
WHERE owner_id = ? AND id = ?;
