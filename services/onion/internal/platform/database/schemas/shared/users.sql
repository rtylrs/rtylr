CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint DEFAULT NULL,
  `email` varchar(320) COLLATE utf8mb4_unicode_ci NOT NULL,
  `username` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `password_hash` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL,
  `full_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `signup_ip_address` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `signup_device_info` text COLLATE utf8mb4_unicode_ci,
  `onboarded` tinyint(1) DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `is_console_admin` tinyint(1) NOT NULL DEFAULT '0',
  `is_banned` tinyint(1) NOT NULL DEFAULT '0',
  `ban_reason` text COLLATE utf8mb4_unicode_ci,
  `banned_at` timestamp NULL DEFAULT NULL,
  `auth_provider` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'password',
  `token_epoch` bigint NOT NULL DEFAULT '0',
  `two_factor_enabled` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `username` (`username`),
  KEY `idx_users_business` (`business_id`),
  KEY `idx_users_provider` (`auth_provider`),
  KEY `idx_users_console_admin` (`is_console_admin`),
  KEY `idx_users_banned` (`is_banned`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



CREATE TABLE IF NOT EXISTS `manual_access_grants` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `plan_id` bigint NOT NULL,
  `granted_by` bigint NOT NULL COMMENT 'console admin user_id who created the grant',
  `note` text COLLATE utf8mb4_unicode_ci COMMENT 'deal notes, e.g. "cash payment $500/yr agreed 2026-05-14"',
  `expires_at` timestamp NULL DEFAULT NULL COMMENT 'NULL = never expires',
  `revoked_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_mag_user` (`user_id`),
  KEY `idx_mag_plan` (`plan_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `console_stats` (
  `id` tinyint NOT NULL,
  `total_users` bigint NOT NULL DEFAULT '0',
  `banned_users` bigint NOT NULL DEFAULT '0',
  `total_businesses` bigint NOT NULL DEFAULT '0',
  `active_subs` bigint NOT NULL DEFAULT '0',
  `trial_subs` bigint NOT NULL DEFAULT '0',
  `manual_grants` bigint NOT NULL DEFAULT '0',
  `manual_addon_grants` bigint NOT NULL DEFAULT '0',
  `console_admins` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  CONSTRAINT `console_stats_chk_1` CHECK ((`id` = 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
