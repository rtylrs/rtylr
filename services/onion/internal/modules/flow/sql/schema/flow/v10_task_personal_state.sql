-- Existing installations record schema.sql once; new task state tables need their own migration.
CREATE TABLE IF NOT EXISTS `flow_task_focus_pins` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `task_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_flow_task_focus_pins` (`task_id`,`user_id`),
  KEY `idx_flow_task_focus_pins_user` (`user_id`,`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `flow_task_dependencies` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `task_id` bigint NOT NULL,
  `depends_on_task_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_flow_task_dependencies` (`task_id`,`depends_on_task_id`),
  KEY `idx_flow_task_dependencies_task` (`task_id`,`business_id`),
  KEY `idx_flow_task_dependencies_depends_on` (`depends_on_task_id`,`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
