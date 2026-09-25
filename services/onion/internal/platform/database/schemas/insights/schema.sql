-- insights_settings stores the tenant's insights-wide preferences.
-- One row per owner; the service layer creates defaults on first read.
CREATE TABLE IF NOT EXISTS insights_settings (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    default_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    fiscal_year_start VARCHAR(5) NOT NULL DEFAULT '01-01',
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    default_period VARCHAR(20) NOT NULL DEFAULT 'this_month',
    compare_previous BOOLEAN NOT NULL DEFAULT FALSE,
    dashboard_widgets JSON NOT NULL DEFAULT (JSON_ARRAY()),
    favorite_reports JSON NOT NULL DEFAULT (JSON_ARRAY()),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_insights_settings_owner (owner_id)
);

-- insights_scheduled_reports stores subscriptions to recurring report
-- deliveries. Each row describes a template + cadence + delivery channel.
CREATE TABLE IF NOT EXISTS insights_scheduled_reports (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    template_id VARCHAR(100) NOT NULL,
    cadence VARCHAR(50) NOT NULL,
    channel VARCHAR(50) NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_scheduled_reports_owner (owner_id),
    INDEX idx_scheduled_reports_active (owner_id, is_active)
);
