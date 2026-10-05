CREATE TABLE IF NOT EXISTS `crm_customer_capture_events` (
  `event_id` BIGINT UNSIGNED NOT NULL,
  `business_id` BIGINT NOT NULL,
  `processed_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`event_id`),
  KEY `idx_crm_customer_capture_business` (`business_id`, `processed_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
