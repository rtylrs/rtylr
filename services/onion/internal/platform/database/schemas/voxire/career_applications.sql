CREATE TABLE IF NOT EXISTS `voxire_career_applications` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `job_id` bigint NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `phone` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `work_history` text COLLATE utf8mb4_unicode_ci,
  `education` text COLLATE utf8mb4_unicode_ci,
  `experience_summary` text COLLATE utf8mb4_unicode_ci,
  `cover_letter` text COLLATE utf8mb4_unicode_ci,
  `cv_key` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `location` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `salary_expectation` decimal(12,2) DEFAULT NULL,
  `salary_currency` varchar(8) COLLATE utf8mb4_unicode_ci DEFAULT 'USD',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'new',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_career_apps_job` (`job_id`),
  KEY `idx_career_apps_job_created` (`job_id`,`created_at`),
  KEY `idx_career_apps_status` (`status`),
  KEY `idx_career_apps_job_status` (`job_id`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `voxhire_admins` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_voxhire_admins_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
