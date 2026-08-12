CREATE TABLE IF NOT EXISTS inventory_cost_layers (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    product_id BIGINT NOT NULL,
    location_id BIGINT NOT NULL,
    quantity_remaining DECIMAL(18,6) NOT NULL,
    unit_cost DECIMAL(18,6) NOT NULL,
    received_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reference_id BIGINT NULL,
    reference_type VARCHAR(50) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_cost_layers_owner_product_location (owner_id, product_id, location_id),
    INDEX idx_cost_layers_owner_received (owner_id, received_at)
);
