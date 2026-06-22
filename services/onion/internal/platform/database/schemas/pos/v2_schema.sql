-- ═══════════════════════════════════════════════════════════════════
-- POS V2 Schema: Multi-branch, multi-terminal, full engine
-- ═══════════════════════════════════════════════════════════════════

-- ── Floors (per branch) ──────────────────────────────────────────
CREATE TABLE IF NOT EXISTS pos_floors (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    sort_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_floors_branch (branch_id)
);

-- ── Add floor_id to tables ───────────────────────────────────────
ALTER TABLE pos_tables
  ADD COLUMN floor_id BIGINT DEFAULT NULL,
  ADD COLUMN sort_order INT DEFAULT 0;

-- ── Terminals (registered POS devices per branch) ────────────────
CREATE TABLE IF NOT EXISTS pos_terminals (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    device_id VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    last_seen_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_terminals_branch (branch_id)
);

-- Runtime session/order columns are defined in the consolidated base schema.
-- Keeping them there avoids duplicate-column errors in schema generation.

-- Update order status enum to include voided and held
ALTER TABLE pos_orders MODIFY COLUMN status
  ENUM('pending','confirmed','preparing','ready','served','paid','cancelled','refunded','voided','held')
  DEFAULT 'pending';

-- ── Order payments (split payment support) ───────────────────────
CREATE TABLE IF NOT EXISTS pos_order_payments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT NOT NULL,
    method ENUM('cash','card','online','gift_card','store_credit','other') NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    reference VARCHAR(255),
    card_last4 VARCHAR(4),
    gift_card_code VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_payments_order (order_id)
);

-- ── Add modifiers JSON to order items ────────────────────────────
ALTER TABLE pos_order_items
  ADD COLUMN modifiers JSON DEFAULT (JSON_ARRAY()),
  ADD COLUMN combo_id BIGINT DEFAULT NULL,
  ADD COLUMN variant_name VARCHAR(255) DEFAULT NULL;

-- ── Audit trail ──────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS pos_audit_log (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    terminal_id BIGINT DEFAULT NULL,
    employee_id BIGINT NOT NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT DEFAULT NULL,
    details JSON DEFAULT (JSON_OBJECT()),
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_audit_business (business_id, created_at),
    INDEX idx_audit_action (action)
);

-- ── Cash movements (drops, pickups, adjustments) ─────────────────
CREATE TABLE IF NOT EXISTS pos_cash_movements (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    terminal_id BIGINT DEFAULT NULL,
    session_id BIGINT NOT NULL,
    employee_id BIGINT NOT NULL,
    type ENUM('drop','pickup','adjustment','float','tip_out','cash_in','cash_out') NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    reason TEXT,
    approved_by BIGINT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_cash_session (session_id)
);

-- ── Loyalty program ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS pos_loyalty_config (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL UNIQUE,
    is_enabled BOOLEAN DEFAULT FALSE,
    points_per_dollar DECIMAL(5,2) DEFAULT 1.00,
    redemption_rate DECIMAL(5,2) DEFAULT 0.01,
    min_redeem_points INT DEFAULT 100,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pos_loyalty_transactions (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    order_id BIGINT DEFAULT NULL,
    type ENUM('earn','redeem','adjust','expire') NOT NULL,
    points INT NOT NULL,
    balance_after INT NOT NULL,
    note TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_loyalty_customer (customer_id)
);

-- ── Store credit ─────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS pos_store_credit (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    balance DECIMAL(10,2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_customer_credit (business_id, customer_id)
);

CREATE TABLE IF NOT EXISTS pos_store_credit_transactions (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    order_id BIGINT DEFAULT NULL,
    type ENUM('issue','redeem','refund_credit','adjust') NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    balance_after DECIMAL(10,2) NOT NULL,
    note TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_credit_customer (customer_id)
);

-- POS stock is tracked through the ERP unified stock layer:
-- stock_locations, stock_levels, and stock_ledger keyed by erp_products.id.

-- ── Scheduled pricing ────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS pos_scheduled_prices (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    branch_id BIGINT DEFAULT NULL,
    price DECIMAL(10,2) NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    days_of_week VARCHAR(20) DEFAULT '0123456',
    name VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_scheduled_product (product_id)
);

-- ── ERP product fields used by POS ───────────────────────────────
-- (Run these against erp_products when bootstrapping an older database)
-- ALTER TABLE erp_products
--   ADD COLUMN IF NOT EXISTS barcode VARCHAR(100) DEFAULT NULL,
--   ADD COLUMN IF NOT EXISTS tax_class VARCHAR(50) DEFAULT 'standard',
--   ADD COLUMN IF NOT EXISTS variant_of BIGINT DEFAULT NULL,
--   ADD COLUMN IF NOT EXISTS variant_name VARCHAR(100) DEFAULT NULL;

-- ── Add loyalty_points and store_credit to customers ─────────────
-- ALTER TABLE pos_customers
--   ADD COLUMN IF NOT EXISTS loyalty_points INT DEFAULT 0,
--   ADD COLUMN IF NOT EXISTS store_credit DECIMAL(10,2) DEFAULT 0;
