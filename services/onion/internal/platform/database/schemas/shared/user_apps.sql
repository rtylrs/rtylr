CREATE TABLE IF NOT EXISTS `user_app_permissions` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint DEFAULT NULL,
  `owner_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `apps` json NOT NULL COMMENT 'Array of app IDs: pos, erp, hr, crm, finance, flow',
  `role` enum('owner','admin','operator','billing','menu_manager','store_manager','staff','viewer') COLLATE utf8mb4_unicode_ci DEFAULT 'staff',
  `is_active` tinyint(1) DEFAULT '1',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_user_perm` (`owner_id`,`user_id`),
  KEY `idx_perms_business` (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
