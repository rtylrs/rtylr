CREATE TABLE IF NOT EXISTS `user_provisioning_states` (
  `user_id` bigint NOT NULL,
  `business_id` bigint NOT NULL,
  `mode` enum('invitation','direct_credentials') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `state` enum('pending_invitation','pending_activation','active','suspended','revoked') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending_activation',
  `email_verified` tinyint(1) NOT NULL DEFAULT '0',
  `password_change_required` tinyint(1) NOT NULL DEFAULT '0',
  `mfa_required` tinyint(1) NOT NULL DEFAULT '1',
  `mfa_enrolled` tinyint(1) NOT NULL DEFAULT '0',
  `recovery_available` tinyint(1) NOT NULL DEFAULT '1',
  `activation_token_hash` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `activation_token_expires_at` timestamp NULL DEFAULT NULL,
  `password_changed_at` timestamp NULL DEFAULT NULL,
  `last_mfa_verified_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`),
  KEY `idx_provisioning_business_state` (`business_id`,`state`),
  KEY `idx_provisioning_activation` (`activation_token_hash`,`activation_token_expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_mfa_enrollments` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `method` enum('totp','webauthn','recovery_code') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `secret_ciphertext` varbinary(1024) DEFAULT NULL,
  `credential_id` varbinary(512) DEFAULT NULL,
  `recovery_code_hash` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `is_active` tinyint(1) NOT NULL DEFAULT '1',
  `enrolled_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `revoked_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_mfa_user_method_credential` (`user_id`,`method`,`credential_id`),
  KEY `idx_mfa_user_active` (`user_id`,`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
