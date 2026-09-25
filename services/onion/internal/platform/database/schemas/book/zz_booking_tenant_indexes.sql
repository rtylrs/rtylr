-- Business-leading schedule indexes for appointment lists, date windows, and
-- status counts. The migration is idempotent for already-provisioned tenants.
-- Roll back only after checking live plans: DROP INDEX <name> ON
-- booking_appointments. Existing owner-leading indexes remain unchanged.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'booking_appointments' AND INDEX_NAME = 'idx_booking_appts_business_start') = 0,
  'ALTER TABLE booking_appointments ADD INDEX idx_booking_appts_business_start (business_id, start_time, id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'booking_appointments' AND INDEX_NAME = 'idx_booking_appts_business_status_start') = 0,
  'ALTER TABLE booking_appointments ADD INDEX idx_booking_appts_business_status_start (business_id, status, start_time)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
