-- name: CreateCheckout :execresult
INSERT INTO store_checkouts (owner_id, business_id, session_id, customer_email, items, subtotal, tax, total)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: CompleteCheckout :execresult
UPDATE store_checkouts SET status = 'completed'
WHERE business_id = ? AND id = ? AND status <> 'completed';

-- name: GetCheckoutForComplete :one
SELECT customer_email, total, status FROM store_checkouts WHERE business_id = ? AND id = ? ORDER BY id ASC LIMIT 1;

-- cross-schema read from erp_products — checkout totals are server-derived
-- name: GetCheckoutProductPricing :one
SELECT name, sale_price, tax_rate
FROM erp_products
WHERE owner_id = ? AND id = ? AND is_active = TRUE;
