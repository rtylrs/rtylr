-- strict is no longer read: it belonged to the typed-number join that the
-- scan flow replaced. It stays so the migration is additive.
CREATE TABLE IF NOT EXISTS `pager_settings` (
  `business_id` bigint NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT '0',
  `strict` tinyint(1) NOT NULL DEFAULT '0',
  `label_mode` varchar(8) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT 'either',
  `review_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `prep_minutes` smallint NOT NULL DEFAULT '0',
  `customer_numbers` tinyint(1) NOT NULL DEFAULT '0',
  `ringtone` varchar(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT 'bell',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- A ticket starts when a phone scans the venue's QR: waiting, unlabelled and
-- holding a match_number such as "427". The cashier then labels it with the
-- order number or the customer's name. A ticket is open while it is waiting or
-- ready. match_word ("panda") and match_code ("A7") are what tickets were given
-- before the numbers; only tickets from then have one.
--
-- number holds the label of a number-kind ticket only, so open_number's unique
-- key allows one open ticket per business and number while names repeat
-- freely. open_match_number holds the match number only while the ticket is
-- open and unlabelled, so it is unique among the phones waiting to be labelled
-- and is free again once the cashier labels the ticket; open_match_word and
-- open_match_code do the same for the old codes. The row carries no customer
-- data beyond the label the cashier typed: no phone or address.
CREATE TABLE IF NOT EXISTS `pager_tickets` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `number` varchar(12) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `label` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `label_kind` varchar(8) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL,
  `match_code` char(2) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL,
  `match_word` varchar(16) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL,
  `match_number` varchar(4) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL,
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'waiting',
  `source` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `public_token` varchar(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `phone_linked` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `ready_at` timestamp(3) NULL DEFAULT NULL,
  `reminded_at` timestamp(3) NULL DEFAULT NULL,
  `acknowledged_at` timestamp(3) NULL DEFAULT NULL,
  `expected_ready_at` timestamp(3) NULL DEFAULT NULL,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `open_number` varchar(12) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci GENERATED ALWAYS AS (CASE WHEN `status` IN ('waiting','ready') THEN `number` ELSE NULL END) VIRTUAL,
  `open_match_code` char(2) CHARACTER SET ascii COLLATE ascii_bin GENERATED ALWAYS AS (CASE WHEN `status` IN ('waiting','ready') AND `label_kind` IS NULL THEN `match_code` ELSE NULL END) VIRTUAL,
  `open_match_word` varchar(16) CHARACTER SET ascii COLLATE ascii_bin GENERATED ALWAYS AS (CASE WHEN `status` IN ('waiting','ready') AND `label_kind` IS NULL THEN `match_word` ELSE NULL END) VIRTUAL,
  `open_match_number` varchar(4) CHARACTER SET ascii COLLATE ascii_bin GENERATED ALWAYS AS (CASE WHEN `status` IN ('waiting','ready') AND `label_kind` IS NULL THEN `match_number` ELSE NULL END) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pager_tickets_public_token` (`public_token`),
  UNIQUE KEY `uk_pager_tickets_open_number` (`business_id`,`open_number`),
  UNIQUE KEY `uk_pager_tickets_open_match_code` (`business_id`,`open_match_code`),
  UNIQUE KEY `uk_pager_tickets_open_match_word` (`business_id`,`open_match_word`),
  UNIQUE KEY `uk_pager_tickets_open_match_number` (`business_id`,`open_match_number`),
  KEY `idx_pager_tickets_business_status` (`business_id`,`status`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
