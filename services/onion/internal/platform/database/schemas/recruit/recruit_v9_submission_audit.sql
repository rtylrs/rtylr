CREATE TABLE IF NOT EXISTS `careers_submission_events` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `submission_id` bigint NOT NULL,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `actor_id` bigint NOT NULL DEFAULT '0',
  `event_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `from_status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `to_status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `metadata_json` json DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_submission_events_submission` (`submission_id`,`created_at`,`id`),
  KEY `idx_submission_events_business` (`business_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
