CREATE TABLE IF NOT EXISTS `pager_settings` (
  `business_id` bigint NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT '0',
  `strict` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- A ticket is open while it is waiting or ready. open_number holds the number
-- only for an open ticket, so its unique key allows one open ticket per
-- business and number next to any amount of closed history. The row carries no
-- customer data: no name, phone or address.
CREATE TABLE IF NOT EXISTS `pager_tickets` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `number` varchar(12) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'waiting',
  `source` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `public_token` varchar(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `phone_linked` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `ready_at` timestamp(3) NULL DEFAULT NULL,
  `reminded_at` timestamp(3) NULL DEFAULT NULL,
  `acknowledged_at` timestamp(3) NULL DEFAULT NULL,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `open_number` varchar(12) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci GENERATED ALWAYS AS (CASE WHEN `status` IN ('waiting','ready') THEN `number` ELSE NULL END) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pager_tickets_public_token` (`public_token`),
  UNIQUE KEY `uk_pager_tickets_open_number` (`business_id`,`open_number`),
  KEY `idx_pager_tickets_business_status` (`business_id`,`status`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
