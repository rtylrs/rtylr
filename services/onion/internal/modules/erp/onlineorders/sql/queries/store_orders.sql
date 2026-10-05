-- name: ListStoreOrders :many
SELECT id, owner_id, business_id, customer_email, customer_name, customer_phone, status, subtotal, tax, shipping_cost, total, shipping_address, notes, created_at, updated_at
FROM store_orders WHERE business_id = ? ORDER BY created_at DESC, id DESC LIMIT 500;

-- name: GetStoreOrder :one
SELECT id, owner_id, business_id, customer_email, customer_name, customer_phone, status, subtotal, tax, shipping_cost, total, shipping_address, notes, created_at, updated_at
FROM store_orders WHERE business_id = ? AND id = ?;

-- name: ListStoreOrderItems :many
SELECT store_order_items.id, store_order_items.order_id, store_order_items.product_id, store_order_items.name, store_order_items.quantity, store_order_items.price, store_order_items.total, store_order_items.created_at
FROM store_order_items
INNER JOIN store_orders ON store_order_items.order_id = store_orders.id
WHERE store_orders.business_id = ? AND store_order_items.order_id = ?
ORDER BY store_order_items.id ASC
LIMIT 500;

-- name: ListStoreOrderItemsForUpdate :many
SELECT store_order_items.id, store_order_items.order_id, store_order_items.product_id, store_order_items.name, store_order_items.quantity, store_order_items.price, store_order_items.total, store_order_items.created_at
FROM store_order_items
INNER JOIN store_orders ON store_order_items.order_id = store_orders.id
WHERE store_orders.business_id = ? AND store_order_items.order_id = ?
ORDER BY store_order_items.id ASC
LIMIT 500
FOR UPDATE;

-- name: GetStoreOrderStatus :one
SELECT owner_id, status, total FROM store_orders WHERE business_id = ? AND id = ? FOR UPDATE;

-- name: CreateStoreOrder :execresult
INSERT INTO store_orders (owner_id, business_id, customer_email, customer_name, customer_phone, status, subtotal, tax, shipping_cost, total, shipping_address, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: CreateStoreOrderItem :execresult
INSERT INTO store_order_items (order_id, product_id, name, quantity, price, total)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateStoreOrder :exec
UPDATE store_orders SET customer_email = ?, customer_name = ?, customer_phone = ?, status = ?, shipping_address = ?, notes = ?
WHERE business_id = ? AND id = ?;

-- cross-schema read from erp_products — all order pricing is server-derived
-- name: GetStoreProductPricing :one
SELECT name, sale_price, tax_rate, track_inventory
FROM erp_products
WHERE owner_id = ? AND id = ? AND is_active = TRUE;

-- name: ListStoreOrdersPaginated :many
SELECT id, owner_id, business_id, customer_email, customer_name, customer_phone, status, subtotal, tax, shipping_cost, total, shipping_address, notes, created_at, updated_at
FROM store_orders WHERE business_id = ? ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountStoreOrders :one
SELECT COUNT(*) FROM store_orders WHERE business_id = ?;
