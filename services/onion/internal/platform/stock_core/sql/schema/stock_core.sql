-- Stock core tables (authoritative DDL: erp/migrations/002_unified_stock.sql).
CREATE TABLE IF NOT EXISTS stock_locations (
    id           BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id     BIGINT NOT NULL,
    business_id  BIGINT NOT NULL DEFAULT 0,
    kind         ENUM('warehouse','pos_branch','virtual') NOT NULL,
    source_table ENUM('erp_warehouses','pos_locations','native') NOT NULL,
    source_id    BIGINT NOT NULL DEFAULT 0,
    name         VARCHAR(255) NOT NULL,
    is_default   BOOLEAN NOT NULL DEFAULT FALSE,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_loc_source (owner_id, source_table, source_id)
);

CREATE TABLE IF NOT EXISTS stock_levels (
    id               BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id         BIGINT NOT NULL,
    business_id      BIGINT NOT NULL DEFAULT 0,
    location_id      BIGINT NOT NULL,
    product_id       BIGINT NOT NULL,
    on_hand          DECIMAL(14,3) NOT NULL DEFAULT 0,
    reserved         DECIMAL(14,3) NOT NULL DEFAULT 0,
    reorder_point    DECIMAL(14,3) NULL,
    reorder_quantity DECIMAL(14,3) NULL,
    updated_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_level (owner_id, location_id, product_id)
);

CREATE TABLE IF NOT EXISTS stock_ledger (
    id              BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id        BIGINT NOT NULL,
    business_id     BIGINT NOT NULL DEFAULT 0,
    location_id     BIGINT NOT NULL,
    product_id      BIGINT NOT NULL,
    entry_type      ENUM('purchase','sale','return','adjustment','waste','count','transfer_in','transfer_out','reserve','release','opening') NOT NULL,
    quantity_change DECIMAL(14,3) NOT NULL,
    quantity_before DECIMAL(14,3) NOT NULL,
    quantity_after  DECIMAL(14,3) NOT NULL,
    reference_id    BIGINT NULL,
    reference_type  VARCHAR(50) NULL,
    transfer_group  BIGINT NULL,
    channel         ENUM('pos','erp') NOT NULL DEFAULT 'erp',
    note            TEXT NULL,
    employee_id     BIGINT NULL,
    created_by      BIGINT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_ledger_reference (owner_id, reference_type, reference_id, entry_type, product_id, location_id)
);

CREATE OR REPLACE VIEW product_stock_total AS
    SELECT owner_id, business_id, product_id, SUM(on_hand) AS total_quantity
    FROM stock_levels
    GROUP BY owner_id, business_id, product_id;
