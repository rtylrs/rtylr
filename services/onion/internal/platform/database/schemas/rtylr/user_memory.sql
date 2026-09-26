CREATE TABLE IF NOT EXISTS `user_memory` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `memory_type` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `key` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `value` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `confidence` float DEFAULT '1',
  `source_session` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`,`memory_type`,`key`),
  KEY `idx_user_memory_user_id` (`user_id`),
  KEY `idx_user_memory_key` (`key`),
  FULLTEXT KEY `idx_user_memory_search` (`key`,`value`,`memory_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



-- idx_user_memory_search is added by zz_user_memory_search_indexes.sql; adding
-- it here too fails a fresh migration run with a duplicate-key error.
