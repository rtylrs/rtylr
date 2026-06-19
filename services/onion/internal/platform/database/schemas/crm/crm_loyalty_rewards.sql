-- Loyalty rewards catalog. Businesses define redeemable rewards
-- (free item, % off, fixed discount) that members can exchange points for.
CREATE TABLE IF NOT EXISTS crm_loyalty_rewards (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(150) NOT NULL,
    description TEXT,
    category VARCHAR(50) DEFAULT 'general' COMMENT 'e.g. food, drink, merch, discount, experience',
    points_cost INT NOT NULL DEFAULT 0,
    stock INT NULL COMMENT 'NULL = unlimited',
    icon VARCHAR(50) COMMENT 'Optional icon key for UI',
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_lr_owner (owner_id),
    INDEX idx_lr_active (owner_id, active),
    INDEX idx_lr_business (business_id)
);
