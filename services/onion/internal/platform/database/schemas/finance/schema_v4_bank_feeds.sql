CREATE TABLE IF NOT EXISTS `finance_bank_feeds` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `bank_account_id` bigint NOT NULL,
  `amount_minor_units` bigint NOT NULL,
  `currency` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'USD',
  `occurred_at` timestamp NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `external_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `matched_expense_id` bigint DEFAULT NULL,
  `matched_invoice_id` bigint DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_finance_bank_feeds_external` (`owner_id`,`external_id`),
  KEY `idx_finance_bank_feeds_owner` (`owner_id`),
  KEY `idx_finance_bank_feeds_bank_account` (`bank_account_id`),
  KEY `idx_finance_bank_feeds_occurred` (`occurred_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
