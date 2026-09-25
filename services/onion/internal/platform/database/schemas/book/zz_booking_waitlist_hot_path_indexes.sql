-- Business-leading ordering index for waitlist history reads.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'booking_waitlist' AND INDEX_NAME = 'idx_booking_waitlist_business_created') = 0,
  'ALTER TABLE booking_waitlist ADD INDEX idx_booking_waitlist_business_created (business_id, created_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
