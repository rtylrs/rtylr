CREATE TABLE IF NOT EXISTS `hr_benefit_enrollments` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `employee_id` bigint NOT NULL,
  `plan_id` bigint NOT NULL,
  `status` enum('active','cancelled') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `enrolled_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `cancelled_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_hr_benefit_enrollment` (`business_id`,`employee_id`,`plan_id`),
  KEY `idx_hr_benefit_enrollments_employee` (`business_id`,`employee_id`,`enrolled_at`),
  KEY `idx_hr_benefit_enrollments_plan` (`business_id`,`plan_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
