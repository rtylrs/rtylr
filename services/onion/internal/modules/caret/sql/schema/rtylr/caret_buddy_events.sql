CREATE TABLE IF NOT EXISTS `caret_buddy_events` (
  `event_id` bigint unsigned NOT NULL,
  `buddy_id` bigint NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`event_id`),
  KEY `idx_caret_buddy_events_buddy` (`buddy_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
