SET @sql := IF(
    (SELECT COUNT(*)
     FROM INFORMATION_SCHEMA.STATISTICS
     WHERE TABLE_SCHEMA = DATABASE()
       AND TABLE_NAME = 'voxire_career_applications'
       AND INDEX_NAME = 'idx_career_apps_job_status'),
    'SELECT 1',
    'ALTER TABLE voxire_career_applications ADD INDEX idx_career_apps_job_status (job_id, status)'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
