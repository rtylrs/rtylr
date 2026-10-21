-- One row per goods receipt. Stock ledger rows reference the receipt, not the
-- purchase order, so a second partial delivery is a new movement instead of a
-- duplicate of the first.
CREATE TABLE IF NOT EXISTS `erp_purchase_order_receipts` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL,
  `purchase_order_id` bigint NOT NULL,
  `received_by` bigint DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_po_receipts_order` (`business_id`, `purchase_order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
