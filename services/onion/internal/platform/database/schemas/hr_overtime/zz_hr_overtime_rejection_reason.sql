-- Preserve the explanation supplied with a rejected overtime decision.
-- The base schema includes the column for fresh databases; this idempotent
-- migration upgrades installations that already have hr_overtime_requests.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_overtime_requests' AND COLUMN_NAME = 'rejection_reason') = 0,
  'ALTER TABLE hr_overtime_requests ADD COLUMN `rejection_reason` TEXT NULL AFTER current_approval_level',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
