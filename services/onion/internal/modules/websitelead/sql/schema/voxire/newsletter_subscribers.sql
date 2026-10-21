CREATE TABLE IF NOT EXISTS `voxire_newsletter_subscribers` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `email` varchar(255) NOT NULL,
  `source` varchar(50) NOT NULL,
  `post_slug` varchar(255) DEFAULT NULL,
  `post_title` varchar(255) DEFAULT NULL,
  `post_category` varchar(100) DEFAULT NULL,
  `language` varchar(5) NOT NULL DEFAULT 'en',
  `resend_contact_id` varchar(64) DEFAULT NULL,
  `resend_synced_at` timestamp NULL DEFAULT NULL,
  `resend_sync_attempts` int NOT NULL DEFAULT '0',
  `resend_sync_error` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_newsletter_email` (`email`),
  KEY `idx_newsletter_resend_pending` (`resend_synced_at`,`resend_sync_attempts`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
