CREATE TABLE IF NOT EXISTS plans (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) NOT NULL UNIQUE,
    apps JSON NOT NULL COMMENT 'Array of app IDs included in plan',
    price_per_seat_cents INT NOT NULL DEFAULT 0,
    trial_days INT NOT NULL DEFAULT 14,
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

-- Tracks each contiguous period of seat usage (Google Workspace model).
-- When seat count changes mid-month the current period is closed (period_end set)
-- and a new period opens (period_end NULL). The invoice for a billing month is
-- computed by summing: seat_count * days_in_period * (price_per_seat_cents / days_in_month).
CREATE TABLE IF NOT EXISTS seat_periods (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    subscription_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    seat_count INT NOT NULL,
    price_per_seat_cents INT NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NULL COMMENT 'NULL = currently active period',
    billing_month DATE NOT NULL COMMENT 'First day of the month this period belongs to',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_sp_subscription (subscription_id),
    INDEX idx_sp_owner_month (owner_id, billing_month)
);

INSERT IGNORE INTO plans (name, slug, apps, price_per_seat_cents, trial_days) VALUES
  ('Flow Solo',        'flow_solo',   '["flow"]',                                           400,  14),
  ('Finance Solo',     'finance_solo','["finance"]',                                         300,  14),
  ('Business Starter', 'starter',     '["dash","flow","finance","crm"]',                     800,  14),
  ('Business Pro',     'pro',         '["dash","flow","finance","crm","hr","erp"]',          1200, 14),
  ('Enterprise',       'enterprise',  '["dash","flow","finance","crm","hr","erp","pos"]',    1600, 14);
