CREATE TABLE IF NOT EXISTS `oauth_clients` (
  `id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `redirect_uris` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `scopes` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_public` tinyint(1) NOT NULL DEFAULT '1',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `oauth_auth_codes` (
  `code` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `client_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` bigint NOT NULL,
  `redirect_uri` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `scope` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `code_challenge` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL,
  `code_challenge_method` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'S256',
  `expires_at` timestamp NOT NULL,
  `used` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`code`),
  KEY `idx_oauth_auth_codes_client` (`client_id`),
  KEY `idx_oauth_auth_codes_user` (`user_id`),
  KEY `idx_oauth_auth_codes_expires` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `oauth_tokens` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `access_token` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL,
  `refresh_token` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `client_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` bigint NOT NULL,
  `scope` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `expires_at` timestamp NOT NULL,
  `revoked` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_oauth_refresh_token` (`refresh_token`),
  KEY `idx_oauth_tokens_access` (`access_token`(64)),
  KEY `idx_oauth_tokens_refresh` (`refresh_token`),
  KEY `idx_oauth_tokens_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO oauth_clients (id, name, redirect_uris, scopes, is_public) VALUES
('rtylr-mcp', 'rtylr MCP Server', '["http://127.0.0.1"]', 'flow:read flow:write crm:read crm:write finance:read finance:write erp:read erp:write', 1);
