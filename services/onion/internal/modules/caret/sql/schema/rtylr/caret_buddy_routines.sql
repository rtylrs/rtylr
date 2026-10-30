CREATE TABLE IF NOT EXISTS `caret_buddy_routines` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `buddy_id` bigint NOT NULL,
  `business_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `name` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `instructions` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `every_minutes` int NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `next_run_at` timestamp NULL DEFAULT NULL,
  `last_run_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_caret_buddy_routines_due` (`enabled`,`next_run_at`),
  KEY `idx_caret_buddy_routines_buddy` (`buddy_id`),
  KEY `idx_caret_buddy_routines_owner` (`business_id`,`user_id`),
  CONSTRAINT `caret_buddy_routines_every` CHECK ((`every_minutes` between 15 and 43200))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
