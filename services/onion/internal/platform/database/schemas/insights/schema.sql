CREATE TABLE IF NOT EXISTS `insights_settings` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `default_currency` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'USD',
  `fiscal_year_start` varchar(5) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '01-01',
  `timezone` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'UTC',
  `default_period` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'this_month',
  `compare_previous` tinyint(1) NOT NULL DEFAULT '0',
  `dashboard_widgets` json NOT NULL DEFAULT (json_array()),
  `favorite_reports` json NOT NULL DEFAULT (json_array()),
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_insights_settings_owner` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `insights_scheduled_reports` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `template_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `cadence` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `channel` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `recipient` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_active` tinyint(1) NOT NULL DEFAULT '1',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_scheduled_reports_owner` (`owner_id`),
  KEY `idx_scheduled_reports_active` (`owner_id`,`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
