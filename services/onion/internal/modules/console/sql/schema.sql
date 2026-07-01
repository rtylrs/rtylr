-- Scaffold for future self_host_licenses sqlc migration.
CREATE TABLE IF NOT EXISTS self_host_licenses (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    license_key VARCHAR(128) NOT NULL,
    owner_id BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_self_host_license_key (license_key)
);
