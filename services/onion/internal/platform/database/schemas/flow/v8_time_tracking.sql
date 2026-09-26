CREATE TABLE IF NOT EXISTS `flow_work_days` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL,
  `employee_id` bigint NOT NULL,
  `date` date NOT NULL,
  `timezone` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `work_mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'onsite',
  `start_time` varchar(5) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `end_time` varchar(5) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `break_minutes` int NOT NULL DEFAULT '0',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'draft',
  `logged_by` bigint DEFAULT NULL,
  `approved_by` bigint DEFAULT NULL,
  `approved_at` timestamp NULL DEFAULT NULL,
  `reopened_by` bigint DEFAULT NULL,
  `reopened_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_work_day` (`owner_id`,`business_id`,`employee_id`,`date`),
  KEY `idx_work_days_tenant_employee_date` (`owner_id`,`business_id`,`employee_id`,`date`),
  KEY `idx_work_days_tenant_status` (`owner_id`,`business_id`,`status`,`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `flow_work_day_actions` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL,
  `work_day_id` bigint NOT NULL,
  `actor_id` bigint NOT NULL,
  `action` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `comment` text COLLATE utf8mb4_unicode_ci,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_work_day_actions_day` (`owner_id`,`business_id`,`work_day_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `flow_work_types` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL,
  `workspace_id` bigint NOT NULL,
  `parent_id` bigint DEFAULT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `hr_category_id` bigint DEFAULT NULL,
  `sort_order` int NOT NULL DEFAULT '0',
  `archived` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_work_types_tenant_workspace` (`owner_id`,`business_id`,`workspace_id`,`archived`,`sort_order`),
  KEY `idx_work_types_parent` (`owner_id`,`business_id`,`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


  

  

  

  

  

  


  


  
