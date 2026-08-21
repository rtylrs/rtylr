-- Hot-path tenant indexes for notification, API-key, and webhook reads.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hub_notifications' AND INDEX_NAME = 'idx_hub_notifications_business_user_created') = 0,
  'ALTER TABLE hub_notifications ADD INDEX idx_hub_notifications_business_user_created (business_id, user_id, created_at, id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hub_api_keys' AND INDEX_NAME = 'idx_hub_api_keys_business') = 0,
  'ALTER TABLE hub_api_keys ADD INDEX idx_hub_api_keys_business (business_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hub_webhooks' AND INDEX_NAME = 'idx_hub_webhooks_business') = 0,
  'ALTER TABLE hub_webhooks ADD INDEX idx_hub_webhooks_business (business_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
