CREATE TABLE IF NOT EXISTS hr_company_config (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    company_name VARCHAR(255) NOT NULL DEFAULT '',
    fiscal_year_start_month INT NOT NULL DEFAULT 1,
    default_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    default_timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    default_work_schedule_id BIGINT,
    overtime_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    overtime_multiplier DECIMAL(4,2) NOT NULL DEFAULT 1.50,
    directory_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_hr_config_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS hr_setup_steps (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    step_key VARCHAR(100) NOT NULL,
    step_label VARCHAR(255) NOT NULL DEFAULT '',
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_hr_setup_steps_owner_step (owner_id, step_key)
);

CREATE TABLE IF NOT EXISTS hr_locations (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(80) NULL,
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100),
    timezone VARCHAR(100),
    latitude DOUBLE NULL,
    longitude DOUBLE NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_hr_locations_owner (owner_id),
    UNIQUE KEY uniq_hr_locations_owner_slug (owner_id, slug)
);

CREATE TABLE IF NOT EXISTS hr_work_schedules (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    schedule_data JSON,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_hr_work_schedules_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS hr_departments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    location_id BIGINT NOT NULL,
    parent_department_id BIGINT NULL,
    name VARCHAR(255) NOT NULL,
    default_work_schedule_id BIGINT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_hr_departments_owner_location (owner_id, location_id)
);

CREATE TABLE IF NOT EXISTS hr_employees (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT,
    department_id BIGINT,
    location_id BIGINT,
    work_schedule_id BIGINT,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    employment_type VARCHAR(50) NOT NULL DEFAULT 'full_time',
    status VARCHAR(50) NOT NULL DEFAULT 'invited',
    salary DECIMAL(10,2),
    hourly_rate DECIMAL(10,2),
    overtime_rate DECIMAL(10,2),
    start_date DATE,
    end_date DATE,
    address TEXT,
    emergency_contact_name VARCHAR(200),
    emergency_contact_phone VARCHAR(50),
    notes TEXT,
    clock_in_pin_hash VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_hr_employees_owner (owner_id),
    INDEX idx_hr_employees_email (owner_id, email),
    INDEX idx_hr_employees_business_email (business_id, email),
    INDEX idx_hr_employees_dept (owner_id, department_id),
    INDEX idx_hr_employees_location (owner_id, location_id),
    UNIQUE KEY uniq_employee_user_id (user_id)
);

CREATE TABLE IF NOT EXISTS hr_roles (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    permissions JSON,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_hr_role_name (owner_id, name),
    INDEX idx_hr_roles_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS hr_employee_roles (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_hr_employee_role (owner_id, employee_id, role_id),
    INDEX idx_hr_employee_roles_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS hr_approval_chains (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    request_type VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_hr_chain_type (owner_id, request_type),
    INDEX idx_hr_chains_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS hr_approval_chain_levels (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    approval_chain_id BIGINT NOT NULL,
    level_order INT NOT NULL,
    approver_type VARCHAR(50) NOT NULL,
    approver_role_id BIGINT,
    approver_employee_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_hr_levels_chain (approval_chain_id)
);

CREATE TABLE IF NOT EXISTS hr_employee_invites (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_id BIGINT NOT NULL,
    token VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    consumed_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_hr_invite_token (token),
    INDEX idx_hr_invites_owner (owner_id),
    INDEX idx_hr_invites_employee (employee_id)
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
