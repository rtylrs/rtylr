-- Provisioning state is separate from authorization. It prevents a direct
-- credential invite from being treated as a fully activated account until the
-- required first-login controls have completed.

CREATE TABLE IF NOT EXISTS user_provisioning_states (
    user_id BIGINT NOT NULL PRIMARY KEY,
    business_id BIGINT NOT NULL,
    mode ENUM('invitation','direct_credentials') NOT NULL,
    state ENUM('pending_invitation','pending_activation','active','suspended','revoked') NOT NULL DEFAULT 'pending_activation',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    password_change_required BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_required BOOLEAN NOT NULL DEFAULT TRUE,
    mfa_enrolled BOOLEAN NOT NULL DEFAULT FALSE,
    recovery_available BOOLEAN NOT NULL DEFAULT TRUE,
    activation_token_hash VARCHAR(128) NULL,
    activation_token_expires_at TIMESTAMP NULL,
    password_changed_at TIMESTAMP NULL,
    last_mfa_verified_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_provisioning_business_state (business_id, state),
    INDEX idx_provisioning_activation (activation_token_hash, activation_token_expires_at)
);

CREATE TABLE IF NOT EXISTS user_mfa_enrollments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    method ENUM('totp','webauthn','recovery_code') NOT NULL,
    secret_ciphertext VARBINARY(1024) NULL,
    credential_id VARBINARY(512) NULL,
    recovery_code_hash VARCHAR(128) NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    enrolled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP NULL,
    UNIQUE KEY uk_mfa_user_method_credential (user_id, method, credential_id),
    INDEX idx_mfa_user_active (user_id, is_active)
);
