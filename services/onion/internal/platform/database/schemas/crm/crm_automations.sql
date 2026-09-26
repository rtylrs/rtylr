CREATE TABLE IF NOT EXISTS `crm_automations` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `trigger_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `trigger_config` json DEFAULT NULL,
  `actions` json DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT '1',
  `last_run_at` timestamp NULL DEFAULT NULL,
  `run_count` int DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_crm_automations_business` (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
