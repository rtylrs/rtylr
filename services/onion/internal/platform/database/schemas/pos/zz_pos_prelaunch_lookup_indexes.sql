-- Business-scoped lookup indexes for pre-launch POS load checks.
-- MySQL has no portable ADD INDEX IF NOT EXISTS, so each ALTER is guarded
-- through INFORMATION_SCHEMA and prepared SQL.

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pos_locations' AND INDEX_NAME = 'idx_pos_locations_business_active_name') = 0,
  'ALTER TABLE pos_locations ADD INDEX idx_pos_locations_business_active_name (business_id, is_active, name)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pos_tables' AND INDEX_NAME = 'idx_pos_tables_business_location_status_name') = 0,
  'ALTER TABLE pos_tables ADD INDEX idx_pos_tables_business_location_status_name (business_id, location_id, status, name)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pos_sessions' AND INDEX_NAME = 'idx_pos_sessions_business_staff_status') = 0,
  'ALTER TABLE pos_sessions ADD INDEX idx_pos_sessions_business_staff_status (business_id, staff_id, status)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pos_sessions' AND INDEX_NAME = 'idx_pos_sessions_business_location_opened') = 0,
  'ALTER TABLE pos_sessions ADD INDEX idx_pos_sessions_business_location_opened (business_id, location_id, opened_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pos_orders' AND INDEX_NAME = 'idx_pos_orders_business_order_number') = 0,
  'ALTER TABLE pos_orders ADD INDEX idx_pos_orders_business_order_number (business_id, order_number)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pos_orders' AND INDEX_NAME = 'idx_pos_orders_business_paid_at') = 0,
  'ALTER TABLE pos_orders ADD INDEX idx_pos_orders_business_paid_at (business_id, payment_status, paid_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Supports customer lifetime-value aggregation by tenant, customer, and paid status.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pos_orders' AND INDEX_NAME = 'idx_pos_orders_business_customer_payment') = 0,
  'ALTER TABLE pos_orders ADD INDEX idx_pos_orders_business_customer_payment (business_id, customer_id, payment_status)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pos_orders' AND INDEX_NAME = 'idx_pos_orders_session_payment') = 0,
  'ALTER TABLE pos_orders ADD INDEX idx_pos_orders_session_payment (session_id, payment_status)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pos_users' AND INDEX_NAME = 'idx_pos_users_business_name') = 0,
  'ALTER TABLE pos_users ADD INDEX idx_pos_users_business_name (business_id, name)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pos_users' AND INDEX_NAME = 'idx_pos_users_business_location_active_name') = 0,
  'ALTER TABLE pos_users ADD INDEX idx_pos_users_business_location_active_name (business_id, location_id, is_active, name)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pos_users' AND INDEX_NAME = 'idx_pos_users_business_location_user_active') = 0,
  'ALTER TABLE pos_users ADD INDEX idx_pos_users_business_location_user_active (business_id, location_id, user_id, is_active)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pos_member_locations' AND INDEX_NAME = 'idx_member_loc_business_user_location') = 0,
  'ALTER TABLE pos_member_locations ADD INDEX idx_member_loc_business_user_location (business_id, user_id, location_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pos_member_locations' AND INDEX_NAME = 'idx_member_loc_business_location_user') = 0,
  'ALTER TABLE pos_member_locations ADD INDEX idx_member_loc_business_location_user (business_id, location_id, user_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
