CREATE TABLE IF NOT EXISTS crm_customers (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(50),
    date_of_birth DATE,
    gender ENUM('male','female','other','prefer_not_to_say'),
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100),
    loyalty_points INT DEFAULT 0,
    total_spent DECIMAL(10,2) DEFAULT 0,
    visit_count INT DEFAULT 0,
    last_visit_at TIMESTAMP NULL,
    loyalty_tier ENUM('bronze','silver','gold','platinum') DEFAULT 'bronze',
    tags JSON DEFAULT (JSON_ARRAY()),
    notes TEXT,
    consent JSON,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    -- idx_crm_customers_business_created / _business_tier / the FULLTEXT search
    -- index are added by zz_crm_customer_search_indexes.sql; declaring them here
    -- as well fails a fresh migration run with duplicate-key errors.
    KEY idx_crm_customers_business (business_id)
);

CREATE TABLE IF NOT EXISTS crm_loyalty_transactions (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    customer_id BIGINT NOT NULL,
    type ENUM('earn','redeem','adjust','expire','award') NOT NULL,
    points INT NOT NULL,
    balance_before INT NOT NULL DEFAULT 0,
    balance_after INT NOT NULL DEFAULT 0,
    description VARCHAR(255),
    reference_id BIGINT,
    reference_type VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    KEY idx_crm_loyalty_transactions_business (business_id)
);

CREATE TABLE IF NOT EXISTS crm_loyalty_config (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    is_enabled BOOLEAN DEFAULT TRUE,
    points_per_currency_unit DECIMAL(5,2) DEFAULT 1.0,
    currency_per_point DECIMAL(8,4) DEFAULT 0.01,
    min_redemption_points INT DEFAULT 0,
    bronze_min INT DEFAULT 0,
    silver_min INT DEFAULT 500,
    gold_min INT DEFAULT 1500,
    platinum_min INT DEFAULT 5000,
    bronze_discount_pct DECIMAL(5,2) DEFAULT 0,
    silver_discount_pct DECIMAL(5,2) DEFAULT 5,
    gold_discount_pct DECIMAL(5,2) DEFAULT 10,
    platinum_discount_pct DECIMAL(5,2) DEFAULT 15,
    points_expiry_days INT DEFAULT 365,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_crm_loyalty_config_business (business_id)
);

CREATE TABLE IF NOT EXISTS crm_segments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    rules JSON NOT NULL COMMENT 'Array of filter rules',
    customer_count INT DEFAULT 0,
    last_calculated_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_crm_segments_business (business_id)
);

CREATE TABLE IF NOT EXISTS crm_campaigns (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    channel ENUM('email','sms','push','whatsapp') NOT NULL,
    status ENUM('draft','scheduled','sent','active','paused','completed','cancelled') DEFAULT 'draft',
    segment_id BIGINT,
    subject VARCHAR(500),
    content TEXT NOT NULL,
    scheduled_at TIMESTAMP NULL,
    sent_at TIMESTAMP NULL,
    audience_count INT DEFAULT 0,
    sent_count INT DEFAULT 0,
    delivered_count INT DEFAULT 0,
    opened_count INT DEFAULT 0,
    clicked_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_crm_campaigns_business (business_id)
);

CREATE TABLE IF NOT EXISTS crm_campaign_sends (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    campaign_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    customer_id BIGINT NOT NULL,
    status ENUM('queued','sent','delivered','opened','clicked','bounced','failed') DEFAULT 'queued',
    sent_at TIMESTAMP NULL,
    opened_at TIMESTAMP NULL,
    clicked_at TIMESTAMP NULL,
    error_message TEXT,
    KEY idx_crm_campaign_sends_business (business_id)
);
