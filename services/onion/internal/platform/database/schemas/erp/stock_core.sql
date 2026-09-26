CREATE TABLE IF NOT EXISTS `stock_locations` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `kind` enum('warehouse','pos_branch','virtual') COLLATE utf8mb4_unicode_ci NOT NULL,
  `source_table` enum('erp_warehouses','pos_locations','native') COLLATE utf8mb4_unicode_ci NOT NULL,
  `source_id` bigint NOT NULL DEFAULT '0',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_default` tinyint(1) NOT NULL DEFAULT '0',
  `is_active` tinyint(1) NOT NULL DEFAULT '1',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_loc_source` (`owner_id`,`source_table`,`source_id`),
  KEY `idx_loc_owner` (`owner_id`,`is_active`),
  KEY `idx_loc_owner_kind` (`owner_id`,`kind`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `stock_levels` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `location_id` bigint NOT NULL,
  `product_id` bigint NOT NULL,
  `on_hand` decimal(14,3) NOT NULL DEFAULT '0.000',
  `reserved` decimal(14,3) NOT NULL DEFAULT '0.000',
  `reorder_point` decimal(14,3) DEFAULT NULL,
  `reorder_quantity` decimal(14,3) DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_level` (`owner_id`,`location_id`,`product_id`),
  KEY `idx_level_owner_product` (`owner_id`,`product_id`),
  KEY `idx_level_location` (`owner_id`,`location_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `stock_ledger` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `location_id` bigint NOT NULL,
  `product_id` bigint NOT NULL,
  `entry_type` enum('purchase','sale','return','adjustment','waste','count','transfer_in','transfer_out','reserve','release','opening') COLLATE utf8mb4_unicode_ci NOT NULL,
  `quantity_change` decimal(14,3) NOT NULL,
  `quantity_before` decimal(14,3) NOT NULL,
  `quantity_after` decimal(14,3) NOT NULL,
  `reference_id` bigint DEFAULT NULL,
  `reference_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `transfer_group` bigint DEFAULT NULL,
  `channel` enum('pos','erp') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'erp',
  `note` text COLLATE utf8mb4_unicode_ci,
  `employee_id` bigint DEFAULT NULL,
  `created_by` bigint DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_ledger_reference` (`owner_id`,`reference_type`,`reference_id`,`entry_type`,`product_id`,`location_id`),
  KEY `idx_ledger_owner_product` (`owner_id`,`product_id`,`created_at`),
  KEY `idx_ledger_owner_loc` (`owner_id`,`location_id`,`created_at`),
  KEY `idx_ledger_ref` (`reference_type`,`reference_id`),
  KEY `idx_ledger_transfer` (`transfer_group`),
  KEY `idx_stock_ledger_business_product_created` (`business_id`,`product_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;






CREATE OR REPLACE VIEW `product_stock_total` AS select `stock_levels`.`owner_id` AS `owner_id`,`stock_levels`.`business_id` AS `business_id`,`stock_levels`.`product_id` AS `product_id`,sum(`stock_levels`.`on_hand`) AS `total_quantity` from `stock_levels` group by `stock_levels`.`owner_id`,`stock_levels`.`business_id`,`stock_levels`.`product_id`;
