CREATE TABLE IF NOT EXISTS hr_leave_types (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(100) NOT NULL,
    is_paid BOOLEAN NOT NULL DEFAULT TRUE,
    requires_approval BOOLEAN NOT NULL DEFAULT TRUE,
    accrual_rule VARCHAR(50),
    accrual_amount DECIMAL(10,2) DEFAULT 0,
    max_carryover_days INT NOT NULL DEFAULT 0,
    min_notice_days INT NOT NULL DEFAULT 0,
    requires_documentation_after_days INT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_leave_type_name (owner_id, name),
    INDEX idx_leave_types_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS hr_leave_requests (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_id BIGINT NOT NULL,
    leave_type_id BIGINT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    total_days DECIMAL(5,2) NOT NULL,
    notes TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    current_approval_level INT NOT NULL DEFAULT 0,
    rejection_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_leave_requests_owner (owner_id),
    INDEX idx_leave_requests_employee (owner_id, employee_id),
    INDEX idx_leave_requests_status (owner_id, status),
    INDEX idx_leave_requests_dates (owner_id, start_date, end_date)
);

CREATE TABLE IF NOT EXISTS hr_leave_balances (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_id BIGINT NOT NULL,
    leave_type_id BIGINT NOT NULL,
    year INT NOT NULL,
    entitled_days DECIMAL(10,2) NOT NULL DEFAULT 0,
    used_days DECIMAL(10,2) NOT NULL DEFAULT 0,
    pending_days DECIMAL(10,2) NOT NULL DEFAULT 0,
    last_accrual_month INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_leave_balance (owner_id, employee_id, leave_type_id, year),
    INDEX idx_leave_balances_owner (owner_id),
    INDEX idx_leave_balances_employee (owner_id, employee_id)
);
