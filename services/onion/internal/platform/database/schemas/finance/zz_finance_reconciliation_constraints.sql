

CREATE TABLE IF NOT EXISTS `finance_reconciliation_targets` (
  `business_id` bigint NOT NULL,
  `target_type` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL,
  `target_id` bigint NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`business_id`,`target_type`,`target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO finance_reconciliation_targets (business_id, target_type, target_id)
SELECT business_id, target_type, target_id
FROM finance_reconciliation_matches;
