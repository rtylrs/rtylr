CREATE TABLE IF NOT EXISTS `user_groups` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `color` varchar(7) COLLATE utf8mb4_unicode_ci DEFAULT '#6366f1',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_owner_name` (`owner_id`,`name`),
  KEY `idx_user_groups_owner` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_group_resource_permissions` (
  `group_id` bigint NOT NULL,
  `resource_key` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `permission_level` enum('viewer','editor') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'viewer',
  PRIMARY KEY (`group_id`,`resource_key`),
  KEY `idx_ugrp_group` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_group_members` (
  `group_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `added_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `added_by` bigint DEFAULT NULL,
  PRIMARY KEY (`group_id`,`user_id`),
  KEY `idx_ugm_user` (`user_id`,`owner_id`),
  KEY `idx_ugm_group` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_group_app_permissions` (
  `group_id` bigint NOT NULL,
  `app_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `permission_level` enum('use','edit','own') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'use',
  PRIMARY KEY (`group_id`,`app_id`),
  KEY `idx_ugap_group` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
