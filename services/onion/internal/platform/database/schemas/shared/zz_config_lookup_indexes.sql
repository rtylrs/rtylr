-- Config is cached for normal reads, but cache misses and cold starts still
-- need indexed tenant lookups. Guard each ALTER through INFORMATION_SCHEMA.

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'config' AND INDEX_NAME = 'idx_config_business') = 0,
  'ALTER TABLE config ADD INDEX idx_config_business (business_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'config' AND INDEX_NAME = 'idx_config_owner') = 0,
  'ALTER TABLE config ADD INDEX idx_config_owner (owner_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
