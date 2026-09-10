CREATE TABLE IF NOT EXISTS user_email_verification_tokens (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token_hash CHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    consumed_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_email_verification_tokens_lookup (token_hash, consumed_at, expires_at),
    INDEX idx_user_email_verification_tokens_user (user_id)
);
