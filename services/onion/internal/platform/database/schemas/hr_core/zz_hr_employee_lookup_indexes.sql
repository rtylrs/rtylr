-- Tenant-leading identity lookup indexes for staff/employee synchronization.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_employees' AND INDEX_NAME = 'idx_hr_employees_business_email') = 0,
  'ALTER TABLE hr_employees ADD INDEX idx_hr_employees_business_email (business_id, email)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
