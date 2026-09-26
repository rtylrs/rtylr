CREATE TABLE IF NOT EXISTS `hr_payroll_config` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `period_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'monthly',
  `pay_day_of_month` int NOT NULL DEFAULT '28',
  `overtime_enabled` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payroll_config_owner` (`owner_id`),
  KEY `idx_hr_payroll_config_business` (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `hr_payroll_runs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `period_start` date NOT NULL,
  `period_end` date NOT NULL,
  `status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'draft',
  `created_by` bigint NOT NULL,
  `approved_by` bigint DEFAULT NULL,
  `exported_at` timestamp NULL DEFAULT NULL,
  `total_gross` decimal(15,2) NOT NULL DEFAULT '0.00',
  `total_net` decimal(15,2) NOT NULL DEFAULT '0.00',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payroll_runs_business_period` (`business_id`,`period_start`,`period_end`),
  KEY `idx_payroll_runs_owner` (`owner_id`),
  KEY `idx_payroll_runs_period` (`owner_id`,`period_start`,`period_end`),
  KEY `idx_payroll_runs_status` (`owner_id`,`status`),
  KEY `idx_hr_payroll_runs_business_status_created` (`business_id`,`status`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `hr_payroll_entries` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `payroll_run_id` bigint NOT NULL,
  `employee_id` bigint NOT NULL,
  `hours_worked` decimal(8,2) NOT NULL DEFAULT '0.00',
  `overtime_hours` decimal(8,2) NOT NULL DEFAULT '0.00',
  `base_pay` decimal(15,2) NOT NULL DEFAULT '0.00',
  `overtime_pay` decimal(15,2) NOT NULL DEFAULT '0.00',
  `bonuses` decimal(15,2) NOT NULL DEFAULT '0.00',
  `deductions` decimal(15,2) NOT NULL DEFAULT '0.00',
  `net_pay` decimal(15,2) NOT NULL DEFAULT '0.00',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payroll_entries_owner_run_employee` (`owner_id`,`payroll_run_id`,`employee_id`),
  KEY `idx_payroll_entries_owner` (`owner_id`),
  KEY `idx_payroll_entries_run` (`owner_id`,`payroll_run_id`),
  KEY `idx_payroll_entries_employee` (`owner_id`,`employee_id`),
  KEY `idx_hr_payroll_entries_business_run_employee` (`business_id`,`payroll_run_id`,`employee_id`),
  KEY `idx_hr_payroll_entries_business_employee` (`business_id`,`employee_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
