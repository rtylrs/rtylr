CREATE TABLE IF NOT EXISTS `careers_form_fields` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `career_id` bigint NOT NULL,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `kind` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `label` varchar(240) COLLATE utf8mb4_unicode_ci NOT NULL,
  `placeholder` varchar(240) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `required` tinyint(1) NOT NULL DEFAULT '0',
  `options_json` json DEFAULT NULL,
  `position` int NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_careers_form_fields_career` (`career_id`),
  KEY `idx_careers_form_fields_owner` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


