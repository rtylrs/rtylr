CREATE TABLE IF NOT EXISTS `user_business_access` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `business_id` bigint NOT NULL,
  `role` enum('owner','admin','viewer') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'owner',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_uba_user_business` (`user_id`,`business_id`),
  KEY `idx_uba_user` (`user_id`),
  KEY `idx_uba_business` (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
