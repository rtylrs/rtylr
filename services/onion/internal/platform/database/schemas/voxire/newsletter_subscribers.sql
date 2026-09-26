CREATE TABLE IF NOT EXISTS `voxire_newsletter_subscribers` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `email` varchar(255) NOT NULL,
  `source` varchar(50) NOT NULL,
  `post_slug` varchar(255) DEFAULT NULL,
  `post_title` varchar(255) DEFAULT NULL,
  `post_category` varchar(100) DEFAULT NULL,
  `language` varchar(5) NOT NULL DEFAULT 'en',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_newsletter_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
