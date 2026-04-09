CREATE TABLE IF NOT EXISTS hr_departments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    manager_staff_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS hr_staff (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    user_id BIGINT,
    department_id BIGINT,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    role VARCHAR(100) NOT NULL,
    employment_type ENUM('full_time','part_time','contractor','intern') DEFAULT 'full_time',
    status ENUM('active','inactive','on_leave','terminated') DEFAULT 'active',
    salary DECIMAL(10,2),
    hourly_rate DECIMAL(10,2),
    overtime_rate DECIMAL(10,2),
    start_date DATE NOT NULL,
    end_date DATE,
    address TEXT,
    emergency_contact_name VARCHAR(200),
    emergency_contact_phone VARCHAR(50),
    apps JSON COMMENT 'Which rtylr apps this staff member can access',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS hr_shifts (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    staff_id BIGINT NOT NULL,
    department_id BIGINT,
    date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    break_duration_minutes INT DEFAULT 0,
    status ENUM('scheduled','in_progress','completed','absent','cancelled') DEFAULT 'scheduled',
    clock_in_at TIMESTAMP NULL,
    clock_out_at TIMESTAMP NULL,
    hours_worked DECIMAL(5,2),
    overtime_hours DECIMAL(5,2) DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS hr_time_off_requests (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    staff_id BIGINT NOT NULL,
    type ENUM('vacation','sick','personal','unpaid','other') NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    days_count DECIMAL(4,1) NOT NULL,
    status ENUM('pending','approved','rejected','cancelled') DEFAULT 'pending',
    reason TEXT,
    approved_by BIGINT,
    approved_at TIMESTAMP NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS hr_payroll_records (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    staff_id BIGINT NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    period_label VARCHAR(50) NOT NULL,
    regular_hours DECIMAL(6,2) DEFAULT 0,
    overtime_hours DECIMAL(6,2) DEFAULT 0,
    base_pay DECIMAL(10,2) NOT NULL DEFAULT 0,
    overtime_pay DECIMAL(10,2) DEFAULT 0,
    bonuses DECIMAL(10,2) DEFAULT 0,
    deductions DECIMAL(10,2) DEFAULT 0,
    tax_withheld DECIMAL(10,2) DEFAULT 0,
    net_pay DECIMAL(10,2) NOT NULL DEFAULT 0,
    status ENUM('draft','processed','paid') DEFAULT 'draft',
    paid_at TIMESTAMP NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
