-- Product media gallery. One row per image; is_primary marks the main image.
CREATE TABLE IF NOT EXISTS erp_product_images (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    product_id BIGINT NOT NULL,
    url VARCHAR(1024) NOT NULL,
    alt VARCHAR(255) DEFAULT NULL,
    position INT NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_product_images (owner_id, product_id, position),
    INDEX idx_product_images_business_product_primary (business_id, product_id, is_primary, position, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Per-business custom product attributes (key/value), so a restaurant and a
-- law firm can capture different fields on the same product model.
CREATE TABLE IF NOT EXISTS erp_product_attributes (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    product_id BIGINT NOT NULL,
    name VARCHAR(120) NOT NULL,
    value VARCHAR(1024) NOT NULL,
    position INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_product_attributes (owner_id, product_id, position),
    INDEX idx_product_attributes_business_product_position (business_id, product_id, position, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
