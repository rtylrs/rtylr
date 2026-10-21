-- name: GetErpSettings :one
SELECT id, business_id, currency, tax_rate, low_stock_threshold, default_warehouse_id, notify_on_low_stock, notify_on_expiry, created_at, updated_at
FROM erp_business_settings
WHERE business_id = ?
LIMIT 1;

-- name: UpsertErpSettings :exec
INSERT INTO erp_business_settings
  (business_id, currency, tax_rate, low_stock_threshold, default_warehouse_id, notify_on_low_stock, notify_on_expiry)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  currency = VALUES(currency),
  tax_rate = VALUES(tax_rate),
  low_stock_threshold = VALUES(low_stock_threshold),
  default_warehouse_id = VALUES(default_warehouse_id),
  notify_on_low_stock = VALUES(notify_on_low_stock),
  notify_on_expiry = VALUES(notify_on_expiry),
  updated_at = CURRENT_TIMESTAMP;
