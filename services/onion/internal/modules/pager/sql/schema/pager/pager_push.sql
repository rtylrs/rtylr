-- Web Push subscriptions of the phones following an open ticket. A row holds
-- what the browser's push service needs and nothing about the person: no
-- name, phone or IP. Rows go when the ticket is collected, when the push
-- service reports them gone, and once older than a ticket's lifetime.
CREATE TABLE IF NOT EXISTS `pager_push_subscriptions` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `ticket_id` bigint NOT NULL,
  `endpoint` varchar(1024) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `endpoint_hash` binary(32) NOT NULL,
  `p256dh` varchar(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `auth` varchar(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `created_at` timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pager_push_ticket_endpoint` (`ticket_id`,`endpoint_hash`),
  KEY `idx_pager_push_business_ticket` (`business_id`,`ticket_id`),
  KEY `idx_pager_push_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
