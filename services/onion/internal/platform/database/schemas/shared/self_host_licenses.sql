CREATE TABLE IF NOT EXISTS self_host_licenses (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_email VARCHAR(255) NOT NULL,
    customer_name VARCHAR(255) NULL,
    plan VARCHAR(100) NOT NULL DEFAULT 'self_hosted',
    features JSON NULL,
    seat_limit INT NULL,
    expires_at TIMESTAMP NULL,
    grace_until TIMESTAMP NULL,
    status ENUM('active','revoked') NOT NULL DEFAULT 'active',
    license_key TEXT NOT NULL,
    key_hash CHAR(64) NOT NULL,
    created_by BIGINT NULL,
    updated_by BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    regenerated_at TIMESTAMP NULL,
    revoked_at TIMESTAMP NULL,
    UNIQUE KEY uq_self_host_licenses_owner_email (owner_email),
    UNIQUE KEY uq_self_host_licenses_key_hash (key_hash),
    INDEX idx_self_host_licenses_status (status),
    INDEX idx_self_host_licenses_created_at (created_at)
);

-- The FULLTEXT search index is added by zz_self_host_license_search_indexes.sql
-- rather than inline: MySQL has no ADD INDEX IF NOT EXISTS, so declaring it here
-- as well makes the zz migration fail on a fresh database with a duplicate key
-- error and stops the whole migration run. Every other zz_* search-index
-- migration in this repo owns its index the same way.
