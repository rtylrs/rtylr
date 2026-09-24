CREATE TABLE IF NOT EXISTS hr_equipment_requests (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_id BIGINT NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    reason TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    current_approval_level INT NOT NULL DEFAULT 0,
    approved_by BIGINT,
    rejection_reason TEXT,
    needed_by_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_equipment_requests_owner (owner_id),
    INDEX idx_equipment_requests_employee_status (owner_id, employee_id, status)
);

CREATE TABLE IF NOT EXISTS hr_assets (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    serial_number VARCHAR(255),
    location_id BIGINT,
    assigned_to BIGINT,
    status VARCHAR(20) NOT NULL DEFAULT 'available',
    `condition` VARCHAR(20) NOT NULL DEFAULT 'good',
    purchase_date DATE,
    purchase_price DECIMAL(12,2),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_assets_owner_status (owner_id, status),
    INDEX idx_assets_owner_location (owner_id, location_id)
);
