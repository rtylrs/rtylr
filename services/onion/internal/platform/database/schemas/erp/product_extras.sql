CREATE TABLE IF NOT EXISTS `erp_product_images` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `product_id` bigint NOT NULL,
  `url` varchar(1024) NOT NULL,
  `alt` varchar(255) DEFAULT NULL,
  `position` int NOT NULL DEFAULT '0',
  `is_primary` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_product_images` (`owner_id`,`product_id`,`position`),
  KEY `idx_product_images_business_product_primary` (`business_id`,`product_id`,`is_primary`,`position`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `erp_product_attributes` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `product_id` bigint NOT NULL,
  `name` varchar(120) NOT NULL,
  `value` varchar(1024) NOT NULL,
  `position` int NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_product_attributes` (`owner_id`,`product_id`,`position`),
  KEY `idx_product_attributes_business_product_position` (`business_id`,`product_id`,`position`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
