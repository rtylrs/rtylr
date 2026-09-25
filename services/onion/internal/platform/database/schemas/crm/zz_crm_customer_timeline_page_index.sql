SET @sql := IF(
    (SELECT COUNT(*)
     FROM INFORMATION_SCHEMA.STATISTICS
     WHERE TABLE_SCHEMA = DATABASE()
       AND TABLE_NAME = 'crm_customer_timeline'
       AND INDEX_NAME = 'idx_crm_customer_timeline_page') = 0,
    'ALTER TABLE crm_customer_timeline ADD INDEX idx_crm_customer_timeline_page (business_id, customer_id, created_at, id)',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
