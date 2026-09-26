CREATE TABLE IF NOT EXISTS `hr_overtime_requests` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `employee_id` bigint NOT NULL,
  `date` date NOT NULL,
  `planned_extra_minutes` int NOT NULL,
  `reason` text COLLATE utf8mb4_unicode_ci,
  `status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending',
  `current_approval_level` int NOT NULL DEFAULT '0',
  `rejection_reason` text COLLATE utf8mb4_unicode_ci,
  `approved_by` bigint DEFAULT NULL,
  `multiplier_applied` decimal(5,2) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_overtime_requests_owner` (`owner_id`),
  KEY `idx_overtime_requests_employee` (`owner_id`,`employee_id`),
  KEY `idx_overtime_requests_status` (`owner_id`,`status`),
  KEY `idx_overtime_requests_date` (`owner_id`,`date`),
  KEY `idx_hr_overtime_business_employee_date_status` (`business_id`,`employee_id`,`date`,`status`),
  KEY `idx_hr_overtime_business_status_date` (`business_id`,`status`,`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `hr_overtime_config` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `daily_threshold_minutes` int NOT NULL DEFAULT '480',
  `default_multiplier` decimal(5,2) NOT NULL DEFAULT '1.50',
  `auto_detect_enabled` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_overtime_config_owner` (`owner_id`),
  KEY `idx_hr_overtime_config_business` (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
