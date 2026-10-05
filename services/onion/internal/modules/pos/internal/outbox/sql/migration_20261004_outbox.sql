CREATE TABLE IF NOT EXISTS `pos_outbox` (
 `id` bigint unsigned NOT NULL AUTO_INCREMENT,
 `business_id` bigint NOT NULL,
 `event_type` varchar(64) NOT NULL,
 `subject_type` varchar(64) NOT NULL,
 `subject_id` bigint NOT NULL,
 `payload` json NOT NULL,
 `dedupe_key` varchar(191) NOT NULL,
 `occurred_at` timestamp NOT NULL,
 `relayed_at` timestamp NULL,
 PRIMARY KEY (`id`),
 UNIQUE KEY `uk_pos_outbox_dedupe` (`business_id`, `dedupe_key`),
 KEY `idx_pos_outbox_pending` (`relayed_at`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
