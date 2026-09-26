CREATE TABLE IF NOT EXISTS `careers_submission_interviews` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `submission_id` bigint NOT NULL,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `interview_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'video',
  `title` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `scheduled_at` datetime NOT NULL,
  `duration_minutes` int NOT NULL DEFAULT '60',
  `timezone` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'UTC',
  `meeting_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `interviewer_name` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'scheduled',
  `feedback` text COLLATE utf8mb4_unicode_ci,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_submission_interviews_submission` (`submission_id`,`scheduled_at`),
  KEY `idx_submission_interviews_business_time` (`business_id`,`scheduled_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `careers_submission_notes` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `submission_id` bigint NOT NULL,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `author_id` bigint NOT NULL DEFAULT '0',
  `body` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_submission_notes_submission` (`submission_id`,`created_at`),
  KEY `idx_submission_notes_business` (`business_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `careers_submission_offers` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `submission_id` bigint NOT NULL,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `title` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `salary` decimal(12,2) DEFAULT NULL,
  `salary_currency` varchar(8) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'USD',
  `start_date` date DEFAULT NULL,
  `expires_at` date DEFAULT NULL,
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'draft',
  `terms` text COLLATE utf8mb4_unicode_ci,
  `sent_at` datetime DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_submission_offers_submission` (`submission_id`,`created_at`),
  KEY `idx_submission_offers_business_status` (`business_id`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
