CREATE TABLE IF NOT EXISTS users (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NULL,
    email VARCHAR(320) UNIQUE NOT NULL,
    password_hash VARCHAR(512) NOT NULL,
    full_name VARCHAR(255) NULL,
    signup_ip_address VARCHAR(45) NULL,
    signup_device_info TEXT NULL,
    onboarded BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    is_console_admin BOOLEAN NOT NULL DEFAULT FALSE,
    is_banned BOOLEAN NOT NULL DEFAULT FALSE,
    ban_reason TEXT NULL,
    banned_at TIMESTAMP NULL,
    auth_provider VARCHAR(16) NOT NULL DEFAULT 'password',
    two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    INDEX idx_users_business (business_id),
    INDEX idx_users_console_admin (is_console_admin),
    INDEX idx_users_banned (is_banned)
);

CREATE TABLE IF NOT EXISTS businesses (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    owner_id BIGINT NOT NULL,
    industry VARCHAR(100),
    country VARCHAR(100),
    timezone VARCHAR(50) DEFAULT 'UTC',
    currency VARCHAR(10) DEFAULT 'USD',
    logo_url VARCHAR(512),
    slug VARCHAR(80) NULL UNIQUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_business_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS plans (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) NOT NULL UNIQUE,
    apps JSON NOT NULL,
    price_per_seat_cents INT NOT NULL DEFAULT 0,
    paddle_price_id VARCHAR(100) NULL,
    trial_days INT NOT NULL DEFAULT 14,
    tagline VARCHAR(200) NOT NULL DEFAULT '',
    annual_price_per_seat_cents INT NOT NULL DEFAULT 0,
    max_seats INT NULL,
    max_locations INT NULL,
    ai_credits_per_month INT NOT NULL DEFAULT 0,
    features JSON NULL,
    tier_order INT NOT NULL DEFAULT 100,
    is_solo BOOLEAN NOT NULL DEFAULT FALSE,
    is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL UNIQUE,
    plan_id BIGINT NOT NULL,
    status ENUM('active','trial','cancelled','expired') DEFAULT 'trial',
    seat_count INT NOT NULL DEFAULT 1,
    trial_ends_at TIMESTAMP NULL,
    next_billing_date DATE NULL,
    current_period_ends_at TIMESTAMP NULL,
    paddle_subscription_id VARCHAR(100) NULL,
    paddle_customer_id VARCHAR(100) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_sub_owner (owner_id),
    INDEX idx_sub_plan (plan_id)
);

CREATE TABLE IF NOT EXISTS manual_access_grants (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    plan_id BIGINT NOT NULL,
    granted_by BIGINT NOT NULL,
    note TEXT NULL,
    expires_at TIMESTAMP NULL,
    revoked_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_mag_user (user_id),
    INDEX idx_mag_plan (plan_id)
);

-- One row, maintained by shared-schema triggers in production. Keeping the
-- shape here lets SQLC type-check the bounded console stats lookup.
CREATE TABLE IF NOT EXISTS console_stats (
    id TINYINT NOT NULL PRIMARY KEY,
    total_users BIGINT NOT NULL DEFAULT 0,
    banned_users BIGINT NOT NULL DEFAULT 0,
    total_businesses BIGINT NOT NULL DEFAULT 0,
    active_subs BIGINT NOT NULL DEFAULT 0,
    trial_subs BIGINT NOT NULL DEFAULT 0,
    manual_grants BIGINT NOT NULL DEFAULT 0,
    manual_addon_grants BIGINT NOT NULL DEFAULT 0,
    console_admins BIGINT NOT NULL DEFAULT 0,
    CHECK (id = 1)
);

CREATE TABLE IF NOT EXISTS addons (
    id VARCHAR(50) NOT NULL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    price DECIMAL(10,2) DEFAULT 0.00,
    is_free BOOLEAN DEFAULT FALSE,
    is_default_enabled BOOLEAN DEFAULT FALSE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS business_addons (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    addon_id VARCHAR(50) NOT NULL,
    is_enabled BOOLEAN DEFAULT TRUE,
    enabled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    disabled_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_business_addon (business_id, addon_id),
    INDEX idx_business_addons_business (business_id)
);

CREATE TABLE IF NOT EXISTS manual_addon_grants (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL,
    addon_id VARCHAR(50) NOT NULL,
    granted_by BIGINT NOT NULL,
    note TEXT NULL,
    expires_at TIMESTAMP NULL,
    revoked_at TIMESTAMP NULL,
    enabled_by_grant BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_manual_addon_grants_user (user_id),
    INDEX idx_manual_addon_grants_business (business_id)
);
