CREATE TABLE IF NOT EXISTS `webhook_deliveries` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `webhook_id` bigint NOT NULL,
  `event_id` bigint unsigned NOT NULL,
  `status` enum('pending','delivering','delivered','failed','dead') NOT NULL DEFAULT 'pending',
  `attempts` tinyint unsigned NOT NULL DEFAULT '0',
  `max_attempts` tinyint unsigned NOT NULL DEFAULT '5',
  `next_attempt_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `response_status` smallint DEFAULT NULL,
  `last_error` text,
  `delivered_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_webhook_deliveries_target` (`webhook_id`,`event_id`),
  KEY `idx_webhook_deliveries_due` (`status`,`next_attempt_at`,`id`),
  KEY `idx_webhook_deliveries_business` (`business_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
