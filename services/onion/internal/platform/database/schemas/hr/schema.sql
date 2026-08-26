CREATE TABLE IF NOT EXISTS hr_departments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    manager_staff_id BIGINT,
    location_id BIGINT DEFAULT NULL,
    parent_department_id BIGINT DEFAULT NULL,
    default_work_schedule_id BIGINT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS hr_staff (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT,
    employee_id BIGINT DEFAULT NULL,
    department_id BIGINT,
    manager_staff_id BIGINT NULL,
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
    apps JSON DEFAULT (JSON_ARRAY()) COMMENT 'Which rtylr apps this staff member can access',
    notes TEXT COMMENT 'HR-admin commentary. Not the profile bio: see bio below',
    photo_url VARCHAR(512) NULL,
    bio TEXT NULL COMMENT 'What the staff member says about themselves, shown on their profile',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    -- idx_hr_staff_search is added by zz_hr_staff_search_indexes.sql; declaring
    -- it here as well fails a fresh migration run with a duplicate-key error.
    INDEX idx_hr_staff_employee (owner_id, employee_id)
);

-- hr_shifts is defined in schemas/hr_shifts/schema.sql, which is its single
-- canonical definition. It used to be declared here as well; because migrate.sh
-- applies the hr directory before hr_shifts, this copy won and the hr_shifts
-- copy was a silent no-op, letting the two definitions drift apart. Modules
-- that query hr_shifts list schemas/hr_shifts in their own sqlc.yaml.

CREATE TABLE IF NOT EXISTS hr_time_off_requests (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
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
    business_id BIGINT NOT NULL DEFAULT 0,
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

CREATE TABLE IF NOT EXISTS hr_reviews (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    staff_id BIGINT NOT NULL,
    reviewer_id BIGINT,
    period VARCHAR(100),
    rating DECIMAL(3,1) DEFAULT 0,
    goals_rating DECIMAL(3,1) DEFAULT 0,
    values_rating DECIMAL(3,1) DEFAULT 0,
    skills_rating DECIMAL(3,1) DEFAULT 0,
    notes TEXT,
    status ENUM('draft','submitted','acknowledged') DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS hr_one_on_ones (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    manager_id BIGINT NOT NULL,
    staff_id BIGINT NOT NULL,
    scheduled_at TIMESTAMP NOT NULL,
    notes TEXT,
    action_items JSON DEFAULT (JSON_ARRAY()),
    status ENUM('scheduled','completed','cancelled') DEFAULT 'scheduled',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS hr_documents (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    staff_id BIGINT,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    file_url TEXT,
    expiry_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS hr_benefit_plans (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    type ENUM('health','dental','vision','401k','life','other') DEFAULT 'other',
    provider VARCHAR(255),
    cost_employer DECIMAL(10,2) DEFAULT 0,
    cost_employee DECIMAL(10,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS hr_training_courses (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    duration_hours DECIMAL(6,1) DEFAULT 0,
    is_required BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS hr_training_progress (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    course_id BIGINT NOT NULL,
    staff_id BIGINT NOT NULL,
    status ENUM('not_started','in_progress','completed') DEFAULT 'not_started',
    completed_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS hr_dashboard_widgets (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    role_id BIGINT NOT NULL DEFAULT 0,
    widget_type VARCHAR(100) NOT NULL,
    position_x INT NOT NULL DEFAULT 0,
    position_y INT NOT NULL DEFAULT 0,
    width INT NOT NULL DEFAULT 1,
    height INT NOT NULL DEFAULT 1,
    config JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_hr_widgets_owner (owner_id),
    INDEX idx_hr_widgets_owner_role (owner_id, role_id)
);

CREATE TABLE IF NOT EXISTS hr_employee_documents (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_id BIGINT NOT NULL,
    document_type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    storage_mode VARCHAR(50) NOT NULL DEFAULT 'url',
    file_url TEXT,
    external_url TEXT,
    expiry_date DATE NULL,
    uploaded_by BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_hr_docs_owner (owner_id),
    INDEX idx_hr_docs_employee (owner_id, employee_id),
    INDEX idx_hr_docs_expiry (owner_id, expiry_date)
);
