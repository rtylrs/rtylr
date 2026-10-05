CREATE TABLE IF NOT EXISTS `finance_reconciliation_matches` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `bank_feed_id` bigint NOT NULL,
  `target_type` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `target_id` bigint NOT NULL,
  `matched_by` bigint DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_finance_reconciliation_business_bank` (`business_id`,`bank_feed_id`),
  KEY `idx_finance_reconciliation_owner` (`owner_id`),
  KEY `idx_finance_reconciliation_bank` (`bank_feed_id`),
  KEY `idx_finance_reconciliation_target` (`target_type`,`target_id`),
  KEY `idx_finance_reconciliation_business_bank` (`business_id`,`bank_feed_id`),
  KEY `idx_finance_reconciliation_business_target` (`business_id`,`target_type`,`target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
