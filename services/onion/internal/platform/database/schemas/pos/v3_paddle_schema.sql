CREATE TABLE IF NOT EXISTS `pos_paddle_events_processed` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `transaction_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `event_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `owner_id` bigint NOT NULL,
  `processed_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_pos_paddle_events_tx` (`transaction_id`),
  KEY `idx_pos_paddle_events_owner` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
