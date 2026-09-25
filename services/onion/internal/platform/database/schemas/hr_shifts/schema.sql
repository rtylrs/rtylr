-- hr_shifts holds two kinds of row, told apart by whether staff_id is set:
--
--   template rows   - a reusable shift for a location. Carry location_id, name
--                     and color; staff_id and date are NULL. Written by the
--                     hr_shifts module and referenced by roster assignments.
--   attendance rows - one staff member's worked shift. Carry staff_id and date
--                     plus the clock/hours columns; location_id and name are
--                     NULL. Written by the hr_shifts module (formerly the
--                     separate shifts module) and by staff clock-in/clock-out.
--
-- This is the single canonical definition. It was previously also declared in
-- schemas/hr/schema.sql, which won on apply order and left this copy a no-op,
-- so the two drifted: that one declared start_time/end_time as TIME and
-- location_id/name as nullable, while this one claimed VARCHAR(10) and NOT NULL.
-- The column types below match the live table, so re-applying is a no-op.
--
-- Consumers outside this module (timesheets, payroll, staff) list
-- schemas/hr_shifts in their own sqlc.yaml to resolve this table.
CREATE TABLE IF NOT EXISTS hr_shifts (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    staff_id BIGINT DEFAULT NULL,
    department_id BIGINT DEFAULT NULL,
    date DATE DEFAULT NULL,
    start_time TIME NOT NULL DEFAULT '00:00:00',
    end_time TIME NOT NULL DEFAULT '00:00:00',
    break_duration_minutes INT DEFAULT 0,
    status ENUM('scheduled','in_progress','completed','absent','cancelled') DEFAULT 'scheduled',
    clock_in_at TIMESTAMP NULL,
    clock_out_at TIMESTAMP NULL,
    hours_worked DECIMAL(5,2) DEFAULT NULL,
    overtime_hours DECIMAL(5,2) DEFAULT 0,
    notes TEXT,
    location_id BIGINT DEFAULT NULL,
    name VARCHAR(100) DEFAULT NULL,
    color VARCHAR(10) DEFAULT '#4CAF50',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_shifts_owner (owner_id),
    INDEX idx_shifts_business_status_date (business_id, status, date)
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
