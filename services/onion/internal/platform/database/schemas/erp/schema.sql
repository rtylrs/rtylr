CREATE TABLE IF NOT EXISTS erp_categories (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_erp_categories_business_name (business_id, name)
);

CREATE TABLE IF NOT EXISTS erp_suppliers (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100),
    tax_id VARCHAR(100),
    payment_terms VARCHAR(100),
    notes TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_erp_suppliers_business_active_name (business_id, is_active, name)
);

CREATE TABLE IF NOT EXISTS erp_products (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    category_id BIGINT,
    supplier_id BIGINT,
    brand_id BIGINT DEFAULT NULL,
    name VARCHAR(255) NOT NULL,
    sku VARCHAR(100),
    barcode VARCHAR(100),
    description TEXT,
    unit VARCHAR(50) DEFAULT 'unit',
    cost_price DECIMAL(10,2) NOT NULL DEFAULT 0,
    sale_price DECIMAL(10,2) NOT NULL DEFAULT 0,
    stock_quantity DECIMAL(10,3) NOT NULL DEFAULT 0,
    low_stock_threshold DECIMAL(10,3) DEFAULT 10,
    reorder_quantity DECIMAL(10,3) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    track_inventory BOOLEAN DEFAULT TRUE,
    -- PIM extensions (migration 005_product_pim.sql)
    compare_at_price DECIMAL(10,2) NOT NULL DEFAULT 0,
    tax_rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    weight DECIMAL(10,3) NOT NULL DEFAULT 0,
    weight_unit VARCHAR(8) NOT NULL DEFAULT 'kg',
    length DECIMAL(10,2) NOT NULL DEFAULT 0,
    width DECIMAL(10,2) NOT NULL DEFAULT 0,
    height DECIMAL(10,2) NOT NULL DEFAULT 0,
    dimension_unit VARCHAR(8) NOT NULL DEFAULT 'cm',
    country_of_origin VARCHAR(2) DEFAULT NULL,
    hs_code VARCHAR(20) DEFAULT NULL,
    allow_backorder BOOLEAN NOT NULL DEFAULT FALSE,
    lead_time_days INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    -- idx_erp_products_business_name / _business_category_name / _business_active
    -- and the FULLTEXT search index are added by
    -- zz_erp_product_search_indexes.sql; declaring them here as well fails a
    -- fresh migration run with duplicate-key errors.
    INDEX idx_erp_products_business_brand (business_id, brand_id),
    INDEX idx_erp_products_business_sku (business_id, sku),
    INDEX idx_erp_products_business_barcode (business_id, barcode),
    INDEX idx_erp_products_business_inventory_name (business_id, is_active, track_inventory, name),
    INDEX idx_erp_products_business_inventory_stock (business_id, is_active, track_inventory, stock_quantity)
);

CREATE TABLE IF NOT EXISTS erp_inventory_transactions (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    product_id BIGINT NOT NULL,
    type ENUM('purchase','sale','adjustment','return','waste') NOT NULL,
    quantity_change DECIMAL(10,3) NOT NULL,
    quantity_before DECIMAL(10,3) NOT NULL,
    quantity_after DECIMAL(10,3) NOT NULL,
    reference_id BIGINT,
    reference_type VARCHAR(50),
    notes TEXT,
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_erp_inventory_tx_business_product_created (business_id, product_id, created_at),
    INDEX idx_erp_inventory_tx_business_reference (business_id, reference_type, reference_id)
);

CREATE TABLE IF NOT EXISTS erp_purchase_orders (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    supplier_id BIGINT NOT NULL,
    po_number VARCHAR(100) NOT NULL,
    status ENUM('draft','sent','approved','partial','received','cancelled') DEFAULT 'draft',
    subtotal DECIMAL(10,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    total DECIMAL(10,2) DEFAULT 0,
    expected_delivery DATE,
    received_at TIMESTAMP NULL,
    notes TEXT,
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_erp_purchase_orders_business_created (business_id, created_at),
    INDEX idx_erp_purchase_orders_business_status_created (business_id, status, created_at)
);

CREATE TABLE IF NOT EXISTS erp_purchase_order_items (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    purchase_order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity_ordered DECIMAL(10,3) NOT NULL,
    quantity_received DECIMAL(10,3) DEFAULT 0,
    unit_cost DECIMAL(10,2) NOT NULL,
    total_cost DECIMAL(10,2) NOT NULL,
    INDEX idx_erp_po_items_order (purchase_order_id),
    INDEX idx_erp_po_items_product (product_id)
);

CREATE TABLE IF NOT EXISTS erp_business_settings (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    tax_rate DOUBLE NOT NULL DEFAULT 8.25,
    low_stock_threshold DOUBLE NOT NULL DEFAULT 10,
    default_warehouse_id BIGINT NOT NULL DEFAULT 1,
    notify_on_low_stock BOOLEAN NOT NULL DEFAULT TRUE,
    notify_on_expiry BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_erp_business_settings_business (business_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
