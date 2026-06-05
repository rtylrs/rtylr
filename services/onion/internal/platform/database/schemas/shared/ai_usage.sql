-- Monthly AI token usage, scoped to the canonical tenant (businesses.owner_id).
-- user_id distinguishes the allocation bucket:
--   user_id = 0  → the shared team pool (default mode)
--   user_id > 0  → that user's personal bucket (per-person mode)
-- The monthly allowance is derived from the owner's plan
-- (plans.ai_credits_per_month + a per-seat bonus, converted to tokens); this
-- table only tracks consumption.
CREATE TABLE IF NOT EXISTS ai_usage_monthly (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL COMMENT 'Canonical tenant = businesses.owner_id',
    user_id BIGINT NOT NULL DEFAULT 0 COMMENT '0 = shared team pool; >0 = per-person bucket',
    billing_month DATE NOT NULL COMMENT 'First day of the calendar month (UTC)',
    tokens_used BIGINT NOT NULL DEFAULT 0 COMMENT 'Combined prompt + completion tokens',
    requests INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_ai_usage_owner_user_month (owner_id, user_id, billing_month),
    INDEX idx_ai_usage_owner (owner_id)
);

-- Per-tenant AI allocation policy, set by an admin in dash settings.
--   pool       → the whole team shares one monthly allowance (default)
--   per_person → the monthly allowance is split evenly across seats
CREATE TABLE IF NOT EXISTS ai_quota_settings (
    owner_id BIGINT NOT NULL PRIMARY KEY COMMENT 'Canonical tenant = businesses.owner_id',
    mode ENUM('pool','per_person') NOT NULL DEFAULT 'pool',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
