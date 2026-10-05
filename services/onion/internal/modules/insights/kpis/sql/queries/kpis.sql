-- cross-schema read from hr_staff — active staff headcount for the KPI summary
-- name: CountActiveStaff :one
SELECT COUNT(*) FROM hr_staff WHERE business_id = ? AND status = 'active';

-- cross-schema read from crm_customers — total customers for the KPI summary
-- name: CountCustomers :one
SELECT COUNT(*) FROM crm_customers WHERE business_id = ?;

-- cross-schema read from products — intentional KPI projection; count low-stock catalog products without importing product repository internals.
-- name: CountLowStockProducts :one
SELECT COUNT(*) FROM erp_products
WHERE business_id = ? AND is_active = TRUE AND track_inventory = TRUE
  AND low_stock_threshold IS NOT NULL
  AND stock_quantity <= low_stock_threshold;

-- cross-schema read from pos_orders — orders created today for the KPI summary
-- name: CountOrdersToday :one
SELECT COUNT(*) FROM pos_orders
WHERE business_id = ?
  AND created_at >= CURDATE()
  AND created_at < CURDATE() + INTERVAL 1 DAY;

-- cross-schema read from pos_orders — total revenue from orders created today
-- name: SumRevenueToday :one
SELECT COALESCE(SUM(total), 0) FROM pos_orders
WHERE business_id = ?
  AND created_at >= CURDATE()
  AND created_at < CURDATE() + INTERVAL 1 DAY;
