CREATE TABLE IF NOT EXISTS `erp_stock_transfers` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `from_warehouse_id` bigint NOT NULL,
  `to_warehouse_id` bigint NOT NULL,
  `status` enum('draft','in_transit','received','cancelled') COLLATE utf8mb4_unicode_ci DEFAULT 'draft',
  `notes` text COLLATE utf8mb4_unicode_ci,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `erp_stock_transfer_items` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `transfer_id` bigint NOT NULL,
  `product_id` bigint NOT NULL,
  `quantity` decimal(10,3) NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
