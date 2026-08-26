CREATE TABLE IF NOT EXISTS pos_locations (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    phone VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    timezone VARCHAR(50) DEFAULT 'UTC',
    currency VARCHAR(10) DEFAULT 'USD',
    tax_rate DECIMAL(5,2) DEFAULT 0.00,
    receipt_header TEXT,
    receipt_footer TEXT,
    business_id BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pos_locations_owner (owner_id),
    INDEX idx_pos_locations_business_active_name (business_id, is_active, name)
);

CREATE TABLE IF NOT EXISTS pos_tables (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    location_id BIGINT,
    name VARCHAR(50) NOT NULL,
    capacity INT DEFAULT 4,
    status ENUM('available','occupied','reserved','cleaning') DEFAULT 'available',
    current_session_order_id BIGINT DEFAULT NULL,
    qr_code_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_pos_tables_owner (owner_id),
    INDEX idx_pos_tables_owner_location (owner_id, location_id),
    INDEX idx_pos_tables_business_location_status_name (business_id, location_id, status, name)
);

CREATE TABLE IF NOT EXISTS pos_sessions (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    location_id BIGINT,
    staff_id BIGINT NOT NULL,
    terminal_id BIGINT DEFAULT NULL,
    status ENUM('open','closed') DEFAULT 'open',
    opening_cash DECIMAL(10,2) DEFAULT 0,
    closing_cash DECIMAL(10,2),
    expected_cash DECIMAL(10,2),
    cash_difference DECIMAL(10,2),
    total_sales DECIMAL(10,2) DEFAULT 0,
    total_orders INT DEFAULT 0,
    total_cash_sales DECIMAL(10,2) DEFAULT 0,
    total_card_sales DECIMAL(10,2) DEFAULT 0,
    total_online_sales DECIMAL(10,2) DEFAULT 0,
    total_refunds DECIMAL(10,2) DEFAULT 0,
    opened_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMP NULL,
    notes TEXT,
    INDEX idx_pos_sessions_owner_status (owner_id, status),
    INDEX idx_pos_sessions_owner_location (owner_id, location_id),
    INDEX idx_pos_sessions_staff (staff_id),
    INDEX idx_pos_sessions_business_staff_status (business_id, staff_id, status),
    INDEX idx_pos_sessions_business_location_opened (business_id, location_id, opened_at)
);

