CREATE TABLE IF NOT EXISTS `crm_loyalty_tiers` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `min_points` int DEFAULT '0',
  `discount_pct` decimal(5,2) DEFAULT '0.00',
  `perks` json DEFAULT NULL,
  `sort_order` int DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_crm_loyalty_tiers_business` (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
