-- T-366 part B: each member's linked GitHub account, GitHub webhook delivery
-- ids for dedupe, and webhook health on the installation.
--
-- A GitHub account links to at most one member per business. The user access
-- token used to read the account is never stored.

CREATE TABLE IF NOT EXISTS `flow_github_user_links` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `github_user_id` bigint NOT NULL,
  `github_login` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `github_avatar_url` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_flow_github_user_links_user` (`business_id`,`user_id`),
  UNIQUE KEY `uq_flow_github_user_links_github_user` (`business_id`,`github_user_id`),
  KEY `idx_flow_github_user_links_github_user` (`github_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- One row per processed X-GitHub-Delivery, written in the same transaction as
-- the delivery's effects, so a delivery that failed can be redelivered and a
-- repeat of one that succeeded is ignored. business_id is 0 for app-level
-- events that belong to no installation (github_app_authorization).
CREATE TABLE IF NOT EXISTS `flow_github_webhook_deliveries` (
  `delivery_id` varchar(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `event` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `received_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`delivery_id`),
  KEY `idx_flow_github_webhook_deliveries_received` (`received_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Columns v15 now creates, added to databases that ran v15 before them.
SET @flow_gh_suspended_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_github_installations'
      AND column_name = 'suspended_at'
);
SET @flow_gh_suspended_sql := IF(
    @flow_gh_suspended_missing,
    'ALTER TABLE `flow_github_installations` ADD COLUMN `suspended_at` timestamp NULL DEFAULT NULL AFTER `installed_by`',
    'SELECT 1'
);
PREPARE flow_gh_suspended_statement FROM @flow_gh_suspended_sql;
EXECUTE flow_gh_suspended_statement;
DEALLOCATE PREPARE flow_gh_suspended_statement;

SET @flow_gh_webhook_at_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_github_installations'
      AND column_name = 'last_webhook_at'
);
SET @flow_gh_webhook_at_sql := IF(
    @flow_gh_webhook_at_missing,
    'ALTER TABLE `flow_github_installations` ADD COLUMN `last_webhook_at` timestamp NULL DEFAULT NULL AFTER `suspended_at`',
    'SELECT 1'
);
PREPARE flow_gh_webhook_at_statement FROM @flow_gh_webhook_at_sql;
EXECUTE flow_gh_webhook_at_statement;
DEALLOCATE PREPARE flow_gh_webhook_at_statement;

SET @flow_gh_webhook_event_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_github_installations'
      AND column_name = 'last_webhook_event'
);
SET @flow_gh_webhook_event_sql := IF(
    @flow_gh_webhook_event_missing,
    'ALTER TABLE `flow_github_installations` ADD COLUMN `last_webhook_event` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '''' AFTER `last_webhook_at`',
    'SELECT 1'
);
PREPARE flow_gh_webhook_event_statement FROM @flow_gh_webhook_event_sql;
EXECUTE flow_gh_webhook_event_statement;
DEALLOCATE PREPARE flow_gh_webhook_event_statement;
