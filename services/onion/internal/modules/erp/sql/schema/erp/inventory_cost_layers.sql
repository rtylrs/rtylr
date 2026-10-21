CREATE TABLE IF NOT EXISTS `inventory_cost_layers` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `product_id` bigint NOT NULL,
  `location_id` bigint NOT NULL,
  `quantity_remaining` decimal(18,6) NOT NULL,
  `unit_cost` decimal(18,6) NOT NULL,
  `received_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `reference_id` bigint DEFAULT NULL,
  `reference_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_cost_layers_owner_product_location` (`owner_id`,`product_id`,`location_id`),
  KEY `idx_cost_layers_owner_received` (`owner_id`,`received_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
