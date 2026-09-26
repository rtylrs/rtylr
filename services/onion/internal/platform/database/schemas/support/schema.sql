CREATE TABLE IF NOT EXISTS `support_tickets` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `customer_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `customer_email` varchar(320) COLLATE utf8mb4_unicode_ci NOT NULL,
  `subject` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` enum('open','pending','resolved','closed') COLLATE utf8mb4_unicode_ci DEFAULT 'open',
  `priority` enum('low','medium','high','urgent') COLLATE utf8mb4_unicode_ci DEFAULT 'medium',
  `assigned_to` bigint DEFAULT NULL,
  `channel` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'email',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_support_tickets_business` (`business_id`),
  KEY `idx_support_tickets_business_status` (`business_id`,`status`),
  KEY `idx_support_tickets_business_updated` (`business_id`,`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `support_messages` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `ticket_id` bigint NOT NULL,
  `sender_type` enum('customer','agent','system') COLLATE utf8mb4_unicode_ci NOT NULL,
  `sender_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `body` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_internal` tinyint(1) DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_support_messages_ticket` (`ticket_id`),
  KEY `idx_support_messages_ticket_created` (`ticket_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `support_customers` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(320) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `phone` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `customer_id` bigint DEFAULT NULL,
  `total_tickets` int DEFAULT '0',
  `last_ticket_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_support_customers_business_email` (`business_id`,`email`),
  KEY `idx_support_customers_customer` (`business_id`,`customer_id`),
  KEY `idx_support_customers_business` (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `support_articles` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `title` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL,
  `slug` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL,
  `body` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `category` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `is_published` tinyint(1) DEFAULT '0',
  `views` int DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_support_articles_business_slug` (`business_id`,`slug`),
  KEY `idx_support_articles_business_title_updated` (`business_id`,`title`,`updated_at`,`id`),
  KEY `idx_support_articles_business` (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `support_canned_responses` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `body` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `category` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `shortcut` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_support_canned_business` (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `support_settings` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `auto_assign` tinyint(1) DEFAULT '0',
  `sla_hours` int DEFAULT '24',
  `business_hours` json DEFAULT (json_object()),
  `notification_email` varchar(320) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `business_id` (`business_id`),
  CONSTRAINT `support_settings_chk_1` CHECK ((`sla_hours` >= 0))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
