-- Adds indexed self-host license filtering support.

SET @sql := IF(
    (SELECT COUNT(*)
     FROM INFORMATION_SCHEMA.STATISTICS
     WHERE TABLE_SCHEMA = DATABASE()
       AND TABLE_NAME = 'self_host_licenses'
       AND INDEX_NAME = 'idx_self_host_licenses_search'),
    'SELECT 1',
    'ALTER TABLE self_host_licenses ADD FULLTEXT INDEX idx_self_host_licenses_search (owner_email, customer_name)'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
