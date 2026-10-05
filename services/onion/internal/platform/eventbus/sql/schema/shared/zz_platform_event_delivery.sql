CREATE TABLE IF NOT EXISTS `platform_events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `event_type` varchar(64) NOT NULL,
  `subject_type` varchar(64) NOT NULL,
  `subject_id` bigint NOT NULL,
  `payload` json NOT NULL,
  `dedupe_key` varchar(191) NOT NULL,
  `occurred_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `dispatched_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_platform_events_dedupe` (`business_id`,`dedupe_key`),
  KEY `idx_platform_events_pending` (`dispatched_at`,`id`),
  KEY `idx_platform_events_business_type` (`business_id`,`event_type`,`occurred_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
