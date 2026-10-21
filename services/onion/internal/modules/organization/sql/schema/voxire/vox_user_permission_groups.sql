CREATE TABLE IF NOT EXISTS `vox_user_permission_groups` (
  `user_id` int NOT NULL,
  `permission_group_id` int NOT NULL,
  `assigned_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `assigned_by` int DEFAULT NULL,
  PRIMARY KEY (`user_id`,`permission_group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
