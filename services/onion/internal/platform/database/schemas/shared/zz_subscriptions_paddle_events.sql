
  

CREATE TABLE IF NOT EXISTS `paddle_webhook_events` (
  `event_id` varchar(255) NOT NULL,
  `paddle_subscription_id` varchar(255) NOT NULL,
  `occurred_at` timestamp NOT NULL,
  `received_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`event_id`),
  KEY `idx_paddle_events_subscription_occurred` (`paddle_subscription_id`,`occurred_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
