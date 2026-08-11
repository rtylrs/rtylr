-- Unified stock model — Slice 0 (foundation: the three tables stock_core needs).
-- Introduces one location namespace, one per-location balance table, and one
-- ledger that BOTH POS and ERP share (POS menu items already ARE erp_products).
-- These tables are live and required at runtime. stock_core.ApplyMovementTx
-- (internal/platform/stock_core/stock_core.go) is the only sanctioned write
-- path, and 8 modules route through it: products, inventory, purchase_orders,
-- stock_takes, stock_transfers, pos_orders, pos_stock_movements, store_orders.
-- Each movement row-locks the stock_levels row, appends to stock_ledger, and
-- adds the same delta to erp_products.stock_quantity in the same transaction
-- (an incremental add, deliberately not a SUM recompute, so it cannot be
-- clobbered by a stale REPEATABLE READ snapshot). The legacy scalar therefore
-- stays equal to the owner-wide SUM(on_hand) by construction.
-- Additive and reversible.
-- Slice status: backfill, dual-write (adjust, PO receive, stock-take), POS sale
-- deduction and transfers have all shipped. A read cutover is NOT required:
-- existing ERP reads still select the scalar and are already correct via that
-- sync, while POS branch-inventory reads select stock_levels.on_hand directly
-- today. The scalar is not yet decommissioned. Still outstanding: POS manual
-- ops and POS stock reads that go through pos_branch_inventory, which need a
-- fail-closed business_id -> owner_id reconciliation first.
-- See vault/rtylr/erp-readiness.md.

CREATE TABLE IF NOT EXISTS stock_locations (
    id           BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id     BIGINT NOT NULL,                 -- canonical tenant = users.id = businesses.owner_id
    business_id  BIGINT NOT NULL DEFAULT 0,       -- denormalized for POS joins; 0 = unset
    kind         ENUM('warehouse','pos_branch','virtual') NOT NULL,
    source_table ENUM('erp_warehouses','pos_locations','native') NOT NULL,
    source_id    BIGINT NOT NULL DEFAULT 0,       -- by-convention link to erp_warehouses.id / pos_locations.id
    name         VARCHAR(255) NOT NULL,
    is_default   BOOLEAN NOT NULL DEFAULT FALSE,  -- the location legacy scalar stock migrates into
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_loc_source (owner_id, source_table, source_id),
    INDEX idx_loc_owner (owner_id, is_active),
    INDEX idx_loc_owner_kind (owner_id, kind)
);

CREATE TABLE IF NOT EXISTS stock_levels (
    id               BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id         BIGINT NOT NULL,
    business_id      BIGINT NOT NULL DEFAULT 0,
    location_id      BIGINT NOT NULL,             -- -> stock_locations.id
    product_id       BIGINT NOT NULL,             -- -> erp_products.id (the only product identity)
    on_hand          DECIMAL(14,3) NOT NULL DEFAULT 0,
    reserved         DECIMAL(14,3) NOT NULL DEFAULT 0,
    reorder_point    DECIMAL(14,3) NULL,
    reorder_quantity DECIMAL(14,3) NULL,
    updated_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_level (owner_id, location_id, product_id),  -- the FOR UPDATE row-lock target
    INDEX idx_level_owner_product (owner_id, product_id),
    INDEX idx_level_location (owner_id, location_id)
);

CREATE TABLE IF NOT EXISTS stock_ledger (
    id              BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id        BIGINT NOT NULL,
    business_id     BIGINT NOT NULL DEFAULT 0,
    location_id     BIGINT NOT NULL,             -- -> stock_locations.id
    product_id      BIGINT NOT NULL,             -- -> erp_products.id
    entry_type      ENUM('purchase','sale','return','adjustment','waste','count','transfer_in','transfer_out','reserve','release','opening') NOT NULL,
    quantity_change DECIMAL(14,3) NOT NULL,       -- signed: +receive, -sale
    quantity_before DECIMAL(14,3) NOT NULL,
    quantity_after  DECIMAL(14,3) NOT NULL,
    reference_id    BIGINT NULL,
    reference_type  VARCHAR(50) NULL,             -- 'purchase_order','stock_take','stock_transfer','pos_order','manual','migration'
    transfer_group  BIGINT NULL,                  -- pairs transfer_out + transfer_in into one logical move
    channel         ENUM('pos','erp') NOT NULL DEFAULT 'erp',  -- provenance only, NEVER a balance partition
    note            TEXT NULL,
    employee_id     BIGINT NULL,
    created_by      BIGINT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ledger_owner_product (owner_id, product_id, created_at),
    INDEX idx_ledger_owner_loc (owner_id, location_id, created_at),
    INDEX idx_ledger_ref (reference_type, reference_id),
    INDEX idx_ledger_transfer (transfer_group)
);

-- Purchase orders gain an optional receiving location (defaults to the tenant's
-- default warehouse when null). Nullable + additive.
ALTER TABLE erp_purchase_orders ADD COLUMN destination_location_id BIGINT NULL;

-- Transitional read view: owner-wide total across locations (matches the legacy
-- scalar's semantics) for readers during the migration.
CREATE OR REPLACE VIEW product_stock_total AS
    SELECT owner_id, business_id, product_id, SUM(on_hand) AS total_quantity
    FROM stock_levels
    GROUP BY owner_id, business_id, product_id;
