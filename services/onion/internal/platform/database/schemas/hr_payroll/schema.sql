CREATE TABLE IF NOT EXISTS hr_payroll_config (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    period_type VARCHAR(50) NOT NULL DEFAULT 'monthly',
    pay_day_of_month INT NOT NULL DEFAULT 28,
    overtime_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_payroll_config_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS hr_payroll_runs (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    created_by BIGINT NOT NULL,
    approved_by BIGINT,
    exported_at TIMESTAMP NULL,
    total_gross DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_net DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_payroll_runs_owner (owner_id),
    INDEX idx_payroll_runs_period (owner_id, period_start, period_end),
    INDEX idx_payroll_runs_status (owner_id, status),
    UNIQUE KEY uk_payroll_runs_business_period (business_id, period_start, period_end)
);

CREATE TABLE IF NOT EXISTS hr_payroll_entries (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    payroll_run_id BIGINT NOT NULL,
    employee_id BIGINT NOT NULL,
    hours_worked DECIMAL(8,2) NOT NULL DEFAULT 0,
    overtime_hours DECIMAL(8,2) NOT NULL DEFAULT 0,
    base_pay DECIMAL(15,2) NOT NULL DEFAULT 0,
    overtime_pay DECIMAL(15,2) NOT NULL DEFAULT 0,
    bonuses DECIMAL(15,2) NOT NULL DEFAULT 0,
    deductions DECIMAL(15,2) NOT NULL DEFAULT 0,
    net_pay DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_payroll_entries_owner_run_employee (owner_id, payroll_run_id, employee_id),
    INDEX idx_payroll_entries_owner (owner_id),
    INDEX idx_payroll_entries_run (owner_id, payroll_run_id),
    INDEX idx_payroll_entries_employee (owner_id, employee_id)
);
