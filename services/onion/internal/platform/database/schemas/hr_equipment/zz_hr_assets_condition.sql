-- Keep the physical condition independent from the asset lifecycle status.
-- The base schema includes the column for fresh databases; this idempotent
-- migration upgrades installations that already have hr_assets.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_assets' AND COLUMN_NAME = 'condition') = 0,
  'ALTER TABLE hr_assets ADD COLUMN `condition` VARCHAR(20) NOT NULL DEFAULT ''good'' AFTER status',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
