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

CREATE TABLE IF NOT EXISTS `automation_runs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `automation_id` bigint NOT NULL,
  `event_id` bigint unsigned NOT NULL,
  `status` enum('pending','running','succeeded','failed','dead') NOT NULL DEFAULT 'pending',
  `attempts` tinyint unsigned NOT NULL DEFAULT '0',
  `max_attempts` tinyint unsigned NOT NULL DEFAULT '3',
  `next_attempt_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `last_error` text,
  `completed_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_automation_runs_target` (`automation_id`,`event_id`),
  KEY `idx_automation_runs_due` (`status`,`next_attempt_at`,`id`),
  KEY `idx_automation_runs_business` (`business_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
