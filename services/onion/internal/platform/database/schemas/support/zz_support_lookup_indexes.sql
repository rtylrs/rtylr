-- Tenant-leading exact lookup indexes used by support ticket and knowledge APIs.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'support_customers' AND INDEX_NAME = 'idx_support_customers_business_email') = 0,
  'ALTER TABLE support_customers ADD INDEX idx_support_customers_business_email (business_id, email)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'support_articles' AND INDEX_NAME = 'idx_support_articles_business_slug') = 0,
  'ALTER TABLE support_articles ADD INDEX idx_support_articles_business_slug (business_id, slug)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'support_articles' AND INDEX_NAME = 'idx_support_articles_business_title_updated') = 0,
  'ALTER TABLE support_articles ADD INDEX idx_support_articles_business_title_updated (business_id, title, updated_at, id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
