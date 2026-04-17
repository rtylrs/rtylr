CREATE TABLE IF NOT EXISTS plans (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) NOT NULL UNIQUE,
    apps JSON NOT NULL COMMENT 'Array of app IDs included in plan',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL UNIQUE,
    plan_id BIGINT NOT NULL,
    status ENUM('active','trial','cancelled','expired') DEFAULT 'trial',
    trial_ends_at TIMESTAMP NULL,
    current_period_ends_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_sub_owner (owner_id),
    INDEX idx_sub_plan (plan_id)
);

INSERT IGNORE INTO plans (name, slug, apps) VALUES
  ('Flow Solo',       'flow_solo',   '["flow"]'),
  ('Finance Solo',    'finance_solo','["finance"]'),
  ('Business Starter','starter',     '["hub","flow","finance","crm"]'),
  ('Business Pro',    'pro',         '["hub","flow","finance","crm","hr","erp"]'),
  ('Enterprise',      'enterprise',  '["hub","flow","finance","crm","hr","erp","pos"]');
