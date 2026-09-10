-- Unified stock tables are platform-owned because POS and ERP share them.
-- Keep the DDL in the main migration stream so every fresh test database and
-- every restored database has the same cross-module stock surface.
CREATE TABLE IF NOT EXISTS stock_locations (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    kind ENUM('warehouse','pos_branch','virtual') NOT NULL,
    source_table ENUM('erp_warehouses','pos_locations','native') NOT NULL,
    source_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_loc_source (owner_id, source_table, source_id),
    INDEX idx_loc_owner (owner_id, is_active),
    INDEX idx_loc_owner_kind (owner_id, kind)
);

CREATE TABLE IF NOT EXISTS stock_levels (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    location_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    on_hand DECIMAL(14,3) NOT NULL DEFAULT 0,
    reserved DECIMAL(14,3) NOT NULL DEFAULT 0,
    reorder_point DECIMAL(14,3) NULL,
    reorder_quantity DECIMAL(14,3) NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_level (owner_id, location_id, product_id),
    INDEX idx_level_owner_product (owner_id, product_id),
    INDEX idx_level_location (owner_id, location_id)
);

CREATE TABLE IF NOT EXISTS stock_ledger (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    location_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    entry_type ENUM('purchase','sale','return','adjustment','waste','count','transfer_in','transfer_out','reserve','release','opening') NOT NULL,
    quantity_change DECIMAL(14,3) NOT NULL,
    quantity_before DECIMAL(14,3) NOT NULL,
    quantity_after DECIMAL(14,3) NOT NULL,
    reference_id BIGINT NULL,
    reference_type VARCHAR(50) NULL,
    transfer_group BIGINT NULL,
    channel ENUM('pos','erp') NOT NULL DEFAULT 'erp',
    note TEXT NULL,
    employee_id BIGINT NULL,
    created_by BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_ledger_reference (owner_id, reference_type, reference_id, entry_type),
    INDEX idx_ledger_owner_product (owner_id, product_id, created_at),
    INDEX idx_ledger_owner_loc (owner_id, location_id, created_at),
    INDEX idx_ledger_ref (reference_type, reference_id),
    INDEX idx_ledger_transfer (transfer_group)
);

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'erp_purchase_orders'
       AND COLUMN_NAME = 'destination_location_id') = 0,
    'ALTER TABLE erp_purchase_orders ADD COLUMN destination_location_id BIGINT NULL',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE OR REPLACE VIEW product_stock_total AS
    SELECT owner_id, business_id, product_id, SUM(on_hand) AS total_quantity
    FROM stock_levels
    GROUP BY owner_id, business_id, product_id;
