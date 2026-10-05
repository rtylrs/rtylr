-- 005_product_pim.sql
-- Run MANUALLY on prod before deploying the PIM backend (Stage 2/3).
-- Adds rich product fields + the media/attributes tables. MySQL 8 has no
-- "ADD COLUMN IF NOT EXISTS"; run once. Re-running the ALTER will error on
-- duplicate columns (safe to ignore if already applied).

ALTER TABLE erp_products
    ADD COLUMN compare_at_price DECIMAL(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN tax_rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    ADD COLUMN weight DECIMAL(10,3) NOT NULL DEFAULT 0,
    ADD COLUMN weight_unit VARCHAR(8) NOT NULL DEFAULT 'kg',
    ADD COLUMN length DECIMAL(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN width DECIMAL(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN height DECIMAL(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN dimension_unit VARCHAR(8) NOT NULL DEFAULT 'cm',
    ADD COLUMN country_of_origin VARCHAR(2) DEFAULT NULL,
    ADD COLUMN hs_code VARCHAR(20) DEFAULT NULL,
    ADD COLUMN allow_backorder BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN lead_time_days INT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS erp_product_images (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    url VARCHAR(1024) NOT NULL,
    alt VARCHAR(255) DEFAULT NULL,
    position INT NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_product_images (owner_id, product_id, position)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS erp_product_attributes (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    name VARCHAR(120) NOT NULL,
    value VARCHAR(1024) NOT NULL,
    position INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_product_attributes (owner_id, product_id, position)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
