-- Ensure hr_departments has the expected business-scoped columns and align
-- owner_id with business_id on existing rows. Safe to re-run.

SET @sql := IF(
    EXISTS (
        SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_departments' AND COLUMN_NAME = 'location_id'
    ),
    'SELECT 1',
    'ALTER TABLE hr_departments ADD COLUMN location_id BIGINT DEFAULT NULL'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    EXISTS (
        SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_departments' AND COLUMN_NAME = 'parent_department_id'
    ),
    'SELECT 1',
    'ALTER TABLE hr_departments ADD COLUMN parent_department_id BIGINT DEFAULT NULL'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    EXISTS (
        SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_departments' AND COLUMN_NAME = 'default_work_schedule_id'
    ),
    'SELECT 1',
    'ALTER TABLE hr_departments ADD COLUMN default_work_schedule_id BIGINT DEFAULT NULL'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE hr_departments d
JOIN businesses b ON b.owner_id = d.owner_id
SET d.business_id = b.id
WHERE d.business_id = 0;

UPDATE hr_departments
SET owner_id = business_id
WHERE business_id > 0 AND owner_id <> business_id;
