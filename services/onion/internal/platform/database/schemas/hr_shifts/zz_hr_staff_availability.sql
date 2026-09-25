-- hr_staff_availability holds the windows in which a staff member can, or
-- explicitly cannot, be scheduled for a shift. It backs hr.shift.v1's
-- AvailabilityPayload and the List/Set/DeleteAvailability RPCs.
--
-- A row is one of two kinds, exactly as the contract describes:
--
--   recurring rule  - day_of_week set, specific_date NULL. The staff member's
--                     standing weekly pattern.
--   dated exception - specific_date set, day_of_week NULL. Overrides the
--                     recurring rules for that one calendar day.
--
-- day_of_week is 0 = Monday through 6 = Sunday, matching the contract. It is
-- deliberately not JavaScript's Date.getDay(), where 0 = Sunday; Monday-first
-- indexes the same week the roster already uses, since GetRosterRequest.week_start
-- is a Monday.
--
-- Scoped by business_id to match hr_shifts and hr_shift_assignments, which are
-- the tables this one is read alongside. owner_id is carried for parity with the
-- rest of the HR schema.
CREATE TABLE IF NOT EXISTS hr_staff_availability (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    -- References hr_staff.id, not hr_employees.id. These are different tables
    -- bridged by the nullable hr_staff.employee_id; see the note below.
    staff_id BIGINT NOT NULL,
    day_of_week TINYINT DEFAULT NULL,
    specific_date DATE DEFAULT NULL,
    start_time TIME NOT NULL,
    -- An end_time earlier than start_time means the window crosses midnight and
    -- ends the following day, which is how availability for a night shift is
    -- expressed. Readers must not assume end_time > start_time.
    end_time TIME NOT NULL,
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- Exactly one of day_of_week and specific_date is set on every row. The
    -- contract states this invariant but `.onk` has no cross-field decorator to
    -- express it, so it is enforced in the handler and again here.
    CONSTRAINT chk_hr_staff_availability_kind
        CHECK ((day_of_week IS NULL) <> (specific_date IS NULL)),

    -- The natural key the contract's upsert relies on is
    -- (business_id, staff_id, start_time, and whichever of day_of_week /
    -- specific_date is set). A plain UNIQUE KEY over the nullable columns would
    -- enforce nothing, because MySQL treats each NULL as distinct and so admits
    -- unlimited duplicates. These STORED generated columns fold the NULL into a
    -- sentinel so the constraint actually binds.
    day_key TINYINT GENERATED ALWAYS AS (IFNULL(day_of_week, -1)) STORED,
    date_key DATE GENERATED ALWAYS AS (IFNULL(specific_date, '1000-01-01')) STORED,
    UNIQUE KEY uk_hr_staff_availability (business_id, staff_id, day_key, date_key, start_time),

    INDEX idx_hr_staff_availability_owner (owner_id),
    -- Covers the self-service read and the scheduler's per-staff read.
    INDEX idx_hr_staff_availability_staff (business_id, staff_id),
    -- Covers the assignment-time conflict lookup, which filters on a date.
    INDEX idx_hr_staff_availability_date (business_id, specific_date)
);
