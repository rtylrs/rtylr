CREATE TABLE IF NOT EXISTS `activity_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `app` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'hub' COMMENT 'Which app: hub, pos, erp, hr, crm, finance, flow',
  `entity_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `entity_id` bigint DEFAULT NULL,
  `action` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `user_id` bigint DEFAULT NULL,
  `old_values` json DEFAULT NULL,
  `new_values` json DEFAULT NULL,
  `ip_address` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_activity_owner` (`owner_id`),
  KEY `idx_activity_app` (`owner_id`,`app`),
  KEY `idx_activity_created` (`owner_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
