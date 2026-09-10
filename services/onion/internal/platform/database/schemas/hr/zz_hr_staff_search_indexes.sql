-- Adds indexed search support for staff lists.

SET @sql := IF(
    (SELECT COUNT(*)
     FROM INFORMATION_SCHEMA.STATISTICS
     WHERE TABLE_SCHEMA = DATABASE()
       AND TABLE_NAME = 'hr_staff'
       AND INDEX_NAME = 'idx_hr_staff_search'),
    'SELECT 1',
    'ALTER TABLE hr_staff ADD FULLTEXT INDEX idx_hr_staff_search (first_name, last_name, email, role)'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
