CREATE TABLE IF NOT EXISTS `ai_usage_monthly` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL COMMENT 'Canonical tenant = businesses.owner_id',
  `user_id` bigint NOT NULL DEFAULT '0' COMMENT '0 = shared team pool; >0 = per-person bucket',
  `billing_month` date NOT NULL COMMENT 'First day of the calendar month (UTC)',
  `tokens_used` bigint NOT NULL DEFAULT '0' COMMENT 'Combined prompt + completion tokens',
  `requests` int NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_ai_usage_owner_user_month` (`owner_id`,`user_id`,`billing_month`),
  KEY `idx_ai_usage_owner` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ai_quota_settings` (
  `owner_id` bigint NOT NULL COMMENT 'Canonical tenant = businesses.owner_id',
  `mode` enum('pool','per_person') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pool',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
