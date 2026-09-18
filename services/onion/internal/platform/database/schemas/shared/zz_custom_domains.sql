CREATE TABLE IF NOT EXISTS custom_domains (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL,
    domain VARCHAR(255) NOT NULL,
    verification_token VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending_verification',
    ssl_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    verified_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_custom_domains_domain (domain),
    UNIQUE KEY uk_custom_domains_business (business_id),
    INDEX idx_custom_domains_status (status)
);
