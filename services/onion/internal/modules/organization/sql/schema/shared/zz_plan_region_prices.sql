-- Per-region price overrides for a plan. A plan with no row for a visitor's
-- region falls back to the plan's own (USD) price. Prices are in the minor unit
-- of `currency` (cents). paddle_price_id is the Paddle price that checkout must
-- use for this region, so the number on the page is the number charged.
CREATE TABLE IF NOT EXISTS `plan_region_prices` (
  `plan_id` bigint NOT NULL,
  `region` char(2) CHARACTER SET ascii COLLATE ascii_general_ci NOT NULL COMMENT 'ISO 3166-1 alpha-2, upper case',
  `currency` char(3) CHARACTER SET ascii COLLATE ascii_general_ci NOT NULL COMMENT 'ISO 4217, upper case',
  `price_per_seat_cents` int NOT NULL DEFAULT '0',
  `annual_price_per_seat_cents` int NOT NULL DEFAULT '0',
  `paddle_price_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `is_active` tinyint(1) NOT NULL DEFAULT '1',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`plan_id`,`region`),
  KEY `idx_plan_region_prices_region` (`region`,`is_active`),
  CONSTRAINT `fk_plan_region_prices_plan` FOREIGN KEY (`plan_id`) REFERENCES `plans` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
