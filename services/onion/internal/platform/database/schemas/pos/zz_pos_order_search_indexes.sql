-- Adds business-scoped pagination and indexed search support for POS order lists.
-- Duplicate-index errors are safe under the migration runner's --force mode.

ALTER TABLE pos_orders ADD INDEX idx_pos_orders_business_status_created (business_id, status, created_at);
ALTER TABLE pos_orders ADD INDEX idx_pos_orders_business_session_created (business_id, session_id, created_at);
ALTER TABLE pos_orders ADD INDEX idx_pos_orders_business_created (business_id, created_at);
ALTER TABLE pos_orders ADD FULLTEXT INDEX idx_pos_orders_search (order_number, customer_name, customer_phone);
