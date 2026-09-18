CREATE TABLE IF NOT EXISTS hr_overtime_requests (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_id BIGINT NOT NULL,
    date DATE NOT NULL,
    planned_extra_minutes INT NOT NULL,
    reason TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    current_approval_level INT NOT NULL DEFAULT 0,
    approved_by BIGINT,
    multiplier_applied DECIMAL(5,2),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_overtime_requests_owner (owner_id),
    INDEX idx_overtime_requests_employee (owner_id, employee_id),
    INDEX idx_overtime_requests_status (owner_id, status),
    INDEX idx_overtime_requests_date (owner_id, date)
);

CREATE TABLE IF NOT EXISTS hr_overtime_config (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    daily_threshold_minutes INT NOT NULL DEFAULT 480,
    default_multiplier DECIMAL(5,2) NOT NULL DEFAULT 1.50,
    auto_detect_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_overtime_config_owner (owner_id)
);
