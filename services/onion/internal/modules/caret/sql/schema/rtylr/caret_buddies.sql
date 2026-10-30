CREATE TABLE IF NOT EXISTS `caret_buddies` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `run_as_user_id` bigint NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT '0',
  `instructions` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `interval_minutes` int NOT NULL DEFAULT '60',
  `daily_run_cap` int NOT NULL DEFAULT '24',
  `next_run_at` timestamp NULL DEFAULT NULL,
  `last_run_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_caret_buddies_business` (`business_id`),
  KEY `idx_caret_buddies_due` (`enabled`,`next_run_at`),
  CONSTRAINT `caret_buddies_interval` CHECK ((`interval_minutes` between 15 and 1440)),
  CONSTRAINT `caret_buddies_cap` CHECK ((`daily_run_cap` between 1 and 96))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
