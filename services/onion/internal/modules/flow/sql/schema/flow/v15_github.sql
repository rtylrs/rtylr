-- T-366: a business's GitHub App installation and the repositories each Flow
-- project picks from it.
--
-- installation_id is unique across businesses so one GitHub account cannot be
-- attached to two tenants: whoever completes the install first holds it, and a
-- second business guessing the same id is rejected.

CREATE TABLE IF NOT EXISTS `flow_github_installations` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `installation_id` bigint NOT NULL,
  `account_login` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `account_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `installed_by` bigint NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_flow_github_installations_business` (`business_id`),
  UNIQUE KEY `uq_flow_github_installations_installation` (`installation_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `flow_project_repositories` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `project_id` bigint NOT NULL,
  `installation_id` bigint NOT NULL,
  `repository_id` bigint NOT NULL,
  `full_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_private` tinyint(1) NOT NULL DEFAULT '0',
  `added_by` bigint NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_flow_project_repositories_project_repo` (`project_id`,`repository_id`),
  KEY `idx_flow_project_repositories_business_project` (`business_id`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
