-- Add parent_id to businesses table
ALTER TABLE businesses ADD COLUMN parent_id BIGINT NULL AFTER owner_id;
ALTER TABLE businesses ADD INDEX idx_businesses_parent (parent_id);

-- New table: tracks which users can access which businesses (beyond direct ownership)
CREATE TABLE IF NOT EXISTS user_business_access (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL,
    role ENUM('owner','admin','viewer') NOT NULL DEFAULT 'owner',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_uba_user_business (user_id, business_id),
    INDEX idx_uba_user (user_id),
    INDEX idx_uba_business (business_id)
);
