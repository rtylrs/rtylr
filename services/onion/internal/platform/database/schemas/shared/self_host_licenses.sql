CREATE TABLE IF NOT EXISTS `self_host_licenses` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `customer_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `plan` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'self_hosted',
  `features` json DEFAULT NULL,
  `seat_limit` int DEFAULT NULL,
  `expires_at` timestamp NULL DEFAULT NULL,
  `grace_until` timestamp NULL DEFAULT NULL,
  `status` enum('active','revoked') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `license_key` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `key_hash` char(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_by` bigint DEFAULT NULL,
  `updated_by` bigint DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `regenerated_at` timestamp NULL DEFAULT NULL,
  `revoked_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_self_host_licenses_owner_email` (`owner_email`),
  UNIQUE KEY `uq_self_host_licenses_key_hash` (`key_hash`),
  KEY `idx_self_host_licenses_status` (`status`),
  KEY `idx_self_host_licenses_created_at` (`created_at`),
  FULLTEXT KEY `idx_self_host_licenses_search` (`owner_email`,`customer_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- The FULLTEXT search index is added by zz_self_host_license_search_indexes.sql
-- rather than inline: MySQL has no ADD INDEX IF NOT EXISTS, so declaring it here
-- as well makes the zz migration fail on a fresh database with a duplicate key
-- error and stops the whole migration run. Every other zz_* search-index
-- migration in this repo owns its index the same way.
