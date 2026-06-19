CREATE TABLE IF NOT EXISTS hr_shifts (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    location_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    start_time VARCHAR(10) NOT NULL,
    end_time VARCHAR(10) NOT NULL,
    color VARCHAR(10) NOT NULL DEFAULT '#4CAF50',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_shifts_owner (owner_id),
    INDEX idx_shifts_location (owner_id, location_id)
);

CREATE TABLE IF NOT EXISTS hr_shift_assignments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_id BIGINT NOT NULL,
    shift_id BIGINT NOT NULL,
    date DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'scheduled',
    published BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_shift_assignment (owner_id, employee_id, date),
    INDEX idx_shift_assignments_owner (owner_id),
    INDEX idx_shift_assignments_employee (owner_id, employee_id),
    INDEX idx_shift_assignments_date (owner_id, date)
);
