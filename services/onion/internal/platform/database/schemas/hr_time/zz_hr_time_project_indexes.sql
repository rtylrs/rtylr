-- Project labour-cost and date-range reads filter by business first.
SET @sql := IF((SELECT COUNT(*)
                FROM INFORMATION_SCHEMA.STATISTICS
                WHERE TABLE_SCHEMA = DATABASE()
                  AND TABLE_NAME = 'hr_time_entries'
                  AND INDEX_NAME = 'idx_hr_time_entries_business_project_date') = 0,
               'ALTER TABLE hr_time_entries ADD INDEX idx_hr_time_entries_business_project_date (business_id, project_id, date)',
               'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