CREATE TABLE IF NOT EXISTS pos_orders (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    session_id BIGINT,
    table_id BIGINT,
    location_id BIGINT,
    order_number VARCHAR(50) NOT NULL,
    type ENUM('dine_in','takeaway','delivery','online') DEFAULT 'dine_in',
    status ENUM('pending','confirmed','preparing','ready','served','paid','cancelled','refunded','voided','held') DEFAULT 'pending',
    customer_id BIGINT,
    customer_name VARCHAR(255),
    customer_phone VARCHAR(50),
    staff_id BIGINT,
    subtotal DECIMAL(10,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    total DECIMAL(10,2) DEFAULT 0,
    payment_method ENUM('cash','card','online','split') DEFAULT 'cash',
    payment_status ENUM('unpaid','partial','paid','refunded') DEFAULT 'unpaid',
    amount_paid DECIMAL(10,2) DEFAULT 0,
    change_due DECIMAL(10,2) DEFAULT 0,
    notes TEXT,
    kitchen_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    paid_at TIMESTAMP NULL,
    terminal_id BIGINT DEFAULT NULL,
    branch_id BIGINT DEFAULT NULL,
    tip_amount DECIMAL(10,2) DEFAULT 0,
    void_reason TEXT DEFAULT NULL,
    voided_by BIGINT DEFAULT NULL,
    voided_at TIMESTAMP NULL,
    exchange_order_id BIGINT DEFAULT NULL,
    paddle_transaction_id VARCHAR(64) DEFAULT NULL,
    UNIQUE KEY uq_pos_orders_number (owner_id, order_number),
    INDEX idx_pos_orders_owner_status (owner_id, status),
    INDEX idx_pos_orders_owner_session (owner_id, session_id),
    INDEX idx_pos_orders_owner_location (owner_id, location_id),
    INDEX idx_pos_orders_owner_created (owner_id, created_at),
    -- idx_pos_orders_business_status_created / _business_session_created /
    -- _business_created and the FULLTEXT search index are added by
    -- zz_pos_order_search_indexes.sql; declaring them here as well fails a
    -- fresh migration run with duplicate-key errors.
    INDEX idx_pos_orders_business_order_number (business_id, order_number),
    INDEX idx_pos_orders_business_paid_at (business_id, payment_status, paid_at),
    INDEX idx_pos_orders_business_customer_payment (business_id, customer_id, payment_status),
    INDEX idx_pos_orders_session_payment (session_id, payment_status),
    INDEX idx_pos_orders_paddle_transaction (paddle_transaction_id),
    INDEX idx_pos_orders_customer (customer_id),
    INDEX idx_pos_orders_staff (staff_id)
);

CREATE TABLE IF NOT EXISTS pos_order_items (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT NOT NULL,
    menu_item_id BIGINT,
    name VARCHAR(255) NOT NULL,
    sku VARCHAR(100),
    quantity DECIMAL(8,3) NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    discount_pct DECIMAL(5,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    tax_rate DECIMAL(5,2) DEFAULT 0,
    total DECIMAL(10,2) NOT NULL,
    notes TEXT,
    status ENUM('pending','preparing','ready','served','cancelled') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pos_order_items_order (order_id),
    INDEX idx_pos_order_items_order_status (order_id, status),
    INDEX idx_pos_order_items_menu_item_order (menu_item_id, order_id)
);

CREATE TABLE IF NOT EXISTS pos_refunds (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    order_id BIGINT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    reason TEXT,
    refunded_by BIGINT,
    refunded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pos_refunds_owner (owner_id),
    INDEX idx_pos_refunds_order (order_id),
    INDEX idx_pos_refunds_business_refunded_at (business_id, refunded_at)
);

CREATE TABLE IF NOT EXISTS pos_modifiers (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    price_delta DECIMAL(10,2) NOT NULL DEFAULT 0,
    group_name VARCHAR(255),
    is_required BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_pos_modifiers_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS pos_combos (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL DEFAULT 0,
    items JSON,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_pos_combos_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS pos_gift_cards (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    customer_id BIGINT,
    code VARCHAR(50) NOT NULL UNIQUE,
    initial_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    current_balance DECIMAL(10,2) NOT NULL DEFAULT 0,
    customer_name VARCHAR(255),
    customer_email VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_pos_gift_cards_owner (owner_id),
    INDEX idx_pos_gift_cards_customer (owner_id, customer_id)
);

-- Gift-card liability journal. Every issuance, redemption, top-up, refund,
-- or adjustment must retain the business order/tender reference and the
-- resulting balance so the stored-value liability can be audited and
-- reconciled without relying on the mutable card balance alone.
CREATE TABLE IF NOT EXISTS pos_gift_card_transactions (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    gift_card_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    type ENUM('issue','redeem','top_up','refund','adjustment') NOT NULL,
    order_ref VARCHAR(128) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    balance_after DECIMAL(10,2) NOT NULL,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_pos_gift_card_transaction_ref (gift_card_id, type, order_ref),
    INDEX idx_pos_gift_card_transactions_business (business_id, created_at),
    INDEX idx_pos_gift_card_transactions_card (gift_card_id, created_at),
    INDEX idx_pos_gift_card_transactions_order_ref (business_id, order_ref)
);

CREATE TABLE IF NOT EXISTS pos_manager_pins (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    staff_id BIGINT NOT NULL,
    pin VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pos_manager_pins_owner_staff (owner_id, staff_id)
);

-- pos_users is a team member's PIN credential at a single location. user_id
-- links back to the users table (the team member identity); a person has at
-- most one pos_users row per location. role/capability flags default from the
-- team role at PIN setup but are editable per-user in the staff screen.
CREATE TABLE IF NOT EXISTS pos_users (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL,
    location_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    pin_hash VARCHAR(255) DEFAULT NULL,
    role ENUM('manager','cashier','server','bartender','kitchen') DEFAULT 'cashier',
    is_active BOOLEAN DEFAULT TRUE,
    hourly_rate DECIMAL(10,2) DEFAULT 0,
    max_discount_pct DECIMAL(5,2) DEFAULT 0,
    can_void BOOLEAN DEFAULT FALSE,
    can_refund BOOLEAN DEFAULT FALSE,
    can_open_drawer BOOLEAN DEFAULT FALSE,
    can_apply_discount BOOLEAN DEFAULT FALSE,
    created_by BIGINT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_pos_user_location (user_id, location_id),
    INDEX idx_pos_users_owner (owner_id),
    INDEX idx_pos_users_location (location_id),
    INDEX idx_pos_users_business_name (business_id, name),
    INDEX idx_pos_users_business_location_active_name (business_id, location_id, is_active, name),
    INDEX idx_pos_users_business_location_user_active (business_id, location_id, user_id, is_active)
);

CREATE TABLE IF NOT EXISTS pos_member_locations (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    location_id BIGINT NOT NULL,
    can_manage_pos_users BOOLEAN DEFAULT FALSE,
    can_manage_terminals BOOLEAN DEFAULT FALSE,
    can_view_reports BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_member_location (user_id, location_id),
    INDEX idx_member_loc_business (business_id),
    INDEX idx_member_loc_business_user_location (business_id, user_id, location_id),
    INDEX idx_member_loc_business_location_user (business_id, location_id, user_id)
);

-- POS-local customer directory (loyalty/store-credit/visit tracking).
CREATE TABLE IF NOT EXISTS pos_customers (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) DEFAULT NULL,
    phone VARCHAR(50) DEFAULT NULL,
    notes TEXT,
    loyalty_points INT DEFAULT 0,
    store_credit DECIMAL(10,2) DEFAULT 0.00,
    visit_count INT DEFAULT 0,
    total_spent DECIMAL(10,2) DEFAULT 0.00,
    last_visit_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_pos_customers_business (business_id),
    INDEX idx_pos_customers_phone (phone)
);

CREATE TABLE IF NOT EXISTS pos_business_settings (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    business_profile ENUM('retail','restaurant','quick_service') NOT NULL DEFAULT 'retail',
    theme ENUM('dark','light') NOT NULL DEFAULT 'dark',
    features JSON NOT NULL,
    receipt_footer TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_pos_business_settings_business (business_id),
    INDEX idx_pos_business_settings_profile (business_profile)
);
