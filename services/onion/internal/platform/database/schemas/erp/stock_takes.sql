CREATE TABLE IF NOT EXISTS erp_stock_takes (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    warehouse_id BIGINT,
    status ENUM('draft','in_progress','completed') DEFAULT 'draft',
    notes TEXT,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS erp_stock_take_items (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    stock_take_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    expected_qty DECIMAL(10,3) NOT NULL DEFAULT 0,
    counted_qty DECIMAL(10,3) NULL,
    variance DECIMAL(10,3) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
