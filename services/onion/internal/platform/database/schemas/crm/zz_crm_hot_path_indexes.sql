-- Business-leading ordering indexes for CRM template and segment lists.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'crm_templates' AND INDEX_NAME = 'idx_crm_templates_business_created') = 0,
  'ALTER TABLE crm_templates ADD INDEX idx_crm_templates_business_created (business_id, created_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'crm_segments' AND INDEX_NAME = 'idx_crm_segments_business_created') = 0,
  'ALTER TABLE crm_segments ADD INDEX idx_crm_segments_business_created (business_id, created_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
