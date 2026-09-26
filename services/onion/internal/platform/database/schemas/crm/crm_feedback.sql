CREATE TABLE IF NOT EXISTS `crm_feedback` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `customer_id` bigint NOT NULL,
  `type` enum('nps','review') COLLATE utf8mb4_unicode_ci NOT NULL,
  `score` int DEFAULT '0',
  `comment` text COLLATE utf8mb4_unicode_ci,
  `source` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_crm_feedback_business` (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
