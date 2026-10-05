CREATE TABLE IF NOT EXISTS `webhook_dead_delivery_notices` (
  `business_id` bigint NOT NULL,
  `webhook_id` bigint NOT NULL,
  `notice_date` date NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`business_id`,`webhook_id`,`notice_date`),
  KEY `idx_webhook_dead_delivery_notices_date` (`notice_date`,`webhook_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
