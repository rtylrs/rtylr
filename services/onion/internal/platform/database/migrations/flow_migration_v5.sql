-- Flow schema migration v5: business entity
CREATE TABLE IF NOT EXISTS flow_businesses (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(20) DEFAULT '#64C5B9',
    icon VARCHAR(50) DEFAULT '🏢',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_businesses_owner (owner_id)
);
-- workspace already has business_id column (added in initial schema)
