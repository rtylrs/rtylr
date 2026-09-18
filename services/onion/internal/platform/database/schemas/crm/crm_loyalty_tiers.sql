CREATE TABLE IF NOT EXISTS crm_loyalty_tiers (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(100) NOT NULL,
    min_points INT DEFAULT 0,
    discount_pct DECIMAL(5,2) DEFAULT 0,
    perks JSON,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_crm_loyalty_tiers_business (business_id)
);
