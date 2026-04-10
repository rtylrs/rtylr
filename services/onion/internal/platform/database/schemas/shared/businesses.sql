-- Businesses are the top-level tenant. Every piece of data belongs to a business.
-- A user who registers creates a business and becomes its owner.
-- Invited users join an existing business.

CREATE TABLE IF NOT EXISTS businesses (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    owner_id BIGINT NOT NULL COMMENT 'The user who created this business',
    industry VARCHAR(100),
    country VARCHAR(100),
    timezone VARCHAR(50) DEFAULT 'UTC',
    currency VARCHAR(10) DEFAULT 'USD',
    logo_url VARCHAR(512),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_business_owner (owner_id)
);

-- Add business_id to users so we know which business each user belongs to
ALTER TABLE users ADD COLUMN business_id BIGINT NULL AFTER id;
ALTER TABLE users ADD INDEX idx_users_business (business_id);

-- Add business_id to user_app_permissions for cleaner multi-tenancy
ALTER TABLE user_app_permissions ADD COLUMN business_id BIGINT NULL AFTER id;
ALTER TABLE user_app_permissions ADD INDEX idx_perms_business (business_id);

-- Add business_id to invite_tokens
ALTER TABLE invite_tokens ADD COLUMN business_id BIGINT NULL AFTER owner_id;
