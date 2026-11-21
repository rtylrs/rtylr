-- Custom fields (T-580): a project defines its own fields and every task in
-- the project can carry one value per field. Both statements are idempotent.

-- Field names are unique per project. The column collation is case-insensitive,
-- so the unique key refuses "Severity" beside "severity".
CREATE TABLE IF NOT EXISTS `flow_custom_fields` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `project_id` bigint NOT NULL,
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `options` json DEFAULT NULL,
  `required` tinyint(1) NOT NULL DEFAULT '0',
  `default_value` json DEFAULT NULL,
  `sort_order` int NOT NULL DEFAULT '0',
  `created_by` bigint NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_flow_custom_fields_project_name` (`project_id`, `name`),
  KEY `idx_flow_custom_fields_business_project` (`business_id`, `project_id`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- One row per task and field. The value sits in the column that matches the
-- field type so it can be filtered and sorted in SQL: value_text holds text,
-- url and a select's option id; value_number holds number and checkbox (0/1);
-- value_date holds date; value_json holds a JSON array of strings (option ids
-- for multi_select, user ids for person).
CREATE TABLE IF NOT EXISTS `flow_task_field_values` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `task_id` bigint NOT NULL,
  `field_id` bigint NOT NULL,
  `value_text` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `value_number` decimal(24,6) DEFAULT NULL,
  `value_date` date DEFAULT NULL,
  `value_json` json DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_flow_task_field_values_task_field` (`task_id`, `field_id`),
  KEY `idx_flow_task_field_values_number` (`field_id`, `value_number`),
  KEY `idx_flow_task_field_values_date` (`field_id`, `value_date`),
  KEY `idx_flow_task_field_values_text` (`field_id`, `value_text`(191)),
  KEY `idx_flow_task_field_values_business` (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
