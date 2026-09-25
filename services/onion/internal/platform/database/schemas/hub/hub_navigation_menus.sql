CREATE TABLE IF NOT EXISTS hub_navigation_menus (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL,
    items JSON NOT NULL DEFAULT (JSON_ARRAY()),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_hub_navigation_menus_business_id (business_id)
);

CREATE INDEX idx_hub_navigation_menus_owner_id ON hub_navigation_menus(owner_id);
