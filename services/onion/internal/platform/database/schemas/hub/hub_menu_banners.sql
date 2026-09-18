CREATE TABLE IF NOT EXISTS hub_menu_banners (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NULL,
    image_url VARCHAR(2048) NULL,
    link_url VARCHAR(2048) NULL,
    position INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    starts_at TIMESTAMP NULL,
    ends_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_hub_menu_banners_business_active ON hub_menu_banners(business_id, is_active);
CREATE INDEX idx_hub_menu_banners_business_position ON hub_menu_banners(business_id, position, id);
