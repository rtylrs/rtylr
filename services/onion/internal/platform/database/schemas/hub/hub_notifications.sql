CREATE TABLE IF NOT EXISTS `hub_notifications` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `user_id` bigint NOT NULL,
  `title` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `body` text COLLATE utf8mb4_unicode_ci,
  `type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'info',
  `is_read` tinyint(1) NOT NULL DEFAULT '0',
  `link` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_hub_notifications_owner_id` (`owner_id`),
  KEY `idx_hub_notifications_user_id` (`user_id`),
  KEY `idx_hub_notifications_is_read` (`is_read`),
  KEY `idx_hub_notifications_business_user_created` (`business_id`,`user_id`,`created_at`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;




