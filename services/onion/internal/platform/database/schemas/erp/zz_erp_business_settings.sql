-- ERP business settings are tenant-scoped and deliberately contain no owner
-- foreign key: the authenticated business scope is applied by the service.
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
